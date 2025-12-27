package services

import (
	"context"
	"fmt"
	"time"

	"github.com/unburdy/invoice-module/entities"
	"gorm.io/gorm"
)

// InvoiceService handles invoice operations
type InvoiceService struct {
	db *gorm.DB
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(db *gorm.DB) *InvoiceService {
	return &InvoiceService{db: db}
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
