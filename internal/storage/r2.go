package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// R2Storage implements StorageDriver for Cloudflare R2 (S3-compatible)
type R2Storage struct {
	client    *s3.Client
	bucket    string
	accountID string
	publicURL string // Custom public URL for CDN access
}

// NewR2Storage creates a new R2 storage driver
func NewR2Storage(cfg *Config) (*R2Storage, error) {
	if cfg.R2Bucket == "" {
		return nil, fmt.Errorf("R2 bucket name is required")
	}

	if cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" {
		return nil, fmt.Errorf("R2 credentials are required")
	}

	if cfg.R2AccountID == "" {
		return nil, fmt.Errorf("R2 account ID is required")
	}

	// R2 endpoint format: https://{accountId}.r2.cloudflarestorage.com
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)

	// Create AWS config with R2 endpoint
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"), // R2 uses "auto" region
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2AccessKeyID,
			cfg.R2SecretAccessKey,
			"",
		)),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               endpoint,
					SigningRegion:     "auto",
					HostnameImmutable: true,
				}, nil
			})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	// Create S3-compatible client for R2
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true // R2 requires path-style URLs
	})

	return &R2Storage{
		client:    client,
		bucket:    cfg.R2Bucket,
		accountID: cfg.R2AccountID,
		publicURL: cfg.R2PublicURL, // Optional CDN URL like https://pub-xxxxx.r2.dev
	}, nil
}

// Upload uploads a file to R2
func (r *R2Storage) Upload(ctx context.Context, file io.Reader, path string) (string, string, error) {
	// Read the entire file into memory for R2 upload
	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %w", err)
	}

	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	// Upload to R2
	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(path),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(getContentType(path)),
		ACL:         types.ObjectCannedACLPublicRead, // Make publicly accessible
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	// Generate public URL
	var publicURL string
	if r.publicURL != "" {
		// Use custom CDN URL if provided
		publicURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(r.publicURL, "/"), path)
	} else {
		// Use default R2 public URL format
		publicURL = fmt.Sprintf("https://pub-%s.r2.dev/%s", r.bucket, path)
	}

	return path, publicURL, nil
}

// Delete removes a file from R2
func (r *R2Storage) Delete(ctx context.Context, path string) error {
	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}

	return nil
}

// GetPublicURL returns the public URL for R2 storage
func (r *R2Storage) GetPublicURL(path string) string {
	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	if r.publicURL != "" {
		// Use custom CDN URL if provided
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(r.publicURL, "/"), path)
	}

	// Use default R2 public URL format
	return fmt.Sprintf("https://pub-%s.r2.dev/%s", r.bucket, path)
}

// Exists checks if a file exists in R2
func (r *R2Storage) Exists(ctx context.Context, path string) (bool, error) {
	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check R2 object existence: %w", err)
	}

	return true, nil
}

// GetReader returns a reader for the file from R2
func (r *R2Storage) GetReader(ctx context.Context, path string) (io.ReadCloser, error) {
	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from R2: %w", err)
	}

	return result.Body, nil
}
