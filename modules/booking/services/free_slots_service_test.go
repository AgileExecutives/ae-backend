package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupFreeSlotsTestDB creates an in-memory SQLite database for testing
func setupFreeSlotsTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create calendar_entries table
	err = db.Exec(`
		CREATE TABLE calendar_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			calendar_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			start_time DATETIME,
			end_time DATETIME,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	return db
}

// createTestTemplate creates a basic booking template for testing
func createTestTemplate() *entities.BookingTemplate {
	return &entities.BookingTemplate{
		ID:                 1,
		UserID:             1,
		CalendarID:         1,
		TenantID:           1,
		SlotDuration:       30,
		BufferTime:         10,
		MaxSeriesBookings:  5,
		AdvanceBookingDays: 90,
		MinNoticeHours:     24,
		Timezone:           "UTC",
		WeeklyAvailability: entities.WeeklyAvailability{
			Monday: []entities.TimeRange{
				{Start: "09:00", End: "12:00"},
				{Start: "14:00", End: "17:00"},
			},
			Tuesday: []entities.TimeRange{
				{Start: "09:00", End: "17:00"},
			},
			Wednesday: []entities.TimeRange{
				{Start: "10:00", End: "16:00"},
			},
			Thursday: []entities.TimeRange{},
			Friday: []entities.TimeRange{
				{Start: "09:00", End: "12:00"},
			},
			Saturday: []entities.TimeRange{},
			Sunday:   []entities.TimeRange{},
		},
		AllowedIntervals: entities.IntervalArray{entities.IntervalWeekly},
		BlockDates:       entities.DateRangeArray{},
	}
}

func TestCalculateFreeSlots_BasicSlotGeneration(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()

	// Set a specific date range for testing (Monday to Wednesday)
	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2026, 2, 4, 23, 59, 59, 0, time.UTC) // Wednesday

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  startDate,
		EndDate:    endDate,
		Timezone:   "UTC",
	}

	// Override min notice to allow testing past dates
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	result, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify slots are generated
	assert.NotEmpty(t, result.Slots, "Should generate some slots")

	// Verify configuration
	assert.Equal(t, 30, result.Config.Duration)
	assert.Equal(t, 10, result.Config.BufferTime)
	assert.Equal(t, "weekly", result.Config.Interval)

	// Verify month data
	assert.Equal(t, 2026, result.MonthData.Year)
	assert.Equal(t, 2, result.MonthData.Month)
	assert.NotEmpty(t, result.MonthData.Days)
}

func TestGenerateAllSlots_RespectWeeklyAvailability(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()

	// Test for a full week
	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2026, 2, 8, 23, 59, 59, 0, time.UTC) // Sunday

	req := FreeSlotsRequest{
		StartDate: startDate,
		EndDate:   endDate,
		Timezone:  "UTC",
	}

	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	slots := service.generateAllSlots(req, template, time.UTC, template.WeeklyAvailability)

	// Group slots by day of week
	slotsByDay := make(map[string]int)
	for _, slot := range slots {
		slotTime, _ := time.Parse(time.RFC3339, slot.StartTime)
		dayName := slotTime.Weekday().String()
		slotsByDay[dayName]++
	}

	// Monday: 09:00-12:00 (3h) + 14:00-17:00 (3h) = 6h = 9 slots (30min each + 10min buffer = 40min per slot)
	assert.Greater(t, slotsByDay["Monday"], 0, "Monday should have slots")

	// Tuesday: 09:00-17:00 (8h) = more slots
	assert.Greater(t, slotsByDay["Tuesday"], slotsByDay["Monday"], "Tuesday should have more slots than Monday")

	// Wednesday: 10:00-16:00 (6h)
	assert.Greater(t, slotsByDay["Wednesday"], 0, "Wednesday should have slots")

	// Thursday: no availability
	assert.Equal(t, 0, slotsByDay["Thursday"], "Thursday should have no slots")

	// Friday: 09:00-12:00 (3h)
	assert.Greater(t, slotsByDay["Friday"], 0, "Friday should have slots")

	// Saturday and Sunday: no availability
	assert.Equal(t, 0, slotsByDay["Saturday"], "Saturday should have no slots")
	assert.Equal(t, 0, slotsByDay["Sunday"], "Sunday should have no slots")
}

func TestGenerateAllSlots_SlotDurationAndBuffer(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.SlotDuration = 60 // 1 hour
	template.BufferTime = 15   // 15 minutes

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

	req := FreeSlotsRequest{
		StartDate: startDate,
		EndDate:   endDate,
		Timezone:  "UTC",
	}

	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	slots := service.generateAllSlots(req, template, time.UTC, template.WeeklyAvailability)

	if len(slots) > 1 {
		// Verify slot duration
		slot := slots[0]
		slotStart, _ := time.Parse(time.RFC3339, slot.StartTime)
		slotEnd, _ := time.Parse(time.RFC3339, slot.EndTime)
		duration := slotEnd.Sub(slotStart)
		assert.Equal(t, 60*time.Minute, duration, "Slot duration should be 60 minutes")

		// Verify buffer between slots (slot end + buffer = next slot start)
		if len(slots) > 1 {
			nextSlot := slots[1]
			nextSlotStart, _ := time.Parse(time.RFC3339, nextSlot.StartTime)
			timeBetween := nextSlotStart.Sub(slotEnd)
			assert.Equal(t, 15*time.Minute, timeBetween, "Buffer time should be 15 minutes")
		}
	}
}

func TestFilterConflictingSlots_RemovesConflicts(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Create test slots
	baseTime := time.Date(2026, 2, 2, 9, 0, 0, 0, time.UTC)
	slots := []entities.TimeSlot{
		{
			ID:        "slot-1",
			StartTime: baseTime.Format(time.RFC3339),
			EndTime:   baseTime.Add(30 * time.Minute).Format(time.RFC3339),
			Available: true,
		},
		{
			ID:        "slot-2",
			StartTime: baseTime.Add(50 * time.Minute).Format(time.RFC3339), // 09:50
			EndTime:   baseTime.Add(80 * time.Minute).Format(time.RFC3339), // 10:20
			Available: true,
		},
		{
			ID:        "slot-3",
			StartTime: baseTime.Add(150 * time.Minute).Format(time.RFC3339), // 11:30
			EndTime:   baseTime.Add(180 * time.Minute).Format(time.RFC3339), // 12:00
			Available: true,
		},
	}

	// Create a conflicting calendar entry that overlaps with slot-2 (10:00-10:30)
	conflictStart := baseTime.Add(60 * time.Minute) // 10:00
	conflictEnd := baseTime.Add(90 * time.Minute)   // 10:30
	existingEntries := []CalendarEntry{
		{
			ID:         1,
			CalendarID: 1,
			TenantID:   1,
			StartTime:  &conflictStart,
			EndTime:    &conflictEnd,
		},
	}

	bufferTime := 10
	result := service.filterConflictingSlots(slots, existingEntries, bufferTime)

	// Should have only 2 slots (slot-1 and slot-3), slot-2 should be filtered out
	assert.Equal(t, 2, len(result), "Should filter out conflicting slot")
	assert.Equal(t, "slot-1", result[0].ID)
	assert.Equal(t, "slot-3", result[1].ID)
}

func TestFilterConflictingSlots_BufferTime(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	baseTime := time.Date(2026, 2, 2, 9, 0, 0, 0, time.UTC)

	// Slot from 09:00 to 09:30
	slots := []entities.TimeSlot{
		{
			ID:        "slot-1",
			StartTime: baseTime.Format(time.RFC3339),
			EndTime:   baseTime.Add(30 * time.Minute).Format(time.RFC3339),
			Available: true,
		},
	}

	// Entry from 09:35 to 09:45 (starts 5 minutes after slot ends)
	entryStart := baseTime.Add(35 * time.Minute)
	entryEnd := baseTime.Add(45 * time.Minute)
	existingEntries := []CalendarEntry{
		{
			ID:         1,
			CalendarID: 1,
			TenantID:   1,
			StartTime:  &entryStart,
			EndTime:    &entryEnd,
		},
	}

	// With 10 minute buffer, slot should be filtered (09:30 + 10min buffer = 09:40, overlaps with 09:35 entry)
	result := service.filterConflictingSlots(slots, existingEntries, 10)
	assert.Equal(t, 0, len(result), "Slot should be filtered due to buffer time")

	// With 4 minute buffer, slot should remain (09:30 + 4min buffer = 09:34, no overlap with 09:35 entry)
	result = service.filterConflictingSlots(slots, existingEntries, 4)
	assert.Equal(t, 1, len(result), "Slot should remain with smaller buffer")
}

func TestIsSlotBookable_MinNoticeHours(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Slot in the past
	pastSlot := time.Now().Add(-1 * time.Hour)
	assert.False(t, service.isSlotBookable(pastSlot, 90, 24), "Past slots should not be bookable")

	// Slot in 12 hours (less than 24h min notice)
	nearFutureSlot := time.Now().Add(12 * time.Hour)
	assert.False(t, service.isSlotBookable(nearFutureSlot, 90, 24), "Slot within min notice period should not be bookable")

	// Slot in 48 hours (more than 24h min notice)
	futureSlot := time.Now().Add(48 * time.Hour)
	assert.True(t, service.isSlotBookable(futureSlot, 90, 24), "Slot after min notice period should be bookable")
}

func TestIsSlotBookable_AdvanceBookingDays(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Slot 100 days in future (beyond 90 day advance booking limit)
	farFutureSlot := time.Now().Add(100 * 24 * time.Hour)
	assert.False(t, service.isSlotBookable(farFutureSlot, 90, 0), "Slot beyond advance booking limit should not be bookable")

	// Slot 60 days in future (within 90 day advance booking limit)
	futureSlot := time.Now().Add(60 * 24 * time.Hour)
	assert.True(t, service.isSlotBookable(futureSlot, 90, 0), "Slot within advance booking limit should be bookable")
}

func TestIsDateBlocked(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	blockedDates := []entities.DateRange{
		{Start: "2026-12-24", End: "2026-12-26"}, // Christmas
		{Start: "2026-12-31", End: "2026-01-01"}, // New Year
	}

	// Date within blocked range
	blockedDate := time.Date(2026, 12, 25, 10, 0, 0, 0, time.UTC)
	assert.True(t, service.isDateBlocked(blockedDate, blockedDates), "Christmas should be blocked")

	// Date on start boundary
	boundaryStart := time.Date(2026, 12, 24, 10, 0, 0, 0, time.UTC)
	assert.True(t, service.isDateBlocked(boundaryStart, blockedDates), "Start boundary should be blocked")

	// Date on end boundary
	boundaryEnd := time.Date(2026, 12, 26, 10, 0, 0, 0, time.UTC)
	assert.True(t, service.isDateBlocked(boundaryEnd, blockedDates), "End boundary should be blocked")

	// Date not in blocked range
	normalDate := time.Date(2026, 2, 20, 10, 0, 0, 0, time.UTC)
	assert.False(t, service.isDateBlocked(normalDate, blockedDates), "Normal date should not be blocked")
}

func TestGetWeekdayAvailability(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	weeklyAvail := entities.WeeklyAvailability{
		Monday: []entities.TimeRange{
			{Start: "09:00", End: "12:00"},
		},
		Tuesday: []entities.TimeRange{
			{Start: "10:00", End: "16:00"},
		},
		Wednesday: []entities.TimeRange{},
	}

	// Test Monday
	monday := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	mondayAvail := service.getWeekdayAvailability(monday, weeklyAvail)
	assert.Len(t, mondayAvail, 1)
	assert.Equal(t, "09:00", mondayAvail[0].Start)
	assert.Equal(t, "12:00", mondayAvail[0].End)

	// Test Tuesday
	tuesday := time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC)
	tuesdayAvail := service.getWeekdayAvailability(tuesday, weeklyAvail)
	assert.Len(t, tuesdayAvail, 1)
	assert.Equal(t, "10:00", tuesdayAvail[0].Start)

	// Test Wednesday (no availability)
	wednesday := time.Date(2026, 2, 4, 0, 0, 0, 0, time.UTC)
	wednesdayAvail := service.getWeekdayAvailability(wednesday, weeklyAvail)
	assert.Len(t, wednesdayAvail, 0)
}

func TestGenerateMonthData(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Create some test slots for different dates
	slots := []entities.TimeSlot{
		{Date: "2026-02-02", StartTime: "2026-02-02T09:00:00Z", EndTime: "2026-02-02T09:30:00Z"},
		{Date: "2026-02-02", StartTime: "2026-02-02T10:00:00Z", EndTime: "2026-02-02T10:30:00Z"},
		{Date: "2026-02-03", StartTime: "2026-02-03T09:00:00Z", EndTime: "2026-02-03T09:30:00Z"},
		// No slots for 2026-02-04
	}

	startDate := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	monthData := service.generateMonthData(slots, startDate, time.UTC)

	assert.Equal(t, 2026, monthData.Year)
	assert.Equal(t, 2, monthData.Month)
	assert.Equal(t, 28, len(monthData.Days), "February 2026 should have 28 days")

	// Check specific dates
	var day02, day03, day04 entities.DayData
	for _, day := range monthData.Days {
		switch day.Date {
		case "2026-02-02":
			day02 = day
		case "2026-02-03":
			day03 = day
		case "2026-02-04":
			day04 = day
		}
	}

	assert.Equal(t, 2, day02.AvailableCount, "Feb 02 should have 2 slots")
	assert.NotEqual(t, "none", day02.Status)

	assert.Equal(t, 1, day03.AvailableCount, "Feb 03 should have 1 slot")

	assert.Equal(t, 0, day04.AvailableCount, "Feb 04 should have 0 slots")
	assert.Equal(t, "none", day04.Status)
}

func TestCalculateFreeSlots_WithExistingEntries(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()

	// Insert a calendar entry that will conflict
	entryStart := time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC)
	entryEnd := time.Date(2026, 2, 2, 11, 0, 0, 0, time.UTC)

	err := db.Exec(`
		INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
		VALUES (?, ?, ?, ?)
	`, 1, 1, entryStart, entryEnd).Error
	require.NoError(t, err)

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  startDate,
		EndDate:    endDate,
		Timezone:   "UTC",
	}

	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	result, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)

	// Verify that slots overlapping with 10:00-11:00 are filtered out
	for _, slot := range result.Slots {
		slotStart, _ := time.Parse(time.RFC3339, slot.StartTime)
		slotEnd, _ := time.Parse(time.RFC3339, slot.EndTime)

		// Check that no slot overlaps with the existing entry (considering buffer)
		// Slot should end before entry starts (with buffer) OR start after entry ends (with buffer)
		bufferDuration := time.Duration(template.BufferTime) * time.Minute
		assert.True(t,
			slotEnd.Add(bufferDuration).Before(entryStart) || slotEnd.Add(bufferDuration).Equal(entryStart) ||
				slotStart.Add(-bufferDuration).After(entryEnd) || slotStart.Add(-bufferDuration).Equal(entryEnd),
			"Slot should not overlap with existing entry considering buffer time")
	}
}

func TestClassifyTimeOfDay(t *testing.T) {
	assert.Equal(t, "morning", entities.ClassifyTimeOfDay(8))
	assert.Equal(t, "morning", entities.ClassifyTimeOfDay(11))
	assert.Equal(t, "afternoon", entities.ClassifyTimeOfDay(12))
	assert.Equal(t, "afternoon", entities.ClassifyTimeOfDay(17))
	assert.Equal(t, "evening", entities.ClassifyTimeOfDay(18))
	assert.Equal(t, "evening", entities.ClassifyTimeOfDay(23))
}

func TestDayStatus(t *testing.T) {
	// No slots available
	assert.Equal(t, "none", entities.DayStatus(0, 10))

	// More than 50% available
	assert.Equal(t, "available", entities.DayStatus(8, 10))

	// 50% or less available
	assert.Equal(t, "partial", entities.DayStatus(5, 10))
	assert.Equal(t, "partial", entities.DayStatus(3, 10))
}

func TestDetermineInterval(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Weekly interval
	assert.Equal(t, "weekly", service.determineInterval(entities.IntervalArray{entities.IntervalWeekly}))

	// Monthly interval
	assert.Equal(t, "monthly", service.determineInterval(entities.IntervalArray{entities.IntervalMonthlyDate}))

	// Yearly interval
	assert.Equal(t, "yearly", service.determineInterval(entities.IntervalArray{entities.IntervalYearly}))

	// None interval
	assert.Equal(t, "none", service.determineInterval(entities.IntervalArray{entities.IntervalNone}))

	// Empty array
	assert.Equal(t, "none", service.determineInterval(entities.IntervalArray{}))
}

func TestCalculateFreeSlots_TimezoneHandling(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.Timezone = "America/New_York"

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  startDate,
		EndDate:    endDate,
		Timezone:   "America/New_York",
	}

	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	result, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)

	// Verify slots have the correct timezone
	for _, slot := range result.Slots {
		assert.Equal(t, "America/New_York", slot.Timezone)
	}
}

func TestCalculateFreeSlots_InvalidTimezone(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  startDate,
		EndDate:    endDate,
		Timezone:   "Invalid/Timezone",
	}

	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	// Should not error, should fall back to UTC
	result, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestCalculateFreeSlots_MultipleAvailabilityWindows(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()

	// Override to have multiple windows on Monday
	template.WeeklyAvailability.Monday = []entities.TimeRange{
		{Start: "08:00", End: "10:00"},
		{Start: "11:00", End: "13:00"},
		{Start: "15:00", End: "17:00"},
	}

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  startDate,
		EndDate:    endDate,
		Timezone:   "UTC",
	}

	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	slots := service.generateAllSlots(req, template, time.UTC, template.WeeklyAvailability)

	// Verify slots are generated in all windows
	hasEarlyMorning := false // 08:00-10:00
	hasLateMorning := false  // 11:00-13:00
	hasAfternoon := false    // 15:00-17:00

	for _, slot := range slots {
		slotTime, _ := time.Parse(time.RFC3339, slot.StartTime)
		hour := slotTime.Hour()

		if hour >= 8 && hour < 10 {
			hasEarlyMorning = true
		}
		if hour >= 11 && hour < 13 {
			hasLateMorning = true
		}
		if hour >= 15 && hour < 17 {
			hasAfternoon = true
		}
	}

	assert.True(t, hasEarlyMorning, "Should have slots in 08:00-10:00 window")
	assert.True(t, hasLateMorning, "Should have slots in 11:00-13:00 window")
	assert.True(t, hasAfternoon, "Should have slots in 15:00-17:00 window")
}

// Test fallback behavior: Template has availability
func TestCalculateFreeSlots_UseTemplateAvailability(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Setup calendars table with different availability
	err := db.Exec(`
		CREATE TABLE calendars (
			id INTEGER PRIMARY KEY,
			tenant_id INTEGER,
			weekly_availability JSON,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Calendar has 08:00-20:00 availability
	calendarAvailability := entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "08:00", End: "20:00"}},
	}
	availJSON, _ := calendarAvailability.Value()
	err = db.Exec(`
		INSERT INTO calendars (id, tenant_id, weekly_availability)
		VALUES (?, ?, ?)
	`, 1, 1, availJSON).Error
	require.NoError(t, err)

	// Template has 10:00-12:00 availability (should take precedence)
	template := createTestTemplate()
	template.WeeklyAvailability = entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "10:00", End: "12:00"}},
	}
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

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
	assert.NotEmpty(t, result.Slots, "Should have slots")

	// Verify slots are only in template's 10:00-12:00 window, not calendar's 08:00-20:00
	for _, slot := range result.Slots {
		slotTime, _ := time.Parse(time.RFC3339, slot.StartTime)
		hour := slotTime.Hour()
		assert.GreaterOrEqual(t, hour, 10, "Slot should be at or after 10:00")
		assert.Less(t, hour, 12, "Slot should be before 12:00")
	}

	// Verify config reflects template availability
	assert.Equal(t, 1, len(result.Config.WeeklyAvailability.Monday))
	assert.Equal(t, "10:00", result.Config.WeeklyAvailability.Monday[0].Start)
	assert.Equal(t, "12:00", result.Config.WeeklyAvailability.Monday[0].End)
}

