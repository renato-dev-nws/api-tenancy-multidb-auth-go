package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/saas-multi-database-api/internal/config"
	adminService "github.com/saas-multi-database-api/internal/services/admin"
)

// Worker responsável por processar eventos de provisionamento de tenants
func main() {
	log.Println("Iniciando Worker de Provisionamento de Tenants...")

	// Carregar configuração
	cfg := config.Load()

	// Conectar ao Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Conectar ao Master DB
	masterPool, err := pgxpool.New(context.Background(), cfg.MasterDB.ConnectionString())
	if err != nil {
		log.Fatalf("Erro ao conectar no Master DB: %v", err)
	}
	defer masterPool.Close()

	// Conectar ao Admin DB (usado para criar novos bancos)
	adminPool, err := pgxpool.New(context.Background(), cfg.AdminDB.ConnectionString())
	if err != nil {
		log.Fatalf("Erro ao conectar no Admin DB: %v", err)
	}
	defer adminPool.Close()

	log.Println("Conexões estabelecidas. Worker pronto para processar eventos.")

	// Canal para receber sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Canal para sinalizar que o worker deve parar
	stopChan := make(chan bool)

	// Goroutine para processar eventos
	go processEvents(redisClient, masterPool, adminPool, stopChan)

	// Aguardar sinal de interrupção
	<-sigChan
	log.Println("Recebido sinal de interrupção. Encerrando worker...")
	close(stopChan)

	// Aguardar um pouco para processar eventos pendentes
	time.Sleep(2 * time.Second)
	log.Println("Worker encerrado.")
}

// processEvents processa eventos da fila do Redis
func processEvents(redisClient *redis.Client, masterPool, adminPool *pgxpool.Pool, stopChan chan bool) {
	ctx := context.Background()
	queueKey := "tenant:provision:queue"

	for {
		select {
		case <-stopChan:
			log.Println("Parando processamento de eventos...")
			return
		default:
			// Bloquear por até 5 segundos esperando por eventos
			result, err := redisClient.BRPop(ctx, 5*time.Second, queueKey).Result()
			if err != nil {
				if err == redis.Nil {
					// Timeout, continuar loop
					continue
				}
				log.Printf("Erro ao ler da fila: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// result[0] é a chave, result[1] é o valor
			eventJSON := result[1]

			var event adminService.ProvisionEvent
			if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
				log.Printf("Erro ao deserializar evento: %v", err)
				continue
			}

			log.Printf("Processando provisionamento do tenant: %s (db_code: %s)", event.URLCode, event.DBCode)

			// Provisionar o tenant
			if err := provisionTenant(ctx, event, masterPool, adminPool); err != nil {
				log.Printf("Erro ao provisionar tenant %s: %v", event.URLCode, err)

				// Atualizar status para 'failed'
				updateTenantStatus(ctx, masterPool, event.TenantID, "failed")

				// Poderia republicar na fila com retry count
				continue
			}

			log.Printf("Tenant %s provisionado com sucesso!", event.URLCode)
		}
	}
}

// provisionTenant cria o banco de dados do tenant e aplica migrations
func provisionTenant(ctx context.Context, event adminService.ProvisionEvent, masterPool, adminPool *pgxpool.Pool) error {
	// Substituir hífens por underscores no db_code para nome válido de database
	dbCode := strings.ReplaceAll(event.DBCode, "-", "_")
	dbName := fmt.Sprintf("db_tenant_%s", dbCode)

	// 1. Criar banco de dados
	log.Printf("Criando database: %s", dbName)
	createDBQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := adminPool.Exec(ctx, createDBQuery)
	if err != nil {
		return fmt.Errorf("erro ao criar database: %w", err)
	}

	// 2. Conectar ao novo banco (direto no postgres, não via pgbouncer)
	tenantDSN := fmt.Sprintf("postgres://postgres:postgres@postgres:5432/%s?sslmode=disable", dbName)
	tenantPool, err := pgxpool.New(ctx, tenantDSN)
	if err != nil {
		return fmt.Errorf("erro ao conectar no tenant DB: %w", err)
	}
	defer tenantPool.Close()

	// 3. Aplicar schema do tenant
	log.Printf("Aplicando schema no database: %s", dbName)
	schema := getTenantSchema()
	_, err = tenantPool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("erro ao aplicar schema: %w", err)
	}

	// 4. Atualizar status para 'active' no Master DB
	log.Printf("Atualizando status do tenant para 'active'")
	if err := updateTenantStatus(ctx, masterPool, event.TenantID, "active"); err != nil {
		return fmt.Errorf("erro ao atualizar status: %w", err)
	}

	return nil
}

