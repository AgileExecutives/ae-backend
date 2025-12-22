package booking

import (
	"context"

	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	bookingHandler := handlers.NewBookingHandler(bookingService, bookingLinkService, freeSlotsSvc, db)

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

	// Initialize services
	m.bookingService = services.NewBookingService(ctx.DB)

	// Try to get BookingLinkService from service registry (created by Factory)
	// If not available, create it directly
	if bookingLinkSvcRaw, ok := ctx.Services.Get("booking-link-service"); ok {
		ctx.Logger.Info("✅ Initialize: Found BookingLinkService in service registry")
		if bookingLinkSvc, ok := bookingLinkSvcRaw.(*services.BookingLinkService); ok {
			m.bookingLinkService = bookingLinkSvc
			ctx.Logger.Info("✅ Initialize: Using BookingLinkService from registry")
		} else {
			ctx.Logger.Error("❌ Initialize: BookingLinkService type assertion failed")
		}
	}

	// If service not in registry, create it (shouldn't happen in normal flow)
	if m.bookingLinkService == nil {
		ctx.Logger.Warn("⚠️ Initialize: BookingLinkService not in registry, creating directly")
		if ctx.TokenService != nil {
			ctx.Logger.Info("Using unified TokenService for booking link generation")
			tokenServiceAdapter := &tokenServiceAdapter{service: ctx.TokenService}
			m.bookingLinkService = services.NewBookingLinkServiceWithTokenService(ctx.DB, tokenServiceAdapter)
		} else {
			jwtSecret := "booking-link-secret-key-change-in-production"
			ctx.Logger.Warn("TokenService not available - using legacy booking token implementation")
			m.bookingLinkService = services.NewBookingLinkService(ctx.DB, jwtSecret)
		}
	}

	m.freeSlotsSvc = services.NewFreeSlotsService(ctx.DB)

	// Initialize middleware
	m.tokenMiddleware = middleware.NewBookingTokenMiddleware(m.bookingLinkService, ctx.DB)

	// Initialize handlers
	m.bookingHandler = handlers.NewBookingHandler(m.bookingService, m.bookingLinkService, m.freeSlotsSvc, ctx.DB)

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
	return []core.ServiceProvider{
		&bookingLinkServiceProvider{module: m},
	}
}

// bookingLinkServiceProvider implements core.ServiceProvider for BookingLinkService
type bookingLinkServiceProvider struct {
	module *Module
}

func (p *bookingLinkServiceProvider) ServiceName() string {
	return "booking-link-service"
}

func (p *bookingLinkServiceProvider) ServiceInterface() interface{} {
	return p.module.bookingLinkService
}

func (p *bookingLinkServiceProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	// Create service with unified TokenService if available
	if ctx.TokenService != nil {
		ctx.Logger.Info("✅ Factory: Creating BookingLinkService with unified TokenService")
		tokenServiceAdapter := &tokenServiceAdapter{service: ctx.TokenService}
		p.module.bookingLinkService = services.NewBookingLinkServiceWithTokenService(ctx.DB, tokenServiceAdapter)
	} else {
		// Fallback to legacy implementation (should not happen in modern setup)
		ctx.Logger.Warn("⚠️ Factory: TokenService not available - using legacy booking token implementation")
		jwtSecret := "booking-link-secret-key-change-in-production"
		p.module.bookingLinkService = services.NewBookingLinkService(ctx.DB, jwtSecret)
	}
	ctx.Logger.Info("✅ Factory: BookingLinkService created and stored in module")
	return p.module.bookingLinkService, nil
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

// tokenServiceAdapter adapts the base TokenService to BookingLinkService interface
type tokenServiceAdapter struct {
	service core.TokenService
}

func (a *tokenServiceAdapter) GenerateToken(claims jwt.Claims) (string, error) {
	return a.service.GenerateToken(interface{}(claims))
}

func (a *tokenServiceAdapter) ValidateToken(tokenString string, claims jwt.Claims) error {
	return a.service.ValidateToken(tokenString, interface{}(claims))
}
