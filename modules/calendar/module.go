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

// FullModule implements the complete core.Module interface for auto-migration support
type FullModule struct {
	db              *gorm.DB
	routeProvider   *routes.RouteProvider
	calendarService *services.CalendarService
	calendarHandler *handlers.CalendarHandler
}

// NewFullModule creates a new calendar module with auto-migration support
func NewFullModule(db *gorm.DB) *FullModule {
	// Initialize services
	calendarService := services.NewCalendarService(db)

	// Initialize handlers
	calendarHandler := handlers.NewCalendarHandler(calendarService)

	// Initialize route provider
	routeProvider := routes.NewRouteProvider(calendarHandler)

	return &FullModule{
		db:              db,
		routeProvider:   routeProvider,
		calendarService: calendarService,
		calendarHandler: calendarHandler,
	}
}

// Name returns the module name
func (m *FullModule) Name() string {
	return "calendar"
}

// Version returns the module version
func (m *FullModule) Version() string {
	return "1.0.0"
}

// Dependencies returns module dependencies
func (m *FullModule) Dependencies() []string {
	return []string{"base"} // Depends on base module for users/tenants
}

// Initialize initializes the module
func (m *FullModule) Initialize(ctx core.ModuleContext) error {
	// Any initialization logic here
	return nil
}

// Start starts the module
func (m *FullModule) Start(ctx context.Context) error {
	// Any startup logic here
	return nil
}

// Stop stops the module
func (m *FullModule) Stop(ctx context.Context) error {
	// Any cleanup logic here
	return nil
}

// Entities returns database entities for auto-migration
func (m *FullModule) Entities() []core.Entity {
	return []core.Entity{
		entities.NewCalendarEntity(),
		entities.NewCalendarEntryEntity(),
		entities.NewCalendarSeriesEntity(),
		entities.NewExternalCalendarEntity(),
	}
}

// GetEntitiesForMigration returns GORM models for auto-migration (implements ModuleWithEntities interface)
func (m *FullModule) GetEntitiesForMigration() []interface{} {
	coreEntities := m.Entities() // Get core entities
	entities := make([]interface{}, len(coreEntities))
	for i, entity := range coreEntities {
		entities[i] = entity.GetModel()
	}
	return entities
}

// Routes returns route providers
func (m *FullModule) Routes() []core.RouteProvider {
	// For now, return empty slice as we use the simple interface
	return []core.RouteProvider{}
}

// EventHandlers returns event handlers
func (m *FullModule) EventHandlers() []core.EventHandler {
	return []core.EventHandler{}
}

// Middleware returns middleware providers
func (m *FullModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

// Services returns service providers
func (m *FullModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{}
}

// SwaggerPaths returns Swagger documentation paths
func (m *FullModule) SwaggerPaths() []string {
	return []string{
		"/calendar",
		"/calendar-entries",
		"/calendar-series",
		"/external-calendars",
	}
}

// Legacy support methods for compatibility with existing baseAPI.ModuleRouteProvider interface

// RegisterRoutes implements compatibility with baseAPI.ModuleRouteProvider
func (m *FullModule) RegisterRoutes(router *gin.RouterGroup) {
	m.routeProvider.RegisterRoutes(router)
}

// GetPrefix implements compatibility with baseAPI.ModuleRouteProvider
func (m *FullModule) GetPrefix() string {
	return m.routeProvider.GetPrefix()
}
