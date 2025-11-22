package entities

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// BookingTemplate represents the booking template/configuration for a user's calendar
type BookingTemplate struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// User/Calendar identification
	UserID     uint `gorm:"not null;index" json:"user_id"`
	CalendarID uint `gorm:"not null;index" json:"calendar_id"`
	TenantID   uint `gorm:"not null;index" json:"tenant_id"`

	// Template identification
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`

	// Slot settings
	SlotDuration int `gorm:"not null;default:30" json:"slot_duration"` // Duration in minutes (15, 30, 60)
	BufferTime   int `gorm:"not null;default:0" json:"buffer_time"`    // Buffer between slots in minutes

	// Series/Recurrence settings
	MaxSeriesBookings int           `gorm:"not null;default:1" json:"max_series_bookings"` // Maximum number of series slots per booking
	AllowedIntervals  IntervalArray `gorm:"type:jsonb" json:"allowed_intervals"`           // Which intervals are allowed
	NumberOfIntervals int           `gorm:"not null;default:1" json:"number_of_intervals"` // Multiplier for intervals (e.g., 2 weeks = biweekly)

	// Weekly availability schedule
	WeeklyAvailability WeeklyAvailability `gorm:"type:jsonb;not null" json:"weekly_availability"`

	// Booking window
	AdvanceBookingDays int `gorm:"not null;default:90" json:"advance_booking_days"` // How many days in advance can be booked
	MinNoticeHours     int `gorm:"not null;default:24" json:"min_notice_hours"`     // Minimum notice required for booking

	// Timezone
	Timezone string `gorm:"type:varchar(100);not null;default:'UTC'" json:"timezone"` // e.g., "Europe/Berlin"

	// Optional: Advanced settings
	MaxBookingsPerDay *int           `gorm:"default:null" json:"max_bookings_per_day,omitempty"` // Limit bookings per day
	AllowBackToBack   *bool          `gorm:"default:null" json:"allow_back_to_back,omitempty"`   // Allow bookings without buffer time
	BlockDates        DateRangeArray `gorm:"type:jsonb" json:"block_dates,omitempty"`            // Array of dates to block

	// Allowed start minute marks within an hour (e.g., [0,15,30,45]).
	// Empty means all minute marks are allowed based on slot cadence.
	AllowedStartMinutes MinutesArray `gorm:"type:jsonb" json:"allowed_start_minutes,omitempty"`
}

// IntervalType represents allowed booking intervals
type IntervalType string

const (
	IntervalNone        IntervalType = "none"
	IntervalWeekly      IntervalType = "weekly"
	IntervalMonthlyDate IntervalType = "monthly-date"
	IntervalMonthlyDay  IntervalType = "monthly-day"
	IntervalYearly      IntervalType = "yearly"
)

// IntervalArray is a custom type for storing interval types as JSONB
type IntervalArray []IntervalType

// Value implements the driver.Valuer interface
func (a IntervalArray) Value() (driver.Value, error) {
	if a == nil {
		return json.Marshal([]IntervalType{})
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *IntervalArray) Scan(value interface{}) error {
	if value == nil {
		*a = []IntervalType{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, a)
}

// WeeklyAvailability represents the weekly schedule
type WeeklyAvailability struct {
	Monday    []TimeRange `json:"monday,omitempty"`
	Tuesday   []TimeRange `json:"tuesday,omitempty"`
	Wednesday []TimeRange `json:"wednesday,omitempty"`
	Thursday  []TimeRange `json:"thursday,omitempty"`
	Friday    []TimeRange `json:"friday,omitempty"`
	Saturday  []TimeRange `json:"saturday,omitempty"`
	Sunday    []TimeRange `json:"sunday,omitempty"`
}

// Value implements the driver.Valuer interface
func (w WeeklyAvailability) Value() (driver.Value, error) {
	return json.Marshal(w)
}

// Scan implements the sql.Scanner interface
func (w *WeeklyAvailability) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, w)
}

// TimeRange represents a time range with start and end times
type TimeRange struct {
	Start string `json:"start"` // Time in HH:mm format (e.g., "09:00")
	End   string `json:"end"`   // Time in HH:mm format (e.g., "17:00")
}

// DateRange represents a date range for blocking dates
type DateRange struct {
	Start string `json:"start"` // Date in YYYY-MM-DD format
	End   string `json:"end"`   // Date in YYYY-MM-DD format
}

// DateRangeArray is a custom type for storing date ranges as JSONB
type DateRangeArray []DateRange

// Value implements the driver.Valuer interface
func (d DateRangeArray) Value() (driver.Value, error) {
	if d == nil {
		return json.Marshal([]DateRange{})
	}
	return json.Marshal(d)
}

// Scan implements the sql.Scanner interface
func (d *DateRangeArray) Scan(value interface{}) error {
	if value == nil {
		*d = []DateRange{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, d)
}

// MinutesArray is a custom type for storing minute marks as JSONB
type MinutesArray []int

// Value implements the driver.Valuer interface for MinutesArray
func (m MinutesArray) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal([]int{})
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for MinutesArray
func (m *MinutesArray) Scan(value interface{}) error {
	if value == nil {
		*m = []int{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// TableName specifies the table name for GORM
func (BookingTemplate) TableName() string {
	return "booking_templates"
}
