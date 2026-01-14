package services

import (
	"testing"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Integration tests for the complete template system

func setupIntegrationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.Template{}, &entities.TemplateContract{})
	require.NoError(t, err)

	return db
}

func TestContractRegistration_Integration(t *testing.T) {
	db := setupIntegrationTestDB(t)
	contractRegistrar := NewContractRegistrar(db)

	// Test registering contracts from different modules
	tenantID := uint(1)

	// Register email contracts (simulated)
	err := registerEmailContractsForTest(contractRegistrar, tenantID)
	assert.NoError(t, err)

	// Register base contracts (simulated)
	err = registerBaseContractsForTest(contractRegistrar, tenantID)
	assert.NoError(t, err)

	// Verify contracts were registered
	var contracts []entities.TemplateContract
	err = db.Find(&contracts).Error
	assert.NoError(t, err)
	assert.True(t, len(contracts) > 0)

	// Verify we have contracts from different modules
	moduleMap := make(map[string]bool)
	for _, contract := range contracts {
		moduleMap[contract.Module] = true
	}
	assert.True(t, moduleMap["email"])
	assert.True(t, moduleMap["base"])
}

func TestTemplateSystemEndToEnd_Integration(t *testing.T) {
	db := setupIntegrationTestDB(t)

	// Setup the complete system
	contractRegistrar := NewContractRegistrar(db)

	tenantID := uint(1)

	// 1. Register contracts
	err := registerEmailContractsForTest(contractRegistrar, tenantID)
	assert.NoError(t, err)

	// 2. Test contract retrieval
	contract, err := contractRegistrar.GetContractByKey("email", "welcome")
	assert.NoError(t, err)
	assert.NotNil(t, contract)
	assert.Equal(t, "email", contract.Module)
	assert.Equal(t, "welcome", contract.TemplateKey)

	// 3. Test getting contract information
	contractProvider := NewDBContractProvider(db)
	contracts, err := contractProvider.GetAvailableContracts()
	assert.NoError(t, err)
	assert.Contains(t, contracts, "welcome")

	// 4. Test getting sample data
	sampleData, err := contractProvider.GetSampleData("welcome")
	assert.NoError(t, err)
	assert.NotEmpty(t, sampleData)
}

func TestContractValidation_Integration(t *testing.T) {
	db := setupIntegrationTestDB(t)
	contractRegistrar := NewContractRegistrar(db)

	tenantID := uint(1)

	// Register a contract with strict validation
	registration := entities.ContractRegistration{
		ModuleName:  "email",
		TemplateKey: "strict_validation",
		Name:        "Strict Validation Template",
		Description: "Template with strict validation rules",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"email": map[string]interface{}{
					"type":   "string",
					"format": "email",
				},
				"name": map[string]interface{}{
					"type":      "string",
					"minLength": 1,
					"maxLength": 100,
				},
			},
			"required": []string{"email", "name"},
		},
		Version: "1.0.0",
	}

	sampleData := map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	}

	err := contractRegistrar.RegisterContractWithSampleData(tenantID, registration, sampleData)
	assert.NoError(t, err)

	// Test validation with DB contract provider
	contractProvider := NewDBContractProvider(db)

	// Verify the contract exists and can be retrieved
	contractSchema, err := contractProvider.GetContract("strict_validation")
	assert.NoError(t, err)
	assert.NotNil(t, contractSchema)
	assert.Equal(t, "object", contractSchema["type"])
}

// Helper functions to simulate module contract registration

func registerEmailContractsForTest(registrar *ContractRegistrar, tenantID uint) error {
	contracts := []entities.ContractRegistration{
		{
			ModuleName:  "email",
			TemplateKey: "welcome",
			Name:        "Welcome Template",
			Description: "Welcome email template",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"FirstName":        map[string]interface{}{"type": "string"},
					"LastName":         map[string]interface{}{"type": "string"},
					"OrganizationName": map[string]interface{}{"type": "string"},
				},
			},
			Version: "1.0.0",
		},
		{
			ModuleName:  "email",
			TemplateKey: "password_reset",
			Name:        "Password Reset Template",
			Description: "Password reset email template",
			Schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"AppName":       map[string]interface{}{"type": "string"},
					"RecipientName": map[string]interface{}{"type": "string"},
					"ResetURL":      map[string]interface{}{"type": "string"},
				},
			},
			Version: "1.0.0",
		},
	}

	for _, contract := range contracts {
		sampleData := map[string]interface{}{
			"FirstName":        "Test",
			"LastName":         "User",
			"OrganizationName": "Test Org",
			"AppName":          "Test App",
			"RecipientName":    "Test User",
			"ResetURL":         "https://test.com/reset",
		}
		if err := registrar.RegisterContractWithSampleData(tenantID, contract, sampleData); err != nil {
			return err
		}
	}
	return nil
}

func registerBaseContractsForTest(registrar *ContractRegistrar, tenantID uint) error {
	contract := entities.ContractRegistration{
		ModuleName:  "base",
		TemplateKey: "invoice",
		Name:        "Invoice Template",
		Description: "Invoice template",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"invoice_number": map[string]interface{}{"type": "string"},
				"invoice_date":   map[string]interface{}{"type": "string"},
			},
		},
		Version: "1.0.0",
	}

	sampleData := map[string]interface{}{
		"invoice_number": "INV-001",
		"invoice_date":   "2024-01-01",
	}

	return registrar.RegisterContractWithSampleData(tenantID, contract, sampleData)
}