// Test fallback behavior: Template is empty, use calendar
func TestCalculateFreeSlots_FallbackToCalendarAvailability(t *testing.T) {
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

	// Calendar has specific availability
	calendarAvailability := entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "13:00", End: "15:00"}},
	}
	availJSON, _ := calendarAvailability.Value()
	err = db.Exec(`
		INSERT INTO calendars (id, tenant_id, weekly_availability)
		VALUES (?, ?, ?)
	`, 1, 1, availJSON).Error
	require.NoError(t, err)

	// Template with EMPTY availability
	template := createTestTemplate()
	template.WeeklyAvailability = entities.WeeklyAvailability{} // Empty
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

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
	assert.NotEmpty(t, result.Slots, "Should have slots from calendar availability")

	// Verify slots are only in calendar's 13:00-15:00 window
	for _, slot := range result.Slots {
		slotTime, _ := time.Parse(time.RFC3339, slot.StartTime)
		hour := slotTime.Hour()
		assert.GreaterOrEqual(t, hour, 13, "Slot should be at or after 13:00")
		assert.Less(t, hour, 15, "Slot should be before 15:00")
	}

	// Verify config reflects calendar availability
	assert.Equal(t, 1, len(result.Config.WeeklyAvailability.Monday))
	assert.Equal(t, "13:00", result.Config.WeeklyAvailability.Monday[0].Start)
	assert.Equal(t, "15:00", result.Config.WeeklyAvailability.Monday[0].End)
}

