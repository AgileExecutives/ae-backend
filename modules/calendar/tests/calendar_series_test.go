package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/services"
	"github.com/unburdy/calendar-module/tests/mocks"
)

func TestCalendarService_CreateCalendarSeries(t *testing.T) {
	fixtures := NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		request       entities.CreateCalendarSeriesRequest
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful series creation",
			request:  fixtures.CreateCalendarSeriesRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				// Mock calendar verification
				mockCalendar := fixtures.CreateMockCalendar()
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(nil, true, mockCalendar)

				// Mock series creation
				mockDB.On("Create", mock.AnythingOfType("*entities.CalendarSeries")).Return(nil, true)
			},
			expectedError: "",
		},
		{
			name:     "calendar not found",
			request:  fixtures.CreateCalendarSeriesRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar not found or access denied",
		},
		{
			name:     "database error during creation",
			request:  fixtures.CreateCalendarSeriesRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				// Mock calendar verification success
				mockCalendar := fixtures.CreateMockCalendar()
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(nil, true, mockCalendar)

				// Mock series creation failure
				mockDB.On("Create", mock.AnythingOfType("*entities.CalendarSeries")).Return(mocks.ErrDatabase, false)
			},
			expectedError: "failed to create calendar series",
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
			result, err := service.CreateCalendarSeries(tt.request, tt.tenantID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.Title, result.Title)
				assert.Equal(t, tt.request.CalendarID, result.CalendarID)
				assert.Equal(t, tt.tenantID, result.TenantID)
				assert.Equal(t, tt.userID, result.UserID)
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCalendarService_GetCalendarSeriesByID(t *testing.T) {
	fixtures := NewTestFixtures()
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
			name:     "successful series retrieval",
			id:       1,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockSeries := fixtures.CreateMockCalendarSeries()
				mockDB.On("Preload", "Calendar").Return(mockDB)
				mockDB.On("Preload", "Entries").Return(mockDB)
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarSeries")).Return(nil, true, mockSeries)
			},
			expectedError: "",
		},
		{
			name:     "series not found",
			id:       999,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Preload", "Calendar").Return(mockDB)
				mockDB.On("Preload", "Entries").Return(mockDB)
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(999), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarSeries")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar series with ID 999 not found",
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
			result, err := service.GetCalendarSeriesByID(tt.id, tt.tenantID, tt.userID)

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

func TestCalendarService_UpdateCalendarSeries(t *testing.T) {
	fixtures := NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		id            uint
		request       entities.UpdateCalendarSeriesRequest
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful series update",
			id:       1,
			request:  fixtures.CreateUpdateCalendarSeriesRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockSeries := fixtures.CreateMockCalendarSeries()

				// Mock finding the series
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarSeries")).Return(nil, true, mockSeries)

				// Mock saving the series
				mockDB.On("Save", mock.AnythingOfType("*entities.CalendarSeries")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:     "series not found",
			id:       999,
			request:  fixtures.CreateUpdateCalendarSeriesRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(999), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarSeries")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar series with ID 999 not found",
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
			result, err := service.UpdateCalendarSeries(tt.id, tt.tenantID, tt.userID, tt.request)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.request.Title != nil {
					assert.Equal(t, *tt.request.Title, result.Title)
				}
				if tt.request.Description != nil {
					assert.Equal(t, *tt.request.Description, result.Description)
				}
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCalendarService_DeleteCalendarSeries(t *testing.T) {
	fixtures := NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		id            uint
		tenantID      uint
		userID        uint
		req           entities.DeleteCalendarSeriesRequest
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful series deletion - all mode",
			id:       1,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			req: entities.DeleteCalendarSeriesRequest{
				DeleteMode: "all",
			},
			setupMock: func() {
				mockSeries := fixtures.CreateMockCalendarSeries()

				// Mock finding the series
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarSeries")).Return(nil, true, mockSeries)

				// Mock deleting related entries
				mockDB.On("Where", "series_id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("Delete", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil)

				// Mock deleting the series
				mockDB.On("Delete", mock.AnythingOfType("*entities.CalendarSeries")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:     "series not found",
			id:       999,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			req: entities.DeleteCalendarSeriesRequest{
				DeleteMode: "all",
			},
			setupMock: func() {
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(999), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarSeries")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar series with ID 999 not found",
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
			err := service.DeleteCalendarSeries(tt.id, tt.tenantID, tt.userID, tt.req)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCalendarService_GenerateSeriesEntries(t *testing.T) {
	fixtures := NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		seriesID      uint
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful series entry generation",
			seriesID: 1,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				// Mock finding the series
				mockSeries := fixtures.CreateMockCalendarSeries()
				mockSeries.Frequency = "weekly"
				mockSeries.Interval = 1
				mockSeries.Count = 5
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarSeries")).Return(nil, true, mockSeries)

				// Mock deleting existing entries
				mockDB.On("Where", "series_id = ?", uint(1)).Return(mockDB)
				mockDB.On("Delete", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil)

				// Mock creating new entries (will be called multiple times)
				mockDB.On("Create", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil, true).Maybe()
			},
			expectedError: "",
		},
		{
			name:     "series not found",
			seriesID: 999,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(999), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarSeries")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar series with ID 999 not found",
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
			err := service.GenerateSeriesEntries(tt.seriesID, tt.tenantID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}
