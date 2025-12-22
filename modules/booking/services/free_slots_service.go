package services

import (
	"fmt"
	"sort"
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

	// Get weekly availability - prioritize template, fallback to calendar, then default to all-day
	weeklyAvailability := s.getWeeklyAvailabilityWithFallback(req.CalendarID, req.TenantID, template.WeeklyAvailability)

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
	allSlots := s.generateAllSlots(req, template, loc, weeklyAvailability)

	// Filter out slots that conflict with existing entries
	availableSlots := s.filterConflictingSlots(allSlots, existingEntries, template.BufferTime)

	// Calculate available recurrences for series bookings
	if len(template.AllowedIntervals) > 0 && template.MaxSeriesBookings > 0 {
		availableSlots = s.calculateAvailableRecurrences(availableSlots, existingEntries, template, req.EndDate, loc)
	}

	// Generate month data
	monthData := s.generateMonthData(availableSlots, req.StartDate, loc)

	// Build configuration response
	config := entities.SlotConfiguration{
		Duration:           template.SlotDuration,
		Interval:           s.determineInterval(template.AllowedIntervals),
		NumberMax:          template.MaxSeriesBookings,
		BufferTime:         template.BufferTime,
		WeeklyAvailability: weeklyAvailability,
	}

	// Convert template to response
	templateResponse := template.ToResponse()

	return &entities.FreeSlotsResponse{
		Slots:     availableSlots,
		MonthData: monthData,
		Config:    config,
		Template:  &templateResponse,
	}, nil
}

// getWeeklyAvailabilityWithFallback returns weekly availability from template, calendar, or default all-day
func (s *FreeSlotsService) getWeeklyAvailabilityWithFallback(calendarID, tenantID uint, templateAvailability entities.WeeklyAvailability) entities.WeeklyAvailability {
	// Check if template has any availability defined
	if s.hasAvailability(templateAvailability) {
		return templateAvailability
	}

	// Fallback to calendar availability
	var calendar struct {
		WeeklyAvailability []byte `gorm:"column:weekly_availability"`
	}

	err := s.db.Table("calendars").
		Select("weekly_availability").
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", calendarID, tenantID).
		Take(&calendar).Error

	if err == nil && len(calendar.WeeklyAvailability) > 0 {
		var calendarAvailability entities.WeeklyAvailability
		if err := calendarAvailability.Scan(calendar.WeeklyAvailability); err == nil {
			if s.hasAvailability(calendarAvailability) {
				return calendarAvailability
			}
		}
	}

	// Default: all days 00:00-23:59 (full availability)
	return s.getDefaultAllDayAvailability()
}

// hasAvailability checks if WeeklyAvailability has any time ranges defined
func (s *FreeSlotsService) hasAvailability(wa entities.WeeklyAvailability) bool {
	return len(wa.Monday) > 0 || len(wa.Tuesday) > 0 || len(wa.Wednesday) > 0 ||
		len(wa.Thursday) > 0 || len(wa.Friday) > 0 || len(wa.Saturday) > 0 || len(wa.Sunday) > 0
}

// getDefaultAllDayAvailability returns availability for all days, all hours
func (s *FreeSlotsService) getDefaultAllDayAvailability() entities.WeeklyAvailability {
	allDay := []entities.TimeRange{{Start: "00:00", End: "23:59"}}
	return entities.WeeklyAvailability{
		Monday:    allDay,
		Tuesday:   allDay,
		Wednesday: allDay,
		Thursday:  allDay,
		Friday:    allDay,
		Saturday:  allDay,
		Sunday:    allDay,
	}
}

