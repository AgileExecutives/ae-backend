package routes

import (
	"github.com/ae-base-server/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/unburdy-server-api/modules/client_management/handlers"
)

// RouteProvider provides routing functionality for client management
type RouteProvider struct {
	clientHandler       *handlers.ClientHandler
	costProviderHandler *handlers.CostProviderHandler
	sessionHandler      *handlers.SessionHandler
	db                  *gorm.DB
}

// NewRouteProvider creates a new route provider
func NewRouteProvider(clientHandler *handlers.ClientHandler, costProviderHandler *handlers.CostProviderHandler, sessionHandler *handlers.SessionHandler, db *gorm.DB) *RouteProvider {
	return &RouteProvider{
		clientHandler:       clientHandler,
		costProviderHandler: costProviderHandler,
		sessionHandler:      sessionHandler,
		db:                  db,
	}
}

// RegisterRoutes registers the client management routes with the provided router group
func (rp *RouteProvider) RegisterRoutes(router *gin.RouterGroup) {
	// Note: Auth middleware is already applied at the router group level
	// Client management endpoints (authenticated)
	clients := router.Group("/clients")
	{
		clients.POST("", rp.clientHandler.CreateClient)
		clients.GET("", rp.clientHandler.GetAllClients)
		clients.GET("/search", rp.clientHandler.SearchClients)
		clients.GET("/:id/sessions", rp.sessionHandler.GetSessionsByClient)
		clients.GET("/:id", rp.clientHandler.GetClient)
		clients.PUT("/:id", rp.clientHandler.UpdateClient)
		clients.DELETE("/:id", rp.clientHandler.DeleteClient)
	}

	// Cost provider management endpoints (authenticated)
	costProviders := router.Group("/cost-providers")
	{
		costProviders.POST("", rp.costProviderHandler.CreateCostProvider)
		costProviders.GET("", rp.costProviderHandler.GetAllCostProviders)
		costProviders.GET("/search", rp.costProviderHandler.SearchCostProviders)
		costProviders.GET("/:id", rp.costProviderHandler.GetCostProvider)
		costProviders.PUT("/:id", rp.costProviderHandler.UpdateCostProvider)
		costProviders.DELETE("/:id", rp.costProviderHandler.DeleteCostProvider)
	}

	// Session management endpoints (authenticated)
	sessions := router.Group("/sessions")
	{
		sessions.POST("", rp.sessionHandler.CreateSession)
		sessions.POST("/book", rp.sessionHandler.BookSessions)
		sessions.GET("", rp.sessionHandler.GetAllSessions)
		sessions.GET("/:id", rp.sessionHandler.GetSession)
		sessions.PUT("/:id", rp.sessionHandler.UpdateSession)
		sessions.DELETE("/:id", rp.sessionHandler.DeleteSession)
	}
}

// GetPrefix returns the route prefix for client management endpoints
func (rp *RouteProvider) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware to apply to all routes
func (rp *RouteProvider) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthMiddleware(rp.db), // Require authentication for all client management routes
	}
}

// GetSwaggerTags returns swagger tags for the routes
func (rp *RouteProvider) GetSwaggerTags() []string {
	return []string{"clients", "cost-providers", "sessions"}
}
