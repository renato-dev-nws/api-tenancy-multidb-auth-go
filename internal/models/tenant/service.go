package tenant

import (
	"time"

	"github.com/google/uuid"
)

// Service representa um serviço no banco de dados do tenant
type Service struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     *string   `json:"description,omitempty"`
	DurationMinutes *int      `json:"duration_minutes,omitempty"`
	Price           float64   `json:"price"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateServiceRequest DTO para criação de serviço
type CreateServiceRequest struct {
	Name            string  `json:"name" binding:"required,min=3,max=255"`
	Description     *string `json:"description,omitempty"`
	DurationMinutes *int    `json:"duration_minutes,omitempty" binding:"omitempty,min=1"`
	Price           float64 `json:"price" binding:"required,min=0"`
	Active          *bool   `json:"active,omitempty"`
}

// UpdateServiceRequest DTO para atualização de serviço
type UpdateServiceRequest struct {
	Name            *string  `json:"name,omitempty" binding:"omitempty,min=3,max=255"`
	Description     *string  `json:"description,omitempty"`
	DurationMinutes *int     `json:"duration_minutes,omitempty" binding:"omitempty,min=1"`
	Price           *float64 `json:"price,omitempty" binding:"omitempty,min=0"`
	Active          *bool    `json:"active,omitempty"`
}

// ServiceListResponse retorna lista paginada de serviços
type ServiceListResponse struct {
	Services   []Service `json:"services"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}
