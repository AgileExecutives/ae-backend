package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/services"
	"github.com/unburdy/calendar-module/tests"
	"github.com/unburdy/calendar-module/tests/mocks"
)

func TestCalendarService_GetCalendarsWithDeepPreload(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	testCases := []struct {
		name                  string
		tenantID              uint
		userID                uint
		setupMock             func()
		expectedCalendarCount int
		expectedError         string
		validateDeepPreload   bool
	}{
		{
			name:     "successful calendars retrieval with deep preload",
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				// Create mock calendars with nested relationships
				mockCalendar1 := fixtures.CreateMockCalendar()
				mockCalendar1.ID = 1

				mockCalendar2 := fixtures.CreateMockCalendar()
				mockCalendar2.ID = 2
				mockCalendar2.Title = "Work Calendar"

				// Create mock calendar entries with series
				mockEntry1 := fixtures.CreateMockCalendarEntry()
				mockEntry1.ID = 1
				mockEntry1.CalendarID = 1
				mockEntry1.SeriesID = &[]uint{1}[0] // Reference to series ID 1

				mockEntry2 := fixtures.CreateMockCalendarEntry()
				mockEntry2.ID = 2
				mockEntry2.CalendarID = 1
				mockEntry2.SeriesID = nil // Standalone entry

				// Create mock calendar series with entries
				mockSeries1 := fixtures.CreateMockCalendarSeries()
				mockSeries1.ID = 1
				mockSeries1.CalendarID = 1

				// Simulate 2-level preloading structure
				mockCalendar1.CalendarEntries = []entities.CalendarEntry{*mockEntry1, *mockEntry2}
				mockCalendar1.CalendarSeries = []entities.CalendarSeries{*mockSeries1}

				// Simulate nested preloads (entries->series, series->entries)
				mockCalendar1.CalendarEntries[0].Series = mockSeries1                                   // Entry points to series
				mockCalendar1.CalendarSeries[0].CalendarEntries = []entities.CalendarEntry{*mockEntry1} // Series points to entries

				mockCalendars := []entities.Calendar{*mockCalendar1, *mockCalendar2}

				// Setup mock expectations for deep preloading (exact order matters)
				mockDB.On("Preload", "CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "CalendarEntries.Series").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries.CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "ExternalCalendars").Return(mockDB)
				mockDB.On("Where", "tenant_id = ? AND user_id = ?", fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]entities.Calendar")).Return(nil, true, mockCalendars)
			},
			expectedCalendarCount: 2,
			expectedError:         "",
			validateDeepPreload:   true,
		},
		{
			name:     "empty calendars result",
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockCalendars := []entities.Calendar{}

				mockDB.On("Preload", "CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "CalendarEntries.Series").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries.CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "ExternalCalendars").Return(mockDB)
				mockDB.On("Where", "tenant_id = ? AND user_id = ?", fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]entities.Calendar")).Return(nil, true, mockCalendars)
			},
			expectedCalendarCount: 0,
			expectedError:         "",
			validateDeepPreload:   false,
		},
		{
			name:     "database error during retrieval",
			tenantID: fixtures.TenantID,
			userID:   fixtures.UserID,
			setupMock: func() {
				mockDB.On("Preload", "CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "CalendarEntries.Series").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries.CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "ExternalCalendars").Return(mockDB)
				mockDB.On("Where", "tenant_id = ? AND user_id = ?", fixtures.TenantID, fixtures.UserID).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]entities.Calendar")).Return(mocks.ErrDatabase, false, []entities.Calendar{})
			},
			expectedCalendarCount: 0,
			expectedError:         "failed to fetch calendars with deep preload",
			validateDeepPreload:   false,
		},
		{
			name:     "tenant isolation - no cross-tenant data",
			tenantID: 999, // Different tenant
			userID:   fixtures.UserID,
			setupMock: func() {
				mockCalendars := []entities.Calendar{} // Empty because different tenant

				mockDB.On("Preload", "CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "CalendarEntries.Series").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries.CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "ExternalCalendars").Return(mockDB)
				mockDB.On("Where", "tenant_id = ? AND user_id = ?", uint(999), fixtures.UserID).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]entities.Calendar")).Return(nil, true, mockCalendars)
			},
			expectedCalendarCount: 0,
			expectedError:         "",
			validateDeepPreload:   false,
		},
		{
			name:     "user isolation - no cross-user data",
			tenantID: fixtures.TenantID,
			userID:   999, // Different user
			setupMock: func() {
				mockCalendars := []entities.Calendar{} // Empty because different user

				mockDB.On("Preload", "CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "CalendarEntries.Series").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries").Return(mockDB)
				mockDB.On("Preload", "CalendarSeries.CalendarEntries").Return(mockDB)
				mockDB.On("Preload", "ExternalCalendars").Return(mockDB)
				mockDB.On("Where", "tenant_id = ? AND user_id = ?", fixtures.TenantID, uint(999)).Return(mockDB)
				mockDB.On("Find", mock.AnythingOfType("*[]entities.Calendar")).Return(nil, true, mockCalendars)
			},
			expectedCalendarCount: 0,
			expectedError:         "",
			validateDeepPreload:   false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDB.ExpectedCalls = nil
			mockDB.Calls = nil

			// Setup mock expectations
			tt.setupMock()

			// Execute the method being tested
			result, err := service.GetCalendarsWithDeepPreload(tt.tenantID, tt.userID)

			// Assert basic results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, 0, len(result))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCalendarCount, len(result))

				// Validate tenant/user isolation
				for _, calendar := range result {
					assert.Equal(t, tt.tenantID, calendar.TenantID)
					assert.Equal(t, tt.userID, calendar.UserID)
				}

				// Validate deep preloading structure (2-level deep)
				if tt.validateDeepPreload && len(result) > 0 {
					calendar := result[0] // Test first calendar

					// Level 1: Direct relationships should be loaded
					assert.NotNil(t, calendar.CalendarEntries, "CalendarEntries should be preloaded")
					assert.NotNil(t, calendar.CalendarSeries, "CalendarSeries should be preloaded")

					// Level 2: Nested relationships should be loaded
					if len(calendar.CalendarEntries) > 0 {
						entry := calendar.CalendarEntries[0]
						if entry.SeriesID != nil {
							assert.NotNil(t, entry.Series, "Entry.Series should be preloaded (2nd level)")
							assert.Equal(t, *entry.SeriesID, entry.Series.ID, "Entry should reference correct series")
						}
					}

					if len(calendar.CalendarSeries) > 0 {
						series := calendar.CalendarSeries[0]
						assert.NotNil(t, series.CalendarEntries, "Series.CalendarEntries should be preloaded (2nd level)")

						// Validate bidirectional relationship
						if len(series.CalendarEntries) > 0 {
							assert.Equal(t, series.ID, *series.CalendarEntries[0].SeriesID, "Series entries should reference back to series")
						}
					}
				}
			}

			// Verify all expected preload calls were made in the correct order
			mockDB.AssertExpectations(t)
		})
	}
}

