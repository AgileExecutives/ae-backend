package handlers

import (
	"fmt"
	"net/http"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/services"
	"gorm.io/gorm"
)

// PDFHandler handles PDF generation requests
type PDFHandler struct {
	pdfService *services.PDFService
	db         *gorm.DB
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(pdfService *services.PDFService, db *gorm.DB) *PDFHandler {
	return &PDFHandler{
		pdfService: pdfService,
		db:         db,
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

	var req services.GeneratePDFFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.TenantID = uint(tenantID)

	result, err := h.pdfService.GeneratePDFFromTemplate(c.Request.Context(), &req)
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

// GenerateInvoicePDF godoc
// @Summary Generate invoice PDF
// @Description Generate a PDF invoice using the default invoice template
// @Tags PDFs
// @Accept json
// @Produce application/pdf
// @Param organization_id query int false "Organization ID (for template selection)"
// @Param request body map[string]interface{} true "Invoice data"
// @Success 200 {file} binary "PDF file"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /pdfs/invoice [post]
// @ID generateInvoicePDF
func (h *PDFHandler) GenerateInvoicePDF(c *gin.Context) {
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant_id required"})
		return
	}

	// Get organization ID from query params
	var organizationID *uint
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		var orgID uint
		if _, err := fmt.Sscanf(orgIDStr, "%d", &orgID); err == nil {
			organizationID = &orgID
		}
	}

	// Parse invoice data from request body
	var invoiceData map[string]interface{}
	if err := c.ShouldBindJSON(&invoiceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.pdfService.GenerateInvoicePDF(c.Request.Context(), uint(tenantID), organizationID, invoiceData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get invoice number for filename if available
	filename := "invoice.pdf"
	if invoiceNum, ok := invoiceData["invoice_number"].(string); ok {
		filename = fmt.Sprintf("invoice_%s.pdf", invoiceNum)
	}

	// Return PDF as downloadable file
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Length", string(rune(result.SizeBytes)))
	c.Data(http.StatusOK, "application/pdf", result.PDFData)
}
