package tenant

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/models/tenant"
	tenantRepo "github.com/saas-multi-database-api/internal/repository/tenant"
)

type SettingHandler struct {
	settingRepo *tenantRepo.SettingRepository
}

func NewSettingHandler() *SettingHandler {
	return &SettingHandler{
		settingRepo: tenantRepo.NewSettingRepository(),
	}
}

// List retrieves all settings
func (h *SettingHandler) List(c *gin.Context) {
	pool := c.MustGet("tenant_pool").(*pgxpool.Pool)

	settings, err := h.settingRepo.List(c.Request.Context(), pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return empty array if no settings
	if settings == nil {
		settings = []tenant.Setting{}
	}

	c.JSON(http.StatusOK, tenant.SettingListResponse{
		Settings: settings,
	})
}

// GetByKey retrieves a specific setting by key
func (h *SettingHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")
	pool := c.MustGet("tenant_pool").(*pgxpool.Pool)

	setting, err := h.settingRepo.GetByKey(c.Request.Context(), pool, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// Update updates a setting value
func (h *SettingHandler) Update(c *gin.Context) {
	key := c.Param("key")
	pool := c.MustGet("tenant_pool").(*pgxpool.Pool)

	var req tenant.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.settingRepo.Update(c.Request.Context(), pool, key, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// Upsert creates or updates a setting
func (h *SettingHandler) Upsert(c *gin.Context) {
	key := c.Param("key")
	pool := c.MustGet("tenant_pool").(*pgxpool.Pool)

	var req tenant.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := h.settingRepo.Upsert(c.Request.Context(), pool, key, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

// Delete removes a setting
func (h *SettingHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	pool := c.MustGet("tenant_pool").(*pgxpool.Pool)

	if err := h.settingRepo.Delete(c.Request.Context(), pool, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "setting deleted successfully"})
}
