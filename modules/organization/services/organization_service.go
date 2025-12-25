package services

import (
	"errors"
	"fmt"

	"github.com/unburdy/organization-module/entities"
	"gorm.io/gorm"
)

// OrganizationService handles business logic for organizations
type OrganizationService struct {
	db *gorm.DB
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(req entities.CreateOrganizationRequest, tenantID, userID uint) (*entities.Organization, error) {
	organization := entities.Organization{
		TenantID:                 tenantID,
		UserID:                   userID,
		Name:                     req.Name,
		OwnerName:                req.OwnerName,
		OwnerTitle:               req.OwnerTitle,
		StreetAddress:            req.StreetAddress,
		Zip:                      req.Zip,
		City:                     req.City,
		Email:                    req.Email,
		Phone:                    req.Phone,
		TaxID:                    req.TaxID,
		TaxRate:                  req.TaxRate,
		TaxUstID:                 req.TaxUstID,
		UnitPrice:                req.UnitPrice,
		BankAccountOwner:         req.BankAccountOwner,
		BankAccountBank:          req.BankAccountBank,
		BankAccountBIC:           req.BankAccountBIC,
		BankAccountIBAN:          req.BankAccountIBAN,
		AdditionalPaymentMethods: req.AdditionalPaymentMethods,
		InvoiceContent:           req.InvoiceContent,
	}

	if err := s.db.Create(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &organization, nil
}

// GetOrganizationByID returns an organization by ID
func (s *OrganizationService) GetOrganizationByID(id, tenantID, userID uint) (*entities.Organization, error) {
	var organization entities.Organization
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}
	return &organization, nil
}

// GetOrganizations returns all organizations for a tenant and user with pagination
// This method is exposed for use by other modules
func (s *OrganizationService) GetOrganizations(page, limit int, tenantID, userID uint) ([]entities.Organization, int64, error) {
	var organizations []entities.Organization
	var total int64

	offset := (page - 1) * limit

	// Count total records
	if err := s.db.Model(&entities.Organization{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Get paginated records
	if err := s.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Offset(offset).
		Limit(limit).
		Find(&organizations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch organizations: %w", err)
	}

	return organizations, total, nil
}

// UpdateOrganization updates an existing organization
func (s *OrganizationService) UpdateOrganization(id, tenantID, userID uint, req entities.UpdateOrganizationRequest) (*entities.Organization, error) {
	var organization entities.Organization
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		organization.Name = *req.Name
	}
	if req.OwnerName != nil {
		organization.OwnerName = *req.OwnerName
	}
	if req.OwnerTitle != nil {
		organization.OwnerTitle = *req.OwnerTitle
	}
	if req.StreetAddress != nil {
		organization.StreetAddress = *req.StreetAddress
	}
	if req.Zip != nil {
		organization.Zip = *req.Zip
	}
	if req.City != nil {
		organization.City = *req.City
	}
	if req.Email != nil {
		organization.Email = *req.Email
	}
	if req.Phone != nil {
		organization.Phone = *req.Phone
	}
	if req.TaxID != nil {
		organization.TaxID = *req.TaxID
	}
	if req.TaxRate != nil {
		organization.TaxRate = req.TaxRate
	}
	if req.TaxUstID != nil {
		organization.TaxUstID = *req.TaxUstID
	}
	if req.UnitPrice != nil {
		organization.UnitPrice = req.UnitPrice
	}
	if req.BankAccountOwner != nil {
		organization.BankAccountOwner = *req.BankAccountOwner
	}
	if req.BankAccountBank != nil {
		organization.BankAccountBank = *req.BankAccountBank
	}
	if req.BankAccountBIC != nil {
		organization.BankAccountBIC = *req.BankAccountBIC
	}
	if req.BankAccountIBAN != nil {
		organization.BankAccountIBAN = *req.BankAccountIBAN
	}
	if req.AdditionalPaymentMethods != nil {
		organization.AdditionalPaymentMethods = req.AdditionalPaymentMethods
	}
	if req.InvoiceContent != nil {
		organization.InvoiceContent = req.InvoiceContent
	}

	if err := s.db.Save(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return &organization, nil
}

// DeleteOrganization deletes an organization (soft delete)
func (s *OrganizationService) DeleteOrganization(id, tenantID, userID uint) error {
	result := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).Delete(&entities.Organization{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete organization: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("organization with ID %d not found", id)
	}
	return nil
}
