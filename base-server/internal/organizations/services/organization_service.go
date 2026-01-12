package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/modules/templates/services"
	"gorm.io/gorm"
)

// OrganizationService handles business logic for organizations
type OrganizationService struct {
	db              *gorm.DB
	templateService *services.TemplateService
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

// SetTemplateService sets the template service for copying templates
func (s *OrganizationService) SetTemplateService(templateService *services.TemplateService) {
	s.templateService = templateService
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(req models.CreateOrganizationRequest, tenantID uint) (*models.Organization, error) {
	organization := models.Organization{
		TenantID:                 tenantID,
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

	// Copy templates from tenant 2 org 2 for new organizations (but not for initial seed orgs)
	// Skip tenant 1 org 1 (Unburdy Verwaltung) and tenant 2 org 2 (Standard Organisation)
	if s.templateService != nil && !(tenantID == 1 && organization.ID == 1) && !(tenantID == 2 && organization.ID == 2) {
		log.Printf("📋 Copying templates from tenant 2 org 2 to new organization (tenant %d, org %d)...", tenantID, organization.ID)
		ctx := context.Background()
		if err := s.templateService.CopyTemplatesFromTenant2Org2(ctx, tenantID, organization.ID); err != nil {
			log.Printf("⚠️  Warning: Failed to copy templates for new organization: %v", err)
			// Don't fail organization creation if template copy fails
		}
	}

	return &organization, nil
}

// GetOrganizationByID returns an organization by ID
func (s *OrganizationService) GetOrganizationByID(id, tenantID uint) (*models.Organization, error) {
	var organization models.Organization
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}
	return &organization, nil
}

// GetOrganizations returns all organizations for a tenant with pagination
// This method is exposed for use by other modules
func (s *OrganizationService) GetOrganizations(page, limit int, tenantID uint) ([]models.Organization, int64, error) {
	var organizations []models.Organization
	var total int64

	offset := (page - 1) * limit

	// Count total records
	if err := s.db.Model(&models.Organization{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count organizations: %w", err)
	}

	// Get paginated records
	if err := s.db.Where("tenant_id = ?", tenantID).
		Offset(offset).
		Limit(limit).
		Find(&organizations).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch organizations: %w", err)
	}

	return organizations, total, nil
}

// UpdateOrganization updates an existing organization
func (s *OrganizationService) UpdateOrganization(id, tenantID uint, req models.UpdateOrganizationRequest) (*models.Organization, error) {
	var organization models.Organization
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&organization).Error; err != nil {
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

// UpdateBillingConfig updates an organization's billing configuration for extra efforts
func (s *OrganizationService) UpdateBillingConfig(id, tenantID uint, req models.UpdateBillingConfigRequest) (*models.Organization, error) {
	var organization models.Organization
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("organization with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}

	// Update billing configuration fields
	if req.ExtraEffortsBillingMode != nil {
		organization.ExtraEffortsBillingMode = *req.ExtraEffortsBillingMode
	}
	if req.ExtraEffortsConfig != nil {
		organization.ExtraEffortsConfig = req.ExtraEffortsConfig
	}
	if req.LineItemSingleUnitText != nil {
		organization.LineItemSingleUnitText = *req.LineItemSingleUnitText
	}
	if req.LineItemDoubleUnitText != nil {
		organization.LineItemDoubleUnitText = *req.LineItemDoubleUnitText
	}

	if err := s.db.Save(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to update billing configuration: %w", err)
	}

	return &organization, nil
}

// DeleteOrganization deletes an organization (soft delete)
func (s *OrganizationService) DeleteOrganization(id, tenantID uint) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&models.Organization{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete organization: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("organization with ID %d not found", id)
	}
	return nil
}
