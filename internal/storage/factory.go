package storage

import (
	"fmt"
)

// NewStorageDriver creates a storage driver based on configuration
func NewStorageDriver(cfg *Config) (StorageDriver, error) {
	switch cfg.Driver {
	case "local", "":
		// Default to local storage
		uploadsPath := cfg.UploadsPath
		if uploadsPath == "" {
			uploadsPath = "./uploads"
		}
		return NewLocalStorage(uploadsPath), nil

	case "s3":
		// TODO: Implement S3 storage
		// return NewS3Storage(cfg), nil
		return nil, fmt.Errorf("S3 storage not yet implemented")

	case "r2":
		// TODO: Implement R2 storage (Cloudflare R2 uses S3-compatible API)
		// return NewR2Storage(cfg), nil
		return nil, fmt.Errorf("R2 storage not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}