// generateAllSlots creates all possible time slots based on template configuration
func (s *FreeSlotsService) generateAllSlots(req FreeSlotsRequest, template *entities.BookingTemplate, loc *time.Location, weeklyAvailability entities.WeeklyAvailability) []entities.TimeSlot {
	var slots []entities.TimeSlot

	// Prepare allowed start minutes helpers (if any)
	allowedMinutes := []int(template.AllowedStartMinutes)
	sort.Ints(allowedMinutes)
	allowedSet := map[int]bool{}
	for _, m := range allowedMinutes {
		if m >= 0 && m < 60 {
			allowedSet[m] = true
		}
	}

	// Helper to align a time to the next allowed minute mark
	alignToNextAllowed := func(t time.Time) time.Time {
		if len(allowedMinutes) == 0 {
			return t
		}
		m := t.Minute()
		if allowedSet[m] {
			return t
		}
		// Find next allowed minute in the current hour
		for _, am := range allowedMinutes {
			if am >= m {
				return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), am, 0, 0, t.Location())
			}
		}
		// No allowed minute left in this hour, roll to next hour at first allowed minute
		nextHour := t.Add(time.Hour)
		return time.Date(nextHour.Year(), nextHour.Month(), nextHour.Day(), nextHour.Hour(), allowedMinutes[0], 0, 0, t.Location())
	}

	// Iterate through each day in the range
	currentDate := req.StartDate
	for currentDate.Before(req.EndDate) || currentDate.Equal(req.EndDate) {
		// Get the weekday availability
		weekdayAvail := s.getWeekdayAvailability(currentDate, weeklyAvailability)

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

			// Align initial slot start to allowed minutes if configured
			if len(allowedMinutes) > 0 {
				slotStart = alignToNextAllowed(slotStart)
			}

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

				// Move to next slot start time
				if len(allowedMinutes) > 0 {
					// When allowed minutes are configured, find the next allowed minute mark
					// Buffer time determines minimum gap: if 0, slots can overlap; if >0, enforce gap
					currentHour := slotStart.Hour()
					found := false

					// If buffer time is 0, just find the next allowed minute after current start
					// If buffer time > 0, find next allowed minute after (slotEnd + buffer)
					var minNextStart time.Time
					if template.BufferTime == 0 {
						// Allow overlapping slots - just find next allowed minute mark
						minNextStart = slotStart.Add(1 * time.Minute) // Just after current start
					} else {
						// Respect buffer time - next slot must start after current slot ends + buffer
						minNextStart = slotEnd.Add(time.Duration(template.BufferTime) * time.Minute)
					}

					// Try remaining allowed minutes in current hour
					for _, am := range allowedMinutes {
						nextTime := time.Date(slotStart.Year(), slotStart.Month(), slotStart.Day(),
							currentHour, am, 0, 0, slotStart.Location())
						if (nextTime.After(minNextStart) || nextTime.Equal(minNextStart)) && nextTime.After(slotStart) {
							slotStart = nextTime
							found = true
							break
						}
					}

					if !found {
						// No allowed minute in current hour works, try next hours
						for hour := currentHour + 1; hour < 24 && !found; hour++ {
							for _, am := range allowedMinutes {
								nextTime := time.Date(slotStart.Year(), slotStart.Month(), slotStart.Day(),
									hour, am, 0, 0, slotStart.Location())
								if nextTime.After(minNextStart) || nextTime.Equal(minNextStart) {
									slotStart = nextTime
									found = true
									break
								}
							}
						}
					}

					if !found {
						// Couldn't find valid slot in this day, break the loop
						break
					}
				} else {
					// No allowed minutes restriction, advance by slot duration + buffer
					slotStart = slotEnd.Add(time.Duration(template.BufferTime) * time.Minute)
				}
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

// calculateAvailableRecurrences calculates how many future recurrences are available for each slot
func (s *FreeSlotsService) calculateAvailableRecurrences(slots []entities.TimeSlot, existingEntries []CalendarEntry, template *entities.BookingTemplate, endDate time.Time, loc *time.Location) []entities.TimeSlot {
	// For each slot, calculate the number of available recurrences
	for i := range slots {
		slot := &slots[i]

		// Parse slot start time
		slotStart, err := time.ParseInLocation(time.RFC3339, slot.StartTime, loc)
		if err != nil {
			continue
		}

		// Parse slot end time
		slotEnd, err := time.ParseInLocation(time.RFC3339, slot.EndTime, loc)
		if err != nil {
			continue
		}

		slotDuration := slotEnd.Sub(slotStart)

		// Calculate available recurrences for each allowed interval type
		maxRecurrences := 0
		for _, intervalType := range template.AllowedIntervals {
			recurrences := s.countAvailableRecurrences(slotStart, slotEnd, slotDuration, intervalType, template.MaxSeriesBookings, existingEntries, endDate, loc, template.BufferTime)
			if recurrences > maxRecurrences {
				maxRecurrences = recurrences
			}
		}

		slot.AvailableRecurrences = maxRecurrences
	}

	return slots
}

// countAvailableRecurrences counts how many recurrences of a specific interval type are available
func (s *FreeSlotsService) countAvailableRecurrences(slotStart, slotEnd time.Time, duration time.Duration, intervalType entities.IntervalType, maxBookings int, existingEntries []CalendarEntry, endDate time.Time, loc *time.Location, bufferTime int) int {
	count := 0
	currentStart := slotStart

	// Check up to maxBookings recurrences
	for count < maxBookings {
		// Check if we're beyond the search end date
		if currentStart.After(endDate) {
			break
		}

		currentEnd := currentStart.Add(duration)

		// Check if this slot conflicts with any existing entries
		conflicts := false
		for _, entry := range existingEntries {
			// Entry times are already parsed as *time.Time
			if entry.StartTime == nil || entry.EndTime == nil {
				continue
			}

			entryStart := *entry.StartTime
			entryEnd := *entry.EndTime

			// Apply buffer time
			bufferDuration := time.Duration(bufferTime) * time.Minute

			// Check for overlap (with buffer)
			if currentStart.Before(entryEnd.Add(bufferDuration)) && currentEnd.Add(bufferDuration).After(entryStart) {
				conflicts = true
				break
			}
		}

		if !conflicts {
			count++
		}

		// Move to next recurrence based on interval type
		var shouldContinue bool
		switch intervalType {
		case entities.IntervalWeekly:
			currentStart = currentStart.AddDate(0, 0, 7)
			shouldContinue = true
		case entities.IntervalMonthlyDate:
			// Same date next month
			currentStart = currentStart.AddDate(0, 1, 0)
			shouldContinue = true
		case entities.IntervalMonthlyDay:
			// Same weekday, same week of month
			currentStart = s.addMonthSameWeekday(currentStart)
			shouldContinue = true
		case entities.IntervalYearly:
			currentStart = currentStart.AddDate(1, 0, 0)
			shouldContinue = true
		case entities.IntervalNone:
			// Only count once for non-recurring
			shouldContinue = false
		default:
			shouldContinue = false
		}

		if !shouldContinue {
			break
		}
	}

	return count
}

// addMonthSameWeekday adds one month while keeping the same weekday and week of month
func (s *FreeSlotsService) addMonthSameWeekday(t time.Time) time.Time {
	// Get the weekday and the week of the month
	weekday := t.Weekday()
	weekOfMonth := (t.Day()-1)/7 + 1

	// Move to next month, first day
	nextMonth := time.Date(t.Year(), t.Month()+1, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	// Find the first occurrence of the same weekday in the next month
	daysUntilWeekday := (int(weekday) - int(nextMonth.Weekday()) + 7) % 7
	firstOccurrence := nextMonth.AddDate(0, 0, daysUntilWeekday)

	// Add weeks to get to the same week of the month
	targetDate := firstOccurrence.AddDate(0, 0, (weekOfMonth-1)*7)

	// If we've gone past the end of the month, use the last occurrence
	if targetDate.Month() != nextMonth.Month() {
		targetDate = targetDate.AddDate(0, 0, -7)
	}

	return targetDate
}
