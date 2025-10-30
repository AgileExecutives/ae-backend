package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/ae-saas-basic/ae-saas-basic/pkg/eventbus"
) // CalendarHandler is an example event handler that creates calendars when users are created
type CalendarHandler struct {
	name string
}

// NewCalendarHandler creates a new calendar event handler
func NewCalendarHandler() *CalendarHandler {
	return &CalendarHandler{
		name: "calendar_handler",
	}
}

// Handle processes events for the calendar plugin
func (h *CalendarHandler) Handle(ctx context.Context, event eventbus.Event) error {
	switch event.GetType() {
	case eventbus.EventUserCreated:
		return h.handleUserCreated(ctx, event)
	case eventbus.EventUserDeleted:
		return h.handleUserDeleted(ctx, event)
	default:
		return fmt.Errorf("unhandled event type: %s", event.GetType())
	}
}

// GetEventTypes returns the event types this handler is interested in
func (h *CalendarHandler) GetEventTypes() []string {
	return []string{
		eventbus.EventUserCreated,
		eventbus.EventUserDeleted,
	}
}

// GetName returns the handler name
func (h *CalendarHandler) GetName() string {
	return h.name
}

// handleUserCreated processes user created events by creating a calendar
func (h *CalendarHandler) handleUserCreated(ctx context.Context, event eventbus.Event) error {
	payload, err := eventbus.GetUserCreatedPayload(event)
	if err != nil {
		return fmt.Errorf("failed to get user created payload: %w", err)
	}

	log.Printf("Calendar Plugin: Creating calendar for user %s (email: %s, tenant: %s)",
		payload.UserID, payload.Email, payload.TenantID)

	// Simulate calendar creation logic
	if err := h.createCalendar(ctx, payload); err != nil {
		return fmt.Errorf("failed to create calendar for user %s: %w", payload.UserID, err)
	}

	log.Printf("Calendar Plugin: Successfully created calendar for user %s", payload.UserID)
	return nil
}

// handleUserDeleted processes user deleted events by removing the calendar
func (h *CalendarHandler) handleUserDeleted(ctx context.Context, event eventbus.Event) error {
	payload, err := eventbus.GetUserDeletedPayload(event)
	if err != nil {
		return fmt.Errorf("failed to get user deleted payload: %w", err)
	}

	log.Printf("Calendar Plugin: Removing calendar for deleted user %s (email: %s, tenant: %s)",
		payload.UserID, payload.Email, payload.TenantID)

	// Simulate calendar removal logic
	if err := h.removeCalendar(ctx, payload); err != nil {
		return fmt.Errorf("failed to remove calendar for user %s: %w", payload.UserID, err)
	}

	log.Printf("Calendar Plugin: Successfully removed calendar for user %s", payload.UserID)
	return nil
}

// createCalendar simulates creating a calendar for a user
func (h *CalendarHandler) createCalendar(ctx context.Context, payload *eventbus.UserCreatedPayload) error {
	// In a real implementation, this would:
	// 1. Connect to calendar service (Google Calendar, Outlook, etc.)
	// 2. Create a new calendar for the user
	// 3. Set up default events or configurations
	// 4. Store calendar metadata in database

	// For demonstration purposes, we'll just simulate the work
	log.Printf("  - Connecting to calendar service for tenant %s", payload.TenantID)
	log.Printf("  - Creating calendar with name: '%s's Calendar'", payload.Email)
	log.Printf("  - Setting up default calendar permissions")
	log.Printf("  - Saving calendar metadata to database")

	// Simulate some processing time
	// time.Sleep(100 * time.Millisecond)

	return nil
}

// removeCalendar simulates removing a calendar for a user
func (h *CalendarHandler) removeCalendar(ctx context.Context, payload *eventbus.UserDeletedPayload) error {
	// In a real implementation, this would:
	// 1. Connect to calendar service
	// 2. Find user's calendar
	// 3. Archive or delete calendar data
	// 4. Remove calendar metadata from database

	// For demonstration purposes, we'll just simulate the work
	log.Printf("  - Connecting to calendar service for tenant %s", payload.TenantID)
	log.Printf("  - Finding calendar for user %s", payload.UserID)
	log.Printf("  - Archiving calendar data")
	log.Printf("  - Removing calendar metadata from database")

	return nil
}
