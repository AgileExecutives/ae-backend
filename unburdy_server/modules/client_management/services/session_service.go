package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	calendarEntities "github.com/unburdy/calendar-module/entities"
	calendarServices "github.com/unburdy/calendar-module/services"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
)

type SessionService struct {
	db *gorm.DB
}

// NewSessionService creates a new session service
func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

// CreateSession creates a new session
func (s *SessionService) CreateSession(req entities.CreateSessionRequest, tenantID uint) (*entities.Session, error) {
	// Verify client exists and belongs to tenant
	var client entities.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", req.ClientID, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client with ID %d not found", req.ClientID)
		}
		return nil, fmt.Errorf("failed to verify client: %w", err)
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = "scheduled"
	}

	session := entities.Session{
		TenantID:        tenantID,
		ClientID:        req.ClientID,
		CalendarEntryID: req.CalendarEntryID,
		DurationMin:     req.DurationMin,
		Type:            req.Type,
		NumberUnits:     req.NumberUnits,
		Status:          status,
		Documentation:   req.Documentation,
	}

	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetSessionByID returns a session by ID
func (s *SessionService) GetSessionByID(id, tenantID uint) (*entities.Session, error) {
	var session entities.Session
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch session: %w", err)
	}
	return &session, nil
}

// GetAllSessions returns all sessions for a tenant with pagination
func (s *SessionService) GetAllSessions(page, limit int, tenantID uint) ([]entities.Session, int, error) {
	var sessions []entities.Session
	var total int64

	offset := (page - 1) * limit

	if err := s.db.Model(&entities.Session{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	if err := s.db.Where("tenant_id = ?", tenantID).Offset(offset).Limit(limit).Find(&sessions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	return sessions, int(total), nil
}

// GetSessionsByClientID returns all sessions for a specific client
func (s *SessionService) GetSessionsByClientID(clientID, tenantID uint, page, limit int) ([]entities.Session, int, error) {
	var sessions []entities.Session
	var total int64

	offset := (page - 1) * limit

	if err := s.db.Model(&entities.Session{}).Where("client_id = ? AND tenant_id = ?", clientID, tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	if err := s.db.Where("client_id = ? AND tenant_id = ?", clientID, tenantID).
		Offset(offset).Limit(limit).Find(&sessions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	return sessions, int(total), nil
}

// UpdateSession updates an existing session
func (s *SessionService) UpdateSession(id, tenantID uint, req entities.UpdateSessionRequest) (*entities.Session, error) {
	var session entities.Session
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Update fields if provided
	if req.DurationMin != nil {
		session.DurationMin = *req.DurationMin
	}
	if req.Type != nil {
		session.Type = *req.Type
	}
	if req.NumberUnits != nil {
		session.NumberUnits = *req.NumberUnits
	}
	if req.Status != nil {
		session.Status = *req.Status
	}
	if req.Documentation != nil {
		session.Documentation = *req.Documentation
	}

	if err := s.db.Save(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &session, nil
}

// DeleteSession soft deletes a session
func (s *SessionService) DeleteSession(id, tenantID uint) error {
	var session entities.Session
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("session with ID %d not found", id)
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	if err := s.db.Delete(&session).Error; err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// BookSessions creates a calendar series and corresponding sessions for a client
func (s *SessionService) BookSessions(req entities.BookSessionsRequest, tenantID, userID uint) (*uint, []entities.Session, error) {
	// Verify client exists and belongs to tenant
	var client entities.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", req.ClientID, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("client with ID %d not found", req.ClientID)
		}
		return nil, nil, fmt.Errorf("failed to verify client: %w", err)
	}

	// Parse time fields
	var startTime, endTime, lastDate *time.Time

	if st, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
		startTime = &st
	} else {
		return nil, nil, fmt.Errorf("invalid start_time format: %w", err)
	}

	if et, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
		endTime = &et
	} else {
		return nil, nil, fmt.Errorf("invalid end_time format: %w", err)
	}

	if req.LastDate != "" {
		if ld, err := time.Parse(time.RFC3339, req.LastDate); err == nil {
			lastDate = &ld
		} else {
			return nil, nil, fmt.Errorf("invalid last_date format: %w", err)
		}
	}

	// Create calendar series request
	seriesReq := calendarEntities.CreateCalendarSeriesRequest{
		CalendarID:    req.CalendarID,
		Title:         req.Title,
		IntervalType:  req.IntervalType,
		IntervalValue: req.IntervalValue,
		LastDate:      lastDate,
		StartTime:     startTime,
		EndTime:       endTime,
		Description:   req.Description,
		Location:      req.Location,
		Timezone:      req.Timezone,
	}

	// Call calendar service to create series with entries
	calendarService := calendarServices.NewCalendarService(s.db)
	series, entries, err := calendarService.CreateCalendarSeriesWithEntries(seriesReq, tenantID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create calendar series: %w", err)
	}

	// Create sessions for each calendar entry
	var sessions []entities.Session
	for _, entry := range entries {
		session := entities.Session{
			TenantID:        tenantID,
			ClientID:        req.ClientID,
			CalendarEntryID: entry.ID,
			DurationMin:     req.DurationMin,
			Type:            req.Type,
			NumberUnits:     req.NumberUnits,
			Status:          "scheduled",
			Documentation:   "",
		}

		if err := s.db.Create(&session).Error; err != nil {
			// If session creation fails, we should ideally rollback the series/entries
			// For now, log the error and continue
			return nil, nil, fmt.Errorf("failed to create session for entry %d: %w", entry.ID, err)
		}

		sessions = append(sessions, session)
	}

	return &series.ID, sessions, nil
}

