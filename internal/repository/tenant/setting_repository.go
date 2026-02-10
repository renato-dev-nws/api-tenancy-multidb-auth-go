package tenant

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/models/tenant"
)

type SettingRepository struct{}

func NewSettingRepository() *SettingRepository {
	return &SettingRepository{}
}

// GetByKey retrieves a setting by its key
func (r *SettingRepository) GetByKey(ctx context.Context, pool *pgxpool.Pool, key string) (*tenant.Setting, error) {
	query := `
		SELECT key, value, updated_at
		FROM settings
		WHERE key = $1
	`

	var setting tenant.Setting
	err := pool.QueryRow(ctx, query, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	return &setting, nil
}

// List retrieves all settings
func (r *SettingRepository) List(ctx context.Context, pool *pgxpool.Pool) ([]tenant.Setting, error) {
	query := `
		SELECT key, value, updated_at
		FROM settings
		ORDER BY key
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}
	defer rows.Close()

	var settings []tenant.Setting
	for rows.Next() {
		var setting tenant.Setting
		if err := rows.Scan(
			&setting.Key,
			&setting.Value,
			&setting.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating settings: %w", err)
	}

	return settings, nil
}

// Update updates a setting's value
func (r *SettingRepository) Update(ctx context.Context, pool *pgxpool.Pool, key string, value []byte) (*tenant.Setting, error) {
	query := `
		UPDATE settings
		SET value = $2, updated_at = CURRENT_TIMESTAMP
		WHERE key = $1
		RETURNING key, value, updated_at
	`

	var setting tenant.Setting
	err := pool.QueryRow(ctx, query, key, value).Scan(
		&setting.Key,
		&setting.Value,
		&setting.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update setting: %w", err)
	}

	return &setting, nil
}

// Upsert creates or updates a setting
func (r *SettingRepository) Upsert(ctx context.Context, pool *pgxpool.Pool, key string, value []byte) (*tenant.Setting, error) {
	query := `
		INSERT INTO settings (key, value, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (key)
		DO UPDATE SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
		RETURNING key, value, updated_at
	`

	var setting tenant.Setting
	err := pool.QueryRow(ctx, query, key, value).Scan(
		&setting.Key,
		&setting.Value,
		&setting.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert setting: %w", err)
	}

	return &setting, nil
}

// Delete removes a setting by key
func (r *SettingRepository) Delete(ctx context.Context, pool *pgxpool.Pool, key string) error {
	query := `DELETE FROM settings WHERE key = $1`

	result, err := pool.Exec(ctx, query, key)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("setting not found")
	}

	return nil
}
