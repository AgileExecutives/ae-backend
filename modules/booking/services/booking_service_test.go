package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/unburdy/booking-module/entities"
	"github.com/unburdy/booking-module/services"
)

// setupBookingTestDB creates an in-memory SQLite database for booking tests
func setupBookingTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the BookingTemplate schema
	err = db.AutoMigrate(&entities.BookingTemplate{})
	require.NoError(t, err)

	return db
}

// TestBookingService_CreateConfiguration tests creating a booking configuration
func TestBookingService_CreateConfiguration(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	tenantID := uint(1)
	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Test Configuration",
		Description:       "Test booking configuration",
		SlotDuration:      30,
		BufferTime:        5,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalWeekly},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Monday: []entities.TimeRange{
				{Start: "09:00", End: "17:00"},
			},
		},
		AdvanceBookingDays: 30,
		MinNoticeHours:     24,
		Timezone:           "Europe/Berlin",
	}

	config, err := service.CreateConfiguration(req, tenantID)

	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "Test Configuration", config.Name)
	assert.Equal(t, 30, config.SlotDuration)
	assert.Equal(t, "Europe/Berlin", config.Timezone)
	assert.Equal(t, tenantID, config.TenantID)
}

// TestBookingService_GetConfiguration tests retrieving a configuration
func TestBookingService_GetConfiguration(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	tenantID := uint(1)
	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Get Test",
		SlotDuration:      60,
		BufferTime:        0,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Tuesday: []entities.TimeRange{
				{Start: "10:00", End: "16:00"},
			},
		},
		AdvanceBookingDays: 60,
		MinNoticeHours:     12,
		Timezone:           "UTC",
	}

	created, err := service.CreateConfiguration(req, tenantID)
	require.NoError(t, err)

	// Retrieve the configuration
	retrieved, err := service.GetConfiguration(created.ID, tenantID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, "Get Test", retrieved.Name)
}

// TestBookingService_UpdateConfiguration tests updating a configuration
func TestBookingService_UpdateConfiguration(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	tenantID := uint(1)
	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Original Name",
		Description:       "Original description",
		SlotDuration:      30,
		BufferTime:        5,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalWeekly},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Wednesday: []entities.TimeRange{
				{Start: "09:00", End: "17:00"},
			},
		},
		AdvanceBookingDays: 30,
		MinNoticeHours:     24,
		Timezone:           "Europe/Berlin",
	}

	created, err := service.CreateConfiguration(req, tenantID)
	require.NoError(t, err)

	// Update the configuration
	newName := "Updated Name"
	newDescription := "Updated description"
	newSlotDuration := 45
	updateReq := entities.UpdateBookingTemplateRequest{
		Name:         &newName,
		Description:  &newDescription,
		SlotDuration: &newSlotDuration,
	}

	updated, err := service.UpdateConfiguration(created.ID, tenantID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, newDescription, updated.Description)
	assert.Equal(t, newSlotDuration, updated.SlotDuration)
}

// TestBookingService_DeleteConfiguration tests deleting a configuration
func TestBookingService_DeleteConfiguration(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	tenantID := uint(1)
	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Delete Test",
		SlotDuration:      30,
		BufferTime:        0,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Thursday: []entities.TimeRange{
				{Start: "08:00", End: "12:00"},
			},
		},
		AdvanceBookingDays: 30,
		MinNoticeHours:     24,
		Timezone:           "UTC",
	}

	created, err := service.CreateConfiguration(req, tenantID)
	require.NoError(t, err)

	// Delete the configuration
	err = service.DeleteConfiguration(created.ID, tenantID)
	require.NoError(t, err)

	// Try to get the deleted configuration
	_, err = service.GetConfiguration(created.ID, tenantID)
	assert.Error(t, err)
}

