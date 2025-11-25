package events

import (
	"fmt"

	"github.com/ae-base-server/pkg/core"
	"gorm.io/gorm"
)

// CalendarEntryDeletedHandler handles calendar entry deletion events
type CalendarEntryDeletedHandler struct {
	db     *gorm.DB
	logger core.Logger
}

// NewCalendarEntryDeletedHandler creates a new calendar entry deleted event handler
func NewCalendarEntryDeletedHandler(db *gorm.DB, logger core.Logger) core.EventHandler {
	return &CalendarEntryDeletedHandler{
		db:     db,
		logger: logger,
	}
}

// EventType returns the event type this handler subscribes to
func (h *CalendarEntryDeletedHandler) EventType() string {
	return "calendar.entry.deleted"
}

// Handle processes the calendar entry deleted event
func (h *CalendarEntryDeletedHandler) Handle(event interface{}) error {
	payload, ok := event.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event payload type")
	}

	calendarEntryID, ok := payload["calendar_entry_id"].(uint)
	if !ok {
		return fmt.Errorf("calendar_entry_id not found or invalid type in event payload")
	}

	h.logger.Info("Processing calendar entry deleted event", "calendar_entry_id", calendarEntryID)

	// Update all sessions with this calendar_entry_id to canceled status if they are scheduled
	result := h.db.Table("sessions").
		Where("calendar_entry_id = ? AND status = ?", calendarEntryID, "scheduled").
		Update("status", "canceled")

	if result.Error != nil {
		h.logger.Error("Failed to cancel sessions for deleted calendar entry", "error", result.Error, "calendar_entry_id", calendarEntryID)
		return fmt.Errorf("failed to cancel sessions: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		h.logger.Info("Canceled sessions for deleted calendar entry", "calendar_entry_id", calendarEntryID, "sessions_updated", result.RowsAffected)
	}

	return nil
}

// Priority returns the handler priority (higher = executed first)
func (h *CalendarEntryDeletedHandler) Priority() int {
	return 100 // Default priority
}
