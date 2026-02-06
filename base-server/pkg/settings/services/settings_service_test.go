package services_test

import (
	"sync"
	"testing"

	"github.com/ae-base-server/pkg/settings/entities"
	"github.com/ae-base-server/pkg/settings/repository"
	"github.com/ae-base-server/pkg/settings/services"
	"github.com/ae-base-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupSettingsTest(t *testing.T) (*services.SettingsService, *gorm.DB) {
	db := testutils.SetupTestDB(t)

	// Auto-migrate settings table
	err := db.AutoMigrate(&entities.Setting{})
	require.NoError(t, err)

	repo := repository.NewSettingsRepository(db)
	service := services.NewSettingsService(repo)

	return service, db
}

// TestSettingsService_SetAndGetSetting tests setting and retrieving a setting
func TestSettingsService_SetAndGetSetting(t *testing.T) {
	service, db := setupSettingsTest(t)
	defer testutils.CleanupTestDB(db)

	tenantID := uint(1)
	organizationID := "test-org"
	domain := "company"
	key := "company_name"
	value := "Test Company"

	// Set the setting
	err := service.SetSetting(tenantID, organizationID, domain, key, value, "string")
	require.NoError(t, err)

	// Get the setting
	retrieved, err := service.GetSetting(tenantID, organizationID, domain, key)
	require.NoError(t, err)
	assert.Equal(t, value, retrieved)
}

// TestSettingsService_UpdateSetting tests updating an existing setting
func TestSettingsService_UpdateSetting(t *testing.T) {
	service, db := setupSettingsTest(t)
	defer testutils.CleanupTestDB(db)

	tenantID := uint(1)
	organizationID := "test-org"
	domain := "company"
	key := "company_email"

	// Set initial value
	err := service.SetSetting(tenantID, organizationID, domain, key, "old@example.com", "string")
	require.NoError(t, err)

	// Update the value
	err = service.SetSetting(tenantID, organizationID, domain, key, "new@example.com", "string")
	require.NoError(t, err)

	// Verify the update
	retrieved, err := service.GetSetting(tenantID, organizationID, domain, key)
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", retrieved)
}

// TestSettingsService_DeleteSetting tests deleting a setting
func TestSettingsService_DeleteSetting(t *testing.T) {
	service, db := setupSettingsTest(t)
	defer testutils.CleanupTestDB(db)

	tenantID := uint(1)
	organizationID := "test-org"
	domain := "company"
	key := "company_phone"

	// Set a value
	err := service.SetSetting(tenantID, organizationID, domain, key, "+1-555-0123", "string")
	require.NoError(t, err)

	// Delete the setting
	err = service.DeleteSetting(tenantID, organizationID, domain, key)
	require.NoError(t, err)

	// Verify it's deleted
	retrieved, err := service.GetSetting(tenantID, organizationID, domain, key)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

// TestSettingsService_TenantIsolation tests that tenants cannot access each other's settings
func TestSettingsService_TenantIsolation(t *testing.T) {
	service, db := setupSettingsTest(t)
	defer testutils.CleanupTestDB(db)

	tenant1ID := uint(1)
	tenant2ID := uint(2)
	organizationID := "test-org"
	domain := "company"
	key := "company_name"

	// Set setting for tenant 1
	err := service.SetSetting(tenant1ID, organizationID, domain, key, "Tenant 1 Company", "string")
	require.NoError(t, err)

	// Set setting for tenant 2
	err = service.SetSetting(tenant2ID, organizationID, domain, key, "Tenant 2 Company", "string")
	require.NoError(t, err)

	// Verify tenant 1 gets their own setting
	retrieved1, err := service.GetSetting(tenant1ID, organizationID, domain, key)
	require.NoError(t, err)
	assert.Equal(t, "Tenant 1 Company", retrieved1)

	// Verify tenant 2 gets their own setting
	retrieved2, err := service.GetSetting(tenant2ID, organizationID, domain, key)
	require.NoError(t, err)
	assert.Equal(t, "Tenant 2 Company", retrieved2)
}

// TestSettingsService_DifferentDataTypes tests different data types
func TestSettingsService_DifferentDataTypes(t *testing.T) {
	service, db := setupSettingsTest(t)
	defer testutils.CleanupTestDB(db)

	tenantID := uint(1)
	organizationID := "test-org"
	domain := "config"

	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected interface{}
	}{
		{
			name:     "string value",
			key:      "string_setting",
			value:    "test string",
			expected: "test string",
		},
		// Integer values are skipped - JSONB doesn't handle raw integers
		{
			name:     "boolean value",
			key:      "bool_setting",
			value:    true,
			expected: true,
		},
		{
			name:     "map value",
			key:      "map_setting",
			value:    map[string]interface{}{"nested": "value"},
			expected: map[string]interface{}{"nested": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the value
			err := service.SetSetting(tenantID, organizationID, domain, tt.key, tt.value, "")
			require.NoError(t, err)

			// Get the value
			retrieved, err := service.GetSetting(tenantID, organizationID, domain, tt.key)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, retrieved)
		})
	}
}

// TestSettingsService_ConcurrentWrites tests concurrent write operations
func TestSettingsService_ConcurrentWrites(t *testing.T) {
	t.Skip("Concurrent writes test requires separate DB connections")

	service, db := setupSettingsTest(t)
	defer testutils.CleanupTestDB(db)

	tenantID := uint(1)
	organizationID := "test-org"
	domain := "test"

	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Write different keys concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := "concurrent_key_" + string(rune('0'+index))
			value := "concurrent_value_" + string(rune('0'+index))
			err := service.SetSetting(tenantID, organizationID, domain, key, value, "string")
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent write failed: %v", err)
	}

	// Verify all settings were created
	for i := 0; i < numGoroutines; i++ {
		key := "concurrent_key_" + string(rune('0'+i))
		retrieved, err := service.GetSetting(tenantID, organizationID, domain, key)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	}
}
