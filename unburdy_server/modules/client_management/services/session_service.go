package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	emailServices "github.com/ae-base-server/modules/email/services"
	bookingServices "github.com/unburdy/booking-module/services"
	calendarEntities "github.com/unburdy/calendar-module/entities"
	calendarServices "github.com/unburdy/calendar-module/services"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
)

type SessionService struct {
	db                 *gorm.DB
	bookingLinkService *bookingServices.BookingLinkService
	emailService       *emailServices.EmailService
}

// NewSessionService creates a new session service
func NewSessionService(db *gorm.DB, emailService *emailServices.EmailService) *SessionService {
	if emailService == nil {
		fmt.Println("⚠️ NewSessionService: emailService is NIL!")
	} else {
		fmt.Println("✅ NewSessionService: emailService is SET!")
	}
	return &SessionService{
		db:           db,
		emailService: emailService,
	}
}

// SetBookingLinkService sets the booking link service for token validation
func (s *SessionService) SetBookingLinkService(bookingLinkService *bookingServices.BookingLinkService) {
	s.bookingLinkService = bookingLinkService
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

	// Parse original date and start time (UTC)
	originalDate, err := time.Parse(time.RFC3339, req.OriginalDate)
	if err != nil {
		return nil, fmt.Errorf("invalid original_date format (expected RFC3339/UTC): %w", err)
	}

	originalStartTime, err := time.Parse(time.RFC3339, req.OriginalStartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid original_start_time format (expected RFC3339/UTC): %w", err)
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = "scheduled"
	}

	calendarEntryID := req.CalendarEntryID
	session := entities.Session{
		TenantID:          tenantID,
		ClientID:          req.ClientID,
		CalendarEntryID:   &calendarEntryID,
		OriginalDate:      originalDate.UTC(),
		OriginalStartTime: originalStartTime.UTC(),
		DurationMin:       req.DurationMin,
		Type:              req.Type,
		NumberUnits:       req.NumberUnits,
		Status:            status,
		Documentation:     req.Documentation,
	}

	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetSessionByID retrieves a session by ID
func (s *SessionService) GetSessionByID(id, tenantID uint) (*entities.Session, error) {
	var session entities.Session
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// GetSessionByCalendarEntryID retrieves a session by calendar entry ID
func (s *SessionService) GetSessionByCalendarEntryID(calendarEntryID, tenantID uint) (*entities.Session, error) {
	var session entities.Session
	if err := s.db.Where("calendar_entry_id = ? AND tenant_id = ?", calendarEntryID, tenantID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session with calendar entry ID %d not found", calendarEntryID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// GetAllSessions retrieves all sessions with pagination
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

	// Verify calendar exists and user has access to it
	var calendar struct {
		ID       uint
		TenantID uint
		UserID   uint
	}
	if err := s.db.Table("calendars").
		Select("id, tenant_id, user_id").
		Where("id = ? AND tenant_id = ? AND user_id = ? AND deleted_at IS NULL", req.CalendarID, tenantID, userID).
		First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("calendar with ID %d not found or access denied", req.CalendarID)
		}
		return nil, nil, fmt.Errorf("failed to verify calendar: %w", err)
	}

	// Parse time fields and ensure UTC
	var startTime, endTime, lastDate *time.Time

	if st, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
		utcTime := st.UTC()
		startTime = &utcTime
	} else {
		return nil, nil, fmt.Errorf("invalid start_time format (expected RFC3339/UTC): %w", err)
	}

	if et, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
		utcTime := et.UTC()
		endTime = &utcTime
	} else {
		return nil, nil, fmt.Errorf("invalid end_time format (expected RFC3339/UTC): %w", err)
	}

	if req.LastDate != "" {
		if ld, err := time.Parse(time.RFC3339, req.LastDate); err == nil {
			utcTime := ld.UTC()
			lastDate = &utcTime
		} else {
			return nil, nil, fmt.Errorf("invalid last_date format (expected RFC3339/UTC): %w", err)
		}
	}

	calendarService := calendarServices.NewCalendarService(s.db)
	var entries []calendarEntities.CalendarEntry
	var seriesID *uint

	// Check if this is a recurring series or a single entry
	if req.IntervalType != "" && req.IntervalType != "none" {
		// Create recurring series with multiple entries
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

		series, generatedEntries, err := calendarService.CreateCalendarSeriesWithEntries(seriesReq, tenantID, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create calendar series: %w", err)
		}
		seriesID = &series.ID
		entries = generatedEntries
	} else {
		// Create single calendar entry (no series)
		entryReq := calendarEntities.CreateCalendarEntryRequest{
			CalendarID:  req.CalendarID,
			Title:       req.Title,
			StartTime:   startTime,
			EndTime:     endTime,
			Description: req.Description,
			Location:    req.Location,
			Timezone:    req.Timezone,
		}

		entry, err := calendarService.CreateCalendarEntry(entryReq, tenantID, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create calendar entry: %w", err)
		}
		entries = []calendarEntities.CalendarEntry{*entry}
	}

	// Create sessions for each calendar entry
	var sessions []entities.Session
	for _, entry := range entries {
		// Extract original date and start time from calendar entry
		// Ensure they are in UTC
		var originalDate, originalStartTime time.Time
		if entry.StartTime != nil {
			originalStartTime = entry.StartTime.UTC()
			// Set original date to the date part of start time (at midnight UTC)
			originalDate = time.Date(
				entry.StartTime.Year(),
				entry.StartTime.Month(),
				entry.StartTime.Day(),
				0, 0, 0, 0, time.UTC,
			)
		} else {
			// Fallback if start time is not set (shouldn't happen)
			now := time.Now().UTC()
			originalStartTime = now
			originalDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		}

		entryID := entry.ID
		session := entities.Session{
			TenantID:          tenantID,
			ClientID:          req.ClientID,
			CalendarEntryID:   &entryID,
			OriginalDate:      originalDate,
			OriginalStartTime: originalStartTime,
			DurationMin:       req.DurationMin,
			Type:              req.Type,
			NumberUnits:       req.NumberUnits,
			Status:            "scheduled",
			Documentation:     "",
		}

		if err := s.db.Create(&session).Error; err != nil {
			// If session creation fails, we should ideally rollback the series/entries
			// For now, log the error and continue
			return nil, nil, fmt.Errorf("failed to create session for entry %d: %w", entry.ID, err)
		}

		sessions = append(sessions, session)
	}

	return seriesID, sessions, nil
}