// Test fallback behavior: Both template and calendar empty, use default all-day
func TestCalculateFreeSlots_FallbackToDefaultAllDay(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Setup calendars table with empty availability
	err := db.Exec(`
		CREATE TABLE calendars (
			id INTEGER PRIMARY KEY,
			tenant_id INTEGER,
			weekly_availability JSON,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Calendar with empty availability
	err = db.Exec(`
		INSERT INTO calendars (id, tenant_id, weekly_availability)
		VALUES (?, ?, ?)
	`, 1, 1, nil).Error
	require.NoError(t, err)

	// Template with EMPTY availability
	template := createTestTemplate()
	template.WeeklyAvailability = entities.WeeklyAvailability{} // Empty
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365
	template.SlotDuration = 120 // 2 hour slots to keep test manageable

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

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
	assert.NotEmpty(t, result.Slots, "Should have slots for all-day availability")

	// Verify we have slots throughout the day (not just business hours)
	hasEarlyMorning := false // Before 9am
	hasAfternoon := false    // After 5pm

	for _, slot := range result.Slots {
		slotTime, _ := time.Parse(time.RFC3339, slot.StartTime)
		hour := slotTime.Hour()

		if hour < 9 {
			hasEarlyMorning = true
		}
		if hour >= 17 {
			hasAfternoon = true
		}
	}

	assert.True(t, hasEarlyMorning, "Should have slots before 9am (all-day)")
	assert.True(t, hasAfternoon, "Should have slots after 5pm (all-day)")

	// Verify config reflects default all-day availability
	assert.Equal(t, 1, len(result.Config.WeeklyAvailability.Monday))
	assert.Equal(t, "00:00", result.Config.WeeklyAvailability.Monday[0].Start)
	assert.Equal(t, "23:59", result.Config.WeeklyAvailability.Monday[0].End)
}

// Test fallback behavior: Calendar not found, use default
func TestCalculateFreeSlots_CalendarNotFoundUsesDefault(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)

	// Setup calendars table but don't insert the calendar
	err := db.Exec(`
		CREATE TABLE calendars (
			id INTEGER PRIMARY KEY,
			tenant_id INTEGER,
			weekly_availability JSON,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Template with EMPTY availability
	template := createTestTemplate()
	template.WeeklyAvailability = entities.WeeklyAvailability{} // Empty
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC) // Monday
	endDate := time.Date(2026, 2, 2, 23, 59, 59, 0, time.UTC)

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 999, // Non-existent calendar
		StartDate:  startDate,
		EndDate:    endDate,
		Timezone:   "UTC",
	}

	result, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Slots, "Should have slots with default all-day availability")

	// Verify config reflects default all-day availability
	assert.Equal(t, "00:00", result.Config.WeeklyAvailability.Monday[0].Start)
	assert.Equal(t, "23:59", result.Config.WeeklyAvailability.Monday[0].End)
}

