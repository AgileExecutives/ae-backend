package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	baseAPI "github.com/ae-base-server/api"
	documentStorage "github.com/unburdy/documents-module/services/storage"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/gorm"
)

// InvoiceService handles business logic for invoices
type InvoiceService struct {
	db         *gorm.DB
	pdfService *InvoicePDFService
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(db *gorm.DB) *InvoiceService {
	return &InvoiceService{
		db: db,
	}
}

// SetPDFService sets the PDF service (for dependency injection)
func (s *InvoiceService) SetPDFService(pdfService *InvoicePDFService) {
	s.pdfService = pdfService
}

// GetPDFService returns the PDF service
func (s *InvoiceService) GetPDFService() *InvoicePDFService {
	return s.pdfService
}

// SetDocumentStorage initializes the PDF service with document storage
func (s *InvoiceService) SetDocumentStorage(storage documentStorage.DocumentStorage) {
	if s.pdfService == nil {
		s.pdfService = NewInvoicePDFService(s.db, storage)
	}
}

// GenerateInvoicePDF generates PDF bytes for an invoice
func (s *InvoiceService) GenerateInvoicePDF(invoiceID, tenantID uint) ([]byte, error) {
	if s.pdfService == nil {
		return nil, errors.New("PDF service not available")
	}

	// Get invoice with all related data
	var invoice entities.Invoice
	if err := s.db.Preload("Organization").
		Preload("InvoiceItems").
		Preload("ClientInvoices.Client").
		Preload("ClientInvoices.CostProvider").
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	ctx := context.Background()
	pdfBytes, err := s.pdfService.GenerateInvoicePDF(ctx, &invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfBytes, nil
}

// CreateDraftInvoice creates a new draft invoice with sessions, extra efforts, and custom line items
func (s *InvoiceService) CreateDraftInvoice(req entities.CreateDraftInvoiceRequest, tenantID, userID uint) (*entities.Invoice, error) {
	// Verify client exists and get cost provider and organization
	var client entities.Client
	if err := s.db.Preload("CostProvider").Where("id = ? AND tenant_id = ?", req.ClientID, tenantID).First(&client).Error; err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	if client.CostProviderID == nil {
		return nil, errors.New("client must have a cost provider")
	}

	// Get the first organization for this tenant
	var organization struct {
		ID      uint
		TaxRate *float64
	}
	if err := s.db.Table("organizations").
		Select("id, tax_rate").
		Where("tenant_id = ?", tenantID).
		First(&organization).Error; err != nil {
		return nil, fmt.Errorf("no organization found for tenant: %w", err)
	}

	// Get billing settings
	settingsHelper := NewSettingsHelper(s.db)
	taxSettings, err := settingsHelper.GetBillingTax(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax settings: %w", err)
	}

	// Validate at least one billable item
	if len(req.SessionIDs) == 0 && len(req.ExtraEffortIDs) == 0 && len(req.CustomLineItems) == 0 {
		return nil, errors.New("invoice must have at least one session, extra effort, or custom line item")
	}

	// Verify all sessions exist, belong to client, and are in 'conducted' status
	var sessions []entities.Session
	if len(req.SessionIDs) > 0 {
		if err := s.db.Where("id IN ? AND client_id = ? AND tenant_id = ? AND status = ?",
			req.SessionIDs, req.ClientID, tenantID, "conducted").Find(&sessions).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch sessions: %w", err)
		}

		if len(sessions) != len(req.SessionIDs) {
			// Provide detailed error information for debugging
			var foundSessionIDs []uint
			for _, session := range sessions {
				foundSessionIDs = append(foundSessionIDs, session.ID)
			}

			// Check which sessions have issues and categorize them
			var allSessionsDebug []struct {
				ID       uint   `json:"id"`
				ClientID uint   `json:"client_id"`
				TenantID uint   `json:"tenant_id"`
				Status   string `json:"status"`
			}
			s.db.Table("sessions").Select("id, client_id, tenant_id, status").
				Where("id IN ?", req.SessionIDs).Find(&allSessionsDebug)

			var wrongStatus, wrongClient, notFound []uint
			foundIDs := make(map[uint]bool)
			for _, dbSession := range allSessionsDebug {
				foundIDs[dbSession.ID] = true
				if dbSession.TenantID != tenantID {
					wrongClient = append(wrongClient, dbSession.ID)
				} else if dbSession.ClientID != req.ClientID {
					wrongClient = append(wrongClient, dbSession.ID)
				} else if dbSession.Status != "conducted" {
					wrongStatus = append(wrongStatus, dbSession.ID)
				}
			}

			for _, requestedID := range req.SessionIDs {
				if !foundIDs[requestedID] {
					notFound = append(notFound, requestedID)
				}
			}

			errorMsg := "Cannot create invoice: "
			if len(notFound) > 0 {
				errorMsg += fmt.Sprintf("Sessions not found: %v. ", notFound)
			}
			if len(wrongClient) > 0 {
				errorMsg += fmt.Sprintf("Sessions don't belong to this client: %v. ", wrongClient)
			}
			if len(wrongStatus) > 0 {
				errorMsg += fmt.Sprintf("Sessions not in 'conducted' status: %v (they need to be marked as completed first). ", wrongStatus)
			}

			return nil, errors.New(errorMsg)
		}
	}

	// Verify all extra efforts exist, belong to client, and are in 'delivered' status
	var extraEfforts []entities.ExtraEffort
	if len(req.ExtraEffortIDs) > 0 {
		if err := s.db.Where("id IN ? AND client_id = ? AND tenant_id = ? AND billing_status = ? AND billable = ?",
			req.ExtraEffortIDs, req.ClientID, tenantID, "delivered", true).Find(&extraEfforts).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch extra efforts: %w", err)
		}

		if len(extraEfforts) != len(req.ExtraEffortIDs) {
			return nil, errors.New("some extra efforts not found, don't belong to this client, or are not billable/delivered")
		}
	}

	// Calculate totals
	var subtotal float64
	var taxAmount float64
	var numberUnits int
	invoiceItems := make([]entities.InvoiceItem, 0)

	// Add session line items
	for _, session := range sessions {
		numberUnits += session.NumberUnits
		// Get unit price from organization or cost provider (placeholder for now)
		unitPrice := 150.0 // TODO: Get from organization.UnitPrice or CostProvider
		totalItemAmount := float64(session.NumberUnits) * unitPrice
		subtotal += totalItemAmount

		item := entities.InvoiceItem{
			ItemType:        "session",
			SessionID:       &session.ID,
			Description:     fmt.Sprintf("Therapiesitzung vom %s", session.OriginalDate.Format("02.01.2006")),
			NumberUnits:     float64(session.NumberUnits),
			UnitPrice:       unitPrice,
			TotalAmount:     totalItemAmount,
			UnitDurationMin: &session.DurationMin,
			IsEditable:      false, // Session-based items not editable
			VATExempt:       taxSettings.VATExempt,
		}

		// Set VAT rate: 0 if exempt, otherwise use settings rate
		if item.VATExempt {
			item.VATRate = 0
			item.VATExemptionText = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
		} else {
			item.VATRate = taxSettings.VATRate
		}

		invoiceItems = append(invoiceItems, item)
	}

	// Add extra effort line items
	for _, effort := range extraEfforts {
		// Calculate unit price based on duration (placeholder logic)
		unitPrice := 150.0 / 60.0 * float64(effort.DurationMin) // Proportional to hour rate
		totalItemAmount := unitPrice

		subtotal += totalItemAmount

		item := entities.InvoiceItem{
			ItemType:        "extra_effort",
			SourceEffortID:  &effort.ID,
			Description:     fmt.Sprintf("%s: %s", effort.EffortType, effort.Description),
			NumberUnits:     1,
			UnitPrice:       unitPrice,
			TotalAmount:     totalItemAmount,
			UnitDurationMin: &effort.DurationMin,
			IsEditable:      false,
			VATExempt:       taxSettings.VATExempt,
		}

		// Set VAT rate: 0 if exempt, otherwise use settings rate
		if item.VATExempt {
			item.VATRate = 0
			item.VATExemptionText = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
		} else {
			item.VATRate = taxSettings.VATRate
		}

		invoiceItems = append(invoiceItems, item)
	}

	// Add custom line items
	vatService := NewVATService()

	for _, customItem := range req.CustomLineItems {
		totalItemAmount := customItem.NumberUnits * customItem.UnitPrice
		subtotal += totalItemAmount

		item := entities.InvoiceItem{
			ItemType:    "custom",
			Description: customItem.Description,
			NumberUnits: customItem.NumberUnits,
			UnitPrice:   customItem.UnitPrice,
			TotalAmount: totalItemAmount,
			IsEditable:  true, // Custom items are editable
		}

		// Apply VAT category if provided
		if customItem.VATCategory != "" {
			err := vatService.ApplyVATCategory(&item, VATCategory(customItem.VATCategory))
			if err != nil {
				// Fallback to manual VAT settings if category is invalid
				item.VATRate = customItem.VATRate
				item.VATExempt = customItem.VATExempt
				item.VATExemptionText = customItem.VATExemptionText
			}
		} else {
			// Manual VAT settings
			vat_rate := customItem.VATRate
			if vat_rate == 0 && !customItem.VATExempt {
				vat_rate = taxSettings.VATRate
			}
			item.VATRate = vat_rate
			item.VATExempt = customItem.VATExempt
			item.VATExemptionText = customItem.VATExemptionText
		}

		// Set default exemption text if needed
		vatService.SetDefaultVATExemptionText(&item)

		invoiceItems = append(invoiceItems, item)
	}

	// Calculate tax using VAT service
	vatSummary := vatService.CalculateInvoiceVAT(invoiceItems)
	taxAmount = vatSummary.TaxAmount
	totalAmount := vatSummary.TotalAmount

	// Populate customer fields from request or cost provider or client
	var customerName, customerAddress, customerAddressExt, customerZip, customerCity, customerCountry, customerContactPerson, customerDepartment, customerEmail string

	// Priority 1: Use values from request if provided
	if req.CustomerName != "" {
		customerName = req.CustomerName
		customerAddress = req.CustomerAddress
		customerAddressExt = req.CustomerAddressExt
		customerZip = req.CustomerZip
		customerCity = req.CustomerCity
		customerCountry = req.CustomerCountry
		customerContactPerson = req.CustomerContactPerson
		customerDepartment = req.CustomerDepartment
		customerEmail = req.CustomerEmail
	} else {
		// Priority 2: Use cost provider if available and has data
		if client.CostProvider != nil {
			if client.CostProvider.Organization != "" {
				customerName = client.CostProvider.Organization
				customerAddress = client.CostProvider.StreetAddress
				customerZip = client.CostProvider.Zip
				customerCity = client.CostProvider.City
				customerContactPerson = client.CostProvider.ContactName
				customerDepartment = client.CostProvider.Department
				// Cost provider doesn't have email or country, leave empty
			}
		}

		// Priority 3: Fall back to client contact if cost provider didn't provide data
		if customerName == "" {
			// Use client name as customer
			customerName = client.FirstName + " " + client.LastName
			customerAddress = client.StreetAddress
			customerZip = client.Zip
			customerCity = client.City

			// Use client contact person if available
			if client.ContactFirstName != "" || client.ContactLastName != "" {
				customerContactPerson = strings.TrimSpace(client.ContactFirstName + " " + client.ContactLastName)
			}

			customerEmail = client.ContactEmail
			if customerEmail == "" {
				customerEmail = client.Email // Fall back to client's own email
			}
		}
	}

	// Create invoice (without invoice number - assigned on finalization)
	invoice := entities.Invoice{
		TenantID:              tenantID,
		UserID:                userID,
		OrganizationID:        organization.ID,
		InvoiceDate:           time.Now(),
		InvoiceNumber:         fmt.Sprintf("DRAFT-%d-%d", tenantID, time.Now().UnixNano()), // Unique temporary number
		NumberUnits:           numberUnits,
		SumAmount:             subtotal,
		TaxAmount:             taxAmount,
		TotalAmount:           totalAmount,
		Status:                entities.InvoiceStatusDraft,
		NumReminders:          0,
		CustomerName:          customerName,
		CustomerAddress:       customerAddress,
		CustomerAddressExt:    customerAddressExt,
		CustomerZip:           customerZip,
		CustomerCity:          customerCity,
		CustomerCountry:       customerCountry,
		CustomerContactPerson: customerContactPerson,
		CustomerDepartment:    customerDepartment,
		CustomerEmail:         customerEmail,
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Save invoice
	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Save invoice items
	for i := range invoiceItems {
		invoiceItems[i].InvoiceID = invoice.ID
		if err := tx.Create(&invoiceItems[i]).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}
	}

	// Create client_invoices relationships for all sessions
	// This is required for PDF generation to load client and cost_provider data
	for idx, session := range sessions {
		sessionID := session.ID
		clientInvoice := entities.ClientInvoice{
			InvoiceID:      invoice.ID,
			ClientID:       req.ClientID,
			CostProviderID: *client.CostProviderID,
			SessionID:      &sessionID,
			InvoiceItemID:  invoiceItems[idx].ID, // Match session to its invoice item
		}

		if err := tx.Create(&clientInvoice).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create client_invoices relationship: %w", err)
		}
	}

	// If there are extra efforts, create client_invoice records for them too
	for i, effort := range extraEfforts {
		// Find the corresponding invoice item (after sessions)
		itemIdx := len(sessions) + i
		if itemIdx < len(invoiceItems) {
			clientInvoice := entities.ClientInvoice{
				InvoiceID:      invoice.ID,
				ClientID:       req.ClientID,
				CostProviderID: *client.CostProviderID,
				SessionID:      nil, // NULL for extra efforts
				InvoiceItemID:  invoiceItems[itemIdx].ID,
				ExtraEffortID:  &effort.ID,
			}

			if err := tx.Create(&clientInvoice).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create client_invoices relationship for extra effort: %w", err)
			}
		}
	}

	// Update session statuses to 'invoice-draft'
	if len(req.SessionIDs) > 0 {
		if err := tx.Model(&entities.Session{}).
			Where("id IN ?", req.SessionIDs).
			Update("status", "invoice-draft").Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update session statuses: %w", err)
		}
	}

	// Update extra effort statuses to 'invoice-draft'
	if len(req.ExtraEffortIDs) > 0 {
		if err := tx.Model(&entities.ExtraEffort{}).
			Where("id IN ?", req.ExtraEffortIDs).
			Update("billing_status", "invoice-draft").Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update extra effort statuses: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload invoice with items, organization, and client relationships
	if err := s.db.Preload("InvoiceItems").
		Preload("Organization").
		Preload("ClientInvoices.Client").
		Preload("ClientInvoices.CostProvider").
		Preload("ClientInvoices.Session").
		First(&invoice, invoice.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// UpdateDraftInvoice updates an existing draft invoice by adding/removing items
func (s *InvoiceService) UpdateDraftInvoice(invoiceID, tenantID, userID uint, req entities.UpdateDraftInvoiceRequest) (*entities.Invoice, error) {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.Preload("InvoiceItems").Where("id = ? AND tenant_id = ?", invoiceID, tenantID).First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Verify invoice is in draft status
	if invoice.Status != entities.InvoiceStatusDraft {
		return nil, errors.New("can only edit invoices in draft status")
	}

	// Get organization ID
	var organization struct {
		ID uint
	}
	if err := s.db.Table("organizations").
		Select("id").
		Where("id = ?", invoice.OrganizationID).
		First(&organization).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}

	// Get billing settings
	settingsHelper := NewSettingsHelper(s.db)
	taxSettings, err := settingsHelper.GetBillingTax(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tax settings: %w", err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Handle removing sessions
	if len(req.RemoveSessionIDs) > 0 {
		// Find invoice items linked to these sessions
		var itemsToRemove []entities.InvoiceItem
		if err := tx.Where("invoice_id = ? AND session_id IN ?", invoiceID, req.RemoveSessionIDs).Find(&itemsToRemove).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to find items to remove: %w", err)
		}

		// Delete the invoice items
		if err := tx.Where("invoice_id = ? AND session_id IN ?", invoiceID, req.RemoveSessionIDs).Delete(&entities.InvoiceItem{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to remove invoice items: %w", err)
		}

		// Revert session statuses from 'invoice-draft' back to 'conducted'
		if err := tx.Model(&entities.Session{}).Where("id IN ?", req.RemoveSessionIDs).Update("status", "conducted").Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to revert session statuses: %w", err)
		}
	}

	// Handle removing extra efforts
	if len(req.RemoveExtraEffortIDs) > 0 {
		// Delete the invoice items
		if err := tx.Where("invoice_id = ? AND source_effort_id IN ?", invoiceID, req.RemoveExtraEffortIDs).Delete(&entities.InvoiceItem{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to remove effort items: %w", err)
		}

		// Revert extra effort statuses from 'invoice-draft' back to 'delivered'
		if err := tx.Model(&entities.ExtraEffort{}).Where("id IN ?", req.RemoveExtraEffortIDs).Update("billing_status", "delivered").Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to revert extra effort statuses: %w", err)
		}
	}

	// Handle adding sessions
	if len(req.AddSessionIDs) > 0 {
		// Get client ID from existing invoice
		var existingItem entities.InvoiceItem
		if err := tx.Where("invoice_id = ?", invoiceID).First(&existingItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to get existing item: %w", err)
		}

		// Verify sessions exist and are in 'conducted' status
		var sessions []entities.Session
		if err := tx.Where("id IN ? AND tenant_id = ? AND status = ?", req.AddSessionIDs, tenantID, "conducted").Find(&sessions).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to fetch sessions: %w", err)
		}

		if len(sessions) != len(req.AddSessionIDs) {
			tx.Rollback()
			return nil, errors.New("some sessions not found or not in conducted status")
		}

		// Create invoice items for new sessions
		for _, session := range sessions {
			unitPrice := 150.0 // TODO: Get from organization
			totalItemAmount := float64(session.NumberUnits) * unitPrice

			item := entities.InvoiceItem{
				InvoiceID:       invoiceID,
				ItemType:        "session",
				SessionID:       &session.ID,
				Description:     fmt.Sprintf("Therapiesitzung vom %s", session.OriginalDate.Format("02.01.2006")),
				NumberUnits:     float64(session.NumberUnits),
				UnitPrice:       unitPrice,
				TotalAmount:     totalItemAmount,
				UnitDurationMin: &session.DurationMin,
				IsEditable:      false,
				VATExempt:       taxSettings.VATExempt,
			}

			// Set VAT rate: 0 if exempt, otherwise use settings rate
			if item.VATExempt {
				item.VATRate = 0
				item.VATExemptionText = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
			} else {
				item.VATRate = taxSettings.VATRate
			}

			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create session item: %w", err)
			}
		}

		// Update session statuses to 'invoice-draft'
		if err := tx.Model(&entities.Session{}).Where("id IN ?", req.AddSessionIDs).Update("status", "invoice-draft").Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update session statuses: %w", err)
		}
	}

	// Handle adding extra efforts
	if len(req.AddExtraEffortIDs) > 0 {
		// Verify extra efforts exist and are in 'delivered' status
		var extraEfforts []entities.ExtraEffort
		if err := tx.Where("id IN ? AND tenant_id = ? AND billing_status = ? AND billable = ?",
			req.AddExtraEffortIDs, tenantID, "delivered", true).Find(&extraEfforts).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to fetch extra efforts: %w", err)
		}

		if len(extraEfforts) != len(req.AddExtraEffortIDs) {
			tx.Rollback()
			return nil, errors.New("some extra efforts not found or not billable/delivered")
		}

		// Create invoice items for new extra efforts
		for _, effort := range extraEfforts {
			unitPrice := 150.0 / 60.0 * float64(effort.DurationMin)
			totalItemAmount := unitPrice

			item := entities.InvoiceItem{
				InvoiceID:       invoiceID,
				ItemType:        "extra_effort",
				SourceEffortID:  &effort.ID,
				Description:     fmt.Sprintf("%s: %s", effort.EffortType, effort.Description),
				NumberUnits:     1,
				UnitPrice:       unitPrice,
				TotalAmount:     totalItemAmount,
				UnitDurationMin: &effort.DurationMin,
				IsEditable:      false,
				VATExempt:       taxSettings.VATExempt,
			}

			// Set VAT rate: 0 if exempt, otherwise use settings rate
			if item.VATExempt {
				item.VATRate = 0
				item.VATExemptionText = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
			} else {
				item.VATRate = taxSettings.VATRate
			}

			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create effort item: %w", err)
			}
		}

		// Update extra effort statuses to 'invoice-draft'
		if err := tx.Model(&entities.ExtraEffort{}).Where("id IN ?", req.AddExtraEffortIDs).Update("billing_status", "invoice-draft").Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update extra effort statuses: %w", err)
		}
	}

	// Handle updating custom line items (remove old ones and add new ones)
	if len(req.CustomLineItems) > 0 {
		// Remove existing custom items
		if err := tx.Where("invoice_id = ? AND item_type = ?", invoiceID, "custom").Delete(&entities.InvoiceItem{}).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to remove old custom items: %w", err)
		}

		// Add new custom items
		vatService := NewVATService()
		for _, customItem := range req.CustomLineItems {
			totalItemAmount := customItem.NumberUnits * customItem.UnitPrice

			item := entities.InvoiceItem{
				InvoiceID:   invoiceID,
				ItemType:    "custom",
				Description: customItem.Description,
				NumberUnits: customItem.NumberUnits,
				UnitPrice:   customItem.UnitPrice,
				TotalAmount: totalItemAmount,
				IsEditable:  true,
			}

			// Apply VAT category if specified
			if customItem.VATCategory != "" {
				if err := vatService.ApplyVATCategory(&item, VATCategory(customItem.VATCategory)); err == nil {
					// VAT category applied successfully
				} else {
					// Fall back to manual VAT settings if category is invalid
					item.VATRate = customItem.VATRate
					item.VATExempt = customItem.VATExempt
					item.VATExemptionText = customItem.VATExemptionText
				}
			} else {
				// Use manual VAT settings
				item.VATRate = customItem.VATRate
				item.VATExempt = customItem.VATExempt
				item.VATExemptionText = customItem.VATExemptionText
			}

			// Apply default VAT rate if none specified
			if !item.VATExempt && item.VATRate == 0 {
				item.VATRate = taxSettings.VATRate
			}

			// Ensure exempt items have 0% rate and exemption text
			if item.VATExempt {
				item.VATRate = 0
				if item.VATExemptionText == "" {
					item.VATExemptionText = "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG"
				}
			}

			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create custom item: %w", err)
			}
		}
	}

	// Recalculate totals
	var items []entities.InvoiceItem
	if err := tx.Where("invoice_id = ?", invoiceID).Find(&items).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to reload items for calculation: %w", err)
	}

	var subtotal float64
	var numberUnits int
	for _, item := range items {
		subtotal += item.TotalAmount
		if item.ItemType == "session" {
			numberUnits += int(item.NumberUnits)
		}
	}

	// Use VAT service for accurate tax calculation
	vatService := NewVATService()
	vatSummary := vatService.CalculateInvoiceVAT(items)
	taxAmount := vatSummary.TaxAmount
	totalAmount := subtotal + taxAmount

	// Update invoice totals
	if err := tx.Model(&invoice).Updates(map[string]interface{}{
		"number_units":    numberUnits,
		"subtotal_amount": subtotal,
		"tax_amount":      taxAmount,
		"total_amount":    totalAmount,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update invoice totals: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload invoice with all associations
	if err := s.db.Preload("InvoiceItems").Preload("Organization").First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// CancelDraftInvoice cancels a draft invoice and reverts item statuses
func (s *InvoiceService) CancelDraftInvoice(invoiceID, tenantID, userID uint) error {
	// Load existing invoice with items
	var invoice entities.Invoice
	if err := s.db.Preload("InvoiceItems").Where("id = ? AND tenant_id = ?", invoiceID, tenantID).First(&invoice).Error; err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Verify invoice is in draft status
	if invoice.Status != entities.InvoiceStatusDraft {
		return errors.New("can only cancel invoices in draft status")
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Collect all session IDs and extra effort IDs to revert statuses
	var sessionIDs []uint
	var effortIDs []uint

	for _, item := range invoice.InvoiceItems {
		if item.SessionID != nil {
			sessionIDs = append(sessionIDs, *item.SessionID)
		}
		if item.SourceEffortID != nil {
			effortIDs = append(effortIDs, *item.SourceEffortID)
		}
	}

	// Revert session statuses from 'invoice-draft' back to 'conducted'
	if len(sessionIDs) > 0 {
		if err := tx.Model(&entities.Session{}).
			Where("id IN ? AND status = ?", sessionIDs, "invoice-draft").
			Update("status", "conducted").Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to revert session statuses: %w", err)
		}
	}

	// Revert extra effort statuses from 'invoice-draft' back to 'delivered'
	if len(effortIDs) > 0 {
		if err := tx.Model(&entities.ExtraEffort{}).
			Where("id IN ? AND billing_status = ?", effortIDs, "invoice-draft").
			Update("billing_status", "delivered").Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to revert extra effort statuses: %w", err)
		}
	}

	// Delete client_invoices (hard delete so sessions can be re-invoiced)
	if err := tx.Unscoped().Where("invoice_id = ?", invoiceID).Delete(&entities.ClientInvoice{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete client invoices: %w", err)
	}

	// Delete all invoice items
	if err := tx.Where("invoice_id = ?", invoiceID).Delete(&entities.InvoiceItem{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete invoice items: %w", err)
	}

	// Soft delete the invoice and set cancelled_at timestamp
	now := time.Now()
	if err := tx.Model(&invoice).Updates(map[string]interface{}{
		"status":       entities.InvoiceStatusCancelled,
		"cancelled_at": &now,
		"deleted_at":   gorm.DeletedAt{Time: now, Valid: true},
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to cancel invoice: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getVATRate returns the VAT rate or default 19%
func getVATRate(rate *float64) float64 {
	if rate != nil {
		return *rate
	}
	return 19.00
}

// FinalizeInvoice finalizes a draft invoice by generating an invoice number and changing status to 'finalized'
func (s *InvoiceService) FinalizeInvoice(invoiceID, tenantID, userID uint, req *entities.FinalizeInvoiceRequest) (*entities.Invoice, error) {
	// Load existing invoice with items and client invoices
	var invoice entities.Invoice
	if err := s.db.Preload("InvoiceItems").
		Preload("ClientInvoices.Client.CostProvider").
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Precondition validation
	if invoice.Status != entities.InvoiceStatusDraft {
		return nil, errors.New("can only finalize invoices in draft status")
	}

	if len(invoice.InvoiceItems) == 0 {
		return nil, errors.New("invoice must have at least one line item")
	}

	// Validate VAT configuration using VAT service
	vatService := NewVATService()
	if err := vatService.ValidateVATConfiguration(invoice.InvoiceItems); err != nil {
		return nil, fmt.Errorf("VAT validation failed: %w", err)
	}

	// Validate government customer fields if applicable
	if len(invoice.ClientInvoices) > 0 {
		for _, clientInvoice := range invoice.ClientInvoices {
			if clientInvoice.Client != nil && clientInvoice.Client.CostProvider != nil {
				if clientInvoice.Client.CostProvider.IsGovernmentCustomer {
					if clientInvoice.Client.CostProvider.LeitwegID == "" {
						return nil, errors.New("government customer must have a Leitweg-ID")
					}
				}
			}
		}
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate invoice number WITHIN the transaction to ensure atomicity
	invoiceNumberService := NewInvoiceNumberService(tx)
	invoiceNumber, err := invoiceNumberService.GenerateInvoiceNumber(invoice.OrganizationID, tenantID, invoice.InvoiceDate)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Collect all session IDs and extra effort IDs to update statuses
	var sessionIDs []uint
	var effortIDs []uint

	for _, item := range invoice.InvoiceItems {
		if item.SessionID != nil {
			sessionIDs = append(sessionIDs, *item.SessionID)
		}
		if item.SourceEffortID != nil {
			effortIDs = append(effortIDs, *item.SourceEffortID)
		}
	}

	// Update session statuses from 'invoice-draft' to 'billed'
	if len(sessionIDs) > 0 {
		if err := tx.Model(&entities.Session{}).
			Where("id IN ? AND status = ?", sessionIDs, "invoice-draft").
			Update("status", "billed").Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update session statuses: %w", err)
		}
	}

	// Update extra effort statuses from 'invoice-draft' to 'billed'
	if len(effortIDs) > 0 {
		if err := tx.Model(&entities.ExtraEffort{}).
			Where("id IN ? AND billing_status = ?", effortIDs, "invoice-draft").
			Update("billing_status", "billed").Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update extra effort statuses: %w", err)
		}
	}

	// Update invoice: set number, status, finalized_at timestamp, and customer fields if provided
	now := time.Now()
	updateFields := map[string]interface{}{
		"invoice_number": invoiceNumber,
		"status":         entities.InvoiceStatusFinalized,
		"finalized_at":   &now,
	}

	// Update customer fields if provided in request
	if req != nil {
		if req.CustomerName != "" {
			updateFields["customer_name"] = req.CustomerName
		}
		if req.CustomerAddress != "" {
			updateFields["customer_address"] = req.CustomerAddress
		}
		if req.CustomerAddressExt != "" {
			updateFields["customer_address_ext"] = req.CustomerAddressExt
		}
		if req.CustomerZip != "" {
			updateFields["customer_zip"] = req.CustomerZip
		}
		if req.CustomerCity != "" {
			updateFields["customer_city"] = req.CustomerCity
		}
		if req.CustomerCountry != "" {
			updateFields["customer_country"] = req.CustomerCountry
		}
		if req.CustomerContactPerson != "" {
			updateFields["customer_contact_person"] = req.CustomerContactPerson
		}
		if req.CustomerDepartment != "" {
			updateFields["customer_department"] = req.CustomerDepartment
		}
		if req.CustomerEmail != "" {
			updateFields["customer_email"] = req.CustomerEmail
		}
	}

	if err := tx.Model(&invoice).Updates(updateFields).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to finalize invoice: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload invoice with all associations
	if err := s.db.Preload("InvoiceItems").
		Preload("Organization").
		Preload("ClientInvoices.Client").
		Preload("ClientInvoices.CostProvider").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	// Generate and store final PDF in MinIO (if PDF service is available)
	if s.pdfService != nil {
		ctx := context.Background()
		storageKey, err := s.pdfService.StoreFinalPDFToMinIO(ctx, &invoice)
		if err != nil {
			// Log error but don't fail the finalization
			fmt.Printf("Warning: Failed to store final PDF in MinIO: %v\n", err)
		} else {
			// Update invoice with document storage key
			// Note: We could store this in a new field like document_storage_key
			// For now, we'll just log it
			fmt.Printf("Final PDF stored in MinIO: %s\n", storageKey)
		}
	}

	return &invoice, nil
}

// MarkAsSent marks a finalized invoice as sent
func (s *InvoiceService) MarkAsSent(invoiceID, tenantID uint) (*entities.Invoice, error) {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.Where("id = ? AND tenant_id = ?", invoiceID, tenantID).First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Validate status - can only mark finalized invoices as sent
	if invoice.Status != entities.InvoiceStatusFinalized {
		return nil, errors.New("can only mark finalized invoices as sent")
	}

	// Update invoice status to sent and set sent_at timestamp
	now := time.Now()
	if err := s.db.Model(&invoice).Updates(map[string]interface{}{
		"status":  entities.InvoiceStatusSent,
		"sent_at": &now,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to mark invoice as sent: %w", err)
	}

	// Reload invoice with all associations
	if err := s.db.Preload("InvoiceItems").
		Preload("Organization").
		Preload("ClientInvoices.Client").
		Preload("ClientInvoices.CostProvider").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// SendInvoiceEmail sends an invoice via email
func (s *InvoiceService) SendInvoiceEmail(invoiceID, tenantID, userID uint) error {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.Preload("ClientInvoices.Client.CostProvider").
		Preload("Organization").
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Precondition: invoice must be sent, paid, or overdue
	if invoice.Status != entities.InvoiceStatusSent &&
		invoice.Status != entities.InvoiceStatusPaid &&
		invoice.Status != entities.InvoiceStatusOverdue {
		return errors.New("can only send email for invoices in sent, paid, or overdue status")
	}

	// TODO: Generate PDF and send email
	// For now, we'll just set the sent_at timestamp
	// In a future phase, this will:
	// 1. Generate PDF using template service
	// 2. Render email template with invoice data
	// 3. Attach PDF to email
	// 4. Send via email service

	// Update sent_at timestamp
	now := time.Now()
	if err := s.db.Model(&invoice).Update("sent_at", &now).Error; err != nil {
		return fmt.Errorf("failed to update sent_at: %w", err)
	}

	return nil
}

// MarkInvoiceAsPaid marks an invoice as paid
func (s *InvoiceService) MarkInvoiceAsPaid(invoiceID, tenantID, userID uint, paymentDate *time.Time, paymentReference string) (*entities.Invoice, error) {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.Where("id = ? AND tenant_id = ?", invoiceID, tenantID).First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Precondition: invoice must be sent or overdue
	if invoice.Status != entities.InvoiceStatusSent && invoice.Status != entities.InvoiceStatusOverdue {
		return nil, errors.New("can only mark invoices as paid if they are in sent or overdue status")
	}

	// Default payment date to today if not provided
	var payDate time.Time
	if paymentDate != nil {
		payDate = *paymentDate
	} else {
		payDate = time.Now()
	}

	// Update invoice status and payment date
	updates := map[string]interface{}{
		"status":       entities.InvoiceStatusPaid,
		"payment_date": payDate,
	}

	if err := s.db.Model(&invoice).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	// Reload invoice with all associations
	if err := s.db.Preload("InvoiceItems").
		Preload("Organization").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// MarkInvoiceAsOverdue marks an invoice as overdue if payment is past due
func (s *InvoiceService) MarkInvoiceAsOverdue(invoiceID, tenantID, userID uint) (*entities.Invoice, error) {
	// Load existing invoice with organization to get payment terms
	var invoice entities.Invoice
	if err := s.db.Preload("Organization").
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Precondition: invoice must be sent
	if invoice.Status != entities.InvoiceStatusSent {
		return nil, errors.New("can only mark invoices as overdue if they are in sent status")
	}

	// Get organization payment terms
	var org struct {
		PaymentDueDays int
	}
	if err := s.db.Table("organizations").
		Select("payment_due_days").
		Where("id = ?", invoice.OrganizationID).
		First(&org).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}

	// Calculate due date
	dueDate := invoice.InvoiceDate.AddDate(0, 0, org.PaymentDueDays)
	today := time.Now()

	// Check if invoice is actually overdue
	if today.Before(dueDate) || today.Equal(dueDate) {
		return nil, fmt.Errorf("invoice is not yet overdue (due date: %s)", dueDate.Format("2006-01-02"))
	}

	// Update invoice status to overdue
	if err := s.db.Model(&invoice).Update("status", entities.InvoiceStatusOverdue).Error; err != nil {
		return nil, fmt.Errorf("failed to mark invoice as overdue: %w", err)
	}

	// Reload invoice with all associations
	if err := s.db.Preload("InvoiceItems").
		Preload("Organization").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// SendReminder sends a payment reminder for an overdue invoice
func (s *InvoiceService) SendReminder(invoiceID, tenantID, userID uint) error {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.Where("id = ? AND tenant_id = ?", invoiceID, tenantID).First(&invoice).Error; err != nil {
		return fmt.Errorf("invoice not found: %w", err)
	}

	// Precondition: invoice must be overdue
	if invoice.Status != entities.InvoiceStatusOverdue {
		return errors.New("can only send reminders for invoices in overdue status")
	}

	// Increment reminder counter and update timestamps
	now := time.Now()
	updates := map[string]interface{}{
		"num_reminders":    invoice.NumReminders + 1,
		"latest_reminder":  &now,
		"reminder_sent_at": &now,
	}

	if err := s.db.Model(&invoice).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update reminder information: %w", err)
	}

	// TODO: Generate reminder PDF and send email
	// For now, we'll just update the counters
	// In a future phase, this will:
	// 1. Select template based on num_reminders (reminder_1, reminder_2, reminder_3)
	// 2. Generate reminder PDF with overdue details
	// 3. Send reminder email
	// 4. Create audit log entry

	return nil
}

// CreateCreditNote creates a credit note for an existing invoice
func (s *InvoiceService) CreateCreditNote(originalInvoiceID, tenantID, userID uint, req entities.CreateCreditNoteRequest) (*entities.Invoice, error) {
	// Load original invoice with items
	var originalInvoice entities.Invoice
	if err := s.db.Preload("InvoiceItems").
		Preload("Organization").
		Where("id = ? AND tenant_id = ?", originalInvoiceID, tenantID).
		First(&originalInvoice).Error; err != nil {
		return nil, fmt.Errorf("original invoice not found: %w", err)
	}

	// Precondition: original invoice must be sent, paid, or overdue
	if originalInvoice.Status != entities.InvoiceStatusSent &&
		originalInvoice.Status != entities.InvoiceStatusPaid &&
		originalInvoice.Status != entities.InvoiceStatusOverdue {
		return nil, errors.New("can only create credit notes for invoices in sent, paid, or overdue status")
	}

	// Verify all line items belong to the original invoice
	var itemsToCredit []entities.InvoiceItem
	if err := s.db.Where("id IN ? AND invoice_id = ?", req.LineItemIDs, originalInvoiceID).
		Find(&itemsToCredit).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch line items: %w", err)
	}

	if len(itemsToCredit) != len(req.LineItemIDs) {
		return nil, errors.New("some line items not found or do not belong to this invoice")
	}

	// Default credit date to today if not provided
	var creditDate time.Time
	if req.CreditDate != nil {
		creditDate = *req.CreditDate
	} else {
		creditDate = time.Now()
	}

	// Generate invoice number for the credit note
	invoiceNumberService := NewInvoiceNumberService(s.db)
	invoiceNumber, err := invoiceNumberService.GenerateInvoiceNumber(originalInvoice.OrganizationID, tenantID, creditDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate credit note number: %w", err)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create credit note invoice
	creditNote := entities.Invoice{
		TenantID:              tenantID,
		UserID:                userID,
		OrganizationID:        originalInvoice.OrganizationID,
		InvoiceDate:           creditDate,
		InvoiceNumber:         invoiceNumber,
		Status:                entities.InvoiceStatusSent, // Credit notes are immediately finalized
		IsCreditNote:          true,
		CreditNoteReferenceID: &originalInvoiceID,
		NumberUnits:           0,
		SumAmount:             0,
		TaxAmount:             0,
		TotalAmount:           0,
	}

	// Set finalized_at since credit notes are immediately sent
	now := time.Now()
	creditNote.FinalizedAt = &now

	if err := tx.Create(&creditNote).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create credit note: %w", err)
	}

	// Create negative invoice items for each credited line item
	var subtotal float64
	var totalUnits int

	for _, originalItem := range itemsToCredit {
		creditItem := entities.InvoiceItem{
			InvoiceID:        creditNote.ID,
			ItemType:         originalItem.ItemType,
			SessionID:        originalItem.SessionID,
			SourceEffortID:   originalItem.SourceEffortID,
			Description:      fmt.Sprintf("Gutschrift: %s (Grund: %s)", originalItem.Description, req.Reason),
			NumberUnits:      -originalItem.NumberUnits, // Negative quantity
			UnitPrice:        originalItem.UnitPrice,
			TotalAmount:      -originalItem.TotalAmount, // Negative total
			UnitDurationMin:  originalItem.UnitDurationMin,
			IsEditable:       false,
			VATRate:          originalItem.VATRate,
			VATExempt:        originalItem.VATExempt,
			VATExemptionText: originalItem.VATExemptionText,
		}

		if err := tx.Create(&creditItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create credit note item: %w", err)
		}

		subtotal += creditItem.TotalAmount
		if creditItem.ItemType == "session" {
			totalUnits += int(creditItem.NumberUnits)
		}
	}

	// Calculate tax amount (negative)
	var taxAmount float64
	if !itemsToCredit[0].VATExempt && len(itemsToCredit) > 0 {
		taxRate := itemsToCredit[0].VATRate / 100.0
		taxAmount = subtotal * taxRate
	}

	totalAmount := subtotal + taxAmount

	// Update credit note totals
	if err := tx.Model(&creditNote).Updates(map[string]interface{}{
		"number_units":    totalUnits,
		"subtotal_amount": subtotal,
		"tax_amount":      taxAmount,
		"total_amount":    totalAmount,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update credit note totals: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload credit note with all associations
	if err := s.db.Preload("InvoiceItems").
		Preload("Organization").
		Preload("ClientInvoices.Client").
		Preload("ClientInvoices.CostProvider").
		First(&creditNote, creditNote.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload credit note: %w", err)
	}

	// Generate and store credit note PDF in MinIO (if PDF service is available)
	if s.pdfService != nil {
		ctx := context.Background()
		storageKey, err := s.pdfService.StoreCreditNotePDFToMinIO(ctx, &creditNote, originalInvoice.InvoiceNumber, req.Reason)
		if err != nil {
			// Log error but don't fail the credit note creation
			fmt.Printf("Warning: Failed to store credit note PDF in MinIO: %v\n", err)
		} else {
			fmt.Printf("Credit note PDF stored in MinIO: %s\n", storageKey)
		}
	}

	return &creditNote, nil
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
		sessionIDPtr := sessionID
		clientInvoice := entities.ClientInvoice{
			InvoiceID:      invoice.ID,
			ClientID:       req.ClientID,
			CostProviderID: *client.CostProviderID,
			SessionID:      &sessionIDPtr,
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

			sessionIDPtr := sessionID
			clientInvoice := entities.ClientInvoice{
				InvoiceID:      invoice.ID,
				ClientID:       client.ID,
				CostProviderID: costProviderID,
				SessionID:      &sessionIDPtr,
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
		if len(req.SessionIDs) > 0 {
			return nil, errors.New("cannot update both status and invoice items in the same request")
		}

		invoice.Status = *req.Status

		// If status is being set to paid, set paid date
		if *req.Status == entities.InvoiceStatusPaid && invoice.PayedDate == nil {
			now := time.Now()
			invoice.PayedDate = &now
		}

		// If status is being set to reminder, increment reminder count
		if *req.Status == entities.InvoiceStatusOverdue {
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
	if len(req.SessionIDs) > 0 {
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
			sessionIDPtr := sessionID
			clientInvoice := entities.ClientInvoice{
				InvoiceID:      id,
				ClientID:       firstClientInvoice.ClientID,
				CostProviderID: firstClientInvoice.CostProviderID,
				SessionID:      &sessionIDPtr,
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
// Also reverts sessions and extra efforts back to their original state
func (s *InvoiceService) DeleteInvoice(id, tenantID, userID uint) error {
	// Verify invoice exists and belongs to user/tenant
	var invoice entities.Invoice
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invoice not found")
		}
		return fmt.Errorf("failed to fetch invoice: %w", err)
	}

	// Start transaction to ensure atomicity
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get all client_invoices for this invoice to revert session/effort states
	var clientInvoices []entities.ClientInvoice
	if err := tx.Where("invoice_id = ?", id).Find(&clientInvoices).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch client invoices: %w", err)
	}

	// Collect session IDs and extra effort IDs to revert
	sessionIDs := make([]uint, 0)
	extraEffortIDs := make([]uint, 0)

	for _, ci := range clientInvoices {
		if ci.SessionID != nil && *ci.SessionID > 0 {
			sessionIDs = append(sessionIDs, *ci.SessionID)
		}
		if ci.ExtraEffortID != nil {
			extraEffortIDs = append(extraEffortIDs, *ci.ExtraEffortID)
		}
	}

	// Revert sessions back to conducted status (so they appear in billable list again)
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

	// Delete client_invoices (hard delete so sessions can be re-invoiced)
	if err := tx.Unscoped().Where("invoice_id = ?", id).Delete(&entities.ClientInvoice{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete client invoices: %w", err)
	}

	// Delete invoice items (soft delete)
	if err := tx.Where("invoice_id = ?", id).Delete(&entities.InvoiceItem{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete invoice items: %w", err)
	}

	// Delete invoice
	if err := tx.Delete(&invoice).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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
		if ci.SessionID != nil && *ci.SessionID > 0 {
			sessionIDs = append(sessionIDs, *ci.SessionID)
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

	// Delete client_invoices (hard delete so sessions can be re-invoiced)
	if err := tx.Unscoped().Where("invoice_id = ?", id).Delete(&entities.ClientInvoice{}).Error; err != nil {
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
	// Get all clients with unbilled conducted sessions OR delivered extra efforts
	var clients []entities.Client

	// Use a subquery to find clients with either unbilled sessions or delivered extra efforts
	err := s.db.
		Preload("CostProvider").
		Where("clients.tenant_id = ? AND ("+
			"clients.id IN ("+
			"SELECT DISTINCT client_id FROM sessions "+
			"WHERE tenant_id = ? AND status = 'conducted'"+
			") OR clients.id IN ("+
			"SELECT DISTINCT client_id FROM extra_efforts "+
			"WHERE tenant_id = ? AND billing_status = 'delivered' AND billable = true"+
			")"+
			")",
			tenantID, tenantID, tenantID,
		).
		Group("clients.id").
		Find(&clients).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch clients with unbilled sessions or extra efforts: %w", err)
	}

	// For each client, fetch their unbilled conducted sessions and delivered extra efforts
	result := make([]entities.ClientWithUnbilledSessionsResponse, 0, len(clients))
	for _, client := range clients {
		var sessions []entities.Session
		err := s.db.
			Where("tenant_id = ? AND client_id = ? AND status = ?",
				tenantID,
				client.ID,
				"conducted",
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

		// Fetch delivered extra efforts for this client
		var extraEfforts []entities.ExtraEffort
		err = s.db.
			Where("tenant_id = ? AND client_id = ? AND billing_status = ? AND billable = ?",
				tenantID,
				client.ID,
				"delivered",
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

// GetCostProviderByID retrieves a cost provider by ID for the given tenant
func (s *InvoiceService) GetCostProviderByID(costProviderID, tenantID uint) (*models.CostProvider, error) {
	var costProvider models.CostProvider
	if err := s.db.Where("id = ? AND tenant_id = ?", costProviderID, tenantID).First(&costProvider).Error; err != nil {
		return nil, fmt.Errorf("cost provider not found: %w", err)
	}
	return &costProvider, nil
}

// GetOrganizationByID retrieves an organization by ID for the given tenant
func (s *InvoiceService) GetOrganizationByID(organizationID, tenantID uint) (*baseAPI.Organization, error) {
	var organization baseAPI.Organization
	if err := s.db.Where("id = ? AND tenant_id = ?", organizationID, tenantID).First(&organization).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	return &organization, nil
}