// BookSessionsWithToken creates sessions using a booking token (retrieves client_id and calendar_id from token)
func (s *SessionService) BookSessionsWithToken(token string, req entities.BookSessionsWithTokenRequest) (*uint, []entities.Session, error) {
	// Validate the booking link service is configured
	if s.bookingLinkService == nil {
		return nil, nil, fmt.Errorf("booking link service not configured")
	}

	// Validate the token using the booking link service
	claims, err := s.bookingLinkService.ValidateBookingLink(token)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid or expired booking token: %w", err)
	}

	// Create the full BookSessionsRequest with data from token claims
	fullReq := entities.BookSessionsRequest{
		ClientID:      claims.ClientID,
		CalendarID:    claims.CalendarID,
		Title:         req.Title,
		Description:   req.Description,
		IntervalType:  req.IntervalType,
		IntervalValue: req.IntervalValue,
		LastDate:      req.LastDate,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		DurationMin:   req.DurationMin,
		Type:          req.Type,
		NumberUnits:   req.NumberUnits,
		Location:      req.Location,
		Timezone:      req.Timezone,
	}

	// Use the existing BookSessions method with tenant and user from token
	seriesID, sessions, err := s.BookSessions(fullReq, claims.TenantID, claims.UserID)
	if err != nil {
		return nil, nil, err
	}

	// After successful booking, invalidate one-time tokens to prevent reuse
	if err := s.bookingLinkService.InvalidateOneTimeToken(token, claims); err != nil {
		// Log the error but don't fail the booking - it already succeeded
		// In production, you might want to log this with proper logging
		fmt.Printf("Warning: failed to invalidate one-time token: %v\n", err)
	}

	// Send confirmation email to client
	if err := s.sendBookingConfirmationEmail(claims.ClientID, claims.TenantID, fullReq, sessions); err != nil {
		// Log the error but don't fail the booking - it already succeeded
		fmt.Printf("Warning: failed to send confirmation email: %v\n", err)
	}

	return seriesID, sessions, nil
}

