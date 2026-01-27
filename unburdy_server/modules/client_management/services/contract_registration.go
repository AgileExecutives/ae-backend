package services

import (
	"path/filepath"

	templateServices "github.com/ae-base-server/modules/templates/services"
)

// RegisterClientManagementContracts registers all client management template contracts
func RegisterClientManagementContracts(contractRegistrar *templateServices.ContractRegistrar, tenantID uint) error {
	// Get the module's contracts directory
	contractsDir := "unburdy_server/modules/client_management/contracts"

	// Register all contract files
	contracts := []string{
		"client-invoice-contract.json",
	}

	for _, contractFile := range contracts {
		contractPath := filepath.Join(contractsDir, contractFile)
		if err := contractRegistrar.RegisterContractFromFile(tenantID, "client-management", contractPath); err != nil {
			return err
		}
	}

	return nil
}