// TestBookingService_GetAllConfigurations tests retrieving all configurations with pagination
func TestBookingService_GetAllConfigurations(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	tenantID := uint(1)

	// Create multiple configurations
	for i := 1; i <= 5; i++ {
		req := entities.CreateBookingTemplateRequest{
			UserID:            uint(i),
			CalendarID:        uint(i),
			Name:              "Config " + string(rune('A'+i-1)),
			SlotDuration:      30,
			BufferTime:        0,
			MaxSeriesBookings: 1,
			AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
			NumberOfIntervals: 1,
			WeeklyAvailability: entities.WeeklyAvailability{
				Friday: []entities.TimeRange{
					{Start: "09:00", End: "17:00"},
				},
			},
			AdvanceBookingDays: 30,
			MinNoticeHours:     24,
			Timezone:           "UTC",
		}
		_, err := service.CreateConfiguration(req, tenantID)
		require.NoError(t, err)
	}

	// Get all configurations with pagination
	configs, total, err := service.GetAllConfigurations(tenantID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, configs, 5)
}

// TestBookingService_GetConfigurationsByUser tests retrieving configurations by user
func TestBookingService_GetConfigurationsByUser(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	tenantID := uint(1)
	userID := uint(42)

	// Create configurations for different users
	for i := 1; i <= 3; i++ {
		req := entities.CreateBookingTemplateRequest{
			UserID:            userID,
			CalendarID:        uint(i),
			Name:              "User Config " + string(rune('A'+i-1)),
			SlotDuration:      30,
			BufferTime:        0,
			MaxSeriesBookings: 1,
			AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
			NumberOfIntervals: 1,
			WeeklyAvailability: entities.WeeklyAvailability{
				Monday: []entities.TimeRange{
					{Start: "09:00", End: "17:00"},
				},
			},
			AdvanceBookingDays: 30,
			MinNoticeHours:     24,
			Timezone:           "UTC",
		}
		_, err := service.CreateConfiguration(req, tenantID)
		require.NoError(t, err)
	}

	// Create a configuration for a different user
	otherReq := entities.CreateBookingTemplateRequest{
		UserID:            99,
		CalendarID:        99,
		Name:              "Other User Config",
		SlotDuration:      30,
		BufferTime:        0,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Monday: []entities.TimeRange{
				{Start: "09:00", End: "17:00"},
			},
		},
		AdvanceBookingDays: 30,
		MinNoticeHours:     24,
		Timezone:           "UTC",
	}
	_, err := service.CreateConfiguration(otherReq, tenantID)
	require.NoError(t, err)

	// Get configurations for specific user
	configs, err := service.GetConfigurationsByUser(userID, tenantID)
	require.NoError(t, err)
	assert.Len(t, configs, 3)
	for _, config := range configs {
		assert.Equal(t, userID, config.UserID)
	}
}

// TestBookingService_TenantIsolation tests that tenants cannot access each other's configurations
func TestBookingService_TenantIsolation(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Tenant 1 Config",
		SlotDuration:      30,
		BufferTime:        0,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Monday: []entities.TimeRange{
				{Start: "09:00", End: "17:00"},
			},
		},
		AdvanceBookingDays: 30,
		MinNoticeHours:     24,
		Timezone:           "UTC",
	}

	// Create configuration for tenant 1
	config1, err := service.CreateConfiguration(req, 1)
	require.NoError(t, err)

	// Create configuration for tenant 2
	config2, err := service.CreateConfiguration(req, 2)
	require.NoError(t, err)

	// Tenant 1 should not see tenant 2's configuration
	_, err = service.GetConfiguration(config2.ID, 1)
	assert.Error(t, err)

	// Tenant 2 should not see tenant 1's configuration
	_, err = service.GetConfiguration(config1.ID, 2)
	assert.Error(t, err)
}

