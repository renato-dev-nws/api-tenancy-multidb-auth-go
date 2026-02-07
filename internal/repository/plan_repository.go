package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/models"
)

type PlanRepository struct {
	pool *pgxpool.Pool
}

func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{pool: pool}
}

// GetAllPlans retorna todos os planos
func (r *PlanRepository) GetAllPlans(ctx context.Context) ([]models.Plan, error) {
	query := `
		SELECT id, name, description, price, created_at, updated_at
		FROM plans
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}
	defer rows.Close()

	var plans []models.Plan
	for rows.Next() {
		var plan models.Plan
		if err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.Description,
			&plan.Price,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}
		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plans: %w", err)
	}

	return plans, nil
}

// GetPlanByID retorna um plano por ID
func (r *PlanRepository) GetPlanByID(ctx context.Context, planID uuid.UUID) (*models.Plan, error) {
	query := `
		SELECT id, name, description, price, created_at, updated_at
		FROM plans
		WHERE id = $1
	`

	var plan models.Plan
	err := r.pool.QueryRow(ctx, query, planID).Scan(
		&plan.ID,
		&plan.Name,
		&plan.Description,
		&plan.Price,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	return &plan, nil
}

// CreatePlan cria um novo plano
func (r *PlanRepository) CreatePlan(ctx context.Context, name, description string, price float64) (*models.Plan, error) {
	query := `
		INSERT INTO plans (name, description, price)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, price, created_at, updated_at
	`

	var plan models.Plan
	err := r.pool.QueryRow(ctx, query, name, description, price).Scan(
		&plan.ID,
		&plan.Name,
		&plan.Description,
		&plan.Price,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	return &plan, nil
}

// UpdatePlan atualiza um plano existente
func (r *PlanRepository) UpdatePlan(ctx context.Context, planID uuid.UUID, name, description string, price float64) (*models.Plan, error) {
	query := `
		UPDATE plans
		SET name = $2, description = $3, price = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, price, created_at, updated_at
	`

	var plan models.Plan
	err := r.pool.QueryRow(ctx, query, planID, name, description, price).Scan(
		&plan.ID,
		&plan.Name,
		&plan.Description,
		&plan.Price,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	return &plan, nil
}

// DeletePlan deleta um plano (verifica se está em uso)
func (r *PlanRepository) DeletePlan(ctx context.Context, planID uuid.UUID) error {
	// Primeiro verifica se há tenants usando este plano
	var count int
	checkQuery := `SELECT COUNT(*) FROM tenants WHERE plan_id = $1`
	if err := r.pool.QueryRow(ctx, checkQuery, planID).Scan(&count); err != nil {
		return fmt.Errorf("failed to check plan usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete plan: %d tenants are using this plan", count)
	}

	// Deleta as associações com features
	deleteAssocQuery := `DELETE FROM plan_features WHERE plan_id = $1`
	if _, err := r.pool.Exec(ctx, deleteAssocQuery, planID); err != nil {
		return fmt.Errorf("failed to delete plan features: %w", err)
	}

	// Deleta o plano
	deleteQuery := `DELETE FROM plans WHERE id = $1`
	result, err := r.pool.Exec(ctx, deleteQuery, planID)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("plan not found")
	}

	return nil
}

// GetPlanFeatures retorna as features de um plano
func (r *PlanRepository) GetPlanFeatures(ctx context.Context, planID uuid.UUID) ([]models.Feature, error) {
	query := `
		SELECT f.id, f.title, f.slug, f.description, f.created_at, f.updated_at
		FROM features f
		JOIN plan_features pf ON f.id = pf.feature_id
		WHERE pf.plan_id = $1
		ORDER BY f.title ASC
	`

	rows, err := r.pool.Query(ctx, query, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to query plan features: %w", err)
	}
	defer rows.Close()

	var features []models.Feature
	for rows.Next() {
		var feature models.Feature
		if err := rows.Scan(
			&feature.ID,
			&feature.Title,
			&feature.Slug,
			&feature.Description,
			&feature.CreatedAt,
			&feature.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan feature: %w", err)
		}
		features = append(features, feature)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating features: %w", err)
	}

	return features, nil
}

// SetPlanFeatures associa features a um plano (remove antigas e adiciona novas)
func (r *PlanRepository) SetPlanFeatures(ctx context.Context, planID uuid.UUID, featureIDs []uuid.UUID) error {
	// Inicia transação
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Remove associações antigas
	deleteQuery := `DELETE FROM plan_features WHERE plan_id = $1`
	if _, err := tx.Exec(ctx, deleteQuery, planID); err != nil {
		return fmt.Errorf("failed to delete old plan features: %w", err)
	}

	// Adiciona novas associações
	if len(featureIDs) > 0 {
		insertQuery := `INSERT INTO plan_features (plan_id, feature_id) VALUES ($1, $2)`
		for _, featureID := range featureIDs {
			if _, err := tx.Exec(ctx, insertQuery, planID, featureID); err != nil {
				return fmt.Errorf("failed to insert plan feature: %w", err)
			}
		}
	}

	// Commit
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
