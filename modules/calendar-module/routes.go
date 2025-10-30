// Package calendar provides route registration for calendar module
package calendar

import (
	"github.com/ae-saas-basic/ae-saas-basic/api"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers calendar routes with authentication middleware
// This function allows easy integration into any Gin router
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	handler := NewHandler(db)

	// All calendar routes require authentication from base-server
	calendarGroup := router.Group("/calendar")
	calendarGroup.Use(api.AuthMiddleware(db)) // Use base-server auth middleware
	{
		// Event management routes
		events := calendarGroup.Group("/events")
		{
			events.POST("", handler.CreateEvent) // POST /api/v1/calendar/events
			events.GET("", handler.GetEvents)    // GET /api/v1/calendar/events
			events.GET("/:id", handler.GetEvent) // GET /api/v1/calendar/events/:id
			// Add more routes as needed:
			// events.PUT("/:id", handler.UpdateEvent)
			// events.DELETE("/:id", handler.DeleteEvent)
		}

		// Future calendar features can be added here:
		// calendarGroup.GET("/availability", handler.GetAvailability)
		// calendarGroup.POST("/recurring", handler.CreateRecurringEvent)
	}
}

// RegisterPublicRoutes registers public calendar routes (no authentication)
// Use this for routes that don't require authentication
func RegisterPublicRoutes(router *gin.RouterGroup, db *gorm.DB) {
	handler := NewHandler(db)

	calendarGroup := router.Group("/calendar")
	{
		// Public routes (if needed)
		// calendarGroup.GET("/public-events", handler.GetPublicEvents)
		// calendarGroup.GET("/availability/:user_id", handler.GetPublicAvailability)

		// For now, keep it empty as most calendar functionality requires auth
		_ = handler       // Avoid unused variable warning
		_ = calendarGroup // Avoid unused variable warning
	}
}

// Migrate runs database migrations for calendar module
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Event{})
}
