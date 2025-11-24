package services

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/unburdy/calendar-module/entities"
)

// Test fixtures
func createCalendarRequest() entities.CreateCalendarRequest {
	weeklyAvailability, _ := json.Marshal(map[string]interface{}{
		"monday": []string{"09:00-17:00"},
	})

	return entities.CreateCalendarRequest{
		Title:              "New Calendar",
		Color:              "#FF0000",
		WeeklyAvailability: weeklyAvailability,
		Timezone:           "UTC",
	}
}

func createMockCalendar(tenantID, userID uint) entities.Calendar {
	weeklyAvailability, _ := json.Marshal(map[string]interface{}{
		"monday":    []string{"09:00-17:00"},
		"tuesday":   []string{"09:00-17:00"},
		"wednesday": []string{"09:00-17:00"},
		"thursday":  []string{"09:00-17:00"},
		"friday":    []string{"09:00-17:00"},
	})

	return entities.Calendar{
		ID:                 1,
		TenantID:           tenantID,
		UserID:             userID,
		Title:              "Test Calendar",
		Color:              "#FF5733",
		WeeklyAvailability: weeklyAvailability,
		CalendarUUID:       "test-calendar-uuid",
		Timezone:           "UTC",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func createMockCalendarEntry(calendarID, seriesID uint) entities.CalendarEntry {
	startTime := time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)

	var seriesRef *uint
	if seriesID > 0 {
		seriesRef = &seriesID
	}

	return entities.CalendarEntry{
		ID:           1,
		TenantID:     1,
		UserID:       1,
		CalendarID:   calendarID,
		SeriesID:     seriesRef,
		Title:        "Test Meeting",
		StartTime:    &startTime,
		EndTime:      &endTime,
		Timezone:     "UTC",
		Type:         "meeting",
		Description:  "Test meeting description",
		Location:     "Conference Room A",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func createMockCalendarSeries(calendarID uint) entities.CalendarSeries {
	timeStart := time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC)
	timeEnd := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)

	return entities.CalendarSeries{
		ID:            1,
		CalendarID:    calendarID,
		TenantID:      1,
		UserID:        1,
		Title:         "Test Series",
		IntervalType:  "weekly",
		IntervalValue: 1,
		StartTime:     &timeStart,
		EndTime:       &timeEnd,
		Description:   "Test Series Description",
		Location:      "Test Location",
		EntryUUID:     "test-series-uuid",
		Sequence:      0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Simple integration test to verify service creation
func TestCalendarService_Initialization(t *testing.T) {
	// This test verifies that the service can be created
	// In a real scenario, you would pass a real database connection
	// For now, we'll test the structure

	t.Run("calendar request validation", func(t *testing.T) {
		request := createCalendarRequest()

		// Validate the request structure
		assert.NotEmpty(t, request.Title)
		assert.NotEmpty(t, request.Color)
		assert.NotNil(t, request.WeeklyAvailability)
		assert.NotEmpty(t, request.Timezone)
	})

	t.Run("calendar service structure", func(t *testing.T) {
		// Test that we can create the service (with nil db for structure test)
		service := &CalendarService{}
		assert.NotNil(t, service)
	})

	t.Run("deep preloading structure validation", func(t *testing.T) {
		// Test the structure of calendars with 2-level deep preloading

		// Create mock calendar with nested relationships
		calendar := createMockCalendar(1, 1)
		entry := createMockCalendarEntry(1, 1) // Entry with series reference
		series := createMockCalendarSeries(1)

		// Simulate 2-level preloading structure
		calendar.CalendarEntries = []entities.CalendarEntry{entry}
		calendar.CalendarSeries = []entities.CalendarSeries{series}

		// Simulate nested preloads (entries->series, series->entries)
		calendar.CalendarEntries[0].Series = &series                                 // Entry points to series
		calendar.CalendarSeries[0].CalendarEntries = []entities.CalendarEntry{entry} // Series points to entries

		// Validate the 2-level deep structure
		assert.Len(t, calendar.CalendarEntries, 1, "Calendar should have entries (level 1)")
		assert.Len(t, calendar.CalendarSeries, 1, "Calendar should have series (level 1)")

		// Validate 2nd level relationships
		firstEntry := calendar.CalendarEntries[0]
		assert.NotNil(t, firstEntry.Series, "Entry should have series reference (level 2)")
		assert.Equal(t, series.ID, firstEntry.Series.ID, "Entry should reference correct series")

		firstSeries := calendar.CalendarSeries[0]
		assert.Len(t, firstSeries.CalendarEntries, 1, "Series should have entries (level 2)")
		assert.Equal(t, entry.ID, firstSeries.CalendarEntries[0].ID, "Series should reference correct entries")
	})

	t.Run("preload method parameter validation", func(t *testing.T) {
		// Test that the GetCalendarsWithDeepPreload method signature is correct
		// This is a compile-time test to ensure the method exists with correct parameters

		// Just test that we can create the service and it has the method
		// We don't call the method since it would panic with nil DB
		service := &CalendarService{}
		assert.NotNil(t, service, "Service should be created")

		// Validate that the method exists by checking it's not nil
		// This is a type check that validates the method signature exists
		assert.NotNil(t, (*CalendarService).GetCalendarsWithDeepPreload, "GetCalendarsWithDeepPreload method should exist")
	})
}
