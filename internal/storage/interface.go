package storage

import (
	"context"
	"io"
)

// StorageDriver defines the interface for different storage backends
type StorageDriver interface {
	// Upload uploads a file to storage
	// Returns the storage path and public URL (if applicable)
	Upload(ctx context.Context, file io.Reader, path string) (storagePath string, publicURL string, err error)

	// Delete removes a file from storage
	Delete(ctx context.Context, path string) error

	// GetPublicURL returns the public URL for a file
	// For local storage, this returns the relative path
	// For cloud storage (S3/R2), this returns the full CDN URL
	GetPublicURL(path string) string

	// Exists checks if a file exists in storage
	Exists(ctx context.Context, path string) (bool, error)

	// GetReader returns a reader for the file
	// Used for downloading or processing files
	GetReader(ctx context.Context, path string) (io.ReadCloser, error)
}

// Config holds the storage configuration
type Config struct {
	Driver string // local, s3, r2

	// Local Storage
	UploadsPath string

	// AWS S3
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSBucket          string

	// Cloudflare R2
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2AccountID       string
	R2Bucket          string
	R2PublicURL       string
}
