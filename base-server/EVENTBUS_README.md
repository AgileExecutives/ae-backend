# Event Bus System

This document describes the event-driven architecture implemented in the base-server for modular plugin support.

## Overview

The event bus system enables loose coupling between different parts of the application through an event-driven architecture. When important actions happen (like user creation), events are published that other modules can subscribe to and respond accordingly.

## Architecture

### Core Components

1. **Event Interface** (`pkg/eventbus/types.go`)
   - Defines the contract for all events
   - Contains event metadata (ID, timestamp, type, payload)

2. **EventHandler Interface** (`pkg/eventbus/types.go`)
   - Defines how modules can listen to and handle events
   - Specifies which event types a handler is interested in

3. **EventBus Interface** (`pkg/eventbus/types.go`)
   - Defines the core event bus functionality
   - Supports both synchronous and asynchronous event publishing

4. **In-Memory Implementation** (`pkg/eventbus/memory.go`)
   - Thread-safe in-memory event bus
   - Supports async event processing with buffered channels
   - Graceful shutdown with context cancellation

### Event Flow

```
User Registration → Auth Handler → PublishUserCreated → Calendar Handler → Create Calendar
```

## Usage Examples

### Publishing Events

```go
import "github.com/ae-saas-basic/ae-saas-basic/internal/eventbus"

// Synchronous event publishing
err := eventbus.PublishUserCreated(ctx, userID, email, tenantID)

// Asynchronous event publishing (recommended for non-critical events)
eventbus.PublishUserCreatedAsync(ctx, userID, email, tenantID)
```

### Creating Event Handlers

```go
type MyHandler struct {
    name string
}

func (h *MyHandler) Handle(ctx context.Context, event eventbus.Event) error {
    switch event.GetType() {
    case eventbus.EventUserCreated:
        payload, err := eventbus.GetUserCreatedPayload(event)
        if err != nil {
            return err
        }
        // Handle user created event
        log.Printf("User created: %s", payload.Email)
    }
    return nil
}

func (h *MyHandler) GetEventTypes() []string {
    return []string{eventbus.EventUserCreated}
}

func (h *MyHandler) GetName() string {
    return h.name
}
```

### Registering Handlers

Add your handler to `internal/eventbus/init.go`:

```go
func InitializeEventHandlers() {
    // Your custom handler
    myHandler := handlers.NewMyHandler()
    if err := Subscribe(myHandler); err != nil {
        log.Printf("Failed to subscribe my handler: %v", err)
    }
}
```

## Available Events

### User Events

- **user.created**: Fired when a new user is registered
- **user.updated**: Fired when user data is modified  
- **user.deleted**: Fired when a user is deleted
- **user.logged_in**: Fired when a user logs in

### Event Payloads

Each event type has a specific payload structure defined in `pkg/eventbus/events.go`:

```go
type UserCreatedPayload struct {
    UserID   string `json:"user_id"`
    Email    string `json:"email"`
    TenantID string `json:"tenant_id"`
}
```

## Example Plugins

### Calendar Handler (`internal/eventbus/handlers/calendar.go`)

Demonstrates how to create a plugin that:
- Listens for `user.created` events
- Automatically creates calendars for new users
- Handles cleanup on `user.deleted` events

### Integration Points

The event bus is integrated at these key points:

1. **Server Startup** (`main.go`): Initializes event handlers
2. **User Registration** (`internal/handlers/auth.go`): Publishes `user.created` events
3. **Handler Registration** (`internal/eventbus/init.go`): Registers all event handlers

## Testing

Run the test to see the event bus in action:

```bash
cd test && go run test_eventbus.go
```

This will demonstrate:
- Event handler registration
- Synchronous and asynchronous event publishing  
- Event processing and handler execution
- Graceful shutdown

## Benefits

1. **Modularity**: Plugins can be added/removed without affecting core functionality
2. **Loose Coupling**: Core business logic doesn't need to know about plugins
3. **Scalability**: Async processing prevents plugins from blocking critical paths
4. **Testability**: Event handlers can be tested in isolation
5. **Extensibility**: New event types and handlers can be added easily

## Adding New Plugins

To add a new plugin (e.g., email notifications):

1. Create handler in `internal/eventbus/handlers/email.go`
2. Implement the `EventHandler` interface
3. Register in `internal/eventbus/init.go`
4. Define any new event types in `pkg/eventbus/events.go`

The system will automatically route relevant events to your handler!
