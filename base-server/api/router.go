// Package api provides public access to ae-base-server router setup
package api

import (
	"github.com/ae-base-server/internal/config"
	"github.com/ae-base-server/internal/router"
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
