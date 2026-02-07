package models

import (
	"time"

	"github.com/google/uuid"
)

// SysUser representa um administrador do SaaS (Control Plane)
type SysUser struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Nunca expor
	FullName     string    `json:"full_name"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	Status       string    `json:"status"` // active, suspended, inactive
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SysRole representa uma role de sistema
type SysRole struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SysPermission representa uma permissão de sistema
type SysPermission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SysUserWithRoles representa um sys_user com suas roles
type SysUserWithRoles struct {
	SysUser
	Roles       []SysRole       `json:"roles"`
	Permissions []SysPermission `json:"permissions"`
}
