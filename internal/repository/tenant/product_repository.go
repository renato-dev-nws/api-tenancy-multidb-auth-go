package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tenantModels "github.com/saas-multi-database-api/internal/models/tenant"
)

// ProductRepository handles product data access in tenant databases
type ProductRepository struct{}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

// Create creates a new product in the tenant database
func (r *ProductRepository) Create(ctx context.Context, pool *pgxpool.Pool, req *tenantModels.CreateProductRequest) (*tenantModels.Product, error) {
	query := `
		INSERT INTO products (name, description, price, sku, stock, active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, price, sku, stock, active, created_at, updated_at
	`

	stock := 0
	if req.Stock != nil {
		stock = *req.Stock
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	var product tenantModels.Product
	err := pool.QueryRow(ctx, query,
		req.Name,
		req.Description,
		req.Price,
		req.SKU,
		stock,
		active,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SKU,
		&product.Stock,
		&product.Active,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &product, nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*tenantModels.Product, error) {
	query := `
		SELECT id, name, description, price, sku, stock, active, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product tenantModels.Product
	err := pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SKU,
		&product.Stock,
		&product.Active,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// List retrieves products with pagination and optional filters
func (r *ProductRepository) List(ctx context.Context, pool *pgxpool.Pool, page, pageSize int, isActive *bool) (*tenantModels.ProductListResponse, error) {
	offset := (page - 1) * pageSize

	// Build query with filters
	query := `
		SELECT id, name, description, price, sku, stock, active, created_at, updated_at
		FROM products
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM products WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	if isActive != nil {
		query += fmt.Sprintf(" AND active = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND active = $%d", argIndex)
		args = append(args, *isActive)
		argIndex++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

	// Get total count
	var totalCount int
	err := pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// Get products
	args = append(args, pageSize, offset)
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	products := []tenantModels.Product{}
	for rows.Next() {
		var product tenantModels.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.SKU,
			&product.Stock,
			&product.Active,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating products: %w", rows.Err())
	}

	return &tenantModels.ProductListResponse{
		Products:   products,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, req *tenantModels.UpdateProductRequest) (*tenantModels.Product, error) {
	// Build dynamic update query
	query := "UPDATE products SET "
	args := []interface{}{}
	argIndex := 1
	updates := []string{}

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Price != nil {
		updates = append(updates, fmt.Sprintf("price = $%d", argIndex))
		args = append(args, *req.Price)
		argIndex++
	}
	if req.SKU != nil {
		updates = append(updates, fmt.Sprintf("sku = $%d", argIndex))
		args = append(args, *req.SKU)
		argIndex++
	}
	if req.Stock != nil {
		updates = append(updates, fmt.Sprintf("stock = $%d", argIndex))
		args = append(args, *req.Stock)
		argIndex++
	}
	if req.Active != nil {
		updates = append(updates, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *req.Active)
		argIndex++
	}

	if len(updates) == 0 {
		return r.GetByID(ctx, pool, id)
	}

	query += fmt.Sprintf("%s WHERE id = $%d", joinStrings(updates, ", "), argIndex)
	query += " RETURNING id, name, description, price, sku, stock, active, created_at, updated_at"
	args = append(args, id)

	var product tenantModels.Product
	err := pool.QueryRow(ctx, query, args...).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.SKU,
		&product.Stock,
		&product.Active,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &product, nil
}

// Delete deletes a product (soft delete by setting active to false)
func (r *ProductRepository) Delete(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	query := "UPDATE products SET active = false WHERE id = $1"
	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// HardDelete permanently deletes a product
func (r *ProductRepository) HardDelete(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	query := "DELETE FROM products WHERE id = $1"
	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
