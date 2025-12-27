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
		h.logger.Error("Invalid event payload type", "event", event)
		return fmt.Errorf("invalid event payload type")
	}

	// Try to extract calendar_entry_id - handle both uint and float64 types
	var calendarEntryID uint
	switch v := payload["calendar_entry_id"].(type) {
	case uint:
		calendarEntryID = v
	case float64:
		calendarEntryID = uint(v)
	case int:
		calendarEntryID = uint(v)
	default:
		h.logger.Error("calendar_entry_id not found or invalid type in event payload", "type", fmt.Sprintf("%T", v), "payload", payload)
		return fmt.Errorf("calendar_entry_id not found or invalid type in event payload")
	}

	h.logger.Info("Processing calendar entry deleted event", "calendar_entry_id", calendarEntryID)
	fmt.Printf("Processing calendar entry deleted event: calendar_entry_id=%d\n", calendarEntryID)
	// Update all sessions with this calendar_entry_id:
	// - Set status to 'canceled'
	// - Set calendar_entry_id to NULL (unlink from deleted entry)
	// - Add cancellation note to documentation
	result := h.db.Table("sessions").
		Where("calendar_entry_id = ? AND status = ?", calendarEntryID, "scheduled").
		Updates(map[string]interface{}{
			"status":            "canceled",
			"calendar_entry_id": nil,
			"documentation":     gorm.Expr("CASE WHEN documentation = '' OR documentation IS NULL THEN ? ELSE CONCAT(documentation, ?, ?) END", "Calendar entry deleted", "\n", "Calendar entry deleted"),
		})

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