// sendBookingConfirmationEmail sends a confirmation email to the client after successful booking
func (s *SessionService) sendBookingConfirmationEmail(clientID, tenantID uint, req entities.BookSessionsRequest, sessions []entities.Session) error {
	fmt.Printf("📧 sendBookingConfirmationEmail called - emailService is nil: %v\n", s.emailService == nil)
	if s.emailService == nil {
		return fmt.Errorf("email service not configured")
	}

	// Fetch client information to get email address
	var client entities.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", clientID, tenantID).First(&client).Error; err != nil {
		return fmt.Errorf("failed to fetch client: %w", err)
	}

	// Determine which email to use (prefer primary email, fall back to contact email)
	recipientEmail := client.Email
	if recipientEmail == "" {
		recipientEmail = client.ContactEmail
	}
	if recipientEmail == "" {
		return fmt.Errorf("no email address found for client")
	}

	// Prepare email data
	clientName := fmt.Sprintf("%s %s", client.FirstName, client.LastName)
	isSeries := len(sessions) > 1

	// Use client's timezone for displaying times
	timezone := client.Timezone
	if timezone == "" {
		timezone = "Europe/Berlin"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to Europe/Berlin if timezone is invalid
		loc, _ = time.LoadLocation("Europe/Berlin")
	}

	// Build appointments list for series
	// Times in DB are in UTC, convert to local timezone for display
	var appointments []map[string]interface{}
	if isSeries {
		for _, session := range sessions {
			if session.OriginalStartTime.IsZero() {
				continue
			}
			// Convert UTC time to local timezone
			localStartTime := session.OriginalStartTime.In(loc)
			// Calculate end time from start time + duration
			endTime := localStartTime.Add(time.Duration(session.DurationMin) * time.Minute)
			appointments = append(appointments, map[string]interface{}{
				"Date":     localStartTime.Format("02.01.2006"),
				"TimeFrom": localStartTime.Format("15:04"),
				"TimeTo":   endTime.Format("15:04"),
			})
		}
	}

	// Prepare single appointment data
	var appointmentDate, timeFrom, timeTo string
	if len(sessions) > 0 && !sessions[0].OriginalStartTime.IsZero() {
		localStartTime := sessions[0].OriginalStartTime.In(loc)
		appointmentDate = localStartTime.Format("02.01.2006")
		timeFrom = localStartTime.Format("15:04")
		endTime := localStartTime.Add(time.Duration(sessions[0].DurationMin) * time.Minute)
		timeTo = endTime.Format("15:04")
	}

	// Prepare email data
	emailData := emailServices.EmailData{
		RecipientName: clientName,
		Subject:       fmt.Sprintf("Terminbestätigung - %s", req.Title),
		CustomData: map[string]interface{}{
			"ClientName":       clientName,
			"Title":            req.Title,
			"Description":      req.Description,
			"Duration":         req.DurationMin,
			"Location":         req.Location,
			"Type":             req.Type,
			"IsSeries":         isSeries,
			"AppointmentCount": len(sessions),
			"Appointments":     appointments,
			"AppointmentDate":  appointmentDate,
			"TimeFrom":         timeFrom,
			"TimeTo":           timeTo,
		},
	}

	// Send email using template
	return s.emailService.SendTemplateEmail(recipientEmail, emailServices.TemplateBookingConfirmation, emailData)
}
