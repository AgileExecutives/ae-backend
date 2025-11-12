package booking

import (
	"context"

	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/booking-module/entities"
	"github.com/unburdy/booking-module/handlers"
	"github.com/unburdy/booking-module/middleware"
	"github.com/unburdy/booking-module/routes"
	"github.com/unburdy/booking-module/services"
)

// Module implements the complete core.Module interface for auto-migration support
type Module struct {
	db                 *gorm.DB
	routeProvider      *routes.RouteProvider
	bookingService     *services.BookingService
	bookingLinkService *services.BookingLinkService
	freeSlotsSvc       *services.FreeSlotsService
	tokenMiddleware    *middleware.BookingTokenMiddleware
	bookingHandler     *handlers.BookingHandler
}

// NewCoreModule creates a new booking module for the bootstrap system
// Initialization happens during the Initialize() lifecycle method
func NewCoreModule() *Module {
	return &Module{}
}

// NewModuleWithAutoMigration creates a new booking module with auto-migration support
func NewModuleWithAutoMigration(db *gorm.DB, jwtSecret string) *Module {
	// Initialize services
	bookingService := services.NewBookingService(db)
	bookingLinkService := services.NewBookingLinkService(db, jwtSecret)
	freeSlotsSvc := services.NewFreeSlotsService(db)

	// Initialize middleware
	tokenMiddleware := middleware.NewBookingTokenMiddleware(bookingLinkService, db)

	// Initialize handlers
	bookingHandler := handlers.NewBookingHandler(bookingService, bookingLinkService, freeSlotsSvc)

	// Initialize route provider with database for auth middleware
	routeProvider := routes.NewRouteProvider(bookingHandler, tokenMiddleware, db)

	return &Module{
		db:                 db,
		routeProvider:      routeProvider,
		bookingService:     bookingService,
		bookingLinkService: bookingLinkService,
		freeSlotsSvc:       freeSlotsSvc,
		tokenMiddleware:    tokenMiddleware,
		bookingHandler:     bookingHandler,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "booking"
}

// Version returns the module version
func (m *Module) Version() string {
	return "1.0.0"
}

// Dependencies returns module dependencies
func (m *Module) Dependencies() []string {
	return []string{"base", "calendar"} // Depends on base module for users/tenants and calendar for calendar entities
}

// Initialize initializes the module
func (m *Module) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing booking module...")

	// Store database reference
	m.db = ctx.DB

	// Use a JWT secret (should be configured via environment variable in production)
	// For now, use a default that should be replaced in production
	jwtSecret := "booking-link-secret-key-change-in-production"
	ctx.Logger.Warn("Using default JWT secret for booking links. Set JWT_SECRET environment variable in production.")

	// Initialize services
	m.bookingService = services.NewBookingService(ctx.DB)
	m.bookingLinkService = services.NewBookingLinkService(ctx.DB, jwtSecret)
	m.freeSlotsSvc = services.NewFreeSlotsService(ctx.DB)

	// Initialize middleware
	m.tokenMiddleware = middleware.NewBookingTokenMiddleware(m.bookingLinkService, ctx.DB)

	// Initialize handlers
	m.bookingHandler = handlers.NewBookingHandler(m.bookingService, m.bookingLinkService, m.freeSlotsSvc)

	// Initialize route provider with database for auth middleware
	m.routeProvider = routes.NewRouteProvider(m.bookingHandler, m.tokenMiddleware, ctx.DB)

	ctx.Logger.Info("Booking module initialized successfully")
	return nil
}

// Start starts the module
func (m *Module) Start(ctx context.Context) error {
	// Any startup logic here
	return nil
}

// Stop stops the module
func (m *Module) Stop(ctx context.Context) error {
	// Any cleanup logic here
	return nil
}

// Entities returns database entities for auto-migration
func (m *Module) Entities() []core.Entity {
	return []core.Entity{
		entities.NewBookingTemplateEntity(),
	}
}

// GetEntitiesForMigration returns GORM models for auto-migration (implements ModuleWithEntities interface)
func (m *Module) GetEntitiesForMigration() []interface{} {
	return []interface{}{
		&entities.BookingTemplate{},
	}
}

// Routes returns route providers
func (m *Module) Routes() []core.RouteProvider {
	if m.routeProvider == nil {
		return []core.RouteProvider{}
	}
	return []core.RouteProvider{
		&bookingRouteAdapter{
			provider: m.routeProvider,
		},
	}
}

// EventHandlers returns event handlers
func (m *Module) EventHandlers() []core.EventHandler {
	return []core.EventHandler{}
}

// Middleware returns middleware providers
func (m *Module) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

// Services returns service providers
func (m *Module) Services() []core.ServiceProvider {
	return []core.ServiceProvider{}
}

// SwaggerPaths returns Swagger documentation paths
func (m *Module) SwaggerPaths() []string {
	return []string{
		"/booking/templates",
		"/booking/templates/{id}",
		"/booking/link",
	}
}

// Legacy support methods for compatibility with existing baseAPI.ModuleRouteProvider interface

// RegisterRoutes implements compatibility with baseAPI.ModuleRouteProvider
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	m.routeProvider.RegisterRoutes(router)
}

// GetPrefix implements compatibility with baseAPI.ModuleRouteProvider
func (m *Module) GetPrefix() string {
	return m.routeProvider.GetPrefix()
}

// bookingRouteAdapter adapts the booking routes.RouteProvider to core.RouteProvider
type bookingRouteAdapter struct {
	provider *routes.RouteProvider
}

func (a *bookingRouteAdapter) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	a.provider.RegisterRoutes(router)
}

func (a *bookingRouteAdapter) GetPrefix() string {
	return a.provider.GetPrefix()
}

func (a *bookingRouteAdapter) GetMiddleware() []gin.HandlerFunc {
	// Middleware is handled by the route provider itself
	return a.provider.GetMiddleware()
}

func (a *bookingRouteAdapter) GetSwaggerTags() []string {
	return a.provider.GetSwaggerTags()
}
