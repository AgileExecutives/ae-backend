package routes

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/handlers"
	"github.com/unburdy/documents-module/services"
)

// InvoiceNumberRoutes implements RouteProvider for invoice number endpoints
type InvoiceNumberRoutes struct {
	invoiceNumberService *services.InvoiceNumberService
}

// NewInvoiceNumberRoutes creates a new InvoiceNumberRoutes instance
func NewInvoiceNumberRoutes(invoiceNumberService *services.InvoiceNumberService) *InvoiceNumberRoutes {
	return &InvoiceNumberRoutes{
		invoiceNumberService: invoiceNumberService,
	}
}

// RegisterRoutes registers all invoice number routes
func (r *InvoiceNumberRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	handler := handlers.NewInvoiceNumberHandler(r.invoiceNumberService, ctx.DB)

	// Invoice number routes
	invoiceNumbers := router.Group("/invoice-numbers")
	{
		// Generate next invoice number
		invoiceNumbers.POST("/generate", handler.GenerateInvoiceNumber)

		// Get current sequence without incrementing
		invoiceNumbers.GET("/current", handler.GetCurrentSequence)

		// Get invoice number history
		invoiceNumbers.GET("/history", handler.GetInvoiceNumberHistory)

		// Void an invoice number
		invoiceNumbers.POST("/void", handler.VoidInvoiceNumber)
	}
}

// GetPrefix returns the base path for invoice number routes
func (r *InvoiceNumberRoutes) GetPrefix() string {
	return "/api/v1"
}

// GetMiddleware returns middleware to apply to all invoice number routes
func (r *InvoiceNumberRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// Auth middleware is typically applied globally
	}
}

// GetSwaggerTags returns Swagger tags for documentation
func (r *InvoiceNumberRoutes) GetSwaggerTags() []string {
	return []string{"Invoice Numbers"}
}
