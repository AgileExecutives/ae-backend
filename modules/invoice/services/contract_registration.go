package services

import (
	"path/filepath"

	templateServices "github.com/ae-base-server/modules/templates/services"
)

// RegisterInvoiceContracts registers all invoice module template contracts
func RegisterInvoiceContracts(contractRegistrar *templateServices.ContractRegistrar, tenantID uint) error {
	// Get the module's contracts directory
	contractsDir := "modules/invoice/contracts"

	// Register all contract files
	contracts := []string{
		"std_invoice-contract.json",
	}

	for _, contractFile := range contracts {
		contractPath := filepath.Join(contractsDir, contractFile)
		if err := contractRegistrar.RegisterContractFromFile(tenantID, "invoice", contractPath); err != nil {
			return err
		}
	}

	return nil
}
