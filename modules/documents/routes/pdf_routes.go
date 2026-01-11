package routes

import (
	baseAPI "github.com/ae-base-server/api"
	templateServices "github.com/ae-base-server/modules/templates/services"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/handlers"
	"github.com/unburdy/documents-module/services"
	"gorm.io/gorm"
)

// PDFRoutes implements RouteProvider for PDF generation endpoints
type PDFRoutes struct {
	pdfService      *services.PDFService
	templateService *templateServices.TemplateService
	db              *gorm.DB
}

// NewPDFRoutes creates a new PDFRoutes instance
func NewPDFRoutes(pdfService *services.PDFService, templateService *templateServices.TemplateService, db *gorm.DB) *PDFRoutes {
	return &PDFRoutes{
		pdfService:      pdfService,
		templateService: templateService,
		db:              db,
	}
}

// RegisterRoutes registers all PDF generation routes
func (r *PDFRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	handler := handlers.NewPDFHandler(r.pdfService, r.templateService, ctx.DB)

	// PDF routes
	pdfs := router.Group("/pdfs")
	{
		// Generate PDF from HTML
		pdfs.POST("/generate", handler.GeneratePDFFromHTML)

		// Generate PDF from template
		pdfs.POST("/from-template", handler.GeneratePDFFromTemplate)

		// Generate invoice PDF
		pdfs.POST("/invoice/:invoice_id", handler.GenerateInvoicePDF)
	}
}

// GetPrefix returns the base path for PDF routes
func (r *PDFRoutes) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware to apply to all PDF routes
func (r *PDFRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		baseAPI.AuthMiddleware(r.db), // Require authentication for tenant ID extraction
	}
}

// GetSwaggerTags returns Swagger tags for documentation
func (r *PDFRoutes) GetSwaggerTags() []string {
	return []string{"PDFs"}
}
