package routes

import (
	"github.com/ae-base-server/internal/organizations/handlers"
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RouteProvider provides routing functionality for organization management
type RouteProvider struct {
	handler *handlers.OrganizationHandler
	db      *gorm.DB
}

// NewRouteProvider creates a new route provider
func NewRouteProvider(handler *handlers.OrganizationHandler, db *gorm.DB) *RouteProvider {
	return &RouteProvider{
		handler: handler,
		db:      db,
	}
}

// RegisterRoutes registers the organization management routes
func (rp *RouteProvider) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	authMiddleware := middleware.AuthMiddleware(rp.db)

	organizations := router.Group("/organizations")
	organizations.Use(authMiddleware)
	{
		organizations.GET("/supported-formats", rp.handler.GetSupportedFormats)
		organizations.POST("", rp.handler.CreateOrganization)
		organizations.GET("", rp.handler.GetAllOrganizations)
		organizations.GET("/:id", rp.handler.GetOrganization)
		organizations.PUT("/:id", rp.handler.UpdateOrganization)
		organizations.DELETE("/:id", rp.handler.DeleteOrganization)
	}
}

// GetPrefix returns the route prefix for organization endpoints
func (rp *RouteProvider) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware to apply to all routes
func (rp *RouteProvider) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

// GetSwaggerTags returns swagger tags for the routes
func (rp *RouteProvider) GetSwaggerTags() []string {
	return []string{"organizations"}
}
