package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ae-base-server/pkg/settings/manager"
	"github.com/unburdy/invoice-module/entities"
	"gorm.io/gorm"
)

// InvoiceService handles invoice operations
type InvoiceService struct {
	db              *gorm.DB
	settingsManager *manager.SettingsManager
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(db *gorm.DB, settingsManager *manager.SettingsManager) *InvoiceService {
	return &InvoiceService{
		db:              db,
		settingsManager: settingsManager,
	}
}

// CreateInvoice creates a new invoice
func (s *InvoiceService) CreateInvoice(ctx context.Context, tenantID, userID uint, req *entities.CreateInvoiceRequest) (*entities.Invoice, error) {
	// Calculate amounts
	var subtotal, totalTax float64
	items := make([]entities.InvoiceItem, len(req.Items))

	for i, item := range req.Items {
		itemAmount := item.Quantity * item.UnitPrice
		itemTax := itemAmount * (item.TaxRate / 100)

		items[i] = entities.InvoiceItem{
			Position:    i + 1,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TaxRate:     item.TaxRate,
			Amount:      itemAmount,
		}

		subtotal += itemAmount
		totalTax += itemTax
	}

	invoice := &entities.Invoice{
		TenantID:        tenantID,
		OrganizationID:  req.OrganizationID,
		UserID:          userID,
		InvoiceNumber:   req.InvoiceNumber,
		InvoiceDate:     req.InvoiceDate,
		DueDate:         req.DueDate,
		Status:          entities.InvoiceStatusDraft,
		CustomerName:    req.CustomerName,
		CustomerAddress: req.CustomerAddress,
		CustomerEmail:   req.CustomerEmail,
		CustomerTaxID:   req.CustomerTaxID,
		SubtotalAmount:  subtotal,
		TaxRate:         req.TaxRate,
		TaxAmount:       totalTax,
		TotalAmount:     subtotal + totalTax,
		Currency:        req.Currency,
		PaymentTerms:    req.PaymentTerms,
		PaymentMethod:   req.PaymentMethod,
		Notes:           req.Notes,
		InternalNote:    req.InternalNote,
		Items:           items,
	}

	if invoice.Currency == "" {
		invoice.Currency = "EUR"
	}

	if err := s.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// CreateInvoiceWithAutoNumber creates a new invoice with auto-generated invoice number
func (s *InvoiceService) CreateInvoiceWithAutoNumber(ctx context.Context, tenantID, userID uint, req *entities.CreateInvoiceRequest) (*entities.Invoice, error) {
	// Generate invoice number using settings
	invoiceNumber, err := s.generateInvoiceNumber(ctx, tenantID, req.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Set the generated invoice number
	req.InvoiceNumber = invoiceNumber

	// Use the existing CreateInvoice method
	return s.CreateInvoice(ctx, tenantID, userID, req)
}

// generateInvoiceNumber generates an invoice number using the settings system
func (s *InvoiceService) generateInvoiceNumber(ctx context.Context, tenantID, organizationID uint) (string, error) {
	// Get invoice settings accessor
	invoiceAccessor, err := s.settingsManager.GetModuleAccessor("invoice")
	if err != nil {
		return "", fmt.Errorf("failed to get invoice settings: %w", err)
	}

	// Get invoice prefix from settings
	prefix, err := invoiceAccessor.GetString("invoice_prefix")
	if err != nil {
		prefix = "INV" // Fallback default
	}

	// Get and increment next invoice number
	nextNumber, err := invoiceAccessor.GetInt("next_invoice_number")
	if err != nil {
		nextNumber = 1000 // Fallback default
	}

	// Increment the counter for next use
	if err := invoiceAccessor.SetInt("next_invoice_number", nextNumber+1); err != nil {
		return "", fmt.Errorf("failed to increment invoice number: %w", err)
	}

	// Generate invoice number: PREFIX-YYYY-NNNN (e.g., INV-2025-1000)
	now := time.Now()
	invoiceNumber := fmt.Sprintf("%s-%04d-%04d", prefix, now.Year(), nextNumber)

	return invoiceNumber, nil
}

// GetInvoiceSettings returns current invoice settings for an organization
func (s *InvoiceService) GetInvoiceSettings(ctx context.Context, tenantID, organizationID uint) (map[string]interface{}, error) {
	// Get invoice settings accessor
	invoiceAccessor, err := s.settingsManager.GetModuleAccessor("invoice")
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice settings: %w", err)
	}

	settings := make(map[string]interface{})

	// Get common invoice settings
	if prefix, err := invoiceAccessor.GetString("invoice_prefix"); err == nil {
		settings["invoice_prefix"] = prefix
	}

	if nextNumber, err := invoiceAccessor.GetInt("next_invoice_number"); err == nil {
		settings["next_invoice_number"] = nextNumber
	}

	if paymentTerms, err := invoiceAccessor.GetInt("payment_terms_days"); err == nil {
		settings["payment_terms_days"] = paymentTerms
	}

	if autoSend, err := invoiceAccessor.GetBool("auto_send_invoice"); err == nil {
		settings["auto_send_invoice"] = autoSend
	}

	if template, err := invoiceAccessor.GetString("invoice_template"); err == nil {
		settings["invoice_template"] = template
	}

	if lateFee, err := invoiceAccessor.GetFloat("late_fee_percentage"); err == nil {
		settings["late_fee_percentage"] = lateFee
	}

	if footer, err := invoiceAccessor.GetString("invoice_footer"); err == nil {
		settings["invoice_footer"] = footer
	}

	// Get tax settings as JSON
	var taxSettings map[string]interface{}
	if err := invoiceAccessor.GetJSON("tax_settings", &taxSettings); err == nil {
		settings["tax_settings"] = taxSettings
	}

	return settings, nil
}

// GetInvoice retrieves an invoice by ID
func (s *InvoiceService) GetInvoice(ctx context.Context, tenantID, invoiceID uint) (*entities.Invoice, error) {
	var invoice entities.Invoice
	err := s.db.WithContext(ctx).
		Preload("Items").
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error

	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

// ListInvoices lists invoices with filters
func (s *InvoiceService) ListInvoices(ctx context.Context, tenantID uint, organizationID *uint, status *entities.InvoiceStatus, page, pageSize int) ([]entities.Invoice, int64, error) {
	var invoices []entities.Invoice
	var total int64

	query := s.db.WithContext(ctx).Model(&entities.Invoice{}).Where("tenant_id = ?", tenantID)

	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Items").
		Order("invoice_date DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

// UpdateInvoice updates an invoice
func (s *InvoiceService) UpdateInvoice(ctx context.Context, tenantID, invoiceID uint, req *entities.UpdateInvoiceRequest) (*entities.Invoice, error) {
	invoice, err := s.GetInvoice(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Status != nil {
		invoice.Status = *req.Status
		if *req.Status == entities.InvoiceStatusPaid && invoice.PaymentDate == nil {
			now := time.Now()
			invoice.PaymentDate = &now
		}
	}

	if req.DueDate != nil {
		invoice.DueDate = req.DueDate
	}

	if req.CustomerName != nil {
		invoice.CustomerName = *req.CustomerName
	}

	if req.CustomerAddress != nil {
		invoice.CustomerAddress = *req.CustomerAddress
	}

	if req.CustomerEmail != nil {
		invoice.CustomerEmail = *req.CustomerEmail
	}

	if req.PaymentTerms != nil {
		invoice.PaymentTerms = *req.PaymentTerms
	}

	if req.PaymentMethod != nil {
		invoice.PaymentMethod = *req.PaymentMethod
	}

	if req.PaymentDate != nil {
		invoice.PaymentDate = req.PaymentDate
	}

	if req.Notes != nil {
		invoice.Notes = *req.Notes
	}

	if req.InternalNote != nil {
		invoice.InternalNote = *req.InternalNote
	}

	// Update items if provided
	if req.Items != nil {
		// Delete existing items
		if err := s.db.WithContext(ctx).Where("invoice_id = ?", invoiceID).Delete(&entities.InvoiceItem{}).Error; err != nil {
			return nil, fmt.Errorf("failed to delete old items: %w", err)
		}

		// Recalculate amounts
		var subtotal, totalTax float64
		items := make([]entities.InvoiceItem, len(*req.Items))

		for i, item := range *req.Items {
			itemAmount := item.Quantity * item.UnitPrice
			itemTax := itemAmount * (item.TaxRate / 100)

			items[i] = entities.InvoiceItem{
				InvoiceID:   invoiceID,
				Position:    i + 1,
				Description: item.Description,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				TaxRate:     item.TaxRate,
				Amount:      itemAmount,
			}

			subtotal += itemAmount
			totalTax += itemTax
		}

		invoice.SubtotalAmount = subtotal
		invoice.TaxAmount = totalTax
		invoice.TotalAmount = subtotal + totalTax
		invoice.Items = items
	}

	if err := s.db.WithContext(ctx).Save(invoice).Error; err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	return invoice, nil
}

// DeleteInvoice soft deletes an invoice
func (s *InvoiceService) DeleteInvoice(ctx context.Context, tenantID, invoiceID uint) error {
	return s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		Delete(&entities.Invoice{}).Error
}

// MarkAsPaid marks an invoice as paid
func (s *InvoiceService) MarkAsPaid(ctx context.Context, tenantID, invoiceID uint, paymentDate time.Time) error {
	return s.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		Updates(map[string]interface{}{
			"status":       entities.InvoiceStatusPaid,
			"payment_date": paymentDate,
		}).Error
}

// LinkDocument links a generated PDF document to an invoice
func (s *InvoiceService) LinkDocument(ctx context.Context, tenantID, invoiceID, documentID uint) error {
	return s.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		Update("document_id", documentID).Error
}
