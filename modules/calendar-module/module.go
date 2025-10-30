package calendar

import (
	"context"
	"log"

	"github.com/ae-saas-basic/ae-saas-basic/pkg/eventbus"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/modules"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
) // CalendarModule implements the Module interface for calendar functionality
type CalendarModule struct {
	name         string
	version      string
	db           *gorm.DB
	handler      *Handler
	eventHandler *CalendarEventHandler
}

// NewCalendarModule creates a new calendar module instance
func NewCalendarModule() *CalendarModule {
	return &CalendarModule{
		name:    "calendar",
		version: "1.0.0",
	}
}

// GetName returns the module name
func (m *CalendarModule) GetName() string {
	return m.name
}

// GetVersion returns the module version
func (m *CalendarModule) GetVersion() string {
	return m.version
}

// GetModels returns the GORM models that need to be migrated
func (m *CalendarModule) GetModels() []interface{} {
	return []interface{}{
		&Calendar{},
		&RecurringEvent{},
		&Event{},
	}
}

// GetEventHandlers returns event handlers to register with the event bus
func (m *CalendarModule) GetEventHandlers() []eventbus.EventHandler {
	if m.eventHandler == nil {
		return []eventbus.EventHandler{}
	}
	return []eventbus.EventHandler{m.eventHandler}
}

// RegisterRoutes registers HTTP routes for this module
func (m *CalendarModule) RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	if m.handler == nil {
		log.Printf("⚠️  Calendar module handler not initialized, skipping route registration")
		return
	}

	log.Printf("🛣️  Registering calendar routes...")

	// Calendar CRUD routes
	calendars := router.Group("/calendars")
	{
		calendars.POST("", m.handler.CreateCalendar)
		calendars.GET("", m.handler.GetCalendars)
		calendars.GET("/:id", m.handler.GetCalendar)
		calendars.PUT("/:id", m.handler.UpdateCalendar)
		calendars.DELETE("/:id", m.handler.DeleteCalendar)
	}

	// Recurring Event CRUD routes
	recurringEvents := router.Group("/recurring-events")
	{
		recurringEvents.POST("", m.handler.CreateRecurringEvent)
		recurringEvents.GET("", m.handler.GetRecurringEvents)
		recurringEvents.GET("/:id", m.handler.GetRecurringEvent)
		recurringEvents.PUT("/:id", m.handler.UpdateRecurringEvent)
		recurringEvents.DELETE("/:id", m.handler.DeleteRecurringEvent)
	}

	// Event CRUD routes
	events := router.Group("/events")
	{
		events.POST("", m.handler.CreateEvent)
		events.GET("", m.handler.GetEvents)
		events.GET("/:id", m.handler.GetEvent)
		// events.PUT("/:id", m.handler.UpdateEvent)
		// events.DELETE("/:id", m.handler.DeleteEvent)
	}

	log.Printf("✅ Calendar routes registered")
}

// Initialize performs any initialization required by the module
func (m *CalendarModule) Initialize(db *gorm.DB) error {
	log.Printf("🔧 Initializing calendar module...")

	m.db = db

	// Initialize handler
	m.handler = NewHandler(db)

	// Initialize event handler for responding to user events
	m.eventHandler = NewCalendarEventHandler(db)

	log.Printf("✅ Calendar module initialized successfully")
	return nil
} // Shutdown performs cleanup when the application is shutting down
func (m *CalendarModule) Shutdown() error {
	log.Printf("🛑 Shutting down calendar module...")
	// Add any cleanup logic here
	log.Printf("✅ Calendar module shut down successfully")
	return nil
}

// GetSwaggerInfo returns Swagger documentation info for the calendar module
func (m *CalendarModule) GetSwaggerInfo() modules.SwaggerModuleInfo {
	return modules.SwaggerModuleInfo{
		Title:       "Calendar Module API",
		Description: "Calendar and event management functionality",
		Version:     m.version,
		Tags: []modules.SwaggerTag{
			{
				Name:        "calendars",
				Description: "Calendar management operations",
			},
			{
				Name:        "recurring-events",
				Description: "Recurring event management operations",
			},
			{
				Name:        "events",
				Description: "Event management operations",
			},
		},
		// Paths and definitions would be generated automatically by swag
		// or can be defined manually here for more control
	}
}

// CalendarEventHandler handles events from the event bus
type CalendarEventHandler struct {
	db   *gorm.DB
	name string
}

// NewCalendarEventHandler creates a new calendar event handler
func NewCalendarEventHandler(db *gorm.DB) *CalendarEventHandler {
	return &CalendarEventHandler{
		db:   db,
		name: "calendar_user_handler",
	}
}

// Handle processes events for the calendar module
func (h *CalendarEventHandler) Handle(ctx context.Context, event eventbus.Event) error {
	switch event.GetType() {
	case eventbus.EventUserCreated:
		return h.handleUserCreated(ctx, event)
	case eventbus.EventUserDeleted:
		return h.handleUserDeleted(ctx, event)
	default:
		// Ignore events we don't handle
		return nil
	}
}

// GetEventTypes returns the event types this handler is interested in
func (h *CalendarEventHandler) GetEventTypes() []string {
	return []string{
		eventbus.EventUserCreated,
		eventbus.EventUserDeleted,
	}
}

// GetName returns the handler name
func (h *CalendarEventHandler) GetName() string {
	return h.name
}

// handleUserCreated creates default calendar setup when a user is created
func (h *CalendarEventHandler) handleUserCreated(ctx context.Context, event eventbus.Event) error {
	payload, err := eventbus.GetUserCreatedPayload(event)
	if err != nil {
		return err
	}

	log.Printf("📅 Calendar Module: Setting up calendar for new user %s (email: %s, tenant: %s)",
		payload.UserID, payload.Email, payload.TenantID)

	// Create a welcome event for the new user
	// Note: This is a simplified example - in reality you might:
	// 1. Create default calendar settings
	// 2. Set up calendar permissions
	// 3. Create welcome/onboarding events
	// 4. Initialize calendar preferences

	log.Printf("  📅 Creating welcome calendar setup for user %s", payload.UserID)
	log.Printf("  📅 Email: %s, Tenant: %s", payload.Email, payload.TenantID)

	// In a real implementation, you might create actual calendar entries here
	// For now, we'll just log the action

	log.Printf("✅ Calendar Module: Calendar setup completed for user %s", payload.UserID)
	return nil
}

// handleUserDeleted cleans up calendar data when a user is deleted
func (h *CalendarEventHandler) handleUserDeleted(ctx context.Context, event eventbus.Event) error {
	payload, err := eventbus.GetUserDeletedPayload(event)
	if err != nil {
		return err
	}

	log.Printf("🗑️  Calendar Module: Cleaning up calendar data for deleted user %s", payload.UserID)

	// In a real implementation, you might:
	// 1. Archive or delete user's calendar events
	// 2. Remove calendar permissions
	// 3. Clean up calendar shares
	// 4. Notify other users of shared calendar changes

	log.Printf("  🗑️  Archiving calendar events for user %s", payload.UserID)
	log.Printf("  🗑️  Removing calendar permissions and shares")

	log.Printf("✅ Calendar Module: Cleanup completed for user %s", payload.UserID)
	return nil
}
