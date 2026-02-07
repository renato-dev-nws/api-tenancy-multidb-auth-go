package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	adminModels "github.com/saas-multi-database-api/internal/models/admin"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
)

type PlanHandler struct {
	planRepo *adminRepo.PlanRepository
}

func NewPlanHandler(planRepo *adminRepo.PlanRepository) *PlanHandler {
	return &PlanHandler{
		planRepo: planRepo,
	}
}

// GetAllPlans lista todos os planos
// GET /api/v1/admin/plans
func (h *PlanHandler) GetAllPlans(c *gin.Context) {
	plans, err := h.planRepo.GetAllPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plans", "details": err.Error()})
		return
	}

	// Para cada plano, buscar suas features
	var planResponses []adminModels.PlanResponse
	for _, plan := range plans {
		features, err := h.planRepo.GetPlanFeatures(c.Request.Context(), plan.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan features"})
			return
		}

		planResponses = append(planResponses, adminModels.PlanResponse{
			ID:          plan.ID,
			Name:        plan.Name,
			Description: plan.Description,
			Price:       plan.Price,
			Features:    features,
			CreatedAt:   plan.CreatedAt,
			UpdatedAt:   plan.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, adminModels.PlanListResponse{
		Plans: planResponses,
		Total: len(planResponses),
	})
}

// GetPlanByID retorna um plano específico
// GET /api/v1/admin/plans/:id
func (h *PlanHandler) GetPlanByID(c *gin.Context) {
	planIDStr := c.Param("id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan ID"})
		return
	}

	plan, err := h.planRepo.GetPlanByID(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}

	features, err := h.planRepo.GetPlanFeatures(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan features"})
		return
	}

	c.JSON(http.StatusOK, adminModels.PlanResponse{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Features:    features,
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
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

	// Criar plano
	plan, err := h.planRepo.CreatePlan(c.Request.Context(), req.Name, req.Description, req.Price)
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

		if err := h.planRepo.SetPlanFeatures(c.Request.Context(), plan.ID, featureUUIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set plan features"})
			return
		}
	}

	// Buscar features para retorno
	features, err := h.planRepo.GetPlanFeatures(c.Request.Context(), plan.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan features"})
		return
	}

	c.JSON(http.StatusCreated, adminModels.PlanResponse{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Features:    features,
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
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

	// Atualizar plano
	plan, err := h.planRepo.UpdatePlan(c.Request.Context(), planID, req.Name, req.Description, req.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update plan", "details": err.Error()})
		return
	}

	// Atualizar features
	var featureUUIDs []uuid.UUID
	for _, idStr := range req.FeatureIDs {
		featureID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID", "feature_id": idStr})
			return
		}
		featureUUIDs = append(featureUUIDs, featureID)
	}

	if err := h.planRepo.SetPlanFeatures(c.Request.Context(), planID, featureUUIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set plan features"})
		return
	}

	// Buscar features para retorno
	features, err := h.planRepo.GetPlanFeatures(c.Request.Context(), planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get plan features"})
		return
	}

	c.JSON(http.StatusOK, adminModels.PlanResponse{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Features:    features,
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
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

	if err := h.planRepo.DeletePlan(c.Request.Context(), planID); err != nil {
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
