package eventbus

import (
"fmt"
"github.com/google/uuid"
)

// Event type constants
const (
EventUserCreated  = "user.created"
EventUserUpdated  = "user.updated"
EventUserDeleted  = "user.deleted"
EventUserLoggedIn = "user.logged_in"
)

// UserCreatedPayload contains data for user creation events
type UserCreatedPayload struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
}

// UserUpdatedPayload contains data for user update events
type UserUpdatedPayload struct {
	UserID   string                 `json:"user_id"`
	Email    string                 `json:"email"`
	TenantID string                 `json:"tenant_id"`
	Changes  map[string]interface{} `json:"changes"`
}

// UserDeletedPayload contains data for user deletion events
type UserDeletedPayload struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
}

// UserLoggedInPayload contains data for user login events
type UserLoggedInPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TenantID  string `json:"tenant_id"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(userID, email, tenantID string) Event {
	return NewBaseEvent(
EventUserCreated,
uuid.New().String(),
		&UserCreatedPayload{
			UserID:   userID,
			Email:    email,
			TenantID: tenantID,
		},
	)
}

// NewUserUpdatedEvent creates a new user updated event
func NewUserUpdatedEvent(userID, email, tenantID string, changes map[string]interface{}) Event {
	return NewBaseEvent(
EventUserUpdated,
uuid.New().String(),
		&UserUpdatedPayload{
			UserID:   userID,
			Email:    email,
			TenantID: tenantID,
			Changes:  changes,
		},
	)
}

// NewUserDeletedEvent creates a new user deleted event
func NewUserDeletedEvent(userID, email, tenantID string) Event {
	return NewBaseEvent(
EventUserDeleted,
uuid.New().String(),
		&UserDeletedPayload{
			UserID:   userID,
			Email:    email,
			TenantID: tenantID,
		},
	)
}

// NewUserLoggedInEvent creates a new user logged in event
func NewUserLoggedInEvent(userID, email, tenantID, ipAddress, userAgent string) Event {
	return NewBaseEvent(
EventUserLoggedIn,
uuid.New().String(),
		&UserLoggedInPayload{
			UserID:    userID,
			Email:     email,
			TenantID:  tenantID,
			IPAddress: ipAddress,
			UserAgent: userAgent,
		},
	)
}

// GetUserCreatedPayload safely extracts UserCreatedPayload from an event
func GetUserCreatedPayload(event Event) (*UserCreatedPayload, error) {
	if event.GetType() != EventUserCreated {
		return nil, fmt.Errorf("event type %s is not user.created", event.GetType())
	}
	
	payload, ok := event.GetPayload().(*UserCreatedPayload)
	if !ok {
		return nil, fmt.Errorf("payload is not UserCreatedPayload")
	}
	
	return payload, nil
}

// GetUserUpdatedPayload safely extracts UserUpdatedPayload from an event
func GetUserUpdatedPayload(event Event) (*UserUpdatedPayload, error) {
	if event.GetType() != EventUserUpdated {
		return nil, fmt.Errorf("event type %s is not user.updated", event.GetType())
	}
	
	payload, ok := event.GetPayload().(*UserUpdatedPayload)
	if !ok {
		return nil, fmt.Errorf("payload is not UserUpdatedPayload")
	}
	
	return payload, nil
}

// GetUserDeletedPayload safely extracts UserDeletedPayload from an event
func GetUserDeletedPayload(event Event) (*UserDeletedPayload, error) {
	if event.GetType() != EventUserDeleted {
		return nil, fmt.Errorf("event type %s is not user.deleted", event.GetType())
	}
	
	payload, ok := event.GetPayload().(*UserDeletedPayload)
	if !ok {
		return nil, fmt.Errorf("payload is not UserDeletedPayload")
	}
	
	return payload, nil
}

// GetUserLoggedInPayload safely extracts UserLoggedInPayload from an event
func GetUserLoggedInPayload(event Event) (*UserLoggedInPayload, error) {
	if event.GetType() != EventUserLoggedIn {
		return nil, fmt.Errorf("event type %s is not user.logged_in", event.GetType())
	}
	
	payload, ok := event.GetPayload().(*UserLoggedInPayload)
	if !ok {
		return nil, fmt.Errorf("payload is not UserLoggedInPayload")
	}
	
	return payload, nil
}
