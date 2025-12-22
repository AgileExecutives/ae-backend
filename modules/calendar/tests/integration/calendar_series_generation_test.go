package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/services"
)

// TestCalendarSeriesGenerationStartsFromCorrectDate verifies that series generation
// starts from the requested start_time instead of time.Now()
func TestCalendarSeriesGenerationStartsFromCorrectDate(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(
		&entities.Calendar{},
		&entities.CalendarSeries{},
		&entities.CalendarEntry{},
	)
	require.NoError(t, err)

	// Create test data
	tenantID := uint(1)
	userID := uint(1)

	// Create a calendar
	calendar := entities.Calendar{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Calendar",
	}
	err = db.Create(&calendar).Error
	require.NoError(t, err)

	// Create service
	service := services.NewCalendarService(db)

	// Test case: Create a weekly series starting January 5, 2026 (a Sunday)
	// Should create 4 appointments: Jan 5, 12, 19, 26
	startTime := time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 1, 5, 12, 45, 0, 0, time.UTC)
	lastDate := time.Date(2026, 1, 26, 12, 0, 0, 0, time.UTC)

	req := entities.CreateCalendarSeriesRequest{
		CalendarID:    calendar.ID,
		Title:         "Weekly Therapy",
		IntervalType:  "weekly",
		IntervalValue: 1, // Every 1 week
		StartTime:     &startTime,
		EndTime:       &endTime,
		LastDate:      &lastDate,
		Description:   "Weekly recurring appointment",
		Timezone:      "UTC",
	}

	series, entries, err := service.CreateCalendarSeriesWithEntries(req, tenantID, userID)
	require.NoError(t, err)
	require.NotNil(t, series)
	require.NotNil(t, entries)

	// Verify we got exactly 4 entries (Jan 5, 12, 19, 26)
	assert.Equal(t, 4, len(entries), "Should create exactly 4 weekly appointments")

	// Verify each entry has the correct date
	expectedDates := []time.Time{
		time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 12, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 19, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 26, 12, 0, 0, 0, time.UTC),
	}

	for i, entry := range entries {
		assert.NotNil(t, entry.StartTime, "Entry %d should have a start time", i+1)
		if entry.StartTime != nil {
			assert.Equal(t, expectedDates[i], *entry.StartTime,
				"Entry %d should start on %s but got %s",
				i+1, expectedDates[i].Format("2006-01-02"), entry.StartTime.Format("2006-01-02"))
		}

		// Verify position in series
		assert.NotNil(t, entry.PositionInSeries)
		if entry.PositionInSeries != nil {
			assert.Equal(t, i+1, *entry.PositionInSeries, "Entry %d should have correct position", i+1)
		}

		// Verify series reference
		assert.NotNil(t, entry.SeriesID)
		if entry.SeriesID != nil {
			assert.Equal(t, series.ID, *entry.SeriesID, "Entry should reference the series")
		}
	}
}

// TestCalendarSeriesGenerationDoesNotStartFromToday verifies that series
// starting in the future don't incorrectly start from today
func TestCalendarSeriesGenerationDoesNotStartFromToday(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(
		&entities.Calendar{},
		&entities.CalendarSeries{},
		&entities.CalendarEntry{},
	)
	require.NoError(t, err)

	// Create test data
	tenantID := uint(1)
	userID := uint(1)

	calendar := entities.Calendar{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Calendar",
	}
	err = db.Create(&calendar).Error
	require.NoError(t, err)

	service := services.NewCalendarService(db)

	// Create a series starting far in the future (June 1, 2026)
	startTime := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 6, 1, 11, 0, 0, 0, time.UTC)
	lastDate := time.Date(2026, 6, 29, 10, 0, 0, 0, time.UTC)

	req := entities.CreateCalendarSeriesRequest{
		CalendarID:    calendar.ID,
		Title:         "Future Monthly Series",
		IntervalType:  "weekly",
		IntervalValue: 1,
		StartTime:     &startTime,
		EndTime:       &endTime,
		LastDate:      &lastDate,
	}

	series, entries, err := service.CreateCalendarSeriesWithEntries(req, tenantID, userID)
	require.NoError(t, err)
	require.NotNil(t, series)
	require.NotNil(t, entries)

	// All entries should be in June 2026, not starting from today
	for i, entry := range entries {
		assert.NotNil(t, entry.StartTime, "Entry %d should have a start time", i+1)
		if entry.StartTime != nil {
			assert.Equal(t, 2026, entry.StartTime.Year(), "Entry %d should be in year 2026", i+1)
			assert.Equal(t, time.June, entry.StartTime.Month(), "Entry %d should be in June", i+1)
			assert.True(t, entry.StartTime.Day() >= 1 && entry.StartTime.Day() <= 29,
				"Entry %d should be between June 1-29, got day %d", i+1, entry.StartTime.Day())
		}
	}

	// First entry should be exactly June 1, 2026
	if len(entries) > 0 && entries[0].StartTime != nil {
		assert.Equal(t, startTime, *entries[0].StartTime,
			"First entry should start exactly on the requested start_time")
	}
}

