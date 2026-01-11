package client_management

import (
	"context"
	"fmt"

	baseAPI "github.com/ae-base-server/api"
	emailServices "github.com/ae-base-server/modules/email/services"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	templateServices "github.com/ae-base-server/modules/templates/services"
	bookingServices "github.com/unburdy/booking-module/services"
	documentServices "github.com/unburdy/documents-module/services/storage"
	"github.com/unburdy/unburdy-server-api/internal/services"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/events"
	"github.com/unburdy/unburdy-server-api/modules/client_management/handlers"
	"github.com/unburdy/unburdy-server-api/modules/client_management/routes"
	clientServices "github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// Module implements the baseAPI.ModuleRouteProvider interface
type Module struct {
	routeProvider *routes.RouteProvider
}

// NewModule creates a new client management module
func NewModule(db *gorm.DB) baseAPI.ModuleRouteProvider {
	// Initialize modular services
	clientService := clientServices.NewClientService(db)
	costProviderService := clientServices.NewCostProviderService(db)
	// Pass nil for email service in legacy module (email won't work but won't crash)
	sessionService := clientServices.NewSessionService(db, nil)
	invoiceService := clientServices.NewInvoiceService(db)
	extraEffortService := clientServices.NewExtraEffortService(db)
	xrechnungService := services.NewXRechnungService()

	// Initialize handlers
	clientHandler := handlers.NewClientHandler(clientService)
	costProviderHandler := handlers.NewCostProviderHandler(costProviderService)
	sessionHandler := handlers.NewSessionHandler(sessionService)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceService, xrechnungService, nil) // nil audit service for legacy module
	invoiceAdapterHandler := handlers.NewInvoiceAdapterHandler(db, "")
	extraEffortHandler := handlers.NewExtraEffortHandler(extraEffortService)

	// Initialize route provider with database for auth middleware
	routeProvider := routes.NewRouteProvider(clientHandler, costProviderHandler, sessionHandler, invoiceHandler, invoiceAdapterHandler, extraEffortHandler, db)

	return &Module{
		routeProvider: routeProvider,
	}
}

// RegisterRoutes implements baseAPI.ModuleRouteProvider
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	// Directly call the method to avoid any interface conflicts
	// Pass nil context since old interface doesn't support public routes
	m.routeProvider.RegisterRoutes(router, nil)
}

// GetPrefix implements baseAPI.ModuleRouteProvider
func (m *Module) GetPrefix() string {
	return m.routeProvider.GetPrefix()
}

// GetEntitiesForMigration implements baseAPI.ModuleWithEntities
func (m *Module) GetEntitiesForMigration() []interface{} {
	return []interface{}{
		&entities.Client{},
		&entities.CostProvider{},
		&entities.Session{},
		&entities.Invoice{},
		&entities.InvoiceItem{},
		&entities.ClientInvoice{},
		&entities.ExtraEffort{},
	}
}

// CoreModule implements the core.Module interface for bootstrap system integration
type CoreModule struct {
	db                    *gorm.DB
	logger                core.Logger
	clientHandlers        *handlers.ClientHandler
	costProviderHandler   *handlers.CostProviderHandler
	sessionHandler        *handlers.SessionHandler
	invoiceHandler        *handlers.InvoiceHandler
	invoiceAdapterHandler *handlers.InvoiceAdapterHandler
	routeProvider         *routes.RouteProvider
}

// NewCoreModule creates a new client management module compatible with bootstrap system
func NewCoreModule() core.Module {
	return &CoreModule{}
}

func (m *CoreModule) Name() string {
	return "client-management"
}

func (m *CoreModule) Version() string {
	return "1.0.0"
}

func (m *CoreModule) Dependencies() []string {
	return []string{"base", "email", "booking", "audit"} // Depends on base module for users/tenants, email for confirmations, booking for token validation, and audit for logging
}

