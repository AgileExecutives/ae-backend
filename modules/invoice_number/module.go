package invoicenumber

import (
	"context"

	"github.com/unburdy/invoice-number-module/entities"
	"github.com/unburdy/invoice-number-module/routes"
	"github.com/unburdy/invoice-number-module/services"
	"github.com/ae-base-server/pkg/core"
)

// InvoiceNumberModule represents the invoice number generation module
type InvoiceNumberModule struct {
	invoiceNumberService *services.InvoiceNumberService
	invoiceNumberRoutes  *routes.InvoiceNumberRoutes
}

// NewInvoiceNumberModule creates a new invoice number module instance
func NewInvoiceNumberModule() core.Module {
	return &InvoiceNumberModule{}
}

func (m *InvoiceNumberModule) Name() string {
	return "invoice_number"
}

func (m *InvoiceNumberModule) Version() string {
	return "1.0.0"
}

func (m *InvoiceNumberModule) Dependencies() []string {
	return []string{"base"}
}

func (m *InvoiceNumberModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing invoice number module...")

	// Initialize service
	m.invoiceNumberService = services.NewInvoiceNumberService(ctx.DB)

	// Initialize routes
	m.invoiceNumberRoutes = routes.NewInvoiceNumberRoutes(m.invoiceNumberService, ctx.DB)

	ctx.Logger.Info("Invoice number module initialized successfully")
	return nil
}

func (m *InvoiceNumberModule) Start(ctx context.Context) error {
	return nil
}

func (m *InvoiceNumberModule) Stop(ctx context.Context) error {
	return nil
}

func (m *InvoiceNumberModule) Entities() []core.Entity {
	return []core.Entity{
		entities.NewInvoiceNumberEntity(),
		entities.NewInvoiceNumberLogEntity(),
	}
}

func (m *InvoiceNumberModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{
		m.invoiceNumberRoutes,
	}
}

func (m *InvoiceNumberModule) EventHandlers() []core.EventHandler {
	return []core.EventHandler{}
}

func (m *InvoiceNumberModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{}
}

func (m *InvoiceNumberModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

func (m *InvoiceNumberModule) SwaggerPaths() []string {
	return []string{
		"./modules/invoice_number/handlers",
		"./modules/invoice_number/entities",
	}
}
