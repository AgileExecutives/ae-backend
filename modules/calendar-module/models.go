// Package calendar provides calendar and event management functionality
// This module extends the base SaaS functionality with calendar features
package calendar

import (
	"time"

	"github.com/ae-saas-basic/ae-saas-basic/api"
	"gorm.io/gorm"
)

// Event represents a calendar event
type Event struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	Location    string    `json:"location"`

	// Base SaaS integration
	UserID         uint `json:"user_id" gorm:"index"`
	OrganizationID uint `json:"organization_id" gorm:"index"` // For tenant isolation

	// Event specific fields
	IsAllDay    bool   `json:"is_all_day" gorm:"default:false"`
	RecurringID *uint  `json:"recurring_id,omitempty"`
	EventType   string `json:"event_type" gorm:"default:'appointment'"` // appointment, meeting, reminder, etc.
	Status      string `json:"status" gorm:"default:'scheduled'"`       // scheduled, confirmed, cancelled, completed
	Priority    string `json:"priority" gorm:"default:'medium'"`        // low, medium, high, urgent

	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// EventCreateRequest represents the request to create an event
type EventCreateRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	Location    string    `json:"location"`
	IsAllDay    bool      `json:"is_all_day"`
	EventType   string    `json:"event_type"`
	Priority    string    `json:"priority"`
}

// EventResponse represents the response structure for events
type EventResponse struct {
	Event
	User         *api.UserResponse `json:"user,omitempty"`
	Organization string            `json:"organization,omitempty"`
}

// EventListResponse represents paginated event list
type EventListResponse struct {
	Events     []EventResponse `json:"events"`
	Pagination struct {
		Page       int   `json:"page"`
		Limit      int   `json:"limit"`
		Total      int64 `json:"total"`
		TotalPages int   `json:"total_pages"`
	} `json:"pagination"`
}
