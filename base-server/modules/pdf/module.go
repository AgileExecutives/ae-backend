package pdf

import (
	"context"

	"github.com/ae-base-server/modules/pdf/events"
	"github.com/ae-base-server/modules/pdf/handlers"
	"github.com/ae-base-server/modules/pdf/services"
	"github.com/ae-base-server/pkg/core"
)

// PDFModule represents the PDF generation module
type PDFModule struct {
	pdfHandler    *handlers.PDFHandler
	pdfService    *services.PDFGenerator
	eventHandlers []core.EventHandler
}

// NewPDFModule creates a new PDF module instance
func NewPDFModule() *PDFModule {
	return &PDFModule{}
}

// Name returns the module name
func (m *PDFModule) Name() string {
	return "pdf"
}

// Version returns the module version
func (m *PDFModule) Version() string {
	return "1.0.0"
}

// Description returns the module description
func (m *PDFModule) Description() string {
	return "PDF generation and document management system"
}

// Dependencies returns the module dependencies
func (m *PDFModule) Dependencies() []string {
	return []string{} // PDF module has no dependencies
}

// Initialize initializes the PDF module
func (m *PDFModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing PDF module...")

	// Initialize services
	m.pdfService = services.NewPDFGenerator()

	// Initialize handlers with database for auth middleware
	m.pdfHandler = handlers.NewPDFHandler(m.pdfService, ctx.DB)

	// Initialize event handlers
	m.eventHandlers = []core.EventHandler{
		events.NewPDFGeneratedHandler(ctx.Logger),
		events.NewPDFFailedHandler(ctx.Logger),
	}

	ctx.Logger.Info("PDF module initialized successfully")
	return nil
}

// Start starts the PDF module
func (m *PDFModule) Start(ctx context.Context) error {
	return nil
}

// Stop stops the PDF module
func (m *PDFModule) Stop(ctx context.Context) error {
	return nil
}

// Entities returns the module entities (PDF module has no database entities)
func (m *PDFModule) Entities() []core.Entity {
	return []core.Entity{}
}

// Routes returns the module route providers
func (m *PDFModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{
		// Temporarily disabled - using internal handlers instead
		// m.pdfHandler,
	}
}

// EventHandlers returns the module event handlers
func (m *PDFModule) EventHandlers() []core.EventHandler {
	return m.eventHandlers
}

// Services returns the module service providers
func (m *PDFModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{
		&PDFServiceProvider{
			pdfService: m.pdfService,
		},
	}
}

// Middleware returns the module middleware providers
func (m *PDFModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{
		// PDF module has no middleware
	}
}

// SwaggerPaths returns the swagger documentation paths
func (m *PDFModule) SwaggerPaths() []string {
	return []string{
		// Swagger paths for PDF endpoints
	}
}

// PDFServiceProvider provides the PDF service for dependency injection
type PDFServiceProvider struct {
	pdfService *services.PDFGenerator
}

// ServiceName returns the service name
func (p *PDFServiceProvider) ServiceName() string {
	return "pdf-generator"
}

// ServiceInterface returns the service interface
func (p *PDFServiceProvider) ServiceInterface() interface{} {
	return p.pdfService
}

// Factory creates the service instance
func (p *PDFServiceProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	return p.pdfService, nil
}
