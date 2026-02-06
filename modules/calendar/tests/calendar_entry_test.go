package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarService_CreateCalendarEntry(t *testing.T) {
	service, db := SetupTestService(t)
	fixtures := NewTestFixtures()

	// Create a calendar first
	calendar := fixtures.CreateMockCalendar()
	err := db.Create(calendar).Error
	require.NoError(t, err)

	t.Run("successful entry creation", func(t *testing.T) {
		req := fixtures.CreateCalendarEntryRequest()
		entry, err := service.CreateCalendarEntry(req, fixtures.TenantID, fixtures.UserID)

		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, req.Title, entry.Title)
		assert.Equal(t, req.CalendarID, entry.CalendarID)
		assert.Equal(t, fixtures.TenantID, entry.TenantID)
		assert.Equal(t, fixtures.UserID, entry.UserID)
	})

	t.Run("calendar not found", func(t *testing.T) {
		req := fixtures.CreateCalendarEntryRequest()
		req.CalendarID = 999 // Non-existent calendar

		entry, err := service.CreateCalendarEntry(req, fixtures.TenantID, fixtures.UserID)

		assert.Error(t, err)
		assert.Nil(t, entry)
	})
}

func TestCalendarService_GetCalendarEntryByID(t *testing.T) {
	service, db := SetupTestService(t)
	fixtures := NewTestFixtures()

	// Create a calendar and entry
	calendar := fixtures.CreateMockCalendar()
	err := db.Create(calendar).Error
	require.NoError(t, err)

	req := fixtures.CreateCalendarEntryRequest()
	entry, err := service.CreateCalendarEntry(req, fixtures.TenantID, fixtures.UserID)
	require.NoError(t, err)

	t.Run("successful entry retrieval", func(t *testing.T) {
		result, err := service.GetCalendarEntryByID(entry.ID, fixtures.TenantID, fixtures.UserID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entry.ID, result.ID)
		assert.Equal(t, entry.Title, result.Title)
	})

	t.Run("entry not found", func(t *testing.T) {
		result, err := service.GetCalendarEntryByID(999, fixtures.TenantID, fixtures.UserID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCalendarService_UpdateCalendarEntry(t *testing.T) {
	service, db := SetupTestService(t)
	fixtures := NewTestFixtures()

	// Create a calendar and entry
	calendar := fixtures.CreateMockCalendar()
	err := db.Create(calendar).Error
	require.NoError(t, err)

	req := fixtures.CreateCalendarEntryRequest()
	entry, err := service.CreateCalendarEntry(req, fixtures.TenantID, fixtures.UserID)
	require.NoError(t, err)

	t.Run("successful entry update", func(t *testing.T) {
		updateReq := fixtures.CreateUpdateCalendarEntryRequest()
		result, err := service.UpdateCalendarEntry(entry.ID, fixtures.TenantID, fixtures.UserID, updateReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if updateReq.Title != nil {
			assert.Equal(t, *updateReq.Title, result.Title)
		}
	})
}

func TestCalendarService_DeleteCalendarEntry(t *testing.T) {
	service, db := SetupTestService(t)
	fixtures := NewTestFixtures()

	// Create a calendar and entry
	calendar := fixtures.CreateMockCalendar()
	err := db.Create(calendar).Error
	require.NoError(t, err)

	req := fixtures.CreateCalendarEntryRequest()
	entry, err := service.CreateCalendarEntry(req, fixtures.TenantID, fixtures.UserID)
	require.NoError(t, err)

	t.Run("successful entry deletion", func(t *testing.T) {
		err := service.DeleteCalendarEntry(entry.ID, fixtures.TenantID, fixtures.UserID)
		assert.NoError(t, err)

		result, err := service.GetCalendarEntryByID(entry.ID, fixtures.TenantID, fixtures.UserID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCalendarService_ImportHolidays(t *testing.T) {
	t.Skip("HolidayImportRequest not implemented")
}