// TestBookingService_WeeklyAvailability tests configuration with full week schedule
func TestBookingService_WeeklyAvailability(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	tenantID := uint(1)
	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Full Week Schedule",
		SlotDuration:      60,
		BufferTime:        15,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalWeekly},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Monday:    []entities.TimeRange{{Start: "09:00", End: "12:00"}},
			Tuesday:   []entities.TimeRange{{Start: "09:00", End: "12:00"}, {Start: "14:00", End: "17:00"}},
			Wednesday: []entities.TimeRange{{Start: "10:00", End: "16:00"}},
			Thursday:  []entities.TimeRange{{Start: "09:00", End: "17:00"}},
			Friday:    []entities.TimeRange{{Start: "09:00", End: "15:00"}},
		},
		AdvanceBookingDays: 90,
		MinNoticeHours:     48,
		Timezone:           "America/New_York",
	}

	config, err := service.CreateConfiguration(req, tenantID)

	require.NoError(t, err)
	assert.NotNil(t, config.WeeklyAvailability.Monday)
	assert.Len(t, config.WeeklyAvailability.Tuesday, 2)
	assert.Equal(t, "America/New_York", config.Timezone)
}

// TestBookingService_AdvancedSettings tests configuration with advanced settings
func TestBookingService_AdvancedSettings(t *testing.T) {
	db := setupBookingTestDB(t)
	service := services.NewBookingService(db)

	tenantID := uint(1)
	maxBookings := 5
	allowBackToBack := true

	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Advanced Config",
		SlotDuration:      30,
		BufferTime:        10,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Monday: []entities.TimeRange{{Start: "09:00", End: "17:00"}},
		},
		AdvanceBookingDays: 30,
		MinNoticeHours:     24,
		Timezone:           "Europe/Berlin",
		MaxBookingsPerDay:  &maxBookings,
		AllowBackToBack:    &allowBackToBack,
		BlockDates: []entities.DateRange{
			{Start: "2026-12-24", End: "2026-12-26"},
			{Start: "2027-01-01", End: "2027-01-01"},
		},
		AllowedStartMinutes: []int{0, 15, 30, 45},
	}

	config, err := service.CreateConfiguration(req, tenantID)

	require.NoError(t, err)
	assert.NotNil(t, config.MaxBookingsPerDay)
	assert.Equal(t, maxBookings, *config.MaxBookingsPerDay)
	assert.NotNil(t, config.AllowBackToBack)
	assert.True(t, *config.AllowBackToBack)
	assert.Len(t, config.BlockDates, 2)
	assert.Len(t, config.AllowedStartMinutes, 4)
}

// BenchmarkBookingService_CreateConfiguration benchmarks configuration creation
func BenchmarkBookingService_CreateConfiguration(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&entities.BookingTemplate{})
	service := services.NewBookingService(db)

	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Benchmark Config",
		SlotDuration:      30,
		BufferTime:        0,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Monday: []entities.TimeRange{{Start: "09:00", End: "17:00"}},
		},
		AdvanceBookingDays: 30,
		MinNoticeHours:     24,
		Timezone:           "UTC",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateConfiguration(req, 1)
	}
}

// BenchmarkBookingService_GetConfiguration benchmarks configuration retrieval
func BenchmarkBookingService_GetConfiguration(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&entities.BookingTemplate{})
	service := services.NewBookingService(db)

	req := entities.CreateBookingTemplateRequest{
		UserID:            1,
		CalendarID:        1,
		Name:              "Benchmark Config",
		SlotDuration:      30,
		BufferTime:        0,
		MaxSeriesBookings: 1,
		AllowedIntervals:  []entities.IntervalType{entities.IntervalNone},
		NumberOfIntervals: 1,
		WeeklyAvailability: entities.WeeklyAvailability{
			Monday: []entities.TimeRange{{Start: "09:00", End: "17:00"}},
		},
		AdvanceBookingDays: 30,
		MinNoticeHours:     24,
		Timezone:           "UTC",
	}

	config, _ := service.CreateConfiguration(req, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetConfiguration(config.ID, 1)
	}
}
