package services

import (
	"context"
	"fmt"
	"log"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/services"
	"gorm.io/gorm"
)

// TenantService handles tenant operations
type TenantService struct {
	db                  *gorm.DB
	tenantBucketService *services.TenantBucketService
}

// NewTenantService creates a new tenant service
func NewTenantService(db *gorm.DB, tenantBucketService *services.TenantBucketService) *TenantService {
	return &TenantService{
		db:                  db,
		tenantBucketService: tenantBucketService,
	}
}

// CreateTenant creates a new tenant and its MinIO bucket
func (s *TenantService) CreateTenant(ctx context.Context, req models.TenantCreateRequest) (*models.Tenant, error) {
	// Create tenant in database
	tenant := models.Tenant{
		CustomerID: req.CustomerID,
		Name:       req.Name,
		Slug:       req.Slug,
	}

	if err := s.db.Create(&tenant).Error; err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	log.Printf("✅ Created tenant: %s (ID: %d)", tenant.Name, tenant.ID)

	// Create MinIO bucket for the tenant
	if s.tenantBucketService != nil {
		if err := s.tenantBucketService.CreateTenantBucket(ctx, tenant.ID); err != nil {
			log.Printf("❌ Warning: Failed to create MinIO bucket for tenant %d: %v", tenant.ID, err)
			// Don't fail tenant creation if bucket creation fails
		}
	}

	return &tenant, nil
}

// CreateTenantWithoutBucket creates a tenant without creating MinIO bucket (for migration/seeding)
func (s *TenantService) CreateTenantWithoutBucket(ctx context.Context, req models.TenantCreateRequest) (*models.Tenant, error) {
	tenant := models.Tenant{
		CustomerID: req.CustomerID,
		Name:       req.Name,
		Slug:       req.Slug,
	}

	if err := s.db.Create(&tenant).Error; err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return &tenant, nil
}

// EnsureTenantBucket ensures a tenant has a MinIO bucket created
func (s *TenantService) EnsureTenantBucket(ctx context.Context, tenantID uint) error {
	if s.tenantBucketService == nil {
		return fmt.Errorf("tenant bucket service not available")
	}

	// Check if bucket already exists
	exists, err := s.tenantBucketService.BucketExists(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if exists {
		log.Printf("✅ MinIO bucket already exists for tenant %d", tenantID)
		return nil
	}

	// Create bucket
	return s.tenantBucketService.CreateTenantBucket(ctx, tenantID)
}

// GetTenant retrieves a tenant by ID
func (s *TenantService) GetTenant(ctx context.Context, tenantID uint) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := s.db.First(&tenant, tenantID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return &tenant, nil
}

// GetAllTenants retrieves all tenants
func (s *TenantService) GetAllTenants(ctx context.Context) ([]models.Tenant, error) {
	var tenants []models.Tenant
	if err := s.db.Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("failed to get tenants: %w", err)
	}
	return tenants, nil
}