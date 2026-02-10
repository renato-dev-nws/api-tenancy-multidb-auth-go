package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saas-multi-database-api/internal/models/tenant"
)

type ImageRepository struct {
	pool *pgxpool.Pool
}

func NewImageRepository(pool *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{pool: pool}
}

// Create inserts a new image record
func (r *ImageRepository) Create(ctx context.Context, req *tenant.CreateImageRequest) (*tenant.Image, error) {
	query := `
		INSERT INTO images (
			imageable_type, imageable_id, filename, original_filename, title, alt_text,
			media_type, mime_type, extension, variant, parent_id,
			width, height, file_size, storage_driver, storage_path, public_url,
			processing_status, display_order, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id, created_at, updated_at
	`

	// Set defaults
	variant := req.Variant
	if variant == "" {
		variant = tenant.VariantOriginal
	}

	storageDriver := req.StorageDriver
	if storageDriver == "" {
		storageDriver = "local"
	}

	// Nullable fields
	var originalFilename, title, altText, publicURL *string
	if req.OriginalFilename != "" {
		originalFilename = &req.OriginalFilename
	}
	if req.Title != "" {
		title = &req.Title
	}
	if req.AltText != "" {
		altText = &req.AltText
	}
	if req.PublicURL != "" {
		publicURL = &req.PublicURL
	}

	var width, height *int
	var fileSize *int64
	if req.Width > 0 {
		width = &req.Width
	}
	if req.Height > 0 {
		height = &req.Height
	}
	if req.FileSize > 0 {
		fileSize = &req.FileSize
	}

	image := &tenant.Image{
		ImageableType:    req.ImageableType,
		ImageableID:      req.ImageableID,
		Filename:         req.Filename,
		OriginalFilename: originalFilename,
		Title:            title,
		AltText:          altText,
		MediaType:        tenant.MediaTypeImage,
		MimeType:         req.MimeType,
		Extension:        req.Extension,
		Variant:          variant,
		ParentID:         req.ParentID,
		Width:            width,
		Height:           height,
		FileSize:         fileSize,
		StorageDriver:    storageDriver,
		StoragePath:      req.StoragePath,
		PublicURL:        publicURL,
		ProcessingStatus: tenant.StatusPending,
		DisplayOrder:     req.DisplayOrder,
	}

	err := r.pool.QueryRow(ctx, query,
		image.ImageableType, image.ImageableID, image.Filename, image.OriginalFilename,
		image.Title, image.AltText, image.MediaType, image.MimeType, image.Extension,
		image.Variant, image.ParentID, image.Width, image.Height, image.FileSize,
		image.StorageDriver, image.StoragePath, image.PublicURL, image.ProcessingStatus,
		image.DisplayOrder,
	).Scan(&image.ID, &image.CreatedAt, &image.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}

	return image, nil
}

// GetByID retrieves an image by ID
func (r *ImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*tenant.Image, error) {
	query := `
		SELECT id, imageable_type, imageable_id, filename, original_filename, title, alt_text,
			media_type, mime_type, extension, variant, parent_id, width, height, file_size,
			storage_driver, storage_path, public_url, processing_status, processed_at,
			display_order, created_at, updated_at
		FROM images
		WHERE id = $1
	`

	var image tenant.Image
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&image.ID, &image.ImageableType, &image.ImageableID, &image.Filename,
		&image.OriginalFilename, &image.Title, &image.AltText, &image.MediaType,
		&image.MimeType, &image.Extension, &image.Variant, &image.ParentID,
		&image.Width, &image.Height, &image.FileSize, &image.StorageDriver,
		&image.StoragePath, &image.PublicURL, &image.ProcessingStatus,
		&image.ProcessedAt, &image.DisplayOrder, &image.CreatedAt, &image.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	return &image, nil
}

