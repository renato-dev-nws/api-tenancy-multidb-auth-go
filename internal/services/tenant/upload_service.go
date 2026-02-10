package tenant

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	tenantmodel "github.com/saas-multi-database-api/internal/models/tenant"
	tenantrepo "github.com/saas-multi-database-api/internal/repository/tenant"
	"github.com/saas-multi-database-api/internal/storage"
)

type UploadService struct {
	imageRepo *tenantrepo.ImageRepository
	storage   storage.StorageDriver
}

func NewUploadService(imageRepo *tenantrepo.ImageRepository, storageDriver storage.StorageDriver) *UploadService {
	return &UploadService{
		imageRepo: imageRepo,
		storage:   storageDriver,
	}
}

// UploadOptions contains options for image upload
type UploadOptions struct {
	TenantUUID    string
	ImageableType string
	ImageableID   uuid.UUID
	MaxFileSize   int64 // bytes
	MaxFiles      int
	AllowedTypes  []string
}

// UploadResult contains the result of an upload operation
type UploadResult struct {
	Image    *tenantmodel.Image
	Error    error
	Filename string
}

// ValidateFile validates a file before upload
func (s *UploadService) ValidateFile(file *multipart.FileHeader, opts *UploadOptions) error {
	// Check file size
	if opts.MaxFileSize > 0 && file.Size > opts.MaxFileSize {
		return fmt.Errorf("file %s exceeds maximum size of %d bytes", file.Filename, opts.MaxFileSize)
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if len(opts.AllowedTypes) > 0 {
		allowed := false
		for _, allowedExt := range opts.AllowedTypes {
			if ext == allowedExt || ext == "."+allowedExt {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file type %s not allowed", ext)
		}
	}

	return nil
}

// UploadImage uploads a single image file
func (s *UploadService) UploadImage(ctx context.Context, file *multipart.FileHeader, opts *UploadOptions, title, altText string) (*tenantmodel.Image, error) {
	// Validate file
	if err := s.ValidateFile(file, opts); err != nil {
		return nil, err
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	imageID := uuid.New()
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%s_original%s", imageID.String(), ext)

	// Build storage path: {tenant_uuid}/images/{imageable_type}/{imageable_id}/{filename}
	storagePath := fmt.Sprintf("%s/images/%s/%s/%s",
		opts.TenantUUID,
		opts.ImageableType,
		opts.ImageableID.String(),
		filename,
	)

	// Upload to storage
	finalPath, publicURL, err := s.storage.Upload(ctx, src, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Determine MIME type from extension
	mimeType := getMimeType(ext)

	// Create image record in database
	createReq := &tenantmodel.CreateImageRequest{
		ImageableType:    opts.ImageableType,
		ImageableID:      opts.ImageableID,
		Filename:         filename,
		OriginalFilename: file.Filename,
		Title:            title,
		AltText:          altText,
		MimeType:         mimeType,
		Extension:        strings.TrimPrefix(ext, "."),
		Variant:          tenantmodel.VariantOriginal,
		FileSize:         file.Size,
		StorageDriver:    "local", // TODO: Get from storage driver
		StoragePath:      finalPath,
		PublicURL:        publicURL,
	}

	image, err := s.imageRepo.Create(ctx, createReq)
	if err != nil {
		// Cleanup: delete uploaded file if database insert fails
		_ = s.storage.Delete(ctx, finalPath)
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	return image, nil
}

// UploadMultipleImages uploads multiple image files
func (s *UploadService) UploadMultipleImages(ctx context.Context, files []*multipart.FileHeader, opts *UploadOptions, titles, altTexts []string) ([]UploadResult, error) {
	// Check max files limit
	if opts.MaxFiles > 0 && len(files) > opts.MaxFiles {
		return nil, fmt.Errorf("too many files: maximum %d allowed", opts.MaxFiles)
	}

	results := make([]UploadResult, len(files))

	for i, file := range files {
		title := ""
		altText := ""

		// Get title and alt text if provided
		if i < len(titles) {
			title = titles[i]
		}
		if i < len(altTexts) {
			altText = altTexts[i]
		}

		// Upload each file
		image, err := s.UploadImage(ctx, file, opts, title, altText)
		results[i] = UploadResult{
			Image:    image,
			Error:    err,
			Filename: file.Filename,
		}
	}

	return results, nil
}

// DeleteImage deletes an image and its file from storage
func (s *UploadService) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	// Get image record
	image, err := s.imageRepo.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	// Delete file from storage
	if err := s.storage.Delete(ctx, image.StoragePath); err != nil {
		// Log error but continue with database deletion
		fmt.Printf("WARNING: failed to delete file from storage: %v\n", err)
	}

	// Delete variants if this is an original image
	if image.Variant == tenantmodel.VariantOriginal {
		variants, err := s.imageRepo.GetVariants(ctx, imageID)
		if err == nil {
			for _, variant := range variants {
				_ = s.storage.Delete(ctx, variant.StoragePath)
			}
		}
	}

	// Delete from database (cascade will delete variants)
	if err := s.imageRepo.Delete(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete image record: %w", err)
	}

	return nil
}

// getMimeType returns MIME type based on file extension
func getMimeType(ext string) string {
	ext = strings.ToLower(ext)
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".bmp":  "image/bmp",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
