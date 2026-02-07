package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	LastTenantLogged string    `json:"last_tenant_logged,omitempty"`
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
	ID        uuid.UUID    `json:"id"`
	DBCode    uuid.UUID    `json:"db_code"`
	URLCode   string       `json:"url_code"`
	OwnerID   uuid.UUID    `json:"owner_id"`
	PlanID    uuid.UUID    `json:"plan_id"`
	Status    TenantStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type TenantStatus string

const (
	TenantStatusProvisioning TenantStatus = "provisioning"
	TenantStatusActive       TenantStatus = "active"
	TenantStatusSuspended    TenantStatus = "suspended"
)

type TenantProfile struct {
	TenantID       uuid.UUID              `json:"tenant_id"`
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
	LastTenantLogged string       `json:"last_tenant_logged,omitempty"`
}

type UserTenant struct {
	ID      uuid.UUID `json:"id"`
	URLCode string    `json:"url_code"`
	Name    string    `json:"name,omitempty"`
	Role    string    `json:"role,omitempty"`
}

type TenantConfigResponse struct {
	Features    []string `json:"features"`
	Permissions []string `json:"permissions"`
}
