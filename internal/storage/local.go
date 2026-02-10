package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements StorageDriver for local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage driver
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
	}
}

// Upload uploads a file to local filesystem
func (s *LocalStorage) Upload(ctx context.Context, file io.Reader, path string) (string, string, error) {
	// Full path on disk
	fullPath := filepath.Join(s.basePath, path)

	// Create directory structure if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	out, err := os.Create(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(out, file); err != nil {
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	// For local storage, storagePath = path and publicURL = /uploads/{path}
	publicURL := fmt.Sprintf("/uploads/%s", path)
	return path, publicURL, nil
}

// Delete removes a file from local filesystem
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to remove empty parent directories
	dir := filepath.Dir(fullPath)
	s.removeEmptyDirs(dir)

	return nil
}

// GetPublicURL returns the public URL for local storage
func (s *LocalStorage) GetPublicURL(path string) string {
	return fmt.Sprintf("/uploads/%s", path)
}

// Exists checks if a file exists on local filesystem
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

// GetReader returns a reader for the file
func (s *LocalStorage) GetReader(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// removeEmptyDirs removes empty parent directories up to basePath
func (s *LocalStorage) removeEmptyDirs(dir string) {
	rel, err := filepath.Rel(s.basePath, dir)
	if err != nil || rel == "." {
		return
	}

	// Try to remove the directory (only succeeds if empty)
	if err := os.Remove(dir); err == nil {
		// Recursively try to remove parent directory
		parent := filepath.Dir(dir)
		s.removeEmptyDirs(parent)
	}
}
