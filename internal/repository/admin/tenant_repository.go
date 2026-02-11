package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/models/admin"
)

type TenantRepository struct {
	pool *pgxpool.Pool
}

func NewTenantRepository(pool *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

// GetTenantByURLCode retrieves a tenant by URL code
func (r *TenantRepository) GetTenantByURLCode(ctx context.Context, urlCode string) (*admin.Tenant, error) {
	tenant := &admin.Tenant{}

	query := `
		SELECT id, db_code, url_code, subdomain, owner_id, plan_id, billing_cycle, status, created_at, updated_at
		FROM tenants
		WHERE url_code = $1
	`

	err := r.pool.QueryRow(ctx, query, urlCode).Scan(
		&tenant.ID,
		&tenant.DBCode,
		&tenant.URLCode,
		&tenant.Subdomain,
		&tenant.OwnerID,
		&tenant.PlanID,
		&tenant.BillingCycle,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetTenantBySubdomain retrieves a tenant by subdomain (public site routing)
func (r *TenantRepository) GetTenantBySubdomain(ctx context.Context, subdomain string) (*admin.Tenant, error) {
	tenant := &admin.Tenant{}

	query := `
		SELECT id, db_code, url_code, subdomain, owner_id, plan_id, billing_cycle, status, created_at, updated_at
		FROM tenants
		WHERE subdomain = $1
	`

	err := r.pool.QueryRow(ctx, query, subdomain).Scan(
		&tenant.ID,
		&tenant.DBCode,
		&tenant.URLCode,
		&tenant.Subdomain,
		&tenant.OwnerID,
		&tenant.PlanID,
		&tenant.BillingCycle,
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
	// First, check if the user is the owner of the tenant
	var ownerID uuid.UUID
	err := r.pool.QueryRow(ctx, "SELECT owner_id FROM tenants WHERE id = $1", tenantID).Scan(&ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant owner: %w", err)
	}

	// If user is the owner, return all available permissions
	if userID == ownerID {
		query := `SELECT slug FROM permissions ORDER BY slug`
		rows, err := r.pool.Query(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to get all permissions for owner: %w", err)
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

	// For non-owners, get permissions based on their role
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

// GetUserRole retrieves the role slug for a user in a specific tenant
func (r *TenantRepository) GetUserRole(ctx context.Context, userID, tenantID uuid.UUID) (string, error) {
	// First, check if the user is the owner of the tenant
	var ownerID uuid.UUID
	err := r.pool.QueryRow(ctx, "SELECT owner_id FROM tenants WHERE id = $1", tenantID).Scan(&ownerID)
	if err != nil {
		return "", fmt.Errorf("failed to check tenant owner: %w", err)
	}

	// If user is the owner, return "owner" role
	if userID == ownerID {
		return "owner", nil
	}

	// For non-owners, get role from tenant_members
	query := `
		SELECT r.slug
		FROM roles r
		JOIN tenant_members tm ON tm.role_id = r.id
		WHERE tm.user_id = $1 AND tm.tenant_id = $2
	`

	var roleSlug string
	err = r.pool.QueryRow(ctx, query, userID, tenantID).Scan(&roleSlug)
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}

	return roleSlug, nil
}

// GetUserTenants retrieves all tenants a user has access to
func (r *TenantRepository) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]admin.UserTenant, error) {
	query := `
		SELECT DISTINCT 
			t.id,
			t.url_code,
			t.subdomain,
			COALESCE(tp.custom_settings->>'name', '') as name,
			CASE 
				WHEN t.owner_id = $1 THEN 'owner'
				ELSE r.slug 
			END as role,
			t.created_at
		FROM tenants t
		LEFT JOIN tenant_members tm ON t.id = tm.tenant_id AND tm.user_id = $1
		LEFT JOIN tenant_profiles tp ON t.id = tp.tenant_id
		LEFT JOIN roles r ON tm.role_id = r.id
		WHERE (tm.user_id = $1 OR t.owner_id = $1) AND t.status = 'active'
		ORDER BY t.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}
	defer rows.Close()

	var tenants []admin.UserTenant
	for rows.Next() {
		var tenant admin.UserTenant
		var name, role *string
		var createdAt time.Time // Dummy variable to ignore in scan
		if err := rows.Scan(&tenant.ID, &tenant.URLCode, &tenant.Subdomain, &name, &role, &createdAt); err != nil {
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
		return []admin.UserTenant{}, nil
	}

	return tenants, nil
}

// GetTenantProfile retorna o perfil/configurações do tenant
func (r *TenantRepository) GetTenantProfile(ctx context.Context, tenantID uuid.UUID) (*admin.TenantProfile, error) {
	query := `
		SELECT 
			tenant_id,
			company_name,
			is_company,
			COALESCE(custom_domain, '') as custom_domain,
			COALESCE(logo_url, '') as logo_url,
			COALESCE(custom_settings, '{}'::jsonb) as custom_settings,
			created_at,
			updated_at
		FROM tenant_profiles
		WHERE tenant_id = $1
	`

	var profile admin.TenantProfile
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&profile.TenantID,
		&profile.CompanyName,
		&profile.IsCompany,
		&profile.CustomDomain,
		&profile.LogoURL,
		&profile.CustomSettings,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant profile: %w", err)
	}

	return &profile, nil
}
