package events

import (
	"github.com/ae-base-server/pkg/core"
	"gorm.io/gorm"
)

// EmailSentHandler handles email sent events
type EmailSentHandler struct {
	db *gorm.DB
}

// NewEmailSentHandler creates a new email sent event handler
func NewEmailSentHandler(db *gorm.DB) *EmailSentHandler {
	return &EmailSentHandler{db: db}
}

// EventType returns the event type this handler processes
func (h *EmailSentHandler) EventType() string {
	return "email.sent"
}

// Handle processes the email sent event
func (h *EmailSentHandler) Handle(event interface{}) error {
	// Handle email sent event logic
	// For example, update delivery status, log metrics, etc.
	return nil
}

// Priority returns the handler priority
func (h *EmailSentHandler) Priority() int {
	return 100
}

// EmailFailedHandler handles email failed events
type EmailFailedHandler struct {
	db     *gorm.DB
	logger core.Logger
}

// NewEmailFailedHandler creates a new email failed event handler
func NewEmailFailedHandler(db *gorm.DB, logger core.Logger) *EmailFailedHandler {
	return &EmailFailedHandler{
		db:     db,
		logger: logger,
	}
}

// EventType returns the event type this handler processes
func (h *EmailFailedHandler) EventType() string {
	return "email.failed"
}

// Handle processes the email failed event
func (h *EmailFailedHandler) Handle(event interface{}) error {
	// Handle email failed event logic
	// For example, retry logic, error logging, alerting, etc.
	h.logger.Error("Email sending failed", "event", event)
	return nil
}

// Priority returns the handler priority
func (h *EmailFailedHandler) Priority() int {
	return 100
}

// EmailQueuedEvent represents an email queued event
type EmailQueuedEvent struct {
	EmailID uint   `json:"email_id"`
	To      string `json:"to"`
	Subject string `json:"subject"`
}

// EmailSentEvent represents an email sent event
type EmailSentEvent struct {
	EmailID uint   `json:"email_id"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	SentAt  string `json:"sent_at"`
}

// EmailFailedEvent represents an email failed event
type EmailFailedEvent struct {
	EmailID uint   `json:"email_id"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Error   string `json:"error"`
}
