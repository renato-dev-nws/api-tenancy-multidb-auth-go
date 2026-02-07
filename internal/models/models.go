package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	LastTenantLogged *string   `json:"last_tenant_logged,omitempty"` // Nullable: usuário pode não ter acessado nenhum tenant ainda
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserProfile struct {
	UserID    uuid.UUID `json:"user_id"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tenant struct {
	ID           uuid.UUID    `json:"id"`
	DBCode       uuid.UUID    `json:"db_code"`
	URLCode      string       `json:"url_code"`           // Auto-generated 11-char code for admin routing (ex: FR34JJO390G)
	Subdomain    string       `json:"subdomain"`          // User-chosen subdomain for public site (ex: joao.meusaas.app)
	OwnerID      *uuid.UUID   `json:"owner_id,omitempty"` // Nullable: tenant pode ser criado sem owner pela Admin API
	PlanID       uuid.UUID    `json:"plan_id"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	Status       TenantStatus `json:"status"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type BillingCycle string

const (
	BillingCycleMonthly    BillingCycle = "monthly"
	BillingCycleQuarterly  BillingCycle = "quarterly"
	BillingCycleSemiannual BillingCycle = "semiannual"
	BillingCycleAnnual     BillingCycle = "annual"
)

type TenantStatus string

const (
	TenantStatusProvisioning TenantStatus = "provisioning"
	TenantStatusActive       TenantStatus = "active"
	TenantStatusSuspended    TenantStatus = "suspended"
)

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

type TenantMember struct {
	UserID    uuid.UUID  `json:"user_id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	RoleID    *uuid.UUID `json:"role_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Feature struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Plan struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Role struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DTOs for API requests/responses

// Admin API (Control Plane) - SaaS administrators
type AdminRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

type AdminLoginResponse struct {
	Token   string `json:"token"`
	SysUser struct {
		ID       uuid.UUID `json:"id"`
		Email    string    `json:"email"`
		FullName string    `json:"full_name"`
	} `json:"sys_user"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// Tenant API (Data Plane) - Tenant users
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
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

type UserTenant struct {
	ID        uuid.UUID `json:"id"`
	URLCode   string    `json:"url_code"`  // For admin panel: meusaas.app/adm/{url_code}/dashboard
	Subdomain string    `json:"subdomain"` // For public site: {subdomain}.meusaas.app
	Name      string    `json:"name,omitempty"`
	Role      string    `json:"role,omitempty"`
}

type TenantConfigResponse struct {
	Features    []string `json:"features"`
	Permissions []string `json:"permissions"`
}

// SubscriptionRequest representa o cadastro completo de um novo assinante
type SubscriptionRequest struct {
	PlanID       uuid.UUID    `json:"plan_id" binding:"required"`
	BillingCycle BillingCycle `json:"billing_cycle" binding:"required"`
	Name         string       `json:"name" binding:"required"`
	IsCompany    bool         `json:"is_company"`
	CompanyName  string       `json:"company_name,omitempty"`                    // Se não for empresa, usar Name
	Subdomain    string       `json:"subdomain" binding:"required,min=3,max=50"` // For public site (ex: joao.meusaas.app)
	Email        string       `json:"email" binding:"required,email"`
	Password     string       `json:"password" binding:"required,min=8"`
	CustomDomain string       `json:"custom_domain,omitempty"` // Domínio customizado opcional
}

// SubscriptionResponse retorna o token e dados do tenant criado
type SubscriptionResponse struct {
	Token  string `json:"token"`
	User   User   `json:"user"`
	Tenant Tenant `json:"tenant"`
}

// ===== Admin API - Plans Management =====

// CreatePlanRequest representa os dados para criar um plano
type CreatePlanRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" binding:"required,min=0"`
	FeatureIDs  []string `json:"feature_ids"` // UUIDs das features
}

// UpdatePlanRequest representa os dados para atualizar um plano
type UpdatePlanRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" binding:"required,min=0"`
	FeatureIDs  []string `json:"feature_ids"` // UUIDs das features
}

// PlanResponse retorna um plano com suas features
type PlanResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	Features    []Feature `json:"features"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PlanListResponse retorna lista de planos
type PlanListResponse struct {
	Plans []PlanResponse `json:"plans"`
	Total int            `json:"total"`
}
