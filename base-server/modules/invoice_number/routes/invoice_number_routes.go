package routes

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/ae-base-server/modules/invoice_number/handlers"
	"github.com/ae-base-server/modules/invoice_number/services"
	"gorm.io/gorm"
)

// InvoiceNumberRoutes implements RouteProvider for invoice number endpoints
type InvoiceNumberRoutes struct {
	invoiceNumberService *services.InvoiceNumberService
	db                   *gorm.DB
}

// NewInvoiceNumberRoutes creates a new InvoiceNumberRoutes instance
func NewInvoiceNumberRoutes(invoiceNumberService *services.InvoiceNumberService, db *gorm.DB) *InvoiceNumberRoutes {
	return &InvoiceNumberRoutes{
		invoiceNumberService: invoiceNumberService,
		db:                   db,
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
	return ""
}

// GetMiddleware returns middleware to apply to all invoice number routes
func (r *InvoiceNumberRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthMiddleware(r.db), // Require authentication for tenant ID extraction
	}
}

// GetSwaggerTags returns Swagger tags for documentation
func (r *InvoiceNumberRoutes) GetSwaggerTags() []string {
	return []string{"Invoice Numbers"}
}
