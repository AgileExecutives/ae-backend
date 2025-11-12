package services

import (
	"errors"
	"fmt"

	"github.com/unburdy/booking-module/entities"
	"gorm.io/gorm"
)

type BookingService struct {
	db *gorm.DB
}

func NewBookingService(db *gorm.DB) *BookingService {
	return &BookingService{db: db}
}

// CreateConfiguration creates a new booking configuration
func (s *BookingService) CreateConfiguration(req entities.CreateBookingTemplateRequest, tenantID uint) (*entities.BookingTemplate, error) {
	config := &entities.BookingTemplate{
		UserID:             req.UserID,
		CalendarID:         req.CalendarID,
		TenantID:           tenantID,
		Name:               req.Name,
		Description:        req.Description,
		SlotDuration:       req.SlotDuration,
		BufferTime:         req.BufferTime,
		MaxSeriesBookings:  req.MaxSeriesBookings,
		AllowedIntervals:   req.AllowedIntervals,
		NumberOfIntervals:  req.NumberOfIntervals,
		WeeklyAvailability: req.WeeklyAvailability,
		AdvanceBookingDays: req.AdvanceBookingDays,
		MinNoticeHours:     req.MinNoticeHours,
		Timezone:           req.Timezone,
		MaxBookingsPerDay:  req.MaxBookingsPerDay,
		AllowBackToBack:    req.AllowBackToBack,
		BlockDates:         req.BlockDates,
	}

	if err := s.db.Create(config).Error; err != nil {
		return nil, fmt.Errorf("failed to create booking configuration: %w", err)
	}

	return config, nil
}

// GetConfiguration retrieves a booking configuration by ID
func (s *BookingService) GetConfiguration(id uint, tenantID uint) (*entities.BookingTemplate, error) {
	var config entities.BookingTemplate
	
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("booking configuration not found")
		}
		return nil, fmt.Errorf("failed to retrieve booking configuration: %w", err)
	}

	return &config, nil
}

// GetAllConfigurations retrieves all booking configurations for a tenant
func (s *BookingService) GetAllConfigurations(tenantID uint, page, limit int) ([]entities.BookingTemplate, int64, error) {
	var configs []entities.BookingTemplate
	var total int64

	query := s.db.Model(&entities.BookingTemplate{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count booking configurations: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve booking configurations: %w", err)
	}

	return configs, total, nil
}

// GetConfigurationsByUser retrieves all booking configurations for a specific user
func (s *BookingService) GetConfigurationsByUser(userID uint, tenantID uint) ([]entities.BookingTemplate, error) {
	var configs []entities.BookingTemplate
	
	if err := s.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve booking configurations for user: %w", err)
	}

	return configs, nil
}

// GetConfigurationsByCalendar retrieves all booking configurations for a specific calendar
func (s *BookingService) GetConfigurationsByCalendar(calendarID uint, tenantID uint) ([]entities.BookingTemplate, error) {
	var configs []entities.BookingTemplate
	
	if err := s.db.Where("calendar_id = ? AND tenant_id = ?", calendarID, tenantID).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve booking configurations for calendar: %w", err)
	}

	return configs, nil
}

// UpdateConfiguration updates an existing booking configuration
func (s *BookingService) UpdateConfiguration(id uint, tenantID uint, req entities.UpdateBookingTemplateRequest) (*entities.BookingTemplate, error) {
	var config entities.BookingTemplate
	
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("booking configuration not found")
		}
		return nil, fmt.Errorf("failed to retrieve booking configuration: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		config.Name = *req.Name
	}
	if req.Description != nil {
		config.Description = *req.Description
	}
	if req.SlotDuration != nil {
		config.SlotDuration = *req.SlotDuration
	}
	if req.BufferTime != nil {
		config.BufferTime = *req.BufferTime
	}
	if req.MaxSeriesBookings != nil {
		config.MaxSeriesBookings = *req.MaxSeriesBookings
	}
	if req.AllowedIntervals != nil {
		config.AllowedIntervals = req.AllowedIntervals
	}
	if req.NumberOfIntervals != nil {
		config.NumberOfIntervals = *req.NumberOfIntervals
	}
	if req.WeeklyAvailability != nil {
		config.WeeklyAvailability = *req.WeeklyAvailability
	}
	if req.AdvanceBookingDays != nil {
		config.AdvanceBookingDays = *req.AdvanceBookingDays
	}
	if req.MinNoticeHours != nil {
		config.MinNoticeHours = *req.MinNoticeHours
	}
	if req.Timezone != nil {
		config.Timezone = *req.Timezone
	}
	if req.MaxBookingsPerDay != nil {
		config.MaxBookingsPerDay = req.MaxBookingsPerDay
	}
	if req.AllowBackToBack != nil {
		config.AllowBackToBack = req.AllowBackToBack
	}
	if req.BlockDates != nil {
		config.BlockDates = req.BlockDates
	}

	if err := s.db.Save(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to update booking configuration: %w", err)
	}

	return &config, nil
}

// DeleteConfiguration soft deletes a booking configuration
func (s *BookingService) DeleteConfiguration(id uint, tenantID uint) error {
	result := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entities.BookingTemplate{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete booking configuration: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("booking configuration not found")
	}

	return nil
}
