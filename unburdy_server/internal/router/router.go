package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	// Import ae-saas-basic public API
	baseAPI "github.com/ae-base-server/api"

	// Import unburdy-specific modules
	"github.com/unburdy/unburdy-server-api/internal/handlers"
	"github.com/unburdy/unburdy-server-api/internal/services"
)

// SetupExtendedRouter creates a router with ALL base endpoints PLUS client management
// This ensures unburdy_server has the exact same API as base-server + additional features
func SetupExtendedRouter(db *gorm.DB) *gin.Engine {
	// Start with the complete base router that has all the endpoints
	r := baseAPI.SetupBaseRouterWithConfig(db)

	// Initialize unburdy-specific services and handlers
	clientService := services.NewClientService(db)
	clientHandler := handlers.NewClientHandler(clientService)

	costProviderService := services.NewCostProviderService(db)
	costProviderHandler := handlers.NewCostProviderHandler(costProviderService)

	// Add unburdy-specific protected routes to the existing router
	protected := r.Group("/api/v1")
	protected.Use(baseAPI.AuthMiddleware(db))
	{
		// Client management endpoints (authenticated)
		clients := protected.Group("/clients")
		{
			clients.POST("", clientHandler.CreateClient)
			clients.GET("", clientHandler.GetAllClients)
			clients.GET("/search", clientHandler.SearchClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.PUT("/:id", clientHandler.UpdateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
		}

		// Cost provider management endpoints (authenticated)
		costProviders := protected.Group("/cost-providers")
		{
			costProviders.POST("", costProviderHandler.CreateCostProvider)
			costProviders.GET("", costProviderHandler.GetAllCostProviders)
			costProviders.GET("/search", costProviderHandler.SearchCostProviders)
			costProviders.GET("/:id", costProviderHandler.GetCostProvider)
			costProviders.PUT("/:id", costProviderHandler.UpdateCostProvider)
			costProviders.DELETE("/:id", costProviderHandler.DeleteCostProvider)
		}

		// Add a status endpoint to show unburdy extensions
		protected.GET("/unburdy/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Unburdy Extended API",
				"features": gin.H{
					"client_management":    "✓ Available",
					"base_api_integration": "✓ Complete",
					"all_base_endpoints":   "✓ Available",
					"calendar_module":      "✓ Available",
				},
			})
		})
	}

	return r
}
