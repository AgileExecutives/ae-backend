package routes

import (
	"github.com/ae-base-server/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/calendar-module/handlers"
)

// RouteProvider provides routing functionality for calendar management
type RouteProvider struct {
	calendarHandler *handlers.CalendarHandler
	db              *gorm.DB
}

// NewRouteProvider creates a new route provider
func NewRouteProvider(calendarHandler *handlers.CalendarHandler, db *gorm.DB) *RouteProvider {
	return &RouteProvider{
		calendarHandler: calendarHandler,
		db:              db,
	}
}

// RegisterRoutes registers the calendar management routes with the provided router group
func (rp *RouteProvider) RegisterRoutes(router *gin.RouterGroup) {
	// Calendar CRUD endpoints (authenticated)
	calendar := router.Group("/calendars")
	{
		calendar.POST("", rp.calendarHandler.CreateCalendar)
		calendar.GET("", rp.calendarHandler.GetCalendarsWithMetadata)
		calendar.GET("/:id", rp.calendarHandler.GetCalendar)
		calendar.PUT("/:id", rp.calendarHandler.UpdateCalendar)
		calendar.DELETE("/:id", rp.calendarHandler.DeleteCalendar)

		// Specialized calendar endpoints
		calendar.GET("/week", rp.calendarHandler.GetCalendarWeekView)
		calendar.GET("/year", rp.calendarHandler.GetCalendarYearView)
		calendar.GET("/free-slots", rp.calendarHandler.GetFreeSlots)
		calendar.POST("/:id/import_holidays", rp.calendarHandler.ImportHolidays)
	}

	// Calendar Entry CRUD endpoints (authenticated)
	entries := router.Group("/calendar-entries")
	{
		entries.POST("", rp.calendarHandler.CreateCalendarEntry)
		entries.GET("", rp.calendarHandler.GetAllCalendarEntries)
		entries.GET("/:id", rp.calendarHandler.GetCalendarEntry)
		entries.PUT("/:id", rp.calendarHandler.UpdateCalendarEntry)
		entries.DELETE("/:id", rp.calendarHandler.DeleteCalendarEntry)
	}

	// Calendar Series CRUD endpoints (authenticated)
	series := router.Group("/calendar-series")
	{
		series.POST("", rp.calendarHandler.CreateCalendarSeries)
		series.GET("", rp.calendarHandler.GetAllCalendarSeries)
		series.GET("/:id", rp.calendarHandler.GetCalendarSeries)
		series.PUT("/:id", rp.calendarHandler.UpdateCalendarSeries)
		series.DELETE("/:id", rp.calendarHandler.DeleteCalendarSeries)
	}

	// External Calendar CRUD endpoints (authenticated)
	external := router.Group("/external-calendars")
	{
		external.POST("", rp.calendarHandler.CreateExternalCalendar)
		external.GET("", rp.calendarHandler.GetAllExternalCalendars)
		external.GET("/:id", rp.calendarHandler.GetExternalCalendar)
		external.PUT("/:id", rp.calendarHandler.UpdateExternalCalendar)
		external.DELETE("/:id", rp.calendarHandler.DeleteExternalCalendar)
	}
}

// GetPrefix returns the route prefix for calendar management endpoints
func (rp *RouteProvider) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware to apply to all routes
func (rp *RouteProvider) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthMiddleware(rp.db), // Require authentication for all calendar routes
	}
}

// GetSwaggerTags returns swagger tags for the routes
func (rp *RouteProvider) GetSwaggerTags() []string {
	return []string{"calendar", "calendar-entries", "calendar-series", "external-calendars", "calendar-views", "calendar-availability", "calendar-utilities"}
}
