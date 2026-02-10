package tenant

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	tenantmodel "github.com/saas-multi-database-api/internal/models/tenant"
	tenantrepo "github.com/saas-multi-database-api/internal/repository/tenant"
	tenantservice "github.com/saas-multi-database-api/internal/services/tenant"
)

type ImageHandler struct {
	imageRepo     *tenantrepo.ImageRepository
	uploadService *tenantservice.UploadService
}

func NewImageHandler(imageRepo *tenantrepo.ImageRepository, uploadService *tenantservice.UploadService) *ImageHandler {
	return &ImageHandler{
		imageRepo:     imageRepo,
		uploadService: uploadService,
	}
}

// UploadImages handles multipart image upload
// POST /api/v1/adm/:url_code/images
func (h *ImageHandler) UploadImages(c *gin.Context) {
	// Get tenant UUID from context (set by middleware)
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id not found in context"})
		return
	}

	// Parse form data
	if err := c.Request.ParseMultipartForm(50 << 20); err != nil { // 50 MB max
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form"})
		return
	}

	// Get imageable_type and imageable_id from form
	imageableType := c.PostForm("imageable_type")
	imageableIDStr := c.PostForm("imageable_id")

	if imageableType == "" || imageableIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "imageable_type and imageable_id are required"})
		return
	}

	// Validate imageable_type
	validTypes := []string{"product", "service", "user", "tenant"}
	isValid := false
	for _, vt := range validTypes {
		if imageableType == vt {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid imageable_type. Must be one of: product, service, user, tenant"})
		return
	}

	// Parse imageable_id
	imageableID, err := uuid.Parse(imageableIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid imageable_id format"})
		return
	}

	// Get files from request
	form := c.Request.MultipartForm
	files := form.File["files"]
	if len(files) == 0 {
		// Try singular "file" field as well
		files = form.File["file"]
	}

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	// Get optional titles and alt_texts
	titles := form.Value["titles"]
	altTexts := form.Value["alt_texts"]

	// Setup upload options
	opts := &tenantservice.UploadOptions{
		TenantUUID:    tenantID.(string),
		ImageableType: imageableType,
		ImageableID:   imageableID,
		MaxFileSize:   10 << 20, // 10 MB per file
		MaxFiles:      10,
		AllowedTypes:  []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
	}

	// Upload images
	results, err := h.uploadService.UploadMultipleImages(c.Request.Context(), files, opts, titles, altTexts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process results
	var successfulImages []tenantmodel.Image
	var errors []map[string]interface{}

	for _, result := range results {
		if result.Error != nil {
			errors = append(errors, map[string]interface{}{
				"filename": result.Filename,
				"error":    result.Error.Error(),
			})
		} else if result.Image != nil {
			successfulImages = append(successfulImages, *result.Image)
		}
	}

	// Return response
	response := gin.H{
		"uploaded": len(successfulImages),
		"images":   successfulImages,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	status := http.StatusOK
	if len(successfulImages) == 0 {
		status = http.StatusBadRequest
	}

	c.JSON(status, response)
}

// ListImages retrieves all images for an entity
// GET /api/v1/adm/:url_code/images?imageable_type=product&imageable_id=uuid
func (h *ImageHandler) ListImages(c *gin.Context) {
	imageableType := c.Query("imageable_type")
	imageableIDStr := c.Query("imageable_id")

	if imageableType == "" || imageableIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "imageable_type and imageable_id query parameters are required"})
		return
	}

	imageableID, err := uuid.Parse(imageableIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid imageable_id format"})
		return
	}

	// Get only original images by default (unless variants=true)
	showVariants := c.Query("variants") == "true"

	var images []tenantmodel.Image
	if showVariants {
		images, err = h.imageRepo.ListByImageable(c.Request.Context(), imageableType, imageableID)
	} else {
		images, err = h.imageRepo.ListOriginalsByImageable(c.Request.Context(), imageableType, imageableID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"images": images})
}

// GetImage retrieves a single image by ID
// GET /api/v1/adm/:url_code/images/:id
func (h *ImageHandler) GetImage(c *gin.Context) {
	idStr := c.Param("id")
	imageID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image id"})
		return
	}

	image, err := h.imageRepo.GetByID(c.Request.Context(), imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	// Get variants if this is an original image
	var variants []tenantmodel.Image
	if image.Variant == tenantmodel.VariantOriginal {
		variants, _ = h.imageRepo.GetVariants(c.Request.Context(), imageID)
	}

	response := gin.H{"image": image}
	if len(variants) > 0 {
		response["variants"] = variants
	}

	c.JSON(http.StatusOK, response)
}

// UpdateImage updates image metadata (title, alt_text, display_order)
// PUT /api/v1/adm/:url_code/images/:id
func (h *ImageHandler) UpdateImage(c *gin.Context) {
	idStr := c.Param("id")
	imageID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image id"})
		return
	}

	var req tenantmodel.UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.imageRepo.Update(c.Request.Context(), imageID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update image"})
		return
	}

	// Get updated image
	image, err := h.imageRepo.GetByID(c.Request.Context(), imageID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "image updated successfully"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": image})
}

// DeleteImage deletes an image and its file
// DELETE /api/v1/adm/:url_code/images/:id
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	idStr := c.Param("id")
	imageID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image id"})
		return
	}

	if err := h.uploadService.DeleteImage(c.Request.Context(), imageID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "image deleted successfully"})
}
