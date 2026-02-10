package tenant

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	tenantmodel "github.com/saas-multi-database-api/internal/models/tenant"
	tenantrepo "github.com/saas-multi-database-api/internal/repository/tenant"
	"github.com/saas-multi-database-api/internal/storage"
)

// ImageProcessor handles image processing operations
type ImageProcessor struct {
	imageRepo     *tenantrepo.ImageRepository
	storageDriver storage.StorageDriver
}

// NewImageProcessor creates a new ImageProcessor instance
func NewImageProcessor(imageRepo *tenantrepo.ImageRepository, storageDriver storage.StorageDriver) *ImageProcessor {
	return &ImageProcessor{
		imageRepo:     imageRepo,
		storageDriver: storageDriver,
	}
}

// ProcessImage processes an image: resizes to variants and converts to WebP
func (p *ImageProcessor) ProcessImage(ctx context.Context, imageID uuid.UUID) error {
	// Get original image from database
	img, err := p.imageRepo.GetByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	// Skip if not original variant or not pending
	if img.Variant != tenantmodel.VariantOriginal || img.ProcessingStatus != tenantmodel.StatusPending {
		return fmt.Errorf("image not eligible for processing: variant=%s, status=%s", img.Variant, img.ProcessingStatus)
	}

	// Update status to processing
	if err := p.imageRepo.UpdateStatus(ctx, imageID, tenantmodel.StatusProcessing); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	// Get image reader from storage
	reader, err := p.storageDriver.GetReader(ctx, img.StoragePath)
	if err != nil {
		p.imageRepo.UpdateStatus(ctx, imageID, tenantmodel.StatusFailed)
		return fmt.Errorf("failed to get image reader: %w", err)
	}
	defer reader.Close()

	// Decode original image
	srcImage, format, err := image.Decode(reader)
	if err != nil {
		p.imageRepo.UpdateStatus(ctx, imageID, tenantmodel.StatusFailed)
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := srcImage.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate file size (approximate)
	var fileSize int64
	if img.FileSize != nil {
		fileSize = *img.FileSize
	}

	// Update original image dimensions
	if err := p.imageRepo.UpdateDimensions(ctx, imageID, originalWidth, originalHeight, fileSize); err != nil {
		return fmt.Errorf("failed to update dimensions: %w", err)
	}

	// Process variants
	variants := []struct {
		variant tenantmodel.ImageVariant
		sizes   struct{ MaxWidth, MaxHeight int }
	}{
		{tenantmodel.VariantMedium, tenantmodel.ImageVariantSizes[tenantmodel.VariantMedium]},
		{tenantmodel.VariantSmall, tenantmodel.ImageVariantSizes[tenantmodel.VariantSmall]},
		{tenantmodel.VariantThumb, tenantmodel.ImageVariantSizes[tenantmodel.VariantThumb]},
	}

	for _, v := range variants {
		if err := p.createVariant(ctx, img, srcImage, v.variant, v.sizes.MaxWidth, format); err != nil {
			p.imageRepo.UpdateStatus(ctx, imageID, tenantmodel.StatusFailed)
			return fmt.Errorf("failed to create variant %s: %w", v.variant, err)
		}
	}

	// Update status to completed
	if err := p.imageRepo.UpdateStatus(ctx, imageID, tenantmodel.StatusCompleted); err != nil {
		return fmt.Errorf("failed to update status to completed: %w", err)
	}

	return nil
}

// createVariant creates a resized variant of the original image
func (p *ImageProcessor) createVariant(
	ctx context.Context,
	original *tenantmodel.Image,
	srcImage image.Image,
	variant tenantmodel.ImageVariant,
	maxSize int,
	format string,
) error {
	// Resize image maintaining aspect ratio
	resizedImage := imaging.Fit(srcImage, maxSize, maxSize, imaging.Lanczos)

	// Get new dimensions
	bounds := resizedImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Generate filename for variant
	baseFilename := filepath.Base(original.Filename)
	ext := filepath.Ext(baseFilename)
	nameWithoutExt := baseFilename[:len(baseFilename)-len(ext)]
	variantFilename := fmt.Sprintf("%s_%s%s", nameWithoutExt, variant, ext)

	// Build storage path
	storagePath := filepath.Join(
		filepath.Dir(original.StoragePath),
		variantFilename,
	)

	// Create temporary file for variant
	tmpFile, err := os.CreateTemp("", "variant-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Encode image based on original format
	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(tmpFile, resizedImage, &jpeg.Options{Quality: 90}); err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
	case "png":
		if err := png.Encode(tmpFile, resizedImage); err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
	default:
		// Default to JPEG for unknown formats
		if err := jpeg.Encode(tmpFile, resizedImage, &jpeg.Options{Quality: 90}); err != nil {
			return fmt.Errorf("failed to encode image: %w", err)
		}
	}

	// Reopen file for reading
	tmpFile.Close()
	variantFile, err := os.Open(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to reopen temp file: %w", err)
	}
	defer variantFile.Close()

	// Get file size
	fileInfo, err := variantFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Upload variant to storage
	finalPath, publicURL, err := p.storageDriver.Upload(ctx, variantFile, storagePath)
	if err != nil {
		return fmt.Errorf("failed to upload variant: %w", err)
	}

	// Create variant record in database
	variantReq := &tenantmodel.CreateImageRequest{
		ImageableType:    original.ImageableType,
		ImageableID:      original.ImageableID,
		Filename:         variantFilename,
		OriginalFilename: getStringVal(original.OriginalFilename),
		Title:            getStringVal(original.Title),
		AltText:          getStringVal(original.AltText),
		MimeType:         original.MimeType,
		Extension:        original.Extension,
		Variant:          variant,
		ParentID:         &original.ID,
		Width:            width,
		Height:           height,
		FileSize:         fileSize,
		StorageDriver:    original.StorageDriver,
		StoragePath:      finalPath,
		PublicURL:        publicURL,
	}

	_, err = p.imageRepo.Create(ctx, variantReq)
	if err != nil {
		// Cleanup uploaded file on database error
		p.storageDriver.Delete(ctx, finalPath)
		return fmt.Errorf("failed to create variant record: %w", err)
	}

	return nil
}

// TODO: WebP variant creation
// createWebPVariant creates a WebP version of the image
// Temporarily disabled - needs proper webp encoding library
/*
func (p *ImageProcessor) createWebPVariant(
	ctx context.Context,
	original *tenantmodel.Image,
	srcImage image.Image,
	variant tenantmodel.ImageVariant,
) error {
	// Get dimensions based on variant
	sizes := tenantmodel.ImageVariantSizes[variant]
	resizedImage := imaging.Fit(srcImage, sizes.MaxWidth, sizes.MaxHeight, imaging.Lanczos)

	// Get new dimensions
	bounds := resizedImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Generate WebP filename
	baseFilename := filepath.Base(original.Filename)
	ext := filepath.Ext(baseFilename)
	nameWithoutExt := baseFilename[:len(baseFilename)-len(ext)]
	webpFilename := fmt.Sprintf("%s_%s.webp", nameWithoutExt, variant)

	// Build storage path
	storagePath := filepath.Join(
		filepath.Dir(original.StoragePath),
		webpFilename,
	)

	// Create temporary file for WebP
	tmpFile, err := os.CreateTemp("", "webp-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Encode to WebP (using quality 90)
	options := &webp.Options{Lossless: false, Quality: 90}
	if err := webp.Encode(tmpFile, resizedImage, options); err != nil {
		return fmt.Errorf("failed to encode WebP: %w", err)
	}

	// Reopen file for reading
	tmpFile.Close()
	webpFile, err := os.Open(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("failed to reopen temp file: %w", err)
	}
	defer webpFile.Close()

	// Get file size
	fileInfo, err := webpFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Upload WebP variant to storage
	finalPath, publicURL, err := p.storageDriver.Upload(ctx, webpFile, storagePath)
	if err != nil {
		return fmt.Errorf("failed to upload WebP variant: %w", err)
	}

	// Create WebP variant record in database
	webpReq := &tenantmodel.CreateImageRequest{
		ImageableType:    original.ImageableType,
		ImageableID:      original.ImageableID,
		Filename:         webpFilename,
		OriginalFilename: getStringVal(original.OriginalFilename),
		Title:            getStringVal(original.Title),
		AltText:          getStringVal(original.AltText),
		MimeType:         "image/webp",
		Extension:        "webp",
		Variant:          variant,
		ParentID:         &original.ID,
		Width:            width,
		Height:           height,
		FileSize:         fileSize,
		StorageDriver:    original.StorageDriver,
		StoragePath:      finalPath,
		PublicURL:        publicURL,
	}

	_, err = p.imageRepo.Create(ctx, webpReq)
	if err != nil {
		// Cleanup uploaded file on database error
		p.storageDriver.Delete(ctx, finalPath)
		return fmt.Errorf("failed to create WebP variant record: %w", err)
	}

	return nil
}
*/

// getStringVal returns the value of a string pointer or empty string
func getStringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
