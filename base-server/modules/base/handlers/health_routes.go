package handlers

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
)

// HealthRoutes provides health check routes
type HealthRoutes struct {
	handlers *HealthHandlers
}

// NewHealthRoutes creates new health routes
func NewHealthRoutes(handlers *HealthHandlers) core.RouteProvider {
	return &HealthRoutes{
		handlers: handlers,
	}
}

func (r *HealthRoutes) GetPrefix() string {
	return ""
}

func (r *HealthRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (r *HealthRoutes) GetSwaggerTags() []string {
	return []string{"health"}
}

func (r *HealthRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	// Register health check endpoints at root level
	router.GET("/health", r.handlers.HealthCheck)
	router.GET("/ping", r.handlers.Ping)
}
