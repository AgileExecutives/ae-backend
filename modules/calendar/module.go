package calendar

import (
	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/calendar-module/handlers"
	"github.com/unburdy/calendar-module/routes"
	"github.com/unburdy/calendar-module/services"
)

// Module implements the baseAPI.ModuleRouteProvider interface
type Module struct {
	routeProvider *routes.RouteProvider
}

// NewModule creates a new calendar module
func NewModule(db *gorm.DB) baseAPI.ModuleRouteProvider {
	// Initialize services
	calendarService := services.NewCalendarService(db)

	// Initialize handlers
	calendarHandler := handlers.NewCalendarHandler(calendarService)

	// Initialize route provider
	routeProvider := routes.NewRouteProvider(calendarHandler)

	return &Module{
		routeProvider: routeProvider,
	}
}

// RegisterRoutes implements baseAPI.ModuleRouteProvider
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	// Directly call the method to avoid any interface conflicts
	m.routeProvider.RegisterRoutes(router)
}

// GetPrefix implements baseAPI.ModuleRouteProvider
func (m *Module) GetPrefix() string {
	return m.routeProvider.GetPrefix()
}
