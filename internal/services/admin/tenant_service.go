package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/saas-multi-database-api/internal/models/admin"
	"github.com/saas-multi-database-api/internal/models/shared"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
	"github.com/saas-multi-database-api/internal/utils"
)

type TenantService struct {
	repo        *adminRepo.TenantRepository
	userRepo    *adminRepo.UserRepository
	redisClient *redis.Client
	masterPool  *pgxpool.Pool
}

func NewTenantService(
	repo *adminRepo.TenantRepository,
	userRepo *adminRepo.UserRepository,
	redisClient *redis.Client,
	masterPool *pgxpool.Pool,
) *TenantService {
	return &TenantService{
		repo:        repo,
		userRepo:    userRepo,
		redisClient: redisClient,
		masterPool:  masterPool,
	}
}

// CreateTenantRequest representa os dados para criar um tenant
type CreateTenantRequest struct {
	Name         string              `json:"name" binding:"required"`
	Subdomain    string              `json:"subdomain" binding:"required,min=3,max=50"` // User-chosen for public site
	URLCode      string              `json:"url_code,omitempty"`                        // Auto-generated if empty (admin routing)
	OwnerID      *uuid.UUID          `json:"owner_id,omitempty"`                        // Opcional: pode ser nil quando criado pela Admin API
	PlanID       uuid.UUID           `json:"plan_id" binding:"required"`
	BillingCycle shared.BillingCycle `json:"billing_cycle" binding:"required"`
	CompanyName  string              `json:"company_name"`
	IsCompany    bool                `json:"is_company"`
	CustomDomain string              `json:"custom_domain,omitempty"`
	Industry     string              `json:"industry,omitempty"` // Deprecated: usar custom_settings
}

// ProvisionEvent representa o evento de provisionamento publicado no Redis
type ProvisionEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	DBCode    string    `json:"db_code"`
	URLCode   string    `json:"url_code"`
	Timestamp time.Time `json:"timestamp"`
}

