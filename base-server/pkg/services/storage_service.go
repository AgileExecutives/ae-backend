package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ae-base-server/modules/templates/services/storage"
)

// StorageService is a unified service for all MinIO storage operations
type StorageService struct {
	minioStorage  *storage.MinIOStorage
	bucketService *TenantBucketService
}

// NewStorageService creates a new unified storage service
func NewStorageService(minioStorage *storage.MinIOStorage) *StorageService {
	bucketService := NewTenantBucketService(minioStorage)
	return &StorageService{
		minioStorage:  minioStorage,
		bucketService: bucketService,
	}
}

// GetTenantBucketName returns the standardized bucket name for a tenant (4-digit format)
func (s *StorageService) GetTenantBucketName(tenantID uint) string {
	return s.bucketService.GetTenantBucketName(tenantID)
}

// CreateTenantBucket creates a bucket for a tenant if it doesn't exist
func (s *StorageService) CreateTenantBucket(ctx context.Context, tenantID uint) error {
	return s.bucketService.CreateTenantBucket(ctx, tenantID)
}

// StoreTemplate stores a template file in the tenant's bucket
func (s *StorageService) StoreTemplate(ctx context.Context, tenantID uint, templateType, name, content string) (string, error) {
	// Ensure bucket exists
	if err := s.CreateTenantBucket(ctx, tenantID); err != nil {
		return "", fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	// Generate storage key with simplified path
	storageKey := fmt.Sprintf("templates/%s/%s_%d.html",
		templateType,
		name,
		time.Now().Unix(),
	)

	bucketName := s.GetTenantBucketName(tenantID)
	_, err := s.minioStorage.Store(ctx, storage.StoreRequest{
		Bucket:      bucketName,
		Key:         storageKey,
		Data:        []byte(content),
		ContentType: "text/html",
		Metadata: map[string]string{
			"template_type": templateType,
			"template_name": name,
			"tenant_id":     fmt.Sprintf("%d", tenantID),
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to store template: %w", err)
	}

	return storageKey, nil
}

// StoreTemplateWithKey stores a template with a specific storage key
func (s *StorageService) StoreTemplateWithKey(ctx context.Context, tenantID uint, storageKey, content string, metadata map[string]string) error {
	// Ensure bucket exists
	if err := s.CreateTenantBucket(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	bucketName := s.GetTenantBucketName(tenantID)
	_, err := s.minioStorage.Store(ctx, storage.StoreRequest{
		Bucket:      bucketName,
		Key:         storageKey,
		Data:        []byte(content),
		ContentType: "text/html",
		Metadata:    metadata,
	})

	if err != nil {
		return fmt.Errorf("failed to store template: %w", err)
	}

	return nil
}

// RetrieveTemplate retrieves a template from storage
func (s *StorageService) RetrieveTemplate(ctx context.Context, tenantID uint, storageKey string) ([]byte, error) {
	bucketName := s.GetTenantBucketName(tenantID)
	content, err := s.minioStorage.Retrieve(ctx, bucketName, storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve template: %w", err)
	}
	return content, nil
}

// DeleteTemplate deletes a template from storage
func (s *StorageService) DeleteTemplate(ctx context.Context, tenantID uint, storageKey string) error {
	bucketName := s.GetTenantBucketName(tenantID)
	err := s.minioStorage.Delete(ctx, bucketName, storageKey)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

// GetTemplateURL generates a pre-signed URL for template access
func (s *StorageService) GetTemplateURL(ctx context.Context, tenantID uint, storageKey string, expiresIn time.Duration) (string, error) {
	bucketName := s.GetTenantBucketName(tenantID)
	url, err := s.minioStorage.GetURL(ctx, bucketName, storageKey, expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to generate template URL: %w", err)
	}
	return url, nil
}

// BucketExists checks if a tenant bucket exists
func (s *StorageService) BucketExists(ctx context.Context, tenantID uint) (bool, error) {
	return s.bucketService.BucketExists(ctx, tenantID)
}
