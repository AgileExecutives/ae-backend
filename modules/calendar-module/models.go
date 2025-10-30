// Package calendar provides calendar and event management functionality
// This module extends the base SaaS functionality with calendar features
package calendar

import (
	"time"

	"github.com/ae-saas-basic/ae-saas-basic/api"
	"gorm.io/gorm"
)

// Calendar represents a user's calendar
type Calendar struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	Name               string `json:"name" binding:"required"`
	ScheduleDaysAhead  int    `json:"schedule_days_ahead" gorm:"default:30"` // How many days ahead to schedule
	ScheduleDaysNotice int    `json:"schedule_days_notice" gorm:"default:7"` // Days notice for scheduling
	ScheduleUnit       string `json:"schedule_unit" gorm:"default:'days'"`   // days, weeks, months

	// Base SaaS integration
	UserID   uint `json:"user_id" gorm:"index"`
	TenantID uint `json:"tenant_id" gorm:"index"` // For tenant isolation

	// Relationships
	Events          []Event          `json:"events,omitempty" gorm:"foreignKey:CalendarID"`
	RecurringEvents []RecurringEvent `json:"recurring_events,omitempty" gorm:"foreignKey:CalendarID"`

	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// RecurringEvent represents a recurring event pattern
type RecurringEvent struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	Participants string `json:"participants"` // Comma-separated user IDs or emails

	// Recurrence pattern
	Weekday   int    `json:"weekday" gorm:"comment:'0=Sunday, 1=Monday, ..., 6=Saturday'"`
	Interval  int    `json:"interval" gorm:"default:1;comment:'Repeat every X weeks'"`
	StartTime string `json:"start_time" gorm:"comment:'Time format: HH:MM'"` // e.g., "09:00"
	EndTime   string `json:"end_time" gorm:"comment:'Time format: HH:MM'"`   // e.g., "10:00"

	// Calendar integration
	CalendarID uint      `json:"calendar_id" gorm:"index"`
	Calendar   *Calendar `json:"calendar,omitempty" gorm:"foreignKey:CalendarID"`

	// Base SaaS integration
	UserID   uint `json:"user_id" gorm:"index"`
	TenantID uint `json:"tenant_id" gorm:"index"` // For tenant isolation

	// Relationships
	Events []Event `json:"events,omitempty" gorm:"foreignKey:RecurringEventID"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Event represents a calendar event
type Event struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Title        string    `json:"title" binding:"required"`
	Description  string    `json:"description"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	EndTime      time.Time `json:"end_time" binding:"required"`
	Location     string    `json:"location"`
	Participants string    `json:"participants"` // Comma-separated user IDs or emails

	// Calendar integration
	CalendarID uint      `json:"calendar_id" gorm:"index"`
	Calendar   *Calendar `json:"calendar,omitempty" gorm:"foreignKey:CalendarID"`

	// Base SaaS integration
	UserID   uint `json:"user_id" gorm:"index"`
	TenantID uint `json:"tenant_id" gorm:"index"` // For tenant isolation

	// Event specific fields
	IsException      bool            `json:"is_exception" gorm:"default:false"`
	IsAllDay         bool            `json:"is_all_day" gorm:"default:false"`
	RecurringEventID *uint           `json:"recurring_event_id,omitempty"`
	RecurringEvent   *RecurringEvent `json:"recurring_event,omitempty" gorm:"foreignKey:RecurringEventID"`
	EventType        string          `json:"event_type" gorm:"default:'appointment'"` // appointment, meeting, reminder, etc.
	Status           string          `json:"status" gorm:"default:'scheduled'"`       // scheduled, confirmed, cancelled, completed
	Priority         string          `json:"priority" gorm:"default:'medium'"`        // low, medium, high, urgent

	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// CalendarCreateRequest represents the request to create a calendar
type CalendarCreateRequest struct {
	Name               string `json:"name" binding:"required"`
	ScheduleDaysAhead  int    `json:"schedule_days_ahead"`
	ScheduleDaysNotice int    `json:"schedule_days_notice"`
	ScheduleUnit       string `json:"schedule_unit"`
}

// CalendarUpdateRequest represents the request to update a calendar
type CalendarUpdateRequest struct {
	Name               string `json:"name"`
	ScheduleDaysAhead  int    `json:"schedule_days_ahead"`
	ScheduleDaysNotice int    `json:"schedule_days_notice"`
	ScheduleUnit       string `json:"schedule_unit"`
}

// CalendarResponse represents the response structure for calendars
type CalendarResponse struct {
	Calendar
	User           *api.UserResponse `json:"user,omitempty"`
	EventCount     int64             `json:"event_count"`
	RecurringCount int64             `json:"recurring_count"`
}

// RecurringEventCreateRequest represents the request to create a recurring event
type RecurringEventCreateRequest struct {
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	Participants string `json:"participants"`
	Weekday      int    `json:"weekday" binding:"required,min=0,max=6"`
	Interval     int    `json:"interval" binding:"required,min=1"`
	StartTime    string `json:"start_time" binding:"required"`
	EndTime      string `json:"end_time" binding:"required"`
	CalendarID   uint   `json:"calendar_id" binding:"required"`
}

// RecurringEventUpdateRequest represents the request to update a recurring event
type RecurringEventUpdateRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Location     string `json:"location"`
	Participants string `json:"participants"`
	Weekday      int    `json:"weekday" binding:"min=0,max=6"`
	Interval     int    `json:"interval" binding:"min=1"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	IsActive     *bool  `json:"is_active"`
}

// RecurringEventResponse represents the response structure for recurring events
type RecurringEventResponse struct {
	RecurringEvent
	Calendar   *CalendarResponse `json:"calendar,omitempty"`
	User       *api.UserResponse `json:"user,omitempty"`
	EventCount int64             `json:"generated_events_count"`
	NextEvent  *time.Time        `json:"next_event_date,omitempty"`
}

// EventCreateRequest represents the request to create an event
type EventCreateRequest struct {
	Title            string    `json:"title" binding:"required"`
	Description      string    `json:"description"`
	StartTime        time.Time `json:"start_time" binding:"required"`
	EndTime          time.Time `json:"end_time" binding:"required"`
	Location         string    `json:"location"`
	Participants     string    `json:"participants"`
	IsAllDay         bool      `json:"is_all_day"`
	EventType        string    `json:"event_type"`
	Priority         string    `json:"priority"`
	CalendarID       uint      `json:"calendar_id" binding:"required"`
	RecurringEventID *uint     `json:"recurring_event_id,omitempty"`
}

// EventResponse represents the response structure for events
type EventResponse struct {
	Event
	User         *api.UserResponse `json:"user,omitempty"`
	Organization string            `json:"organization,omitempty"`
}

// CalendarListResponse represents paginated calendar list
type CalendarListResponse struct {
	Calendars  []CalendarResponse `json:"calendars"`
	Pagination struct {
		Page       int   `json:"page"`
		Limit      int   `json:"limit"`
		Total      int64 `json:"total"`
		TotalPages int   `json:"total_pages"`
	} `json:"pagination"`
}

// RecurringEventListResponse represents paginated recurring event list
type RecurringEventListResponse struct {
	RecurringEvents []RecurringEventResponse `json:"recurring_events"`
	Pagination      struct {
		Page       int   `json:"page"`
		Limit      int   `json:"limit"`
		Total      int64 `json:"total"`
		TotalPages int   `json:"total_pages"`
	} `json:"pagination"`
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
