package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ae-base-server/modules/templates/entities"
	"gorm.io/gorm"
)

var (
	ErrContractNotFound   = errors.New("template contract not found")
	ErrDuplicateContract  = errors.New("contract with this module and template_key already exists")
	ErrUnsupportedChannel = errors.New("channel not supported by contract")
	ErrInvalidChannelList = errors.New("supported_channels must contain at least one channel")
	ErrInvalidModule      = errors.New("module name is required")
	ErrInvalidTemplateKey = errors.New("template_key is required")
)

// ContractService handles template contract operations
type ContractService struct {
	db *gorm.DB
}

// NewContractService creates a new contract service
func NewContractService(db *gorm.DB) *ContractService {
	return &ContractService{
		db: db,
	}
}

// RegisterContract registers or updates a template contract
func (s *ContractService) RegisterContract(ctx context.Context, req *entities.RegisterContractRequest) (*entities.TemplateContract, error) {
	// Validate required fields
	if req.Module == "" {
		return nil, ErrInvalidModule
	}
	if req.TemplateKey == "" {
		return nil, ErrInvalidTemplateKey
	}
	if len(req.SupportedChannels) == 0 {
		return nil, ErrInvalidChannelList
	}

	// Convert request data to JSONB
	channelsJSON, err := entities.MarshalJSON(req.SupportedChannels)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal supported_channels: %w", err)
	}

	var schemaJSON, sampleDataJSON []byte
	if req.VariableSchema != nil {
		schemaJSON, err = entities.MarshalJSON(req.VariableSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variable_schema: %w", err)
		}
	}
	if req.DefaultSampleData != nil {
		sampleDataJSON, err = entities.MarshalJSON(req.DefaultSampleData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default_sample_data: %w", err)
		}
	}

	// Check if contract already exists
	var existing entities.TemplateContract
	err = s.db.Where("module = ? AND template_key = ?", req.Module, req.TemplateKey).
		First(&existing).Error

	if err == nil {
		// Update existing contract
		existing.Description = req.Description
		existing.SupportedChannels = channelsJSON
		existing.VariableSchema = schemaJSON
		existing.DefaultSampleData = sampleDataJSON

		if err := s.db.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to update contract: %w", err)
		}
		return &existing, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing contract: %w", err)
	}

	// Create new contract
	contract := &entities.TemplateContract{
		Module:            req.Module,
		TemplateKey:       req.TemplateKey,
		Description:       req.Description,
		SupportedChannels: channelsJSON,
		VariableSchema:    schemaJSON,
		DefaultSampleData: sampleDataJSON,
	}

	if err := s.db.Create(contract).Error; err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	return contract, nil
}

// GetContract retrieves a contract by module and template_key
func (s *ContractService) GetContract(ctx context.Context, module, templateKey string) (*entities.TemplateContract, error) {
	var contract entities.TemplateContract
	err := s.db.Where("module = ? AND template_key = ?", module, templateKey).
		First(&contract).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContractNotFound
		}
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}

	return &contract, nil
}

// GetContractByID retrieves a contract by ID
func (s *ContractService) GetContractByID(ctx context.Context, id uint) (*entities.TemplateContract, error) {
	var contract entities.TemplateContract
	err := s.db.First(&contract, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrContractNotFound
		}
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}

	return &contract, nil
}

// ListContracts retrieves all contracts, optionally filtered by module
func (s *ContractService) ListContracts(ctx context.Context, module string) ([]entities.TemplateContract, error) {
	var contracts []entities.TemplateContract
	query := s.db

	if module != "" {
		query = query.Where("module = ?", module)
	}

	if err := query.Find(&contracts).Error; err != nil {
		return nil, fmt.Errorf("failed to list contracts: %w", err)
	}

	return contracts, nil
}

// ValidateChannel checks if a channel is supported by a contract
func (s *ContractService) ValidateChannel(ctx context.Context, module, templateKey, channel string) error {
	contract, err := s.GetContract(ctx, module, templateKey)
	if err != nil {
		return err
	}

	if !contract.SupportsChannel(channel) {
		return fmt.Errorf("%w: %s not supported for %s.%s",
			ErrUnsupportedChannel, channel, module, templateKey)
	}

	return nil
}

// UpdateContract updates an existing contract
func (s *ContractService) UpdateContract(ctx context.Context, id uint, req *entities.UpdateContractRequest) (*entities.TemplateContract, error) {
	contract, err := s.GetContractByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Description != nil {
		contract.Description = *req.Description
	}

	if req.SupportedChannels != nil {
		channelsJSON, err := entities.MarshalJSON(*req.SupportedChannels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal supported_channels: %w", err)
		}
		contract.SupportedChannels = channelsJSON
	}

	if req.VariableSchema != nil {
		schemaJSON, err := entities.MarshalJSON(*req.VariableSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal variable_schema: %w", err)
		}
		contract.VariableSchema = schemaJSON
	}

	if req.DefaultSampleData != nil {
		sampleDataJSON, err := entities.MarshalJSON(*req.DefaultSampleData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default_sample_data: %w", err)
		}
		contract.DefaultSampleData = sampleDataJSON
	}

	if err := s.db.Save(contract).Error; err != nil {
		return nil, fmt.Errorf("failed to update contract: %w", err)
	}

	return contract, nil
}

// DeleteContract deletes a contract
func (s *ContractService) DeleteContract(ctx context.Context, id uint) error {
	// Check if any templates are using this contract
	var templateCount int64
	err := s.db.Model(&entities.Template{}).
		Where("module = (SELECT module FROM template_contracts WHERE id = ?) AND template_key = (SELECT template_key FROM template_contracts WHERE id = ?)", id, id).
		Count(&templateCount).Error

	if err != nil {
		return fmt.Errorf("failed to check template usage: %w", err)
	}

	if templateCount > 0 {
		return fmt.Errorf("cannot delete contract: %d templates are using it", templateCount)
	}

	if err := s.db.Delete(&entities.TemplateContract{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete contract: %w", err)
	}

	return nil
}
