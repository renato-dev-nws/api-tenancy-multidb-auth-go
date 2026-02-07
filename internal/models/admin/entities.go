package admin

import (
	"time"

	"github.com/google/uuid"
)

// Plan representa um plano de assinatura
type Plan struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Feature representa uma funcionalidade do sistema
type Feature struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Code        string    `json:"code"` // Código curto para permissões (prod, serv)
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Role representa uma role (pode ser de sistema ou de tenant)
type Role struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"` // Null = role de sistema
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Permission representa uma permissão
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