// updateTenantStatus atualiza o status do tenant no Master DB
func updateTenantStatus(ctx context.Context, masterPool *pgxpool.Pool, tenantID interface{}, status string) error {
	query := `UPDATE tenants SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := masterPool.Exec(ctx, query, status, time.Now(), tenantID)
	return err
}

// getTenantSchema retorna o schema SQL para databases de tenant
func getTenantSchema() string {
	return `
		-- Products table
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			sku VARCHAR(100) UNIQUE,
			price DECIMAL(10,2) NOT NULL,
			stock INTEGER DEFAULT 0,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Services table
		CREATE TABLE IF NOT EXISTS services (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			duration_minutes INTEGER,
			price DECIMAL(10,2) NOT NULL,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Customers table
		CREATE TABLE IF NOT EXISTS customers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE,
			phone VARCHAR(50),
			document VARCHAR(50),
			address JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Orders table
		CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			customer_id UUID REFERENCES customers(id),
			total DECIMAL(10,2) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			notes TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Order items table
		CREATE TABLE IF NOT EXISTS order_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
			product_id UUID REFERENCES products(id),
			service_id UUID REFERENCES services(id),
			quantity INTEGER NOT NULL DEFAULT 1,
			unit_price DECIMAL(10,2) NOT NULL,
			subtotal DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Settings table
		CREATE TABLE IF NOT EXISTS settings (
			key VARCHAR(100) PRIMARY KEY,
			value JSONB NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Insert default interface settings
		INSERT INTO settings (key, value) VALUES 
		('interface', '{"logo": null, "primary_color": "#003388", "secondary_color": "#DDDDDD"}');

		-- Images table (Polymorphic Association)
		CREATE TYPE media_type AS ENUM ('image', 'video', 'document');
		CREATE TYPE image_variant AS ENUM ('original', 'medium', 'small', 'thumb');

		CREATE TABLE IF NOT EXISTS images (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			imageable_type VARCHAR(50) NOT NULL,
			imageable_id UUID NOT NULL,
			filename VARCHAR(255) NOT NULL,
			original_filename VARCHAR(255),
			title VARCHAR(255),
			alt_text VARCHAR(255),
			media_type media_type NOT NULL DEFAULT 'image',
			mime_type VARCHAR(100) NOT NULL,
			extension VARCHAR(10) NOT NULL,
			variant image_variant NOT NULL DEFAULT 'original',
			parent_id UUID REFERENCES images(id) ON DELETE CASCADE,
			width INTEGER,
			height INTEGER,
			file_size BIGINT,
			storage_driver VARCHAR(20) NOT NULL DEFAULT 'local',
			storage_path TEXT NOT NULL,
			public_url TEXT,
			processing_status VARCHAR(20) DEFAULT 'pending',
			processed_at TIMESTAMP,
			display_order INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		-- Indexes
		CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
		CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
		CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
		CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
		CREATE INDEX IF NOT EXISTS idx_images_imageable ON images(imageable_type, imageable_id);
		CREATE INDEX IF NOT EXISTS idx_images_variant ON images(variant);
		CREATE INDEX IF NOT EXISTS idx_images_parent ON images(parent_id) WHERE parent_id IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_images_status ON images(processing_status);
		CREATE INDEX IF NOT EXISTS idx_images_display_order ON images(imageable_type, imageable_id, display_order);

		-- Grant permissions to saas_api user
		GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO saas_api;
		GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO saas_api;
	`
}
