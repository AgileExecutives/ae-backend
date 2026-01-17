package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	invoiceNumberService "github.com/ae-base-server/modules/invoice_number/services"
	"github.com/unburdy/invoice-module/entities"
)

// FinalizeInvoice finalizes a draft invoice by generating an invoice number and changing status to finalized
func (s *InvoiceService) FinalizeInvoice(ctx context.Context, tenantID, invoiceID, userID uint) (*entities.Invoice, error) {
	// Load existing invoice with items
	var invoice entities.Invoice
	if err := s.db.WithContext(ctx).
		Preload("Items").
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Precondition validation
	if invoice.Status != entities.InvoiceStatusDraft {
		return nil, errors.New("can only finalize invoices in draft status")
	}

	if len(invoice.Items) == 0 {
		return nil, errors.New("invoice must have at least one line item")
	}

	// Validate VAT configuration
	if err := s.validateVAT(&invoice); err != nil {
		return nil, fmt.Errorf("VAT validation failed: %w", err)
	}

	// Validate business rules
	if err := s.validateInvoice(&invoice); err != nil {
		return nil, fmt.Errorf("invoice validation failed: %w", err)
	}

	// Begin transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate invoice number if still has DRAFT prefix
	var invoiceNumber string
	if invoice.InvoiceNumber == "" || len(invoice.InvoiceNumber) > 5 && invoice.InvoiceNumber[:6] == "DRAFT-" {
		invoiceNumberSvc := invoiceNumberService.NewInvoiceNumberService(tx)
		number, err := invoiceNumberSvc.GenerateInvoiceNumber(
			invoice.OrganizationID,
			tenantID,
			invoice.InvoiceDate,
		)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to generate invoice number: %w", err)
		}
		invoiceNumber = number
	} else {
		invoiceNumber = invoice.InvoiceNumber
	}

	// Update invoice: set number, status, and finalized_at timestamp
	now := time.Now()
	if err := tx.Model(&invoice).Updates(map[string]interface{}{
		"invoice_number": invoiceNumber,
		"status":         entities.InvoiceStatusFinalized,
		"finalized_at":   &now,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to finalize invoice: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload invoice with all associations
	if err := s.db.WithContext(ctx).
		Preload("Items").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// MarkAsSent marks a finalized invoice as sent
func (s *InvoiceService) MarkAsSent(ctx context.Context, tenantID, invoiceID uint) (*entities.Invoice, error) {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Validate status - can only mark finalized invoices as sent
	if invoice.Status != entities.InvoiceStatusFinalized {
		return nil, errors.New("can only mark finalized invoices as sent")
	}

	// Update invoice status to sent and set sent_at timestamp
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&invoice).Updates(map[string]interface{}{
		"status":        entities.InvoiceStatusSent,
		"email_sent_at": &now,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to mark invoice as sent: %w", err)
	}

	// Reload invoice
	if err := s.db.WithContext(ctx).
		Preload("Items").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// MarkAsPaidWithAmount marks an invoice as paid with payment details
func (s *InvoiceService) MarkAsPaidWithAmount(ctx context.Context, tenantID, invoiceID uint, paymentDate time.Time, paymentMethod string) (*entities.Invoice, error) {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Validate status - can mark sent or overdue invoices as paid
	if invoice.Status != entities.InvoiceStatusSent && invoice.Status != entities.InvoiceStatusOverdue {
		return nil, errors.New("can only mark sent or overdue invoices as paid")
	}

	// Update invoice status to paid
	updates := map[string]interface{}{
		"status":       entities.InvoiceStatusPaid,
		"payment_date": paymentDate,
	}

	if paymentMethod != "" {
		updates["payment_method"] = paymentMethod
	}

	if err := s.db.WithContext(ctx).Model(&invoice).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	// Reload invoice
	if err := s.db.WithContext(ctx).
		Preload("Items").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// SendReminder sends a payment reminder and updates the reminder counter
func (s *InvoiceService) SendReminder(ctx context.Context, tenantID, invoiceID uint) (*entities.Invoice, error) {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Validate status - can only send reminders for sent or overdue invoices
	if invoice.Status != entities.InvoiceStatusSent && invoice.Status != entities.InvoiceStatusOverdue {
		return nil, errors.New("can only send reminders for sent or overdue invoices")
	}

	// Check if invoice is overdue
	now := time.Now()
	var newStatus entities.InvoiceStatus
	if invoice.DueDate != nil && invoice.DueDate.Before(now) {
		newStatus = entities.InvoiceStatusOverdue
	} else {
		newStatus = invoice.Status
	}

	// Update reminder counter and timestamp
	if err := s.db.WithContext(ctx).Model(&invoice).Updates(map[string]interface{}{
		"num_reminders":    invoice.NumReminders + 1,
		"reminder_sent_at": &now,
		"status":           newStatus,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update reminder: %w", err)
	}

	// Reload invoice
	if err := s.db.WithContext(ctx).
		Preload("Items").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// CancelInvoice cancels an invoice with an optional reason
func (s *InvoiceService) CancelInvoice(ctx context.Context, tenantID, invoiceID uint, reason string) (*entities.Invoice, error) {
	// Load existing invoice
	var invoice entities.Invoice
	if err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", invoiceID, tenantID).
		First(&invoice).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}

	// Cannot cancel already paid invoices
	if invoice.Status == entities.InvoiceStatusPaid {
		return nil, errors.New("cannot cancel paid invoices")
	}

	// Cannot cancel already cancelled invoices
	if invoice.Status == entities.InvoiceStatusCancelled {
		return nil, errors.New("invoice is already cancelled")
	}

	// Update invoice status to cancelled
	now := time.Now()
	updates := map[string]interface{}{
		"status":       entities.InvoiceStatusCancelled,
		"cancelled_at": &now,
	}

	if reason != "" {
		updates["cancellation_reason"] = reason
	}

	if err := s.db.WithContext(ctx).Model(&invoice).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to cancel invoice: %w", err)
	}

	// Reload invoice
	if err := s.db.WithContext(ctx).
		Preload("Items").
		First(&invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return &invoice, nil
}

// CheckOverdueInvoices checks and updates overdue invoices
func (s *InvoiceService) CheckOverdueInvoices(ctx context.Context) error {
	now := time.Now()

	// Update all sent invoices that are past their due date to overdue status
	return s.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("status = ? AND due_date < ?", entities.InvoiceStatusSent, now).
		Update("status", entities.InvoiceStatusOverdue).Error
}

// validateVAT validates VAT configuration on invoice items
func (s *InvoiceService) validateVAT(invoice *entities.Invoice) error {
	for _, item := range invoice.Items {
		// If VAT exempt, must have exemption text
		if item.VATExempt && item.VATExemptionText == "" {
			return fmt.Errorf("VAT exempt items must have exemption text (item: %s)", item.Description)
		}

		// If not exempt, VAT rate should be set
		if !item.VATExempt && item.VATRate == 0 {
			return fmt.Errorf("non-exempt items must have VAT rate configured (item: %s)", item.Description)
		}

		// VAT rate should be reasonable (0-100%)
		if item.VATRate < 0 || item.VATRate > 100 {
			return fmt.Errorf("invalid VAT rate %v%% for item: %s", item.VATRate, item.Description)
		}
	}

	return nil
}

// validateInvoice validates business rules for an invoice
func (s *InvoiceService) validateInvoice(invoice *entities.Invoice) error {
	// Must have customer name
	if invoice.CustomerName == "" {
		return errors.New("customer name is required")
	}

	// Must have at least one item
	if len(invoice.Items) == 0 {
		return errors.New("invoice must have at least one line item")
	}

	// Due date should be after invoice date if set
	if invoice.DueDate != nil && invoice.DueDate.Before(invoice.InvoiceDate) {
		return errors.New("due date must be after invoice date")
	}

	// Total amount should be positive
	if invoice.TotalAmount <= 0 {
		return errors.New("invoice total must be positive")
	}

	// Validate performance period if set
	if invoice.PerformancePeriodStart != nil && invoice.PerformancePeriodEnd != nil {
		if invoice.PerformancePeriodEnd.Before(*invoice.PerformancePeriodStart) {
			return errors.New("performance period end must be after start")
		}
	}

	return nil
}
