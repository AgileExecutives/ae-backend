package organization

import (
	"context"

	"github.com/ae-base-server/internal/organizations/handlers"
	"github.com/ae-base-server/internal/organizations/routes"
	"github.com/ae-base-server/internal/organizations/services"
	"github.com/ae-base-server/modules/organization/entities"
	"github.com/ae-base-server/pkg/core"
)

// OrganizationModule represents the organization management module
type OrganizationModule struct {
	handler       *handlers.OrganizationHandler
	service       *services.OrganizationService
	routeProvider *routes.RouteProvider
}

// NewOrganizationModule creates a new organization module instance
func NewOrganizationModule() core.Module {
	return &OrganizationModule{}
}

func (m *OrganizationModule) Name() string {
	return "organization"
}

func (m *OrganizationModule) Version() string {
	return "1.0.0"
}

func (m *OrganizationModule) Dependencies() []string {
	return []string{"base"}
}

// Initialize sets up the organization module
func (m *OrganizationModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing organization module...")

	// Create service
	m.service = services.NewOrganizationService(ctx.DB)

	// Create handler
	m.handler = handlers.NewOrganizationHandler(m.service)

	// Create route provider
	m.routeProvider = routes.NewRouteProvider(m.handler, ctx.DB)

	ctx.Logger.Info("Organization module initialized successfully")
	return nil
}

func (m *OrganizationModule) Start(ctx context.Context) error {
	return nil
}

func (m *OrganizationModule) Stop(ctx context.Context) error {
	return nil
}

func (m *OrganizationModule) Entities() []core.Entity {
	return []core.Entity{
		entities.NewOrganizationEntity(),
	}
}

func (m *OrganizationModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{
		m.routeProvider,
	}
}

func (m *OrganizationModule) EventHandlers() []core.EventHandler {
	return []core.EventHandler{}
}

func (m *OrganizationModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{}
}

func (m *OrganizationModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

func (m *OrganizationModule) SwaggerPaths() []string {
	return []string{
		"./internal/organizations/handlers",
		"./internal/models",
	}
}
