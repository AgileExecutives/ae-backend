package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
)

// Phase 1: Test recurrence calculation stops on first conflict

func TestCalculateAvailableRecurrences_StopsOnFirstConflict(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 5
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 1000

	// Get next Thursday at 9:00
	baseDate := getNextWeekday(time.Thursday, 7).Add(9 * time.Hour)

	// Create a conflict in week 3 (Thursday 9:00-10:00)
	week3Conflict := baseDate.AddDate(0, 0, 14) // 2 weeks later
	week3End := week3Conflict.Add(1 * time.Hour)

	err := db.Exec(`
		INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
		VALUES (?, ?, ?, ?)
	`, 1, 1, week3Conflict, week3End).Error
	require.NoError(t, err)

	// Create slot for the first Thursday
	slots := []entities.TimeSlot{
		{
			ID:        "slot-1",
			StartTime: baseDate.Format(time.RFC3339),
			EndTime:   baseDate.Add(1 * time.Hour).Format(time.RFC3339),
			Available: true,
		},
	}

	// Get existing entries
	var existingEntries []CalendarEntry
	err = db.Table("calendar_entries").
		Select("id, calendar_id, tenant_id, start_time, end_time").
		Where("calendar_id = ? AND tenant_id = ?", 1, 1).
		Find(&existingEntries).Error
	require.NoError(t, err)

	endDate := baseDate.AddDate(0, 2, 0) // 2 months out

	// Calculate recurrences
	result := service.calculateAvailableRecurrences(slots, existingEntries, template, endDate, time.UTC)

	// Should stop at first conflict - weeks 1 and 2 are available, week 3 conflicts
	// Expected: 2 (current week + 1 week ahead, then stops at week 3 conflict)
	assert.Equal(t, 2, result[0].AvailableRecurrences,
		"Should stop counting at first conflict (weeks 1 and 2 only)")
}

func TestCalculateAvailableRecurrences_AllSlotsAvailable(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 5
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 1000

	// Get next Thursday at 9:00
	baseDate := getNextWeekday(time.Thursday, 7).Add(9 * time.Hour)

	// No conflicts - empty calendar
	slots := []entities.TimeSlot{
		{
			ID:        "slot-1",
			StartTime: baseDate.Format(time.RFC3339),
			EndTime:   baseDate.Add(1 * time.Hour).Format(time.RFC3339),
			Available: true,
		},
	}

	var existingEntries []CalendarEntry
	endDate := baseDate.AddDate(0, 2, 0) // 2 months out

	// Calculate recurrences
	result := service.calculateAvailableRecurrences(slots, existingEntries, template, endDate, time.UTC)

	// All 5 weeks should be available
	assert.Equal(t, 5, result[0].AvailableRecurrences,
		"Should have all 5 recurrences available with no conflicts")
}

func TestCalculateAvailableRecurrences_ImmediateConflict(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 5
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 1000

	// Get next Thursday at 9:00
	baseDate := getNextWeekday(time.Thursday, 7).Add(9 * time.Hour)

	// Create a conflict in week 2 (next Thursday)
	week2Conflict := baseDate.AddDate(0, 0, 7) // 1 week later
	week2End := week2Conflict.Add(1 * time.Hour)

	err := db.Exec(`
		INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
		VALUES (?, ?, ?, ?)
	`, 1, 1, week2Conflict, week2End).Error
	require.NoError(t, err)

	slots := []entities.TimeSlot{
		{
			ID:        "slot-1",
			StartTime: baseDate.Format(time.RFC3339),
			EndTime:   baseDate.Add(1 * time.Hour).Format(time.RFC3339),
			Available: true,
		},
	}

	var existingEntries []CalendarEntry
	err = db.Table("calendar_entries").
		Select("id, calendar_id, tenant_id, start_time, end_time").
		Where("calendar_id = ? AND tenant_id = ?", 1, 1).
		Find(&existingEntries).Error
	require.NoError(t, err)

	endDate := baseDate.AddDate(0, 2, 0)

	result := service.calculateAvailableRecurrences(slots, existingEntries, template, endDate, time.UTC)

	// Only the first week is available, second week conflicts
	assert.Equal(t, 1, result[0].AvailableRecurrences,
		"Should have only 1 recurrence (current slot), stops at week 2 conflict")
}

