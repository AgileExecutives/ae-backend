package routes

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/invoice-module/handlers"
)

// InvoiceRoutes implements RouteProvider for invoice management
type InvoiceRoutes struct {
	handler *handlers.InvoiceHandler
}

// NewInvoiceRoutes creates a new InvoiceRoutes instance
func NewInvoiceRoutes(handler *handlers.InvoiceHandler) *InvoiceRoutes {
	return &InvoiceRoutes{handler: handler}
}

// RegisterRoutes registers all invoice routes
func (r *InvoiceRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	invoices := router.Group("/invoices")
	{
		// CRUD operations
		invoices.POST("", r.handler.CreateInvoice)
		invoices.GET("", r.handler.ListInvoices)
		invoices.GET("/:id", r.handler.GetInvoice)
		invoices.PUT("/:id", r.handler.UpdateInvoice)
		invoices.DELETE("/:id", r.handler.DeleteInvoice)

		// Workflow operations
		invoices.POST("/:id/finalize", r.handler.FinalizeInvoice)
		invoices.POST("/:id/send", r.handler.MarkInvoiceAsSent)
		invoices.POST("/:id/pay", r.handler.MarkInvoiceAsPaid)
		invoices.POST("/:id/remind", r.handler.SendInvoiceReminder)
		invoices.POST("/:id/cancel", r.handler.CancelInvoice)

		// PDF generation
		invoices.POST("/:id/generate-pdf", r.handler.GenerateInvoicePDF)
	}
}

// GetPrefix returns the route prefix for invoices
func (r *InvoiceRoutes) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware for invoice routes
func (r *InvoiceRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

// GetSwaggerTags returns Swagger tags for invoice routes
func (r *InvoiceRoutes) GetSwaggerTags() []string {
	return []string{"Invoices"}
}
