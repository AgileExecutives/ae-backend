// Package api provides public access to module registration and routing
package api

import (
	"fmt"
	"log"

	"github.com/ae-base-server/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ModuleRouteProvider defines how external modules can register their routes
type ModuleRouteProvider interface {
	RegisterRoutes(router *gin.RouterGroup)
	GetPrefix() string
}

// ModuleWithEntities extends ModuleRouteProvider with auto-migration support
// Modules implementing this interface will have their entities automatically migrated
type ModuleWithEntities interface {
	ModuleRouteProvider
	GetEntitiesForMigration() []interface{} // Returns GORM models for auto-migration
}

// CreateProtectedRouterGroup creates a router group with authentication middleware
func CreateProtectedRouterGroup(baseRouter *gin.Engine, db *gorm.DB, prefix string) *gin.RouterGroup {
	group := baseRouter.Group(prefix)
	group.Use(middleware.AuthMiddleware(db))
	return group
}

// SetupBaseRouterWithConfig creates a basic Gin router with common middleware
func SetupBaseRouterWithConfig(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	return router
}

// RegisterModuleRoutes registers routes for an external module with authentication
func RegisterModuleRoutes(baseRouter *gin.Engine, db *gorm.DB, module ModuleRouteProvider) {
	// Create protected group with the module's prefix
	protectedGroup := CreateProtectedRouterGroup(baseRouter, db, "/api/v1"+module.GetPrefix())

	// Register the module's routes
	module.RegisterRoutes(protectedGroup)
}

// SetupModularRouter creates a base router and allows modules to register routes
// Modules implementing ModuleWithEntities will have their entities auto-migrated
func SetupModularRouter(db *gorm.DB, modules []ModuleRouteProvider) *gin.Engine {
	// Create base router with all base functionality
	baseRouter := SetupBaseRouterWithConfig(db)

	// Auto-migrate entities from modules that support it
	for _, module := range modules {
		if moduleWithEntities, ok := module.(ModuleWithEntities); ok {
			entities := moduleWithEntities.GetEntitiesForMigration()
			if len(entities) > 0 {
				log.Printf("Auto-migrating entities for module %T...", module)
				if err := db.AutoMigrate(entities...); err != nil {
					log.Printf("Failed to migrate entities for module %T: %v", module, err)
					// Don't panic - log error and continue
				} else {
					log.Printf("Successfully migrated %d entities for module %T", len(entities), module)
				}
			}
		}
	}

	// Register all external modules
	for _, module := range modules {
		RegisterModuleRoutes(baseRouter, db, module)
	}

	return baseRouter
}

// MigrateModules explicitly migrates entities from modules that support it
// This can be called separately if you want more control over when migration occurs
func MigrateModules(db *gorm.DB, modules []ModuleRouteProvider) error {
	for _, module := range modules {
		if moduleWithEntities, ok := module.(ModuleWithEntities); ok {
			entities := moduleWithEntities.GetEntitiesForMigration()
			if len(entities) > 0 {
				if err := db.AutoMigrate(entities...); err != nil {
					return fmt.Errorf("failed to migrate entities for module %T: %w", module, err)
				}
				log.Printf("Successfully migrated %d entities for module %T", len(entities), module)
			}
		}
	}
	return nil
}
