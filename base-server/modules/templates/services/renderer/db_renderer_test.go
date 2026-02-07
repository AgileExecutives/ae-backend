package renderer

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockContractProvider is a mock implementation of ContractProvider.
type MockContractProvider struct {
	mock.Mock
}

func (m *MockContractProvider) GetContract(templateKey string) (map[string]interface{}, error) {
	args := m.Called(templateKey)
	if v := args.Get(0); v != nil {
		return v.(map[string]interface{}), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockContractProvider) GetAvailableContracts() ([]string, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestDBRenderer_GetAvailableContracts(t *testing.T) {
	mockProvider := new(MockContractProvider)
	mockProvider.On("GetAvailableContracts").Return([]string{"password_reset", "welcome"}, nil)
	mockProvider.On("GetContract", "password_reset").Return(minimalSchemaWithRequired([]string{"AppName", "RecipientName"}), nil)
	mockProvider.On("GetContract", "welcome").Return(minimalSchemaWithRequired([]string{"RecipientName"}), nil)

	renderer, err := NewDBRenderer("testdata", mockProvider)
	assert.NoError(t, err)

	contracts := renderer.GetAvailableContracts()
	assert.ElementsMatch(t, []string{"password_reset", "welcome"}, contracts)

	mockProvider.AssertExpectations(t)
}

func TestDBRenderer_ValidateContract_Success(t *testing.T) {
	mockProvider := new(MockContractProvider)
	mockProvider.On("GetAvailableContracts").Return([]string{"password_reset"}, nil)
	mockProvider.On("GetContract", "password_reset").Return(minimalSchemaWithRequired([]string{"AppName", "RecipientName"}), nil)

	renderer, err := NewDBRenderer("testdata", mockProvider)
	assert.NoError(t, err)

	data := map[string]interface{}{
		"AppName":       "Test App",
		"RecipientName": "John Doe",
	}
	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)

	err = renderer.validateContract("password_reset", jsonData)
	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
}

func TestDBRenderer_ValidateContract_InvalidData(t *testing.T) {
	mockProvider := new(MockContractProvider)
	mockProvider.On("GetAvailableContracts").Return([]string{"password_reset"}, nil)
	mockProvider.On("GetContract", "password_reset").Return(minimalSchemaWithRequired([]string{"AppName", "RecipientName"}), nil)

	renderer, err := NewDBRenderer("testdata", mockProvider)
	assert.NoError(t, err)

	// Missing RecipientName
	data := map[string]interface{}{
		"AppName": "Test App",
	}
	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)

	err = renderer.validateContract("password_reset", jsonData)
	assert.Error(t, err)
	mockProvider.AssertExpectations(t)
}

func TestDBRenderer_UnknownContract(t *testing.T) {
	mockProvider := new(MockContractProvider)
	mockProvider.On("GetAvailableContracts").Return([]string{}, nil)

	renderer, err := NewDBRenderer("testdata", mockProvider)
	assert.NoError(t, err)

	jsonData, err := json.Marshal(map[string]interface{}{"field": "value"})
	assert.NoError(t, err)

	err = renderer.validateContract("unknown_contract", jsonData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown contract")
	mockProvider.AssertExpectations(t)
}

func minimalSchemaWithRequired(required []string) map[string]interface{} {
	return map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"AppName": map[string]interface{}{"type": "string"},
			"RecipientName": map[string]interface{}{
				"type": "string",
			},
		},
		"required": required,
	}
}
