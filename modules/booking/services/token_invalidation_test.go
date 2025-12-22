package services

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBForInvalidation() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&entities.BookingTemplate{}, &entities.TokenUsage{})
	if err != nil {
		return nil, err
	}

	// Create token_usage table manually for testing
	db.Exec(`CREATE TABLE IF NOT EXISTS booking_token_usage (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		token_id TEXT NOT NULL UNIQUE,
		tenant_id INTEGER NOT NULL,
		template_id INTEGER NOT NULL,
		client_id INTEGER NOT NULL,
		use_count INTEGER NOT NULL DEFAULT 0,
		max_use_count INTEGER NOT NULL DEFAULT 0,
		last_used_at DATETIME,
		expires_at DATETIME
	)`)

	return db, nil
}

func TestInvalidateOneTimeToken(t *testing.T) {
	db, err := setupTestDBForInvalidation()
	require.NoError(t, err)

	// Create a test template
	template := &entities.BookingTemplate{
		UserID:     1,
		TenantID:   1,
		CalendarID: 1,
		Name:       "Test Template",
		Duration:   30,
		Timezone:   "UTC",
	}
	err = db.Create(template).Error
	require.NoError(t, err)

	service := NewBookingLinkService(db, "test-secret-key-32-chars-long!!")

	tests := []struct {
		name              string
		purpose           entities.TokenPurpose
		maxUseCount       int
		expectInvalidated bool
	}{
		{
			name:              "One-time booking link should be invalidated",
			purpose:           entities.OneTimeBookingLink,
			maxUseCount:       0,
			expectInvalidated: true,
		},
		{
			name:              "Permanent link with max_use=1 should be invalidated",
			purpose:           entities.TimedBookingLink,
			maxUseCount:       1,
			expectInvalidated: true,
		},
		{
			name:              "Permanent link with max_use=5 should NOT be invalidated",
			purpose:           entities.TimedBookingLink,
			maxUseCount:       5,
			expectInvalidated: false,
		},
		{
			name:              "Permanent unlimited link should NOT be invalidated",
			purpose:           entities.TimedBookingLink,
			maxUseCount:       0,
			expectInvalidated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a token
			token, err := service.GenerateBookingLinkWithOptions(
				template.ID,
				123, // clientID
				1,   // tenantID
				tt.purpose,
				tt.maxUseCount,
				0, // validityDays
			)
			require.NoError(t, err)

			// Validate it first to ensure it works
			claims, err := service.ValidateBookingLink(token)
			require.NoError(t, err)
			assert.Equal(t, tt.purpose, claims.Purpose)
			assert.Equal(t, tt.maxUseCount, claims.MaxUseCount)

			// Call InvalidateOneTimeToken
			err = service.InvalidateOneTimeToken(token, claims)
			require.NoError(t, err)

			// Check if the token usage record was created/updated
			var count int64
			db.Table("booking_token_usage").Count(&count)

			if tt.expectInvalidated {
				// Should have created a usage record
				assert.Greater(t, count, int64(0), "Expected usage record to be created")

				// Verify the usage count equals max
				var usage entities.TokenUsage
				err = db.Table("booking_token_usage").Where("client_id = ?", 123).Last(&usage).Error
				require.NoError(t, err)

				// For one-time links, max should be 1
				expectedMax := tt.maxUseCount
				if expectedMax == 0 {
					expectedMax = 1
				}
				assert.Equal(t, expectedMax, usage.UseCount, "Use count should equal max use count")
				assert.Equal(t, expectedMax, usage.MaxUseCount)
			}
		})
	}
}

func TestInvalidateOneTimeTokenPreventsReuse(t *testing.T) {
	db, err := setupTestDBForInvalidation()
	require.NoError(t, err)

	// Create a test template
	template := &entities.BookingTemplate{
		UserID:     1,
		TenantID:   1,
		CalendarID: 1,
		Name:       "Test Template",
		Duration:   30,
		Timezone:   "UTC",
	}
	err = db.Create(template).Error
	require.NoError(t, err)

	service := NewBookingLinkService(db, "test-secret-key-32-chars-long!!")

	// Generate a one-time token
	token, err := service.GenerateBookingLinkWithOptions(
		template.ID,
		456, // clientID
		1,   // tenantID
		entities.OneTimeBookingLink,
		1, // maxUseCount
		0, // validityDays
	)
	require.NoError(t, err)

	// Validate it - should work
	claims, err := service.ValidateBookingLink(token)
	require.NoError(t, err)

	// Simulate the middleware using it (first time)
	tokenID := generateTokenIDForTest(token)
	now := time.Now()
	usage := entities.TokenUsage{
		TokenID:     tokenID,
		TenantID:    claims.TenantID,
		TemplateID:  claims.TemplateID,
		ClientID:    claims.ClientID,
		UseCount:    1, // First use
		MaxUseCount: 1,
		LastUsedAt:  &now,
	}
	err = db.Create(&usage).Error
	require.NoError(t, err)

	// Now invalidate it (simulating after booking)
	err = service.InvalidateOneTimeToken(token, claims)
	require.NoError(t, err)

	// Check that the usage count is now at max
	var updatedUsage entities.TokenUsage
	err = db.Where("token_id = ?", tokenID).First(&updatedUsage).Error
	require.NoError(t, err)
	assert.Equal(t, 1, updatedUsage.UseCount)
	assert.Equal(t, 1, updatedUsage.MaxUseCount)

	// Verify it's marked as exhausted
	assert.True(t, updatedUsage.HasReachedLimit())
	assert.False(t, updatedUsage.CanBeUsed())
}

// Helper function to generate token ID for testing
func generateTokenIDForTest(token string) string {
	// This should match the implementation in middleware
	return fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
}
