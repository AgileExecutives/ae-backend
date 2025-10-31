// Package api provides router extension capabilities
package api

import (
	"github.com/ae-base-server/internal/router"
	"github.com/ae-base-server/pkg/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RouterExtensionFunc is a function type for extending the base router
// Uses interface{} for config to avoid internal package exposure
type RouterExtensionFunc func(*gin.Engine, *gorm.DB, interface{})

// SetupExtendedRouter sets up the base router and allows extensions
// This enables unburdy_server to add its own routes to the base API
func SetupExtendedRouter(db *gorm.DB, cfg config.Config, extensions ...RouterExtensionFunc) *gin.Engine {
	// Get the base router with all standard endpoints
	r := router.SetupRouter(db, cfg)

	// Apply any extensions
	for _, extension := range extensions {
		extension(r, db, cfg)
	}

	return r
}

// SetupExtendedRouterWithConfig is a convenience function with auto-loaded config
func SetupExtendedRouterWithConfig(db *gorm.DB, extensions ...RouterExtensionFunc) *gin.Engine {
	cfg := config.Load()
	return SetupExtendedRouter(db, cfg, extensions...)
}
