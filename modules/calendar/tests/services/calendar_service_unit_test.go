package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unburdy/calendar-module/tests"
)

func TestCalendarService_CreateCalendar(t *testing.T) {
	service, _ := tests.SetupTestService(t)
	fixtures := tests.NewTestFixtures()

	t.Run("successful calendar creation", func(t *testing.T) {
		request := fixtures.CreateCalendarRequest()

		result, err := service.CreateCalendar(request, fixtures.TenantID, fixtures.UserID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.Title, result.Title)
		assert.Equal(t, request.Color, result.Color)
		assert.Equal(t, fixtures.TenantID, result.TenantID)
		assert.Equal(t, fixtures.UserID, result.UserID)
		assert.NotEmpty(t, result.CalendarUUID)
	})
}
