package handlers

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
)

// StaticRoutes provides route registration for static file serving
type StaticRoutes struct {
	handlers *StaticHandlers
}

// NewStaticRoutes creates a new static routes instance
func NewStaticRoutes(handlers *StaticHandlers) core.RouteProvider {
	return &StaticRoutes{
		handlers: handlers,
	}
}

// RegisterRoutes registers all static file serving routes
func (r *StaticRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	// JSON file serving endpoints
	router.GET("/static", r.handlers.ListStaticJSON)
	router.GET("/static/:filename", r.handlers.ServeStaticJSON)
}

// GetPrefix returns the route prefix for this provider
func (r *StaticRoutes) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware for static routes (none needed)
func (r *StaticRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

// GetSwaggerTags returns swagger tags for documentation
func (r *StaticRoutes) GetSwaggerTags() []string {
	return []string{"static"}
}
