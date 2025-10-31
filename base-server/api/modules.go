// Package api provides public access to module registration and routing
package api

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ModuleRouteProvider defines how external modules can register their routes
type ModuleRouteProvider interface {
	RegisterRoutes(router *gin.RouterGroup)
	GetPrefix() string
}

// RegisterModuleRoutes registers routes for an external module with authentication
func RegisterModuleRoutes(baseRouter *gin.Engine, db *gorm.DB, module ModuleRouteProvider) {
	// Create protected group with the module's prefix
	protectedGroup := CreateProtectedRouterGroup(baseRouter, db, "/api/v1"+module.GetPrefix())

	// Register the module's routes
	module.RegisterRoutes(protectedGroup)
}

// SetupModularRouter creates a base router and allows modules to register routes
func SetupModularRouter(db *gorm.DB, modules []ModuleRouteProvider) *gin.Engine {
	// Create base router with all base functionality
	baseRouter := SetupBaseRouterWithConfig(db)

	// Register all external modules
	for _, module := range modules {
		RegisterModuleRoutes(baseRouter, db, module)
	}

	return baseRouter
}
