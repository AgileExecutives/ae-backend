package events

import (
	"github.com/ae-base-server/pkg/core"
)

// PDFGeneratedHandler handles PDF generated events
type PDFGeneratedHandler struct {
	logger core.Logger
}

// NewPDFGeneratedHandler creates a new PDF generated event handler
func NewPDFGeneratedHandler(logger core.Logger) *PDFGeneratedHandler {
	return &PDFGeneratedHandler{logger: logger}
}

// EventType returns the event type this handler processes
func (h *PDFGeneratedHandler) EventType() string {
	return "pdf.generated"
}

// Handle processes the PDF generated event
func (h *PDFGeneratedHandler) Handle(event interface{}) error {
	// Handle PDF generated event logic
	// For example, log metrics, notify users, etc.
	h.logger.Info("PDF generated successfully", "event", event)
	return nil
}

// Priority returns the handler priority
func (h *PDFGeneratedHandler) Priority() int {
	return 100
}

// PDFFailedHandler handles PDF generation failed events
type PDFFailedHandler struct {
	logger core.Logger
}

// NewPDFFailedHandler creates a new PDF failed event handler
func NewPDFFailedHandler(logger core.Logger) *PDFFailedHandler {
	return &PDFFailedHandler{logger: logger}
}

// EventType returns the event type this handler processes
func (h *PDFFailedHandler) EventType() string {
	return "pdf.failed"
}

// Handle processes the PDF failed event
func (h *PDFFailedHandler) Handle(event interface{}) error {
	// Handle PDF failed event logic
	// For example, retry logic, error logging, alerting, etc.
	h.logger.Error("PDF generation failed", "event", event)
	return nil
}

// Priority returns the handler priority
func (h *PDFFailedHandler) Priority() int {
	return 100
}

// PDFGeneratedEvent represents a PDF generated event
type PDFGeneratedEvent struct {
	FileName     string `json:"file_name"`
	TemplateName string `json:"template_name"`
	GeneratedAt  string `json:"generated_at"`
}

// PDFFailedEvent represents a PDF generation failed event
type PDFFailedEvent struct {
	FileName     string `json:"file_name"`
	TemplateName string `json:"template_name"`
	Error        string `json:"error"`
	FailedAt     string `json:"failed_at"`
}
