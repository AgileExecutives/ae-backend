package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/services"
	"github.com/unburdy/calendar-module/tests"
	"github.com/unburdy/calendar-module/tests/mocks"
)

func TestCalendarService_CreateCalendar(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		request       entities.CreateCalendarRequest
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful calendar creation",
			request:  fixtures.CreateCalendarRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Create", mock.AnythingOfType("*entities.Calendar")).Return(nil, true)
			},
			expectedError: "",
		},
		{
			name:     "database error during creation",
			request:  fixtures.CreateCalendarRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Create", mock.AnythingOfType("*entities.Calendar")).Return(mocks.ErrDatabase, false)
			},
			expectedError: "failed to create calendar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDB.ExpectedCalls = nil
			mockDB.Calls = nil

			// Setup mock expectations
			tt.setupMock()

			// Execute
			result, err := service.CreateCalendar(tt.request, tt.tenantID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Title, result.Title)
				assert.Equal(t, tt.request.Color, result.Color)
				assert.Equal(t, tt.tenantID, result.TenantID)
				assert.Equal(t, tt.userID, result.UserID)
				assert.NotEmpty(t, result.CalendarUUID)
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCalendarService_GetCalendarByID(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		id            uint
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful calendar retrieval",
			id:       1,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockCalendar := fixtures.CreateMockCalendar()
				mockDB.On("Preload", "CalendarSeries").Return(mockDB)
				mockDB.On("Preload", "CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "ExternalCalendars").Return(mockDB)
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(nil, true, mockCalendar)
			},
			expectedError: "",
		},
		{
			name:     "calendar not found",
			id:       999,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Preload", "CalendarSeries").Return(mockDB)
				mockDB.On("Preload", "CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "ExternalCalendars").Return(mockDB)
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(999), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar with ID 999 not found",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDB.ExpectedCalls = nil
			mockDB.Calls = nil

			// Setup mock expectations
			tt.setupMock()

			// Execute
			result, err := service.GetCalendarByID(tt.id, tt.tenantID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
				assert.Equal(t, tt.tenantID, result.TenantID)
				assert.Equal(t, tt.userID, result.UserID)
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCalendarService_GetCalendarWeekView(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testDate := time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC)        // Wednesday
	startOfWeek := testDate.AddDate(0, 0, -int(testDate.Weekday())) // Sunday
	endOfWeek := startOfWeek.AddDate(0, 0, 6)                       // Saturday

	testCases := []struct {
		name          string
		date          time.Time
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedCount int
		expectedError string
	}{
		{
			name:     "successful week view",
			date:     testDate,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockEntries := []entities.CalendarEntry{*fixtures.CreateMockCalendarEntry()}

				mockDB.On("Preload", "Calendar").Return(mockDB)
				mockDB.On("Preload", "Series").Return(mockDB)
				mockDB.On("Where", "tenant_id = ? AND user_id = ? AND date_from <= ? AND date_to >= ?",
					fixtures.TenantID, fixtures.UserID, endOfWeek, startOfWeek).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]entities.CalendarEntry")).Return(nil, true, mockEntries)
			},
			expectedCount: 1,
			expectedError: "",
		},
		{
			name:     "database error",
			date:     testDate,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Preload", "Calendar").Return(mockDB)
				mockDB.On("Preload", "Series").Return(mockDB)
				mockDB.On("Where", "tenant_id = ? AND user_id = ? AND date_from <= ? AND date_to >= ?",
					fixtures.TenantID, fixtures.UserID, endOfWeek, startOfWeek).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]entities.CalendarEntry")).Return(mocks.ErrDatabase, false, []entities.CalendarEntry{})
			},
			expectedCount: 0,
			expectedError: "failed to fetch week view",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDB.ExpectedCalls = nil
			mockDB.Calls = nil

			// Setup mock expectations
			tt.setupMock()

			// Execute
			result, err := service.GetCalendarWeekView(tt.date, tt.tenantID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, 0, len(result))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(result))
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}
