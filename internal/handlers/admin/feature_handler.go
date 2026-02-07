package admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	adminModels "github.com/saas-multi-database-api/internal/models/admin"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
)

type FeatureHandler struct {
	featureRepo *adminRepo.FeatureRepository
}

func NewFeatureHandler(featureRepo *adminRepo.FeatureRepository) *FeatureHandler {
	return &FeatureHandler{
		featureRepo: featureRepo,
	}
}

// GetAllFeatures lista todas as features
// GET /api/v1/admin/features
func (h *FeatureHandler) GetAllFeatures(c *gin.Context) {
	features, err := h.featureRepo.GetAllFeatures(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get features", "details": err.Error()})
		return
	}

	// Para cada feature, buscar quantos planos usam
	var featureResponses []adminModels.FeatureResponse
	for _, feature := range features {
		planCount, err := h.featureRepo.GetFeaturePlanCount(c.Request.Context(), feature.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan count"})
			return
		}

		featureResponses = append(featureResponses, adminModels.FeatureResponse{
			ID:          feature.ID,
			Title:       feature.Title,
			Slug:        feature.Slug,
			Code:        feature.Code,
			Description: feature.Description,
			IsActive:    feature.IsActive,
			PlanCount:   planCount,
			CreatedAt:   feature.CreatedAt,
			UpdatedAt:   feature.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, adminModels.FeatureListResponse{
		Features: featureResponses,
		Total:    len(featureResponses),
	})
}

// GetFeatureByID retorna uma feature específica
// GET /api/v1/admin/features/:id
func (h *FeatureHandler) GetFeatureByID(c *gin.Context) {
	featureIDStr := c.Param("id")
	featureID, err := uuid.Parse(featureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID"})
		return
	}

	feature, err := h.featureRepo.GetFeatureByID(c.Request.Context(), featureID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "feature not found"})
		return
	}

	planCount, err := h.featureRepo.GetFeaturePlanCount(c.Request.Context(), featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan count"})
		return
	}

	c.JSON(http.StatusOK, adminModels.FeatureResponse{
		ID:          feature.ID,
		Title:       feature.Title,
		Slug:        feature.Slug,
		Code:        feature.Code,
		Description: feature.Description,
		IsActive:    feature.IsActive,
		PlanCount:   planCount,
		CreatedAt:   feature.CreatedAt,
		UpdatedAt:   feature.UpdatedAt,
	})
}

// CreateFeature cria uma nova feature
// POST /api/v1/admin/features
func (h *FeatureHandler) CreateFeature(c *gin.Context) {
	var req adminModels.CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Normalizar slug e code
	req.Slug = strings.ToLower(strings.ReplaceAll(req.Slug, " ", "_"))
	req.Code = strings.ToLower(req.Code)

	// Verificar se slug já existe
	slugExists, err := h.featureRepo.CheckSlugExists(c.Request.Context(), req.Slug, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check slug"})
		return
	}
	if slugExists {
		c.JSON(http.StatusConflict, gin.H{"error": "slug already exists"})
		return
	}

	// Verificar se code já existe
	codeExists, err := h.featureRepo.CheckCodeExists(c.Request.Context(), req.Code, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check code"})
		return
	}
	if codeExists {
		c.JSON(http.StatusConflict, gin.H{"error": "code already exists"})
		return
	}

	// Criar feature
	feature, err := h.featureRepo.CreateFeature(c.Request.Context(), req.Title, req.Slug, req.Code, req.Description, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create feature", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, adminModels.FeatureResponse{
		ID:          feature.ID,
		Title:       feature.Title,
		Slug:        feature.Slug,
		Code:        feature.Code,
		Description: feature.Description,
		IsActive:    feature.IsActive,
		PlanCount:   0,
		CreatedAt:   feature.CreatedAt,
		UpdatedAt:   feature.UpdatedAt,
	})
}

// UpdateFeature atualiza uma feature existente
// PUT /api/v1/admin/features/:id
func (h *FeatureHandler) UpdateFeature(c *gin.Context) {
	featureIDStr := c.Param("id")
	featureID, err := uuid.Parse(featureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID"})
		return
	}

	var req adminModels.UpdateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Normalizar slug e code
	req.Slug = strings.ToLower(strings.ReplaceAll(req.Slug, " ", "_"))
	req.Code = strings.ToLower(req.Code)

	// Verificar se slug já existe (exceto para esta feature)
	slugExists, err := h.featureRepo.CheckSlugExists(c.Request.Context(), req.Slug, &featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check slug"})
		return
	}
	if slugExists {
		c.JSON(http.StatusConflict, gin.H{"error": "slug already exists"})
		return
	}

	// Verificar se code já existe (exceto para esta feature)
	codeExists, err := h.featureRepo.CheckCodeExists(c.Request.Context(), req.Code, &featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check code"})
		return
	}
	if codeExists {
		c.JSON(http.StatusConflict, gin.H{"error": "code already exists"})
		return
	}

	// Atualizar feature
	feature, err := h.featureRepo.UpdateFeature(c.Request.Context(), featureID, req.Title, req.Slug, req.Code, req.Description, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update feature", "details": err.Error()})
		return
	}

	planCount, err := h.featureRepo.GetFeaturePlanCount(c.Request.Context(), featureID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan count"})
		return
	}

	c.JSON(http.StatusOK, adminModels.FeatureResponse{
		ID:          feature.ID,
		Title:       feature.Title,
		Slug:        feature.Slug,
		Code:        feature.Code,
		Description: feature.Description,
		IsActive:    feature.IsActive,
		PlanCount:   planCount,
		CreatedAt:   feature.CreatedAt,
		UpdatedAt:   feature.UpdatedAt,
	})
}

// DeleteFeature deleta uma feature
// DELETE /api/v1/admin/features/:id
func (h *FeatureHandler) DeleteFeature(c *gin.Context) {
	featureIDStr := c.Param("id")
	featureID, err := uuid.Parse(featureIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID"})
		return
	}

	if err := h.featureRepo.DeleteFeature(c.Request.Context(), featureID); err != nil {
		// Verifica se é erro de feature em uso
		if strings.Contains(err.Error(), "plans are using this feature") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete feature", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "feature deleted successfully"})
}
