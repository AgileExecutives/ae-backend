package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"time"

	baseAPI "github.com/ae-base-server/api"
	templateServices "github.com/ae-base-server/modules/templates/services"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/unburdy/invoice-module/entities"
	"gorm.io/gorm"
)

// DefaultPDFService implements PDFService using template service and chromedp
type DefaultPDFService struct {
	db               *gorm.DB
	templateService  *templateServices.TemplateService
	documentService  interface{} // Document storage service
	fallbackTemplate string
}

// NewDefaultPDFService creates a new default PDF service
func NewDefaultPDFService(db *gorm.DB, templateService *templateServices.TemplateService) *DefaultPDFService {
	return &DefaultPDFService{
		db:               db,
		templateService:  templateService,
		fallbackTemplate: "statics/pdf_templates/std_invoice.html",
	}
}

// SetDocumentService sets the document storage service
func (s *DefaultPDFService) SetDocumentService(documentService interface{}) {
	s.documentService = documentService
}

// GeneratePDF generates a PDF from an invoice and returns the document ID
func (s *DefaultPDFService) GeneratePDF(ctx context.Context, invoice *entities.Invoice, templateID *uint) (documentID uint, err error) {
	// Prepare contract data
	contractData := s.convertToContractFormat(invoice)

	var html string

	// Try to use specific template if provided, otherwise use default std_invoice template
	if templateID != nil {
		html, err = s.templateService.RenderTemplate(ctx, invoice.TenantID, *templateID, contractData)
		if err != nil {
			return 0, fmt.Errorf("failed to render template %d: %w", *templateID, err)
		}
	} else {
		// Get the std_invoice template
		templates, count, err := s.templateService.ListTemplates(ctx, invoice.TenantID, nil, "DOCUMENT", "std_invoice", nil, 1, 1)
		if err == nil && count > 0 {
			html, err = s.templateService.RenderTemplate(ctx, invoice.TenantID, templates[0].ID, contractData)
			if err != nil {
				return 0, fmt.Errorf("failed to render std_invoice template: %w", err)
			}
		} else {
			// Fallback to file-based template if no std_invoice template found
			html, err = s.renderFallbackTemplate(contractData)
			if err != nil {
				return 0, fmt.Errorf("failed to render fallback template: %w", err)
			}
		}
	}

	// Convert HTML to PDF
	pdfData, err := s.convertHTMLToPDF(ctx, html)
	if err != nil {
		return 0, fmt.Errorf("failed to convert HTML to PDF: %w", err)
	}

	// Store PDF in document storage
	// TODO: Integrate with document service when available
	// For now, return a placeholder document ID
	documentID = uint(time.Now().Unix())

	return documentID, nil
}

// StorePDF stores a PDF in document storage
func (s *DefaultPDFService) StorePDF(ctx context.Context, invoice *entities.Invoice, pdfData []byte) (storageKey string, err error) {
	// TODO: Implement document storage integration
	storageKey = fmt.Sprintf("invoices/%d/%s.pdf", invoice.TenantID, invoice.InvoiceNumber)
	return storageKey, nil
}

// GetPDFURL gets a download URL for an invoice PDF
func (s *DefaultPDFService) GetPDFURL(ctx context.Context, documentID uint) (string, error) {
	// TODO: Implement document URL retrieval
	return fmt.Sprintf("/api/v1/documents/%d/download", documentID), nil
}