// CreateTenant cria um novo tenant de forma síncrona no Master DB
// e publica evento para provisionamento assíncrono do banco de dados
func (s *TenantService) CreateTenant(ctx context.Context, req CreateTenantRequest) (*admin.Tenant, error) {
	var err error

	// Validar se o owner existe (somente se fornecido)
	if req.OwnerID != nil {
		_, err = s.userRepo.GetUserByID(ctx, *req.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("owner not found: %w", err)
		}
	}

	// Normalizar e validar subdomain (para site público)
	subdomain := utils.NormalizeSlug(req.Subdomain)
	if len(subdomain) < 3 {
		return nil, fmt.Errorf("subdomain muito curto após normalização")
	}
	if len(subdomain) > 50 {
		return nil, fmt.Errorf("subdomain muito longo (máximo 50 caracteres)")
	}

	// Verificar se subdomain já existe
	existingSubdomain, _ := s.repo.GetTenantBySubdomain(ctx, subdomain)
	if existingSubdomain != nil {
		return nil, fmt.Errorf("subdomain já está em uso")
	}

	// Gerar url_code automaticamente (para admin panel)
	// Se foi fornecido (Admin API), usar; senão gerar
	urlCode := req.URLCode
	if urlCode == "" {
		urlCode = utils.GenerateURLCode()
	}

	// Verificar se URL code já existe (garantir unicidade)
	for attempts := 0; attempts < 10; attempts++ {
		existing, _ := s.repo.GetTenantByURLCode(ctx, urlCode)
		if existing == nil {
			break // Code is unique
		}
		urlCode = utils.GenerateURLCode() // Regenerate if collision
		if attempts == 9 {
			return nil, fmt.Errorf("falha ao gerar url_code único após 10 tentativas")
		}
	}

	// Gerar IDs e códigos
	tenantID := uuid.New()
	dbCode := uuid.New().String() // UUID completo como db_code

	// Criar tenant no Master DB com status 'provisioning'
	query := `
		INSERT INTO tenants (id, db_code, url_code, subdomain, owner_id, plan_id, billing_cycle, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, db_code, url_code, subdomain, owner_id, plan_id, billing_cycle, status, created_at, updated_at
	`

	now := time.Now()
	tenant := &admin.Tenant{}

	err = s.masterPool.QueryRow(
		ctx,
		query,
		tenantID,
		dbCode,
		urlCode,
		subdomain,
		req.OwnerID,
		req.PlanID,
		req.BillingCycle,
		"provisioning",
		now,
		now,
	).Scan(
		&tenant.ID,
		&tenant.DBCode,
		&tenant.URLCode,
		&tenant.Subdomain,
		&tenant.OwnerID,
		&tenant.PlanID,
		&tenant.BillingCycle,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar tenant: %w", err)
	}

	// Criar perfil do tenant
	customSettings := map[string]interface{}{
		"name":     req.Name,
		"industry": req.Industry,
	}

	settingsJSON, err := json.Marshal(customSettings)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar custom_settings: %w", err)
	}

	_, err = s.masterPool.Exec(
		ctx,
		`INSERT INTO tenant_profiles (tenant_id, company_name, is_company, custom_domain, custom_settings, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tenantID,
		req.CompanyName,
		req.IsCompany,
		req.CustomDomain,
		settingsJSON,
		now,
		now,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar perfil do tenant: %w", err)
	}

	// Adicionar owner como membro com role de owner (somente se owner_id foi fornecido)
	if req.OwnerID != nil {
		var ownerRoleID uuid.UUID
		err = s.masterPool.QueryRow(ctx, "SELECT id FROM roles WHERE slug = 'owner' LIMIT 1").Scan(&ownerRoleID)
		if err != nil {
			return nil, fmt.Errorf("role 'owner' não encontrada: %w", err)
		}

		_, err = s.masterPool.Exec(
			ctx,
			`INSERT INTO tenant_members (tenant_id, user_id, role_id, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			tenantID,
			req.OwnerID,
			ownerRoleID,
			now,
			now,
		)

		if err != nil {
			return nil, fmt.Errorf("erro ao adicionar owner como membro: %w", err)
		}
	}

	// Publicar evento para provisionamento assíncrono
	event := ProvisionEvent{
		TenantID:  tenantID,
		DBCode:    dbCode,
		URLCode:   urlCode,
		Timestamp: now,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar evento: %w", err)
	}

	// Publicar na fila do Redis
	err = s.redisClient.LPush(ctx, "tenant:provision:queue", eventJSON).Err()
	if err != nil {
		return nil, fmt.Errorf("erro ao publicar evento de provisionamento: %w", err)
	}

	// Cachear mapeamento url_code -> db_code
	cacheKey := fmt.Sprintf("tenant:urlcode:%s", urlCode)
	err = s.redisClient.Set(ctx, cacheKey, dbCode, 24*time.Hour).Err()
	if err != nil {
		// Log erro mas não falha a criação
		fmt.Printf("Warning: erro ao cachear tenant: %v\n", err)
	}

	return tenant, nil
}

// UpdateTenantStatus atualiza o status do tenant (usado pelo Worker)
func (s *TenantService) UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	query := `
		UPDATE tenants
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := s.masterPool.Exec(ctx, query, status, time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status do tenant: %w", err)
	}

	return nil
}

// GetTenantByID retorna um tenant pelo ID
func (s *TenantService) GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*admin.Tenant, error) {
	query := `
		SELECT id, db_code, url_code, owner_id, plan_id, status, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`

	tenant := &admin.Tenant{}
	err := s.masterPool.QueryRow(ctx, query, tenantID).Scan(
		&tenant.ID,
		&tenant.DBCode,
		&tenant.URLCode,
		&tenant.OwnerID,
		&tenant.PlanID,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("tenant não encontrado: %w", err)
	}

	return tenant, nil
}

// ListUserTenants retorna todos os tenants de um usuário
func (s *TenantService) ListUserTenants(ctx context.Context, userID uuid.UUID) ([]admin.UserTenant, error) {
	return s.repo.GetUserTenants(ctx, userID)
}
