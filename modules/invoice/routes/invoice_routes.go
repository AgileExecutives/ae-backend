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
		invoices.POST("", r.handler.CreateInvoice)
		invoices.GET("", r.handler.ListInvoices)
		invoices.GET("/:id", r.handler.GetInvoice)
		invoices.PUT("/:id", r.handler.UpdateInvoice)
		invoices.DELETE("/:id", r.handler.DeleteInvoice)
		invoices.POST("/:id/mark-paid", r.handler.MarkAsPaid)
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
