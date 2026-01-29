package audit

import (
	"context"

	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
)

type CoreModule struct {
	module *Module
}

func NewCoreModule() core.Module {
	return &CoreModule{}
}

func (m *CoreModule) Name() string {
	return "audit"
}

func (m *CoreModule) Version() string {
	return "1.0.0"
}

func (m *CoreModule) Dependencies() []string {
	return []string{"base"}
}

func (m *CoreModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing audit module...")

	m.module = NewModule(ctx.DB)

	if err := m.module.AutoMigrate(); err != nil {
		return err
	}

	ctx.Services.Register("audit-service", m.module.GetService())
	ctx.Logger.Info("Audit module initialized successfully")

	return nil
}

func (m *CoreModule) Start(ctx context.Context) error {
	return nil
}

func (m *CoreModule) Stop(ctx context.Context) error {
	return nil
}

func (m *CoreModule) Entities() []core.Entity {
	return []core.Entity{}
}

func (m *CoreModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{
		&AuditRouteProvider{module: m.module},
	}
}

func (m *CoreModule) EventHandlers() []core.EventHandler {
	return []core.EventHandler{}
}

func (m *CoreModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

func (m *CoreModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{}
}

func (m *CoreModule) SwaggerPaths() []string {
	return []string{}
}

type AuditRouteProvider struct {
	module *Module
}

func (r *AuditRouteProvider) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	if r.module != nil {
		r.module.RegisterRoutes(router)
	}
}

func (r *AuditRouteProvider) GetPrefix() string {
	return ""
}

func (r *AuditRouteProvider) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (r *AuditRouteProvider) GetSwaggerTags() []string {
	return []string{"audit"}
}
