package services

import (
	"encoding/json"
	"testing"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDBContractProvider(t *testing.T) (*gorm.DB, *DBContractProvider) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.TemplateContract{})
	require.NoError(t, err)

	provider := NewDBContractProvider(db)
	return db, provider
}

func TestDBContractProvider_GetContract(t *testing.T) {
	db, provider := setupDBContractProvider(t)

	// Create a test contract
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"AppName":       map[string]interface{}{"type": "string"},
			"RecipientName": map[string]interface{}{"type": "string"},
		},
		"required": []string{"AppName", "RecipientName"},
	}
	schemaJSON, _ := json.Marshal(schema)

	contract := entities.TemplateContract{
		Module:            "email",
		TemplateKey:       "password_reset",
		Description:       "Password reset template",
		SupportedChannels: []byte(`["EMAIL"]`),
		VariableSchema:    schemaJSON,
	}

	err := db.Create(&contract).Error
	require.NoError(t, err)

	// Test GetContract
	retrievedSchema, err := provider.GetContract("password_reset")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSchema)
	assert.Equal(t, "object", retrievedSchema["type"])

	properties := retrievedSchema["properties"].(map[string]interface{})
	assert.Contains(t, properties, "AppName")
	assert.Contains(t, properties, "RecipientName")
}

func TestDBContractProvider_GetContract_NotFound(t *testing.T) {
	_, provider := setupDBContractProvider(t)

	_, err := provider.GetContract("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contract not found")
}

func TestDBContractProvider_GetAvailableContracts(t *testing.T) {
	db, provider := setupDBContractProvider(t)

	// Create test contracts
	contracts := []entities.TemplateContract{
		{
			Module:            "email",
			TemplateKey:       "welcome",
			SupportedChannels: []byte(`["EMAIL"]`),
			VariableSchema:    []byte(`{"type": "object"}`),
		},
		{
			Module:            "email",
			TemplateKey:       "password_reset",
			SupportedChannels: []byte(`["EMAIL"]`),
			VariableSchema:    []byte(`{"type": "object"}`),
		},
		{
			Module:            "booking",
			TemplateKey:       "booking_confirmation",
			SupportedChannels: []byte(`["EMAIL"]`),
			VariableSchema:    []byte(`{"type": "object"}`),
		},
	}

	for _, contract := range contracts {
		err := db.Create(&contract).Error
		require.NoError(t, err)
	}

	// Test GetAvailableContracts
	availableContracts, err := provider.GetAvailableContracts()
	assert.NoError(t, err)
	assert.Len(t, availableContracts, 3)
	assert.Contains(t, availableContracts, "welcome")
	assert.Contains(t, availableContracts, "password_reset")
	assert.Contains(t, availableContracts, "booking_confirmation")
}

func TestDBContractProvider_GetSampleData(t *testing.T) {
	db, provider := setupDBContractProvider(t)

	// Create a test contract with sample data
	sampleData := map[string]interface{}{
		"AppName":       "Test App",
		"RecipientName": "John Doe",
		"ResetURL":      "https://example.com/reset",
	}
	sampleDataJSON, _ := json.Marshal(sampleData)

	contract := entities.TemplateContract{
		Module:            "email",
		TemplateKey:       "password_reset",
		Description:       "Password reset template",
		SupportedChannels: []byte(`["EMAIL"]`),
		VariableSchema:    []byte(`{"type": "object"}`),
		DefaultSampleData: sampleDataJSON,
	}

	err := db.Create(&contract).Error
	require.NoError(t, err)

	// Test GetSampleData
	retrievedSampleData, err := provider.GetSampleData("password_reset")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSampleData)
	assert.Equal(t, "Test App", retrievedSampleData["AppName"])
	assert.Equal(t, "John Doe", retrievedSampleData["RecipientName"])
	assert.Equal(t, "https://example.com/reset", retrievedSampleData["ResetURL"])
}

func TestDBContractProvider_GetSampleData_EmptyData(t *testing.T) {
	db, provider := setupDBContractProvider(t)

	// Create a test contract with empty sample data
	contract := entities.TemplateContract{
		Module:            "email",
		TemplateKey:       "welcome",
		Description:       "Welcome template",
		SupportedChannels: []byte(`["EMAIL"]`),
		VariableSchema:    []byte(`{"type": "object"}`),
		DefaultSampleData: []byte(`{}`),
	}

	err := db.Create(&contract).Error
	require.NoError(t, err)

	// Test GetSampleData
	retrievedSampleData, err := provider.GetSampleData("welcome")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedSampleData)
	assert.Empty(t, retrievedSampleData)
}

func TestDBContractProvider_GetSampleData_NotFound(t *testing.T) {
	_, provider := setupDBContractProvider(t)

	_, err := provider.GetSampleData("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contract not found")
}