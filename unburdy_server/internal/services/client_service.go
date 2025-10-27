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

// CreateClient creates a new client within the user's tenant
func (s *ClientService) CreateClient(req models.CreateClientRequest, userID, tenantID uint) (*models.Client, error) {
	client := models.Client{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		TenantID:    tenantID,
		CreatedBy:   userID,
	}

	if err := s.db.Create(&client).Error; err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Load related data
	if err := s.db.Preload("Tenant").Preload("CreatedByUser").First(&client, client.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load client relations: %w", err)
	}

	return &client, nil
}

// GetClientByID returns a client by ID within a tenant
func (s *ClientService) GetClientByID(id, tenantID uint) (*models.Client, error) {
	var client models.Client
	if err := s.db.Preload("Tenant").Preload("CreatedByUser").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client with ID %d not found in tenant", id)
		}
		return nil, fmt.Errorf("failed to fetch client: %w", err)
	}
	return &client, nil
}

// GetAllClients returns all clients within a tenant with pagination
func (s *ClientService) GetAllClients(page, limit int, tenantID uint) ([]models.Client, int, error) {
	var clients []models.Client
	var total int64

	offset := (page - 1) * limit

	query := s.db.Where("tenant_id = ?", tenantID)

	// Count total records for the tenant
	if err := query.Model(&models.Client{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count clients: %w", err)
	}

	// Get paginated records with relations
	if err := query.Preload("Tenant").Preload("CreatedByUser").
		Offset(offset).Limit(limit).Find(&clients).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch clients: %w", err)
	}

	return clients, int(total), nil
}

// UpdateClient updates an existing client within a tenant
func (s *ClientService) UpdateClient(id, tenantID uint, req models.UpdateClientRequest) (*models.Client, error) {
	var client models.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client with ID %d not found in tenant", id)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Update fields if provided
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

	// Load relations
	if err := s.db.Preload("Tenant").Preload("CreatedByUser").First(&client, client.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load client relations: %w", err)
	}

	return &client, nil
}

// DeleteClient soft deletes a client within a tenant
func (s *ClientService) DeleteClient(id, tenantID uint) error {
	var client models.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("client with ID %d not found in tenant", id)
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

	// Get paginated search results with relations
	if err := searchQuery.Preload("Tenant").Preload("CreatedByUser").
		Offset(offset).Limit(limit).Find(&clients).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search clients: %w", err)
	}

	return clients, total, nil
}
