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

// ProductHandler handles product operations for tenants
type ProductHandler struct {
	productRepo *tenantRepo.ProductRepository
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		productRepo: tenantRepo.NewProductRepository(),
	}
}

// Create creates a new product
func (h *ProductHandler) Create(c *gin.Context) {
	var req tenantModels.CreateProductRequest

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

	product, err := h.productRepo.Create(c.Request.Context(), tenantPool, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetByID retrieves a product by ID
func (h *ProductHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	// Get tenant pool from context
	pool, exists := c.Get("tenant_pool")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant pool not found"})
		return
	}

	tenantPool := pool.(*pgxpool.Pool)

	product, err := h.productRepo.GetByID(c.Request.Context(), tenantPool, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// List retrieves products with pagination and filters
func (h *ProductHandler) List(c *gin.Context) {
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

	result, err := h.productRepo.List(c.Request.Context(), tenantPool, page, pageSize, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Update updates a product
func (h *ProductHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	var req tenantModels.UpdateProductRequest
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

	product, err := h.productRepo.Update(c.Request.Context(), tenantPool, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// Delete soft deletes a product
func (h *ProductHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
		return
	}

	// Get tenant pool from context
	pool, exists := c.Get("tenant_pool")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant pool not found"})
		return
	}

	tenantPool := pool.(*pgxpool.Pool)

	if err := h.productRepo.Delete(c.Request.Context(), tenantPool, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}
