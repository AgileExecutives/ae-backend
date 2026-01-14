package services

import (
	"path/filepath"

	templateServices "github.com/ae-base-server/modules/templates/services"
)

// RegisterBaseContracts registers all base template contracts with the template system
func RegisterBaseContracts(contractRegistrar *templateServices.ContractRegistrar, tenantID uint) error {
	// Get the module's contracts directory
	contractsDir := "modules/base/contracts"

	// Register all contract files
	contracts := []string{
		"invoice-contract.json",
	}

	for _, contractFile := range contracts {
		contractPath := filepath.Join(contractsDir, contractFile)
		if err := contractRegistrar.RegisterContractFromFile(tenantID, "base", contractPath); err != nil {
			return err
		}
	}

	return nil
}
