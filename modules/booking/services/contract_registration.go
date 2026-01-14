package services

import (
	"path/filepath"

	templateServices "github.com/ae-base-server/modules/templates/services"
)

// RegisterBookingContracts registers all booking template contracts with the template system
func RegisterBookingContracts(contractRegistrar *templateServices.ContractRegistrar, tenantID uint) error {
	// Get the module's contracts directory
	contractsDir := "modules/booking/contracts"

	// Register all contract files
	contracts := []string{
		"booking_confirmation-contract.json",
	}

	for _, contractFile := range contracts {
		contractPath := filepath.Join(contractsDir, contractFile)
		if err := contractRegistrar.RegisterContractFromFile(tenantID, "booking", contractPath); err != nil {
			return err
		}
	}

	return nil
}
