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
		return NewS3Storage(cfg)

	case "r2":
		return NewR2Storage(cfg)

	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}
