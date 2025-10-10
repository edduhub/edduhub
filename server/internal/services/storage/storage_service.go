package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
)

type StorageService interface {
	UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error)
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string) (string, error)
	ListFiles(ctx context.Context, prefix string) ([]string, error)
}

type storageService struct {
	minioClient *minio.Client
	bucketName  string
	endpoint    string
	useSSL      bool
}

func NewStorageService(minioClient *minio.Client, bucketName, endpoint string, useSSL bool) StorageService {
	return &storageService{
		minioClient: minioClient,
		bucketName:  bucketName,
		endpoint:    endpoint,
		useSSL:      useSSL,
	}
}

func (s *storageService) UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	if s.minioClient == nil {
		return "", fmt.Errorf("storage service not configured")
	}

	// Ensure bucket exists
	exists, err := s.minioClient.BucketExists(ctx, s.bucketName)
	if err != nil {
		return "", fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = s.minioClient.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Upload file
	_, err = s.minioClient.PutObject(ctx, s.bucketName, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate URL
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	fileURL := fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucketName, objectKey)

	return fileURL, nil
}

func (s *storageService) DeleteFile(ctx context.Context, objectKey string) error {
	if s.minioClient == nil {
		return fmt.Errorf("storage service not configured")
	}

	err := s.minioClient.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *storageService) GetFileURL(ctx context.Context, objectKey string) (string, error) {
	if s.minioClient == nil {
		return "", fmt.Errorf("storage service not configured")
	}

	// Generate presigned URL valid for 1 hour
	url, err := s.minioClient.PresignedGetObject(ctx, s.bucketName, objectKey, time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate URL: %w", err)
	}

	return url.String(), nil
}

func (s *storageService) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	if s.minioClient == nil {
		return nil, fmt.Errorf("storage service not configured")
	}

	var files []string
	objectCh := s.minioClient.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, object.Key)
	}

	return files, nil
}
