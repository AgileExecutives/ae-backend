package routes

import (
	"github.com/ae-base-server/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/booking-module/handlers"
	bookingMiddleware "github.com/unburdy/booking-module/middleware"
)

// RouteProvider provides routing functionality for booking management
type RouteProvider struct {
	bookingHandler  *handlers.BookingHandler
	tokenMiddleware *bookingMiddleware.BookingTokenMiddleware
	db              *gorm.DB
}

// NewRouteProvider creates a new route provider
func NewRouteProvider(bookingHandler *handlers.BookingHandler, tokenMiddleware *bookingMiddleware.BookingTokenMiddleware, db *gorm.DB) *RouteProvider {
	return &RouteProvider{
		bookingHandler:  bookingHandler,
		tokenMiddleware: tokenMiddleware,
		db:              db,
	}
}

// RegisterRoutes registers the booking management routes with the provided router group
func (rp *RouteProvider) RegisterRoutes(router *gin.RouterGroup) {
	// Public endpoints (no authentication) - just token validation
	router.GET("/booking/freeslots/:token", rp.tokenMiddleware.ValidateBookingToken(), rp.bookingHandler.GetFreeSlots)
	router.GET("/client/:token", rp.tokenMiddleware.ValidateBookingToken(), rp.bookingHandler.GetClientByToken)

	// Authenticated endpoints - wrap in auth middleware
	authMiddleware := middleware.AuthMiddleware(rp.db)

	// Booking templates/configurations CRUD endpoints (authenticated)
	templates := router.Group("/booking/templates")
	templates.Use(authMiddleware)
	{
		templates.POST("", rp.bookingHandler.CreateConfiguration)
		templates.GET("", rp.bookingHandler.GetAllConfigurations)
		templates.GET("/:id", rp.bookingHandler.GetConfiguration)
		templates.PUT("/:id", rp.bookingHandler.UpdateConfiguration)
		templates.DELETE("/:id", rp.bookingHandler.DeleteConfiguration)

		// Additional query endpoints
		templates.GET("/by-user", rp.bookingHandler.GetConfigurationsByUser)
		templates.GET("/by-calendar", rp.bookingHandler.GetConfigurationsByCalendar)
	}

	// Booking link generation endpoint (authenticated)
	router.POST("/booking/link", authMiddleware, rp.bookingHandler.CreateBookingLink)
}

// GetPrefix returns the route prefix for booking management endpoints
func (rp *RouteProvider) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware to apply to all routes
// Returns empty array since we apply auth selectively in RegisterRoutes
func (rp *RouteProvider) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

// GetSwaggerTags returns swagger tags for the routes
func (rp *RouteProvider) GetSwaggerTags() []string {
	return []string{"booking"}
}
