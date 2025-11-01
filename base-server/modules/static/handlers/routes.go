package handlers

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StaticRoutes provides route registration for static file serving
type StaticRoutes struct {
	handlers *StaticHandlers
	db       *gorm.DB
}

// NewStaticRoutes creates a new static routes instance
func NewStaticRoutes(handlers *StaticHandlers, db *gorm.DB) core.RouteProvider {
	return &StaticRoutes{
		handlers: handlers,
		db:       db,
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

// GetMiddleware returns middleware for static routes (requires authentication)
func (r *StaticRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{middleware.AuthMiddleware(r.db)}
}

// GetSwaggerTags returns swagger tags for documentation
func (r *StaticRoutes) GetSwaggerTags() []string {
	return []string{"static"}
}
