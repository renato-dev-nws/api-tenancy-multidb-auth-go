package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/models"
)

type TenantRepository struct {
	pool *pgxpool.Pool
}

func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

// GetTenantByURLCode retrieves a tenant by URL code
func (r *TenantRepository) GetTenantByURLCode(ctx context.Context, urlCode string) (*models.Tenant, error) {
	tenant := &models.Tenant{}

	query := `
		SELECT id, db_code, url_code, owner_id, plan_id, status, created_at, updated_at
		FROM tenants
		WHERE url_code = $1
	`

	err := r.pool.QueryRow(ctx, query, urlCode).Scan(
		&tenant.ID,
		&tenant.DBCode,
		&tenant.URLCode,
		&tenant.OwnerID,
		&tenant.PlanID,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// CheckUserAccess verifies if a user has access to a tenant
func (r *TenantRepository) CheckUserAccess(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 FROM tenant_members
			WHERE user_id = $1 AND tenant_id = $2
		)
	`

	err := r.pool.QueryRow(ctx, query, userID, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user access: %w", err)
	}

	return exists, nil
}

// GetTenantFeatures retrieves all features for a tenant's plan
func (r *TenantRepository) GetTenantFeatures(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	query := `
		SELECT f.slug
		FROM features f
		JOIN plan_features pf ON f.id = pf.feature_id
		JOIN tenants t ON t.plan_id = pf.plan_id
		WHERE t.id = $1
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant features: %w", err)
	}
	defer rows.Close()

	var features []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("failed to scan feature: %w", err)
		}
		features = append(features, slug)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating features: %w", err)
	}

	return features, nil
}

// GetUserPermissions retrieves all permissions for a user in a tenant
func (r *TenantRepository) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT p.slug
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN roles r ON rp.role_id = r.id
		JOIN tenant_members tm ON tm.role_id = r.id
		WHERE tm.user_id = $1 AND tm.tenant_id = $2
	`

	rows, err := r.pool.Query(ctx, query, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, slug)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// GetUserTenants retrieves all tenants a user has access to
func (r *TenantRepository) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]models.UserTenant, error) {
	query := `
		SELECT DISTINCT 
			t.id,
			t.url_code,
			COALESCE(tp.custom_settings->>'name', '') as name,
			r.slug as role,
			t.created_at
		FROM tenants t
		JOIN tenant_members tm ON t.id = tm.tenant_id
		LEFT JOIN tenant_profiles tp ON t.id = tp.tenant_id
		LEFT JOIN roles r ON tm.role_id = r.id
		WHERE tm.user_id = $1 AND t.status = 'active'
		ORDER BY t.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}
	defer rows.Close()

	var tenants []models.UserTenant
	for rows.Next() {
		var tenant models.UserTenant
		var name, role *string
		var createdAt time.Time // Dummy variable to ignore in scan
		if err := rows.Scan(&tenant.ID, &tenant.URLCode, &name, &role, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		if name != nil {
			tenant.Name = *name
		}
		if role != nil {
			tenant.Role = *role
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tenants: %w", err)
	}

	// Return empty slice if no tenants found
	if tenants == nil {
		return []models.UserTenant{}, nil
	}

	return tenants, nil
}
