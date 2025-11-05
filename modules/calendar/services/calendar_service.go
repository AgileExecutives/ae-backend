package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/unburdy/calendar-module/entities"
)

type CalendarService struct {
	db *gorm.DB
}

// NewCalendarService creates a new calendar service
func NewCalendarService(db *gorm.DB) *CalendarService {
	return &CalendarService{db: db}
}

// Calendar CRUD Operations

// CreateCalendar creates a new calendar for a specific user
func (s *CalendarService) CreateCalendar(req entities.CreateCalendarRequest, tenantID, userID uint) (*entities.Calendar, error) {
	calendar := entities.Calendar{
		TenantID:           tenantID,
		UserID:             userID,
		Title:              req.Title,
		Color:              req.Color,
		WeeklyAvailability: req.WeeklyAvailability,
		CalendarUUID:       uuid.New().String(),
		Timezone:           req.Timezone,
	}

	if err := s.db.Create(&calendar).Error; err != nil {
		return nil, fmt.Errorf("failed to create calendar: %w", err)
	}

	return &calendar, nil
}

// GetCalendarByID returns a calendar by ID within a tenant and user
func (s *CalendarService) GetCalendarByID(id, tenantID, userID uint) (*entities.Calendar, error) {
	var calendar entities.Calendar
	if err := s.db.Preload("CalendarSeries").Preload("CalendarEntries").Preload("ExternalCalendars").
		Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("calendar with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch calendar: %w", err)
	}
	return &calendar, nil
}

// GetAllCalendars returns all calendars for a user with pagination and 2-level deep preloading
// Preloads: CalendarEntries with their Series, CalendarSeries with their CalendarEntries, ExternalCalendars
func (s *CalendarService) GetAllCalendars(page, limit int, tenantID, userID uint) ([]entities.Calendar, int, error) {
	var calendars []entities.Calendar
	var total int64

	offset := (page - 1) * limit

	// Count total records for the user
	if err := s.db.Model(&entities.Calendar{}).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count calendars: %w", err)
	}

	// Get paginated records with 2-level deep preloaded relationships
	// Level 1: Direct relationships
	// Level 2: Nested relationships (entries->series, series->entries)
	if err := s.db.
		Preload("CalendarEntries").                // Load all calendar entries
		Preload("CalendarEntries.Series").         // Load series for each entry (2nd level)
		Preload("CalendarSeries").                 // Load all calendar series
		Preload("CalendarSeries.CalendarEntries"). // Load entries for each series (2nd level)
		Preload("ExternalCalendars").              // Load external calendars
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Offset(offset).Limit(limit).Find(&calendars).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch calendars: %w", err)
	}

	return calendars, int(total), nil
}

// GetCalendarsWithDeepPreload returns all calendars for a user with 2-level deep preloading (no pagination)
// This method is optimized for the /calendar endpoint that returns all calendar metadata
func (s *CalendarService) GetCalendarsWithDeepPreload(tenantID, userID uint) ([]entities.Calendar, error) {
	var calendars []entities.Calendar

	// Get all records with 2-level deep preloaded relationships
	// Level 1: Direct relationships (CalendarEntries, CalendarSeries, ExternalCalendars)
	// Level 2: Nested relationships (entries->series, series->entries)
	if err := s.db.
		Preload("CalendarEntries").                // Load all calendar entries
		Preload("CalendarEntries.Series").         // Load series for each entry (2nd level)
		Preload("CalendarSeries").                 // Load all calendar series
		Preload("CalendarSeries.CalendarEntries"). // Load entries for each series (2nd level)
		Preload("ExternalCalendars").              // Load external calendars
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Find(&calendars).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch calendars with deep preload: %w", err)
	}

	return calendars, nil
}

// UpdateCalendar updates an existing calendar within a tenant and user
func (s *CalendarService) UpdateCalendar(id, tenantID, userID uint, req entities.UpdateCalendarRequest) (*entities.Calendar, error) {
	var calendar entities.Calendar
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("calendar with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	// Update fields if provided
	if req.Title != nil {
		calendar.Title = *req.Title
	}
	if req.Color != nil {
		calendar.Color = *req.Color
	}
	if req.WeeklyAvailability != nil {
		calendar.WeeklyAvailability = *req.WeeklyAvailability
	}
	if req.Timezone != nil {
		calendar.Timezone = *req.Timezone
	}

	if err := s.db.Save(&calendar).Error; err != nil {
		return nil, fmt.Errorf("failed to update calendar: %w", err)
	}

	// Reload calendar with relationships
	if err := s.db.Preload("CalendarSeries").Preload("CalendarEntries").Preload("ExternalCalendars").
		First(&calendar, calendar.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload calendar: %w", err)
	}

	return &calendar, nil
}

// DeleteCalendar soft deletes a calendar within a tenant and user
func (s *CalendarService) DeleteCalendar(id, tenantID, userID uint) error {
	var calendar entities.Calendar
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("calendar with ID %d not found", id)
		}
		return fmt.Errorf("failed to get calendar: %w", err)
	}

	if err := s.db.Delete(&calendar).Error; err != nil {
		return fmt.Errorf("failed to delete calendar: %w", err)
	}

	return nil
}

