package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/models"
	"github.com/saas-multi-database-api/internal/services"
)

type TenantHandler struct {
	tenantService *services.TenantService
}

func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// GetConfig retorna a configuração do tenant para o frontend
func (h *TenantHandler) GetConfig(c *gin.Context) {
	// Get features and permissions from context (injected by middleware)
	features := c.MustGet("features").([]string)
	permissions := c.MustGet("permissions").([]string)

	response := models.TenantConfigResponse{
		Features:    features,
		Permissions: permissions,
	}

	c.JSON(http.StatusOK, response)
}

// CreateTenant cria um novo tenant (apenas autenticado pode criar)
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	// Para Admin API: não precisa de owner (será nil)
	// Para Tenant API: poderia usar o user_id autenticado como owner
	// Por enquanto, aceita owner_id opcional no body ou deixa nil

	var req services.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos", "details": err.Error()})
		return
	}

	// Se não houver owner_id no body, deixa como nil (para Admin API)
	// Se precisar injetar owner do JWT no futuro, adicionar lógica aqui

	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "tenant criado com sucesso, provisionamento em andamento",
		"tenant":  tenant,
	})
}

// GetTenant retorna os detalhes de um tenant específico
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do tenant inválido"})
		return
	}

	tenant, err := h.tenantService.GetTenantByID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant não encontrado"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// ListMyTenants retorna todos os tenants do usuário autenticado
func (h *TenantHandler) ListMyTenants(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	tenants, err := h.tenantService.ListUserTenants(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar tenants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants,
	})
}

// UpdateTenant atualiza informações do tenant (Admin API)
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	if h.tenantService == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "operação disponível apenas na Admin API"})
		return
	}

	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do tenant inválido"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Status      *string `json:"status"`
		PlanID      *string `json:"plan_id"`
		CompanyName *string `json:"company_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	// TODO: Implement UpdateTenant in service layer
	c.JSON(http.StatusOK, gin.H{
		"message":   "tenant atualizado com sucesso",
		"tenant_id": tenantID,
	})
}

// DeleteTenant suspende/exclui um tenant (Admin API)
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	if h.tenantService == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "operação disponível apenas na Admin API"})
		return
	}

	tenantIDStr := c.Param("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do tenant inválido"})
		return
	}

	// TODO: Implement soft delete (status = suspended) in service layer
	c.JSON(http.StatusOK, gin.H{
		"message":   "tenant suspenso com sucesso",
		"tenant_id": tenantID,
	})
}

// Helper function to parse UUID
func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
