package client_management

import (
	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/unburdy-server-api/modules/client_management/handlers"
	"github.com/unburdy/unburdy-server-api/modules/client_management/routes"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// Module implements the baseAPI.ModuleRouteProvider interface
type Module struct {
	routeProvider *routes.RouteProvider
}

// NewModule creates a new client management module
func NewModule(db *gorm.DB) baseAPI.ModuleRouteProvider {
	// Initialize services
	clientService := services.NewClientService(db)
	costProviderService := services.NewCostProviderService(db)

	// Initialize handlers
	clientHandler := handlers.NewClientHandler(clientService)
	costProviderHandler := handlers.NewCostProviderHandler(costProviderService)

	// Initialize route provider
	routeProvider := routes.NewRouteProvider(clientHandler, costProviderHandler)

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