// Calendar Entry CRUD Operations

// CreateCalendarEntry creates a new calendar entry
func (s *CalendarService) CreateCalendarEntry(req entities.CreateCalendarEntryRequest, tenantID, userID uint) (*entities.CalendarEntry, error) {
	// Verify calendar belongs to user
	var calendar entities.Calendar
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", req.CalendarID, tenantID, userID).First(&calendar).Error; err != nil {
		return nil, fmt.Errorf("calendar not found or access denied")
	}

	entry := entities.CalendarEntry{
		TenantID:     tenantID,
		UserID:       userID,
		CalendarID:   req.CalendarID,
		SeriesID:     req.SeriesID,
		Title:        req.Title,
		IsException:  req.IsException,
		Participants: req.Participants,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Type:         req.Type,
		Description:  req.Description,
		Location:     req.Location,
		Timezone:     req.Timezone,
		IsAllDay:     req.IsAllDay,
	}

	if err := s.db.Create(&entry).Error; err != nil {
		return nil, fmt.Errorf("failed to create calendar entry: %w", err)
	}

	return &entry, nil
}

// GetCalendarEntryByID returns a calendar entry by ID
func (s *CalendarService) GetCalendarEntryByID(id, tenantID, userID uint) (*entities.CalendarEntry, error) {
	var entry entities.CalendarEntry
	if err := s.db.Preload("Calendar").Preload("Series").
		Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("calendar entry with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch calendar entry: %w", err)
	}
	return &entry, nil
}

// GetAllCalendarEntries returns all calendar entries for a user with pagination
func (s *CalendarService) GetAllCalendarEntries(page, limit int, tenantID, userID uint) ([]entities.CalendarEntry, int, error) {
	var entries []entities.CalendarEntry
	var total int64

	offset := (page - 1) * limit

	if err := s.db.Model(&entities.CalendarEntry{}).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count calendar entries: %w", err)
	}

	if err := s.db.Preload("Calendar").Preload("Series").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).Offset(offset).Limit(limit).Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch calendar entries: %w", err)
	}

	return entries, int(total), nil
}

// UpdateCalendarEntry updates an existing calendar entry
func (s *CalendarService) UpdateCalendarEntry(id, tenantID, userID uint, req entities.UpdateCalendarEntryRequest) (*entities.CalendarEntry, error) {
	var entry entities.CalendarEntry
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("calendar entry with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get calendar entry: %w", err)
	}

	// Update fields if provided
	if req.Title != nil {
		entry.Title = *req.Title
	}
	if req.IsException != nil {
		entry.IsException = *req.IsException
	}
	if req.Participants != nil {
		entry.Participants = req.Participants
	}
	if req.StartTime != nil {
		entry.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		entry.EndTime = req.EndTime
	}
	if req.Type != nil {
		entry.Type = *req.Type
	}
	if req.Description != nil {
		entry.Description = *req.Description
	}
	if req.Location != nil {
		entry.Location = *req.Location
	}
	if req.Timezone != nil {
		entry.Timezone = *req.Timezone
	}
	if req.IsAllDay != nil {
		entry.IsAllDay = *req.IsAllDay
	}

	if err := s.db.Save(&entry).Error; err != nil {
		return nil, fmt.Errorf("failed to update calendar entry: %w", err)
	}

	return &entry, nil
}

// DeleteCalendarEntry soft deletes a calendar entry
func (s *CalendarService) DeleteCalendarEntry(id, tenantID, userID uint) error {
	var entry entities.CalendarEntry
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&entry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("calendar entry with ID %d not found", id)
		}
		return fmt.Errorf("failed to get calendar entry: %w", err)
	}

	if err := s.db.Delete(&entry).Error; err != nil {
		return fmt.Errorf("failed to delete calendar entry: %w", err)
	}

	return nil
}

// Calendar Series CRUD Operations

