package entities

import (
	"time"

	"gorm.io/gorm"
)

// Session represents a therapy/appointment session linked to a calendar entry
// All timestamps (CreatedAt, UpdatedAt, OriginalDate, OriginalStartTime) are stored in UTC
type Session struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	TenantID          uint           `gorm:"not null;index:idx_session_tenant" json:"tenant_id"`
	ClientID          uint           `gorm:"not null;index:idx_session_client" json:"client_id"`
	CalendarEntryID   *uint          `gorm:"index:idx_session_calendar" json:"calendar_entry_id"` // Nullable - can be NULL if calendar entry is deleted
	OriginalDate      time.Time      `gorm:"not null" json:"original_date"`                       // UTC - Date of original calendar entry
	OriginalStartTime time.Time      `gorm:"not null" json:"original_start_time"`                 // UTC - Start time from original calendar entry
	DurationMin       int            `gorm:"not null" json:"duration_min"`
	Type              string         `gorm:"type:varchar(50);not null" json:"type"`
	NumberUnits       int            `gorm:"not null;default:1" json:"number_units"`
	Status            string         `gorm:"type:varchar(20);not null;default:'scheduled'" json:"status"` // scheduled, canceled, conducted
	Documentation     string         `gorm:"type:text" json:"documentation"`
	InternalNote      string         `gorm:"type:text" json:"internal_note"` // For therapist reference, not shown on invoices
	CreatedAt         time.Time      `json:"created_at"`                     // Always UTC
	UpdatedAt         time.Time      `json:"updated_at"`                     // Always UTC
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Session model
func (Session) TableName() string {
	return "sessions"
}

// CreateSessionRequest represents the request payload for creating a session
type CreateSessionRequest struct {
	ClientID          uint   `json:"client_id" binding:"required" example:"1"`
	CalendarEntryID   uint   `json:"calendar_entry_id" binding:"required" example:"1"`
	OriginalDate      string `json:"original_date" binding:"required" example:"2025-11-26T00:00:00Z"`       // UTC date from calendar entry
	OriginalStartTime string `json:"original_start_time" binding:"required" example:"2025-11-26T10:00:00Z"` // UTC start time from calendar entry
	DurationMin       int    `json:"duration_min" binding:"required,min=1" example:"60"`
	Type              string `json:"type" binding:"required" example:"therapy"`
	NumberUnits       int    `json:"number_units" binding:"required,min=1" example:"1"`
	Status            string `json:"status" binding:"omitempty,oneof=scheduled canceled conducted" example:"scheduled"`
	Documentation     string `json:"documentation,omitempty" example:"Initial session notes"`
}

// UpdateSessionRequest represents the request payload for updating a session
type UpdateSessionRequest struct {
	DurationMin   *int    `json:"duration_min,omitempty" example:"60"`
	Type          *string `json:"type,omitempty" example:"therapy"`
	NumberUnits   *int    `json:"number_units,omitempty" example:"1"`
	Status        *string `json:"status,omitempty" example:"conducted"`
	Documentation *string `json:"documentation,omitempty" example:"Updated session notes"`
	InternalNote  *string `json:"internal_note,omitempty" example:"Client was engaged"`
}

// BookSessionsRequest represents the request to create series of sessions from calendar series
// All time fields (start_time, end_time, last_date) MUST be in UTC format (RFC3339 with Z suffix)
type BookSessionsRequest struct {
	ClientID      uint   `json:"client_id" binding:"required" example:"1"`
	CalendarID    uint   `json:"calendar_id" binding:"required" example:"1"`
	Title         string `json:"title" binding:"required" example:"Therapy Session"`
	Description   string `json:"description,omitempty" example:"Weekly therapy session"`
	IntervalType  string `json:"interval_type,omitempty" binding:"omitempty,oneof=none weekly monthly-date monthly-day yearly" example:"weekly"`
	IntervalValue int    `json:"interval_value,omitempty" binding:"omitempty,min=1" example:"1"`
	LastDate      string `json:"last_date,omitempty" example:"2025-12-31T23:59:59Z"`           // UTC timestamp (RFC3339)
	StartTime     string `json:"start_time" binding:"required" example:"2025-11-26T09:00:00Z"` // UTC timestamp (RFC3339)
	EndTime       string `json:"end_time" binding:"required" example:"2025-11-26T10:00:00Z"`   // UTC timestamp (RFC3339)
	DurationMin   int    `json:"duration_min" binding:"required,min=1" example:"60"`
	Type          string `json:"type" binding:"required" example:"therapy"`
	NumberUnits   int    `json:"number_units" binding:"required,min=1" example:"1"`
	Location      string `json:"location,omitempty" example:"Office 101"`
	Timezone      string `json:"timezone,omitempty" example:"Europe/Berlin"` // For recurring events that follow local time
}

