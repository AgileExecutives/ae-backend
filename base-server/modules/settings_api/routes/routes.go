package routes

import (
	"github.com/ae-base-server/modules/settings_api/handlers"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
)

// RouteProvider provides routing functionality for settings API.
type RouteProvider struct {
	handler *handlers.SettingsHandler
}

// NewRouteProvider creates a new route provider.
func NewRouteProvider(handler *handlers.SettingsHandler) *RouteProvider {
	return &RouteProvider{handler: handler}
}

// RegisterRoutes registers the settings routes.
func (rp *RouteProvider) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	settingsGroup := router.Group("/settings")
	{
		tenantGroup := settingsGroup.Group("/organizations/:tenant_id")
		{
			tenantGroup.GET("/domains/:domain", rp.handler.GetDomainSettings)
			tenantGroup.POST("/domains/:domain", rp.handler.UpdateDomainSettings)
			tenantGroup.PUT("/domains/:domain", rp.handler.UpdateDomainSettings)
			tenantGroup.GET("/domains/:domain/:key", rp.handler.GetSetting)
			tenantGroup.POST("/domains/:domain/:key", rp.handler.UpdateSetting)
			tenantGroup.PUT("/domains/:domain/:key", rp.handler.UpdateSetting)
		}
	}
}

func (rp *RouteProvider) GetPrefix() string {
	return ""
}

func (rp *RouteProvider) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (rp *RouteProvider) GetSwaggerTags() []string {
	return []string{"settings"}
}