// CreateCalendarSeries creates a new calendar series
func (s *CalendarService) CreateCalendarSeries(req entities.CreateCalendarSeriesRequest, tenantID, userID uint) (*entities.CalendarSeries, error) {
	// Verify calendar belongs to user
	var calendar entities.Calendar
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", req.CalendarID, tenantID, userID).First(&calendar).Error; err != nil {
		return nil, fmt.Errorf("calendar not found or access denied")
	}

	series := entities.CalendarSeries{
		TenantID:             tenantID,
		UserID:               userID,
		CalendarID:           req.CalendarID,
		Title:                req.Title,
		Participants:         req.Participants,
		Weekday:              req.Weekday,
		Interval:             req.Interval,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		Description:          req.Description,
		Location:             req.Location,
		Timezone:             req.Timezone,
		EntryUUID:            uuid.New().String(),
		ExternalUID:          req.ExternalUID,
		ExternalCalendarUUID: req.ExternalCalendarUUID,
	}

	if err := s.db.Create(&series).Error; err != nil {
		return nil, fmt.Errorf("failed to create calendar series: %w", err)
	}

	return &series, nil
}

// GetCalendarSeriesByID returns a calendar series by ID
func (s *CalendarService) GetCalendarSeriesByID(id, tenantID, userID uint) (*entities.CalendarSeries, error) {
	var series entities.CalendarSeries
	if err := s.db.Preload("Calendar").Preload("CalendarEntries").
		Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&series).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("calendar series with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch calendar series: %w", err)
	}
	return &series, nil
}

// GetAllCalendarSeries returns all calendar series for a user with pagination
func (s *CalendarService) GetAllCalendarSeries(page, limit int, tenantID, userID uint) ([]entities.CalendarSeries, int, error) {
	var series []entities.CalendarSeries
	var total int64

	offset := (page - 1) * limit

	if err := s.db.Model(&entities.CalendarSeries{}).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count calendar series: %w", err)
	}

	if err := s.db.Preload("Calendar").Preload("CalendarEntries").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).Offset(offset).Limit(limit).Find(&series).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch calendar series: %w", err)
	}

	return series, int(total), nil
}

// UpdateCalendarSeries updates an existing calendar series
func (s *CalendarService) UpdateCalendarSeries(id, tenantID, userID uint, req entities.UpdateCalendarSeriesRequest) (*entities.CalendarSeries, error) {
	var series entities.CalendarSeries
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&series).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("calendar series with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get calendar series: %w", err)
	}

	// Update fields if provided
	if req.Title != nil {
		series.Title = *req.Title
	}
	if req.Participants != nil {
		series.Participants = req.Participants
	}
	if req.Weekday != nil {
		series.Weekday = *req.Weekday
	}
	if req.Interval != nil {
		series.Interval = *req.Interval
	}
	// Use new UTC timestamp fields if provided
	if req.StartTime != nil {
		series.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		series.EndTime = req.EndTime
	}
	if req.Description != nil {
		series.Description = *req.Description
	}
	if req.Location != nil {
		series.Location = *req.Location
	}
	if req.Timezone != nil {
		series.Timezone = *req.Timezone
	}
	if req.ExternalUID != nil {
		series.ExternalUID = *req.ExternalUID
	}

	if err := s.db.Save(&series).Error; err != nil {
		return nil, fmt.Errorf("failed to update calendar series: %w", err)
	}

	return &series, nil
}

// DeleteCalendarSeries soft deletes a calendar series
func (s *CalendarService) DeleteCalendarSeries(id, tenantID, userID uint) error {
	var series entities.CalendarSeries
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&series).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("calendar series with ID %d not found", id)
		}
		return fmt.Errorf("failed to get calendar series: %w", err)
	}

	if err := s.db.Delete(&series).Error; err != nil {
		return fmt.Errorf("failed to delete calendar series: %w", err)
	}

	return nil
}

// External Calendar CRUD Operations

// CreateExternalCalendar creates a new external calendar
func (s *CalendarService) CreateExternalCalendar(req entities.CreateExternalCalendarRequest, tenantID, userID uint) (*entities.ExternalCalendar, error) {
	// Verify calendar belongs to user
	var calendar entities.Calendar
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", req.CalendarID, tenantID, userID).First(&calendar).Error; err != nil {
		return nil, fmt.Errorf("calendar not found or access denied")
	}

	external := entities.ExternalCalendar{
		TenantID:     tenantID,
		UserID:       userID,
		CalendarID:   req.CalendarID,
		Title:        req.Title,
		URL:          req.URL,
		Settings:     req.Settings,
		Color:        req.Color,
		CalendarUUID: uuid.New().String(),
	}

	if err := s.db.Create(&external).Error; err != nil {
		return nil, fmt.Errorf("failed to create external calendar: %w", err)
	}

	return &external, nil
}

