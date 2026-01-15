package base

import (
	"context"

	"github.com/ae-base-server/modules/base/entities"
	"github.com/ae-base-server/modules/base/events"
	"github.com/ae-base-server/modules/base/handlers"
	"github.com/ae-base-server/modules/base/middleware"
	"github.com/ae-base-server/modules/base/services"
	"github.com/ae-base-server/pkg/core"
	settingsentities "github.com/ae-base-server/pkg/settings/entities"
)

// BaseModule provides core authentication, user management, and contact functionality
type BaseModule struct {
	authHandlers         *handlers.AuthHandlers
	contactHandlers      *handlers.ContactHandlers
	healthHandlers       *handlers.HealthHandlers
	userSettingsHandlers *handlers.UserSettingsHandlers
	authService          *services.AuthService
	eventHandlers        *events.BaseEventHandlers
	authMiddleware       *middleware.AuthMiddleware
	moduleContext        core.ModuleContext
}

// NewBaseModule creates a new base module instance
func NewBaseModule() core.Module {
	return &BaseModule{}
}

func (m *BaseModule) Name() string {
	return "base"
}

func (m *BaseModule) Version() string {
	return "1.0.0"
}

func (m *BaseModule) Dependencies() []string {
	return []string{} // No dependencies - this is the base module
}

func (m *BaseModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing base module...")

	// Store context for later use
	m.moduleContext = ctx

	// Initialize services
	m.authService = services.NewAuthService(ctx.DB, ctx.Logger)

	// Initialize handlers
	m.authHandlers = handlers.NewAuthHandlers(ctx.DB, ctx.Logger)
	m.contactHandlers = handlers.NewContactHandlers(ctx.DB, ctx.Logger)
	m.healthHandlers = handlers.NewHealthHandlers(ctx.DB, ctx.Logger)
	m.userSettingsHandlers = handlers.NewUserSettingsHandlers(ctx.DB, ctx.Logger)

	// Initialize event handlers
	m.eventHandlers = events.NewBaseEventHandlers(ctx.EventBus, ctx.Logger)

	// Initialize middleware
	m.authMiddleware = middleware.NewAuthMiddleware(ctx.DB, ctx.Logger)

	ctx.Logger.Info("Base module initialized successfully")
	return nil
}

func (m *BaseModule) Start(ctx context.Context) error {
	// Set module registry in auth handler now that all modules are initialized
	// This needs to be done in Start() because it happens after all Initialize() calls
	if m.authHandlers != nil && m.moduleContext.ModuleRegistry != nil {
		m.authHandlers.SetModuleRegistry(m.moduleContext.ModuleRegistry)
		m.moduleContext.Logger.Info("Base module started - auth handlers configured with module registry for template seeding")
	} else {
		m.moduleContext.Logger.Warn("Base module started - module registry not available for template seeding")
	}
	return nil
}

func (m *BaseModule) Stop(ctx context.Context) error {
	// Stop any background services if needed
	return nil
}

func (m *BaseModule) Entities() []core.Entity {
	return []core.Entity{
		entities.NewUserEntity(),
		entities.NewTenantEntity(),
		entities.NewContactEntity(),
		entities.NewNewsletterEntity(),
		entities.NewTokenBlacklistEntity(),
		entities.NewUserSettingsEntity(),
		settingsentities.NewSettingDefinitionEntity(),
		settingsentities.NewSettingEntity(),
	}
}

func (m *BaseModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{
		handlers.NewAuthRoutes(m.authHandlers),
		handlers.NewContactRoutes(m.contactHandlers),
		handlers.NewUserSettingsRoutes(m.userSettingsHandlers),
		handlers.NewHealthRoutes(m.healthHandlers),
	}
}

func (m *BaseModule) EventHandlers() []core.EventHandler {
	return []core.EventHandler{
		events.NewUserCreatedHandler(m.eventHandlers),
		events.NewUserLoginHandler(m.eventHandlers),
		events.NewContactFormSubmittedHandler(m.eventHandlers),
	}
}

func (m *BaseModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{
		services.NewAuthServiceProvider(m.authService),
	}
}

func (m *BaseModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{
		middleware.NewAuthMiddlewareProvider(m.authMiddleware),
	}
}

func (m *BaseModule) SwaggerPaths() []string {
	return []string{
		"./modules/base/handlers",
		"./modules/base/entities",
	}
}
