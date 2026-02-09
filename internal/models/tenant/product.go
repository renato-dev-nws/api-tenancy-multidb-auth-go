package tenant

import (
	"time"

	"github.com/google/uuid"
)

// Product representa um produto no banco de dados do tenant
type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Price       float64   `json:"price"`
	SKU         *string   `json:"sku,omitempty"`
	Stock       int       `json:"stock"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequestDTO para criação de produto
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=3,max=255"`
	Description *string `json:"description,omitempty"`
	Price       float64 `json:"price" binding:"required,min=0"`
	SKU         *string `json:"sku,omitempty" binding:"omitempty,max=100"`
	Stock       *int    `json:"stock,omitempty" binding:"omitempty,min=0"`
	Active      *bool   `json:"active,omitempty"`
}

// UpdateProductRequestDTO para atualização de produto
type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty" binding:"omitempty,min=3,max=255"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty" binding:"omitempty,min=0"`
	SKU         *string  `json:"sku,omitempty" binding:"omitempty,max=100"`
	Stock       *int     `json:"stock,omitempty" binding:"omitempty,min=0"`
	Active      *bool    `json:"active,omitempty"`
}

// ProductListResponse retorna lista paginada de produtos
type ProductListResponse struct {
	Products   []Product `json:"products"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}