// Test that verifies the preloading chain follows the expected pattern
func TestCalendarService_GetCalendarsWithDeepPreload_PreloadChain(t *testing.T) {
	fixtures := tests.NewTestFixtures()
	mockDB := &mocks.MockDB{}
	service := services.NewCalendarService(mockDB)

	t.Run("verifies correct preload chain execution", func(t *testing.T) {
		mockCalendars := []entities.Calendar{*fixtures.CreateMockCalendar()}

		// Track the order of preload calls
		var preloadCalls []string

		mockDB.On("Preload", "CalendarEntries").Run(func(args mock.Arguments) {
			preloadCalls = append(preloadCalls, "CalendarEntries")
		}).Return(mockDB)

		mockDB.On("Preload", "CalendarEntries.Series").Run(func(args mock.Arguments) {
			preloadCalls = append(preloadCalls, "CalendarEntries.Series")
		}).Return(mockDB)

		mockDB.On("Preload", "CalendarSeries").Run(func(args mock.Arguments) {
			preloadCalls = append(preloadCalls, "CalendarSeries")
		}).Return(mockDB)

		mockDB.On("Preload", "CalendarSeries.CalendarEntries").Run(func(args mock.Arguments) {
			preloadCalls = append(preloadCalls, "CalendarSeries.CalendarEntries")
		}).Return(mockDB)

		mockDB.On("Preload", "ExternalCalendars").Run(func(args mock.Arguments) {
			preloadCalls = append(preloadCalls, "ExternalCalendars")
		}).Return(mockDB)

		mockDB.On("Where", "tenant_id = ? AND user_id = ?", fixtures.TenantID, fixtures.UserID).Return(mockDB)
		mockDB.On("Find", mock.AnythingOfType("*[]entities.Calendar")).Return(nil, true, mockCalendars)

		// Execute
		result, err := service.GetCalendarsWithDeepPreload(fixtures.TenantID, fixtures.UserID)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		// Verify the preload chain was executed in the expected order
		expectedCalls := []string{
			"CalendarEntries",
			"CalendarEntries.Series",
			"CalendarSeries",
			"CalendarSeries.CalendarEntries",
			"ExternalCalendars",
		}
		assert.Equal(t, expectedCalls, preloadCalls, "Preload calls should be executed in the expected order")

		mockDB.AssertExpectations(t)
	})
}
