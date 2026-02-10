package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/cache"
	"github.com/saas-multi-database-api/internal/config"
	"github.com/saas-multi-database-api/internal/database"
	tenantRepo "github.com/saas-multi-database-api/internal/repository/tenant"
	tenantService "github.com/saas-multi-database-api/internal/services/tenant"
	"github.com/saas-multi-database-api/internal/storage"
)

// Worker responsável por processar imagens de forma assíncrona
func main() {
	log.Println("Iniciando Image Processing Worker...")

	// Carregar configuração
	cfg := config.Load()

	ctx := context.Background()

	// Conectar ao Redis
	redisClient, err := cache.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Erro ao conectar no Redis: %v", err)
	}
	defer redisClient.Close()

	//Inicializar Database Manager
	dbManager := database.GetManager(cfg)

	// Inicializar Master Pool
	if err := dbManager.InitMasterPool(ctx); err != nil {
		log.Fatalf("Erro ao inicializar Master Pool: %v", err)
	}

	// Inicializar Storage Driver
	storageDriver, err := storage.NewStorageDriver(&storage.Config{
		Driver:      cfg.Storage.Driver,
		UploadsPath: cfg.Storage.UploadsPath,
	})
	if err != nil {
		log.Fatalf("Erro ao inicializar storage driver: %v", err)
	}

	log.Println("Conexões estabelecidas. Worker pronto para processar imagens.")

	// Canal para receber sinais de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Context para gerenciar lifecycle
	ctxWorker, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscriber para eventos de processamento de imagens
	pubsub := redisClient.Client.Subscribe(ctxWorker, "image:process")
	defer pubsub.Close()

	// Goroutine para processar mensagens
	go func() {
		for {
			select {
			case <-ctxWorker.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctxWorker)
				if err != nil {
					if ctxWorker.Err() != nil {
						return
					}
					log.Printf("Erro ao receber mensagem: %v", err)
					time.Sleep(1 * time.Second)
					continue
				}

				// Parse event
				var event tenantService.ProcessImageEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					log.Printf("Erro ao fazer parse do evento: %v", err)
					continue
				}

				log.Printf("Processando imagem: tenant=%s, image_id=%s", event.TenantDBCode, event.ImageID)

				// Processar imagem
				if err := processImage(ctxWorker, dbManager, storageDriver, event); err != nil {
					log.Printf("Erro ao processar imagem %s: %v", event.ImageID, err)
				} else {
					log.Printf("Imagem %s processada com sucesso", event.ImageID)
				}
			}
		}
	}()

	log.Println("Worker em execução. Aguardando eventos...")

	// Aguardar sinal de interrupção
	<-quit
	log.Println("Encerrando worker...")

	// Cancel context to stop goroutines
	cancel()

	// Wait a bit for cleanup
	time.Sleep(2 * time.Second)

	log.Println("Worker encerrado.")
}

// processImage processa uma imagem individual
func processImage(
	ctx context.Context,
	dbManager *database.Manager,
	storageDriver storage.StorageDriver,
	event tenantService.ProcessImageEvent,
) error {
	// Get tenant pool
	tenantPool, err := dbManager.GetTenantPool(ctx, event.TenantDBCode)
	if err != nil {
		return fmt.Errorf("failed to get tenant pool: %w", err)
	}

	// Criar repositório e serviço
	imageRepo := tenantRepo.NewImageRepository(tenantPool)
	processor := tenantService.NewImageProcessor(imageRepo, storageDriver)

	// Processar imagem
	if err := processor.ProcessImage(ctx, event.ImageID); err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}

	return nil
}

// ImageProcessEvent represents the event structure from Redis
type ImageProcessEvent struct {
	TenantDBCode string    `json:"tenant_db_code"`
	ImageID      uuid.UUID `json:"image_id"`
}
