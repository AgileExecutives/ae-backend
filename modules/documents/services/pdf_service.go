package services

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/unburdy/documents-module/entities"
	"github.com/unburdy/documents-module/services/storage"
	"gorm.io/gorm"
)

// PDFService handles PDF generation from HTML content
type PDFService struct {
	db             *gorm.DB
	storage        storage.DocumentStorage
	templateService *TemplateService
}

// NewPDFService creates a new PDF service instance
func NewPDFService(db *gorm.DB, storage storage.DocumentStorage, templateService *TemplateService) *PDFService {
	return &PDFService{
		db:             db,
		storage:        storage,
		templateService: templateService,
	}
}

// GeneratePDFFromHTMLRequest represents a request to generate PDF from HTML
type GeneratePDFFromHTMLRequest struct {
	HTML        string                 `json:"html" binding:"required"`
	TenantID    uint                   `json:"-"`
	Filename    string                 `json:"filename"`
	DocumentType string                 `json:"document_type"`
	Metadata    map[string]interface{} `json:"metadata"`
	SaveDocument bool                   `json:"save_document"` // If true, save to documents table
}

// GeneratePDFFromTemplateRequest represents a request to generate PDF from template
type GeneratePDFFromTemplateRequest struct {
	TemplateID   uint                   `json:"template_id" binding:"required"`
	TenantID     uint                   `json:"-"`
	OrganizationID *uint                `json:"organization_id"`
	Data         map[string]interface{} `json:"data"`
	Filename     string                 `json:"filename"`
	DocumentType string                 `json:"document_type"`
	Metadata     map[string]interface{} `json:"metadata"`
	SaveDocument bool                   `json:"save_document"` // If true, save to documents table
}

// PDFGenerationResult contains the result of PDF generation
type PDFGenerationResult struct {
	PDFData    []byte             `json:"-"`
	StorageKey string             `json:"storage_key,omitempty"`
	Document   *entities.Document `json:"document,omitempty"`
	Filename   string             `json:"filename"`
	SizeBytes  int64              `json:"size_bytes"`
}

// GeneratePDFFromHTML generates a PDF from HTML content
func (s *PDFService) GeneratePDFFromHTML(ctx context.Context, req *GeneratePDFFromHTMLRequest) (*PDFGenerationResult, error) {
	// Generate PDF using chromedp
	pdfData, err := s.convertHTMLToPDF(ctx, req.HTML)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to PDF: %w", err)
	}

	filename := req.Filename
	if filename == "" {
		filename = fmt.Sprintf("document_%d.pdf", time.Now().Unix())
	}

	result := &PDFGenerationResult{
		PDFData:   pdfData,
		Filename:  filename,
		SizeBytes: int64(len(pdfData)),
	}

	// Save document if requested
	if req.SaveDocument {
		doc, err := s.saveDocument(ctx, req.TenantID, filename, req.DocumentType, pdfData, req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to save document: %w", err)
		}
		result.Document = doc
		result.StorageKey = doc.StorageKey
	}

	return result, nil
}

