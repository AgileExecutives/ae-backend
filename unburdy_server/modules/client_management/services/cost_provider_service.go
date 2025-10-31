package services

import (
	"errors"
	"fmt"

	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/gorm"
)

type CostProviderService struct {
	db *gorm.DB
}

func NewCostProviderService(db *gorm.DB) *CostProviderService {
	return &CostProviderService{
		db: db,
	}
}

// CreateCostProvider creates a new cost provider for a specific tenant
func (s *CostProviderService) CreateCostProvider(req entities.CreateCostProviderRequest, tenantID uint) (*entities.CostProvider, error) {
	costProvider := entities.CostProvider{
		TenantID:      tenantID,
		Organization:  req.Organization,
		Department:    req.Department,
		ContactName:   req.ContactName,
		StreetAddress: req.StreetAddress,
		Zip:           req.Zip,
		City:          req.City,
	}

	if err := s.db.Create(&costProvider).Error; err != nil {
		return nil, fmt.Errorf("failed to create cost provider: %w", err)
	}

	return &costProvider, nil
}

// GetCostProviderByID returns a cost provider by ID within a tenant
func (s *CostProviderService) GetCostProviderByID(id, tenantID uint) (*entities.CostProvider, error) {
	var costProvider entities.CostProvider
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&costProvider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cost provider with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch cost provider: %w", err)
	}
	return &costProvider, nil
}

// GetAllCostProviders returns all cost providers with pagination for a tenant
func (s *CostProviderService) GetAllCostProviders(page, limit int, tenantID uint) ([]entities.CostProvider, int64, error) {
	var costProviders []entities.CostProvider
	var total int64

	offset := (page - 1) * limit

	// Count total records for the tenant
	if err := s.db.Model(&entities.CostProvider{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count cost providers: %w", err)
	}

	// Get paginated records for the tenant
	if err := s.db.Where("tenant_id = ?", tenantID).Offset(offset).Limit(limit).Find(&costProviders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch cost providers: %w", err)
	}

	return costProviders, total, nil
}

// UpdateCostProvider updates an existing cost provider within a tenant
func (s *CostProviderService) UpdateCostProvider(id, tenantID uint, req entities.UpdateCostProviderRequest) (*entities.CostProvider, error) {
	var costProvider entities.CostProvider
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&costProvider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cost provider with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get cost provider: %w", err)
	}

	// Update fields if provided
	if req.Organization != nil {
		costProvider.Organization = *req.Organization
	}
	if req.Department != nil {
		costProvider.Department = *req.Department
	}
	if req.ContactName != nil {
		costProvider.ContactName = *req.ContactName
	}
	if req.StreetAddress != nil {
		costProvider.StreetAddress = *req.StreetAddress
	}
	if req.Zip != nil {
		costProvider.Zip = *req.Zip
	}
	if req.City != nil {
		costProvider.City = *req.City
	}

	if err := s.db.Save(&costProvider).Error; err != nil {
		return nil, fmt.Errorf("failed to update cost provider: %w", err)
	}

	return &costProvider, nil
}

// DeleteCostProvider soft deletes a cost provider within a tenant
func (s *CostProviderService) DeleteCostProvider(id, tenantID uint) error {
	var costProvider entities.CostProvider
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&costProvider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("cost provider with ID %d not found", id)
		}
		return fmt.Errorf("failed to get cost provider: %w", err)
	}

	if err := s.db.Delete(&costProvider).Error; err != nil {
		return fmt.Errorf("failed to delete cost provider: %w", err)
	}

	return nil
}

// SearchCostProviders searches cost providers by organization name for a tenant
func (s *CostProviderService) SearchCostProviders(query string, page, limit int, tenantID uint) ([]entities.CostProvider, int64, error) {
	var costProviders []entities.CostProvider
	var total int64

	// Build search query for the tenant
	searchPattern := "%" + query + "%"
	searchQuery := s.db.Model(&entities.CostProvider{}).Where(
		"tenant_id = ? AND (LOWER(organization) LIKE LOWER(?) OR LOWER(contact_name) LIKE LOWER(?))",
		tenantID, searchPattern, searchPattern,
	)

	// Count total matching records
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count cost providers: %w", err)
	}

	// Get paginated search results
	offset := (page - 1) * limit
	if err := searchQuery.Offset(offset).Limit(limit).Find(&costProviders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search cost providers: %w", err)
	}

	return costProviders, total, nil
}