// GetExternalCalendarByID returns an external calendar by ID
func (s *CalendarService) GetExternalCalendarByID(id, tenantID, userID uint) (*entities.ExternalCalendar, error) {
	var external entities.ExternalCalendar
	if err := s.db.Preload("Calendar").
		Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&external).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("external calendar with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch external calendar: %w", err)
	}
	return &external, nil
}

// GetAllExternalCalendars returns all external calendars for a user with pagination
func (s *CalendarService) GetAllExternalCalendars(page, limit int, tenantID, userID uint) ([]entities.ExternalCalendar, int, error) {
	var externals []entities.ExternalCalendar
	var total int64

	offset := (page - 1) * limit

	if err := s.db.Model(&entities.ExternalCalendar{}).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count external calendars: %w", err)
	}

	if err := s.db.Preload("Calendar").
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).Offset(offset).Limit(limit).Find(&externals).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch external calendars: %w", err)
	}

	return externals, int(total), nil
}

// UpdateExternalCalendar updates an existing external calendar
func (s *CalendarService) UpdateExternalCalendar(id, tenantID, userID uint, req entities.UpdateExternalCalendarRequest) (*entities.ExternalCalendar, error) {
	var external entities.ExternalCalendar
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&external).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("external calendar with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get external calendar: %w", err)
	}

	// Update fields if provided
	if req.Title != nil {
		external.Title = *req.Title
	}
	if req.URL != nil {
		external.URL = *req.URL
	}
	if req.Settings != nil {
		external.Settings = *req.Settings
	}
	if req.Color != nil {
		external.Color = *req.Color
	}

	if err := s.db.Save(&external).Error; err != nil {
		return nil, fmt.Errorf("failed to update external calendar: %w", err)
	}

	return &external, nil
}

// DeleteExternalCalendar soft deletes an external calendar
func (s *CalendarService) DeleteExternalCalendar(id, tenantID, userID uint) error {
	var external entities.ExternalCalendar
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).First(&external).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("external calendar with ID %d not found", id)
		}
		return fmt.Errorf("failed to get external calendar: %w", err)
	}

	if err := s.db.Delete(&external).Error; err != nil {
		return fmt.Errorf("failed to delete external calendar: %w", err)
	}

	return nil
}

// Specialized Methods

// GetCalendarWeekView returns calendar entries for a specific week
func (s *CalendarService) GetCalendarWeekView(date time.Time, tenantID, userID uint) ([]entities.CalendarEntry, error) {
	// Calculate start and end of week (Sunday to Saturday)
	startOfWeek := date.AddDate(0, 0, -int(date.Weekday()))
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	var entries []entities.CalendarEntry
	if err := s.db.Preload("Calendar").Preload("Series").
		Where("tenant_id = ? AND user_id = ? AND date_from <= ? AND date_to >= ?",
			tenantID, userID, endOfWeek, startOfWeek).
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch week view: %w", err)
	}

	return entries, nil
}

// GetCalendarYearView returns calendar entries for a specific year
func (s *CalendarService) GetCalendarYearView(year int, tenantID, userID uint) ([]entities.CalendarEntry, error) {
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	var entries []entities.CalendarEntry
	if err := s.db.Preload("Calendar").Preload("Series").
		Where("tenant_id = ? AND user_id = ? AND date_from <= ? AND date_to >= ?",
			tenantID, userID, endOfYear, startOfYear).
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch year view: %w", err)
	}

	return entries, nil
}

// GetFreeSlots finds free time slots based on weekly availability and existing entries
func (s *CalendarService) GetFreeSlots(req entities.FreeSlotRequest, tenantID, userID uint) ([]entities.FreeSlot, error) {
	// This is a basic implementation - in a real system you would:
	// 1. Parse weekly_availability JSON from user's calendar
	// 2. Generate time slots based on availability
	// 3. Filter out occupied slots based on existing calendar entries
	// 4. Return available slots

	var freeSlots []entities.FreeSlot

	// For now, return a placeholder implementation
	// In a real implementation, you would query calendar entries and calculate availability
	startTime := time.Now().Add(24 * time.Hour) // Start from tomorrow
	for i := 0; i < req.NumberMax; i++ {
		slotStart := startTime.Add(time.Duration(i*req.Interval) * time.Minute)
		slotEnd := slotStart.Add(time.Duration(req.Duration) * time.Minute)

		freeSlots = append(freeSlots, entities.FreeSlot{
			StartTime: slotStart,
			EndTime:   slotEnd,
			Duration:  req.Duration,
		})
	}

	return freeSlots, nil
}

