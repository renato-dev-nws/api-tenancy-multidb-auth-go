package tenant

import (
	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/models/shared"
)

// ===== Auth Requests/Responses =====

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uuid.UUID `json:"id"`
		Email    string    `json:"email"`
		FullName string    `json:"full_name"`
	} `json:"user"`
	Tenants          []UserTenant `json:"tenants"`
	LastTenantLogged *string      `json:"last_tenant_logged,omitempty"`
}

type UserTenant struct {
	ID        uuid.UUID `json:"id"`
	URLCode   string    `json:"url_code"`
	Subdomain string    `json:"subdomain"`
	Name      string    `json:"name,omitempty"`
	Role      string    `json:"role,omitempty"`
}

type TenantLoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uuid.UUID `json:"id"`
		Email    string    `json:"email"`
		FullName string    `json:"full_name"`
	} `json:"user"`
	Tenants          []UserTenant `json:"tenants"`
	LastTenantLogged *string      `json:"last_tenant_logged,omitempty"`
}

// LoginToTenantResponse retorna dados completos após login em tenant específico
type LoginToTenantResponse struct {
	Message          string       `json:"message"`
	LastTenantLogged string       `json:"last_tenant_logged"`
	Features         []string     `json:"features"`
	Permissions      []string     `json:"permissions"`
	Config           TenantConfig `json:"config"`
}

// TenantConfig contém configurações de layout do tenant para o frontend
type TenantConfig struct {
	LogoURL        string                 `json:"logo_url,omitempty"`
	CompanyName    string                 `json:"company_name,omitempty"`
	CustomSettings map[string]interface{} `json:"custom_settings,omitempty"`
}

// ===== Config Response =====

type ConfigResponse struct {
	Features    []string     `json:"features"`
	Permissions []string     `json:"permissions"`
	Config      TenantConfig `json:"config"`
}

// ===== Subscription Request/Response =====

type SubscriptionRequest struct {
	PlanID       uuid.UUID           `json:"plan_id" binding:"required"`
	BillingCycle shared.BillingCycle `json:"billing_cycle" binding:"required"`
	Name         string              `json:"name" binding:"required"`
	Subdomain    string              `json:"subdomain" binding:"required,min=3,max=50"`
	FullName     string              `json:"full_name" binding:"required"`
	Email        string              `json:"email" binding:"required,email"`
	Password     string              `json:"password" binding:"required,min=8"`
	CompanyName  string              `json:"company_name,omitempty"`
	IsCompany    bool                `json:"is_company"`
	CustomDomain string              `json:"custom_domain,omitempty"`
}

type SubscriptionResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uuid.UUID `json:"id"`
		Email    string    `json:"email"`
		FullName string    `json:"full_name"`
	} `json:"user"`
	Tenant struct {
		ID        uuid.UUID `json:"id"`
		URLCode   string    `json:"url_code"`
		Subdomain string    `json:"subdomain"`
		Name      string    `json:"name"`
	} `json:"tenant"`
}