// Test fallback behavior: Partial template availability
func TestCalculateFreeSlots_PartialTemplateAvailability(t *testing.T) {
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

	// Calendar has availability for all weekdays
	calendarAvailability := entities.WeeklyAvailability{
		Monday:  []entities.TimeRange{{Start: "08:00", End: "18:00"}},
		Tuesday: []entities.TimeRange{{Start: "08:00", End: "18:00"}},
	}
	availJSON, _ := calendarAvailability.Value()
	err = db.Exec(`
		INSERT INTO calendars (id, tenant_id, weekly_availability)
		VALUES (?, ?, ?)
	`, 1, 1, availJSON).Error
	require.NoError(t, err)

	// Template has availability only for Monday (partial)
	template := createTestTemplate()
	template.WeeklyAvailability = entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "10:00", End: "12:00"}},
		// Tuesday is empty, but since template has SOME availability, it should use template
	}
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	startDate := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)  // Monday
	endDate := time.Date(2026, 2, 3, 23, 59, 59, 0, time.UTC) // Tuesday

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

	// Since template has SOME availability (Monday), it should use template for all days
	// This means Tuesday should have NO slots (template has no Tuesday availability)
	mondaySlots := 0
	tuesdaySlots := 0

	for _, slot := range result.Slots {
		slotTime, _ := time.Parse(time.RFC3339, slot.StartTime)
		if slotTime.Weekday() == time.Monday {
			mondaySlots++
		} else if slotTime.Weekday() == time.Tuesday {
			tuesdaySlots++
		}
	}

	assert.Greater(t, mondaySlots, 0, "Should have Monday slots from template")
	assert.Equal(t, 0, tuesdaySlots, "Should have NO Tuesday slots (template takes precedence)")
}

