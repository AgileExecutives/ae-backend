package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
)

func TestGetWeeklyAvailabilityWithFallback_TemplateHasAvailability(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Create template availability
	templateAvailability := entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "09:00", End: "17:00"}},
		Friday: []entities.TimeRange{{Start: "10:00", End: "16:00"}},
	}

	result := service.getWeeklyAvailabilityWithFallback(1, 1, templateAvailability)

	// Should return template availability
	assert.Equal(t, 1, len(result.Monday))
	assert.Equal(t, "09:00", result.Monday[0].Start)
	assert.Equal(t, 1, len(result.Friday))
	assert.Equal(t, "10:00", result.Friday[0].Start)
}

func TestGetWeeklyAvailabilityWithFallback_EmptyTemplateUsesCalendar(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Create a calendar with availability
	calendarAvailability := entities.WeeklyAvailability{
		Monday:  []entities.TimeRange{{Start: "08:00", End: "18:00"}},
		Tuesday: []entities.TimeRange{{Start: "09:00", End: "17:00"}},
	}

	// Insert calendar into database
	err := db.Exec(`
		CREATE TABLE calendars (
			id INTEGER PRIMARY KEY,
			tenant_id INTEGER,
			weekly_availability JSON,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	availJSON, _ := calendarAvailability.Value()
	err = db.Exec(`
		INSERT INTO calendars (id, tenant_id, weekly_availability)
		VALUES (?, ?, ?)
	`, 1, 1, availJSON).Error
	require.NoError(t, err)

	// Empty template availability
	emptyTemplate := entities.WeeklyAvailability{}

	result := service.getWeeklyAvailabilityWithFallback(1, 1, emptyTemplate)

	// Should use calendar availability
	assert.Equal(t, 1, len(result.Monday))
	assert.Equal(t, "08:00", result.Monday[0].Start)
	assert.Equal(t, 1, len(result.Tuesday))
	assert.Equal(t, "09:00", result.Tuesday[0].Start)
}

func TestGetWeeklyAvailabilityWithFallback_NoDataUsesDefault(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Create calendars table but no data
	err := db.Exec(`
		CREATE TABLE calendars (
			id INTEGER PRIMARY KEY,
			tenant_id INTEGER,
			weekly_availability JSON,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Empty template availability
	emptyTemplate := entities.WeeklyAvailability{}

	result := service.getWeeklyAvailabilityWithFallback(999, 1, emptyTemplate)

	// Should use default all-day availability
	assert.Equal(t, 1, len(result.Monday))
	assert.Equal(t, "00:00", result.Monday[0].Start)
	assert.Equal(t, "23:59", result.Monday[0].End)

	assert.Equal(t, 1, len(result.Tuesday))
	assert.Equal(t, "00:00", result.Tuesday[0].Start)

	assert.Equal(t, 1, len(result.Sunday))
	assert.Equal(t, "00:00", result.Sunday[0].Start)
	assert.Equal(t, "23:59", result.Sunday[0].End)
}

func TestHasAvailability(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Empty availability
	empty := entities.WeeklyAvailability{}
	assert.False(t, service.hasAvailability(empty))

	// Has Monday availability
	withMonday := entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "09:00", End: "17:00"}},
	}
	assert.True(t, service.hasAvailability(withMonday))

	// Has Sunday availability
	withSunday := entities.WeeklyAvailability{
		Sunday: []entities.TimeRange{{Start: "10:00", End: "14:00"}},
	}
	assert.True(t, service.hasAvailability(withSunday))
}

func TestGetDefaultAllDayAvailability(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	result := service.getDefaultAllDayAvailability()

	// All days should have 00:00-23:59
	days := [][]entities.TimeRange{
		result.Monday, result.Tuesday, result.Wednesday,
		result.Thursday, result.Friday, result.Saturday, result.Sunday,
	}

	for _, day := range days {
		assert.Equal(t, 1, len(day))
		assert.Equal(t, "00:00", day[0].Start)
		assert.Equal(t, "23:59", day[0].End)
	}
}

func TestCalculateFreeSlots_UsesCalendarFallback(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Setup calendars table
	err := db.Exec(`
		CREATE TABLE calendars (
			id INTEGER PRIMARY KEY,
			tenant_id INTEGER,
			weekly_availability JSON,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Calendar with specific availability
	calendarAvailability := entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "10:00", End: "12:00"}},
	}
	availJSON, _ := calendarAvailability.Value()
	err = db.Exec(`
		INSERT INTO calendars (id, tenant_id, weekly_availability)
		VALUES (?, ?, ?)
	`, 1, 1, availJSON).Error
	require.NoError(t, err)

	// Template with NO availability (should fallback to calendar)
	template := &entities.BookingTemplate{
		SlotDuration:       30,
		BufferTime:         0,
		MinNoticeHours:     0,
		AdvanceBookingDays: 365,
		WeeklyAvailability: entities.WeeklyAvailability{}, // Empty
		AllowedIntervals:   []entities.IntervalType{"none"},
		MaxSeriesBookings:  0,
	}

	startDate := getNextWeekday(time.Monday, 7)
	endDate := startDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  startDate,
		EndDate:    endDate,
		Timezone:   "UTC",
	}

	result, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)

	// Should have slots only in the 10:00-12:00 window
	assert.Greater(t, len(result.Slots), 0, "Should have generated slots")

	for _, slot := range result.Slots {
		slotTime, _ := time.Parse(time.RFC3339, slot.StartTime)
		hour := slotTime.Hour()
		assert.GreaterOrEqual(t, hour, 10, "Slot should be at or after 10:00")
		assert.Less(t, hour, 12, "Slot should be before 12:00")
	}

	// Config should reflect the calendar availability
	assert.Equal(t, 1, len(result.Config.WeeklyAvailability.Monday))
	assert.Equal(t, "10:00", result.Config.WeeklyAvailability.Monday[0].Start)
}
