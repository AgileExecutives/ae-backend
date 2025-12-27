package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/ae-base-server/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	calendarEntities "github.com/unburdy/calendar-module/entities"
	calendarServices "github.com/unburdy/calendar-module/services"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/events"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	require.NoError(t, err)

	// Auto-migrate all required tables
	err = db.AutoMigrate(
		&entities.Client{},
		&entities.CostProvider{},
		&entities.Session{},
		&calendarEntities.Calendar{},
		&calendarEntities.CalendarEntry{},
		&calendarEntities.CalendarSeries{},
		&calendarEntities.ExternalCalendar{},
	)
	require.NoError(t, err)

	return db
}

// mockLogger is a simple logger for testing
type mockLogger struct{}

func (l *mockLogger) Debug(args ...interface{})                      {}
func (l *mockLogger) Info(args ...interface{})                       {}
func (l *mockLogger) Warn(args ...interface{})                       {}
func (l *mockLogger) Error(args ...interface{})                      {}
func (l *mockLogger) Fatal(args ...interface{})                      {}
func (l *mockLogger) With(key string, value interface{}) core.Logger { return l }

// TestSessionCancellationOnCalendarEntryDeletion tests that sessions are canceled when calendar entries are deleted
func TestSessionCancellationOnCalendarEntryDeletion(t *testing.T) {
	db := setupTestDB(t)
	logger := &mockLogger{}

	// Create test tenant and user context
	const (
		tenantID uint = 1
		userID   uint = 1
	)

	// Step 1: Create a client
	client := entities.Client{
		TenantID:  tenantID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}
	err := db.Create(&client).Error
	require.NoError(t, err)
	t.Logf("Created client with ID: %d", client.ID)

	// Step 2: Create a calendar
	calendarService := calendarServices.NewCalendarService(db)
	calendar, err := calendarService.CreateCalendar(calendarEntities.CreateCalendarRequest{
		Title:    "Test Calendar",
		Color:    "#FF5733",
		Timezone: "UTC",
	}, tenantID, userID)
	require.NoError(t, err)
	t.Logf("Created calendar with ID: %d", calendar.ID)

	// Step 3: Create a single calendar entry
	startTime := time.Date(2025, 11, 26, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 26, 11, 0, 0, 0, time.UTC)

	calendarEntry, err := calendarService.CreateCalendarEntry(calendarEntities.CreateCalendarEntryRequest{
		CalendarID:  calendar.ID,
		Title:       "Single Session",
		StartTime:   &startTime,
		EndTime:     &endTime,
		Description: "Single therapy session",
		Location:    "Office 101",
		Timezone:    "UTC",
	}, tenantID, userID)
	require.NoError(t, err)
	t.Logf("Created calendar entry with ID: %d", calendarEntry.ID)

	// Step 4: Create a session linked to the calendar entry
	sessionService := services.NewSessionService(db, nil)
	session, err := sessionService.CreateSession(entities.CreateSessionRequest{
		ClientID:          client.ID,
		CalendarEntryID:   calendarEntry.ID,
		OriginalDate:      startTime.Format(time.RFC3339),
		OriginalStartTime: startTime.Format(time.RFC3339),
		DurationMin:       60,
		Type:              "therapy",
		NumberUnits:       1,
		Status:            "scheduled",
		Documentation:     "Initial session notes",
	}, tenantID)
	require.NoError(t, err)
	require.NotNil(t, session.CalendarEntryID, "Session should have a calendar entry ID")
	t.Logf("Created session with ID: %d, CalendarEntryID: %d, Status: %s", session.ID, *session.CalendarEntryID, session.Status)

	// Verify session is scheduled with correct calendar entry
	assert.Equal(t, "scheduled", session.Status)
	assert.Equal(t, calendarEntry.ID, *session.CalendarEntryID)
	assert.Equal(t, "Initial session notes", session.Documentation)
	assert.Equal(t, startTime.Truncate(time.Second), session.OriginalStartTime.Truncate(time.Second))

	// Step 5: Register the event handler
	eventHandler := events.NewCalendarEntryDeletedHandler(db, logger)

	// Step 6: Simulate calendar entry deletion by triggering the event directly
	// (In production, this event would be published by the calendar service)
	t.Logf("Simulating deletion of calendar entry with ID: %d", calendarEntry.ID)

	// Step 7: Manually trigger the event (simulating event bus)
	eventPayload := map[string]interface{}{
		"calendar_entry_id": calendarEntry.ID,
		"tenant_id":         tenantID,
		"user_id":           userID,
		"calendar_id":       calendar.ID,
	}
	err = eventHandler.Handle(eventPayload)
	require.NoError(t, err)

	// Step 8: Verify the session was updated correctly
	var updatedSession entities.Session
	err = db.First(&updatedSession, session.ID).Error
	require.NoError(t, err)

	calEntryIDStr := "nil"
	if updatedSession.CalendarEntryID != nil {
		calEntryIDStr = fmt.Sprintf("%d", *updatedSession.CalendarEntryID)
	}
	t.Logf("Updated session - ID: %d, CalendarEntryID: %s, Status: %s, Documentation: %s",
		updatedSession.ID, calEntryIDStr, updatedSession.Status, updatedSession.Documentation)

	// Assertions
	assert.Equal(t, "canceled", updatedSession.Status, "Session status should be 'canceled'")
	assert.Nil(t, updatedSession.CalendarEntryID, "Session calendar_entry_id should be nil")
	assert.Contains(t, updatedSession.Documentation, "Calendar entry deleted", "Documentation should contain cancellation note")
	assert.Contains(t, updatedSession.Documentation, "Initial session notes", "Documentation should preserve original notes")
}

