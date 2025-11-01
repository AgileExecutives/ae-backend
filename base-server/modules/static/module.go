package static

import (
	"context"

	"github.com/ae-base-server/modules/static/handlers"
	"github.com/ae-base-server/pkg/core"
	"gorm.io/gorm"
)

// StaticModule provides secure JSON file serving functionality
type StaticModule struct {
	staticHandlers *handlers.StaticHandlers
	db             *gorm.DB
}

// NewStaticModule creates a new static module instance
func NewStaticModule() core.Module {
	return &StaticModule{}
}

func (m *StaticModule) Name() string {
	return "static"
}

func (m *StaticModule) Version() string {
	return "1.0.0"
}

func (m *StaticModule) Dependencies() []string {
	return []string{} // No dependencies
}

func (m *StaticModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing static module...")

	// Store database reference
	m.db = ctx.DB

	// Initialize handlers
	m.staticHandlers = handlers.NewStaticHandlers(ctx.Logger)

	ctx.Logger.Info("Static module initialized successfully")
	return nil
}

func (m *StaticModule) Start(ctx context.Context) error {
	// No background services needed
	return nil
}

func (m *StaticModule) Stop(ctx context.Context) error {
	// No background services to stop
	return nil
}

func (m *StaticModule) Entities() []core.Entity {
	// No database entities needed for static file serving
	return []core.Entity{}
}

func (m *StaticModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{
		handlers.NewStaticRoutes(m.staticHandlers, m.db),
	}
}

func (m *StaticModule) EventHandlers() []core.EventHandler {
	// No events needed for static file serving
	return []core.EventHandler{}
}

func (m *StaticModule) Services() []core.ServiceProvider {
	// No services needed for static file serving
	return []core.ServiceProvider{}
}

func (m *StaticModule) Middleware() []core.MiddlewareProvider {
	// No middleware needed for static file serving
	return []core.MiddlewareProvider{}
}

func (m *StaticModule) SwaggerPaths() []string {
	return []string{
		"./modules/static/handlers",
	}
}
