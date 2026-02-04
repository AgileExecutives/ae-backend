package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/unburdy/booking-module/entities"
)

// CalendarSeries represents a recurring series in the database
type CalendarSeries struct {
	ID            uint
	TenantID      uint
	UserID        uint
	CalendarID    uint
	Title         string
	Description   string
	Location      string
	IntervalType  string
	IntervalValue int
	StartTime     *time.Time
	EndTime       *time.Time
	LastDate      *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SeriesBookingRequest represents a request to create a recurring booking
type SeriesBookingRequest struct {
	TemplateID     uint
	TenantID       uint
	UserID         uint
	CalendarID     uint
	StartTime      time.Time
	EndTime        time.Time
	IntervalType   string // "weekly", "monthly-date", "monthly-day", "yearly"
	NumOccurrences int    // How many occurrences to book
	Title          string
	Description    string
	Location       string
}

// CreateSeriesBooking creates a recurring booking series
func (s *BookingService) CreateSeriesBooking(req SeriesBookingRequest, template *entities.BookingTemplate) (uint, []CalendarEntry, error) {
	// Validate that template allows recurrence
	if template.MaxSeriesBookings == 0 || len(template.AllowedIntervals) == 0 {
		return 0, nil, errors.New("template does not allow recurring bookings")
	}

	// Validate that requested interval is allowed
	intervalAllowed := false
	for _, allowed := range template.AllowedIntervals {
		if string(allowed) == req.IntervalType {
			intervalAllowed = true
			break
		}
	}
	if !intervalAllowed {
		return 0, nil, fmt.Errorf("interval type '%s' is not allowed by template", req.IntervalType)
	}

	// Respect max bookings limit
	maxOccurrences := req.NumOccurrences
	if template.MaxSeriesBookings > 0 && maxOccurrences > template.MaxSeriesBookings {
		maxOccurrences = template.MaxSeriesBookings
	}

	// Create the series record
	series := CalendarSeries{
		TenantID:      req.TenantID,
		UserID:        req.UserID,
		CalendarID:    req.CalendarID,
		Title:         req.Title,
		Description:   req.Description,
		Location:      req.Location,
		IntervalType:  req.IntervalType,
		IntervalValue: 1, // Default to 1 (every week, every month, etc.)
		StartTime:     &req.StartTime,
		EndTime:       &req.EndTime,
	}

	if err := s.db.Table("calendar_series").Create(&series).Error; err != nil {
		return 0, nil, fmt.Errorf("failed to create calendar series: %w", err)
	}

	// Generate entries for each occurrence
	entries := make([]CalendarEntry, 0, maxOccurrences)
	currentStart := req.StartTime
	currentEnd := req.EndTime
	duration := req.EndTime.Sub(req.StartTime)

	for i := 0; i < maxOccurrences; i++ {
		// Check for conflicts before creating entry
		hasConflict, err := s.hasConflict(req.CalendarID, req.TenantID, currentStart, currentEnd)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to check conflicts: %w", err)
		}

		if hasConflict {
			// Stop at first conflict
			if len(entries) == 0 {
				return 0, nil, errors.New("conflict detected at first occurrence")
			}
			break
		}

		// Create calendar entry - make copies of times to avoid pointer issues
		position := i + 1
		entryStart := currentStart
		entryEnd := currentEnd
		entry := CalendarEntry{
			CalendarID:       req.CalendarID,
			TenantID:         req.TenantID,
			StartTime:        &entryStart,
			EndTime:          &entryEnd,
			SeriesID:         &series.ID,
			PositionInSeries: &position,
		}

		if err := s.db.Table("calendar_entries").Create(&entry).Error; err != nil {
			return 0, nil, fmt.Errorf("failed to create calendar entry: %w", err)
		}

		entries = append(entries, entry)

		// Calculate next occurrence based on interval type
		switch req.IntervalType {
		case "weekly":
			currentStart = currentStart.AddDate(0, 0, 7)
		case "monthly-date":
			// Keep the same day of month
			currentStart = currentStart.AddDate(0, 1, 0)
		case "monthly-day":
			// Keep the same weekday (e.g., 2nd Thursday)
			currentStart = addMonthSameWeekday(currentStart)
		case "yearly":
			currentStart = currentStart.AddDate(1, 0, 0)
		default:
			return 0, nil, fmt.Errorf("unsupported interval type: %s", req.IntervalType)
		}
		currentEnd = currentStart.Add(duration)
	}

	// Update series last_date
	if len(entries) > 0 {
		lastEntry := entries[len(entries)-1]
		series.LastDate = lastEntry.EndTime
		s.db.Table("calendar_series").Where("id = ?", series.ID).Update("last_date", series.LastDate)
	}

	return series.ID, entries, nil
}

// hasConflict checks if there's a conflict for the given time slot
func (s *BookingService) hasConflict(calendarID uint, tenantID uint, startTime, endTime time.Time) (bool, error) {
	var count int64
	err := s.db.Table("calendar_entries").
		Where("calendar_id = ?", calendarID).
		Where("tenant_id = ?", tenantID).
		Where("start_time < ?", endTime).
		Where("end_time > ?", startTime).
		Where("deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// addMonthSameWeekday adds one month while keeping the same weekday position
// E.g., 2nd Thursday of the month -> 2nd Thursday of next month
func addMonthSameWeekday(t time.Time) time.Time {
	// Calculate which occurrence this is (1st, 2nd, 3rd, etc.)
	weekdayOccurrence := (t.Day()-1)/7 + 1

	// Move to next month
	nextMonth := t.AddDate(0, 1, 0)
	// Go to first day of that month
	firstOfMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())

	// Find the first occurrence of the target weekday
	targetWeekday := t.Weekday()
	daysUntilWeekday := (int(targetWeekday) - int(firstOfMonth.Weekday()) + 7) % 7
	firstOccurrence := firstOfMonth.AddDate(0, 0, daysUntilWeekday)

	// Add weeks to get to the nth occurrence
	result := firstOccurrence.AddDate(0, 0, 7*(weekdayOccurrence-1))

	return result
}
