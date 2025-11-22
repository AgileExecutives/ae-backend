package entities

// CreateBookingTemplateRequest represents the request to create a booking template
type CreateBookingTemplateRequest struct {
	UserID             uint               `json:"user_id" binding:"required"`
	CalendarID         uint               `json:"calendar_id" binding:"required"`
	Name               string             `json:"name" binding:"required"`
	Description        string             `json:"description"`
	SlotDuration       int                `json:"slot_duration" binding:"required,min=5"`
	BufferTime         int                `json:"buffer_time" binding:"min=0"`
	MaxSeriesBookings  int                `json:"max_series_bookings" binding:"required,min=1"`
	AllowedIntervals   []IntervalType     `json:"allowed_intervals" binding:"required"`
	NumberOfIntervals  int                `json:"number_of_intervals" binding:"required,min=1"`
	WeeklyAvailability WeeklyAvailability `json:"weekly_availability" binding:"required"`
	AdvanceBookingDays int                `json:"advance_booking_days" binding:"required,min=1"`
	MinNoticeHours     int                `json:"min_notice_hours" binding:"required,min=0"`
	Timezone           string             `json:"timezone" binding:"required"`
	MaxBookingsPerDay  *int               `json:"max_bookings_per_day,omitempty"`
	AllowBackToBack    *bool              `json:"allow_back_to_back,omitempty"`
	BlockDates         []DateRange        `json:"block_dates,omitempty"`
	// Allowed start minute marks within an hour (e.g., [0,15,30,45]). Optional.
	// Allowed start minute marks within an hour (e.g., [0,15,30,45]). Optional. Empty means all minute marks permitted.
	// swagger:example [0,15,30,45]
	AllowedStartMinutes []int `json:"allowed_start_minutes,omitempty"`
}

// UpdateBookingTemplateRequest represents the request to update a booking template
type UpdateBookingTemplateRequest struct {
	Name               *string             `json:"name,omitempty"`
	Description        *string             `json:"description,omitempty"`
	SlotDuration       *int                `json:"slot_duration,omitempty" binding:"omitempty,min=5"`
	BufferTime         *int                `json:"buffer_time,omitempty" binding:"omitempty,min=0"`
	MaxSeriesBookings  *int                `json:"max_series_bookings,omitempty" binding:"omitempty,min=1"`
	AllowedIntervals   []IntervalType      `json:"allowed_intervals,omitempty"`
	NumberOfIntervals  *int                `json:"number_of_intervals,omitempty" binding:"omitempty,min=1"`
	WeeklyAvailability *WeeklyAvailability `json:"weekly_availability,omitempty"`
	AdvanceBookingDays *int                `json:"advance_booking_days,omitempty" binding:"omitempty,min=1"`
	MinNoticeHours     *int                `json:"min_notice_hours,omitempty" binding:"omitempty,min=0"`
	Timezone           *string             `json:"timezone,omitempty"`
	MaxBookingsPerDay  *int                `json:"max_bookings_per_day,omitempty"`
	AllowBackToBack    *bool               `json:"allow_back_to_back,omitempty"`
	BlockDates         []DateRange         `json:"block_dates,omitempty"`
	// Allowed start minute marks within an hour (e.g., [0,15,30,45]). Optional.
	// Allowed start minute marks within an hour (e.g., [0,15,30,45]). Optional. Empty means all minute marks permitted.
	// swagger:example [0,15,30,45]
	AllowedStartMinutes []int `json:"allowed_start_minutes,omitempty"`
}

// BookingTemplateResponse represents the response for a booking template
type BookingTemplateResponse struct {
	ID                 uint               `json:"id"`
	UserID             uint               `json:"user_id"`
	CalendarID         uint               `json:"calendar_id"`
	TenantID           uint               `json:"tenant_id"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	SlotDuration       int                `json:"slot_duration"`
	BufferTime         int                `json:"buffer_time"`
	MaxSeriesBookings  int                `json:"max_series_bookings"`
	AllowedIntervals   []IntervalType     `json:"allowed_intervals"`
	NumberOfIntervals  int                `json:"number_of_intervals"`
	WeeklyAvailability WeeklyAvailability `json:"weekly_availability"`
	AdvanceBookingDays int                `json:"advance_booking_days"`
	MinNoticeHours     int                `json:"min_notice_hours"`
	Timezone           string             `json:"timezone"`
	MaxBookingsPerDay  *int               `json:"max_bookings_per_day,omitempty"`
	AllowBackToBack    *bool              `json:"allow_back_to_back,omitempty"`
	BlockDates         []DateRange        `json:"block_dates,omitempty"`
	// Allowed start minute marks within an hour (e.g., [0,15,30,45]). Empty when all allowed.
	// swagger:example [0,15,30,45]
	AllowedStartMinutes []int  `json:"allowed_start_minutes,omitempty"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

// ToResponse converts BookingTemplate to BookingTemplateResponse
func (bc *BookingTemplate) ToResponse() BookingTemplateResponse {
	return BookingTemplateResponse{
		ID:                  bc.ID,
		UserID:              bc.UserID,
		CalendarID:          bc.CalendarID,
		TenantID:            bc.TenantID,
		Name:                bc.Name,
		Description:         bc.Description,
		SlotDuration:        bc.SlotDuration,
		BufferTime:          bc.BufferTime,
		MaxSeriesBookings:   bc.MaxSeriesBookings,
		AllowedIntervals:    bc.AllowedIntervals,
		NumberOfIntervals:   bc.NumberOfIntervals,
		WeeklyAvailability:  bc.WeeklyAvailability,
		AdvanceBookingDays:  bc.AdvanceBookingDays,
		MinNoticeHours:      bc.MinNoticeHours,
		Timezone:            bc.Timezone,
		MaxBookingsPerDay:   bc.MaxBookingsPerDay,
		AllowBackToBack:     bc.AllowBackToBack,
		BlockDates:          bc.BlockDates,
		AllowedStartMinutes: []int(bc.AllowedStartMinutes),
		CreatedAt:           bc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           bc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
