package main

import (
	"net/http"

	"github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PingModule represents the ping module
type PingModule struct {
	db *gorm.DB
}

// NewPingModule creates a new ping module instance
func NewPingModule(db *gorm.DB) *PingModule {
	return &PingModule{db: db}
}

// RegisterRoutes implements the ModuleRouteProvider interface
// Note: All routes registered here will be protected by authentication
func (m *PingModule) RegisterRoutes(router *gin.RouterGroup) {
	// Protected ping endpoint (requires authentication)
	router.GET("/protected-ping", m.handleProtectedPing)
}

// RegisterPublicRoutes allows the module to register public routes on the base router
func (m *PingModule) RegisterPublicRoutes(baseRouter *gin.Engine) {
	// Create a public ping group
	publicGroup := baseRouter.Group("/api/v1/ping")

	// Add public ping endpoint
	publicGroup.GET("/ping", m.handlePing)
}

// GetPrefix returns the module's URL prefix
func (m *PingModule) GetPrefix() string {
	return "/ping"
}

// handlePing handles the basic ping request
func (m *PingModule) handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"module":  "ping",
		"version": "1.0.0",
		"timestamp": gin.H{
			"unix": c.Request.Header.Get("X-Request-Time"),
		},
	})
}

// handleProtectedPing handles authenticated ping requests
func (m *PingModule) handleProtectedPing(c *gin.Context) {
	// This endpoint is automatically protected by the base-server auth middleware
	// We can access user context here if needed

	c.JSON(http.StatusOK, gin.H{
		"message":            "authenticated pong",
		"module":             "ping",
		"user_authenticated": true,
		"endpoints": gin.H{
			"public_ping":    "/api/v1/ping/ping (public)",
			"protected_ping": "/api/v1/ping/protected-ping (requires auth)",
		},
	})
}

// Ensure PingModule implements ModuleRouteProvider interface
var _ api.ModuleRouteProvider = (*PingModule)(nil)