// TestSessionSeriesCancellationOnCalendarEntryDeletion tests canceling multiple sessions from a series
func TestSessionSeriesCancellationOnCalendarEntryDeletion(t *testing.T) {
	db := setupTestDB(t)
	logger := &mockLogger{}

	const (
		tenantID uint = 1
		userID   uint = 1
	)

	// Step 1: Create a client
	client := entities.Client{
		TenantID:  tenantID,
		FirstName: "Jane",
		LastName:  "Smith",
		Email:     "jane.smith@example.com",
	}
	err := db.Create(&client).Error
	require.NoError(t, err)

	// Step 2: Create a calendar
	calendarService := calendarServices.NewCalendarService(db)
	calendar, err := calendarService.CreateCalendar(calendarEntities.CreateCalendarRequest{
		Title:    "Therapy Calendar",
		Color:    "#00FF00",
		Timezone: "UTC",
	}, tenantID, userID)
	require.NoError(t, err)

	// Step 3: Create a recurring series with multiple entries
	startTime := time.Date(2025, 11, 26, 14, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 26, 15, 0, 0, 0, time.UTC)
	lastDate := time.Date(2025, 12, 10, 15, 0, 0, 0, time.UTC)

	series, entries, err := calendarService.CreateCalendarSeriesWithEntries(calendarEntities.CreateCalendarSeriesRequest{
		CalendarID:    calendar.ID,
		Title:         "Weekly Therapy",
		IntervalType:  "weekly",
		IntervalValue: 1,
		StartTime:     &startTime,
		EndTime:       &endTime,
		LastDate:      &lastDate,
		Description:   "Weekly therapy sessions",
		Location:      "Therapy Room",
		Timezone:      "UTC",
	}, tenantID, userID)
	require.NoError(t, err)
	require.Greater(t, len(entries), 1, "Should create multiple entries for the series")
	t.Logf("Created series with ID: %d and %d entries", series.ID, len(entries))

	// Step 4: Create sessions for each calendar entry using BookSessions
	// (BookSessions automatically extracts original_date and original_start_time from calendar entries)
	sessionService := services.NewSessionService(db, nil)
	createdSessions := make([]entities.Session, 0, len(entries))

	for i, entry := range entries {
		// For testing, we create sessions manually to have more control
		// In production, BookSessions would handle this
		session, err := sessionService.CreateSession(entities.CreateSessionRequest{
			ClientID:          client.ID,
			CalendarEntryID:   entry.ID,
			OriginalDate:      entry.StartTime.Format(time.RFC3339),
			OriginalStartTime: entry.StartTime.Format(time.RFC3339),
			DurationMin:       60,
			Type:              "therapy",
			NumberUnits:       1,
			Status:            "scheduled",
			Documentation:     "",
		}, tenantID)
		require.NoError(t, err)
		createdSessions = append(createdSessions, *session)
		t.Logf("Created session %d with ID: %d for CalendarEntry: %d", i+1, session.ID, entry.ID)
	}

	// Verify all sessions are scheduled
	for _, session := range createdSessions {
		assert.Equal(t, "scheduled", session.Status)
		assert.NotNil(t, session.CalendarEntryID, "Session should have a calendar entry ID")
	}

	// Step 5: Delete one calendar entry from the middle of the series (simulated)
	entryToDelete := entries[1] // Delete the second entry
	t.Logf("Simulating deletion of calendar entry ID: %d", entryToDelete.ID)

	// Step 6: Trigger the event handler
	eventHandler := events.NewCalendarEntryDeletedHandler(db, logger)
	eventPayload := map[string]interface{}{
		"calendar_entry_id": entryToDelete.ID,
		"tenant_id":         tenantID,
		"user_id":           userID,
		"calendar_id":       calendar.ID,
	}
	err = eventHandler.Handle(eventPayload)
	require.NoError(t, err)

	// Step 7: Verify only the session for the deleted entry was canceled
	var allSessions []entities.Session
	err = db.Where("client_id = ?", client.ID).Order("id").Find(&allSessions).Error
	require.NoError(t, err)

	canceledCount := 0
	scheduledCount := 0

	for _, session := range allSessions {
		calEntryIDStr := "nil"
		if session.CalendarEntryID != nil {
			calEntryIDStr = fmt.Sprintf("%d", *session.CalendarEntryID)
		}
		t.Logf("Session ID: %d, CalendarEntryID: %s, Status: %s",
			session.ID, calEntryIDStr, session.Status)

		if session.Status == "canceled" {
			canceledCount++
			assert.Nil(t, session.CalendarEntryID, "Canceled session should have calendar_entry_id = nil")
			assert.Contains(t, session.Documentation, "Calendar entry deleted")
		} else if session.Status == "scheduled" {
			scheduledCount++
			assert.NotNil(t, session.CalendarEntryID, "Scheduled session should have calendar_entry_id")
		}
	}

	assert.Equal(t, 1, canceledCount, "Exactly one session should be canceled")
	assert.Equal(t, len(entries)-1, scheduledCount, "Other sessions should remain scheduled")
}

