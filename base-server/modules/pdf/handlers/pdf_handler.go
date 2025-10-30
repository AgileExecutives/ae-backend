package handlers

import (
	"net/http"

	"github.com/ae-base-server/modules/pdf/services"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
)

// PDFHandler handles PDF generation requests
type PDFHandler struct {
	pdfService *services.PDFGenerator
}

// NewPDFHandler creates a new PDF handler
func NewPDFHandler(pdfService *services.PDFGenerator) *PDFHandler {
	return &PDFHandler{pdfService: pdfService}
}

// RegisterRoutes registers PDF-related routes
func (h *PDFHandler) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	pdfRoutes := router.Group("/pdf")
	{
		pdfRoutes.POST("/create", h.GeneratePDFFromTemplate)
	}
}

// GetPrefix returns the route prefix
func (h *PDFHandler) GetPrefix() string {
	return "/api/v1"
}

// GetMiddleware returns route middleware
func (h *PDFHandler) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// TODO: Add auth middleware
	}
}

// GetSwaggerTags returns swagger tags for documentation
func (h *PDFHandler) GetSwaggerTags() []string {
	return []string{"pdf"}
}

// PDFGenerateRequest represents the request structure for PDF generation
type PDFGenerateRequest struct {
	Data         map[string]interface{} `json:"data"`
	TemplateName string                 `json:"templateName" example:"report.html"`
	FileName     string                 `json:"fileName" example:"generated-report"`
}

// PDFGenerateResponse represents the response structure for successful PDF generation
type PDFGenerateResponse struct {
	Success  bool   `json:"success" example:"true"`
	Message  string `json:"message" example:"PDF generated successfully"`
	Filename string `json:"filename" example:"generated-report.pdf"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Template name is required"`
	Details string `json:"details,omitempty" example:"Additional error details"`
}

// GeneratePDFFromTemplate generates a PDF from a specified template and data
// @Summary Generate PDF from template
// @Description Generate a PDF document based on a specified template and data
// @Tags pdf
// @Accept json
// @Produce application/json
// @Security BearerAuth
// @Param request body PDFGenerateRequest true "PDF generation request"
// @Success 200 {object} PDFGenerateResponse "PDF generated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Failed to generate PDF"
// @Router /api/v1/pdf/create [post]
func (h *PDFHandler) GeneratePDFFromTemplate(c *gin.Context) {
	var requestBody PDFGenerateRequest

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate required fields manually to give specific error messages
	if requestBody.Data == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Data is required",
		})
		return
	}

	if requestBody.TemplateName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Template name is required",
		})
		return
	}

	if requestBody.FileName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "File name is required",
		})
		return
	}

	pdfName, err := h.pdfService.GeneratePDF(requestBody.Data, requestBody.TemplateName, requestBody.FileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate PDF",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PDFGenerateResponse{
		Success:  true,
		Message:  "PDF generated successfully",
		Filename: pdfName,
	})
}
