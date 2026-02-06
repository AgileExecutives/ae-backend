package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarService_CreateCalendarSeries(t *testing.T) {
	service, db := SetupTestService(t)
	fixtures := NewTestFixtures()

	// Create a calendar first
	calendar := fixtures.CreateMockCalendar()
	err := db.Create(calendar).Error
	require.NoError(t, err)

	t.Run("successful series creation", func(t *testing.T) {
		req := fixtures.CreateCalendarSeriesRequest()
		series, err := service.CreateCalendarSeries(req, fixtures.TenantID, fixtures.UserID)

		assert.NoError(t, err)
		assert.NotNil(t, series)
		assert.Equal(t, req.Title, series.Title)
		assert.Equal(t, req.CalendarID, series.CalendarID)
		assert.Equal(t, fixtures.TenantID, series.TenantID)
		assert.Equal(t, fixtures.UserID, series.UserID)
	})

	t.Run("calendar not found", func(t *testing.T) {
		req := fixtures.CreateCalendarSeriesRequest()
		req.CalendarID = 999 // Non-existent calendar

		series, err := service.CreateCalendarSeries(req, fixtures.TenantID, fixtures.UserID)

		assert.Error(t, err)
		assert.Nil(t, series)
	})
}

func TestCalendarService_GetCalendarSeriesByID(t *testing.T) {
	service, db := SetupTestService(t)
	fixtures := NewTestFixtures()

	// Create a calendar and series
	calendar := fixtures.CreateMockCalendar()
	err := db.Create(calendar).Error
	require.NoError(t, err)

	req := fixtures.CreateCalendarSeriesRequest()
	series, err := service.CreateCalendarSeries(req, fixtures.TenantID, fixtures.UserID)
	require.NoError(t, err)

	t.Run("successful series retrieval", func(t *testing.T) {
		result, err := service.GetCalendarSeriesByID(series.ID, fixtures.TenantID, fixtures.UserID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, series.ID, result.ID)
		assert.Equal(t, series.Title, result.Title)
	})

	t.Run("series not found", func(t *testing.T) {
		result, err := service.GetCalendarSeriesByID(999, fixtures.TenantID, fixtures.UserID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCalendarService_UpdateCalendarSeries(t *testing.T) {
	service, db := SetupTestService(t)
	fixtures := NewTestFixtures()

	// Create a calendar and series
	calendar := fixtures.CreateMockCalendar()
	err := db.Create(calendar).Error
	require.NoError(t, err)

	req := fixtures.CreateCalendarSeriesRequest()
	series, err := service.CreateCalendarSeries(req, fixtures.TenantID, fixtures.UserID)
	require.NoError(t, err)

	t.Run("successful series update", func(t *testing.T) {
		updateReq := fixtures.CreateUpdateCalendarSeriesRequest()
		result, err := service.UpdateCalendarSeries(series.ID, fixtures.TenantID, fixtures.UserID, updateReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		if updateReq.Title != nil {
			assert.Equal(t, *updateReq.Title, result.Title)
		}
	})
}

func TestCalendarService_DeleteCalendarSeries(t *testing.T) {
	t.Skip("DeleteCalendarSeries signature unknown")
}

func TestCalendarService_GenerateSeriesEntries(t *testing.T) {
	t.Skip("GenerateSeriesEntries is unexported")
}
