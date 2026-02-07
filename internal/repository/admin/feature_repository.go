package admin

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/models/admin"
)

type FeatureRepository struct {
	pool *pgxpool.Pool
}

func NewFeatureRepository(pool *pgxpool.Pool) *FeatureRepository {
	return &FeatureRepository{pool: pool}
}

// GetAllFeatures retorna todas as features
func (r *FeatureRepository) GetAllFeatures(ctx context.Context) ([]admin.Feature, error) {
	query := `
		SELECT id, title, slug, code, description, is_active, created_at, updated_at
		FROM features
		ORDER BY title ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query features: %w", err)
	}
	defer rows.Close()

	var features []admin.Feature
	for rows.Next() {
		var feature admin.Feature
		if err := rows.Scan(
			&feature.ID,
			&feature.Title,
			&feature.Slug,
			&feature.Code,
			&feature.Description,
			&feature.IsActive,
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

// GetFeatureByID retorna uma feature por ID
func (r *FeatureRepository) GetFeatureByID(ctx context.Context, featureID uuid.UUID) (*admin.Feature, error) {
	query := `
		SELECT id, title, slug, code, description, is_active, created_at, updated_at
		FROM features
		WHERE id = $1
	`

	var feature admin.Feature
	err := r.pool.QueryRow(ctx, query, featureID).Scan(
		&feature.ID,
		&feature.Title,
		&feature.Slug,
		&feature.Code,
		&feature.Description,
		&feature.IsActive,
		&feature.CreatedAt,
		&feature.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get feature: %w", err)
	}

	return &feature, nil
}

// CreateFeature cria uma nova feature
func (r *FeatureRepository) CreateFeature(ctx context.Context, title, slug, code, description string, isActive bool) (*admin.Feature, error) {
	query := `
		INSERT INTO features (title, slug, code, description, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, slug, code, description, is_active, created_at, updated_at
	`

	var feature admin.Feature
	err := r.pool.QueryRow(ctx, query, title, slug, code, description, isActive).Scan(
		&feature.ID,
		&feature.Title,
		&feature.Slug,
		&feature.Code,
		&feature.Description,
		&feature.IsActive,
		&feature.CreatedAt,
		&feature.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create feature: %w", err)
	}

	return &feature, nil
}

// UpdateFeature atualiza uma feature existente
func (r *FeatureRepository) UpdateFeature(ctx context.Context, featureID uuid.UUID, title, slug, code, description string, isActive bool) (*admin.Feature, error) {
	query := `
		UPDATE features
		SET title = $2, slug = $3, code = $4, description = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, slug, code, description, is_active, created_at, updated_at
	`

	var feature admin.Feature
	err := r.pool.QueryRow(ctx, query, featureID, title, slug, code, description, isActive).Scan(
		&feature.ID,
		&feature.Title,
		&feature.Slug,
		&feature.Code,
		&feature.Description,
		&feature.IsActive,
		&feature.CreatedAt,
		&feature.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update feature: %w", err)
	}

	return &feature, nil
}

// DeleteFeature deleta uma feature (verifica se está em uso)
func (r *FeatureRepository) DeleteFeature(ctx context.Context, featureID uuid.UUID) error {
	// Primeiro verifica se há planos usando esta feature
	var count int
	checkQuery := `SELECT COUNT(*) FROM plan_features WHERE feature_id = $1`
	if err := r.pool.QueryRow(ctx, checkQuery, featureID).Scan(&count); err != nil {
		return fmt.Errorf("failed to check feature usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete feature: %d plans are using this feature", count)
	}

	// Deleta a feature
	deleteQuery := `DELETE FROM features WHERE id = $1`
	result, err := r.pool.Exec(ctx, deleteQuery, featureID)
	if err != nil {
		return fmt.Errorf("failed to delete feature: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("feature not found")
	}

	return nil
}

// GetFeaturePlanCount retorna quantos planos usam a feature
func (r *FeatureRepository) GetFeaturePlanCount(ctx context.Context, featureID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM plan_features WHERE feature_id = $1`

	var count int
	if err := r.pool.QueryRow(ctx, query, featureID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get feature plan count: %w", err)
	}

	return count, nil
}

// CheckSlugExists verifica se o slug já existe (para outra feature)
func (r *FeatureRepository) CheckSlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	var query string
	var args []interface{}

	if excludeID != nil {
		query = `SELECT EXISTS(SELECT 1 FROM features WHERE slug = $1 AND id != $2)`
		args = []interface{}{slug, *excludeID}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM features WHERE slug = $1)`
		args = []interface{}{slug}
	}

	var exists bool
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}

	return exists, nil
}

// CheckCodeExists verifica se o code já existe (para outra feature)
func (r *FeatureRepository) CheckCodeExists(ctx context.Context, code string, excludeID *uuid.UUID) (bool, error) {
	var query string
	var args []interface{}

	if excludeID != nil {
		query = `SELECT EXISTS(SELECT 1 FROM features WHERE code = $1 AND id != $2)`
		args = []interface{}{code, *excludeID}
	} else {
		query = `SELECT EXISTS(SELECT 1 FROM features WHERE code = $1)`
		args = []interface{}{code}
	}

	var exists bool
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check code existence: %w", err)
	}

	return exists, nil
}
