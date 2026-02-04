package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
	"gorm.io/gorm"
)

// Phase 3: Test series booking creation

func setupSeriesBookingTestDB(t *testing.T) *gorm.DB {
	db := setupFreeSlotsTestDB(t)

	// Create calendar_series table
	err := db.Exec(`
		CREATE TABLE calendar_series (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			calendar_id INTEGER NOT NULL,
			title TEXT,
			participants TEXT,
			interval_type VARCHAR(50),
			interval_value INTEGER,
			start_time DATETIME,
			end_time DATETIME,
			last_date DATETIME,
			description TEXT,
			location TEXT,
			timezone VARCHAR(100),
			entry_uuid VARCHAR(255),
			external_uid VARCHAR(255),
			sequence INTEGER DEFAULT 0,
			external_calendar_uuid VARCHAR(255),
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Modify calendar_entries to add series fields
	err = db.Exec(`
		ALTER TABLE calendar_entries ADD COLUMN series_id INTEGER
	`).Error
	if err != nil {
		// Column might already exist, that's ok
	}

	err = db.Exec(`
		ALTER TABLE calendar_entries ADD COLUMN position_in_series INTEGER
	`).Error
	if err != nil {
		// Column might already exist, that's ok
	}

	return db
}

func TestCreateSeriesBooking_CreatesSeriesAndEntries(t *testing.T) {
	db := setupSeriesBookingTestDB(t)
	service := NewBookingService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 5

	baseTime := getNextWeekday(time.Thursday, 7).Add(9 * time.Hour)

	req := SeriesBookingRequest{
		TemplateID:     1,
		TenantID:       1,
		UserID:         1,
		CalendarID:     1,
		StartTime:      baseTime,
		EndTime:        baseTime.Add(1 * time.Hour),
		IntervalType:   "weekly",
		NumOccurrences: 3,
		Title:          "Recurring Meeting",
		Description:    "Weekly team sync",
		Location:       "Conference Room A",
	}

	seriesID, entries, err := service.CreateSeriesBooking(req, template)
	require.NoError(t, err)
	assert.Greater(t, seriesID, uint(0), "Series ID should be created")
	assert.Equal(t, 3, len(entries), "Should create 3 calendar entries")

	// Verify series was created in database
	var seriesCount int64
	db.Table("calendar_series").Where("id = ?", seriesID).Count(&seriesCount)
	assert.Equal(t, int64(1), seriesCount, "Series should exist in database")

	// Verify all entries link to the series
	for i, entry := range entries {
		assert.Equal(t, seriesID, *entry.SeriesID, "Entry should link to series")
		assert.Equal(t, i+1, *entry.PositionInSeries, "Entry should have correct position")
	}

	// Verify entries are spaced correctly (weekly)
	if len(entries) >= 2 {
		time1 := entries[0].StartTime
		time2 := entries[1].StartTime
		diff := time2.Sub(*time1)
		assert.Equal(t, 7*24*time.Hour, diff, "Weekly entries should be 7 days apart")
	}
}

func TestCreateSeriesBooking_RespectsMaxBookings(t *testing.T) {
	db := setupSeriesBookingTestDB(t)
	service := NewBookingService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 3 // Max is 3

	baseTime := getNextWeekday(time.Thursday, 7).Add(9 * time.Hour)

	req := SeriesBookingRequest{
		TemplateID:     1,
		TenantID:       1,
		UserID:         1,
		CalendarID:     1,
		StartTime:      baseTime,
		EndTime:        baseTime.Add(1 * time.Hour),
		IntervalType:   "weekly",
		NumOccurrences: 10, // Request 10 but template max is 3
		Title:          "Test Meeting",
	}

	_, entries, err := service.CreateSeriesBooking(req, template)
	require.NoError(t, err)
	assert.Equal(t, 3, len(entries), "Should respect template max and create only 3 entries")
}

func TestCreateSeriesBooking_StopsAtFirstConflict(t *testing.T) {
	db := setupSeriesBookingTestDB(t)
	service := NewBookingService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 5

	baseTime := getNextWeekday(time.Thursday, 7).Add(9 * time.Hour)

	// Create a conflict in week 3
	week3Conflict := baseTime.AddDate(0, 0, 14)
	week3End := week3Conflict.Add(1 * time.Hour)
	err := db.Exec(`
		INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
		VALUES (?, ?, ?, ?)
	`, 1, 1, week3Conflict, week3End).Error
	require.NoError(t, err)

	req := SeriesBookingRequest{
		TemplateID:     1,
		TenantID:       1,
		UserID:         1,
		CalendarID:     1,
		StartTime:      baseTime,
		EndTime:        baseTime.Add(1 * time.Hour),
		IntervalType:   "weekly",
		NumOccurrences: 5,
		Title:          "Test Meeting",
	}

	_, entries, err := service.CreateSeriesBooking(req, template)

	// Should either return error or create only 2 entries (weeks 1 and 2)
	if err != nil {
		assert.Contains(t, err.Error(), "conflict", "Error should mention conflict")
	} else {
		assert.Equal(t, 2, len(entries), "Should create only 2 entries before conflict")
	}
}

func TestCreateSeriesBooking_WeeklyInterval(t *testing.T) {
	db := setupSeriesBookingTestDB(t)
	service := NewBookingService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 4

	baseTime := getNextWeekday(time.Thursday, 7).Add(9 * time.Hour)

	req := SeriesBookingRequest{
		TemplateID:     1,
		TenantID:       1,
		UserID:         1,
		CalendarID:     1,
		StartTime:      baseTime,
		EndTime:        baseTime.Add(1 * time.Hour),
		IntervalType:   "weekly",
		NumOccurrences: 4,
		Title:          "Weekly Meeting",
	}

	_, entries, err := service.CreateSeriesBooking(req, template)
	require.NoError(t, err)
	require.Equal(t, 4, len(entries))

	// Verify all entries are on Thursday
	for _, entry := range entries {
		assert.Equal(t, time.Thursday, entry.StartTime.Weekday(), "All entries should be on Thursday")
	}

	// Verify they are exactly 1 week apart
	for i := 1; i < len(entries); i++ {
		diff := entries[i].StartTime.Sub(*entries[i-1].StartTime)
		assert.Equal(t, 7*24*time.Hour, diff, "Entries should be exactly 1 week apart")
	}
}

func TestCreateSeriesBooking_MonthlyDateInterval(t *testing.T) {
	db := setupSeriesBookingTestDB(t)
	service := NewBookingService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalMonthlyDate}
	template.MaxSeriesBookings = 3

	// Start on the 15th of the month
	baseTime := getNextWeekday(time.Monday, 7)
	// Adjust to 15th of current month
	baseTime = time.Date(baseTime.Year(), baseTime.Month(), 15, 10, 0, 0, 0, time.UTC)

	req := SeriesBookingRequest{
		TemplateID:     1,
		TenantID:       1,
		UserID:         1,
		CalendarID:     1,
		StartTime:      baseTime,
		EndTime:        baseTime.Add(1 * time.Hour),
		IntervalType:   "monthly-date",
		NumOccurrences: 3,
		Title:          "Monthly Meeting",
	}

	_, entries, err := service.CreateSeriesBooking(req, template)
	require.NoError(t, err)
	require.Equal(t, 3, len(entries))

	// Verify all entries are on the 15th
	for _, entry := range entries {
		assert.Equal(t, 15, entry.StartTime.Day(), "All entries should be on the 15th")
	}

	// Verify they are in consecutive months
	for i := 1; i < len(entries); i++ {
		prevMonth := entries[i-1].StartTime.Month()
		currMonth := entries[i].StartTime.Month()

		// Handle year rollover
		monthDiff := int(currMonth) - int(prevMonth)
		if monthDiff < 0 {
			monthDiff += 12
		}
		assert.Equal(t, 1, monthDiff, "Entries should be in consecutive months")
	}
}

func TestCreateSeriesBooking_ValidatesTemplateAllowsRecurrence(t *testing.T) {
	db := setupSeriesBookingTestDB(t)
	service := NewBookingService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{} // No recurrence allowed
	template.MaxSeriesBookings = 0

	baseTime := getNextWeekday(time.Thursday, 7).Add(9 * time.Hour)

	req := SeriesBookingRequest{
		TemplateID:     1,
		TenantID:       1,
		UserID:         1,
		CalendarID:     1,
		StartTime:      baseTime,
		EndTime:        baseTime.Add(1 * time.Hour),
		IntervalType:   "weekly",
		NumOccurrences: 3,
		Title:          "Test Meeting",
	}

	_, _, err := service.CreateSeriesBooking(req, template)
	assert.Error(t, err, "Should return error when template doesn't allow recurrence")
	assert.Contains(t, err.Error(), "not allow", "Error should mention recurrence not allowed")
}
