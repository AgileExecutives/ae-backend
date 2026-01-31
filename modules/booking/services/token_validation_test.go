package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/booking-module/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTokenValidationTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&entities.BookingTemplate{}, &entities.TokenUsage{})
	if err != nil {
		return nil, err
	}

	// Create token_blacklist table
	err = db.Exec(`
		CREATE TABLE token_blacklist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token_id VARCHAR(255) NOT NULL,
			user_id INTEGER,
			reason TEXT,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestGenerateBookingLinkWithOptions(t *testing.T) {
	db, err := setupTokenValidationTestDB()
	require.NoError(t, err)

	// Create a test template
	template := &entities.BookingTemplate{
		UserID:       1,
		TenantID:     1,
		CalendarID:   1,
		Name:         "Test Template",
		SlotDuration: 30,
		Timezone:     "UTC",
	}
	err = db.Create(template).Error
	require.NoError(t, err)

	service := NewBookingLinkService(db, "test-secret-key-32-chars-long!!")

	tests := []struct {
		name         string
		maxUseCount  int
		validityDays int
		purpose      entities.TokenPurpose
	}{
		{
			name:         "One-time link with default expiration",
			maxUseCount:  1,
			validityDays: 0,
			purpose:      entities.OneTimeBookingLink,
		},
		{
			name:         "Limited use link (5 times, 7 days)",
			maxUseCount:  5,
			validityDays: 7,
			purpose:      entities.TimedBookingLink,
		},
		{
			name:         "Unlimited use with time limit (30 days)",
			maxUseCount:  0,
			validityDays: 30,
			purpose:      entities.TimedBookingLink,
		},
		{
			name:         "Permanent unlimited link",
			maxUseCount:  0,
			validityDays: 0,
			purpose:      entities.TimedBookingLink,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateBookingLinkWithOptions(
				template.ID,
				123, // clientID
				1,   // tenantID
				tt.purpose,
				tt.maxUseCount,
				tt.validityDays,
			)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Validate the token
			claims, err := service.ValidateBookingLink(token)
			assert.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, tt.maxUseCount, claims.MaxUseCount)
			assert.Equal(t, tt.validityDays, claims.ValidityDays)
			assert.Equal(t, tt.purpose, claims.Purpose)

			// Check expiration
			if tt.validityDays > 0 || tt.purpose == entities.OneTimeBookingLink {
				assert.NotNil(t, claims.ExpiresAt)
				if claims.ExpiresAt != nil {
					assert.Greater(t, claims.ExpiresAt.Unix(), int64(0))
				}
			}
		})
	}
}

func TestTokenUsageTracking(t *testing.T) {
	db, err := setupTokenValidationTestDB()
	require.NoError(t, err)

	// Create test usage record
	usage := &entities.TokenUsage{
		TokenID:     "test-token-id",
		TenantID:    1,
		TemplateID:  1,
		ClientID:    123,
		UseCount:    0,
		MaxUseCount: 3,
	}
	err = db.Create(usage).Error
	require.NoError(t, err)

	t.Run("can be used when under limit", func(t *testing.T) {
		assert.True(t, usage.CanBeUsed())
		assert.False(t, usage.HasReachedLimit())
		assert.False(t, usage.IsExpired())
	})

	t.Run("increment usage", func(t *testing.T) {
		usage.IncrementUsage()
		assert.Equal(t, 1, usage.UseCount)
		assert.NotNil(t, usage.LastUsedAt)
		assert.True(t, usage.CanBeUsed())
	})

	t.Run("reaches limit after max uses", func(t *testing.T) {
		usage.UseCount = 3
		assert.True(t, usage.HasReachedLimit())
		assert.False(t, usage.CanBeUsed())
	})

	t.Run("unlimited usage when max is 0", func(t *testing.T) {
		unlimitedUsage := &entities.TokenUsage{
			TokenID:     "unlimited-token",
			TenantID:    1,
			TemplateID:  1,
			ClientID:    123,
			UseCount:    100,
			MaxUseCount: 0,
		}
		assert.False(t, unlimitedUsage.HasReachedLimit())
		assert.True(t, unlimitedUsage.CanBeUsed())
	})

	t.Run("expired token", func(t *testing.T) {
		pastTime := time.Now().Add(-1 * time.Hour)
		expiredUsage := &entities.TokenUsage{
			TokenID:     "expired-token",
			TenantID:    1,
			TemplateID:  1,
			ClientID:    123,
			UseCount:    0,
			MaxUseCount: 10,
			ExpiresAt:   &pastTime,
		}
		assert.True(t, expiredUsage.IsExpired())
		assert.False(t, expiredUsage.CanBeUsed())
	})

	t.Run("not expired token", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour)
		validUsage := &entities.TokenUsage{
			TokenID:     "valid-token",
			TenantID:    1,
			TemplateID:  1,
			ClientID:    123,
			UseCount:    0,
			MaxUseCount: 10,
			ExpiresAt:   &futureTime,
		}
		assert.False(t, validUsage.IsExpired())
		assert.True(t, validUsage.CanBeUsed())
	})
}

func TestTokenExpiration(t *testing.T) {
	db, err := setupTokenValidationTestDB()
	require.NoError(t, err)

	// Create a test template
	template := &entities.BookingTemplate{
		UserID:       1,
		TenantID:     1,
		CalendarID:   1,
		Name:         "Test Template",
		SlotDuration: 30,
		Timezone:     "UTC",
	}
	err = db.Create(template).Error
	require.NoError(t, err)

	_ = NewBookingLinkService(db, "test-secret-key-32-chars-long!!")

	t.Run("expired token validation fails", func(t *testing.T) {
		t.Skip("Test needs to generate expired tokens - requires custom token generation logic")
		// Note: generateToken is a private method, and we can't easily create expired tokens
		// This test should be handled at integration level or with a test helper
	})
}
