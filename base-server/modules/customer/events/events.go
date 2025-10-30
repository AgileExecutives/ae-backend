package events

import (
	"github.com/ae-base-server/pkg/core"
)

// CustomerEventHandlers provides event handling for customer module
type CustomerEventHandlers struct {
	eventBus core.EventBus
	logger   core.Logger
}

// NewCustomerEventHandlers creates new customer event handlers
func NewCustomerEventHandlers(eventBus core.EventBus, logger core.Logger) *CustomerEventHandlers {
	return &CustomerEventHandlers{
		eventBus: eventBus,
		logger:   logger,
	}
}

// CustomerCreatedHandler handles customer creation events
type CustomerCreatedHandler struct {
	handlers *CustomerEventHandlers
}

// NewCustomerCreatedHandler creates a new customer created event handler
func NewCustomerCreatedHandler(handlers *CustomerEventHandlers) core.EventHandler {
	return &CustomerCreatedHandler{
		handlers: handlers,
	}
}

func (h *CustomerCreatedHandler) EventType() string {
	return "customer.created"
}

func (h *CustomerCreatedHandler) Handle(event interface{}) error {
	h.handlers.logger.Info("Customer created event received", "event", event)
	// Implementation for customer creation handling
	return nil
}

func (h *CustomerCreatedHandler) Priority() int {
	return 100
}

// PlanSubscribedHandler handles plan subscription events
type PlanSubscribedHandler struct {
	handlers *CustomerEventHandlers
}

// NewPlanSubscribedHandler creates a new plan subscribed event handler
func NewPlanSubscribedHandler(handlers *CustomerEventHandlers) core.EventHandler {
	return &PlanSubscribedHandler{
		handlers: handlers,
	}
}

func (h *PlanSubscribedHandler) EventType() string {
	return "plan.subscribed"
}

func (h *PlanSubscribedHandler) Handle(event interface{}) error {
	h.handlers.logger.Info("Plan subscribed event received", "event", event)
	// Implementation for plan subscription handling
	return nil
}

func (h *PlanSubscribedHandler) Priority() int {
	return 100
}
