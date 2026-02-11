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

// S3Storage implements StorageDriver for AWS S3
type S3Storage struct {
	client *s3.Client
	bucket string
	region string
}

// NewS3Storage creates a new S3 storage driver
func NewS3Storage(cfg *Config) (*S3Storage, error) {
	if cfg.AWSBucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}

	if cfg.AWSAccessKeyID == "" || cfg.AWSSecretAccessKey == "" {
		return nil, fmt.Errorf("AWS credentials are required")
	}

	region := cfg.AWSRegion
	if region == "" {
		region = "us-east-1"
	}

	// Create AWS config
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsConfig)

	return &S3Storage{
		client: client,
		bucket: cfg.AWSBucket,
		region: region,
	}, nil
}

// Upload uploads a file to S3
func (s *S3Storage) Upload(ctx context.Context, file io.Reader, path string) (string, string, error) {
	// Read the entire file into memory for S3 upload
	data, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %w", err)
	}

	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	// Upload to S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(path),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(getContentType(path)),
		ACL:         types.ObjectCannedACLPublicRead, // Make publicly accessible
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate public URL
	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, path)

	return path, publicURL, nil
}

// Delete removes a file from S3
func (s *S3Storage) Delete(ctx context.Context, path string) error {
	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// GetPublicURL returns the public URL for S3 storage
func (s *S3Storage) GetPublicURL(path string) string {
	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, path)
}

// Exists checks if a file exists in S3
func (s *S3Storage) Exists(ctx context.Context, path string) (bool, error) {
	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check S3 object existence: %w", err)
	}

	return true, nil
}

// GetReader returns a reader for the file from S3
func (s *S3Storage) GetReader(ctx context.Context, path string) (io.ReadCloser, error) {
	// Clean path (remove leading slash if present)
	path = strings.TrimPrefix(path, "/")

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	return result.Body, nil
}

// getContentType returns the MIME type based on file extension
func getContentType(path string) string {
	if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(path, ".png") {
		return "image/png"
	} else if strings.HasSuffix(path, ".webp") {
		return "image/webp"
	} else if strings.HasSuffix(path, ".gif") {
		return "image/gif"
	}
	return "application/octet-stream"
}
