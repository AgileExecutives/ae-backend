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

func TestCalendarService_CreateCalendarEntry(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		request       entities.CreateCalendarEntryRequest
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful entry creation",
			request:  fixtures.CreateCalendarEntryRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				// Mock calendar verification
				mockCalendar := fixtures.CreateMockCalendar()
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(nil, true, mockCalendar)
				
				// Mock entry creation
				mockDB.On("Create", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil, true)
			},
			expectedError: "",
		},
		{
			name:     "calendar not found",
			request:  fixtures.CreateCalendarEntryRequest(),
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
			request:  fixtures.CreateCalendarEntryRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				// Mock calendar verification success
				mockCalendar := fixtures.CreateMockCalendar()
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(nil, true, mockCalendar)
				
				// Mock entry creation failure
				mockDB.On("Create", mock.AnythingOfType("*entities.CalendarEntry")).Return(mocks.ErrDatabase, false)
			},
			expectedError: "failed to create calendar entry",
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
			result, err := service.CreateCalendarEntry(tt.request, tt.tenantID, tt.userID)

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

func TestCalendarService_GetCalendarEntryByID(t *testing.T) {
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
			name:     "successful entry retrieval",
			id:       1,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockEntry := fixtures.CreateMockCalendarEntry()
				mockDB.On("Preload", "Calendar").Return(mockDB)
				mockDB.On("Preload", "Series").Return(mockDB)
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil, true, mockEntry)
			},
			expectedError: "",
		},
		{
			name:     "entry not found",
			id:       999,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Preload", "Calendar").Return(mockDB)
				mockDB.On("Preload", "Series").Return(mockDB)
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(999), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarEntry")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar entry with ID 999 not found",
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
			result, err := service.GetCalendarEntryByID(tt.id, tt.tenantID, tt.userID)

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

func TestCalendarService_UpdateCalendarEntry(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		id            uint
		request       entities.UpdateCalendarEntryRequest
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful entry update",
			id:       1,
			request:  fixtures.CreateUpdateCalendarEntryRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockEntry := fixtures.CreateMockCalendarEntry()
				
				// Mock finding the entry
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil, true, mockEntry)
				
				// Mock saving the entry
				mockDB.On("Save", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:     "entry not found",
			id:       999,
			request:  fixtures.CreateUpdateCalendarEntryRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(999), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarEntry")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar entry with ID 999 not found",
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
			result, err := service.UpdateCalendarEntry(tt.id, tt.tenantID, tt.userID, tt.request)

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

func TestCalendarService_DeleteCalendarEntry(t *testing.T) {
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
			name:     "successful entry deletion",
			id:       1,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockEntry := fixtures.CreateMockCalendarEntry()
				
				// Mock finding the entry
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(1), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil, true, mockEntry)
				
				// Mock deleting the entry
				mockDB.On("Delete", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil)
			},
			expectedError: "",
		},
		{
			name:     "entry not found",
			id:       999,
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Where", "id = ? AND tenant_id = ? AND user_id = ?", uint(999), fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.CalendarEntry")).Return(gorm.ErrRecordNotFound, false, nil)
			},
			expectedError: "calendar entry with ID 999 not found",
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
			err := service.DeleteCalendarEntry(tt.id, tt.tenantID, tt.userID)

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

func TestCalendarService_GetFreeSlots(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name            string
		request         entities.FreeSlotRequest
		tenantID        uint
		userID          uint
		expectedMinSlots int
		expectedError   string
	}{
		{
			name:             "successful free slot calculation",
			request:          fixtures.CreateFreeSlotRequest(),
			tenantID:         fixtures.TenantID,
			userID:           fixtures.UserID,
			expectedMinSlots: 1, // At least one slot should be returned
			expectedError:    "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := service.GetFreeSlots(tt.request, tt.tenantID, tt.userID)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(result), tt.expectedMinSlots)
				
				// Verify slot structure
				for _, slot := range result {
					assert.True(t, slot.EndTime.After(slot.StartTime))
					assert.Equal(t, tt.request.Duration, slot.Duration)
				}
			}
		})
	}
}

func TestCalendarService_ImportHolidays(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name          string
		request       entities.ImportHolidaysRequest
		tenantID      uint
		userID        uint
		setupMock     func()
		expectedError string
	}{
		{
			name:     "successful holiday import with new calendar",
			request:  fixtures.CreateImportHolidaysRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				// Mock holidays calendar not found (will create new one)
				mockDB.On("Where", "tenant_id = ? AND user_id = ? AND title = ?", fixtures.TenantID, fixtures.UserID, "Holidays").Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(gorm.ErrRecordNotFound, false, nil)
				
				// Mock calendar creation
				mockDB.On("Create", mock.AnythingOfType("*entities.Calendar")).Return(nil, true)
				
				// Mock holiday entry creation
				mockDB.On("Create", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil, true)
			},
			expectedError: "",
		},
		{
			name:     "successful holiday import with existing calendar",
			request:  fixtures.CreateImportHolidaysRequest(),
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				// Mock holidays calendar found
				mockCalendar := fixtures.CreateMockCalendar()
				mockCalendar.Title = "Holidays"
				mockDB.On("Where", "tenant_id = ? AND user_id = ? AND title = ?", fixtures.TenantID, fixtures.UserID, "Holidays").Return(mockDB)
				mockDB.On("First", mock.AnythingOfType("*entities.Calendar")).Return(nil, true, mockCalendar)
				
				// Mock holiday entry creation
				mockDB.On("Create", mock.AnythingOfType("*entities.CalendarEntry")).Return(nil, true)
			},
			expectedError: "",
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
			err := service.ImportHolidays(tt.request, tt.tenantID, tt.userID)

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