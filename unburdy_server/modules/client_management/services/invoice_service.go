package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/gorm"
)

// InvoiceService handles business logic for invoices
type InvoiceService struct {
	db *gorm.DB
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(db *gorm.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

// CreateInvoice creates a new invoice with invoice items
func (s *InvoiceService) CreateInvoice(req entities.CreateInvoiceRequest, tenantID, userID uint) (*entities.Invoice, error) {
	// Verify client exists and get cost provider and organization
	var client entities.Client
	if err := s.db.Preload("CostProvider").Where("id = ? AND tenant_id = ?", req.ClientID, tenantID).First(&client).Error; err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	if client.CostProviderID == nil {
		return nil, errors.New("client must have a cost provider")
	}

	// Get the first organization for this tenant (assuming one org per tenant for now)
	var organization struct {
		ID uint
	}
	if err := s.db.Table("organizations").Select("id").Where("tenant_id = ?", tenantID).First(&organization).Error; err != nil {
		return nil, fmt.Errorf("no organization found for tenant: %w", err)
	}

	// Verify all sessions exist and belong to this client
	var sessions []entities.Session
	if err := s.db.Where("id IN ? AND client_id = ? AND tenant_id = ?", req.SessionIDs, req.ClientID, tenantID).Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	if len(sessions) != len(req.SessionIDs) {
		return nil, errors.New("some sessions not found or don't belong to this client")
	}

	// Check if any sessions are already invoiced
	var existingItems int64
	if err := s.db.Model(&entities.InvoiceItem{}).Where("session_id IN ?", req.SessionIDs).Count(&existingItems).Error; err != nil {
		return nil, fmt.Errorf("failed to check existing invoice items: %w", err)
	}
	if existingItems > 0 {
		return nil, errors.New("one or more sessions are already invoiced")
	}

	// Generate invoice number (simple implementation - can be enhanced)
	invoiceNumber := fmt.Sprintf("INV-%d-%d", tenantID, time.Now().UnixNano())

	// Calculate totals from sessions
	var numberUnits int
	var sumAmount float64
	var taxAmount float64
	var totalAmount float64

	for _ = range sessions {
		numberUnits++
		// Assuming session has cost information - adjust based on actual session structure
		// For now, using placeholder logic
		sessionCost := 100.0 // This should come from session or cost provider
		sumAmount += sessionCost
	}

	// Calculate tax (assuming 19% - should come from organization or settings)
	taxRate := 0.19
	taxAmount = sumAmount * taxRate
	totalAmount = sumAmount + taxAmount

	// Create invoice
	invoice := entities.Invoice{
		TenantID:       tenantID,
		UserID:         userID,
		ClientID:       req.ClientID,
		CostProviderID: *client.CostProviderID,
		OrganizationID: organization.ID,
		InvoiceDate:    time.Now(),
		InvoiceNumber:  invoiceNumber,
		NumberUnits:    numberUnits,
		SumAmount:      sumAmount,
		TaxAmount:      taxAmount,
		TotalAmount:    totalAmount,
		Status:         entities.InvoiceStatusDraft,
		NumReminders:   0,
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create invoice
	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create invoice items
	invoiceItems := make([]entities.InvoiceItem, len(req.SessionIDs))
	for i, sessionID := range req.SessionIDs {
		invoiceItems[i] = entities.InvoiceItem{
			InvoiceID: invoice.ID,
			SessionID: sessionID,
		}
	}

	if err := tx.Create(&invoiceItems).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create invoice items: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload with associations
	return s.GetInvoiceByID(invoice.ID, tenantID, userID)
}

// GetInvoiceByID retrieves an invoice by ID with all associations
func (s *InvoiceService) GetInvoiceByID(id, tenantID, userID uint) (*entities.Invoice, error) {
	var invoice entities.Invoice
	err := s.db.
		Preload("InvoiceItems.Session").
		Preload("Organization").
		Preload("CostProvider").
		Preload("Client").
		Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).
		First(&invoice).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}

	return &invoice, nil
}

// GetInvoices retrieves all invoices with pagination
func (s *InvoiceService) GetInvoices(page, limit int, tenantID, userID uint) ([]entities.Invoice, int64, error) {
	var invoices []entities.Invoice
	var total int64

	offset := (page - 1) * limit

	// Count total
	if err := s.db.Model(&entities.Invoice{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Fetch invoices with associations
	err := s.db.
		Preload("InvoiceItems.Session").
		Preload("Organization").
		Preload("CostProvider").
		Preload("Client").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("invoice_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&invoices).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch invoices: %w", err)
	}

	return invoices, total, nil
}

// UpdateInvoice updates an invoice (status or invoice items)
func (s *InvoiceService) UpdateInvoice(id, tenantID, userID uint, req entities.UpdateInvoiceRequest) (*entities.Invoice, error) {
	// Fetch existing invoice
	var invoice entities.Invoice
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invoice not found")
		}
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}

	// Handle status update
	if req.Status != nil {
		if req.SessionIDs != nil && len(req.SessionIDs) > 0 {
			return nil, errors.New("cannot update both status and invoice items in the same request")
		}

		invoice.Status = *req.Status

		// If status is being set to payed, set payed date
		if *req.Status == entities.InvoiceStatusPayed && invoice.PayedDate == nil {
			now := time.Now()
			invoice.PayedDate = &now
		}

		// If status is being set to reminder, increment reminder count
		if *req.Status == entities.InvoiceStatusReminder {
			invoice.NumReminders++
			now := time.Now()
			invoice.LatestReminder = &now
		}

		if err := s.db.Save(&invoice).Error; err != nil {
			return nil, fmt.Errorf("failed to update invoice status: %w", err)
		}

		return s.GetInvoiceByID(id, tenantID, userID)
	}

	// Handle invoice items update
	if req.SessionIDs != nil && len(req.SessionIDs) > 0 {
		// Verify all sessions exist and belong to the same client
		var sessions []entities.Session
		if err := s.db.Where("id IN ? AND client_id = ? AND tenant_id = ?", req.SessionIDs, invoice.ClientID, tenantID).Find(&sessions).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch sessions: %w", err)
		}

		if len(sessions) != len(req.SessionIDs) {
			return nil, errors.New("some sessions not found or don't belong to this client")
		}

		// Check if any of the new sessions are already invoiced (excluding current invoice)
		var existingItems int64
		if err := s.db.Model(&entities.InvoiceItem{}).
			Where("session_id IN ? AND invoice_id != ?", req.SessionIDs, id).
			Count(&existingItems).Error; err != nil {
			return nil, fmt.Errorf("failed to check existing invoice items: %w", err)
		}
		if existingItems > 0 {
			return nil, errors.New("one or more sessions are already invoiced")
		}

		// Start transaction
		tx := s.db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Delete existing invoice items (permanently, not soft delete)
		if err := tx.Unscoped().Where("invoice_id = ?", id).Delete(&entities.InvoiceItem{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete existing invoice items: %w", err)
		}

		// Create new invoice items
		invoiceItems := make([]entities.InvoiceItem, len(req.SessionIDs))
		for i, sessionID := range req.SessionIDs {
			invoiceItems[i] = entities.InvoiceItem{
				InvoiceID: id,
				SessionID: sessionID,
			}
		}

		if err := tx.Create(&invoiceItems).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create invoice items: %w", err)
		}

		// Recalculate totals
		var numberUnits int
		var sumAmount float64

		for range sessions {
			numberUnits++
			sessionCost := 100.0 // Should come from session or cost provider
			sumAmount += sessionCost
		}

		taxRate := 0.19
		taxAmount := sumAmount * taxRate
		totalAmount := sumAmount + taxAmount

		invoice.NumberUnits = numberUnits
		invoice.SumAmount = sumAmount
		invoice.TaxAmount = taxAmount
		invoice.TotalAmount = totalAmount

		if err := tx.Save(&invoice).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update invoice totals: %w", err)
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		return s.GetInvoiceByID(id, tenantID, userID)
	}

	return nil, errors.New("no update fields provided")
}

