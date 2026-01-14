package services

import (
	"context"
	"fmt"
	"log"

	"github.com/ae-base-server/modules/templates/services/storage"
)

// TenantBucketService handles MinIO bucket operations for tenants
type TenantBucketService struct {
	storage *storage.MinIOStorage
}

// NewTenantBucketService creates a new tenant bucket service
func NewTenantBucketService(storage *storage.MinIOStorage) *TenantBucketService {
	return &TenantBucketService{
		storage: storage,
	}
}

// CreateTenantBucket creates a MinIO bucket for a tenant
func (s *TenantBucketService) CreateTenantBucket(ctx context.Context, tenantID uint) error {
	bucketName := fmt.Sprintf("tenant-%04d", tenantID)

	// Create bucket by storing a dummy object (this will create the bucket if it doesn't exist)
	req := storage.StoreRequest{
		Bucket:      bucketName,
		Key:         ".tenant-init",
		Data:        []byte(fmt.Sprintf("Tenant %d initialized", tenantID)),
		ContentType: "text/plain",
		Metadata: map[string]string{
			"tenant-id":  fmt.Sprintf("%d", tenantID),
			"created-by": "tenant-bucket-service",
			"purpose":    "bucket-initialization",
		},
	}

	_, err := s.storage.Store(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create bucket for tenant %d: %w", tenantID, err)
	}

	log.Printf("✅ Created MinIO bucket for tenant %d: %s", tenantID, bucketName)
	return nil
}

// CreateTenantBuckets creates MinIO buckets for multiple tenants
func (s *TenantBucketService) CreateTenantBuckets(ctx context.Context, tenantIDs []uint) error {
	for _, tenantID := range tenantIDs {
		if err := s.CreateTenantBucket(ctx, tenantID); err != nil {
			log.Printf("❌ Failed to create bucket for tenant %d: %v", tenantID, err)
			// Continue with other tenants instead of failing completely
			continue
		}
	}
	return nil
}

// GetTenantBucketName returns the bucket name for a given tenant ID
func (s *TenantBucketService) GetTenantBucketName(tenantID uint) string {
	return fmt.Sprintf("tenant-%04d", tenantID)
}

// BucketExists checks if a tenant bucket exists
func (s *TenantBucketService) BucketExists(ctx context.Context, tenantID uint) (bool, error) {
	bucketName := s.GetTenantBucketName(tenantID)
	return s.storage.Exists(ctx, bucketName, ".tenant-init")
}
