package services

import (
	"path/filepath"

	templateServices "github.com/ae-base-server/modules/templates/services"
)

// RegisterEmailContracts registers all email template contracts with the template system
func RegisterEmailContracts(contractRegistrar *templateServices.ContractRegistrar, tenantID uint) error {
	// Get the module's contracts directory
	contractsDir := "modules/email/contracts"
	
	// Register all contract files
	contracts := []string{
		"email_verification-contract.json",
		"password_reset-contract.json", 
		"welcome-contract.json",
	}
	
	for _, contractFile := range contracts {
		contractPath := filepath.Join(contractsDir, contractFile)
		if err := contractRegistrar.RegisterContractFromFile(tenantID, "email", contractPath); err != nil {
			return err
		}
	}
	
	return nil
}