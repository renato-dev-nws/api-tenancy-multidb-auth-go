package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	adminModels "github.com/saas-multi-database-api/internal/models/admin"
	adminRepo "github.com/saas-multi-database-api/internal/repository/admin"
)

const (
	// Cache key for all plans with features
	plansListCacheKey = "plans:list:all"
	// Cache TTL: 24 hours (plans change sporadically)
	plansCacheTTL = 24 * time.Hour
)

type PlanService struct {
	planRepo *adminRepo.PlanRepository
	redis    *redis.Client
}

func NewPlanService(planRepo *adminRepo.PlanRepository, redisClient *redis.Client) *PlanService {
	return &PlanService{
		planRepo: planRepo,
		redis:    redisClient,
	}
}

// GetAllPlansWithCache retorna todos os planos com features, usando cache Redis
func (s *PlanService) GetAllPlansWithCache(ctx context.Context) ([]adminModels.PlanResponse, error) {
	// Tentar buscar do cache primeiro
	cached, err := s.redis.Get(ctx, plansListCacheKey).Result()
	if err == nil {
		// Cache hit - deserializar
		var plans []adminModels.PlanResponse
		if err := json.Unmarshal([]byte(cached), &plans); err == nil {
			return plans, nil
		}
		// Se falhou ao deserializar, continua para buscar do BD
	}

	// Cache miss - buscar do banco de dados
	plans, err := s.planRepo.GetAllPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get plans from database: %w", err)
	}

	// Para cada plano, buscar suas features
	var planResponses []adminModels.PlanResponse
	for _, plan := range plans {
		features, err := s.planRepo.GetPlanFeatures(ctx, plan.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get plan features: %w", err)
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

	// Cachear resultado por 24 horas
	jsonData, err := json.Marshal(planResponses)
	if err == nil {
		s.redis.Set(ctx, plansListCacheKey, jsonData, plansCacheTTL)
	}

	return planResponses, nil
}

// GetPlanByIDWithCache retorna um plano específico com features
func (s *PlanService) GetPlanByIDWithCache(ctx context.Context, planID uuid.UUID) (*adminModels.PlanResponse, error) {
	cacheKey := fmt.Sprintf("plans:id:%s", planID.String())

	// Tentar buscar do cache
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var planResponse adminModels.PlanResponse
		if err := json.Unmarshal([]byte(cached), &planResponse); err == nil {
			return &planResponse, nil
		}
	}

	// Cache miss - buscar do BD
	plan, err := s.planRepo.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	features, err := s.planRepo.GetPlanFeatures(ctx, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan features: %w", err)
	}

	planResponse := &adminModels.PlanResponse{
		ID:          plan.ID,
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Features:    features,
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
	}

	// Cachear por 24 horas
	jsonData, err := json.Marshal(planResponse)
	if err == nil {
		s.redis.Set(ctx, cacheKey, jsonData, plansCacheTTL)
	}

	return planResponse, nil
}

// InvalidatePlansCache invalida todo o cache de planos
// Deve ser chamado quando planos são criados, atualizados ou deletados
func (s *PlanService) InvalidatePlansCache(ctx context.Context) error {
	// Invalidar cache da lista completa
	if err := s.redis.Del(ctx, plansListCacheKey).Err(); err != nil {
		return fmt.Errorf("failed to invalidate plans list cache: %w", err)
	}

	// Invalidar cache de todos os planos individuais
	// Usar padrão de chave para deletar todos os caches de planos individuais
	iter := s.redis.Scan(ctx, 0, "plans:id:*", 0).Iterator()
	for iter.Next(ctx) {
		if err := s.redis.Del(ctx, iter.Val()).Err(); err != nil {
			// Log error but continue
			continue
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan plan caches: %w", err)
	}

	return nil
}

// InvalidatePlanCache invalida o cache de um plano específico
func (s *PlanService) InvalidatePlanCache(ctx context.Context, planID uuid.UUID) error {
	cacheKey := fmt.Sprintf("plans:id:%s", planID.String())

	// Invalidar cache individual
	if err := s.redis.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("failed to invalidate plan cache: %w", err)
	}

	// Invalidar cache da lista completa também
	if err := s.redis.Del(ctx, plansListCacheKey).Err(); err != nil {
		return fmt.Errorf("failed to invalidate plans list cache: %w", err)
	}

	return nil
}

// CreatePlan cria um plano e invalida cache
func (s *PlanService) CreatePlan(ctx context.Context, name, description string, price float64) (*adminModels.Plan, error) {
	plan, err := s.planRepo.CreatePlan(ctx, name, description, price)
	if err != nil {
		return nil, err
	}

	// Invalidar cache após criar
	if err := s.InvalidatePlansCache(ctx); err != nil {
		// Log error mas não falha a operação
		fmt.Printf("Warning: failed to invalidate plans cache: %v\n", err)
	}

	return plan, nil
}

// UpdatePlan atualiza um plano e invalida cache
func (s *PlanService) UpdatePlan(ctx context.Context, planID uuid.UUID, name, description string, price float64) (*adminModels.Plan, error) {
	plan, err := s.planRepo.UpdatePlan(ctx, planID, name, description, price)
	if err != nil {
		return nil, err
	}

	// Invalidar cache após atualizar
	if err := s.InvalidatePlanCache(ctx, planID); err != nil {
		fmt.Printf("Warning: failed to invalidate plan cache: %v\n", err)
	}

	return plan, nil
}

// DeletePlan deleta um plano e invalida cache
func (s *PlanService) DeletePlan(ctx context.Context, planID uuid.UUID) error {
	if err := s.planRepo.DeletePlan(ctx, planID); err != nil {
		return err
	}

	// Invalidar cache após deletar
	if err := s.InvalidatePlanCache(ctx, planID); err != nil {
		fmt.Printf("Warning: failed to invalidate plan cache: %v\n", err)
	}

	return nil
}

// AddFeaturesToPlan adiciona múltiplas features a um plano e invalida cache
func (s *PlanService) AddFeaturesToPlan(ctx context.Context, planID uuid.UUID, featureIDs []uuid.UUID) error {
	if err := s.planRepo.SetPlanFeatures(ctx, planID, featureIDs); err != nil {
		return err
	}

	// Invalidar cache após adicionar features
	if err := s.InvalidatePlanCache(ctx, planID); err != nil {
		fmt.Printf("Warning: failed to invalidate plan cache: %v\n", err)
	}

	return nil
}
