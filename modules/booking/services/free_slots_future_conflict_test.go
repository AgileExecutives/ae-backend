package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
)

// TestFreeSlotsWithFutureConflict tests that recurrence calculation detects conflicts
// outside the search window (e.g., a booking in May when searching April slots)
func TestFreeSlotsWithFutureConflict(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 10

	// April 1, 2026 is a Wednesday
	searchStart := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	searchEnd := time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)

	// Create a conflict on Wednesday May 13 at 10:00 (outside search window)
	// This should affect the recurrence count for Wednesday April slots
	conflictDate := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	conflictEnd := conflictDate.Add(55 * time.Minute)

	err := db.Exec(`
		INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
		VALUES (?, ?, ?, ?)
	`, 1, 1, conflictDate, conflictEnd).Error
	require.NoError(t, err)

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  searchStart,
		EndDate:    searchEnd,
		Timezone:   "UTC",
	}

	response, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.Slots, "Should have generated slots")

	// Find Wednesday April 1 at 10:00
	var targetSlot *entities.TimeSlot
	for i := range response.Slots {
		slot := &response.Slots[i]
		if slot.Date == "2026-04-01" && slot.Time == "10:00" {
			targetSlot = slot
			break
		}
	}

	require.NotNil(t, targetSlot, "Should find April 1 Wednesday 10:00 slot")

	// Count weeks from April 1 to May 13:
	// Apr 1, Apr 8, Apr 15, Apr 22, Apr 29, May 6, May 13 (conflict)
	// So we should be able to book 6 recurrences before hitting the conflict
	expectedRecurrences := 6

	assert.Equal(t, expectedRecurrences, targetSlot.AvailableRecurrences,
		"Should count 6 weekly recurrences (Apr 1, 8, 15, 22, 29, May 6) before conflict on May 13")
}

// TestFreeSlotsWithMultipleFutureConflicts tests multiple conflicts outside search window
func TestFreeSlotsWithMultipleFutureConflicts(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 10

	// Search April 2026
	searchStart := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	searchEnd := time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)

	// Create conflicts on Wednesday:
	// - May 6 at 14:00 (week 5 from April 1)
	// - May 20 at 14:00 (week 7 from April 1)
	conflicts := []time.Time{
		time.Date(2026, 5, 6, 14, 0, 0, 0, time.UTC),  // Wednesday May 6
		time.Date(2026, 5, 20, 14, 0, 0, 0, time.UTC), // Wednesday May 20
	}

	for _, conflictDate := range conflicts {
		conflictEnd := conflictDate.Add(55 * time.Minute)
		err := db.Exec(`
			INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
			VALUES (?, ?, ?, ?)
		`, 1, 1, conflictDate, conflictEnd).Error
		require.NoError(t, err)
	}

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  searchStart,
		EndDate:    searchEnd,
		Timezone:   "UTC",
	}

	response, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.Slots, "Should have generated slots")

	// Find first Tuesday at 14:00 in April
	var targetSlot *entities.TimeSlot
	for i := range response.Slots {
		slot := &response.Slots[i]
		// April 1, 2026 is Wednesday, so first Tuesday is April 7
		if slot.Date == "2026-04-01" && slot.Time == "14:00" {
			targetSlot = slot
			break
		}
	}

	require.NotNil(t, targetSlot, "Should find April 1 Wednesday 14:00 slot")

	// Wednesdays from April 1: Apr 1, 8, 15, 22, 29, May 6 (conflict)
	// Algorithm stops at first conflict
	expectedRecurrences := 5

	assert.Equal(t, expectedRecurrences, targetSlot.AvailableRecurrences,
		"Should count 5 weekly recurrences before hitting first conflict on May 6")
}

// TestFreeSlotsWithPastBookingNoEffect tests that past bookings don't affect recurrence count
func TestFreeSlotsWithPastBookingNoEffect(t *testing.T) {
	db := setupFreeSlotsTestDB(t)
	service := NewFreeSlotsService(db)
	template := createTestTemplate()
	template.AllowedIntervals = entities.IntervalArray{entities.IntervalWeekly}
	template.MaxSeriesBookings = 10

	// Search April 2026
	searchStart := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	searchEnd := time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)

	// Create a booking in March (past, but same weekday/time)
	pastBooking := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC) // Wednesday March 18
	pastEnd := pastBooking.Add(55 * time.Minute)

	err := db.Exec(`
		INSERT INTO calendar_entries (calendar_id, tenant_id, start_time, end_time)
		VALUES (?, ?, ?, ?)
	`, 1, 1, pastBooking, pastEnd).Error
	require.NoError(t, err)

	req := FreeSlotsRequest{
		TemplateID: 1,
		TenantID:   1,
		CalendarID: 1,
		StartDate:  searchStart,
		EndDate:    searchEnd,
		Timezone:   "UTC",
	}

	response, err := service.CalculateFreeSlots(req, template)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.Slots, "Should have generated slots")

	// Find Wednesday April 1 at 10:00
	var targetSlot *entities.TimeSlot
	for i := range response.Slots {
		slot := &response.Slots[i]
		if slot.Date == "2026-04-01" && slot.Time == "10:00" {
			targetSlot = slot
			break
		}
	}

	require.NotNil(t, targetSlot, "Should find April 1 Wednesday 10:00 slot")

	// Past bookings should NOT affect recurrence count - should get full 10
	assert.Equal(t, 10, targetSlot.AvailableRecurrences,
		"Past bookings should not reduce available recurrences")
}
