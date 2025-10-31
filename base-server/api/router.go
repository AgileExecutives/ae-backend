// Package api provides public access to ae-base-server router setup
package api

import (
	"github.com/ae-base-server/internal/router"
	"github.com/ae-base-server/pkg/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupBaseRouter sets up the complete base router with all endpoints
// This allows external modules to get the full base API setup
func SetupBaseRouter(db *gorm.DB, cfg config.Config) *gin.Engine {
	return router.SetupRouter(db, cfg)
}

// SetupBaseRouterWithConfig sets up the base router with auto-loaded config
// Convenience function that loads config automatically
func SetupBaseRouterWithConfig(db *gorm.DB) *gin.Engine {
	cfg := config.Load()
	return router.SetupRouter(db, cfg)
}

// CreateProtectedRouterGroup creates a router group with authentication middleware applied
// This allows external modules to add authenticated routes to the base router
func CreateProtectedRouterGroup(baseRouter *gin.Engine, db *gorm.DB, prefix string) *gin.RouterGroup {
	group := baseRouter.Group(prefix)
	group.Use(AuthMiddleware(db))
	return group
}
