package events

import (
	"github.com/ae-base-server/pkg/core"
)

// BaseEventHandlers provides event handling for base module
type BaseEventHandlers struct {
	eventBus core.EventBus
	logger   core.Logger
}

// NewBaseEventHandlers creates new base event handlers
func NewBaseEventHandlers(eventBus core.EventBus, logger core.Logger) *BaseEventHandlers {
	return &BaseEventHandlers{
		eventBus: eventBus,
		logger:   logger,
	}
}

// UserCreatedHandler handles user creation events
type UserCreatedHandler struct {
	handlers *BaseEventHandlers
}

// NewUserCreatedHandler creates a new user created event handler
func NewUserCreatedHandler(handlers *BaseEventHandlers) core.EventHandler {
	return &UserCreatedHandler{
		handlers: handlers,
	}
}

func (h *UserCreatedHandler) EventType() string {
	return "user.created"
}

func (h *UserCreatedHandler) Handle(event interface{}) error {
	h.handlers.logger.Info("User created event received", "event", event)
	// Implementation for user creation handling
	return nil
}

func (h *UserCreatedHandler) Priority() int {
	return 100
}

// UserLoginHandler handles user login events
type UserLoginHandler struct {
	handlers *BaseEventHandlers
}

// NewUserLoginHandler creates a new user login event handler
func NewUserLoginHandler(handlers *BaseEventHandlers) core.EventHandler {
	return &UserLoginHandler{
		handlers: handlers,
	}
}

func (h *UserLoginHandler) EventType() string {
	return "user.login"
}

func (h *UserLoginHandler) Handle(event interface{}) error {
	h.handlers.logger.Info("User login event received", "event", event)
	// Implementation for login event handling
	return nil
}

func (h *UserLoginHandler) Priority() int {
	return 100
}

// ContactFormSubmittedHandler handles contact form submission events
type ContactFormSubmittedHandler struct {
	handlers *BaseEventHandlers
}

// NewContactFormSubmittedHandler creates a new contact form submitted event handler
func NewContactFormSubmittedHandler(handlers *BaseEventHandlers) core.EventHandler {
	return &ContactFormSubmittedHandler{
		handlers: handlers,
	}
}

func (h *ContactFormSubmittedHandler) EventType() string {
	return "contact.form.submitted"
}

func (h *ContactFormSubmittedHandler) Handle(event interface{}) error {
	h.handlers.logger.Info("Contact form submitted event received", "event", event)
	// Implementation for contact form submission handling
	return nil
}

func (h *ContactFormSubmittedHandler) Priority() int {
	return 100
}
