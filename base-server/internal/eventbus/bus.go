package eventbus

import (
	"context"
	"log"
	"sync"

	"github.com/ae-base-server/pkg/eventbus"
)

var (
	globalBus eventbus.EventBus
	once      sync.Once
)

// GetBus returns the global event bus instance (singleton)
func GetBus() eventbus.EventBus {
	once.Do(func() {
		globalBus = eventbus.NewMemoryEventBus()
		log.Println("Global event bus initialized")
	})
	return globalBus
}

// PublishUserCreated publishes a user created event
func PublishUserCreated(ctx context.Context, userID, email, tenantID string) error {
	event := eventbus.NewUserCreatedEvent(userID, email, tenantID)
	return GetBus().Publish(ctx, event)
}

// PublishUserCreatedAsync publishes a user created event asynchronously
func PublishUserCreatedAsync(ctx context.Context, userID, email, tenantID string) {
	event := eventbus.NewUserCreatedEvent(userID, email, tenantID)
	GetBus().PublishAsync(ctx, event)
}

// PublishUserUpdated publishes a user updated event
func PublishUserUpdated(ctx context.Context, userID, email, tenantID string, changes map[string]interface{}) error {
	event := eventbus.NewUserUpdatedEvent(userID, email, tenantID, changes)
	return GetBus().Publish(ctx, event)
}

// PublishUserUpdatedAsync publishes a user updated event asynchronously
func PublishUserUpdatedAsync(ctx context.Context, userID, email, tenantID string, changes map[string]interface{}) {
	event := eventbus.NewUserUpdatedEvent(userID, email, tenantID, changes)
	GetBus().PublishAsync(ctx, event)
}

// PublishUserDeleted publishes a user deleted event
func PublishUserDeleted(ctx context.Context, userID, email, tenantID string) error {
	event := eventbus.NewUserDeletedEvent(userID, email, tenantID)
	return GetBus().Publish(ctx, event)
}

// PublishUserDeletedAsync publishes a user deleted event asynchronously
func PublishUserDeletedAsync(ctx context.Context, userID, email, tenantID string) {
	event := eventbus.NewUserDeletedEvent(userID, email, tenantID)
	GetBus().PublishAsync(ctx, event)
}

// PublishUserLoggedIn publishes a user logged in event
func PublishUserLoggedIn(ctx context.Context, userID, email, tenantID, ipAddress, userAgent string) error {
	event := eventbus.NewUserLoggedInEvent(userID, email, tenantID, ipAddress, userAgent)
	return GetBus().Publish(ctx, event)
}

// PublishUserLoggedInAsync publishes a user logged in event asynchronously
func PublishUserLoggedInAsync(ctx context.Context, userID, email, tenantID, ipAddress, userAgent string) {
	event := eventbus.NewUserLoggedInEvent(userID, email, tenantID, ipAddress, userAgent)
	GetBus().PublishAsync(ctx, event)
}

// Subscribe subscribes an event handler to the global bus
func Subscribe(handler eventbus.EventHandler) error {
	return GetBus().Subscribe(handler)
}

// Unsubscribe unsubscribes an event handler from the global bus
func Unsubscribe(handlerName string) error {
	return GetBus().Unsubscribe(handlerName)
}

// Shutdown gracefully shuts down the global event bus
func Shutdown(ctx context.Context) error {
	if globalBus != nil {
		return globalBus.Shutdown(ctx)
	}
	return nil
}
