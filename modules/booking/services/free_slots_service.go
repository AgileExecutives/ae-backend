package services

import (
	"fmt"
	"time"

	"github.com/unburdy/booking-module/entities"
	"gorm.io/gorm"
)

// CalendarEntry represents a simplified calendar entry for conflict checking
type CalendarEntry struct {
	ID         uint
	CalendarID uint
	TenantID   uint
	StartTime  *time.Time
	EndTime    *time.Time
}

// FreeSlotsService handles free slot calculation
type FreeSlotsService struct {
	db *gorm.DB
}

// NewFreeSlotsService creates a new free slots service
func NewFreeSlotsService(db *gorm.DB) *FreeSlotsService {
	return &FreeSlotsService{db: db}
}

// FreeSlotsRequest contains parameters for free slot calculation
type FreeSlotsRequest struct {
	TemplateID uint
	TenantID   uint
	CalendarID uint
	StartDate  time.Time // Start of range to search
	EndDate    time.Time // End of range to search
	Timezone   string    // Timezone for slot calculation
}

// CalculateFreeSlots generates available time slots based on template configuration
func (s *FreeSlotsService) CalculateFreeSlots(req FreeSlotsRequest, template *entities.BookingTemplate) (*entities.FreeSlotsResponse, error) {
	// Load timezone
	loc, err := time.LoadLocation(req.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Get existing calendar entries for the date range
	var existingEntries []CalendarEntry
	err = s.db.Table("calendar_entries").
		Select("id, calendar_id, tenant_id, start_time, end_time").
		Where("calendar_id = ? AND tenant_id = ? AND start_time >= ? AND start_time <= ? AND deleted_at IS NULL",
			req.CalendarID, req.TenantID, req.StartDate, req.EndDate).
		Find(&existingEntries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch existing entries: %w", err)
	}

	// Generate all possible slots
	allSlots := s.generateAllSlots(req, template, loc)

	// Filter out slots that conflict with existing entries
	availableSlots := s.filterConflictingSlots(allSlots, existingEntries, template.BufferTime)

	// Generate month data
	monthData := s.generateMonthData(availableSlots, req.StartDate, loc)

	// Build configuration response
	config := entities.SlotConfiguration{
		Duration:           template.SlotDuration,
		Interval:           s.determineInterval(template.AllowedIntervals),
		NumberMax:          template.MaxSeriesBookings,
		BufferTime:         template.BufferTime,
		WeeklyAvailability: template.WeeklyAvailability,
	}

	return &entities.FreeSlotsResponse{
		Slots:     availableSlots,
		MonthData: monthData,
		Config:    config,
	}, nil
}

// generateAllSlots creates all possible time slots based on template configuration
func (s *FreeSlotsService) generateAllSlots(req FreeSlotsRequest, template *entities.BookingTemplate, loc *time.Location) []entities.TimeSlot {
	var slots []entities.TimeSlot

	// Iterate through each day in the range
	currentDate := req.StartDate
	for currentDate.Before(req.EndDate) || currentDate.Equal(req.EndDate) {
		// Get the weekday availability
		weekdayAvail := s.getWeekdayAvailability(currentDate, template.WeeklyAvailability)

		// For each availability window in the day
		for _, window := range weekdayAvail {
			// Parse window start and end times
			windowStart, _ := time.Parse("15:04", window.Start)
			windowEnd, _ := time.Parse("15:04", window.End)

			// Create slots within this window
			slotStart := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(),
				windowStart.Hour(), windowStart.Minute(), 0, 0, loc)
			windowEndTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(),
				windowEnd.Hour(), windowEnd.Minute(), 0, 0, loc)

			// Generate slots
			for slotStart.Before(windowEndTime) {
				slotEnd := slotStart.Add(time.Duration(template.SlotDuration) * time.Minute)

				// Only add if slot end is within the window
				if slotEnd.After(windowEndTime) {
					break
				}

				// Check advance booking days and min notice hours
				if !s.isSlotBookable(slotStart, template.AdvanceBookingDays, template.MinNoticeHours) {
					slotStart = slotEnd.Add(time.Duration(template.BufferTime) * time.Minute)
					continue
				}

				// Check if date is blocked
				if s.isDateBlocked(slotStart, template.BlockDates) {
					slotStart = slotEnd.Add(time.Duration(template.BufferTime) * time.Minute)
					continue
				}

				slot := entities.TimeSlot{
					ID:        fmt.Sprintf("slot-%s-%02d-%02d", slotStart.Format("2006-01-02"), slotStart.Hour(), slotStart.Minute()),
					StartTime: slotStart.Format(time.RFC3339),
					EndTime:   slotEnd.Format(time.RFC3339),
					Date:      slotStart.Format("2006-01-02"),
					Time:      slotStart.Format("15:04"),
					Duration:  template.SlotDuration,
					Available: true,
					TimeOfDay: entities.ClassifyTimeOfDay(slotStart.Hour()),
					Timezone:  req.Timezone,
				}

				slots = append(slots, slot)

				// Move to next slot (with buffer time)
				slotStart = slotEnd.Add(time.Duration(template.BufferTime) * time.Minute)
			}
		}

		// Move to next day
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return slots
}

