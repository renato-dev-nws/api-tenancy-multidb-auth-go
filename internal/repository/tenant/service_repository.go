package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	tenantModels "github.com/saas-multi-database-api/internal/models/tenant"
)

// ServiceRepository handles service data access in tenant databases
type ServiceRepository struct{}

func NewServiceRepository() *ServiceRepository {
	return &ServiceRepository{}
}

// Create creates a new service in the tenant database
func (r *ServiceRepository) Create(ctx context.Context, pool *pgxpool.Pool, req *tenantModels.CreateServiceRequest) (*tenantModels.Service, error) {
	query := `
		INSERT INTO services (name, description, duration_minutes, price, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, duration_minutes, price, active, created_at, updated_at
	`

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	var service tenantModels.Service
	err := pool.QueryRow(ctx, query,
		req.Name,
		req.Description,
		req.DurationMinutes,
		req.Price,
		active,
	).Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.DurationMinutes,
		&service.Price,
		&service.Active,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return &service, nil
}

// GetByID retrieves a service by ID
func (r *ServiceRepository) GetByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*tenantModels.Service, error) {
	query := `
		SELECT id, name, description, duration_minutes, price, active, created_at, updated_at
		FROM services
		WHERE id = $1
	`

	var service tenantModels.Service
	err := pool.QueryRow(ctx, query, id).Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.DurationMinutes,
		&service.Price,
		&service.Active,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &service, nil
}

// List retrieves services with pagination and optional filters
func (r *ServiceRepository) List(ctx context.Context, pool *pgxpool.Pool, page, pageSize int, isActive *bool) (*tenantModels.ServiceListResponse, error) {
	offset := (page - 1) * pageSize

	// Build query with filters
	query := `
		SELECT id, name, description, duration_minutes, price, active, created_at, updated_at
		FROM services
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM services WHERE 1=1"
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
		return nil, fmt.Errorf("failed to count services: %w", err)
	}

	// Get services
	args = append(args, pageSize, offset)
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	services := []tenantModels.Service{}
	for rows.Next() {
		var service tenantModels.Service
		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.Description,
			&service.DurationMinutes,
			&service.Price,
			&service.Active,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, service)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating services: %w", rows.Err())
	}

	return &tenantModels.ServiceListResponse{
		Services:   services,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// Update updates a service
func (r *ServiceRepository) Update(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, req *tenantModels.UpdateServiceRequest) (*tenantModels.Service, error) {
	// Build dynamic update query
	query := "UPDATE services SET "
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
	if req.DurationMinutes != nil {
		updates = append(updates, fmt.Sprintf("duration_minutes = $%d", argIndex))
		args = append(args, *req.DurationMinutes)
		argIndex++
	}
	if req.Price != nil {
		updates = append(updates, fmt.Sprintf("price = $%d", argIndex))
		args = append(args, *req.Price)
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
	query += " RETURNING id, name, description, duration_minutes, price, active, created_at, updated_at"
	args = append(args, id)

	var service tenantModels.Service
	err := pool.QueryRow(ctx, query, args...).Scan(
		&service.ID,
		&service.Name,
		&service.Description,
		&service.DurationMinutes,
		&service.Price,
		&service.Active,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	return &service, nil
}

// Delete deletes a service (soft delete by setting active to false)
func (r *ServiceRepository) Delete(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	query := "UPDATE services SET active = false WHERE id = $1"
	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("service not found")
	}

	return nil
}

// HardDelete permanently deletes a service
func (r *ServiceRepository) HardDelete(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	query := "DELETE FROM services WHERE id = $1"
	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete service: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("service not found")
	}

	return nil
}
