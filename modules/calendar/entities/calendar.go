package entities

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Calendar represents a calendar entity
type Calendar struct {
	ID                uint                `gorm:"primarykey" json:"id"`
	TenantID          uint                `gorm:"not null;index" json:"tenant_id"`
	UserID            uint                `gorm:"not null;index" json:"user_id"`
	Title             string              `gorm:"size:255;not null" json:"title" binding:"required" example:"My Calendar"`
	Color             string              `gorm:"size:50" json:"color,omitempty" example:"#FF5733"`
	WeeklyAvailability json.RawMessage    `gorm:"type:json" json:"weekly_availability,omitempty" example:"{}"`
	CalendarUUID      string              `gorm:"size:255;uniqueIndex;not null" json:"calendar_uuid"`
	Timezone          string              `gorm:"size:100" json:"timezone,omitempty" example:"UTC"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	DeletedAt         gorm.DeletedAt      `gorm:"index" json:"-"`

	// Relationships
	CalendarSeries    []CalendarSeries    `gorm:"foreignKey:CalendarID" json:"calendar_series,omitempty"`
	CalendarEntries   []CalendarEntry     `gorm:"foreignKey:CalendarID" json:"calendar_entries,omitempty"`
	ExternalCalendars []ExternalCalendar  `gorm:"foreignKey:CalendarID" json:"external_calendars,omitempty"`
}

// CalendarEntry represents a calendar entry entity
type CalendarEntry struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	CalendarID   uint           `gorm:"not null;index" json:"calendar_id"`
	SeriesID     *uint          `gorm:"index" json:"series_id,omitempty"`
	Title        string         `gorm:"size:255;not null" json:"title" binding:"required" example:"Meeting"`
	IsException  bool           `gorm:"default:false" json:"is_exception" example:"false"`
	Participants json.RawMessage `gorm:"type:json" json:"participants,omitempty" example:"[]"`
	DateFrom     *time.Time     `gorm:"type:date" json:"date_from,omitempty" example:"2025-01-15"`
	DateTo       *time.Time     `gorm:"type:date" json:"date_to,omitempty" example:"2025-01-15"`
	TimeFrom     *time.Time     `gorm:"type:time" json:"time_from,omitempty" example:"09:00"`
	TimeTo       *time.Time     `gorm:"type:time" json:"time_to,omitempty" example:"10:00"`
	Timezone     string         `gorm:"size:100" json:"timezone,omitempty" example:"UTC"`
	Type         string         `gorm:"size:50" json:"type,omitempty" example:"meeting"`
	Description  string         `gorm:"type:text" json:"description,omitempty" example:"Team meeting"`
	Location     string         `gorm:"size:255" json:"location,omitempty" example:"Conference Room A"`
	IsAllDay     bool           `gorm:"default:false" json:"is_all_day" example:"false"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Calendar *Calendar      `gorm:"foreignKey:CalendarID" json:"calendar,omitempty"`
	Series   *CalendarSeries `gorm:"foreignKey:SeriesID" json:"series,omitempty"`
}