func TestGenerateAllSlots_AllowedStartMinutes_BasicAlignment(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()

	// Single window to make expectations clear: 09:00-10:00
	template.WeeklyAvailability = entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "09:00", End: "10:00"}},
	}
	template.SlotDuration = 20 // minutes
	template.BufferTime = 0
	template.AllowedStartMinutes = entities.MinutesArray{0, 30}
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	startDate := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour) // 7 days from now
	// Adjust to next Monday
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, 1)
	}
	endDate := startDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	req := FreeSlotsRequest{StartDate: startDate, EndDate: endDate, Timezone: "UTC"}

	slots := service.generateAllSlots(req, template, time.UTC, template.WeeklyAvailability)

	// Expect starts at 09:00 and 09:30 only (20-min slots)
	var times []string
	for _, s := range slots {
		times = append(times, s.Time)
		st, _ := time.Parse(time.RFC3339, s.StartTime)
		minute := st.Minute()
		assert.Contains(t, []int{0, 30}, minute, "minute must be allowed")
	}
	// Ensure exact expected set
	assert.ElementsMatch(t, times, []string{"09:00", "09:30"})
}

func TestGenerateAllSlots_AllowedStartMinutes_NonStandard(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()

	// Window 09:00-11:00, 30-min duration, allowed minutes at :10 and :40
	template.WeeklyAvailability = entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "09:00", End: "11:00"}},
	}
	template.SlotDuration = 30
	template.BufferTime = 0
	template.AllowedStartMinutes = entities.MinutesArray{10, 40}
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	startDate := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour) // 7 days from now
	// Adjust to next Monday
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, 1)
	}
	endDate := startDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	req := FreeSlotsRequest{StartDate: startDate, EndDate: endDate, Timezone: "UTC"}

	slots := service.generateAllSlots(req, template, time.UTC, template.WeeklyAvailability)

	// Expect: 09:10, 09:40, 10:10 (10:40 would end at 11:10, beyond window)
	var times []string
	for _, s := range slots {
		times = append(times, s.Time)
		st, _ := time.Parse(time.RFC3339, s.StartTime)
		minute := st.Minute()
		assert.Contains(t, []int{10, 40}, minute, "minute must be allowed")
	}
	assert.ElementsMatch(t, times, []string{"09:10", "09:40", "10:10"})
}

