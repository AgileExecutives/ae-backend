package calendar

import (
	"context"

	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/handlers"
	"github.com/unburdy/calendar-module/routes"
	"github.com/unburdy/calendar-module/services"
)

// Module implements the complete core.Module interface for auto-migration support
type Module struct {
	db              *gorm.DB
	routeProvider   *routes.RouteProvider
	calendarService *services.CalendarService
	calendarHandler *handlers.CalendarHandler
}

// NewCoreModule creates a new calendar module for the bootstrap system
// Initialization happens during the Initialize() lifecycle method
func NewCoreModule() *Module {
	return &Module{}
}

// NewModuleWithAutoMigration creates a new calendar module with auto-migration support
func NewModuleWithAutoMigration(db *gorm.DB) *Module {
	// Initialize services
	calendarService := services.NewCalendarService(db)

	// Initialize handlers
	calendarHandler := handlers.NewCalendarHandler(calendarService)

	// Initialize route provider with database for auth middleware
	routeProvider := routes.NewRouteProvider(calendarHandler, db)

	return &Module{
		db:              db,
		routeProvider:   routeProvider,
		calendarService: calendarService,
		calendarHandler: calendarHandler,
	}
}

// Name returns the module name
func (m *Module) Name() string {
	return "calendar"
}

// Version returns the module version
func (m *Module) Version() string {
	return "1.0.0"
}

// Dependencies returns module dependencies
func (m *Module) Dependencies() []string {
	return []string{"base"} // Depends on base module for users/tenants
}

// Initialize initializes the module
func (m *Module) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing calendar module...")

	// Store database reference
	m.db = ctx.DB

	// Initialize services
	m.calendarService = services.NewCalendarService(ctx.DB)
	m.calendarService.SetEventBus(ctx.EventBus)

	// Initialize handlers
	m.calendarHandler = handlers.NewCalendarHandler(m.calendarService)

	// Initialize route provider with database for auth middleware
	m.routeProvider = routes.NewRouteProvider(m.calendarHandler, ctx.DB)

	ctx.Logger.Info("Calendar module initialized successfully")
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
		entities.NewCalendarEntity(),
		entities.NewCalendarEntryEntity(),
		entities.NewCalendarSeriesEntity(),
		entities.NewExternalCalendarEntity(),
	}
}

// GetEntitiesForMigration returns GORM models for auto-migration (implements ModuleWithEntities interface)
func (m *Module) GetEntitiesForMigration() []interface{} {
	coreEntities := m.Entities() // Get core entities
	entities := make([]interface{}, len(coreEntities))
	for i, entity := range coreEntities {
		entities[i] = entity.GetModel()
	}
	return entities
}

// Routes returns route providers
func (m *Module) Routes() []core.RouteProvider {
	if m.routeProvider == nil {
		return []core.RouteProvider{}
	}
	return []core.RouteProvider{
		&calendarRouteAdapter{
			provider: m.routeProvider,
			// ctx will be nil here but set during Initialize
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
		"/calendar",
		"/calendar/{id}/import_holidays",
		"/calendar-entries",
		"/calendar-series",
		"/external-calendars",
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

// calendarRouteAdapter adapts the calendar routes.RouteProvider to core.RouteProvider
type calendarRouteAdapter struct {
	provider *routes.RouteProvider
}

func (a *calendarRouteAdapter) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	a.provider.RegisterRoutes(router)
}

func (a *calendarRouteAdapter) GetPrefix() string {
	return a.provider.GetPrefix()
}

func (a *calendarRouteAdapter) GetMiddleware() []gin.HandlerFunc {
	// Middleware is handled by the route provider itself
	return a.provider.GetMiddleware()
}

func (a *calendarRouteAdapter) GetSwaggerTags() []string {
	return a.provider.GetSwaggerTags()
}
