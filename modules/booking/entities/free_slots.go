package entities

// FreeSlotsResponse represents the complete response for free slots endpoint
type FreeSlotsResponse struct {
	Slots     []TimeSlot               `json:"slots"`
	MonthData MonthData                `json:"monthData"`
	Config    SlotConfiguration        `json:"config"`
	Template  *BookingTemplateResponse `json:"template"`
}

// TimeSlot represents an individual available time slot
type TimeSlot struct {
	ID        string `json:"id"`          // Unique identifier (e.g., "slot-2025-11-05-09-00")
	StartTime string `json:"start_time"`  // ISO datetime (e.g., "2025-11-05T09:00:00+01:00")
	EndTime   string `json:"end_time"`    // ISO datetime (e.g., "2025-11-05T09:30:00+01:00")
	Date      string `json:"date"`        // YYYY-MM-DD format
	Time      string `json:"time"`        // HH:mm format
	Duration  int    `json:"duration"`    // Duration in minutes
	Available bool   `json:"available"`   // Is this slot available?
	TimeOfDay string `json:"time_of_day"` // "morning", "afternoon", "evening"
	Timezone  string `json:"timezone"`    // e.g., "Europe/Berlin"
}

// MonthData provides month overview for calendar grid display
type MonthData struct {
	Year  int       `json:"year"`  // e.g., 2025
	Month int       `json:"month"` // 1-12
	Days  []DayData `json:"days"`  // Array of all days in month
}

// DayData represents daily availability summary
type DayData struct {
	Date           string `json:"date"`           // YYYY-MM-DD format
	AvailableCount int    `json:"availableCount"` // Number of free slots
	Status         string `json:"status"`         // "available", "partial", "none"
}

// SlotConfiguration contains the slot configuration rules
type SlotConfiguration struct {
	Duration           int                `json:"duration"`            // Slot duration in minutes
	Interval           string             `json:"interval"`            // "weekly", "biweekly", "none"
	NumberMax          int                `json:"number_max"`          // Max number of series slots
	BufferTime         int                `json:"buffer_time"`         // Buffer between slots in minutes
	WeeklyAvailability WeeklyAvailability `json:"weekly_availability"` // Weekly availability config
}

// TimeOfDay classifies a time into morning/afternoon/evening
func ClassifyTimeOfDay(hour int) string {
	if hour < 12 {
		return "morning"
	} else if hour < 18 {
		return "afternoon"
	}
	return "evening"
}

// DayStatus determines the status based on available count and capacity
func DayStatus(availableCount, totalPossible int) string {
	if availableCount == 0 {
		return "none"
	}

	percentage := float64(availableCount) / float64(totalPossible)
	if percentage > 0.5 {
		return "available"
	}
	return "partial"
}