// convertToContractFormat converts Invoice entity to std_invoice contract format
func (s *DefaultPDFService) convertToContractFormat(invoice *entities.Invoice) map[string]interface{} {
	// Load organization if not loaded
	var organization baseAPI.Organization
	if err := s.db.First(&organization, invoice.OrganizationID).Error; err != nil {
		organization = baseAPI.Organization{} // Empty fallback
	}

	// Build organization data
	orgData := map[string]interface{}{
		"name":           organization.Name,
		"owner_name":     organization.OwnerName,
		"owner_title":    organization.OwnerTitle,
		"street_address": organization.StreetAddress,
		"zip":            organization.Zip,
		"city":           organization.City,
		"country":        organization.Country,
		"email":          organization.Email,
		"phone":          organization.Phone,
		"tax_id":         organization.TaxID,
	}

	// Add bank account if available
	if organization.BankAccountIBAN != "" {
		orgData["bank_account"] = map[string]interface{}{
			"owner": organization.BankAccountOwner,
			"bank":  organization.BankAccountBank,
			"iban":  organization.BankAccountIBAN,
			"bic":   organization.BankAccountBIC,
		}
	}

	// Build customer data from invoice customer fields
	customerData := map[string]interface{}{
		"name":           invoice.CustomerName,
		"address":        invoice.CustomerAddress,
		"address_ext":    invoice.CustomerAddressExt,
		"zip":            invoice.CustomerZip,
		"city":           invoice.CustomerCity,
		"country":        invoice.CustomerCountry,
		"email":          invoice.CustomerEmail,
		"tax_id":         invoice.CustomerTaxID,
		"contact_person": invoice.CustomerContactPerson,
		"department":     invoice.CustomerDepartment,
	}

	// Build invoice items
	items := make([]map[string]interface{}, 0, len(invoice.Items))
	for _, item := range invoice.Items {
		items = append(items, map[string]interface{}{
			"position":           item.Position,
			"description":        item.Description,
			"quantity":           item.Quantity,
			"unit_price":         item.UnitPrice,
			"tax_rate":           item.TaxRate,
			"amount":             item.Amount,
			"vat_rate":           item.VATRate,
			"vat_exempt":         item.VATExempt,
			"vat_exemption_text": item.VATExemptionText,
		})
	}

	// Build totals
	totalsData := map[string]interface{}{
		"subtotal":   invoice.SubtotalAmount,
		"tax_amount": invoice.TaxAmount,
		"tax_rate":   invoice.TaxRate,
		"total":      invoice.TotalAmount,
		"currency":   invoice.Currency,
	}

	// Build payment info
	paymentData := map[string]interface{}{
		"terms":          invoice.PaymentTerms,
		"net_terms":      invoice.NetTerms,
		"method":         invoice.PaymentMethod,
		"discount_rate":  invoice.DiscountRate,
		"discount_terms": invoice.DiscountTerms,
	}

	// Format dates
	invoiceDate := invoice.InvoiceDate.Format("02.01.2006")
	dueDate := ""
	if invoice.DueDate != nil {
		dueDate = invoice.DueDate.Format("02.01.2006")
	}
	deliveryDate := ""
	if invoice.DeliveryDate != nil {
		deliveryDate = invoice.DeliveryDate.Format("02.01.2006")
	}
	performanceStart := ""
	if invoice.PerformancePeriodStart != nil {
		performanceStart = invoice.PerformancePeriodStart.Format("02.01.2006")
	}
	performanceEnd := ""
	if invoice.PerformancePeriodEnd != nil {
		performanceEnd = invoice.PerformancePeriodEnd.Format("02.01.2006")
	}

	// Determine credit note reference
	creditNoteRef := ""
	if invoice.IsCreditNote && invoice.CreditNoteReferenceID != nil {
		var originalInvoice entities.Invoice
		if err := s.db.Select("invoice_number").First(&originalInvoice, *invoice.CreditNoteReferenceID).Error; err == nil {
			creditNoteRef = originalInvoice.InvoiceNumber
		}
	}

	// Return contract-compatible data format
	return map[string]interface{}{
		"invoice_number":           invoice.InvoiceNumber,
		"invoice_date":             invoiceDate,
		"due_date":                 dueDate,
		"organization":             orgData,
		"customer":                 customerData,
		"subject":                  invoice.Subject,
		"our_reference":            invoice.OurReference,
		"your_reference":           invoice.YourReference,
		"po_number":                invoice.PONumber,
		"delivery_date":            deliveryDate,
		"performance_period_start": performanceStart,
		"performance_period_end":   performanceEnd,
		"invoice_items":            items,
		"totals":                   totalsData,
		"payment":                  paymentData,
		"notes":                    invoice.Notes,
		"is_draft":                 invoice.Status == "draft",
		"is_credit_note":           invoice.IsCreditNote,
		"credit_note_reference":    creditNoteRef,
	}
}

// renderFallbackTemplate renders using file-based template
func (s *DefaultPDFService) renderFallbackTemplate(data map[string]interface{}) (string, error) {
	tmpl, err := template.ParseFiles(s.fallbackTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse fallback template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute fallback template: %w", err)
	}

	return buf.String(), nil
}

// convertHTMLToPDF converts HTML content to PDF using chromedp
func (s *DefaultPDFService) convertHTMLToPDF(ctx context.Context, html string) ([]byte, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create chromedp context
	allocCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var pdfData []byte

	// Encode HTML as base64 data URL
	encodedHTML := base64.StdEncoding.EncodeToString([]byte(html))
	dataURL := "data:text/html;base64," + encodedHTML

	// Navigate to data URL and print to PDF
	err := chromedp.Run(allocCtx,
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfData, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				WithPreferCSSPageSize(false).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfData, nil
}