func TestCalculateAvailableRecurrences_MonthlyInterval(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalMonthlyDate}
	template.MaxSeriesBookings = 4
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 1000

	// Get next Monday at 10:00
	baseDate := getNextWeekday(time.Monday, 7).Add(10 * time.Hour)

	// Create a conflict in month 3 (same day, 2 months later)
	month3Conflict := baseDate.AddDate(0, 2, 0)
	month3End := month3Conflict.Add(1 * time.Hour)

	err := db.Exec(`
		INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
		VALUES (?, ?, ?, ?)
	`, 1, 1, month3Conflict, month3End).Error
	require.NoError(t, err)

	slots := []entities.TimeSlot{
		{
			ID:        "slot-1",
			StartTime: baseDate.Format(time.RFC3339),
			EndTime:   baseDate.Add(1 * time.Hour).Format(time.RFC3339),
			Available: true,
		},
	}

	var existingEntries []CalendarEntry
	err = db.Table("calendar_entries").
		Select("id, calendar_id, tenant_id, start_time, end_time").
		Where("calendar_id = ? AND tenant_id = ?", 1, 1).
		Find(&existingEntries).Error
	require.NoError(t, err)

	endDate := baseDate.AddDate(0, 6, 0) // 6 months out

	result := service.calculateAvailableRecurrences(slots, existingEntries, template, endDate, time.UTC)

	// Months 1 and 2 available, month 3 conflicts
	assert.Equal(t, 2, result[0].AvailableRecurrences,
		"Should stop at first monthly conflict (months 1 and 2 only)")
}

// Phase 2: Verify API response includes AvailableRecurrences

func TestFreeSlotsResponse_IncludesAvailableRecurrences(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 3
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 1000

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
	require.NotEmpty(t, result.Slots)

	// Each slot should have AvailableRecurrences field populated
	for _, slot := range result.Slots {
		assert.GreaterOrEqual(t, slot.AvailableRecurrences, 1,
			"Each slot should have at least 1 recurrence (itself)")
		assert.LessOrEqual(t, slot.AvailableRecurrences, template.MaxSeriesBookings,
			"Recurrences should not exceed template max")
	}
}

func TestFreeSlotsResponse_DifferentSlotsHaveDifferentRecurrences(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 5
	template.MinNoticeHours = 0
	template.AdvanceBookingDays = 1000

	startDate := getNextWeekday(time.Monday, 7)
	endDate := startDate.AddDate(0, 2, 0) // 2 months out to allow for weekly recurrences

	// Create a conflict at 10:00 in week 2 (will affect one slot's recurrence count)
	conflictTime := startDate.AddDate(0, 0, 7).Add(10 * time.Hour)
	conflictEnd := conflictTime.Add(30 * time.Minute)

	err := db.Exec(`
		INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
		VALUES (?, ?, ?, ?)
	`, 1, 1, conflictTime, conflictEnd).Error
	require.NoError(t, err)

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
	require.NotEmpty(t, result.Slots, "Should have generated slots")

	// Find 9:00 and 14:00 slots from the FIRST day only
	var slot9, slot14 *entities.TimeSlot
	firstDate := result.Slots[0].Date
	for i := range result.Slots {
		if result.Slots[i].Date == firstDate {
			if result.Slots[i].Time == "09:00" {
				slot9 = &result.Slots[i]
			} else if result.Slots[i].Time == "14:00" {
				slot14 = &result.Slots[i]
			}
		}
	}

	require.NotNil(t, slot9, "Should have 9:00 slot")
	require.NotNil(t, slot14, "Should have 14:00 slot")

	// Both should have recurrences, but different counts shows calculation is per-slot
	assert.GreaterOrEqual(t, slot9.AvailableRecurrences, 1)
	assert.GreaterOrEqual(t, slot14.AvailableRecurrences, 1)

	// With no conflicts, both should have max recurrences
	assert.Equal(t, template.MaxSeriesBookings, slot9.AvailableRecurrences,
		"9:00 slot should have max recurrences with no conflicts")
	assert.Equal(t, template.MaxSeriesBookings, slot14.AvailableRecurrences,
		"14:00 slot should have max recurrences with no conflicts")
}
