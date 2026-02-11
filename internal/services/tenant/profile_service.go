package tenant

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	tenantmodel "github.com/saas-multi-database-api/internal/models/tenant"
	tenantrepo "github.com/saas-multi-database-api/internal/repository/tenant"
	"github.com/saas-multi-database-api/internal/storage"
)

type ProfileService struct {
	imageRepo     *tenantrepo.ImageRepository
	storageDriver storage.StorageDriver
}

func NewProfileService(imageRepo *tenantrepo.ImageRepository, storageDriver storage.StorageDriver) *ProfileService {
	return &ProfileService{
		imageRepo:     imageRepo,
		storageDriver: storageDriver,
	}
}

// UploadUserAvatar uploads and processes user avatar (200x200px max)
func (s *ProfileService) UploadUserAvatar(ctx context.Context, file *multipart.FileHeader, tenantUUID, tenantDBCode string, userID uuid.UUID) (*tenantmodel.Image, error) {
	// Validate file
	if err := s.validateAvatarFile(file); err != nil {
		return nil, err
	}

	// Process the image
	processedFile, err := s.processAvatarImage(file)
	if err != nil {
		return nil, fmt.Errorf("failed to process avatar: %w", err)
	}
	defer processedFile.Close()

	// Generate storage path for user avatar
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("avatar%s", ext)
	storagePath := fmt.Sprintf("%s/images/profiles/users/%s/%s", tenantUUID, userID.String(), filename)

	// Upload to storage
	finalPath, publicURL, err := s.storageDriver.Upload(ctx, processedFile, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Create image record in database
	createReq := &tenantmodel.CreateImageRequest{
		ImageableType:    "user",
		ImageableID:      userID,
		Filename:         filename,
		OriginalFilename: file.Filename,
		Title:            "User Avatar",
		AltText:          "User profile picture",
		MimeType:         getMimeType(ext),
		Extension:        strings.TrimPrefix(ext, "."),
		Variant:          "avatar",
		FileSize:         0,       // Will be updated after processing
		StorageDriver:    "local", // TODO: Get from storage driver
		StoragePath:      finalPath,
		PublicURL:        publicURL,
	}

	image, err := s.imageRepo.Create(ctx, createReq)
	if err != nil {
		// Cleanup: delete uploaded file if database insert fails
		_ = s.storageDriver.Delete(ctx, finalPath)
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	return image, nil
}

// UploadTenantLogo uploads and processes tenant logo
func (s *ProfileService) UploadTenantLogo(ctx context.Context, file *multipart.FileHeader, tenantUUID, tenantDBCode string, tenantID uuid.UUID) (*tenantmodel.Image, error) {
	// Validate file
	if err := s.validateLogoFile(file); err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var processedFile io.ReadCloser
	var err error

	// Process based on file type
	if ext == ".svg" {
		// SVG files are not processed - upload as-is
		src, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open SVG file: %w", err)
		}
		processedFile = src
	} else {
		// PNG/JPG files need resizing
		processedFile, err = s.processLogoImage(file)
		if err != nil {
			return nil, fmt.Errorf("failed to process logo: %w", err)
		}
	}
	defer processedFile.Close()

	// Generate storage path for tenant logo
	filename := fmt.Sprintf("logo%s", ext)
	storagePath := fmt.Sprintf("%s/images/profiles/tenants/%s/%s", tenantUUID, tenantID.String(), filename)

	// Upload to storage
	finalPath, publicURL, err := s.storageDriver.Upload(ctx, processedFile, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload logo: %w", err)
	}

	// Create image record in database
	createReq := &tenantmodel.CreateImageRequest{
		ImageableType:    "tenant",
		ImageableID:      tenantID,
		Filename:         filename,
		OriginalFilename: file.Filename,
		Title:            "Tenant Logo",
		AltText:          "Company logo",
		MimeType:         getMimeType(ext),
		Extension:        strings.TrimPrefix(ext, "."),
		Variant:          "logo",
		FileSize:         file.Size,
		StorageDriver:    "local", // TODO: Get from storage driver
		StoragePath:      finalPath,
		PublicURL:        publicURL,
	}

	image, err := s.imageRepo.Create(ctx, createReq)
	if err != nil {
		// Cleanup: delete uploaded file if database insert fails
		_ = s.storageDriver.Delete(ctx, finalPath)
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	return image, nil
}

// UploadSysUserAvatar uploads system user avatar (admin users)
func (s *ProfileService) UploadSysUserAvatar(ctx context.Context, file *multipart.FileHeader, sysUserID uuid.UUID) (*tenantmodel.Image, error) {
	// Note: System users don't have tenant isolation, so we use a special path
	// Validate file
	if err := s.validateAvatarFile(file); err != nil {
		return nil, err
	}

	// Process the image
	processedFile, err := s.processAvatarImage(file)
	if err != nil {
		return nil, fmt.Errorf("failed to process avatar: %w", err)
	}
	defer processedFile.Close()

	// Generate storage path for system user avatar
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("avatar%s", ext)
	storagePath := fmt.Sprintf("system/profiles/sys-users/%s/%s", sysUserID.String(), filename)

	// Upload to storage
	finalPath, publicURL, err := s.storageDriver.Upload(ctx, processedFile, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Create image record in database (using master DB for sys users)
	// Note: We might need to create this in a different table or handle differently
	// For now, we'll create a minimal image record
	createReq := &tenantmodel.CreateImageRequest{
		ImageableType:    "sys_user",
		ImageableID:      sysUserID,
		Filename:         filename,
		OriginalFilename: file.Filename,
		Title:            "System User Avatar",
		AltText:          "Administrator profile picture",
		MimeType:         getMimeType(ext),
		Extension:        strings.TrimPrefix(ext, "."),
		Variant:          "avatar",
		FileSize:         0,       // Will be updated after processing
		StorageDriver:    "local", // TODO: Get from storage driver
		StoragePath:      finalPath,
		PublicURL:        publicURL,
	}

	image, err := s.imageRepo.Create(ctx, createReq)
	if err != nil {
		// Cleanup: delete uploaded file if database insert fails
		_ = s.storageDriver.Delete(ctx, finalPath)
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	return image, nil
}

// processAvatarImage resizes image to max 200x200px maintaining aspect ratio
func (s *ProfileService) processAvatarImage(file *multipart.FileHeader) (io.ReadCloser, error) {
	// Open source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Decode image
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to max 200x200px maintaining aspect ratio
	resized := imaging.Fit(img, 200, 200, imaging.Lanczos)

	// Encode to buffer
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(&buf, resized)
	case "gif", "webp":
		// Convert to JPEG for consistency
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85})
	default:
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %w", err)
	}

	return io.NopCloser(&buf), nil
}

// processLogoImage resizes logo to max 120px width or 60px height maintaining aspect ratio
func (s *ProfileService) processLogoImage(file *multipart.FileHeader) (io.ReadCloser, error) {
	// Open source file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Decode image
	img, format, err := image.Decode(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var resized image.Image

	// Calculate new dimensions based on constraints
	if width > 120 || height > 60 {
		// Calculate scale factors
		widthScale := 120.0 / float64(width)
		heightScale := 60.0 / float64(height)

		// Use the smaller scale factor to maintain aspect ratio
		scale := widthScale
		if heightScale < widthScale {
			scale = heightScale
		}

		newWidth := int(float64(width) * scale)
		newHeight := int(float64(height) * scale)

		resized = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	} else {
		// No resizing needed
		resized = img
	}

	// Encode to buffer
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, resized)
	default:
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %w", err)
	}

	return io.NopCloser(&buf), nil
}

// validateAvatarFile validates avatar file constraints
func (s *ProfileService) validateAvatarFile(file *multipart.FileHeader) error {
	// Check file size (max 5MB)
	if file.Size > 5<<20 {
		return fmt.Errorf("avatar file too large: maximum size is 5MB")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isValid := false
	for _, allowedExt := range allowedTypes {
		if ext == allowedExt {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid file type: avatar must be jpg, png, gif, or webp")
	}

	return nil
}

// validateLogoFile validates logo file constraints
func (s *ProfileService) validateLogoFile(file *multipart.FileHeader) error {
	// Check file size (max 2MB)
	if file.Size > 2<<20 {
		return fmt.Errorf("logo file too large: maximum size is 2MB")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedTypes := []string{".jpg", ".jpeg", ".png", ".svg"}
	isValid := false
	for _, allowedExt := range allowedTypes {
		if ext == allowedExt {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid file type: logo must be jpg, png, or svg")
	}

	return nil
}