// DeleteInvoice deletes an invoice and all its invoice items
func (s *InvoiceService) DeleteInvoice(id, tenantID, userID uint) error {
	// Verify invoice exists and belongs to user/tenant
	var invoice entities.Invoice
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invoice not found")
		}
		return fmt.Errorf("failed to fetch invoice: %w", err)
	}

	// Delete invoice items first (soft delete)
	if err := s.db.Where("invoice_id = ?", id).Delete(&entities.InvoiceItem{}).Error; err != nil {
		return fmt.Errorf("failed to delete invoice items: %w", err)
	}

	// Delete invoice
	if err := s.db.Delete(&invoice).Error; err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	return nil
}

// GetClientsWithUnbilledSessions retrieves all clients with conducted sessions that don't have invoice items
func (s *InvoiceService) GetClientsWithUnbilledSessions(tenantID, userID uint) ([]entities.ClientWithUnbilledSessionsResponse, error) {
	// Get all clients with unbilled conducted sessions
	var clients []entities.Client

	err := s.db.
		Preload("CostProvider").
		Joins("INNER JOIN sessions ON sessions.client_id = clients.id").
		Where("clients.tenant_id = ?", tenantID).
		Where("sessions.tenant_id = ?", tenantID).
		Where("sessions.status = ?", "conducted").
		Where("sessions.id NOT IN (?)",
			s.db.Table("invoice_items").Select("session_id"),
		).
		Group("clients.id").
		Find(&clients).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch clients with unbilled sessions: %w", err)
	}

	// For each client, fetch their unbilled conducted sessions
	result := make([]entities.ClientWithUnbilledSessionsResponse, 0, len(clients))
	for _, client := range clients {
		var sessions []entities.Session
		err := s.db.
			Where("tenant_id = ? AND client_id = ? AND status = ? AND id NOT IN (?)",
				tenantID,
				client.ID,
				"conducted",
				s.db.Table("invoice_items").Select("session_id"),
			).
			Order("original_date ASC").
			Find(&sessions).Error

		if err != nil {
			return nil, fmt.Errorf("failed to fetch unbilled sessions for client %d: %w", client.ID, err)
		}

		sessionResponses := make([]entities.SessionResponse, len(sessions))
		for i, session := range sessions {
			sessionResponses[i] = session.ToResponse()
		}

		result = append(result, entities.ClientWithUnbilledSessionsResponse{
			ClientResponse: client.ToResponse(),
			Sessions:       sessionResponses,
		})
	}

	return result, nil
}