func (m *CoreModule) Initialize(ctx core.ModuleContext) error {
	fmt.Println("\n🚀🚀🚀 CLIENT MANAGEMENT MODULE INITIALIZE CALLED 🚀🚀🚀")
	ctx.Logger.Info("Initializing client management module...")
	m.db = ctx.DB
	m.logger = ctx.Logger

	// Initialize modular services
	clientService := clientServices.NewClientService(ctx.DB)
	costProviderService := clientServices.NewCostProviderService(ctx.DB)

	// Get email service from registry
	var emailService *emailServices.EmailService
	fmt.Println("🔍 CLIENT MANAGEMENT: Looking for email-service in registry...")
	ctx.Logger.Info("🔍 Client Management: Looking for email-service in registry...")
	if emailSvcRaw, ok := ctx.Services.Get("email-service"); ok {
		fmt.Printf("✅ CLIENT MANAGEMENT: Email service found in registry! Raw value ptr=%p\n", emailSvcRaw)
		ctx.Logger.Info("✅ Client Management: Email service raw object found in registry")
		if emailSvc, ok := emailSvcRaw.(*emailServices.EmailService); ok {
			fmt.Printf("✅ CLIENT MANAGEMENT: Type assertion succeeded! emailSvc ptr=%p\n", emailSvc)
			emailService = emailSvc
			fmt.Printf("✅ CLIENT MANAGEMENT: After assignment, emailService ptr=%p\n", emailService)
			fmt.Println("✅ CLIENT MANAGEMENT: Email service type assertion SUCCESS!")
			ctx.Logger.Info("✅ Client Management: Email service successfully type-asserted")
		} else {
			fmt.Println("❌ CLIENT MANAGEMENT: Email service type assertion FAILED!")
			ctx.Logger.Warn("❌ Client Management: Email service found but type assertion failed")
		}
	} else {
		fmt.Println("❌ CLIENT MANAGEMENT: Email service NOT found in registry!")
		ctx.Logger.Warn("❌ Client Management: Email service NOT found in registry")
	}

	fmt.Printf("🔧 About to create SessionService with emailService ptr=%p\n", emailService)
	sessionService := clientServices.NewSessionService(ctx.DB, emailService)
	fmt.Printf("🔧 SessionService created successfully\n")

	// Try to get booking link service from service registry (if available)
	// This is optional - if booking module isn't loaded, token booking won't work
	if bookingLinkSvcRaw, ok := ctx.Services.Get("booking-link-service"); ok {
		ctx.Logger.Info("✅ Client Management: Booking link service found in registry")
		// Type assert to the actual BookingLinkService type
		if bookingLinkSvc, ok := bookingLinkSvcRaw.(*bookingServices.BookingLinkService); ok {
			sessionService.SetBookingLinkService(bookingLinkSvc)
			ctx.Logger.Info("✅ Client Management: Successfully injected booking link service into session service")
		} else {
			ctx.Logger.Error("❌ Client Management: Booking link service type assertion failed")
		}
	} else {
		ctx.Logger.Warn("⚠️ Client Management: Booking link service not found in registry - token-based booking will not be available")
	}

	// Initialize invoice service
	invoiceService := clientServices.NewInvoiceService(ctx.DB)

	// Initialize extra effort service
	extraEffortService := clientServices.NewExtraEffortService(ctx.DB)

	// Get document storage from registry
	if docStorageRaw, ok := ctx.Services.Get("document-storage"); ok {
		if docStorage, ok := docStorageRaw.(documentServices.DocumentStorage); ok {
			ctx.Logger.Info("✅ Client Management: Document storage successfully retrieved from registry")
			invoiceService.SetDocumentStorage(docStorage)
			ctx.Logger.Info("✅ Client Management: Invoice PDF service initialized")
		} else {
			ctx.Logger.Warn("⚠️ Client Management: Document storage type assertion failed")
		}
	} else {
		ctx.Logger.Warn("⚠️ Client Management: Document storage not found in registry")
	}

	// Get template service from registry (for PDF generation)
	// var templateService *templateServices.TemplateService
	if templateSvcRaw, ok := ctx.Services.Get("template_service"); ok {
		if _, ok := templateSvcRaw.(*templateServices.TemplateService); ok {
			// templateService = templateSvc
			ctx.Logger.Info("✅ Client Management: Template service successfully retrieved from registry")
		} else {
			ctx.Logger.Warn("⚠️ Client Management: Template service type assertion failed")
		}
	} else {
		ctx.Logger.Warn("⚠️ Client Management: Template service not found in registry - invoice PDF generation will not work")
	}

	// Get audit service from registry (for audit logging)
	var auditService interface{}
	if auditSvcRaw, ok := ctx.Services.Get("audit-service"); ok {
		auditService = auditSvcRaw
		ctx.Logger.Info("✅ Client Management: Audit service successfully retrieved from registry")
	} else {
		ctx.Logger.Warn("⚠️ Client Management: Audit service not found in registry - audit logging will not work")
	}

	// Initialize handlers
	m.clientHandlers = handlers.NewClientHandler(clientService)
	m.costProviderHandler = handlers.NewCostProviderHandler(costProviderService)
	m.sessionHandler = handlers.NewSessionHandler(sessionService)
	xrechnungService := services.NewXRechnungService()
	m.invoiceHandler = handlers.NewInvoiceHandler(invoiceService, xrechnungService, auditService)
	extraEffortHandler := handlers.NewExtraEffortHandler(extraEffortService)

	// Initialize invoice adapter handler (for invoice module integration)
	// TODO: Get invoice module URL from config
	m.invoiceAdapterHandler = handlers.NewInvoiceAdapterHandler(ctx.DB, "")

	// Initialize route provider with database for auth middleware
	m.routeProvider = routes.NewRouteProvider(m.clientHandlers, m.costProviderHandler, m.sessionHandler, m.invoiceHandler, m.invoiceAdapterHandler, extraEffortHandler, ctx.DB)

	ctx.Logger.Info("Client management module initialized successfully")
	return nil
}

