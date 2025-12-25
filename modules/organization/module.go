package organization

import (
"context"

"github.com/ae-base-server/pkg/core"
"github.com/gin-gonic/gin"
"github.com/unburdy/organization-module/entities"
"github.com/unburdy/organization-module/handlers"
"github.com/unburdy/organization-module/routes"
"github.com/unburdy/organization-module/services"
"gorm.io/gorm"
)

// Module implements the core.Module interface
type Module struct {
	db            *gorm.DB
	handler       *handlers.OrganizationHandler
	service       *services.OrganizationService
	routeProvider *routes.RouteProvider
}

// NewCoreModule creates a new organization module
func NewCoreModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "organization"
}

func (m *Module) Version() string {
	return "1.0.0"
}

func (m *Module) Dependencies() []string {
	return []string{"base"}
}

func (m *Module) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing organization module...")
	m.db = ctx.DB
	m.service = services.NewOrganizationService(m.db)
	m.handler = handlers.NewOrganizationHandler(m.service)
	m.routeProvider = routes.NewRouteProvider(m.handler, m.db)
	ctx.Logger.Info("Organization module initialized successfully")
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	// Start method for lifecycle management
	return nil
}

func (m *Module) Stop(ctx context.Context) error {
	return nil
}

func (m *Module) Entities() []core.Entity {
	return []core.Entity{
		entities.NewOrganizationEntity(),
	}
}

func (m *Module) Routes() []core.RouteProvider {
	if m.routeProvider == nil {
		return []core.RouteProvider{}
	}
	return []core.RouteProvider{
		&organizationRouteAdapter{provider: m.routeProvider},
	}
}

func (m *Module) EventHandlers() []core.EventHandler {
	return []core.EventHandler{}
}

func (m *Module) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

func (m *Module) Services() []core.ServiceProvider {
	return []core.ServiceProvider{}
}

func (m *Module) SwaggerPaths() []string {
	return []string{
		"/organizations",
		"/organizations/{id}",
	}
}

// Legacy support methods

func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	m.routeProvider.RegisterRoutes(router)
}

func (m *Module) GetPrefix() string {
	return m.routeProvider.GetPrefix()
}

// GetService returns the organization service for use by other modules
func (m *Module) GetService() *services.OrganizationService {
	return m.service
}

// GetEntitiesForMigration returns GORM models for auto-migration
func (m *Module) GetEntitiesForMigration() []interface{} {
	return []interface{}{
		&entities.Organization{},
	}
}

// organizationRouteAdapter adapts the organization routes.RouteProvider to core.RouteProvider
type organizationRouteAdapter struct {
	provider *routes.RouteProvider
}

func (a *organizationRouteAdapter) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	a.provider.RegisterRoutes(router)
}

func (a *organizationRouteAdapter) GetPrefix() string {
	return a.provider.GetPrefix()
}

func (a *organizationRouteAdapter) GetMiddleware() []gin.HandlerFunc {
	return a.provider.GetMiddleware()
}

func (a *organizationRouteAdapter) GetSwaggerTags() []string {
	return a.provider.GetSwaggerTags()
}
