package database

import (
	"log"

	baseAPI "github.com/ae-saas-basic/ae-saas-basic/api"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"gorm.io/gorm"
)

// SeedDatabase seeds the database with data from ae-saas seed-data.json
// plus unburdy-specific data
func SeedDatabase(db *gorm.DB) error {
	log.Println("🌱 Seeding database with ae-saas-basic data from seed-data.json...")

	// Use ae-saas-basic's seed function to load data from seed-data.json
	// This creates: tenants, users, plans from the JSON file
	if err := baseAPI.SeedBaseData(db); err != nil {
		log.Printf("⚠️  Warning: ae-saas seed failed (may already be seeded): %v", err)
		// Don't return error - continue with unburdy-specific seeding
	}

	// Get the first tenant to use as default for clients
	var tenant baseAPI.Tenant
	if err := db.First(&tenant).Error; err != nil {
		log.Printf("⚠️  Warning: No tenant found after seeding: %v", err)
		return nil // Return early if no tenant exists
	}

	// Get the first user to use as default creator for clients
	var user baseAPI.User
	if err := db.First(&user).Error; err != nil {
		log.Printf("⚠️  Warning: No user found after seeding: %v", err)
		return nil // Return early if no user exists
	}

	log.Printf("✅ Using tenant '%s' (ID: %d) and user '%s' (ID: %d) for client references",
		tenant.Name, tenant.ID, user.Email, user.ID)

	// Update existing clients to have proper tenant and user references
	var orphanedClients []models.Client
	if err := db.Where("tenant_id = 0 OR created_by = 0").Find(&orphanedClients).Error; err != nil {
		return err
	}

	for _, client := range orphanedClients {
		client.TenantID = tenant.ID
		client.CreatedBy = user.ID
		if err := db.Save(&client).Error; err != nil {
			log.Printf("Warning: Failed to update client %d: %v", client.ID, err)
		} else {
			log.Printf("✅ Updated client %d with tenant and user references", client.ID)
		}
	}

	return nil
}
