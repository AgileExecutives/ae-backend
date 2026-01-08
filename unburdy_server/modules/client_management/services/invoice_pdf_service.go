package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	documentStorage "github.com/unburdy/documents-module/services/storage"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/gorm"
)

// InvoicePDFService handles PDF generation for invoices
type InvoicePDFService struct {
	db           *gorm.DB
	storage      documentStorage.DocumentStorage
	templatePath string
}

// NewInvoicePDFService creates a new invoice PDF service
func NewInvoicePDFService(db *gorm.DB, storage documentStorage.DocumentStorage) *InvoicePDFService {
	// Get template path from environment or use default
	templatePath := os.Getenv("INVOICE_TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "statics/pdf_templates/invoice_units_template.html"
	}

	return &InvoicePDFService{
		db:           db,
		storage:      storage,
		templatePath: templatePath,
	}
}

// InvoicePDFData represents the data structure for invoice PDF templates
type InvoicePDFData struct {
	Invoice               *entities.Invoice
	Organization          interface{}
	Client                interface{}
	CostProvider          interface{}
	Sessions              []map[string]interface{}
	InvoiceItems          []entities.InvoiceItem
	IsDraft               bool
	IsCreditNote          bool
	OriginalInvoiceNumber string
	CreditNoteReason      string
	VATExempt             bool
	VATExemptionText      string
	PaymentDueDate        string
}

// GenerateDraftPDF generates a PDF for a draft invoice with watermark
func (s *InvoicePDFService) GenerateDraftPDF(ctx context.Context, invoice *entities.Invoice) ([]byte, error) {
	data, err := s.prepareInvoiceData(invoice, true)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare invoice data: %w", err)
	}

	return s.generatePDF(ctx, data)
}

// GenerateFinalPDF generates an immutable PDF for a finalized invoice
func (s *InvoicePDFService) GenerateFinalPDF(ctx context.Context, invoice *entities.Invoice) ([]byte, error) {
	data, err := s.prepareInvoiceData(invoice, false)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare invoice data: %w", err)
	}

	return s.generatePDF(ctx, data)
}

// GenerateCreditNotePDF generates a PDF for a credit note
func (s *InvoicePDFService) GenerateCreditNotePDF(ctx context.Context, creditNote *entities.Invoice, originalInvoiceNumber, reason string) ([]byte, error) {
	data, err := s.prepareInvoiceData(creditNote, false)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare credit note data: %w", err)
	}

	data.IsCreditNote = true
	data.OriginalInvoiceNumber = originalInvoiceNumber
	data.CreditNoteReason = reason

	return s.generatePDF(ctx, data)
}

