// Package calendar provides calendar and event management functionality
// This package can be imported and used by any application built on top of ae-saas-basic
package calendar

// Re-export main types and functions for easy importing
type (
	CalendarEvent   = Event
	CalendarHandler = Handler
)

// Main package exports
var (
	NewCalendarHandler           = NewHandler
	RegisterCalendarRoutes       = RegisterRoutes
	RegisterCalendarPublicRoutes = RegisterPublicRoutes
	MigrateCalendar              = Migrate
)
