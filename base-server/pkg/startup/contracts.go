package startup

import (
	"fmt"

	baseServices "github.com/ae-base-server/modules/base/services"
	emailServices "github.com/ae-base-server/modules/email/services"
	templateServices "github.com/ae-base-server/modules/templates/services"
	"gorm.io/gorm"
)

// RegisterAllContracts registers all module contracts with the template system
func RegisterAllContracts(db *gorm.DB) error {
	contractRegistrar := templateServices.NewContractRegistrar(db)

	fmt.Println("🔧 Registering template contracts for all tenants...")

	// Get all tenants from database
	var tenants []struct {
		ID uint
	}
	if err := db.Table("tenants").Select("id").Where("deleted_at IS NULL").Find(&tenants).Error; err != nil {
		return fmt.Errorf("failed to fetch tenants: %w", err)
	}

	fmt.Printf("Found %d tenants for contract registration\n", len(tenants))

	// Register contracts from each module
	modules := map[string]func(*templateServices.ContractRegistrar, uint) error{
		"email": emailServices.RegisterEmailContracts,
		"base":  baseServices.RegisterBaseContracts,
	}

	// Register for each tenant
	for _, tenant := range tenants {
		fmt.Printf("📝 Registering contracts for tenant ID: %d\n", tenant.ID)
		for moduleName, registerFunc := range modules {
			fmt.Printf("  - Module: %s\n", moduleName)
			if err := registerFunc(contractRegistrar, tenant.ID); err != nil {
				return fmt.Errorf("failed to register %s contracts for tenant %d: %w", moduleName, tenant.ID, err)
			}
		}
	}

	// Verify contracts were created
	var contractCount int64
	db.Table("template_contracts").Count(&contractCount)
	fmt.Printf("✅ All contracts registered successfully - Total contracts in database: %d\n", contractCount)
	return nil
}
