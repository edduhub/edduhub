package storage

import (
	"context"
	"errors"
	"io"
	"net/url"
	"runtime"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string `json:"bucket"`
	Region    string
}

type MinioClient struct {
	client *minio.Client
	config *MinioConfig
}

func NewMinioClient(config *MinioConfig) (*MinioClient, error) {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, config.Bucket)
	if err != nil {
		return nil, errors.New("unable to check if bucket exists: " + err.Error())
	}

	if !exists {
		err := minioClient.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{
			Region: config.Region,
		})
		if err != nil {
			return nil, err
		}
	}

	return &MinioClient{
		client: minioClient,
		config: config,
	}, nil
}

func (m *MinioClient) Upload(ctx context.Context, filePath string, bucketName string, objectName string, contentType string) error {
	_, err := m.client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (m *MinioClient) Get(ctx context.Context, filePath string, bucketName string, objectName, contentType string) error {
	err := m.client.FGetObject(ctx, bucketName, objectName, filePath, minio.GetObjectOptions{})
	return err
}

func (m *MinioClient) GetPresignedURL(ctx context.Context, bucketName, objectName string, expirySeconds int64) (string, error) {
	reqParams := url.Values{}
	presignedURL, err := m.client.PresignedGetObject(
		ctx,
		bucketName,
		objectName,
		time.Duration(expirySeconds)*time.Second,
		reqParams,
	)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func (m *MinioClient) UploadFromReader(ctx context.Context, reader io.Reader, bucketName, objectName string, size int64, contentType string) error {
	info, err := m.client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
		NumThreads:  uint(runtime.NumCPU()), // Use the number of CPU cores for parallel uploads
	})
	if err != nil {
		return err
	}
	_ = info
	return nil
}

func (m *MinioClient) GetBucketName() string {
	return m.config.Bucket
}

func (m *MinioClient) GetFileSize(ctx context.Context, bucketName, objectName string) (int, error) {
	info, err := m.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return int(info.Size), nil
}
