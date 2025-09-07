package storage

import (
	"context"
	"errors"

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
		errors.Is("Unable to check if bucket exists")
	}
	if !exists {
		err := minioClient.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{
			Region: config.Region,
		})
		if err != nil {
			return &MinioClient{}, err
		}
	}
	return &MinioClient{
		client: minioClient,
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