// TestSessionNotCanceledIfAlreadyConducted tests that conducted sessions are not affected
func TestSessionNotCanceledIfAlreadyConducted(t *testing.T) {
	db := setupTestDB(t)
	logger := &mockLogger{}

	const (
		tenantID uint = 1
		userID   uint = 1
	)

	// Create client
	client := entities.Client{
		TenantID:  tenantID,
		FirstName: "Bob",
		LastName:  "Johnson",
		Email:     "bob@example.com",
	}
	err := db.Create(&client).Error
	require.NoError(t, err)

	// Create calendar
	calendarService := calendarServices.NewCalendarService(db)
	calendar, err := calendarService.CreateCalendar(calendarEntities.CreateCalendarRequest{
		Title:    "Test Calendar",
		Color:    "#0000FF",
		Timezone: "UTC",
	}, tenantID, userID)
	require.NoError(t, err)

	// Create calendar entry
	startTime := time.Date(2025, 11, 20, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2025, 11, 20, 11, 0, 0, 0, time.UTC)

	calendarEntry, err := calendarService.CreateCalendarEntry(calendarEntities.CreateCalendarEntryRequest{
		CalendarID: calendar.ID,
		Title:      "Past Session",
		StartTime:  &startTime,
		EndTime:    &endTime,
		Timezone:   "UTC",
	}, tenantID, userID)
	require.NoError(t, err)

	// Create session with "conducted" status
	sessionService := services.NewSessionService(db, nil)
	session, err := sessionService.CreateSession(entities.CreateSessionRequest{
		ClientID:          client.ID,
		CalendarEntryID:   calendarEntry.ID,
		OriginalDate:      startTime.Format(time.RFC3339),
		OriginalStartTime: startTime.Format(time.RFC3339),
		DurationMin:       60,
		Type:              "therapy",
		NumberUnits:       1,
		Status:            "conducted",
		Documentation:     "Session completed successfully",
	}, tenantID)
	require.NoError(t, err)
	t.Logf("Created conducted session with ID: %d", session.ID)

	// Delete the calendar entry (simulated)
	t.Logf("Simulating deletion of calendar entry ID: %d", calendarEntry.ID)

	// Trigger the event handler
	eventHandler := events.NewCalendarEntryDeletedHandler(db, logger)
	eventPayload := map[string]interface{}{
		"calendar_entry_id": calendarEntry.ID,
		"tenant_id":         tenantID,
		"user_id":           userID,
		"calendar_id":       calendar.ID,
	}
	err = eventHandler.Handle(eventPayload)
	require.NoError(t, err)

	// Verify the session was NOT modified (since it was already conducted)
	var updatedSession entities.Session
	err = db.First(&updatedSession, session.ID).Error
	require.NoError(t, err)

	assert.Equal(t, "conducted", updatedSession.Status, "Conducted session should not be changed to canceled")
	require.NotNil(t, updatedSession.CalendarEntryID, "Conducted session should keep its calendar_entry_id")
	assert.Equal(t, calendarEntry.ID, *updatedSession.CalendarEntryID, "Conducted session should keep its calendar_entry_id")
	assert.Equal(t, "Session completed successfully", updatedSession.Documentation, "Documentation should not be modified")
	assert.NotContains(t, updatedSession.Documentation, "Calendar entry deleted", "Should not add deletion note to conducted session")
}
