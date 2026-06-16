package services

import (
	"testing"
	"time"

	calendarEntities "github.com/ae/shared-modules/calendar/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSessionDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entities.Client{}, &entities.Session{}, &entities.CostProvider{}))
	return db
}

func createTestClient(t *testing.T, db *gorm.DB, tenantID uint) *entities.Client {
	t.Helper()
	client := entities.Client{
		TenantID:  tenantID,
		FirstName: "Test",
		LastName:  "Client",
	}
	require.NoError(t, db.Create(&client).Error)
	return &client
}

func makeSessionReq(clientID uint) entities.CreateSessionRequest {
	now := time.Now().UTC()
	return entities.CreateSessionRequest{
		ClientID:          clientID,
		OriginalDate:      now.Format(time.RFC3339),
		OriginalStartTime: now.Format(time.RFC3339),
		DurationMin:       50,
		Type:              "individual",
		NumberUnits:       1,
		Status:            "scheduled",
	}
}

func TestSessionService_CreateSession(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	client := createTestClient(t, db, 1)

	t.Run("creates session successfully", func(t *testing.T) {
		s, err := svc.CreateSession(makeSessionReq(client.ID), 1)
		require.NoError(t, err)
		assert.NotZero(t, s.ID)
		assert.Equal(t, client.ID, s.ClientID)
		assert.Equal(t, "scheduled", s.Status)
	})

	t.Run("client not found returns error", func(t *testing.T) {
		_, err := svc.CreateSession(makeSessionReq(9999), 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("invalid date format returns error", func(t *testing.T) {
		req := makeSessionReq(client.ID)
		req.OriginalDate = "not-a-date"
		_, err := svc.CreateSession(req, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "original_date")
	})

	t.Run("invalid start time format returns error", func(t *testing.T) {
		req := makeSessionReq(client.ID)
		req.OriginalStartTime = "not-a-time"
		_, err := svc.CreateSession(req, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "original_start_time")
	})

	t.Run("default status when empty", func(t *testing.T) {
		req := makeSessionReq(client.ID)
		req.Status = ""
		s, err := svc.CreateSession(req, 1)
		require.NoError(t, err)
		assert.Equal(t, "scheduled", s.Status)
	})
}

func TestSessionService_GetSessionByID(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	client := createTestClient(t, db, 1)

	created, err := svc.CreateSession(makeSessionReq(client.ID), 1)
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		s, err := svc.GetSessionByID(created.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, created.ID, s.ID)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetSessionByID(9999, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		_, err := svc.GetSessionByID(created.ID, 2)
		require.Error(t, err)
	})
}

func TestSessionService_GetSessionByCalendarEntryID(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	client := createTestClient(t, db, 1)

	req := makeSessionReq(client.ID)
	req.CalendarEntryID = 42
	created, err := svc.CreateSession(req, 1)
	require.NoError(t, err)

	t.Run("found by calendar entry ID", func(t *testing.T) {
		s, err := svc.GetSessionByCalendarEntryID(42, 1)
		require.NoError(t, err)
		assert.Equal(t, created.ID, s.ID)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetSessionByCalendarEntryID(9999, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestSessionService_GetAllSessions(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	c1 := createTestClient(t, db, 1)
	c2 := createTestClient(t, db, 2)

	for i := 0; i < 3; i++ {
		_, err := svc.CreateSession(makeSessionReq(c1.ID), 1)
		require.NoError(t, err)
	}
	_, err := svc.CreateSession(makeSessionReq(c2.ID), 2)
	require.NoError(t, err)

	t.Run("returns tenant sessions", func(t *testing.T) {
		sessions, total, err := svc.GetAllSessions(1, 10, 1)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, sessions, 3)
	})

	t.Run("pagination", func(t *testing.T) {
		sessions, total, err := svc.GetAllSessions(1, 2, 1)
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, sessions, 2)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		sessions, total, err := svc.GetAllSessions(1, 10, 2)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, sessions, 1)
	})
}

func TestSessionService_GetSessionsByClientID(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	c1 := createTestClient(t, db, 1)
	c2 := createTestClient(t, db, 1)

	for i := 0; i < 2; i++ {
		_, err := svc.CreateSession(makeSessionReq(c1.ID), 1)
		require.NoError(t, err)
	}
	_, err := svc.CreateSession(makeSessionReq(c2.ID), 1)
	require.NoError(t, err)

	sessions, total, err := svc.GetSessionsByClientID(c1.ID, 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, sessions, 2)

	sessions2, total2, err := svc.GetSessionsByClientID(c2.ID, 1, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total2)
	assert.Len(t, sessions2, 1)
}

func TestSessionService_UpdateSession(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	client := createTestClient(t, db, 1)

	created, err := svc.CreateSession(makeSessionReq(client.ID), 1)
	require.NoError(t, err)

	newStatus := "conducted"
	newDuration := 60
	newDoc := "session notes"

	t.Run("update fields", func(t *testing.T) {
		updated, err := svc.UpdateSession(created.ID, 1, entities.UpdateSessionRequest{
			Status:      &newStatus,
			DurationMin: &newDuration,
			Documentation: &newDoc,
		})
		require.NoError(t, err)
		assert.Equal(t, "conducted", updated.Status)
		assert.Equal(t, 60, updated.DurationMin)
		assert.Equal(t, "session notes", updated.Documentation)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.UpdateSession(9999, 1, entities.UpdateSessionRequest{Status: &newStatus})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		_, err := svc.UpdateSession(created.ID, 2, entities.UpdateSessionRequest{Status: &newStatus})
		require.Error(t, err)
	})
}

func TestSessionService_DeleteSession(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	client := createTestClient(t, db, 1)

	created, err := svc.CreateSession(makeSessionReq(client.ID), 1)
	require.NoError(t, err)

	t.Run("delete succeeds", func(t *testing.T) {
		err := svc.DeleteSession(created.ID, 1)
		require.NoError(t, err)
		_, err2 := svc.GetSessionByID(created.ID, 1)
		require.Error(t, err2)
	})

	t.Run("not found returns error", func(t *testing.T) {
		err := svc.DeleteSession(9999, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestSessionService_BulkUpdateSessionStatus(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	client := createTestClient(t, db, 1)

	var sessionIDs []uint
	for i := 0; i < 3; i++ {
		s, err := svc.CreateSession(makeSessionReq(client.ID), 1)
		require.NoError(t, err)
		sessionIDs = append(sessionIDs, s.ID)
	}

	t.Run("bulk update to conducted", func(t *testing.T) {
		err := svc.BulkUpdateSessionStatus(sessionIDs, 1, client.ID, "conducted")
		require.NoError(t, err)
		for _, id := range sessionIDs {
			s, err := svc.GetSessionByID(id, 1)
			require.NoError(t, err)
			assert.Equal(t, "conducted", s.Status)
		}
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		err := svc.BulkUpdateSessionStatus(sessionIDs, 1, client.ID, "INVALID")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("empty IDs returns error", func(t *testing.T) {
		err := svc.BulkUpdateSessionStatus([]uint{}, 1, client.ID, "canceled")
		require.Error(t, err)
	})

	t.Run("wrong client returns error", func(t *testing.T) {
		err := svc.BulkUpdateSessionStatus(sessionIDs, 1, 9999, "canceled")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Expected to update")
	})
}

func TestSessionService_MarkSessionsAsConducted(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	client := createTestClient(t, db, 1)

	s, err := svc.CreateSession(makeSessionReq(client.ID), 1)
	require.NoError(t, err)

	err = svc.MarkSessionsAsConducted([]uint{s.ID}, 1, client.ID)
	require.NoError(t, err)

	updated, err := svc.GetSessionByID(s.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "conducted", updated.Status)
}

func TestSessionService_GetDetailedSessionsUpcoming7Days(t *testing.T) {
	db := setupSessionDB(t)
	svc := NewSessionService(db, nil)
	client := createTestClient(t, db, 1)

	now := time.Now().UTC()
	req := makeSessionReq(client.ID)
	req.OriginalStartTime = now.Add(24 * time.Hour).Format(time.RFC3339)
	req.OriginalDate = now.Add(24 * time.Hour).Format(time.RFC3339)
	_, err := svc.CreateSession(req, 1)
	require.NoError(t, err)

	// Session in the future but beyond 7 days
	req2 := makeSessionReq(client.ID)
	req2.OriginalStartTime = now.Add(10 * 24 * time.Hour).Format(time.RFC3339)
	req2.OriginalDate = req2.OriginalStartTime
	_, err = svc.CreateSession(req2, 1)
	require.NoError(t, err)

	results, err := svc.GetDetailedSessionsUpcoming7Days(1, &now)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func setupSessionWithCalendarDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&calendarEntities.Calendar{},
		&calendarEntities.CalendarEntry{},
		&calendarEntities.CalendarSeries{},
		&entities.Client{},
		&entities.Session{},
		&entities.CostProvider{},
	))
	return db
}

func TestSessionService_BookSessions_ClientNotFound(t *testing.T) {
	db := setupSessionWithCalendarDB(t)
	svc := NewSessionService(db, nil)

	req := entities.BookSessionsRequest{
		ClientID:    9999,
		CalendarID:  1,
		Title:       "Test Session",
		StartTime:   time.Now().UTC().Format(time.RFC3339),
		EndTime:     time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		DurationMin: 60,
		Type:        "therapy",
		NumberUnits: 1,
	}
	_, _, err := svc.BookSessions(req, 1, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSessionService_BookSessions_CalendarNotFound(t *testing.T) {
	db := setupSessionWithCalendarDB(t)
	svc := NewSessionService(db, nil)

	client := createTestClient(t, db, 1)

	req := entities.BookSessionsRequest{
		ClientID:    client.ID,
		CalendarID:  9999,
		Title:       "Test Session",
		StartTime:   time.Now().UTC().Format(time.RFC3339),
		EndTime:     time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		DurationMin: 60,
		Type:        "therapy",
		NumberUnits: 1,
	}
	_, _, err := svc.BookSessions(req, 1, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSessionService_BookSessions_SingleSession(t *testing.T) {
	db := setupSessionWithCalendarDB(t)
	svc := NewSessionService(db, nil)

	client := createTestClient(t, db, 1)

	// Create calendar directly
	cal := calendarEntities.Calendar{
		TenantID:     1,
		UserID:       10,
		Title:        "Test Calendar",
		CalendarUUID: "test-uuid-book-1234",
		Timezone:     "UTC",
	}
	require.NoError(t, db.Create(&cal).Error)
	calID := cal.ID

	req := entities.BookSessionsRequest{
		ClientID:    client.ID,
		CalendarID:  calID,
		Title:       "Therapy Session",
		StartTime:   time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		EndTime:     time.Now().UTC().Add(25 * time.Hour).Format(time.RFC3339),
		DurationMin: 60,
		Type:        "therapy",
		NumberUnits: 1,
	}
	_, sessions, err := svc.BookSessions(req, 1, 10)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "scheduled", sessions[0].Status)
}

func TestSessionService_BookSessions_InvalidStartTime(t *testing.T) {
	db := setupSessionWithCalendarDB(t)
	svc := NewSessionService(db, nil)

	client := createTestClient(t, db, 1)

	req := entities.BookSessionsRequest{
		ClientID:    client.ID,
		CalendarID:  1,
		Title:       "Session",
		StartTime:   "not-a-time",
		EndTime:     time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		DurationMin: 60,
		Type:        "therapy",
		NumberUnits: 1,
	}
	_, _, err := svc.BookSessions(req, 1, 10)
	require.Error(t, err)
}

func TestSessionService_BookSessionsWithToken_NoService(t *testing.T) {
	db := setupSessionWithCalendarDB(t)
	svc := NewSessionService(db, nil)

	req := entities.BookSessionsWithTokenRequest{
		Title:       "Test",
		StartTime:   time.Now().UTC().Format(time.RFC3339),
		EndTime:     time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		DurationMin: 60,
		Type:        "therapy",
		NumberUnits: 1,
	}
	_, _, err := svc.BookSessionsWithToken("some-token", req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "booking link service not configured")
}
