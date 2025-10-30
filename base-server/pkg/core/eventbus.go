package core

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// eventHandler holds a handler function with its priority
type eventHandler struct {
	handler  func(interface{}) error
	priority int
}

// simpleEventBus implements EventBus interface
type simpleEventBus struct {
	handlers map[string][]eventHandler
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewEventBus creates a new event bus
func NewEventBus() EventBus {
	return &simpleEventBus{
		handlers: make(map[string][]eventHandler),
	}
}

// Publish publishes an event to all subscribers
func (bus *simpleEventBus) Publish(eventType string, event interface{}) error {
	bus.mutex.RLock()
	handlers := bus.handlers[eventType]
	bus.mutex.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	// Sort handlers by priority (higher priority first)
	sortedHandlers := make([]eventHandler, len(handlers))
	copy(sortedHandlers, handlers)
	sort.Slice(sortedHandlers, func(i, j int) bool {
		return sortedHandlers[i].priority > sortedHandlers[j].priority
	})

	// Execute handlers
	for _, h := range sortedHandlers {
		if err := h.handler(event); err != nil {
			return fmt.Errorf("event handler failed for %s: %w", eventType, err)
		}
	}

	return nil
}

// Subscribe subscribes to an event type
func (bus *simpleEventBus) Subscribe(eventType string, handler func(interface{}) error, priority int) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.handlers[eventType] = append(bus.handlers[eventType], eventHandler{
		handler:  handler,
		priority: priority,
	})

	return nil
}

// Unsubscribe removes a handler (note: this is a simplified implementation)
func (bus *simpleEventBus) Unsubscribe(eventType string, handler func(interface{}) error) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	handlers := bus.handlers[eventType]
	for i, h := range handlers {
		// Note: This comparison might not work as expected due to Go's function comparison limitations
		// In a production system, you'd want to use a different approach (e.g., handler IDs)
		if fmt.Sprintf("%p", h.handler) == fmt.Sprintf("%p", handler) {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return nil
}

// Start starts the event bus
func (bus *simpleEventBus) Start(ctx context.Context) error {
	bus.ctx, bus.cancel = context.WithCancel(ctx)
	return nil
}

// Stop stops the event bus
func (bus *simpleEventBus) Stop(ctx context.Context) error {
	if bus.cancel != nil {
		bus.cancel()
	}
	return nil
}
