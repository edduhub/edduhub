// Package config provides file storage configuration for the EduHub application.
// This file implements Minio/S3-compatible storage configuration management.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// StorageConfig holds file storage configuration parameters.
// It supports Minio and S3-compatible storage services.
type StorageConfig struct {
	// Endpoint is the storage server endpoint (e.g., "localhost:9000" for Minio)
	Endpoint string

	// Bucket is the default bucket name for storing files
	Bucket string

	// AccessKey is the access key for storage authentication
	AccessKey string

	// SecretKey is the secret key for storage authentication
	SecretKey string

	// UseSSL indicates whether to use SSL/TLS for connections
	UseSSL bool

	// Region is the storage region (mainly for S3 compatibility)
	Region string

	// PresignedURLExpirySeconds is how long presigned URLs remain valid (default: 3600)
	PresignedURLExpirySeconds int64
}

// LoadStorageConfig loads storage configuration from environment variables.
// It supports both local Minio and cloud S3-compatible storage services.
//
// Environment variables:
//   - STORAGE_ENDPOINT: Storage server endpoint (default: "localhost:9000")
//   - STORAGE_BUCKET: Default bucket name (default: "eduhub")
//   - STORAGE_ACCESS_KEY: Storage access key (optional, for cloud storage)
//   - STORAGE_SECRET_KEY: Storage secret key (optional, for cloud storage)
//   - STORAGE_USE_SSL: Use SSL/TLS (default: "false")
//   - STORAGE_REGION: Storage region (default: "us-east-1")
//   - STORAGE_PRESIGNED_URL_EXPIRY: Presigned URL expiry in seconds (default: "3600")
//
// Returns:
//   - *StorageConfig: The loaded storage configuration
//   - error: Any validation errors
func LoadStorageConfig() (*StorageConfig, error) {
	expiryStr := getEnvOrDefault("STORAGE_PRESIGNED_URL_EXPIRY", "3600")
	expirySeconds, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid STORAGE_PRESIGNED_URL_EXPIRY value: %s", expiryStr)
	}

	config := &StorageConfig{
		Endpoint:                  getEnvOrDefault("STORAGE_ENDPOINT", "localhost:9000"),
		Bucket:                    getEnvOrDefault("STORAGE_BUCKET", "eduhub"),
		AccessKey:                 os.Getenv("STORAGE_ACCESS_KEY"),
		SecretKey:                 os.Getenv("STORAGE_SECRET_KEY"),
		UseSSL:                    getEnvOrDefault("STORAGE_USE_SSL", "false") == "true",
		Region:                    getEnvOrDefault("STORAGE_REGION", "us-east-1"),
		PresignedURLExpirySeconds: expirySeconds,
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("storage config validation failed: %w", err)
	}

	return config, nil
}

// Validate performs validation on the StorageConfig.
// It ensures required parameters are present and valid.
//
// For cloud storage (with access keys), all authentication parameters are required.
// For local Minio, some parameters can be optional if Minio runs without authentication.
//
// Returns an error if validation fails.
func (c *StorageConfig) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("STORAGE_ENDPOINT is required")
	}
	if c.Bucket == "" {
		return fmt.Errorf("STORAGE_BUCKET is required")
	}

	// If access keys are provided, secret key is also required and vice versa
	if (c.AccessKey != "" && c.SecretKey == "") || (c.AccessKey == "" && c.SecretKey != "") {
		return fmt.Errorf("both STORAGE_ACCESS_KEY and STORAGE_SECRET_KEY must be provided together")
	}

	if c.PresignedURLExpirySeconds <= 0 {
		return fmt.Errorf("STORAGE_PRESIGNED_URL_EXPIRY must be greater than 0")
	}

	return nil
}