// BookSessionsWithTokenRequest represents the request to book sessions using a token (without client_id/calendar_id)
// All time fields (start_time, end_time, last_date) MUST be in UTC format (RFC3339 with Z suffix)
type BookSessionsWithTokenRequest struct {
	Title         string `json:"title" binding:"required" example:"Therapy Session"`
	Description   string `json:"description,omitempty" example:"Weekly therapy session"`
	IntervalType  string `json:"interval_type,omitempty" binding:"omitempty,oneof=none weekly monthly-date monthly-day yearly" example:"weekly"`
	IntervalValue int    `json:"interval_value,omitempty" binding:"omitempty,min=1" example:"1"`
	LastDate      string `json:"last_date,omitempty" example:"2025-12-31T23:59:59Z"`           // UTC timestamp (RFC3339)
	StartTime     string `json:"start_time" binding:"required" example:"2025-11-26T09:00:00Z"` // UTC timestamp (RFC3339)
	EndTime       string `json:"end_time" binding:"required" example:"2025-11-26T10:00:00Z"`   // UTC timestamp (RFC3339)
	DurationMin   int    `json:"duration_min" binding:"required,min=1" example:"60"`
	Type          string `json:"type" binding:"required" example:"therapy"`
	NumberUnits   int    `json:"number_units" binding:"required,min=1" example:"1"`
	Location      string `json:"location,omitempty" example:"Office 101"`
	Timezone      string `json:"timezone,omitempty" example:"Europe/Berlin"` // For recurring events that follow local time
}

// BookSessionsResponse represents the response for booking multiple sessions
type BookSessionsResponse struct {
	SeriesID *uint             `json:"series_id,omitempty"`
	Sessions []SessionResponse `json:"sessions"`
}

// SessionResponse represents the response format for session data
type SessionResponse struct {
	ID                uint      `json:"id"`
	TenantID          uint      `json:"tenant_id"`
	ClientID          uint      `json:"client_id"`
	CalendarEntryID   *uint     `json:"calendar_entry_id"`   // Nullable - NULL if calendar entry was deleted
	OriginalDate      time.Time `json:"original_date"`       // UTC - Date of original calendar entry
	OriginalStartTime time.Time `json:"original_start_time"` // UTC - Start time from original calendar entry
	DurationMin       int       `json:"duration_min"`
	Type              string    `json:"type"`
	NumberUnits       int       `json:"number_units"`
	Status            string    `json:"status"`
	Documentation     string    `json:"documentation"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ToResponse converts a Session to SessionResponse
// Ensures all timestamps are returned in UTC
func (s *Session) ToResponse() SessionResponse {
	return SessionResponse{
		ID:                s.ID,
		TenantID:          s.TenantID,
		ClientID:          s.ClientID,
		CalendarEntryID:   s.CalendarEntryID,
		OriginalDate:      s.OriginalDate.UTC(),
		OriginalStartTime: s.OriginalStartTime.UTC(),
		DurationMin:       s.DurationMin,
		Type:              s.Type,
		NumberUnits:       s.NumberUnits,
		Status:            s.Status,
		Documentation:     s.Documentation,
		CreatedAt:         s.CreatedAt.UTC(),
		UpdatedAt:         s.UpdatedAt.UTC(),
	}
}

// SessionDetailResponse represents a detailed session response with client and session navigation
type SessionDetailResponse struct {
	SessionResponse
	Client          *ClientResponse  `json:"client,omitempty"`
	PreviousSession *SessionResponse `json:"previous_session,omitempty"`
	NextSession     *SessionResponse `json:"next_session,omitempty"`
}