// TestCalendarSeriesGenerationMonthlyInterval tests monthly recurring series
func TestCalendarSeriesGenerationMonthlyInterval(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(
		&entities.Calendar{},
		&entities.CalendarSeries{},
		&entities.CalendarEntry{},
	)
	require.NoError(t, err)

	tenantID := uint(1)
	userID := uint(1)

	calendar := entities.Calendar{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Calendar",
	}
	err = db.Create(&calendar).Error
	require.NoError(t, err)

	service := services.NewCalendarService(db)

	// Create monthly series on the 15th of each month
	startTime := time.Date(2026, 1, 15, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 1, 15, 15, 0, 0, 0, time.UTC)
	lastDate := time.Date(2026, 4, 15, 14, 0, 0, 0, time.UTC)

	req := entities.CreateCalendarSeriesRequest{
		CalendarID:    calendar.ID,
		Title:         "Monthly Meeting",
		IntervalType:  "monthly-date",
		IntervalValue: 1,
		StartTime:     &startTime,
		EndTime:       &endTime,
		LastDate:      &lastDate,
	}

	series, entries, err := service.CreateCalendarSeriesWithEntries(req, tenantID, userID)
	require.NoError(t, err)
	require.NotNil(t, series)
	require.NotNil(t, entries)

	// Should create 4 entries: Jan 15, Feb 15, Mar 15, Apr 15
	assert.Equal(t, 4, len(entries), "Should create 4 monthly entries")

	expectedMonths := []time.Month{time.January, time.February, time.March, time.April}
	for i, entry := range entries {
		assert.NotNil(t, entry.StartTime)
		if entry.StartTime != nil {
			assert.Equal(t, 2026, entry.StartTime.Year())
			assert.Equal(t, expectedMonths[i], entry.StartTime.Month())
			assert.Equal(t, 15, entry.StartTime.Day(), "Should be on the 15th of each month")
			assert.Equal(t, 14, entry.StartTime.Hour())
		}
	}
}

// TestCalendarSeriesGenerationYearlyInterval tests yearly recurring series
func TestCalendarSeriesGenerationYearlyInterval(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate
	err = db.AutoMigrate(
		&entities.Calendar{},
		&entities.CalendarSeries{},
		&entities.CalendarEntry{},
	)
	require.NoError(t, err)

	tenantID := uint(1)
	userID := uint(1)

	calendar := entities.Calendar{
		TenantID: tenantID,
		UserID:   userID,
		Title:    "Test Calendar",
	}
	err = db.Create(&calendar).Error
	require.NoError(t, err)

	service := services.NewCalendarService(db)

	// Create yearly series
	startTime := time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	lastDate := time.Date(2029, 3, 20, 9, 0, 0, 0, time.UTC)

	req := entities.CreateCalendarSeriesRequest{
		CalendarID:    calendar.ID,
		Title:         "Annual Review",
		IntervalType:  "yearly",
		IntervalValue: 1,
		StartTime:     &startTime,
		EndTime:       &endTime,
		LastDate:      &lastDate,
	}

	series, entries, err := service.CreateCalendarSeriesWithEntries(req, tenantID, userID)
	require.NoError(t, err)
	require.NotNil(t, series)
	require.NotNil(t, entries)

	// Should create 4 entries: 2026, 2027, 2028, 2029
	assert.Equal(t, 4, len(entries), "Should create 4 yearly entries")

	for i, entry := range entries {
		assert.NotNil(t, entry.StartTime)
		if entry.StartTime != nil {
			assert.Equal(t, 2026+i, entry.StartTime.Year())
			assert.Equal(t, time.March, entry.StartTime.Month())
			assert.Equal(t, 20, entry.StartTime.Day())
		}
	}
}
