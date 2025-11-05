package integration_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/unburdy/calendar-module/entities"
	"github.com/unburdy/calendar-module/services"
)

var (
	testDB      *gorm.DB
	testService *services.CalendarService
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default,
	})
	require.NoError(t, err, "Failed to create test database")

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&entities.Calendar{},
		&entities.CalendarEntry{},
		&entities.CalendarSeries{},
		&entities.ExternalCalendar{},
	)
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

// TestCalendarService_CreateCalendar tests creating a calendar
func TestCalendarService_CreateCalendar(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewCalendarService(db)

	weeklyAvailability := map[string]interface{}{
		"monday": map[string]string{
			"start": "09:00",
			"end":   "17:00",
		},
	}
	availabilityJSON, _ := json.Marshal(weeklyAvailability)

	req := entities.CreateCalendarRequest{
		Title:              "Test Calendar",
		Color:              "#FF0000",
		WeeklyAvailability: availabilityJSON,
		Timezone:           "UTC",
	}

	calendar, err := service.CreateCalendar(req, 1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, calendar)
	assert.Equal(t, "Test Calendar", calendar.Title)
	assert.Equal(t, "#FF0000", calendar.Color)
	assert.Equal(t, uint(1), calendar.TenantID)
	assert.Equal(t, uint(1), calendar.UserID)
	assert.NotZero(t, calendar.ID)
	assert.NotEmpty(t, calendar.CalendarUUID)
}

// TestCalendarService_GetCalendarByID tests retrieving a calendar
func TestCalendarService_GetCalendarByID(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewCalendarService(db)

	// Create a calendar first
	weeklyAvailability := map[string]interface{}{
		"monday": map[string]string{"start": "09:00", "end": "17:00"},
	}
	availabilityJSON, _ := json.Marshal(weeklyAvailability)

	req := entities.CreateCalendarRequest{
		Title:              "Test Calendar",
		Color:              "#FF0000",
		WeeklyAvailability: availabilityJSON,
		Timezone:           "UTC",
	}

	created, err := service.CreateCalendar(req, 1, 1)
	require.NoError(t, err)

	// Retrieve it
	retrieved, err := service.GetCalendarByID(created.ID, 1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, "Test Calendar", retrieved.Title)
}

// TestCalendarService_UpdateCalendar tests updating a calendar
func TestCalendarService_UpdateCalendar(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewCalendarService(db)

	// Create a calendar
	weeklyAvailability := map[string]interface{}{
		"monday": map[string]string{"start": "09:00", "end": "17:00"},
	}
	availabilityJSON, _ := json.Marshal(weeklyAvailability)

	createReq := entities.CreateCalendarRequest{
		Title:              "Original Title",
		Color:              "#FF0000",
		WeeklyAvailability: availabilityJSON,
		Timezone:           "UTC",
	}

	created, err := service.CreateCalendar(createReq, 1, 1)
	require.NoError(t, err)

	// Update it
	newTitle := "Updated Title"
	newColor := "#00FF00"
	updateReq := entities.UpdateCalendarRequest{
		Title: &newTitle,
		Color: &newColor,
	}

	updated, err := service.UpdateCalendar(created.ID, 1, 1, updateReq)

	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, "#00FF00", updated.Color)
}

// TestCalendarService_DeleteCalendar tests deleting a calendar
func TestCalendarService_DeleteCalendar(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewCalendarService(db)

	// Create a calendar
	weeklyAvailability := map[string]interface{}{
		"monday": map[string]string{"start": "09:00", "end": "17:00"},
	}
	availabilityJSON, _ := json.Marshal(weeklyAvailability)

	req := entities.CreateCalendarRequest{
		Title:              "To Delete",
		Color:              "#FF0000",
		WeeklyAvailability: availabilityJSON,
		Timezone:           "UTC",
	}

	created, err := service.CreateCalendar(req, 1, 1)
	require.NoError(t, err)

	// Delete it
	err = service.DeleteCalendar(created.ID, 1, 1)
	assert.NoError(t, err)

	// Verify it's gone
	_, err = service.GetCalendarByID(created.ID, 1, 1)
	assert.Error(t, err)
}