// CalendarSeries represents a calendar series entity for recurring events
type CalendarSeries struct {
	ID                   uint                `gorm:"primarykey" json:"id"`
	TenantID             uint                `gorm:"not null;index" json:"tenant_id"`
	UserID               uint                `gorm:"not null;index" json:"user_id"`
	CalendarID           uint                `gorm:"not null;index" json:"calendar_id"`
	Title                string              `gorm:"size:255;not null" json:"title" binding:"required" example:"Weekly Meeting"`
	Participants         json.RawMessage      `gorm:"type:json" json:"participants,omitempty" example:"[]"`
	Weekday              int                 `gorm:"not null" json:"weekday" example:"1"` // 0-6 (Sunday-Saturday)
	Interval             int                 `gorm:"not null;default:1" json:"interval" example:"1"` // Every N weeks
	TimeStart            *time.Time          `gorm:"type:time" json:"time_start,omitempty" example:"09:00"`
	TimeEnd              *time.Time          `gorm:"type:time" json:"time_end,omitempty" example:"10:00"`
	Description          string              `gorm:"type:text" json:"description,omitempty" example:"Weekly team meeting"`
	Location             string              `gorm:"size:255" json:"location,omitempty" example:"Conference Room A"`
	EntryUUID            string              `gorm:"size:255;uniqueIndex;not null" json:"entry_uuid"`
	ExternalUID          string              `gorm:"size:255" json:"external_uid,omitempty" example:"ext-123"`
	Sequence             int                 `gorm:"default:0" json:"sequence" example:"0"`
	ExternalCalendarUUID string              `gorm:"size:255" json:"external_calendar_uuid,omitempty" example:"ext-cal-123"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
	DeletedAt            gorm.DeletedAt      `gorm:"index" json:"-"`

	// Relationships
	Calendar        *Calendar       `gorm:"foreignKey:CalendarID" json:"calendar,omitempty"`
	CalendarEntries []CalendarEntry `gorm:"foreignKey:SeriesID" json:"calendar_entries,omitempty"`
}

// ExternalCalendar represents an external calendar entity
type ExternalCalendar struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	TenantID     uint           `gorm:"not null;index" json:"tenant_id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	CalendarID   uint           `gorm:"not null;index" json:"calendar_id"`
	Title        string         `gorm:"size:255;not null" json:"title" binding:"required" example:"External Calendar"`
	URL          string         `gorm:"size:500" json:"url,omitempty" example:"https://calendar.google.com/ical/..."`
	Settings     json.RawMessage `gorm:"type:json" json:"settings,omitempty" example:"{}"`
	SyncLastRun  *time.Time     `gorm:"type:timestamp" json:"sync_last_run,omitempty"`
	Color        string         `gorm:"size:50" json:"color,omitempty" example:"#33FF57"`
	CalendarUUID string         `gorm:"size:255;uniqueIndex;not null" json:"calendar_uuid"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Calendar *Calendar `gorm:"foreignKey:CalendarID" json:"calendar,omitempty"`
}

// CreateCalendarRequest represents the request payload for creating a calendar
type CreateCalendarRequest struct {
	Title             string         `json:"title" binding:"required" example:"My Calendar"`
	Color             string         `json:"color,omitempty" example:"#FF5733"`
	WeeklyAvailability json.RawMessage `json:"weekly_availability,omitempty" example:"{}"`
	Timezone          string         `json:"timezone,omitempty" example:"UTC"`
}

// UpdateCalendarRequest represents the request payload for updating a calendar
type UpdateCalendarRequest struct {
	Title             *string         `json:"title,omitempty" example:"My Updated Calendar"`
	Color             *string         `json:"color,omitempty" example:"#FF5733"`
	WeeklyAvailability *json.RawMessage `json:"weekly_availability,omitempty" example:"{}"`
	Timezone          *string         `json:"timezone,omitempty" example:"UTC"`
}

// CreateCalendarEntryRequest represents the request payload for creating a calendar entry
type CreateCalendarEntryRequest struct {
	CalendarID   uint           `json:"calendar_id" binding:"required" example:"1"`
	SeriesID     *uint          `json:"series_id,omitempty" example:"1"`
	Title        string         `json:"title" binding:"required" example:"Meeting"`
	IsException  bool           `json:"is_exception,omitempty" example:"false"`
	Participants json.RawMessage `json:"participants,omitempty" example:"[]"`
	DateFrom     *time.Time     `json:"date_from,omitempty" example:"2025-01-15T09:00:00Z"`
	DateTo       *time.Time     `json:"date_to,omitempty" example:"2025-01-15T10:00:00Z"`
	TimeFrom     *time.Time     `json:"time_from,omitempty" example:"09:00:00"`
	TimeTo       *time.Time     `json:"time_to,omitempty" example:"10:00:00"`
	Timezone     string         `json:"timezone,omitempty" example:"UTC"`
	Type         string         `json:"type,omitempty" example:"meeting"`
	Description  string         `json:"description,omitempty" example:"Team meeting"`
	Location     string         `json:"location,omitempty" example:"Conference Room A"`
	IsAllDay     bool           `json:"is_all_day,omitempty" example:"false"`
}

// UpdateCalendarEntryRequest represents the request payload for updating a calendar entry
type UpdateCalendarEntryRequest struct {
	Title        *string         `json:"title,omitempty" example:"Updated Meeting"`
	IsException  *bool           `json:"is_exception,omitempty" example:"false"`
	Participants *json.RawMessage `json:"participants,omitempty" example:"[]"`
	DateFrom     *time.Time      `json:"date_from,omitempty" example:"2025-01-15T09:00:00Z"`
	DateTo       *time.Time      `json:"date_to,omitempty" example:"2025-01-15T10:00:00Z"`
	TimeFrom     *time.Time      `json:"time_from,omitempty" example:"09:00:00"`
	TimeTo       *time.Time      `json:"time_to,omitempty" example:"10:00:00"`
	Timezone     *string         `json:"timezone,omitempty" example:"UTC"`
	Type         *string         `json:"type,omitempty" example:"meeting"`
	Description  *string         `json:"description,omitempty" example:"Updated team meeting"`
	Location     *string         `json:"location,omitempty" example:"Conference Room A"`
	IsAllDay     *bool           `json:"is_all_day,omitempty" example:"false"`
}

// CreateCalendarSeriesRequest represents the request payload for creating a calendar series
type CreateCalendarSeriesRequest struct {
	CalendarID           uint           `json:"calendar_id" binding:"required" example:"1"`
	Title                string         `json:"title" binding:"required" example:"Weekly Meeting"`
	Participants         json.RawMessage `json:"participants,omitempty" example:"[]"`
	Weekday              int            `json:"weekday" binding:"required,min=0,max=6" example:"1"`
	Interval             int            `json:"interval" binding:"required,min=1" example:"1"`
	TimeStart            *time.Time     `json:"time_start,omitempty" example:"09:00:00"`
	TimeEnd              *time.Time     `json:"time_end,omitempty" example:"10:00:00"`
	Description          string         `json:"description,omitempty" example:"Weekly team meeting"`
	Location             string         `json:"location,omitempty" example:"Conference Room A"`
	ExternalUID          string         `json:"external_uid,omitempty" example:"ext-123"`
	ExternalCalendarUUID string         `json:"external_calendar_uuid,omitempty" example:"ext-cal-123"`
}

// UpdateCalendarSeriesRequest represents the request payload for updating a calendar series
type UpdateCalendarSeriesRequest struct {
	Title                *string         `json:"title,omitempty" example:"Updated Weekly Meeting"`
	Participants         *json.RawMessage `json:"participants,omitempty" example:"[]"`
	Weekday              *int            `json:"weekday,omitempty" example:"1"`
	Interval             *int            `json:"interval,omitempty" example:"1"`
	TimeStart            *time.Time      `json:"time_start,omitempty" example:"09:00:00"`
	TimeEnd              *time.Time      `json:"time_end,omitempty" example:"10:00:00"`
	Description          *string         `json:"description,omitempty" example:"Updated weekly team meeting"`
	Location             *string         `json:"location,omitempty" example:"Conference Room A"`
	ExternalUID          *string         `json:"external_uid,omitempty" example:"ext-123"`
	ExternalCalendarUUID *string         `json:"external_calendar_uuid,omitempty" example:"ext-cal-123"`
}

// CreateExternalCalendarRequest represents the request payload for creating an external calendar
type CreateExternalCalendarRequest struct {
	CalendarID uint           `json:"calendar_id" binding:"required" example:"1"`
	Title      string         `json:"title" binding:"required" example:"External Calendar"`
	URL        string         `json:"url,omitempty" example:"https://calendar.google.com/ical/..."`
	Settings   json.RawMessage `json:"settings,omitempty" example:"{}"`
	Color      string         `json:"color,omitempty" example:"#33FF57"`
}

// UpdateExternalCalendarRequest represents the request payload for updating an external calendar
type UpdateExternalCalendarRequest struct {
	Title    *string         `json:"title,omitempty" example:"Updated External Calendar"`
	URL      *string         `json:"url,omitempty" example:"https://calendar.google.com/ical/..."`
	Settings *json.RawMessage `json:"settings,omitempty" example:"{}"`
	Color    *string         `json:"color,omitempty" example:"#33FF57"`
}

// CalendarResponse represents the response format for calendar data
type CalendarResponse struct {
	ID                uint                       `json:"id"`
	TenantID          uint                       `json:"tenant_id"`
	UserID            uint                       `json:"user_id"`
	Title             string                     `json:"title"`
	Color             string                     `json:"color,omitempty"`
	WeeklyAvailability json.RawMessage           `json:"weekly_availability,omitempty"`
	CalendarUUID      string                     `json:"calendar_uuid"`
	Timezone          string                     `json:"timezone,omitempty"`
	CalendarSeries    []CalendarSeriesResponse   `json:"calendar_series,omitempty"`
	CalendarEntries   []CalendarEntryResponse    `json:"calendar_entries,omitempty"`
	ExternalCalendars []ExternalCalendarResponse `json:"external_calendars,omitempty"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

// CalendarEntryResponse represents the response format for calendar entry data
type CalendarEntryResponse struct {
	ID           uint            `json:"id"`
	TenantID     uint            `json:"tenant_id"`
	UserID       uint            `json:"user_id"`
	CalendarID   uint            `json:"calendar_id"`
	SeriesID     *uint           `json:"series_id,omitempty"`
	Title        string          `json:"title"`
	IsException  bool            `json:"is_exception"`
	Participants json.RawMessage `json:"participants,omitempty"`
	DateFrom     *time.Time      `json:"date_from,omitempty"`
	DateTo       *time.Time      `json:"date_to,omitempty"`
	TimeFrom     *time.Time      `json:"time_from,omitempty"`
	TimeTo       *time.Time      `json:"time_to,omitempty"`
	Timezone     string          `json:"timezone,omitempty"`
	Type         string          `json:"type,omitempty"`
	Description  string          `json:"description,omitempty"`
	Location     string          `json:"location,omitempty"`
	IsAllDay     bool            `json:"is_all_day"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CalendarSeriesResponse represents the response format for calendar series data
type CalendarSeriesResponse struct {
	ID                   uint            `json:"id"`
	TenantID             uint            `json:"tenant_id"`
	UserID               uint            `json:"user_id"`
	CalendarID           uint            `json:"calendar_id"`
	Title                string          `json:"title"`
	Participants         json.RawMessage `json:"participants,omitempty"`
	Weekday              int             `json:"weekday"`
	Interval             int             `json:"interval"`
	TimeStart            *time.Time      `json:"time_start,omitempty"`
	TimeEnd              *time.Time      `json:"time_end,omitempty"`
	Description          string          `json:"description,omitempty"`
	Location             string          `json:"location,omitempty"`
	EntryUUID            string          `json:"entry_uuid"`
	ExternalUID          string          `json:"external_uid,omitempty"`
	Sequence             int             `json:"sequence"`
	ExternalCalendarUUID string          `json:"external_calendar_uuid,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// ExternalCalendarResponse represents the response format for external calendar data
type ExternalCalendarResponse struct {
	ID           uint            `json:"id"`
	TenantID     uint            `json:"tenant_id"`
	UserID       uint            `json:"user_id"`
	CalendarID   uint            `json:"calendar_id"`
	Title        string          `json:"title"`
	URL          string          `json:"url,omitempty"`
	Settings     json.RawMessage `json:"settings,omitempty"`
	SyncLastRun  *time.Time      `json:"sync_last_run,omitempty"`
	Color        string          `json:"color,omitempty"`
	CalendarUUID string          `json:"calendar_uuid"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ToResponse converts a Calendar model to CalendarResponse
func (c *Calendar) ToResponse() CalendarResponse {
	var seriesResponses []CalendarSeriesResponse
	for _, series := range c.CalendarSeries {
		seriesResponses = append(seriesResponses, series.ToResponse())
	}

	var entryResponses []CalendarEntryResponse
	for _, entry := range c.CalendarEntries {
		entryResponses = append(entryResponses, entry.ToResponse())
	}

	var externalResponses []ExternalCalendarResponse
	for _, external := range c.ExternalCalendars {
		externalResponses = append(externalResponses, external.ToResponse())
	}

	return CalendarResponse{
		ID:                c.ID,
		TenantID:          c.TenantID,
		UserID:            c.UserID,
		Title:             c.Title,
		Color:             c.Color,
		WeeklyAvailability: c.WeeklyAvailability,
		CalendarUUID:      c.CalendarUUID,
		Timezone:          c.Timezone,
		CalendarSeries:    seriesResponses,
		CalendarEntries:   entryResponses,
		ExternalCalendars: externalResponses,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}

// ToResponse converts a CalendarEntry model to CalendarEntryResponse
func (ce *CalendarEntry) ToResponse() CalendarEntryResponse {
	return CalendarEntryResponse{
		ID:           ce.ID,
		TenantID:     ce.TenantID,
		UserID:       ce.UserID,
		CalendarID:   ce.CalendarID,
		SeriesID:     ce.SeriesID,
		Title:        ce.Title,
		IsException:  ce.IsException,
		Participants: ce.Participants,
		DateFrom:     ce.DateFrom,
		DateTo:       ce.DateTo,
		TimeFrom:     ce.TimeFrom,
		TimeTo:       ce.TimeTo,
		Timezone:     ce.Timezone,
		Type:         ce.Type,
		Description:  ce.Description,
		Location:     ce.Location,
		IsAllDay:     ce.IsAllDay,
		CreatedAt:    ce.CreatedAt,
		UpdatedAt:    ce.UpdatedAt,
	}
}

// ToResponse converts a CalendarSeries model to CalendarSeriesResponse
func (cs *CalendarSeries) ToResponse() CalendarSeriesResponse {
	return CalendarSeriesResponse{
		ID:                   cs.ID,
		TenantID:             cs.TenantID,
		UserID:               cs.UserID,
		CalendarID:           cs.CalendarID,
		Title:                cs.Title,
		Participants:         cs.Participants,
		Weekday:              cs.Weekday,
		Interval:             cs.Interval,
		TimeStart:            cs.TimeStart,
		TimeEnd:              cs.TimeEnd,
		Description:          cs.Description,
		Location:             cs.Location,
		EntryUUID:            cs.EntryUUID,
		ExternalUID:          cs.ExternalUID,
		Sequence:             cs.Sequence,
		ExternalCalendarUUID: cs.ExternalCalendarUUID,
		CreatedAt:            cs.CreatedAt,
		UpdatedAt:            cs.UpdatedAt,
	}
}

// ToResponse converts an ExternalCalendar model to ExternalCalendarResponse
func (ec *ExternalCalendar) ToResponse() ExternalCalendarResponse {
	return ExternalCalendarResponse{
		ID:           ec.ID,
		TenantID:     ec.TenantID,
		UserID:       ec.UserID,
		CalendarID:   ec.CalendarID,
		Title:        ec.Title,
		URL:          ec.URL,
		Settings:     ec.Settings,
		SyncLastRun:  ec.SyncLastRun,
		Color:        ec.Color,
		CalendarUUID: ec.CalendarUUID,
		CreatedAt:    ec.CreatedAt,
		UpdatedAt:    ec.UpdatedAt,
	}
}

// FreeSlotRequest represents the request for finding free time slots
type FreeSlotRequest struct {
	Duration   int `form:"duration" binding:"required,min=1" example:"60"`      // Duration in minutes
	Interval   int `form:"interval" binding:"required,min=1" example:"30"`      // Interval between slots in minutes
	NumberMax  int `form:"number_max" binding:"required,min=1" example:"10"`    // Maximum number of slots to return
}

// FreeSlot represents a free time slot
type FreeSlot struct {
	StartTime time.Time `json:"start_time" example:"2025-01-15T09:00:00Z"`
	EndTime   time.Time `json:"end_time" example:"2025-01-15T10:00:00Z"`
	Duration  int       `json:"duration" example:"60"`
}

// WeekViewRequest represents the request for week view
type WeekViewRequest struct {
	Date string `form:"date" binding:"required" example:"2025-01-15"` // Date in YYYY-MM-DD format
}

// YearViewRequest represents the request for year view
type YearViewRequest struct {
	Year int `form:"year" binding:"required,min=1900,max=2100" example:"2025"`
}

// ImportHolidaysRequest represents the request for importing holidays
type ImportHolidaysRequest struct {
	Country string `json:"country" binding:"required" example:"US"`
	Year    int    `json:"year" binding:"required,min=1900,max=2100" example:"2025"`
}