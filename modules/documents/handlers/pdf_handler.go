package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	baseAPI "github.com/ae-base-server/api"
	templateServices "github.com/ae-base-server/modules/templates/services"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/services"
	"gorm.io/gorm"
)

// PDFHandler handles PDF generation requests
type PDFHandler struct {
	pdfService      *services.PDFService
	templateService *templateServices.TemplateService
	db              *gorm.DB
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(pdfService *services.PDFService, templateService *templateServices.TemplateService, db *gorm.DB) *PDFHandler {
	return &PDFHandler{
		pdfService:      pdfService,
		templateService: templateService,
		db:              db,
	}
}

// GeneratePDFFromHTML godoc
// @Summary Generate PDF from HTML
// @Description Generate a PDF document from HTML content using Chromedp
// @Tags PDFs
// @Accept json
// @Produce application/pdf
// @Param request body services.GeneratePDFFromHTMLRequest true "HTML and metadata"
// @Success 200 {file} binary "PDF file"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /pdfs/generate [post]
// @ID generatePDFFromHTML
func (h *PDFHandler) GeneratePDFFromHTML(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	var req services.GeneratePDFFromHTMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.TenantID = uint(tenantID)

	result, err := h.pdfService.GeneratePDFFromHTML(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return PDF as downloadable file
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+result.Filename+"\"")
	c.Header("Content-Length", string(rune(result.SizeBytes)))
	c.Data(http.StatusOK, "application/pdf", result.PDFData)
}

// GeneratePDFFromTemplate godoc
// @Summary Generate PDF from template
// @Description Generate a PDF document from a template with data
// @Tags PDFs
// @Accept json
// @Produce application/pdf
// @Param request body services.GeneratePDFFromTemplateRequest true "Template ID and data"
// @Success 200 {file} binary "PDF file"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /pdfs/from-template [post]
// @ID generatePDFFromTemplate
func (h *PDFHandler) GeneratePDFFromTemplate(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	user, err := baseAPI.GetUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user required"})
		return
	}

	var req services.GeneratePDFFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.TenantID = uint(tenantID)
	if req.UserID == 0 {
		req.UserID = user.ID
	}

	// Step 1: Get template and render HTML
	html, err := h.templateService.RenderTemplate(c.Request.Context(), req.TenantID, req.TemplateID, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to render template: %v", err)})
		return
	}

	// Get template for filename
	tmpl, err := h.templateService.GetTemplate(c.Request.Context(), req.TenantID, req.TemplateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get template: %v", err)})
		return
	}

	filename := req.Filename
	if filename == "" {
		filename = fmt.Sprintf("%s_%d.pdf", tmpl.Name, time.Now().Unix())
	}

	// Step 2: Generate PDF from HTML
	htmlReq := services.GeneratePDFFromHTMLRequest{
		TenantID:     req.TenantID,
		UserID:       req.UserID,
		HTML:         html,
		Filename:     filename,
		DocumentType: req.DocumentType,
		SaveDocument: req.SaveDocument,
		Metadata:     req.Metadata,
	}

	result, err := h.pdfService.GeneratePDFFromHTML(c.Request.Context(), &htmlReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return PDF as downloadable file
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+result.Filename+"\"")
	c.Header("Content-Length", string(rune(result.SizeBytes)))
	c.Data(http.StatusOK, "application/pdf", result.PDFData)
}

// GenerateInvoicePDF generates a PDF invoice using default templates
// @Summary Generate invoice PDF
// @Description Generate a PDF invoice document using predefined templates
// @Tags PDFs
// @Accept json
// @Produce application/pdf
// @Param invoice_id path int true "Invoice ID"
// @Success 200 {file} binary "PDF file"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Invoice not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security Bearer
// @Router /documents/pdf/invoice/{invoice_id} [post]
func (h *PDFHandler) GenerateInvoicePDF(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	// Get invoice ID from URL parameter
	invoiceIDParam := c.Param("invoice_id")
	invoiceID, err := strconv.ParseUint(invoiceIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invoice_id"})
		return
	}

	// For now, use a simple HTML template for invoice
	// In the future, this should fetch invoice data and use a proper template
	invoiceHTML := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Invoice %d</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			.header { text-align: center; margin-bottom: 30px; }
			.invoice-details { margin-bottom: 20px; }
			.footer { margin-top: 30px; font-size: 12px; color: #666; }
		</style>
	</head>
	<body>
		<div class="header">
			<h1>Invoice #%d</h1>
			<p>Generated on %s</p>
		</div>
		<div class="invoice-details">
			<h3>Invoice Details</h3>
			<p>Invoice ID: %d</p>
			<p>This is a sample invoice PDF generated from the default template.</p>
		</div>
		<div class="footer">
			<p>Generated by the invoice system</p>
		</div>
	</body>
	</html>
	`, invoiceID, invoiceID, time.Now().Format("January 2, 2006"), invoiceID)

	// Create request for PDF generation from HTML
	filename := fmt.Sprintf("invoice_%d_%s.pdf", invoiceID, time.Now().Format("20060102"))
	req := services.GeneratePDFFromHTMLRequest{
		HTML:         invoiceHTML,
		Filename:     filename,
		DocumentType: "invoice",
		SaveDocument: true,
		Metadata: map[string]interface{}{
			"invoice_id": invoiceID,
			"type":       "invoice_pdf",
		},
	}

	result, err := h.pdfService.GeneratePDFFromHTML(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return PDF as downloadable file
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", result.Filename))
	c.Header("Content-Length", string(rune(result.SizeBytes)))
	c.Data(http.StatusOK, "application/pdf", result.PDFData)
}
