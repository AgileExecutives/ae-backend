package services

import (
	"errors"
	"fmt"

	"github.com/unburdy/unburdy-server-api/internal/models"
	"gorm.io/gorm"
)

type ClientService struct {
	db *gorm.DB
}

// NewClientService creates a new client service
func NewClientService(db *gorm.DB) *ClientService {
	return &ClientService{db: db}
}

// CreateClient creates a new client for a specific tenant
func (s *ClientService) CreateClient(req models.CreateClientRequest, tenantID uint) (*models.Client, error) {
	client := models.Client{
		TenantID:             tenantID,
		CostProviderID:       req.CostProviderID,
		FirstName:            req.FirstName,
		LastName:             req.LastName,
		DateOfBirth:          req.DateOfBirth,
		Gender:               req.Gender,
		PrimaryLanguage:      req.PrimaryLanguage,
		ContactFirstName:     req.ContactFirstName,
		ContactLastName:      req.ContactLastName,
		ContactEmail:         req.ContactEmail,
		ContactPhone:         req.ContactPhone,
		AlternativeFirstName: req.AlternativeFirstName,
		AlternativeLastName:  req.AlternativeLastName,
		AlternativePhone:     req.AlternativePhone,
		AlternativeEmail:     req.AlternativeEmail,
		StreetAddress:        req.StreetAddress,
		Zip:                  req.Zip,
		City:                 req.City,
		Email:                req.Email,
		Phone:                req.Phone,
		TherapyTitle:         req.TherapyTitle,
		ProviderApprovalCode: req.ProviderApprovalCode,
		UnitPrice:            req.UnitPrice,
		Status:               req.Status,
		AdmissionDate:        req.AdmissionDate,
		ReferralSource:       req.ReferralSource,
		Notes:                req.Notes,
	}

	// Set default values if not provided
	if client.Gender == "" {
		client.Gender = "undisclosed"
	}
	if client.Status == "" {
		client.Status = "waiting"
	}
	if req.InvoicedIndividually != nil {
		client.InvoicedIndividually = *req.InvoicedIndividually
	}

	if err := s.db.Create(&client).Error; err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &client, nil
}

// GetClientByID returns a client by ID within a tenant with preloaded cost provider
func (s *ClientService) GetClientByID(id, tenantID uint) (*models.Client, error) {
	var client models.Client
	if err := s.db.Preload("CostProvider").Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch client: %w", err)
	}
	return &client, nil
}

// GetAllClients returns all clients with pagination for a tenant with preloaded cost providers
func (s *ClientService) GetAllClients(page, limit int, tenantID uint) ([]models.Client, int, error) {
	var clients []models.Client
	var total int64

	offset := (page - 1) * limit

	// Count total records for the tenant
	if err := s.db.Model(&models.Client{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count clients: %w", err)
	}

	// Get paginated records for the tenant with preloaded cost provider
	if err := s.db.Preload("CostProvider").Where("tenant_id = ?", tenantID).Offset(offset).Limit(limit).Find(&clients).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch clients: %w", err)
	}

	return clients, int(total), nil
}

// UpdateClient updates an existing client within a tenant
func (s *ClientService) UpdateClient(id, tenantID uint, req models.UpdateClientRequest) (*models.Client, error) {
	var client models.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Update fields if provided
	if req.CostProviderID != nil {
		client.CostProviderID = req.CostProviderID
	}
	if req.FirstName != nil {
		client.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		client.LastName = *req.LastName
	}
	if req.DateOfBirth != nil {
		client.DateOfBirth = req.DateOfBirth
	}

	if err := s.db.Save(&client).Error; err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	// Reload client to get updated data with cost provider
	if err := s.db.Preload("CostProvider").First(&client, client.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload client: %w", err)
	}

	return &client, nil
}

// DeleteClient soft deletes a client within a tenant
func (s *ClientService) DeleteClient(id, tenantID uint) error {
	var client models.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("client with ID %d not found", id)
		}
		return fmt.Errorf("failed to get client: %w", err)
	}

	if err := s.db.Delete(&client).Error; err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	return nil
}

// SearchClients searches clients by first name or last name within a tenant
func (s *ClientService) SearchClients(query string, page, limit int, tenantID uint) ([]models.Client, int64, error) {
	var clients []models.Client
	var total int64

	// Build search query - use LIKE for SQLite compatibility, ILIKE for PostgreSQL
	searchPattern := "%" + query + "%"
	searchQuery := s.db.Model(&models.Client{}).Where(
		"tenant_id = ? AND (LOWER(first_name) LIKE LOWER(?) OR LOWER(last_name) LIKE LOWER(?))",
		tenantID, searchPattern, searchPattern,
	)

	// Count total matching records
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count clients: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get paginated search results with cost provider preload
	if err := searchQuery.Preload("CostProvider").Offset(offset).Limit(limit).Find(&clients).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search clients: %w", err)
	}

	return clients, total, nil
}
