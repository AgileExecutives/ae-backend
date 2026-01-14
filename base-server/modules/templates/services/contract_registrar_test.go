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

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.TemplateContract{})
	require.NoError(t, err)

	return db
}

func TestContractRegistrar_RegisterContract(t *testing.T) {
	db := setupTestDB(t)
	registrar := NewContractRegistrar(db)

	// Create a contract registration
	registration := entities.ContractRegistration{
		ModuleName:  "email",
		TemplateKey: "welcome",
		Name:        "Welcome Template",
		Description: "Welcome email template",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"FirstName": map[string]interface{}{"type": "string"},
				"LastName":  map[string]interface{}{"type": "string"},
			},
		},
		Version: "1.0.0",
	}

	// Register the contract
	tenantID := uint(1)
	err := registrar.RegisterContract(tenantID, registration)
	assert.NoError(t, err)

	// Verify the contract was created
	var contract entities.TemplateContract
	err = db.Where("module = ? AND template_key = ?", "email", "welcome").First(&contract).Error
	assert.NoError(t, err)
	assert.Equal(t, "email", contract.Module)
	assert.Equal(t, "welcome", contract.TemplateKey)
}

func TestContractRegistrar_RegisterContractWithSampleData(t *testing.T) {
	db := setupTestDB(t)
	registrar := NewContractRegistrar(db)

	registration := entities.ContractRegistration{
		ModuleName:  "email",
		TemplateKey: "password_reset",
		Name:        "Password Reset Template",
		Description: "Password reset email template",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"AppName":       map[string]interface{}{"type": "string"},
				"RecipientName": map[string]interface{}{"type": "string"},
			},
		},
		Version: "1.0.0",
	}

	sampleData := map[string]interface{}{
		"AppName":       "Test App",
		"RecipientName": "John Doe",
		"ResetURL":      "https://test.com/reset",
	}

	// Register contract with sample data
	tenantID := uint(1)
	err := registrar.RegisterContractWithSampleData(tenantID, registration, sampleData)
	assert.NoError(t, err)

	// Verify the contract and sample data were stored
	var contract entities.TemplateContract
	err = db.Where("module = ? AND template_key = ?", "email", "password_reset").First(&contract).Error
	assert.NoError(t, err)

	var storedSampleData map[string]interface{}
	err = json.Unmarshal(contract.DefaultSampleData, &storedSampleData)
	assert.NoError(t, err)
	assert.Equal(t, "Test App", storedSampleData["AppName"])
	assert.Equal(t, "John Doe", storedSampleData["RecipientName"])
}

func TestContractRegistrar_UpdateExistingContract(t *testing.T) {
	db := setupTestDB(t)
	registrar := NewContractRegistrar(db)

	// Create initial contract
	initialRegistration := entities.ContractRegistration{
		ModuleName:  "email",
		TemplateKey: "welcome",
		Name:        "Welcome Template",
		Description: "Initial description",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"FirstName": map[string]interface{}{"type": "string"},
			},
		},
		Version: "1.0.0",
	}

	tenantID := uint(1)
	err := registrar.RegisterContract(tenantID, initialRegistration)
	assert.NoError(t, err)

	// Update the contract
	updatedRegistration := entities.ContractRegistration{
		ModuleName:  "email",
		TemplateKey: "welcome",
		Name:        "Welcome Template",
		Description: "Updated description",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"FirstName": map[string]interface{}{"type": "string"},
				"LastName":  map[string]interface{}{"type": "string"},
			},
		},
		Version: "1.1.0",
	}

	err = registrar.RegisterContract(tenantID, updatedRegistration)
	assert.NoError(t, err)

	// Verify the contract was updated
	var contract entities.TemplateContract
	err = db.Where("module = ? AND template_key = ?", "email", "welcome").First(&contract).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated description", contract.Description)

	// Verify schema was updated
	var schema map[string]interface{}
	err = json.Unmarshal(contract.VariableSchema, &schema)
	assert.NoError(t, err)
	properties := schema["properties"].(map[string]interface{})
	assert.Contains(t, properties, "FirstName")
	assert.Contains(t, properties, "LastName")
}

func TestContractRegistrar_GetContractByKey(t *testing.T) {
	db := setupTestDB(t)
	registrar := NewContractRegistrar(db)

	// Register a contract
	registration := entities.ContractRegistration{
		ModuleName:  "booking",
		TemplateKey: "booking_confirmation",
		Name:        "Booking Confirmation Template",
		Description: "Booking confirmation email template",
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"AppName":   map[string]interface{}{"type": "string"},
				"BookingId": map[string]interface{}{"type": "string"},
			},
		},
		Version: "1.0.0",
	}

	tenantID := uint(1)
	err := registrar.RegisterContract(tenantID, registration)
	assert.NoError(t, err)

	// Retrieve the contract
	contract, err := registrar.GetContractByKey("booking", "booking_confirmation")
	assert.NoError(t, err)
	assert.NotNil(t, contract)
	assert.Equal(t, "booking", contract.Module)
	assert.Equal(t, "booking_confirmation", contract.TemplateKey)

	// Test non-existent contract
	contract, err = registrar.GetContractByKey("nonexistent", "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, contract)
}

func TestContractRegistrar_MultipleContracts(t *testing.T) {
	db := setupTestDB(t)
	registrar := NewContractRegistrar(db)

	// Register multiple contracts
	contracts := []entities.ContractRegistration{
		{
			ModuleName:  "email",
			TemplateKey: "welcome",
			Name:        "Welcome Template",
			Schema:      map[string]interface{}{"type": "object"},
			Version:     "1.0.0",
		},
		{
			ModuleName:  "booking",
			TemplateKey: "booking_confirmation",
			Name:        "Booking Confirmation Template",
			Schema:      map[string]interface{}{"type": "object"},
			Version:     "1.0.0",
		},
		{
			ModuleName:  "base",
			TemplateKey: "invoice",
			Name:        "Invoice Template",
			Schema:      map[string]interface{}{"type": "object"},
			Version:     "1.0.0",
		},
	}

	tenantID := uint(1)
	for _, contract := range contracts {
		err := registrar.RegisterContract(tenantID, contract)
		assert.NoError(t, err)
	}

	// Verify contracts can be retrieved individually
	emailContract, err := registrar.GetContractByKey("email", "welcome")
	assert.NoError(t, err)
	assert.Equal(t, "email", emailContract.Module)

	bookingContract, err := registrar.GetContractByKey("booking", "booking_confirmation")
	assert.NoError(t, err)
	assert.Equal(t, "booking", bookingContract.Module)

	baseContract, err := registrar.GetContractByKey("base", "invoice")
	assert.NoError(t, err)
	assert.Equal(t, "base", baseContract.Module)

	// Verify we can count contracts in database
	var count int64
	db.Model(&entities.TemplateContract{}).Count(&count)
	assert.Equal(t, int64(3), count)
}