// ListByImageable retrieves all images for a specific entity
func (r *ImageRepository) ListByImageable(ctx context.Context, imageableType string, imageableID uuid.UUID) ([]tenant.Image, error) {
	query := `
		SELECT id, imageable_type, imageable_id, filename, original_filename, title, alt_text,
			media_type, mime_type, extension, variant, parent_id, width, height, file_size,
			storage_driver, storage_path, public_url, processing_status, processed_at,
			display_order, created_at, updated_at
		FROM images
		WHERE imageable_type = $1 AND imageable_id = $2
		ORDER BY display_order ASC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, imageableType, imageableID)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	defer rows.Close()

	var images []tenant.Image
	for rows.Next() {
		var image tenant.Image
		err := rows.Scan(
			&image.ID, &image.ImageableType, &image.ImageableID, &image.Filename,
			&image.OriginalFilename, &image.Title, &image.AltText, &image.MediaType,
			&image.MimeType, &image.Extension, &image.Variant, &image.ParentID,
			&image.Width, &image.Height, &image.FileSize, &image.StorageDriver,
			&image.StoragePath, &image.PublicURL, &image.ProcessingStatus,
			&image.ProcessedAt, &image.DisplayOrder, &image.CreatedAt, &image.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, image)
	}

	return images, nil
}

// ListOriginalsByImageable retrieves only original images (no variants)
func (r *ImageRepository) ListOriginalsByImageable(ctx context.Context, imageableType string, imageableID uuid.UUID) ([]tenant.Image, error) {
	query := `
		SELECT id, imageable_type, imageable_id, filename, original_filename, title, alt_text,
			media_type, mime_type, extension, variant, parent_id, width, height, file_size,
			storage_driver, storage_path, public_url, processing_status, processed_at,
			display_order, created_at, updated_at
		FROM images
		WHERE imageable_type = $1 AND imageable_id = $2 AND variant = 'original'
		ORDER BY display_order ASC, created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, imageableType, imageableID)
	if err != nil {
		return nil, fmt.Errorf("failed to list original images: %w", err)
	}
	defer rows.Close()

	var images []tenant.Image
	for rows.Next() {
		var image tenant.Image
		err := rows.Scan(
			&image.ID, &image.ImageableType, &image.ImageableID, &image.Filename,
			&image.OriginalFilename, &image.Title, &image.AltText, &image.MediaType,
			&image.MimeType, &image.Extension, &image.Variant, &image.ParentID,
			&image.Width, &image.Height, &image.FileSize, &image.StorageDriver,
			&image.StoragePath, &image.PublicURL, &image.ProcessingStatus,
			&image.ProcessedAt, &image.DisplayOrder, &image.CreatedAt, &image.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, image)
	}

	return images, nil
}

// Update updates image metadata
func (r *ImageRepository) Update(ctx context.Context, id uuid.UUID, req *tenant.UpdateImageRequest) error {
	query := `
		UPDATE images
		SET title = COALESCE($2, title),
			alt_text = COALESCE($3, alt_text),
			display_order = COALESCE($4, display_order),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, req.Title, req.AltText, req.DisplayOrder)
	if err != nil {
		return fmt.Errorf("failed to update image: %w", err)
	}

	return nil
}

// UpdateStatus updates the processing status of an image
func (r *ImageRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status tenant.ProcessingStatus) error {
	query := `
		UPDATE images
		SET processing_status = $2,
			processed_at = CASE WHEN $2 IN ('completed', 'failed') THEN CURRENT_TIMESTAMP ELSE processed_at END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update image status: %w", err)
	}

	return nil
}

// Delete soft deletes an image (actually hard delete for now, can be changed to soft delete later)
func (r *ImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM images WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// DeleteByImageable deletes all images associated with an entity
func (r *ImageRepository) DeleteByImageable(ctx context.Context, imageableType string, imageableID uuid.UUID) error {
	query := `DELETE FROM images WHERE imageable_type = $1 AND imageable_id = $2`

	_, err := r.pool.Exec(ctx, query, imageableType, imageableID)
	if err != nil {
		return fmt.Errorf("failed to delete images: %w", err)
	}

	return nil
}

// GetVariants retrieves all variants of an original image
func (r *ImageRepository) GetVariants(ctx context.Context, parentID uuid.UUID) ([]tenant.Image, error) {
	query := `
		SELECT id, imageable_type, imageable_id, filename, original_filename, title, alt_text,
			media_type, mime_type, extension, variant, parent_id, width, height, file_size,
			storage_driver, storage_path, public_url, processing_status, processed_at,
			display_order, created_at, updated_at
		FROM images
		WHERE parent_id = $1
		ORDER BY 
			CASE variant
				WHEN 'medium' THEN 1
				WHEN 'small' THEN 2
				WHEN 'thumb' THEN 3
				ELSE 4
			END
	`

	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image variants: %w", err)
	}
	defer rows.Close()

	var images []tenant.Image
	for rows.Next() {
		var image tenant.Image
		err := rows.Scan(
			&image.ID, &image.ImageableType, &image.ImageableID, &image.Filename,
			&image.OriginalFilename, &image.Title, &image.AltText, &image.MediaType,
			&image.MimeType, &image.Extension, &image.Variant, &image.ParentID,
			&image.Width, &image.Height, &image.FileSize, &image.StorageDriver,
			&image.StoragePath, &image.PublicURL, &image.ProcessingStatus,
			&image.ProcessedAt, &image.DisplayOrder, &image.CreatedAt, &image.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image variant: %w", err)
		}
		images = append(images, image)
	}

	return images, nil
}

// UpdateDimensions updates width and height of an image
func (r *ImageRepository) UpdateDimensions(ctx context.Context, id uuid.UUID, width, height int, fileSize int64) error {
	query := `
		UPDATE images
		SET width = $2,
			height = $3,
			file_size = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id, width, height, fileSize)
	if err != nil {
		return fmt.Errorf("failed to update image dimensions: %w", err)
	}

	return nil
}
