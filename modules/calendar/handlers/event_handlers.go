package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/ae-base-server/pkg/eventbus"
	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/services"
)

// CalendarEventHandler handles calendar-related events
type CalendarEventHandler struct {
	calendarService *services.CalendarService
}

// NewCalendarEventHandler creates a new calendar event handler
func NewCalendarEventHandler(calendarService *services.CalendarService) *CalendarEventHandler {
	return &CalendarEventHandler{
		calendarService: calendarService,
	}
}

// EventType returns the event type this handler listens to
func (h *CalendarEventHandler) EventType() string {
	return eventbus.EventUserCreated
}

// Handle handles incoming events
func (h *CalendarEventHandler) Handle(event interface{}) error {
	return h.handleUserCreated(event)
}

// Priority returns the handler priority (higher runs first)
func (h *CalendarEventHandler) Priority() int {
	return 100 // Default priority
}

// handleUserCreated creates a default calendar when a user is created
func (h *CalendarEventHandler) handleUserCreated(event interface{}) error {
	log.Println("📅 CalendarEventHandler: Received UserCreated event")

	// Parse the event payload
	var payload eventbus.UserCreatedPayload
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("❌ CalendarEventHandler: Failed to marshal event payload: %v", err)
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal user created payload: %w", err)
	}

	// Convert string IDs to uint
	userID, err := strconv.ParseUint(payload.UserID, 10, 32)
	if err != nil {
		return fmt.Errorf("failed to parse user ID: %w", err)
	}

	tenantID, err := strconv.ParseUint(payload.TenantID, 10, 32)
	if err != nil {
		return fmt.Errorf("failed to parse tenant ID: %w", err)
	}

	log.Printf("📅 Creating default calendar for user %d (email: %s, tenant: %d)",
		userID, payload.Email, tenantID)

	// Create default calendar request
	req := entities.CreateCalendarRequest{
		Title:    "My Calendar",
		Color:    "#4285F4", // Google Calendar blue
		Timezone: "Europe/Berlin",
		WeeklyAvailability: json.RawMessage(`{
			"monday": {"enabled": true, "slots": [{"start": "09:00", "end": "17:00"}]},
			"tuesday": {"enabled": true, "slots": [{"start": "09:00", "end": "17:00"}]},
			"wednesday": {"enabled": true, "slots": [{"start": "09:00", "end": "17:00"}]},
			"thursday": {"enabled": true, "slots": [{"start": "09:00", "end": "17:00"}]},
			"friday": {"enabled": true, "slots": [{"start": "09:00", "end": "17:00"}]},
			"saturday": {"enabled": false, "slots": []},
			"sunday": {"enabled": false, "slots": []}
		}`),
	}

	// Create the calendar using the service
	calendar, err := h.calendarService.CreateCalendar(req, uint(tenantID), uint(userID))
	if err != nil {
		return fmt.Errorf("failed to create default calendar for user %d: %w", userID, err)
	}

	log.Printf("✅ Successfully created default calendar (ID: %d) for user %d (email: %s)",
		calendar.ID, userID, payload.Email)

	return nil
}
