package settingsapi

import (
	"context"

	"github.com/ae-base-server/modules/settings_api/entities"
	"github.com/ae-base-server/modules/settings_api/handlers"
	"github.com/ae-base-server/modules/settings_api/routes"
	"github.com/ae-base-server/pkg/core"
)

// SettingsAPIModule exposes a small tenant-scoped settings HTTP API.
//
// This module exists to keep Unburdy's legacy settings endpoints available while
// the newer settings system is being integrated.
type SettingsAPIModule struct {
	handler       *handlers.SettingsHandler
	routeProvider *routes.RouteProvider
}

// NewSettingsAPIModule creates a new settings API module instance.
func NewSettingsAPIModule() core.Module {
	return &SettingsAPIModule{}
}

func (m *SettingsAPIModule) Name() string { return "settings_api" }

func (m *SettingsAPIModule) Version() string { return "1.0.0" }

func (m *SettingsAPIModule) Dependencies() []string { return []string{"base"} }

func (m *SettingsAPIModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing settings_api module...")

	m.handler = handlers.NewSettingsHandler(ctx.DB)
	m.routeProvider = routes.NewRouteProvider(m.handler)

	ctx.Logger.Info("settings_api module initialized successfully")
	return nil
}

func (m *SettingsAPIModule) Start(ctx context.Context) error { return nil }

func (m *SettingsAPIModule) Stop(ctx context.Context) error { return nil }

func (m *SettingsAPIModule) Entities() []core.Entity {
	return []core.Entity{
		entities.NewSettingDefinitionEntity(),
		entities.NewSettingEntity(),
	}
}

func (m *SettingsAPIModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{m.routeProvider}
}

func (m *SettingsAPIModule) EventHandlers() []core.EventHandler { return []core.EventHandler{} }

func (m *SettingsAPIModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

func (m *SettingsAPIModule) Services() []core.ServiceProvider { return []core.ServiceProvider{} }

func (m *SettingsAPIModule) SwaggerPaths() []string {
	return []string{
		"./modules/settings_api",
		"./pkg/settings",
	}
}
