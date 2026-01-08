package handlers

import (
	"fmt"
	"net/http"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/services"
	templateServices "github.com/unburdy/templates-module/services"
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
