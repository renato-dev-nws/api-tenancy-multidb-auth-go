package tenant

import (
	"time"

	"github.com/google/uuid"
)

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeDocument MediaType = "document"
)

// ImageVariant represents the size variant of an image
type ImageVariant string

const (
	VariantOriginal ImageVariant = "original"
	VariantMedium   ImageVariant = "medium"
	VariantSmall    ImageVariant = "small"
	VariantThumb    ImageVariant = "thumb"
)

// ProcessingStatus represents the image processing status
type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
)

// Image represents a polymorphic image entity
type Image struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	ImageableType    string           `json:"imageable_type" db:"imageable_type"`
	ImageableID      uuid.UUID        `json:"imageable_id" db:"imageable_id"`
	Filename         string           `json:"filename" db:"filename"`
	OriginalFilename *string          `json:"original_filename,omitempty" db:"original_filename"`
	Title            *string          `json:"title,omitempty" db:"title"`
	AltText          *string          `json:"alt_text,omitempty" db:"alt_text"`
	MediaType        MediaType        `json:"media_type" db:"media_type"`
	MimeType         string           `json:"mime_type" db:"mime_type"`
	Extension        string           `json:"extension" db:"extension"`
	Variant          ImageVariant     `json:"variant" db:"variant"`
	ParentID         *uuid.UUID       `json:"parent_id,omitempty" db:"parent_id"`
	Width            *int             `json:"width,omitempty" db:"width"`
	Height           *int             `json:"height,omitempty" db:"height"`
	FileSize         *int64           `json:"file_size,omitempty" db:"file_size"`
	StorageDriver    string           `json:"storage_driver" db:"storage_driver"`
	StoragePath      string           `json:"storage_path" db:"storage_path"`
	PublicURL        *string          `json:"public_url,omitempty" db:"public_url"`
	ProcessingStatus ProcessingStatus `json:"processing_status" db:"processing_status"`
	ProcessedAt      *time.Time       `json:"processed_at,omitempty" db:"processed_at"`
	DisplayOrder     int              `json:"display_order" db:"display_order"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// CreateImageRequest represents the request to create an image record
type CreateImageRequest struct {
	ImageableType    string       `json:"imageable_type" binding:"required"`
	ImageableID      uuid.UUID    `json:"imageable_id" binding:"required"`
	Filename         string       `json:"filename" binding:"required"`
	OriginalFilename string       `json:"original_filename"`
	Title            string       `json:"title"`
	AltText          string       `json:"alt_text"`
	MimeType         string       `json:"mime_type" binding:"required"`
	Extension        string       `json:"extension" binding:"required"`
	Variant          ImageVariant `json:"variant"`
	ParentID         *uuid.UUID   `json:"parent_id"`
	Width            int          `json:"width"`
	Height           int          `json:"height"`
	FileSize         int64        `json:"file_size"`
	StorageDriver    string       `json:"storage_driver"`
	StoragePath      string       `json:"storage_path" binding:"required"`
	PublicURL        string       `json:"public_url"`
	DisplayOrder     int          `json:"display_order"`
}

// UpdateImageRequest represents the request to update image metadata
type UpdateImageRequest struct {
	Title        *string `json:"title"`
	AltText      *string `json:"alt_text"`
	DisplayOrder *int    `json:"display_order"`
}

// ImageUploadRequest represents the multipart upload request
type ImageUploadRequest struct {
	ImageableType string    `form:"imageable_type" binding:"required"` // product, service
	ImageableID   uuid.UUID `form:"imageable_id" binding:"required"`
	Titles        []string  `form:"titles"`    // Optional titles for each file
	AltTexts      []string  `form:"alt_texts"` // Optional alt texts
}

// ImageUploadResponse represents the response after uploading images
type ImageUploadResponse struct {
	Uploaded int     `json:"uploaded"`
	Images   []Image `json:"images"`
}

// ImageVariantSizes defines the dimensions for each variant
var ImageVariantSizes = map[ImageVariant]struct {
	MaxWidth  int
	MaxHeight int
}{
	VariantOriginal: {MaxWidth: 1400, MaxHeight: 1400},
	VariantMedium:   {MaxWidth: 800, MaxHeight: 800},
	VariantSmall:    {MaxWidth: 350, MaxHeight: 350},
	VariantThumb:    {MaxWidth: 100, MaxHeight: 100},
}
