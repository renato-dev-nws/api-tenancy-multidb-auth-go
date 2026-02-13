package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	adminModels "github.com/saas-multi-database-api/internal/models/admin"
	adminService "github.com/saas-multi-database-api/internal/services/admin"
)

type PlanHandler struct {
	planService *adminService.PlanService
}

func NewPlanHandler(planService *adminService.PlanService) *PlanHandler {
	return &PlanHandler{
		planService: planService,
	}
}

// GetAllPlans lista todos os planos (com cache Redis)
// GET /api/v1/admin/plans
func (h *PlanHandler) GetAllPlans(c *gin.Context) {
	planResponses, err := h.planService.GetAllPlansWithCache(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plans", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, adminModels.PlanListResponse{
		Plans: planResponses,
		Total: len(planResponses),
	})
}

// GetPlanByID retorna um plano específico (com cache Redis)
// GET /api/v1/admin/plans/:id
func (h *PlanHandler) GetPlanByID(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan ID"})
		return
	}

	planResponse, err := h.planService.GetPlanByIDWithCache(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	c.JSON(http.StatusOK, adminModels.PlanResponse{
		ID:          planResponse.ID,
		Name:        planResponse.Name,
		Description: planResponse.Description,
		Price:       planResponse.Price,
		Features:    planResponse.Features,
		CreatedAt:   planResponse.CreatedAt,
		UpdatedAt:   planResponse.UpdatedAt,
	})
}

// CreatePlan cria um novo plano
// POST /api/v1/admin/plans
func (h *PlanHandler) CreatePlan(c *gin.Context) {
	var req adminModels.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Criar plano (com invalidação de cache)
	plan, err := h.planService.CreatePlan(c.Request.Context(), req.Name, req.Description, req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create plan", "details": err.Error()})
		return
	}

	// Associar features se fornecidas
	if len(req.FeatureIDs) > 0 {
		var featureUUIDs []uuid.UUID
		for _, idStr := range req.FeatureIDs {
			featureID, err := uuid.Parse(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID", "feature_id": idStr})
				return
			}
			featureUUIDs = append(featureUUIDs, featureID)
		}
		if err := h.planService.AddFeaturesToPlan(c.Request.Context(), plan.ID, featureUUIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add features to plan"})
			return
		}
	}

	// Buscar plano atualizado do cache
	planResponse, err := h.planService.GetPlanByIDWithCache(c.Request.Context(), plan.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get created plan"})
		return
	}

	c.JSON(http.StatusCreated, adminModels.PlanResponse{
		ID:          planResponse.ID,
		Name:        planResponse.Name,
		Description: planResponse.Description,
		Price:       planResponse.Price,
		Features:    planResponse.Features,
		CreatedAt:   planResponse.CreatedAt,
		UpdatedAt:   planResponse.UpdatedAt,
	})
}

// UpdatePlan atualiza um plano existente
// PUT /api/v1/admin/plans/:id
func (h *PlanHandler) UpdatePlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan ID"})
		return
	}

	var req adminModels.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Atualizar plano (com invalidação de cache)
	_, err = h.planService.UpdatePlan(c.Request.Context(), planID, req.Name, req.Description, req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update plan", "details": err.Error()})
		return
	}

	// Atualizar features (remove todas e adiciona novas)
	var featureUUIDs []uuid.UUID
	for _, idStr := range req.FeatureIDs {
		featureID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID", "feature_id": idStr})
			return
		}
		featureUUIDs = append(featureUUIDs, featureID)
	}
	if err := h.planService.AddFeaturesToPlan(c.Request.Context(), planID, featureUUIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update plan features"})
		return
	}

	// Buscar plano atualizado do cache
	planResponse, err := h.planService.GetPlanByIDWithCache(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updated plan"})
		return
	}

	c.JSON(http.StatusOK, adminModels.PlanResponse{
		ID:          planResponse.ID,
		Name:        planResponse.Name,
		Description: planResponse.Description,
		Price:       planResponse.Price,
		Features:    planResponse.Features,
		CreatedAt:   planResponse.CreatedAt,
		UpdatedAt:   planResponse.UpdatedAt,
	})
}

// DeletePlan deleta um plano
// DELETE /api/v1/admin/plans/:id
func (h *PlanHandler) DeletePlan(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan ID"})
		return
	}

	if err := h.planService.DeletePlan(c.Request.Context(), planID); err != nil {
		// Verifica se é erro de plano em uso
		if err.Error() == "cannot delete plan: tenants are using this plan" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete plan", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "plan deleted successfully"})
}
