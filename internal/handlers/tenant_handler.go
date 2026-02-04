package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/models"
	"github.com/saas-multi-database-api/internal/repository"
)

type TenantHandler struct {
	tenantRepo *repository.TenantRepository
}

func NewTenantHandler(tenantRepo *repository.TenantRepository) *TenantHandler {
	return &TenantHandler{
		tenantRepo: tenantRepo,
	}
}

// GetConfig returns the tenant configuration for the frontend
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

// Helper function to parse UUID
func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