func TestGenerateAllSlots_AllowedStartMinutes_WithBuffer(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()

	// Window 09:00-12:00, 45-min duration, 15-min buffer, allowed minutes :00 and :30
	template.WeeklyAvailability = entities.WeeklyAvailability{
		Monday: []entities.TimeRange{{Start: "09:00", End: "12:00"}},
	}
	template.SlotDuration = 45
	template.BufferTime = 15
	template.AllowedStartMinutes = entities.MinutesArray{0, 30}
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 365

	startDate := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour) // 7 days from now
	// Adjust to next Monday
	for startDate.Weekday() != time.Monday {
		startDate = startDate.AddDate(0, 0, 1)
	}
	endDate := startDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	req := FreeSlotsRequest{StartDate: startDate, EndDate: endDate, Timezone: "UTC"}

	slots := service.generateAllSlots(req, template, time.UTC, template.WeeklyAvailability)

	// Expect starts at 09:00, 10:00, 11:00
	expected := []string{"09:00", "10:00", "11:00"}
	var times []string
	for _, s := range slots {
		times = append(times, s.Time)
		st, _ := time.Parse(time.RFC3339, s.StartTime)
		minute := st.Minute()
		assert.Contains(t, []int{0, 30}, minute, "minute must be allowed")
	}
	assert.ElementsMatch(t, times, expected)
}