func (m *CoreModule) Start(ctx context.Context) error {
	// Start any background services if needed
	return nil
}

func (m *CoreModule) Stop(ctx context.Context) error {
	// Stop any background services if needed
	return nil
}

func (m *CoreModule) Entities() []core.Entity {
	return []core.Entity{
		entities.NewClientEntity(),
		entities.NewCostProviderEntity(),
		entities.NewSessionEntity(),
		entities.NewInvoiceEntity(),
		entities.NewInvoiceItemEntity(),
		entities.NewClientInvoiceEntity(),
		entities.NewExtraEffortEntity(),
	}
}

func (m *CoreModule) Routes() []core.RouteProvider {
	if m.routeProvider == nil {
		return []core.RouteProvider{}
	}
	return []core.RouteProvider{
		&clientManagementRouteAdapter{provider: m.routeProvider},
	}
}

func (m *CoreModule) EventHandlers() []core.EventHandler {
	if m.db == nil {
		return []core.EventHandler{}
	}
	return []core.EventHandler{
		events.NewCalendarEntryDeletedHandler(m.db, m.logger),
	}
}

func (m *CoreModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{}
}

func (m *CoreModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

func (m *CoreModule) SwaggerPaths() []string {
	return []string{
		"/clients",
		"/cost-providers",
		"/sessions",
		"/client-invoices",
		"/extra-efforts",
	}
}

// clientManagementRouteAdapter adapts the client management routes.RouteProvider to core.RouteProvider
type clientManagementRouteAdapter struct {
	provider *routes.RouteProvider
}

func (a *clientManagementRouteAdapter) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	a.provider.RegisterRoutes(router, &ctx)
}

func (a *clientManagementRouteAdapter) GetPrefix() string {
	return a.provider.GetPrefix()
}

func (a *clientManagementRouteAdapter) GetMiddleware() []gin.HandlerFunc {
	return a.provider.GetMiddleware()
}

func (a *clientManagementRouteAdapter) GetSwaggerTags() []string {
	return a.provider.GetSwaggerTags()
}
