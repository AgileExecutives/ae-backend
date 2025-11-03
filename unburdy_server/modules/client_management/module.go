package client_management

import (
	"context"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/unburdy-server-api/internal/services"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/handlers"
	"github.com/unburdy/unburdy-server-api/modules/client_management/routes"
	moduleServices "github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// Module implements the baseAPI.ModuleRouteProvider interface
type Module struct {
	routeProvider *routes.RouteProvider
}

// NewModule creates a new client management module
func NewModule(db *gorm.DB) baseAPI.ModuleRouteProvider {
	// Initialize services (using internal for client, modular for cost provider)
	clientService := services.NewClientService(db)
	costProviderService := moduleServices.NewCostProviderService(db)

	// Initialize handlers
	clientHandler := handlers.NewClientHandler(clientService)
	costProviderHandler := handlers.NewCostProviderHandler(costProviderService)

	// Initialize route provider with database for auth middleware
	routeProvider := routes.NewRouteProvider(clientHandler, costProviderHandler, db)

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

// GetEntitiesForMigration implements baseAPI.ModuleWithEntities
func (m *Module) GetEntitiesForMigration() []interface{} {
	return []interface{}{
		&entities.Client{},
		&entities.CostProvider{},
	}
}

// CoreModule implements the core.Module interface for bootstrap system integration
type CoreModule struct {
	db                  *gorm.DB
	clientHandlers      *handlers.ClientHandler
	costProviderHandler *handlers.CostProviderHandler
	routeProvider       *routes.RouteProvider
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
	return []string{"base"} // Depends on base module for users/tenants
}

func (m *CoreModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing client management module...")
	m.db = ctx.DB

	// Initialize services (using internal for client, modular for cost provider)
	clientService := services.NewClientService(ctx.DB)
	costProviderService := moduleServices.NewCostProviderService(ctx.DB)

	// Initialize handlers
	m.clientHandlers = handlers.NewClientHandler(clientService)
	m.costProviderHandler = handlers.NewCostProviderHandler(costProviderService)

	// Initialize route provider with database for auth middleware
	m.routeProvider = routes.NewRouteProvider(m.clientHandlers, m.costProviderHandler, ctx.DB)

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
	return []core.EventHandler{}
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
	}
}

// clientManagementRouteAdapter adapts the client management routes.RouteProvider to core.RouteProvider
type clientManagementRouteAdapter struct {
	provider *routes.RouteProvider
}

func (a *clientManagementRouteAdapter) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	a.provider.RegisterRoutes(router)
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
