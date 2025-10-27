package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ae-saas-basic/ae-saas-basic/internal/services"
	"github.com/gin-gonic/gin"
)

// PDFHandler handles PDF generation requests
type PDFHandler struct {
	pdfService *services.PDFService
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(pdfService *services.PDFService) *PDFHandler {
	return &PDFHandler{
		pdfService: pdfService,
	}
}

// GeneratePDFRequest represents the request structure for PDF generation
type GeneratePDFRequest struct {
	Template   string                 `json:"template" binding:"required"`
	Data       map[string]interface{} `json:"data"`
	Config     *services.PDFConfig    `json:"config,omitempty"`
	OutputPath string                 `json:"output_path,omitempty"`
	Save       bool                   `json:"save,omitempty"` // Whether to save the file on server
}

// PDFResponse represents the response structure for PDF generation
type PDFResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	FilePath  string `json:"file_path,omitempty"`
	FileName  string `json:"file_name,omitempty"`
	Size      int    `json:"size,omitempty"`
	Generated string `json:"generated"`
}

// GeneratePDF generates a PDF and returns it as binary data or saves it
func (h *PDFHandler) GeneratePDF(c *gin.Context) {
	var req GeneratePDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Validate template data
	templateData := services.PDFTemplateData{
		Template:   req.Template,
		Data:       req.Data,
		Config:     req.Config,
		OutputPath: req.OutputPath,
	}

	if err := h.pdfService.ValidateTemplateData(templateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid template data",
			"details": err.Error(),
		})
		return
	}

	// Set timeout for PDF generation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	if req.Save {
		// Generate and save PDF
		filePath, err := h.pdfService.GenerateAndSavePDF(ctx, templateData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to generate PDF",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, PDFResponse{
			Success:   true,
			Message:   "PDF generated successfully",
			FilePath:  filePath,
			Generated: time.Now().Format(time.RFC3339),
		})
	} else {
		// Generate PDF and return as binary
		pdfBytes, err := h.pdfService.GeneratePDF(ctx, templateData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to generate PDF",
				"details": err.Error(),
			})
			return
		}

		// Set headers for PDF download
		filename := fmt.Sprintf("%s_%s.pdf", req.Template, time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	}
}

// GeneratePDFFromHTML generates a PDF directly from HTML content
func (h *PDFHandler) GeneratePDFFromHTML(c *gin.Context) {
	type HTMLRequest struct {
		HTML   string              `json:"html" binding:"required"`
		Config *services.PDFConfig `json:"config,omitempty"`
		Save   bool                `json:"save,omitempty"`
	}

	var req HTMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Use default config if not provided
	config := services.DefaultPDFConfig()
	if req.Config != nil {
		config = *req.Config
	}

	// Set timeout for PDF generation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Generate PDF from HTML
	pdfBytes, err := h.pdfService.GeneratePDFFromHTML(ctx, req.HTML, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate PDF from HTML",
			"details": err.Error(),
		})
		return
	}

	if req.Save {
		// Save PDF to file
		filename := fmt.Sprintf("html_pdf_%s.pdf", time.Now().Format("20060102_150405"))
		filePath, err := h.pdfService.SavePDF(pdfBytes, filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to save PDF",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, PDFResponse{
			Success:   true,
			Message:   "PDF generated and saved successfully",
			FilePath:  filePath,
			FileName:  filename,
			Size:      len(pdfBytes),
			Generated: time.Now().Format(time.RFC3339),
		})
	} else {
		// Return PDF as binary
		filename := fmt.Sprintf("generated_%s.pdf", time.Now().Format("20060102_150405"))
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	}
}

// PreviewTemplate returns rendered HTML for template preview
func (h *PDFHandler) PreviewTemplate(c *gin.Context) {
	type PreviewRequest struct {
		Template string                 `json:"template" binding:"required"`
		Data     map[string]interface{} `json:"data"`
	}

	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Render template
	html, err := h.pdfService.PreviewTemplate(req.Template, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to render template",
			"details": err.Error(),
		})
		return
	}

	// Return HTML for preview
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// ListTemplates returns available PDF templates
func (h *PDFHandler) ListTemplates(c *gin.Context) {
	templates := h.pdfService.GetAvailableTemplates()

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"templates": templates,
		"count":     len(templates),
	})
}

// GetTemplateInfo returns information about a specific template
func (h *PDFHandler) GetTemplateInfo(c *gin.Context) {
	templateName := c.Param("template")
	if templateName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Template name is required",
		})
		return
	}

	info, err := h.pdfService.GetTemplateInfo(templateName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Template not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"template": info,
	})
}

// GetPDFConfig returns current PDF configuration
func (h *PDFHandler) GetPDFConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"config":  h.pdfService.Config,
		"default": services.DefaultPDFConfig(),
	})
}

// StreamPDF streams PDF generation for large files
func (h *PDFHandler) StreamPDF(c *gin.Context) {
	var req GeneratePDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Validate template data
	templateData := services.PDFTemplateData{
		Template:   req.Template,
		Data:       req.Data,
		Config:     req.Config,
		OutputPath: req.OutputPath,
	}

	if err := h.pdfService.ValidateTemplateData(templateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid template data",
			"details": err.Error(),
		})
		return
	}

	// Set timeout for PDF generation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second) // Longer timeout for streaming
	defer cancel()

	// Set headers for PDF streaming
	filename := fmt.Sprintf("%s_%s.pdf", req.Template, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Transfer-Encoding", "chunked")

	// Stream PDF generation
	if err := h.pdfService.StreamPDF(ctx, templateData, c.Writer); err != nil {
		// If streaming fails, we can't send JSON anymore, so log the error
		fmt.Printf("PDF streaming failed: %v\n", err)
		return
	}
}