// prepareInvoiceData prepares the template data from invoice entity
func (s *InvoicePDFService) prepareInvoiceData(invoice *entities.Invoice, isDraft bool) (*InvoicePDFData, error) {
	// Load all necessary relationships if not already loaded
	if invoice.Organization == nil || len(invoice.InvoiceItems) == 0 {
		if err := s.db.Preload("InvoiceItems").
			Preload("Organization").
			Preload("ClientInvoices.Client").
			Preload("ClientInvoices.CostProvider").
			First(invoice, invoice.ID).Error; err != nil {
			return nil, fmt.Errorf("failed to reload invoice: %w", err)
		}
	}

	// Get client and cost provider from first ClientInvoice
	var client interface{}
	var costProvider interface{}
	if len(invoice.ClientInvoices) > 0 {
		client = invoice.ClientInvoices[0].Client
		costProvider = invoice.ClientInvoices[0].CostProvider
	}

	// Check if VAT exempt
	vatExempt := false
	vatExemptionText := ""
	if len(invoice.InvoiceItems) > 0 {
		vatExempt = invoice.InvoiceItems[0].VATExempt
		vatExemptionText = invoice.InvoiceItems[0].VATExemptionText
	}

	// Calculate payment due date
	paymentDueDate := ""
	if !isDraft && invoice.Organization != nil {
		// Get payment due days from organization
		var org struct {
			PaymentDueDays int
		}
		if err := s.db.Table("organizations").
			Select("payment_due_days").
			Where("id = ?", invoice.OrganizationID).
			First(&org).Error; err == nil {
			dueDate := invoice.InvoiceDate.AddDate(0, 0, org.PaymentDueDays)
			paymentDueDate = dueDate.Format("02.01.2006")
		}
	}

	// Convert sessions for template (if using session-based items)
	sessions := make([]map[string]interface{}, 0)
	for _, item := range invoice.InvoiceItems {
		if item.ItemType == "session" && item.SessionID != nil {
			// Load session data if needed
			var session struct {
				ID            uint
				OriginalDate  time.Time
				Documentation string
				DurationMin   int
			}
			if err := s.db.Table("sessions").
				Select("id, original_date, documentation, duration_min").
				Where("id = ?", *item.SessionID).
				First(&session).Error; err == nil {
				sessions = append(sessions, map[string]interface{}{
					"OriginalDate":  session.OriginalDate.Format("02.01.2006"),
					"Documentation": session.Documentation,
					"DurationMin":   session.DurationMin,
				})
			}
		}
	}

	return &InvoicePDFData{
		Invoice:          invoice,
		Organization:     invoice.Organization,
		Client:           client,
		CostProvider:     costProvider,
		Sessions:         sessions,
		InvoiceItems:     invoice.InvoiceItems,
		IsDraft:          isDraft,
		IsCreditNote:     invoice.IsCreditNote,
		VATExempt:        vatExempt,
		VATExemptionText: vatExemptionText,
		PaymentDueDate:   paymentDueDate,
	}, nil
}

// generatePDF renders HTML template and converts to PDF
func (s *InvoicePDFService) generatePDF(ctx context.Context, data *InvoicePDFData) ([]byte, error) {
	// Render HTML from template
	html, err := s.renderTemplate(data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Convert HTML to PDF using chromedp
	return s.convertHTMLToPDF(ctx, html)
}

// renderTemplate renders the invoice HTML template
func (s *InvoicePDFService) renderTemplate(data *InvoicePDFData) (string, error) {
	// Parse template file
	tmpl, err := template.ParseFiles(s.templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// convertHTMLToPDF converts HTML content to PDF using chromedp
func (s *InvoicePDFService) convertHTMLToPDF(ctx context.Context, html string) ([]byte, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create chromedp context
	allocCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var pdfData []byte

	// Navigate to data URL and print to PDF
	err := chromedp.Run(allocCtx,
		chromedp.Navigate("data:text/html,"+html),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfData, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithPaperWidth(8.27).   // A4 width in inches
				WithPaperHeight(11.69). // A4 height in inches
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfData, nil
}

// SavePDFToFile saves PDF bytes to a file (for development/testing)
func (s *InvoicePDFService) SavePDFToFile(pdfData []byte, filename string) error {
	outputDir := "tmp"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outputPath, pdfData, 0644); err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	return nil
}

// StoreDraftPDFToMinIO generates and stores a draft invoice PDF in MinIO
// Returns the storage key
func (s *InvoicePDFService) StoreDraftPDFToMinIO(ctx context.Context, invoice *entities.Invoice) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}

	// Generate PDF
	pdfData, err := s.GenerateDraftPDF(ctx, invoice)
	if err != nil {
		return "", fmt.Errorf("failed to generate draft PDF: %w", err)
	}

	// Generate storage key
	storageKey := fmt.Sprintf("invoices/drafts/%d/%s-draft.pdf", invoice.TenantID, invoice.InvoiceNumber)

	// Store in MinIO
	storeReq := documentStorage.StoreRequest{
		Bucket:      "invoices",
		Key:         storageKey,
		Data:        pdfData,
		ContentType: "application/pdf",
		Metadata: map[string]string{
			"invoice_id":     fmt.Sprintf("%d", invoice.ID),
			"tenant_id":      fmt.Sprintf("%d", invoice.TenantID),
			"invoice_number": invoice.InvoiceNumber,
			"status":         "draft",
		},
		ACL: "private",
	}

	key, err := s.storage.Store(ctx, storeReq)
	if err != nil {
		return "", fmt.Errorf("failed to store draft PDF in MinIO: %w", err)
	}

	return key, nil
}

// StoreFinalPDFToMinIO generates and stores a final invoice PDF in MinIO
// Returns the storage key
func (s *InvoicePDFService) StoreFinalPDFToMinIO(ctx context.Context, invoice *entities.Invoice) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}

	// Generate PDF
	pdfData, err := s.GenerateFinalPDF(ctx, invoice)
	if err != nil {
		return "", fmt.Errorf("failed to generate final PDF: %w", err)
	}

	// Generate storage key
	storageKey := fmt.Sprintf("invoices/final/%d/%s.pdf", invoice.TenantID, invoice.InvoiceNumber)

	// Store in MinIO
	storeReq := documentStorage.StoreRequest{
		Bucket:      "invoices",
		Key:         storageKey,
		Data:        pdfData,
		ContentType: "application/pdf",
		Metadata: map[string]string{
			"invoice_id":     fmt.Sprintf("%d", invoice.ID),
			"tenant_id":      fmt.Sprintf("%d", invoice.TenantID),
			"invoice_number": invoice.InvoiceNumber,
			"status":         "final",
			"generated_at":   time.Now().Format(time.RFC3339),
		},
		ACL: "private",
	}

	key, err := s.storage.Store(ctx, storeReq)
	if err != nil {
		return "", fmt.Errorf("failed to store final PDF in MinIO: %w", err)
	}

	return key, nil
}

