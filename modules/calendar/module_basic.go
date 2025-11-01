package calendar

import (
	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/calendar-module/handlers"
	"github.com/unburdy/calendar-module/routes"
	"github.com/unburdy/calendar-module/services"
)

// BasicModule implements the baseAPI.ModuleRouteProvider interface (legacy compatibility)
type BasicModule struct {
	routeProvider *routes.RouteProvider
}

// NewBasicModule creates a new basic calendar module (legacy compatibility)
func NewBasicModule(db *gorm.DB) baseAPI.ModuleRouteProvider {
	// Initialize services
	calendarService := services.NewCalendarService(db)

	// Initialize handlers
	calendarHandler := handlers.NewCalendarHandler(calendarService)

	// Initialize route provider
	routeProvider := routes.NewRouteProvider(calendarHandler)

	return &BasicModule{
		routeProvider: routeProvider,
	}
}

// RegisterRoutes implements baseAPI.ModuleRouteProvider
func (m *BasicModule) RegisterRoutes(router *gin.RouterGroup) {
	// Directly call the method to avoid any interface conflicts
	m.routeProvider.RegisterRoutes(router)
}

// GetPrefix implements baseAPI.ModuleRouteProvider
func (m *BasicModule) GetPrefix() string {
	return m.routeProvider.GetPrefix()
}

// NewModule creates a new calendar module (legacy compatibility - calls NewBasicModule)
// Deprecated: Use NewBasicModule for basic routing or NewFullModule for auto-migration
func NewModule(db *gorm.DB) baseAPI.ModuleRouteProvider {
	return NewBasicModule(db)
}
