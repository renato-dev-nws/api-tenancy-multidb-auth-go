package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/saas-multi-database-api/internal/config"
	"github.com/saas-multi-database-api/internal/storage"
)

func main() {
	fmt.Println("=== Storage Driver Test ===")

	// Load configuration
	cfg := config.Load()
	
	fmt.Printf("Storage Driver: %s\n", cfg.Storage.Driver)
	fmt.Printf("Uploads Path: %s\n", cfg.Storage.UploadsPath)

	// Test storage driver initialization
	storageDriver, err := storage.NewStorageDriver(&storage.Config{
		Driver:             cfg.Storage.Driver,
		UploadsPath:        cfg.Storage.UploadsPath,
		AWSAccessKeyID:     cfg.Storage.AWSAccessKeyID,
		AWSSecretAccessKey: cfg.Storage.AWSSecretAccessKey,
		AWSRegion:          cfg.Storage.AWSRegion,
		AWSBucket:          cfg.Storage.AWSBucket,
		R2AccessKeyID:      cfg.Storage.R2AccessKeyID,
		R2SecretAccessKey:  cfg.Storage.R2SecretAccessKey,
		R2AccountID:        cfg.Storage.R2AccountID,
		R2Bucket:           cfg.Storage.R2Bucket,
		R2PublicURL:        cfg.Storage.R2PublicURL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize storage driver: %v", err)
	}

	fmt.Printf("✅ Storage driver initialized successfully!\n")

	// Test public URL generation
	testPath := "test-tenant-uuid/images/product/123/image.jpg"
	publicURL := storageDriver.GetPublicURL(testPath)
	fmt.Printf("✅ Public URL for '%s': %s\n", testPath, publicURL)

	// Test file existence check (for a non-existent file)
	ctx := context.Background()
	exists, err := storageDriver.Exists(ctx, testPath)
	if err != nil {
		fmt.Printf("❌ Error checking file existence: %v\n", err)
	} else {
		fmt.Printf("✅ File existence check works (exists: %v)\n", exists)
	}

	// Test upload with a simple string reader
	testContent := "Hello, World! This is a test file."
	reader := strings.NewReader(testContent)
	uploadPath := "test-uploads/test-file.txt"
	
	storagePath, publicURL, err := storageDriver.Upload(ctx, reader, uploadPath)
	if err != nil {
		fmt.Printf("❌ Error uploading test file: %v\n", err)
	} else {
		fmt.Printf("✅ Test file uploaded successfully!\n")
		fmt.Printf("   Storage Path: %s\n", storagePath)
		fmt.Printf("   Public URL: %s\n", publicURL)
		
		// Clean up test file
		if err := storageDriver.Delete(ctx, storagePath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to delete test file: %v\n", err)
		} else {
			fmt.Printf("✅ Test file deleted successfully\n")
		}
	}

	fmt.Println("\n=== Test completed ===")
}