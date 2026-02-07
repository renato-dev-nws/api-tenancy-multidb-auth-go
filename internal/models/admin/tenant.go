package admin

import (
	"time"

	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/models/shared"
)

// Tenant representa um tenant no Master DB (Control Plane)
type Tenant struct {
	ID           uuid.UUID           `json:"id"`
	DBCode       uuid.UUID           `json:"db_code"`
	URLCode      string              `json:"url_code"`           // Auto-generated 11-char code for admin routing
	Subdomain    string              `json:"subdomain"`          // User-chosen subdomain for public site
	OwnerID      *uuid.UUID          `json:"owner_id,omitempty"` // Nullable: pode ser criado sem owner pela Admin API
	PlanID       uuid.UUID           `json:"plan_id"`
	BillingCycle shared.BillingCycle `json:"billing_cycle"`
	Status       shared.TenantStatus `json:"status"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// TenantProfile contém dados adicionais do tenant
type TenantProfile struct {
	TenantID       uuid.UUID              `json:"tenant_id"`
	CompanyName    string                 `json:"company_name,omitempty"`
	IsCompany      bool                   `json:"is_company"`
	CustomDomain   string                 `json:"custom_domain,omitempty"`
	LogoURL        string                 `json:"logo_url,omitempty"`
	CustomSettings map[string]interface{} `json:"custom_settings,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// TenantMember representa a associação entre usuário e tenant
type TenantMember struct {
	UserID    uuid.UUID  `json:"user_id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	RoleID    *uuid.UUID `json:"role_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
