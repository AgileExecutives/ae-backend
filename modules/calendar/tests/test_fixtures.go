package tests

import (
	"encoding/json"
	"time"

	"github.com/unburdy/calendar-module/entities"
)

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
	startTime := time.Date(2025, 11, 1, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 1, 15, 0, 0, 0, time.UTC)

	participants, _ := json.Marshal([]string{"john@example.com"})

	return entities.CreateCalendarEntryRequest{
		CalendarID:   1,
		Title:        "New Meeting",
		IsException:  false,
		Participants: participants,
		StartTime:    &startTime,
		EndTime:      &endTime,
		Type:         "meeting",
		Description:  "New meeting description",
		Location:     "Room B",
		IsAllDay:     false,
	}
}

func (f *TestFixtures) CreateFreeSlotRequest() entities.FreeSlotRequest {
	return entities.FreeSlotRequest{
		Duration:  60, // 60 minutes
		Interval:  30, // 30 minutes between slots
		NumberMax: 5,  // Maximum 5 slots
	}
}

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
	startTime := time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)

	return entities.CreateCalendarSeriesRequest{
		CalendarID:  1,
		Title:       "Test Series",
		Weekday:     1, // Monday
		Interval:    1,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Description: "Test Series Description",
		Location:    "Test Location",
	}
}

func (f *TestFixtures) CreateUpdateCalendarSeriesRequest() entities.UpdateCalendarSeriesRequest {
	title := "Updated Test Series"
	description := "Updated Series Description"
	weekday := 2 // Tuesday
	interval := 2
	location := "Updated Location"

	return entities.UpdateCalendarSeriesRequest{
		Title:       &title,
		Description: &description,
		Weekday:     &weekday,
		Interval:    &interval,
		Location:    &location,
	}
}

func (f *TestFixtures) CreateMockCalendarSeries() *entities.CalendarSeries {
	startTime := time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)

	return &entities.CalendarSeries{
		ID:          1,
		CalendarID:  1,
		TenantID:    f.TenantID,
		UserID:      f.UserID,
		Title:       "Test Series",
		Weekday:     1, // Monday
		Interval:    1,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Description: "Test Series Description",
		Location:    "Test Location",
		EntryUUID:   "test-series-uuid",
		Sequence:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
