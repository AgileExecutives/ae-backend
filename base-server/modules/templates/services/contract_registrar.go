package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ae-base-server/modules/templates/entities"
	"gorm.io/gorm"
)

// ContractRegistrar handles registration of contracts from modules
type ContractRegistrar struct {
	db *gorm.DB
}

// NewContractRegistrar creates a new contract registrar
func NewContractRegistrar(db *gorm.DB) *ContractRegistrar {
	return &ContractRegistrar{
		db: db,
	}
}

// RegisterContract registers a contract from a module
func (r *ContractRegistrar) RegisterContract(tenantID uint, registration entities.ContractRegistration) error {
	return r.RegisterContractWithSampleData(tenantID, registration, map[string]interface{}{})
}

// RegisterContractWithSampleData registers a contract with sample data from a module
func (r *ContractRegistrar) RegisterContractWithSampleData(tenantID uint, registration entities.ContractRegistration, sampleData map[string]interface{}) error {
	// Check if contract already exists
	var existingContract entities.TemplateContract
	err := r.db.Where("module = ? AND template_key = ?", registration.ModuleName, registration.TemplateKey).First(&existingContract).Error
	
	if err == nil {
		// Update existing contract
		return r.updateContractWithSampleData(&existingContract, tenantID, registration, sampleData)
	} else if err == gorm.ErrRecordNotFound {
		// Create new contract
		return r.createContractWithSampleData(tenantID, registration, sampleData)
	}
	
	return fmt.Errorf("failed to check existing contract: %w", err)
}

// RegisterContractFromFile registers a contract from a JSON file
func (r *ContractRegistrar) RegisterContractFromFile(tenantID uint, moduleName, contractFilePath string) error {
	// Read contract file
	contractData, err := ioutil.ReadFile(contractFilePath)
	if err != nil {
		return fmt.Errorf("failed to read contract file %s: %w", contractFilePath, err)
	}

	// Parse JSON schema
	var schema map[string]interface{}
	if err := json.Unmarshal(contractData, &schema); err != nil {
		return fmt.Errorf("failed to parse contract JSON: %w", err)
	}

	// Extract template key from filename
	filename := filepath.Base(contractFilePath)
	templateKey := filename[:len(filename)-len(filepath.Ext(filename))]
	if templateKey[len(templateKey)-9:] == "-contract" {
		templateKey = templateKey[:len(templateKey)-9]
	}

	// Try to load sample data from startupseed directory
	sampleDataPath := filepath.Join("startupseed", templateKey+".json")
	var sampleData map[string]interface{}
	if sampleDataBytes, err := ioutil.ReadFile(sampleDataPath); err == nil {
		if err := json.Unmarshal(sampleDataBytes, &sampleData); err != nil {
			fmt.Printf("Warning: Failed to parse sample data for %s: %v\n", templateKey, err)
			sampleData = map[string]interface{}{}
		}
	} else {
		fmt.Printf("Warning: No sample data found for %s\n", templateKey)
		sampleData = map[string]interface{}{}
	}

	// Create registration
	registration := entities.ContractRegistration{
		ModuleName:  moduleName,
		TemplateKey: templateKey,
		Name:        fmt.Sprintf("%s Template", templateKey),
		Description: fmt.Sprintf("Template contract for %s", templateKey),
		Schema:      schema,
		Version:     "1.0.0",
	}

	return r.RegisterContractWithSampleData(tenantID, registration, sampleData)
}

// createContractWithSampleData creates a new contract in the database with sample data
func (r *ContractRegistrar) createContractWithSampleData(tenantID uint, registration entities.ContractRegistration, sampleData map[string]interface{}) error {
	schemaJSON, err := json.Marshal(registration.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	sampleDataJSON, err := json.Marshal(sampleData)
	if err != nil {
		return fmt.Errorf("failed to marshal sample data: %w", err)
	}

	contract := entities.TemplateContract{
		Module:            registration.ModuleName,
		TemplateKey:       registration.TemplateKey,
		Description:       registration.Description,
		SupportedChannels: []byte(`["EMAIL", "DOCUMENT"]`), // Default supported channels
		VariableSchema:    schemaJSON,
		DefaultSampleData: sampleDataJSON,
	}

	if err := r.db.Create(&contract).Error; err != nil {
		return fmt.Errorf("failed to create contract: %w", err)
	}

	return nil
}

// updateContractWithSampleData updates an existing contract with sample data
func (r *ContractRegistrar) updateContractWithSampleData(contract *entities.TemplateContract, tenantID uint, registration entities.ContractRegistration, sampleData map[string]interface{}) error {
	schemaJSON, err := json.Marshal(registration.Schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	sampleDataJSON, err := json.Marshal(sampleData)
	if err != nil {
		return fmt.Errorf("failed to marshal sample data: %w", err)
	}

	contract.Description = registration.Description
	contract.VariableSchema = schemaJSON
	contract.DefaultSampleData = sampleDataJSON

	if err := r.db.Save(contract).Error; err != nil {
		return fmt.Errorf("failed to update contract: %w", err)
	}

	return nil
}

// GetContractByKey retrieves a contract by module and template key
func (r *ContractRegistrar) GetContractByKey(moduleName, templateKey string) (*entities.TemplateContract, error) {
	var contract entities.TemplateContract
	err := r.db.Where("module = ? AND template_key = ?", moduleName, templateKey).First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}