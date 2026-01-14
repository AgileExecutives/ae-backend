package services

import (
	"encoding/json"
	"fmt"

	"github.com/ae-base-server/modules/templates/entities"
	"gorm.io/gorm"
)

// DBContractProvider provides contracts from the database
type DBContractProvider struct {
	db *gorm.DB
}

// NewDBContractProvider creates a new database contract provider
func NewDBContractProvider(db *gorm.DB) *DBContractProvider {
	return &DBContractProvider{
		db: db,
	}
}

// GetContract retrieves a contract schema from the database
func (p *DBContractProvider) GetContract(templateKey string) (map[string]interface{}, error) {
	var contract entities.TemplateContract
	if err := p.db.Where("template_key = ?", templateKey).First(&contract).Error; err != nil {
		return nil, fmt.Errorf("contract not found for template_key %s: %w", templateKey, err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(contract.VariableSchema, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract schema: %w", err)
	}

	return schema, nil
}

// GetAvailableContracts returns all available contract template keys from the database
func (p *DBContractProvider) GetAvailableContracts() ([]string, error) {
	var contracts []entities.TemplateContract
	if err := p.db.Select("template_key").Find(&contracts).Error; err != nil {
		return nil, fmt.Errorf("failed to get available contracts: %w", err)
	}

	keys := make([]string, len(contracts))
	for i, contract := range contracts {
		keys[i] = contract.TemplateKey
	}

	return keys, nil
}

// GetSampleData retrieves sample data for a template from the database
func (p *DBContractProvider) GetSampleData(templateKey string) (map[string]interface{}, error) {
	var contract entities.TemplateContract
	if err := p.db.Where("template_key = ?", templateKey).First(&contract).Error; err != nil {
		return nil, fmt.Errorf("contract not found for template_key %s: %w", templateKey, err)
	}

	var sampleData map[string]interface{}
	if len(contract.DefaultSampleData) > 0 {
		if err := json.Unmarshal(contract.DefaultSampleData, &sampleData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal sample data: %w", err)
		}
	}

	return sampleData, nil
}