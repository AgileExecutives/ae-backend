package routes

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/handlers"
	"github.com/unburdy/documents-module/services"
)

// PDFRoutes implements RouteProvider for PDF generation endpoints
type PDFRoutes struct {
	pdfService *services.PDFService
}

// NewPDFRoutes creates a new PDFRoutes instance
func NewPDFRoutes(pdfService *services.PDFService) *PDFRoutes {
	return &PDFRoutes{
		pdfService: pdfService,
	}
}

// RegisterRoutes registers all PDF generation routes
func (r *PDFRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	handler := handlers.NewPDFHandler(r.pdfService, ctx.DB)

	// PDF routes
	pdfs := router.Group("/pdfs")
	{
		// Generate PDF from HTML
		pdfs.POST("/generate", handler.GeneratePDFFromHTML)

		// Generate PDF from template
		pdfs.POST("/from-template", handler.GeneratePDFFromTemplate)

		// Generate invoice PDF (convenience endpoint)
		pdfs.POST("/invoice", handler.GenerateInvoicePDF)
	}
}

// GetPrefix returns the base path for PDF routes
func (r *PDFRoutes) GetPrefix() string {
	return "/api/v1"
}

// GetMiddleware returns middleware to apply to all PDF routes
func (r *PDFRoutes) GetMiddleware() []gin.HandlerFunc {
	// Auth middleware is typically applied globally
	return []gin.HandlerFunc{}
}

// GetSwaggerTags returns Swagger tags for documentation
func (r *PDFRoutes) GetSwaggerTags() []string {
	return []string{"PDFs"}
}
