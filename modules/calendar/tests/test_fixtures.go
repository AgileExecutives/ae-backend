package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/services"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	require.NoError(t, err, "Failed to create test database")

	// Customize the migrator to override timestamptz with datetime
	db.Config.Dialector = &sqlite.Dialector{
		DSN:        ":memory:",
		DriverName: sqlite.DriverName,
	}

	// Auto-migrate the calendar entities
	err = db.AutoMigrate(
		&entities.Calendar{},
		&entities.CalendarEntry{},
		&entities.CalendarSeries{},
	)
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// SetupTestService creates a test service with an in-memory database
func SetupTestService(t *testing.T) (*services.CalendarService, *gorm.DB) {
	db := SetupTestDB(t)
	service := services.NewCalendarService(db)
	return service, db
}

// TestFixtures contains common test data
type TestFixtures struct {
	TenantID uint
	UserID   uint
}

// NewTestFixtures creates a new test fixtures instance
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{
		TenantID: 1,
		UserID:   1,
	}
}

// Calendar test fixtures

func (f *TestFixtures) CreateMockCalendar() *entities.Calendar {
	weeklyAvailability, _ := json.Marshal(map[string]interface{}{
		"monday":    []string{"09:00-17:00"},
		"tuesday":   []string{"09:00-17:00"},
		"wednesday": []string{"09:00-17:00"},
		"thursday":  []string{"09:00-17:00"},
		"friday":    []string{"09:00-17:00"},
	})

	return &entities.Calendar{
		ID:                 1,
		TenantID:           f.TenantID,
		UserID:             f.UserID,
		Title:              "Test Calendar",
		Color:              "#FF5733",
		WeeklyAvailability: weeklyAvailability,
		CalendarUUID:       "test-calendar-uuid",
		Timezone:           "UTC",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func (f *TestFixtures) CreateMockCalendars() []entities.Calendar {
	return []entities.Calendar{
		*f.CreateMockCalendar(),
		{
			ID:           2,
			TenantID:     f.TenantID,
			UserID:       f.UserID,
			Title:        "Work Calendar",
			Color:        "#00FF00",
			CalendarUUID: "work-calendar-uuid",
			Timezone:     "UTC",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
}

func (f *TestFixtures) CreateCalendarRequest() entities.CreateCalendarRequest {
	weeklyAvailability, _ := json.Marshal(map[string]interface{}{
		"monday": []string{"09:00-17:00"},
	})

	return entities.CreateCalendarRequest{
		Title:              "New Calendar",
		Color:              "#FF0000",
		WeeklyAvailability: weeklyAvailability,
		Timezone:           "UTC",
	}
}

func (f *TestFixtures) CreateUpdateCalendarRequest() entities.UpdateCalendarRequest {
	title := "Updated Calendar"
	color := "#00FF00"

	return entities.UpdateCalendarRequest{
		Title: &title,
		Color: &color,
	}
}

// Calendar Entry test fixtures

func (f *TestFixtures) CreateMockCalendarEntry() *entities.CalendarEntry {
	startTime := time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)

	participants, _ := json.Marshal([]string{"user1@example.com", "user2@example.com"})

	return &entities.CalendarEntry{
		ID:           1,
		TenantID:     f.TenantID,
		UserID:       f.UserID,
		CalendarID:   1,
		Title:        "Test Meeting",
		IsException:  false,
		Participants: participants,
		StartTime:    &startTime,
		EndTime:      &endTime,
		Type:         "meeting",
		Description:  "Test meeting description",
		Location:     "Conference Room A",
		IsAllDay:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Additional fixture methods...

func (f *TestFixtures) CreateCalendarEntryRequest() entities.CreateCalendarEntryRequest {
	participants, _ := json.Marshal([]string{"john@example.com"})

	return entities.CreateCalendarEntryRequest{
		CalendarID:   1,
		Title:        "New Meeting",
		IsException:  false,
		Participants: participants,
		// Skip time fields to avoid SQLite scanning issues with timestamptz
		Type:        "meeting",
		Description: "New meeting description",
		Location:    "Room B",
		IsAllDay:    false,
	}
}

// Note: FreeSlotRequest moved to booking module

func (f *TestFixtures) CreateUpdateCalendarEntryRequest() entities.UpdateCalendarEntryRequest {
	title := "Updated Meeting"
	description := "Updated meeting description"
	location := "Updated Room C"

	return entities.UpdateCalendarEntryRequest{
		Title:       &title,
		Description: &description,
		Location:    &location,
	}
}

func (f *TestFixtures) CreateImportHolidaysRequest() entities.ImportHolidaysRequest {
	return entities.ImportHolidaysRequest{
		State:    "BW",
		YearFrom: 2024,
		YearTo:   2025,
		Holidays: entities.UnburdyHolidaysData{
			SchoolHolidays: make(map[string]map[string][2]string),
			PublicHolidays: make(map[string]map[string]string),
		},
	}
}

// Calendar Series fixtures

func (f *TestFixtures) CreateCalendarSeriesRequest() entities.CreateCalendarSeriesRequest {
	return entities.CreateCalendarSeriesRequest{
		CalendarID:    1,
		Title:         "Test Series",
		IntervalType:  "weekly",
		IntervalValue: 1,
		// Skip time fields to avoid SQLite scanning issues
		Description: "Test Series Description",
		Location:    "Test Location",
	}
}

func (f *TestFixtures) CreateUpdateCalendarSeriesRequest() entities.UpdateCalendarSeriesRequest {
	title := "Updated Test Series"
	description := "Updated Series Description"
	intervalType := "weekly"
	intervalValue := 2
	location := "Updated Location"

	return entities.UpdateCalendarSeriesRequest{
		Title:         &title,
		Description:   &description,
		IntervalType:  &intervalType,
		IntervalValue: &intervalValue,
		Location:      &location,
		// Skip time fields to avoid SQLite scanning issues
	}
}

func (f *TestFixtures) CreateMockCalendarSeries() *entities.CalendarSeries {
	startTime := time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)
	lastDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	return &entities.CalendarSeries{
		ID:            1,
		CalendarID:    1,
		TenantID:      f.TenantID,
		UserID:        f.UserID,
		Title:         "Test Series",
		IntervalType:  "weekly",
		IntervalValue: 1,
		StartTime:     &startTime,
		EndTime:       &endTime,
		LastDate:      &lastDate,
		Description:   "Test Series Description",
		Location:      "Test Location",
		EntryUUID:     "test-series-uuid",
		Sequence:      0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