// ImportHolidaysToCalendar imports holidays into a specific calendar using unburdy format
func (s *CalendarService) ImportHolidaysToCalendar(calendarID uint, req entities.ImportHolidaysRequest, tenantID, userID uint) (*entities.HolidayImportResult, error) {
	// Verify calendar exists and belongs to user
	var calendar entities.Calendar
	if err := s.db.Where("id = ? AND tenant_id = ? AND user_id = ?", calendarID, tenantID, userID).First(&calendar).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("calendar not found")
		}
		return nil, fmt.Errorf("failed to query calendar: %w", err)
	}

	result := &entities.HolidayImportResult{
		ImportedYears: []string{},
		Errors:        []string{},
	}

	// Process years within the requested range
	for year := req.YearFrom; year <= req.YearTo; year++ {
		yearStr := fmt.Sprintf("%d", year)
		yearImported := false

		// Import school holidays for this year
		if schoolHolidays, exists := req.Holidays.SchoolHolidays[yearStr]; exists {
			imported, err := s.importSchoolHolidays(calendar.ID, tenantID, userID, year, schoolHolidays, req.State)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("School holidays %d: %v", year, err))
			} else {
				result.SchoolHolidays += imported
				result.TotalImported += imported
				yearImported = true
			}
		}

		// Import public holidays for this year
		if publicHolidays, exists := req.Holidays.PublicHolidays[yearStr]; exists {
			imported, err := s.importPublicHolidays(calendar.ID, tenantID, userID, year, publicHolidays, req.State)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Public holidays %d: %v", year, err))
			} else {
				result.PublicHolidays += imported
				result.TotalImported += imported
				yearImported = true
			}
		}

		if yearImported {
			result.ImportedYears = append(result.ImportedYears, yearStr)
		}
	}

	return result, nil
}

// importSchoolHolidays imports school holidays for a specific year
func (s *CalendarService) importSchoolHolidays(calendarID, tenantID, userID uint, year int, holidays map[string][2]string, state string) (int, error) {
	imported := 0

	for holidayName, dates := range holidays {
		// Parse start and end dates
		startDate, err := time.Parse("2006-01-02", dates[0])
		if err != nil {
			return imported, fmt.Errorf("invalid start date for %s: %w", holidayName, err)
		}

		endDate, err := time.Parse("2006-01-02", dates[1])
		if err != nil {
			return imported, fmt.Errorf("invalid end date for %s: %w", holidayName, err)
		}

		// Create calendar entry for the holiday period
		entry := entities.CalendarEntry{
			TenantID:    tenantID,
			UserID:      userID,
			CalendarID:  calendarID,
			Title:       holidayName,
			StartTime:   &startDate,
			EndTime:     &endDate,
			Type:        "school_holiday",
			Description: fmt.Sprintf("%s - School holidays in %s", holidayName, state),
			IsAllDay:    true,
			Location:    state,
		}

		if err := s.db.Create(&entry).Error; err != nil {
			return imported, fmt.Errorf("failed to create school holiday entry: %w", err)
		}

		imported++
	}

	return imported, nil
}

// importPublicHolidays imports public holidays for a specific year
func (s *CalendarService) importPublicHolidays(calendarID, tenantID, userID uint, year int, holidays map[string]string, state string) (int, error) {
	imported := 0

	for holidayName, dateStr := range holidays {
		// Parse the holiday date
		holidayDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return imported, fmt.Errorf("invalid date for %s: %w", holidayName, err)
		}

		// Create calendar entry for the public holiday
		entry := entities.CalendarEntry{
			TenantID:    tenantID,
			UserID:      userID,
			CalendarID:  calendarID,
			Title:       holidayName,
			StartTime:   &holidayDate,
			EndTime:     &holidayDate,
			Type:        "public_holiday",
			Description: fmt.Sprintf("%s - Public holiday in %s", holidayName, state),
			IsAllDay:    true,
			Location:    state,
		}

		if err := s.db.Create(&entry).Error; err != nil {
			return imported, fmt.Errorf("failed to create public holiday entry: %w", err)
		}

		imported++
	}

	return imported, nil
}
