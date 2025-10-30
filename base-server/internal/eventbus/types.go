package eventbus

import (
	"context"
	"time"
)

// Event represents a generic event in the system
type Event interface {
	GetType() string
	GetID() string
	GetTimestamp() time.Time
	GetPayload() interface{}
	GetMetadata() map[string]interface{}
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	GetEventTypes() []string
	GetName() string
}

// EventBus defines the interface for the event bus
type EventBus interface {
	// Publish an event to all registered handlers
	Publish(ctx context.Context, event Event) error

	// PublishAsync publishes an event asynchronously
	PublishAsync(ctx context.Context, event Event)

	// Subscribe registers an event handler for specific event types
	Subscribe(handler EventHandler) error

	// Unsubscribe removes an event handler
	Unsubscribe(handlerName string) error

	// GetSubscribers returns list of handlers for an event type
	GetSubscribers(eventType string) []EventHandler

	// Shutdown gracefully shuts down the event bus
	Shutdown(ctx context.Context) error
}

// BaseEvent provides a default implementation of the Event interface
type BaseEvent struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   interface{}            `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func (e *BaseEvent) GetType() string {
	return e.Type
}

func (e *BaseEvent) GetID() string {
	return e.ID
}

func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *BaseEvent) GetPayload() interface{} {
	return e.Payload
}

func (e *BaseEvent) GetMetadata() map[string]interface{} {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	return e.Metadata
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType, id string, payload interface{}) *BaseEvent {
	return &BaseEvent{
		Type:      eventType,
		ID:        id,
		Timestamp: time.Now(),
		Payload:   payload,
		Metadata:  make(map[string]interface{}),
	}
}