// StoreCreditNotePDFToMinIO generates and stores a credit note PDF in MinIO
// Returns the storage key
func (s *InvoicePDFService) StoreCreditNotePDFToMinIO(ctx context.Context, creditNote *entities.Invoice, originalInvoiceNumber, reason string) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}

	// Generate PDF
	pdfData, err := s.GenerateCreditNotePDF(ctx, creditNote, originalInvoiceNumber, reason)
	if err != nil {
		return "", fmt.Errorf("failed to generate credit note PDF: %w", err)
	}

	// Generate storage key
	storageKey := fmt.Sprintf("invoices/credit-notes/%d/%s.pdf", creditNote.TenantID, creditNote.InvoiceNumber)

	// Store in MinIO
	storeReq := documentStorage.StoreRequest{
		Bucket:      "invoices",
		Key:         storageKey,
		Data:        pdfData,
		ContentType: "application/pdf",
		Metadata: map[string]string{
			"invoice_id":              fmt.Sprintf("%d", creditNote.ID),
			"tenant_id":               fmt.Sprintf("%d", creditNote.TenantID),
			"invoice_number":          creditNote.InvoiceNumber,
			"status":                  "credit_note",
			"original_invoice_number": originalInvoiceNumber,
			"reason":                  reason,
			"generated_at":            time.Now().Format(time.RFC3339),
		},
		ACL: "private",
	}

	key, err := s.storage.Store(ctx, storeReq)
	if err != nil {
		return "", fmt.Errorf("failed to store credit note PDF in MinIO: %w", err)
	}

	return key, nil
}

// GetPDFURL returns a pre-signed URL for accessing an invoice PDF
func (s *InvoicePDFService) GetPDFURL(ctx context.Context, storageKey string, expiresIn time.Duration) (string, error) {
	if s.storage == nil {
		return "", fmt.Errorf("storage not configured")
	}

	url, err := s.storage.GetURL(ctx, "invoices", storageKey, expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to generate PDF URL: %w", err)
	}

	return url, nil
}

// DeletePDF removes an invoice PDF from MinIO
func (s *InvoicePDFService) DeletePDF(ctx context.Context, storageKey string) error {
	if s.storage == nil {
		return fmt.Errorf("storage not configured")
	}

	err := s.storage.Delete(ctx, "invoices", storageKey)
	if err != nil {
		return fmt.Errorf("failed to delete PDF: %w", err)
	}

	return nil
}
