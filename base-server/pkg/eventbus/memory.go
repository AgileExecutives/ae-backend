package eventbus

import (
"context"
"fmt"
"log"
"sync"
)

// MemoryEventBus is an in-memory implementation of the EventBus interface
type MemoryEventBus struct {
	mu        sync.RWMutex
	handlers  map[string][]EventHandler // eventType -> []handlers
	asyncChan chan asyncEvent
	shutdown  chan struct{}
	wg        sync.WaitGroup
	running   bool
}

type asyncEvent struct {
	ctx   context.Context
	event Event
}

// NewMemoryEventBus creates a new in-memory event bus
func NewMemoryEventBus() *MemoryEventBus {
	bus := &MemoryEventBus{
		handlers:  make(map[string][]EventHandler),
		asyncChan: make(chan asyncEvent, 1000), // Buffer for async events
		shutdown:  make(chan struct{}),
	}
	
	// Start async event processing goroutine
	bus.start()
	return bus
}

// start begins the async event processing goroutine
func (b *MemoryEventBus) start() {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return
	}
	b.running = true
	b.mu.Unlock()
	
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case asyncEvent := <-b.asyncChan:
				if err := b.Publish(asyncEvent.ctx, asyncEvent.event); err != nil {
					log.Printf("Error processing async event %s: %v", asyncEvent.event.GetType(), err)
				}
			case <-b.shutdown:
				return
			}
		}
	}()
}

// Publish publishes an event to all registered handlers synchronously
func (b *MemoryEventBus) Publish(ctx context.Context, event Event) error {
	b.mu.RLock()
	handlers, exists := b.handlers[event.GetType()]
	b.mu.RUnlock()
	
	if !exists || len(handlers) == 0 {
		// No handlers registered for this event type
		return nil
	}
	
	var errs []error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			errs = append(errs, fmt.Errorf("handler %s failed: %w", handler.GetName(), err))
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("event handling errors: %v", errs)
	}
	
	return nil
}

// PublishAsync publishes an event asynchronously
func (b *MemoryEventBus) PublishAsync(ctx context.Context, event Event) {
	select {
	case b.asyncChan <- asyncEvent{ctx: ctx, event: event}:
		// Event queued successfully
	default:
		// Channel is full, log error
		log.Printf("Warning: async event channel full, dropping event %s", event.GetType())
	}
}

// Subscribe registers an event handler for specific event types
func (b *MemoryEventBus) Subscribe(handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}
	
	eventTypes := handler.GetEventTypes()
	if len(eventTypes) == 0 {
		return fmt.Errorf("handler must specify at least one event type")
	}
	
	b.mu.Lock()
	defer b.mu.Unlock()
	
	for _, eventType := range eventTypes {
		// Check if handler is already subscribed to this event type
		for _, existingHandler := range b.handlers[eventType] {
			if existingHandler.GetName() == handler.GetName() {
				return fmt.Errorf("handler %s already subscribed to event type %s", 
handler.GetName(), eventType)
			}
		}
		
		b.handlers[eventType] = append(b.handlers[eventType], handler)
	}
	
	return nil
}

// Unsubscribe removes an event handler by name
func (b *MemoryEventBus) Unsubscribe(handlerName string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	removed := false
	for eventType, handlers := range b.handlers {
		newHandlers := make([]EventHandler, 0, len(handlers))
		for _, handler := range handlers {
			if handler.GetName() != handlerName {
				newHandlers = append(newHandlers, handler)
			} else {
				removed = true
			}
		}
		b.handlers[eventType] = newHandlers
	}
	
	if !removed {
		return fmt.Errorf("handler %s not found", handlerName)
	}
	
	return nil
}

// GetSubscribers returns list of handlers for an event type
func (b *MemoryEventBus) GetSubscribers(eventType string) []EventHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	handlers, exists := b.handlers[eventType]
	if !exists {
		return nil
	}
	
	// Return a copy to avoid race conditions
	result := make([]EventHandler, len(handlers))
	copy(result, handlers)
	return result
}

// Shutdown gracefully shuts down the event bus
func (b *MemoryEventBus) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	if !b.running {
		b.mu.Unlock()
		return nil
	}
	b.running = false
	b.mu.Unlock()
	
	close(b.shutdown)
	
	// Wait for async processor to finish with timeout
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
