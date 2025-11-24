package entities

import (
	"time"

	"gorm.io/gorm"
)

// Session represents a therapy/appointment session linked to a calendar entry
type Session struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	TenantID        uint           `gorm:"not null;index:idx_session_tenant" json:"tenant_id"`
	ClientID        uint           `gorm:"not null;index:idx_session_client" json:"client_id"`
	CalendarEntryID uint           `gorm:"not null;index:idx_session_calendar" json:"calendar_entry_id"`
	DurationMin     int            `gorm:"not null" json:"duration_min"`
	Type            string         `gorm:"type:varchar(50);not null" json:"type"`
	NumberUnits     int            `gorm:"not null;default:1" json:"number_units"`
	Status          string         `gorm:"type:varchar(20);not null;default:'scheduled'" json:"status"` // scheduled, canceled, conducted
	Documentation   string         `gorm:"type:text" json:"documentation"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Session model
func (Session) TableName() string {
	return "sessions"
}

// CreateSessionRequest represents the request payload for creating a session
type CreateSessionRequest struct {
	ClientID        uint   `json:"client_id" binding:"required" example:"1"`
	CalendarEntryID uint   `json:"calendar_entry_id" binding:"required" example:"1"`
	DurationMin     int    `json:"duration_min" binding:"required,min=1" example:"60"`
	Type            string `json:"type" binding:"required" example:"therapy"`
	NumberUnits     int    `json:"number_units" binding:"required,min=1" example:"1"`
	Status          string `json:"status" binding:"omitempty,oneof=scheduled canceled conducted" example:"scheduled"`
	Documentation   string `json:"documentation,omitempty" example:"Initial session notes"`
}

// UpdateSessionRequest represents the request payload for updating a session
type UpdateSessionRequest struct {
	DurationMin   *int    `json:"duration_min,omitempty" example:"60"`
	Type          *string `json:"type,omitempty" example:"therapy"`
	NumberUnits   *int    `json:"number_units,omitempty" example:"1"`
	Status        *string `json:"status,omitempty" example:"conducted"`
	Documentation *string `json:"documentation,omitempty" example:"Updated session notes"`
}

// BookSessionsRequest represents the request to create series of sessions from calendar series
type BookSessionsRequest struct {
	ClientID      uint   `json:"client_id" binding:"required" example:"1"`
	CalendarID    uint   `json:"calendar_id" binding:"required" example:"1"`
	Title         string `json:"title" binding:"required" example:"Therapy Session"`
	Description   string `json:"description,omitempty" example:"Weekly therapy session"`
	IntervalType  string `json:"interval_type" binding:"required,oneof=none weekly monthly-date monthly-day yearly" example:"weekly"`
	IntervalValue int    `json:"interval_value" binding:"required,min=1" example:"1"`
	LastDate      string `json:"last_date,omitempty" example:"2025-12-31T23:59:59Z"`
	StartTime     string `json:"start_time" binding:"required" example:"2025-11-25T09:00:00Z"`
	EndTime       string `json:"end_time" binding:"required" example:"2025-11-25T10:00:00Z"`
	DurationMin   int    `json:"duration_min" binding:"required,min=1" example:"60"`
	Type          string `json:"type" binding:"required" example:"therapy"`
	NumberUnits   int    `json:"number_units" binding:"required,min=1" example:"1"`
	Location      string `json:"location,omitempty" example:"Office 101"`
	Timezone      string `json:"timezone,omitempty" example:"Europe/Berlin"`
}

// BookSessionsResponse represents the response for booking multiple sessions
type BookSessionsResponse struct {
	SeriesID uint              `json:"series_id"`
	Sessions []SessionResponse `json:"sessions"`
}

// SessionResponse represents the response format for session data
type SessionResponse struct {
	ID              uint      `json:"id"`
	TenantID        uint      `json:"tenant_id"`
	ClientID        uint      `json:"client_id"`
	CalendarEntryID uint      `json:"calendar_entry_id"`
	DurationMin     int       `json:"duration_min"`
	Type            string    `json:"type"`
	NumberUnits     int       `json:"number_units"`
	Status          string    `json:"status"`
	Documentation   string    `json:"documentation"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ToResponse converts a Session to SessionResponse
func (s *Session) ToResponse() SessionResponse {
	return SessionResponse{
		ID:              s.ID,
		TenantID:        s.TenantID,
		ClientID:        s.ClientID,
		CalendarEntryID: s.CalendarEntryID,
		DurationMin:     s.DurationMin,
		Type:            s.Type,
		NumberUnits:     s.NumberUnits,
		Status:          s.Status,
		Documentation:   s.Documentation,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}
