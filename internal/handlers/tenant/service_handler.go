package tenant

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tenantModels "github.com/saas-multi-database-api/internal/models/tenant"
	tenantRepo "github.com/saas-multi-database-api/internal/repository/tenant"
)

// ServiceHandler handles service operations for tenants
type ServiceHandler struct {
	serviceRepo *tenantRepo.ServiceRepository
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{
		serviceRepo: tenantRepo.NewServiceRepository(),
	}
}

// Create creates a new service
func (h *ServiceHandler) Create(c *gin.Context) {
	var req tenantModels.CreateServiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant pool from context (injected by TenantMiddleware)
	pool, exists := c.Get("tenant_pool")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant pool not found"})
		return
	}

	tenantPool := pool.(*pgxpool.Pool)

	service, err := h.serviceRepo.Create(c.Request.Context(), tenantPool, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create service", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, service)
}

// GetByID retrieves a service by ID
func (h *ServiceHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service ID"})
		return
	}

	// Get tenant pool from context
	pool, exists := c.Get("tenant_pool")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant pool not found"})
		return
	}

	tenantPool := pool.(*pgxpool.Pool)

	service, err := h.serviceRepo.GetByID(c.Request.Context(), tenantPool, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	c.JSON(http.StatusOK, service)
}

// List retrieves services with pagination and filters
func (h *ServiceHandler) List(c *gin.Context) {
	// Parse pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse filters
	var isActive *bool
	if active := c.Query("active"); active != "" {
		if active == "true" {
			val := true
			isActive = &val
		} else if active == "false" {
			val := false
			isActive = &val
		}
	}

	// Get tenant pool from context
	pool, exists := c.Get("tenant_pool")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant pool not found"})
		return
	}

	tenantPool := pool.(*pgxpool.Pool)

	result, err := h.serviceRepo.List(c.Request.Context(), tenantPool, page, pageSize, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list services", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update updates a service
func (h *ServiceHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service ID"})
		return
	}

	var req tenantModels.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant pool from context
	pool, exists := c.Get("tenant_pool")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant pool not found"})
		return
	}

	tenantPool := pool.(*pgxpool.Pool)

	service, err := h.serviceRepo.Update(c.Request.Context(), tenantPool, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update service", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, service)
}

// Delete soft deletes a service
func (h *ServiceHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service ID"})
		return
	}

	// Get tenant pool from context
	pool, exists := c.Get("tenant_pool")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant pool not found"})
		return
	}

	tenantPool := pool.(*pgxpool.Pool)

	if err := h.serviceRepo.Delete(c.Request.Context(), tenantPool, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete service", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service deleted successfully"})
}