// getWeekdayAvailability returns the availability windows for a given day
func (s *FreeSlotsService) getWeekdayAvailability(date time.Time, weeklyAvail entities.WeeklyAvailability) []entities.TimeRange {
	weekday := date.Weekday().String()
	weekdayLower := map[string][]entities.TimeRange{
		"Monday":    weeklyAvail.Monday,
		"Tuesday":   weeklyAvail.Tuesday,
		"Wednesday": weeklyAvail.Wednesday,
		"Thursday":  weeklyAvail.Thursday,
		"Friday":    weeklyAvail.Friday,
		"Saturday":  weeklyAvail.Saturday,
		"Sunday":    weeklyAvail.Sunday,
	}

	if avail, exists := weekdayLower[weekday]; exists {
		return avail
	}
	return []entities.TimeRange{}
}

// isSlotBookable checks if a slot meets advance booking and min notice requirements
func (s *FreeSlotsService) isSlotBookable(slotStart time.Time, advanceBookingDays, minNoticeHours int) bool {
	now := time.Now()

	// Check minimum notice period
	minNoticeTime := now.Add(time.Duration(minNoticeHours) * time.Hour)
	if slotStart.Before(minNoticeTime) {
		return false
	}

	// Check advance booking limit
	maxAdvanceTime := now.AddDate(0, 0, advanceBookingDays)
	if slotStart.After(maxAdvanceTime) {
		return false
	}

	return true
}

// isDateBlocked checks if a date is in the blocked dates list
func (s *FreeSlotsService) isDateBlocked(slotStart time.Time, blockDates []entities.DateRange) bool {
	for _, blocked := range blockDates {
		blockedStart, _ := time.Parse("2006-01-02", blocked.Start)
		blockedEnd, _ := time.Parse("2006-01-02", blocked.End)

		if slotStart.After(blockedStart) && slotStart.Before(blockedEnd) {
			return true
		}
		if slotStart.Format("2006-01-02") == blocked.Start || slotStart.Format("2006-01-02") == blocked.End {
			return true
		}
	}

	return false
}

// filterConflictingSlots removes slots that conflict with existing entries
func (s *FreeSlotsService) filterConflictingSlots(slots []entities.TimeSlot, existingEntries []CalendarEntry, bufferTime int) []entities.TimeSlot {
	var available []entities.TimeSlot

	for _, slot := range slots {
		slotStart, _ := time.Parse(time.RFC3339, slot.StartTime)
		slotEnd, _ := time.Parse(time.RFC3339, slot.EndTime)

		// Add buffer time to slot for conflict checking
		slotStartWithBuffer := slotStart.Add(-time.Duration(bufferTime) * time.Minute)
		slotEndWithBuffer := slotEnd.Add(time.Duration(bufferTime) * time.Minute)

		isAvailable := true
		for _, entry := range existingEntries {
			if entry.StartTime == nil || entry.EndTime == nil {
				continue
			}

			// Check if there's any overlap
			if slotStartWithBuffer.Before(*entry.EndTime) && slotEndWithBuffer.After(*entry.StartTime) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			available = append(available, slot)
		}
	}

	return available
}

// generateMonthData creates month overview data for calendar display
func (s *FreeSlotsService) generateMonthData(slots []entities.TimeSlot, startDate time.Time, loc *time.Location) entities.MonthData {
	year := startDate.Year()
	month := int(startDate.Month())

	// Group slots by date
	slotsByDate := make(map[string]int)
	totalPossibleByDate := make(map[string]int)

	for _, slot := range slots {
		slotsByDate[slot.Date]++
		totalPossibleByDate[slot.Date]++ // Simplified: count all generated slots as "possible"
	}

	// Generate day data for all days in the month
	var days []entities.DayData
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	lastDay := firstDay.AddDate(0, 1, -1)

	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		availCount := slotsByDate[dateStr]
		totalPossible := totalPossibleByDate[dateStr]
		if totalPossible == 0 {
			totalPossible = 1 // Avoid division by zero
		}

		status := entities.DayStatus(availCount, totalPossible)

		days = append(days, entities.DayData{
			Date:           dateStr,
			AvailableCount: availCount,
			Status:         status,
		})
	}

	return entities.MonthData{
		Year:  year,
		Month: month,
		Days:  days,
	}
}

// determineInterval converts AllowedIntervals to a readable string
func (s *FreeSlotsService) determineInterval(intervals []entities.IntervalType) string {
	if len(intervals) == 0 {
		return "none"
	}

	// Return the first interval as the primary
	switch intervals[0] {
	case entities.IntervalWeekly:
		return "weekly"
	case entities.IntervalMonthlyDate, entities.IntervalMonthlyDay:
		return "monthly"
	case entities.IntervalYearly:
		return "yearly"
	default:
		return "none"
	}
}