// TestCalendarService_CreateCalendarEntry tests creating a calendar entry
func TestCalendarService_CreateCalendarEntry(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewCalendarService(db)

	// Create a calendar first
	weeklyAvailability := map[string]interface{}{
		"monday": map[string]string{"start": "09:00", "end": "17:00"},
	}
	availabilityJSON, _ := json.Marshal(weeklyAvailability)

	calReq := entities.CreateCalendarRequest{
		Title:              "Test Calendar",
		Color:              "#FF0000",
		WeeklyAvailability: availabilityJSON,
		Timezone:           "UTC",
	}

	calendar, err := service.CreateCalendar(calReq, 1, 1)
	require.NoError(t, err)

	// Create an entry
	startTime := time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)
	participants, _ := json.Marshal([]string{"user@example.com"})

	entryReq := entities.CreateCalendarEntryRequest{
		CalendarID:   calendar.ID,
		Title:        "Test Meeting",
		StartTime:    &startTime,
		EndTime:      &endTime,
		Participants: participants,
		Type:         "meeting",
		Description:  "Test meeting description",
		Location:     "Conference Room",
		IsAllDay:     false,
	}

	entry, err := service.CreateCalendarEntry(entryReq, 1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "Test Meeting", entry.Title)
	assert.Equal(t, calendar.ID, entry.CalendarID)
	assert.NotZero(t, entry.ID)
}

// TestCalendarService_GetCalendarsWithDeepPreload tests deep preloading
func TestCalendarService_GetCalendarsWithDeepPreload(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewCalendarService(db)

	// Create a calendar
	weeklyAvailability := map[string]interface{}{
		"monday": map[string]string{"start": "09:00", "end": "17:00"},
	}
	availabilityJSON, _ := json.Marshal(weeklyAvailability)

	calReq := entities.CreateCalendarRequest{
		Title:              "Test Calendar",
		Color:              "#FF0000",
		WeeklyAvailability: availabilityJSON,
		Timezone:           "UTC",
	}

	calendar, err := service.CreateCalendar(calReq, 1, 1)
	require.NoError(t, err)

	// Create an entry
	startTime := time.Date(2025, 11, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)
	participants, _ := json.Marshal([]string{"user@example.com"})

	entryReq := entities.CreateCalendarEntryRequest{
		CalendarID:   calendar.ID,
		Title:        "Test Meeting",
		StartTime:    &startTime,
		EndTime:      &endTime,
		Participants: participants,
		Type:         "meeting",
	}

	_, err = service.CreateCalendarEntry(entryReq, 1, 1)
	require.NoError(t, err)

	// Test deep preload
	calendars, err := service.GetCalendarsWithDeepPreload(1, 1)

	assert.NoError(t, err)
	assert.Len(t, calendars, 1)
	assert.Equal(t, "Test Calendar", calendars[0].Title)
	// Entries should be preloaded
	assert.NotNil(t, calendars[0].CalendarEntries)
}

// TestCalendarService_TenantIsolation tests that calendars are isolated by tenant
func TestCalendarService_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewCalendarService(db)

	weeklyAvailability := map[string]interface{}{
		"monday": map[string]string{"start": "09:00", "end": "17:00"},
	}
	availabilityJSON, _ := json.Marshal(weeklyAvailability)

	// Create calendars for different tenants
	req1 := entities.CreateCalendarRequest{
		Title:              "Tenant 1 Calendar",
		Color:              "#FF0000",
		WeeklyAvailability: availabilityJSON,
		Timezone:           "UTC",
	}
	cal1, err := service.CreateCalendar(req1, 1, 1)
	require.NoError(t, err)

	req2 := entities.CreateCalendarRequest{
		Title:              "Tenant 2 Calendar",
		Color:              "#00FF00",
		WeeklyAvailability: availabilityJSON,
		Timezone:           "UTC",
	}
	cal2, err := service.CreateCalendar(req2, 2, 1)
	require.NoError(t, err)

	// Tenant 1 should not see Tenant 2's calendar
	_, err = service.GetCalendarByID(cal2.ID, 1, 1)
	assert.Error(t, err)

	// Tenant 2 should not see Tenant 1's calendar
	_, err = service.GetCalendarByID(cal1.ID, 2, 1)
	assert.Error(t, err)

	// Each should see their own
	retrieved1, err := service.GetCalendarByID(cal1.ID, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, "Tenant 1 Calendar", retrieved1.Title)

	retrieved2, err := service.GetCalendarByID(cal2.ID, 2, 1)
	assert.NoError(t, err)
	assert.Equal(t, "Tenant 2 Calendar", retrieved2.Title)
}
