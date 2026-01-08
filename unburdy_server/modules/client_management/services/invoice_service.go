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

	// Check if any sessions are already invoiced (now through client_invoices table)
	var existingItems int64
	if err := s.db.Model(&entities.ClientInvoice{}).Where("session_id IN ?", req.SessionIDs).Count(&existingItems).Error; err != nil {
		return nil, fmt.Errorf("failed to check existing client invoices: %w", err)
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

	// Create invoice items and client_invoices entries
	for _, sessionID := range req.SessionIDs {
		// Create invoice item
		invoiceItem := entities.InvoiceItem{
			InvoiceID: invoice.ID,
		}
		if err := tx.Create(&invoiceItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}

		// Create client_invoice entry
		clientInvoice := entities.ClientInvoice{
			InvoiceID:      invoice.ID,
			ClientID:       req.ClientID,
			CostProviderID: *client.CostProviderID,
			SessionID:      sessionID,
			InvoiceItemID:  invoiceItem.ID,
		}
		if err := tx.Create(&clientInvoice).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create client invoice: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload with associations
	return s.GetInvoiceByID(invoice.ID, tenantID, userID)
}

// CreateInvoiceDirect creates a new invoice directly from frontend data
func (s *InvoiceService) CreateInvoiceDirect(req entities.CreateInvoiceDirectRequest, tenantID, userID uint) (*entities.Invoice, error) {
	// Extract client and sessions from unbilledClient
	client := req.UnbilledClient
	params := req.Parameters

	// Verify client exists (using client.ID as client_id)
	var dbClient entities.Client
	if err := s.db.Preload("CostProvider").Where("id = ? AND tenant_id = ?", client.ID, tenantID).First(&dbClient).Error; err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	// Get the organization with full billing configuration
	var org struct {
		ID                      uint
		UnitPrice               *float64
		ExtraEffortsBillingMode string
		ExtraEffortsConfig      []byte
		LineItemSingleUnitText  string
		LineItemDoubleUnitText  string
	}
	if err := s.db.Table("organizations").
		Select("id, unit_price, extra_efforts_billing_mode, extra_efforts_config, line_item_single_unit_text, line_item_double_unit_text").
		Where("tenant_id = ?", tenantID).
		First(&org).Error; err != nil {
		return nil, fmt.Errorf("no organization found for tenant: %w", err)
	}

	// Parse invoice date or use current date
	var invoiceDate time.Time
	if params.InvoiceDate != "" {
		var err error
		invoiceDate, err = time.Parse("2006-01-02", params.InvoiceDate)
		if err != nil {
			return nil, fmt.Errorf("invalid invoice date format: %w", err)
		}
	} else {
		invoiceDate = time.Now()
	}

	// Extract session IDs from sessions array
	sessionIDs := make([]uint, len(client.Sessions))
	for i, session := range client.Sessions {
		sessionIDs[i] = session.ID
	}

	// Verify all sessions exist and belong to this client
	var sessions []entities.Session
	if err := s.db.Where("id IN ? AND client_id = ? AND tenant_id = ?", sessionIDs, client.ID, tenantID).Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	if len(sessions) != len(sessionIDs) {
		return nil, errors.New("some sessions not found or don't belong to this client")
	}

	// Check if any sessions are already invoiced
	var existingClientInvoices int64
	if err := s.db.Model(&entities.ClientInvoice{}).Where("session_id IN ?", sessionIDs).Count(&existingClientInvoices).Error; err != nil {
		return nil, fmt.Errorf("failed to check existing client invoices: %w", err)
	}
	if existingClientInvoices > 0 {
		return nil, errors.New("one or more sessions are already invoiced")
	}

	// Extract extra effort IDs and verify they exist
	effortIDs := make([]uint, len(client.ExtraEfforts))
	for i, effort := range client.ExtraEfforts {
		effortIDs[i] = effort.ID
	}

	var extraEfforts []entities.ExtraEffort
	if len(effortIDs) > 0 {
		if err := s.db.Where("id IN ? AND client_id = ? AND tenant_id = ? AND billing_status = ?",
			effortIDs, client.ID, tenantID, "unbilled").Find(&extraEfforts).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch extra efforts: %w", err)
		}

		if len(extraEfforts) != len(effortIDs) {
			return nil, errors.New("some extra efforts not found or already billed")
		}
	}

	// Use invoice items generator to calculate line items
	taxRate := params.TaxRate
	if taxRate == 0 {
		taxRate = 19.0 // Default tax rate
	}

	unitPrice := 100.0 // Default
	if org.UnitPrice != nil {
		unitPrice = *org.UnitPrice
	} else if client.UnitPrice != nil {
		unitPrice = *client.UnitPrice
	}

	billingConfig := OrganizationBillingConfig{
		BillingMode:    org.ExtraEffortsBillingMode,
		Config:         org.ExtraEffortsConfig,
		UnitPrice:      unitPrice,
		SingleUnitText: org.LineItemSingleUnitText,
		DoubleUnitText: org.LineItemDoubleUnitText,
	}

	generator, err := NewInvoiceItemsGenerator(billingConfig, taxRate)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice generator: %w", err)
	}

	result, err := generator.GenerateItems(sessions, extraEfforts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice items: %w", err)
	}

	// Generate invoice number if not provided
	invoiceNumber := params.InvoiceNumber
	if invoiceNumber == "" {
		invoiceNumber = fmt.Sprintf("INV-%d-%d", time.Now().Year(), time.Now().Unix())
	}

	// Get cost provider ID
	costProviderID := uint(0)
	if client.CostProviderID != nil {
		costProviderID = *client.CostProviderID
	}

	// Create invoice
	invoice := entities.Invoice{
		TenantID:       tenantID,
		UserID:         userID,
		OrganizationID: org.ID,
		InvoiceDate:    invoiceDate,
		InvoiceNumber:  invoiceNumber,
		NumberUnits:    result.TotalUnits,
		SumAmount:      result.SubTotal,
		TaxAmount:      result.TaxAmount,
		TotalAmount:    result.GrandTotal,
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

	// Create invoice items and client_invoices entries based on generated line items
	for _, lineItem := range result.LineItems {
		// Create invoice item with all details
		var unitDurationMin *int
		if lineItem.UnitDurationMin > 0 {
			duration := lineItem.UnitDurationMin
			unitDurationMin = &duration
		}

		invoiceItem := entities.InvoiceItem{
			InvoiceID:       invoice.ID,
			ItemType:        lineItem.ItemType,
			Description:     lineItem.Description,
			NumberUnits:     float64(lineItem.NumberUnits),
			UnitPrice:       lineItem.UnitPrice,
			TotalAmount:     lineItem.TotalAmount,
			UnitDurationMin: unitDurationMin,
			IsEditable:      lineItem.IsEditable,
		}
		if err := tx.Create(&invoiceItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}

		// Create client_invoice entries for each session in this line item
		for _, sessionID := range lineItem.SessionIDs {
			var extraEffortID *uint
			if len(lineItem.ExtraEffortIDs) > 0 {
				extraEffortID = &lineItem.ExtraEffortIDs[0]
			}

			clientInvoice := entities.ClientInvoice{
				InvoiceID:      invoice.ID,
				ClientID:       client.ID,
				CostProviderID: costProviderID,
				SessionID:      sessionID,
				InvoiceItemID:  invoiceItem.ID,
				ExtraEffortID:  extraEffortID,
			}
			if err := tx.Create(&clientInvoice).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create client invoice: %w", err)
			}
		}

		// Mark extra efforts as billed
		if len(lineItem.ExtraEffortIDs) > 0 {
			if err := tx.Model(&entities.ExtraEffort{}).
				Where("id IN ?", lineItem.ExtraEffortIDs).
				Updates(map[string]interface{}{
					"billing_status":  "billed",
					"invoice_item_id": invoiceItem.ID,
				}).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to mark extra efforts as billed: %w", err)
			}
		}
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
		Preload("InvoiceItems").
		Preload("ClientInvoices.Client").
		Preload("ClientInvoices.CostProvider").
		Preload("ClientInvoices.Session").
		Preload("ClientInvoices.InvoiceItem").
		Preload("Organization").
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
		Preload("InvoiceItems").
		Preload("ClientInvoices.Client").
		Preload("ClientInvoices.CostProvider").
		Preload("ClientInvoices.Session").
		Preload("ClientInvoices.InvoiceItem").
		Preload("Organization").
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

		// If status is being set to paid, set paid date
		if *req.Status == entities.InvoiceStatusPaid && invoice.PayedDate == nil {
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
		// Get the first client from existing client_invoices to verify new sessions
		var firstClientInvoice entities.ClientInvoice
		if err := s.db.Where("invoice_id = ?", id).First(&firstClientInvoice).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch existing client invoice: %w", err)
		}

		// Verify all sessions exist and belong to the same client
		var sessions []entities.Session
		if err := s.db.Where("id IN ? AND client_id = ? AND tenant_id = ?", req.SessionIDs, firstClientInvoice.ClientID, tenantID).Find(&sessions).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch sessions: %w", err)
		}

		if len(sessions) != len(req.SessionIDs) {
			return nil, errors.New("some sessions not found or don't belong to this client")
		}

		// Check if any of the new sessions are already invoiced (excluding current invoice)
		var existingItems int64
		if err := s.db.Model(&entities.ClientInvoice{}).
			Where("session_id IN ? AND invoice_id != ?", req.SessionIDs, id).
			Count(&existingItems).Error; err != nil {
			return nil, fmt.Errorf("failed to check existing client invoices: %w", err)
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

		// Delete existing client_invoices and invoice items (permanently, not soft delete)
		if err := tx.Unscoped().Where("invoice_id = ?", id).Delete(&entities.ClientInvoice{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete existing client invoices: %w", err)
		}

		if err := tx.Unscoped().Where("invoice_id = ?", id).Delete(&entities.InvoiceItem{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete existing invoice items: %w", err)
		}

		// Create new invoice items and client_invoices
		for _, sessionID := range req.SessionIDs {
			// Create invoice item
			invoiceItem := entities.InvoiceItem{
				InvoiceID: id,
			}
			if err := tx.Create(&invoiceItem).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create invoice item: %w", err)
			}

			// Create client_invoice entry
			clientInvoice := entities.ClientInvoice{
				InvoiceID:      id,
				ClientID:       firstClientInvoice.ClientID,
				CostProviderID: firstClientInvoice.CostProviderID,
				SessionID:      sessionID,
				InvoiceItemID:  invoiceItem.ID,
			}
			if err := tx.Create(&clientInvoice).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create client invoice: %w", err)
			}
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

// UpdateInvoiceDocumentID updates the document_id for an invoice
func (s *InvoiceService) UpdateInvoiceDocumentID(invoiceID, documentID, tenantID, userID uint) error {
	result := s.db.Model(&entities.Invoice{}).
		Where("id = ? AND tenant_id = ? AND user_id = ?", invoiceID, tenantID, userID).
		Update("document_id", documentID)

	if result.Error != nil {
		return fmt.Errorf("failed to update invoice document_id: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("invoice not found or access denied")
	}

	return nil
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

// CancelInvoice cancels a draft invoice and reverts session/effort states
func (s *InvoiceService) CancelInvoice(id, tenantID, userID uint) error {
	// Verify invoice exists and belongs to user/tenant
	var invoice entities.Invoice
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invoice not found")
		}
		return fmt.Errorf("failed to fetch invoice: %w", err)
	}

	// Only allow canceling draft invoices
	if invoice.Status != entities.InvoiceStatusDraft {
		return fmt.Errorf("only draft invoices can be cancelled, current status: %s", invoice.Status)
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get all client_invoices for this invoice
	var clientInvoices []entities.ClientInvoice
	if err := tx.Where("invoice_id = ?", id).Find(&clientInvoices).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch client invoices: %w", err)
	}

	// Collect session IDs and extra effort IDs
	sessionIDs := make([]uint, 0)
	extraEffortIDs := make([]uint, 0)

	for _, ci := range clientInvoices {
		if ci.SessionID > 0 {
			sessionIDs = append(sessionIDs, ci.SessionID)
		}
		if ci.ExtraEffortID != nil {
			extraEffortIDs = append(extraEffortIDs, *ci.ExtraEffortID)
		}
	}

	// Revert sessions back to conducted status
	if len(sessionIDs) > 0 {
		if err := tx.Model(&entities.Session{}).
			Where("id IN ?", sessionIDs).
			Update("status", "conducted").Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to revert session statuses: %w", err)
		}
	}

	// Revert extra efforts back to unbilled status
	if len(extraEffortIDs) > 0 {
		if err := tx.Model(&entities.ExtraEffort{}).
			Where("id IN ?", extraEffortIDs).
			Updates(map[string]interface{}{
				"billing_status":  "unbilled",
				"invoice_item_id": nil,
			}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to revert extra effort statuses: %w", err)
		}
	}

	// Delete client_invoices (so sessions/efforts can be re-invoiced)
	if err := tx.Where("invoice_id = ?", id).Delete(&entities.ClientInvoice{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete client invoices: %w", err)
	}

	// Set invoice status to cancelled instead of deleting
	if err := tx.Model(&invoice).Update("status", entities.InvoiceStatusCancelled).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetClientsWithUnbilledSessions retrieves all clients with conducted sessions or extra efforts that don't have invoice items
func (s *InvoiceService) GetClientsWithUnbilledSessions(tenantID, userID uint) ([]entities.ClientWithUnbilledSessionsResponse, error) {
	// Get all clients with unbilled conducted sessions OR unbilled extra efforts
	var clients []entities.Client

	// Use a subquery to find clients with either unbilled sessions or unbilled extra efforts
	err := s.db.
		Preload("CostProvider").
		Where("clients.tenant_id = ? AND ("+
			"clients.id IN ("+
			"SELECT DISTINCT client_id FROM sessions "+
			"WHERE tenant_id = ? AND status = 'conducted' AND id NOT IN (SELECT session_id FROM client_invoices WHERE session_id IS NOT NULL)"+
			") OR clients.id IN ("+
			"SELECT DISTINCT client_id FROM extra_efforts "+
			"WHERE tenant_id = ? AND billing_status = 'unbilled' AND billable = true"+
			")"+
			")",
			tenantID, tenantID, tenantID,
		).
		Group("clients.id").
		Find(&clients).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch clients with unbilled sessions or extra efforts: %w", err)
	}

	// For each client, fetch their unbilled conducted sessions and extra efforts
	result := make([]entities.ClientWithUnbilledSessionsResponse, 0, len(clients))
	for _, client := range clients {
		var sessions []entities.Session
		err := s.db.
			Where("tenant_id = ? AND client_id = ? AND status = ? AND id NOT IN (?)",
				tenantID,
				client.ID,
				"conducted",
				s.db.Table("client_invoices").Select("session_id"),
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

		// Fetch unbilled extra efforts for this client
		var extraEfforts []entities.ExtraEffort
		err = s.db.
			Where("tenant_id = ? AND client_id = ? AND billing_status = ? AND billable = ?",
				tenantID,
				client.ID,
				"unbilled",
				true,
			).
			Order("effort_date ASC").
			Find(&extraEfforts).Error

		if err != nil {
			return nil, fmt.Errorf("failed to fetch unbilled extra efforts for client %d: %w", client.ID, err)
		}

		effortResponses := make([]entities.ExtraEffortResponse, len(extraEfforts))
		for i, effort := range extraEfforts {
			effortResponses[i] = entities.ExtraEffortResponse{
				ID:            effort.ID,
				ClientID:      effort.ClientID,
				SessionID:     effort.SessionID,
				EffortType:    effort.EffortType,
				EffortDate:    effort.EffortDate,
				DurationMin:   effort.DurationMin,
				Description:   effort.Description,
				Billable:      effort.Billable,
				BillingStatus: effort.BillingStatus,
				CreatedAt:     effort.CreatedAt,
			}
		}

		result = append(result, entities.ClientWithUnbilledSessionsResponse{
			ClientResponse: client.ToResponse(),
			Sessions:       sessionResponses,
			ExtraEfforts:   effortResponses,
		})
	}

	return result, nil
}