// GeneratePDFFromTemplate generates a PDF from a template
func (s *PDFService) GeneratePDFFromTemplate(ctx context.Context, req *GeneratePDFFromTemplateRequest) (*PDFGenerationResult, error) {
	// Render template with data
	html, err := s.templateService.RenderTemplate(ctx, req.TenantID, req.TemplateID, req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Generate PDF from rendered HTML
	pdfData, err := s.convertHTMLToPDF(ctx, html)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to PDF: %w", err)
	}

	// Get template for filename
	tmpl, err := s.templateService.GetTemplate(ctx, req.TenantID, req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	filename := req.Filename
	if filename == "" {
		filename = fmt.Sprintf("%s_%d.pdf", tmpl.Name, time.Now().Unix())
	}

	result := &PDFGenerationResult{
		PDFData:   pdfData,
		Filename:  filename,
		SizeBytes: int64(len(pdfData)),
	}

	// Save document if requested
	if req.SaveDocument {
		doc, err := s.saveDocument(ctx, req.TenantID, filename, req.DocumentType, pdfData, req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to save document: %w", err)
		}
		result.Document = doc
		result.StorageKey = doc.StorageKey
	}

	return result, nil
}

// convertHTMLToPDF converts HTML content to PDF using chromedp
func (s *PDFService) convertHTMLToPDF(ctx context.Context, html string) ([]byte, error) {
	// Create chromedp context with timeout
	chromedpCtx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// Set timeout for PDF generation
	chromedpCtx, cancel = context.WithTimeout(chromedpCtx, 30*time.Second)
	defer cancel()

	var pdfBuffer []byte

	// Configure PDF print parameters
	printParams := page.PrintToPDFParams{
		PrintBackground:     true,
		Landscape:           false,
		MarginTop:           0.4,
		MarginBottom:        0.4,
		MarginLeft:          0.4,
		MarginRight:         0.4,
		PaperWidth:          8.27, // A4 width in inches
		PaperHeight:         11.7, // A4 height in inches
		PreferCSSPageSize:   false,
		DisplayHeaderFooter: false,
	}

	// Execute chromedp tasks
	err := chromedp.Run(chromedpCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Set the HTML content
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}

			return page.SetDocumentContent(frameTree.Frame.ID, html).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Generate PDF
			var err error
			pdfBuffer, _, err = printParams.Do(ctx)
			return err
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("chromedp execution failed: %w", err)
	}

	return pdfBuffer, nil
}

// saveDocument saves PDF data to storage and creates a database record
func (s *PDFService) saveDocument(
	ctx context.Context,
	tenantID uint,
	filename string,
	documentType string,
	pdfData []byte,
	metadata map[string]interface{},
) (*entities.Document, error) {
	// Generate storage key
	storageKey := fmt.Sprintf("tenants/%d/documents/%s/%s_%d.pdf",
		tenantID,
		documentType,
		filename,
		time.Now().Unix(),
	)

	// Store in MinIO
	_, err := s.storage.Store(ctx, storage.StoreRequest{
		Bucket:      "documents",
		Key:         storageKey,
		Data:        pdfData,
		ContentType: "application/pdf",
		Metadata: map[string]string{
			"tenant_id":     fmt.Sprintf("%d", tenantID),
			"document_type": documentType,
			"filename":      filename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store PDF: %w", err)
	}

	// Convert metadata to JSONB format
	metadataJSON, err := entities.MarshalJSON(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create database record
	doc := &entities.Document{
		TenantID:      tenantID,
		FileName:      filename,
		DocumentType:  documentType,
		ContentType:   "application/pdf",
		FileSizeBytes:     int64(len(pdfData)),
		StorageBucket: "documents",
		StorageKey:    storageKey,
		Metadata:      metadataJSON,
	}

	if err := s.db.Create(doc).Error; err != nil {
		// Rollback storage if DB insert fails
		s.storage.Delete(ctx, "documents", storageKey)
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	return doc, nil
}

// GenerateInvoicePDF generates a PDF for an invoice
func (s *PDFService) GenerateInvoicePDF(
	ctx context.Context,
	tenantID uint,
	organizationID *uint,
	invoiceData map[string]interface{},
) (*PDFGenerationResult, error) {
	// Get the default invoice template
	tmpl, err := s.templateService.GetDefaultTemplate(ctx, tenantID, organizationID, "invoice")
	if err != nil {
		return nil, fmt.Errorf("no invoice template found: %w", err)
	}

	// Generate PDF from template
	return s.GeneratePDFFromTemplate(ctx, &GeneratePDFFromTemplateRequest{
		TemplateID:     tmpl.ID,
		TenantID:       tenantID,
		OrganizationID: organizationID,
		Data:           invoiceData,
		DocumentType:   "invoice",
		SaveDocument:   true,
		Metadata: map[string]interface{}{
			"template_id":   tmpl.ID,
			"template_name": tmpl.Name,
			"generated_at":  time.Now().Format(time.RFC3339),
		},
	})
}
