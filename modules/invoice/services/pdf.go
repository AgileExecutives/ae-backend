package services

import (
	"context"

	"github.com/unburdy/invoice-module/entities"
)

// PDFService defines the interface for PDF generation
// This allows different PDF implementations to be plugged in
type PDFService interface {
	// GeneratePDF generates a PDF from an invoice and returns the document ID
	GeneratePDF(ctx context.Context, invoice *entities.Invoice, templateID *uint) (documentID uint, err error)

	// StorePDF stores a PDF in document storage (e.g., MinIO) and returns storage key
	StorePDF(ctx context.Context, invoice *entities.Invoice, pdfData []byte) (storageKey string, err error)

	// GetPDFURL gets a download URL for an invoice PDF
	GetPDFURL(ctx context.Context, documentID uint) (string, error)
}

// SetPDFService sets the PDF service implementation
func (s *InvoiceService) SetPDFService(pdfService PDFService) {
	s.pdfService = pdfService
}

// GenerateInvoicePDF generates and stores a PDF for an invoice
func (s *InvoiceService) GenerateInvoicePDF(ctx context.Context, tenantID, invoiceID uint, templateID *uint) (*entities.Invoice, error) {
	// Load invoice
	invoice, err := s.GetInvoice(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}

	// Generate PDF if service is available
	if s.pdfService != nil {
		documentID, err := s.pdfService.GeneratePDF(ctx, invoice, templateID)
		if err != nil {
			return nil, err
		}

		// Update invoice with document ID
		invoice.DocumentID = &documentID
		if err := s.db.WithContext(ctx).Save(invoice).Error; err != nil {
			return nil, err
		}
	}

	return invoice, nil
}

// GetInvoicePDFURL gets the PDF download URL for an invoice
func (s *InvoiceService) GetInvoicePDFURL(ctx context.Context, tenantID, invoiceID uint) (string, error) {
	// Load invoice
	invoice, err := s.GetInvoice(ctx, tenantID, invoiceID)
	if err != nil {
		return "", err
	}

	if invoice.DocumentID == nil {
		return "", nil
	}

	if s.pdfService == nil {
		return "", nil
	}

	return s.pdfService.GetPDFURL(ctx, *invoice.DocumentID)
}
