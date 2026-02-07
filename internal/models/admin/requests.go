package admin

import (
	"time"

	"github.com/google/uuid"
	"github.com/saas-multi-database-api/internal/models/shared"
)

// ===== SysUser Requests/Responses =====

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

type CreateSysUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
	RoleIDs  []int  `json:"role_ids,omitempty"`
}

type UpdateSysUserRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	FullName  string  `json:"full_name" binding:"required"`
	AvatarURL *string `json:"avatar_url"`
	Status    string  `json:"status" binding:"required,oneof=active suspended inactive"`
	RoleIDs   []int   `json:"role_ids,omitempty"`
}

type SysUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url"`
	Status    string    `json:"status"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SysUserListResponse struct {
	Users []SysUserResponse `json:"users"`
	Total int               `json:"total"`
}

// ===== Plan Requests/Responses =====

type CreatePlanRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" binding:"required,min=0"`
	FeatureIDs  []string `json:"feature_ids"` // UUIDs das features
}

type UpdatePlanRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" binding:"required,min=0"`
	FeatureIDs  []string `json:"feature_ids"`
}

type PlanResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	Features    []Feature `json:"features"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PlanListResponse struct {
	Plans []PlanResponse `json:"plans"`
	Total int            `json:"total"`
}

// ===== Feature Requests/Responses =====

type CreateFeatureRequest struct {
	Title       string `json:"title" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Code        string `json:"code" binding:"required,min=2,max=10"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type UpdateFeatureRequest struct {
	Title       string `json:"title" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Code        string `json:"code" binding:"required,min=2,max=10"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type FeatureResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Code        string    `json:"code"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	PlanCount   int       `json:"plan_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FeatureListResponse struct {
	Features []FeatureResponse `json:"features"`
	Total    int               `json:"total"`
}

// ===== Subscription Requests/Responses =====

type SubscriptionRequest struct {
	PlanID       uuid.UUID           `json:"plan_id" binding:"required"`
	BillingCycle shared.BillingCycle `json:"billing_cycle" binding:"required"`
	Name         string              `json:"name" binding:"required"`
	IsCompany    bool                `json:"is_company"`
	CompanyName  string              `json:"company_name,omitempty"`
	Subdomain    string              `json:"subdomain" binding:"required,min=3,max=50"`
	Email        string              `json:"email" binding:"required,email"`
	Password     string              `json:"password" binding:"required,min=8"`
	CustomDomain string              `json:"custom_domain,omitempty"`
}

type SubscriptionResponse struct {
	Token  string `json:"token"`
	User   User   `json:"user"`
	Tenant Tenant `json:"tenant"`
}

// ===== Auth Requests (Generic) =====

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ===== Config Responses =====

type TenantConfigResponse struct {
	Features    []string `json:"features"`
	Permissions []string `json:"permissions"`
}
