package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage implements DocumentStorage using MinIO/S3
type MinIOStorage struct {
	client *minio.Client
	config MinIOConfig
}

// MinIOConfig holds MinIO connection configuration
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
}

// NewMinIOStorage creates a new MinIO storage client
func NewMinIOStorage(config MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinIOStorage{
		client: client,
		config: config,
	}, nil
}

// Store saves document bytes to MinIO
func (s *MinIOStorage) Store(ctx context.Context, req StoreRequest) (string, error) {
	// Ensure bucket exists
	exists, err := s.client.BucketExists(ctx, req.Bucket)
	if err != nil {
		return "", fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, req.Bucket, minio.MakeBucketOptions{
			Region: s.config.Region,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Prepare upload options
	opts := minio.PutObjectOptions{
		ContentType:  req.ContentType,
		UserMetadata: req.Metadata,
	}

	// Upload object
	reader := bytes.NewReader(req.Data)
	_, err = s.client.PutObject(ctx, req.Bucket, req.Key, reader, int64(len(req.Data)), opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	return req.Key, nil
}

// Retrieve gets document bytes from MinIO
func (s *MinIOStorage) Retrieve(ctx context.Context, bucket, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return data, nil
}

// GetURL returns a pre-signed URL for direct download
func (s *MinIOStorage) GetURL(ctx context.Context, bucket, key string, expiresIn time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, key, expiresIn, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// Delete removes a document from MinIO
func (s *MinIOStorage) Delete(ctx context.Context, bucket, key string) error {
	err := s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// List returns documents matching prefix
func (s *MinIOStorage) List(ctx context.Context, bucket, prefix string) ([]DocumentMeta, error) {
	var documents []DocumentMeta

	objectCh := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}

		documents = append(documents, DocumentMeta{
			Key:          object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			ContentType:  object.ContentType,
			ETag:         object.ETag,
		})
	}

	return documents, nil
}

// Exists checks if a document exists in MinIO
func (s *MinIOStorage) Exists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}
