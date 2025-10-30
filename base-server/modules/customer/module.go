package customer

import (
	"context"

	"github.com/ae-base-server/modules/customer/entities"
	"github.com/ae-base-server/modules/customer/events"
	"github.com/ae-base-server/modules/customer/handlers"
	"github.com/ae-base-server/modules/customer/services"
	"github.com/ae-base-server/pkg/core"
)

// CustomerModule provides customer and plan management functionality
type CustomerModule struct {
	customerHandlers *handlers.CustomerHandlers
	planHandlers     *handlers.PlanHandlers
	customerService  *services.CustomerService
	eventHandlers    *events.CustomerEventHandlers
}

// NewCustomerModule creates a new customer module instance
func NewCustomerModule() core.Module {
	return &CustomerModule{}
}

func (m *CustomerModule) Name() string {
	return "customer"
}

func (m *CustomerModule) Version() string {
	return "1.0.0"
}

func (m *CustomerModule) Dependencies() []string {
	return []string{"base"} // Depends on base module for auth
}

func (m *CustomerModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing customer module...")

	// Initialize services
	m.customerService = services.NewCustomerService(ctx.DB, ctx.Logger)

	// Initialize handlers
	m.customerHandlers = handlers.NewCustomerHandlers(ctx.DB, ctx.Logger)
	m.planHandlers = handlers.NewPlanHandlers(ctx.DB, ctx.Logger)

	// Initialize event handlers
	m.eventHandlers = events.NewCustomerEventHandlers(ctx.EventBus, ctx.Logger)

	ctx.Logger.Info("Customer module initialized successfully")
	return nil
}

func (m *CustomerModule) Start(ctx context.Context) error {
	// Start any background services if needed
	return nil
}

func (m *CustomerModule) Stop(ctx context.Context) error {
	// Stop any background services if needed
	return nil
}

func (m *CustomerModule) Entities() []core.Entity {
	return []core.Entity{
		entities.NewCustomerEntity(),
		entities.NewPlanEntity(),
	}
}

func (m *CustomerModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{
		// Temporarily disabled - using internal handlers instead
		// handlers.NewCustomerRoutes(m.customerHandlers),
		// handlers.NewPlanRoutes(m.planHandlers),
	}
}

func (m *CustomerModule) EventHandlers() []core.EventHandler {
	return []core.EventHandler{
		events.NewCustomerCreatedHandler(m.eventHandlers),
		events.NewPlanSubscribedHandler(m.eventHandlers),
	}
}

func (m *CustomerModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{
		services.NewCustomerServiceProvider(m.customerService),
	}
}

func (m *CustomerModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{} // No custom middleware for now
}

func (m *CustomerModule) SwaggerPaths() []string {
	return []string{
		"./modules/customer/handlers",
		"./modules/customer/entities",
	}
}
