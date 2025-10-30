package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	baseAPI "github.com/ae-base-server/api"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"gorm.io/gorm"
)

// SeedAppData represents the structure of the seed_app_data.json file
type SeedAppData struct {
	CostProviders []CostProviderSeed `json:"cost_providers"`
	Clients       []ClientSeed       `json:"clients"`
}

// CostProviderSeed represents cost provider data from JSON
type CostProviderSeed struct {
	Organization string `json:"organization"`
	Department   string `json:"department"`
	Street       string `json:"street"`
	Zip          string `json:"zip"`
	City         string `json:"city"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	State        string `json:"state"`
}

// ClientSeed represents client data from JSON
type ClientSeed struct {
	FirstName            string     `json:"first_name"`
	LastName             string     `json:"last_name"`
	DateOfBirth          *time.Time `json:"date_of_birth"`
	Gender               string     `json:"gender"`
	PrimaryLanguage      string     `json:"primary_language"`
	ContactFirstName     string     `json:"contact_first_name"`
	ContactLastName      string     `json:"contact_last_name"`
	ContactEmail         string     `json:"contact_email"`
	ContactPhone         string     `json:"contact_phone"`
	AlternativeFirstName string     `json:"alternative_first_name"`
	AlternativeLastName  string     `json:"alternative_last_name"`
	AlternativePhone     string     `json:"alternative_phone"`
	AlternativeEmail     string     `json:"alternative_email"`
	StreetAddress        string     `json:"street_address"`
	City                 string     `json:"city"`
	State                string     `json:"state"`
	PostalCode           string     `json:"postal_code"`
	Country              string     `json:"country"`
	InsuranceID          string     `json:"insurance_id"`
	SessionRate          float64    `json:"session_rate"`
	TherapyType          string     `json:"therapy_type"`
	TherapyTitle         string     `json:"therapy_title"` // New field for therapy title from diagnostics
	Diagnosis            string     `json:"diagnosis"`
	TreatmentGoals       string     `json:"treatment_goals"`
	Status               string     `json:"status"`
	Notes                string     `json:"notes"`
	Email                string     `json:"email"` // Direct email field
	Phone                string     `json:"phone"` // Direct phone field
	Zip                  string     `json:"zip"`   // Direct zip field
}

// SeedDatabase seeds the database with data from ae-base-server seed-data.json
// plus unburdy-specific data
func SeedDatabase(db *gorm.DB) error {
	log.Println("🌱 Seeding database with ae-base-server data from seed-data.json...")

	// Use ae-base-server's seed function to load data from seed-data.json
	// This creates: tenants, users, plans from the JSON file
	if err := baseAPI.SeedBaseData(db); err != nil {
		log.Printf("⚠️  Warning: ae-base-server seed failed (may already be seeded): %v", err)
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

	// Seed unburdy-specific data from seed_app_data.json
	if err := seedAppSpecificData(db, tenant.ID); err != nil {
		log.Printf("⚠️  Warning: Failed to seed app-specific data: %v", err)
	}

	return nil
}

// seedAppSpecificData loads and seeds cost providers and clients from seed_app_data.json
func seedAppSpecificData(db *gorm.DB, tenantID uint) error {
	log.Println("🌱 Seeding cost providers and clients from seed_app_data.json...")

	// Read the seed_app_data.json file
	filePath := filepath.Join(".", "seed_app_data.json")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read seed_app_data.json: %w", err)
	}

	var seedData SeedAppData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return fmt.Errorf("failed to parse seed_app_data.json: %w", err)
	}

	// Seed cost providers first
	var createdProviders []models.CostProvider
	for _, providerData := range seedData.CostProviders {
		// Check if this cost provider already exists
		var existingProvider models.CostProvider
		if err := db.Where("organization = ? AND tenant_id = ?", providerData.Organization, tenantID).First(&existingProvider).Error; err == nil {
			// Provider already exists, skip
			createdProviders = append(createdProviders, existingProvider)
			continue
		}

		provider := models.CostProvider{
			TenantID:      tenantID,
			Organization:  providerData.Organization,
			Department:    providerData.Department,
			ContactName:   "", // Not provided in JSON
			StreetAddress: providerData.Street,
			Zip:           providerData.Zip,
			City:          providerData.City,
		}

		if err := db.Create(&provider).Error; err != nil {
			log.Printf("⚠️  Failed to create cost provider '%s': %v", provider.Organization, err)
			continue
		}

		createdProviders = append(createdProviders, provider)
		log.Printf("✅ Created cost provider: %s", provider.Organization)
	}

	// Seed clients and link them to random cost providers
	rand.Seed(time.Now().UnixNano())
	var createdClients []models.Client

	for i, clientData := range seedData.Clients {
		// Check if this client already exists
		var existingClient models.Client
		if err := db.Where("first_name = ? AND last_name = ? AND tenant_id = ?", clientData.FirstName, clientData.LastName, tenantID).First(&existingClient).Error; err == nil {
			// Client already exists, skip
			createdClients = append(createdClients, existingClient)
			continue
		}

		// Randomly assign a cost provider to active clients
		var costProviderID *uint
		if clientData.Status == "active" && len(createdProviders) > 0 {
			randomProvider := createdProviders[rand.Intn(len(createdProviders))]
			costProviderID = &randomProvider.ID
		}

		// Debug log to check therapy title value
		if clientData.TherapyTitle != "" {
			log.Printf("🔍 Client %s has therapy title: '%s'", clientData.FirstName, clientData.TherapyTitle)
		} else {
			log.Printf("⚠️  Client %s has empty therapy title", clientData.FirstName)
		}

		client := models.Client{
			TenantID:             tenantID,
			FirstName:            clientData.FirstName,
			LastName:             clientData.LastName,
			DateOfBirth:          clientData.DateOfBirth,
			Gender:               clientData.Gender,
			PrimaryLanguage:      clientData.PrimaryLanguage,
			ContactFirstName:     clientData.ContactFirstName,
			ContactLastName:      clientData.ContactLastName,
			ContactEmail:         clientData.ContactEmail,
			ContactPhone:         clientData.ContactPhone,
			AlternativeFirstName: clientData.AlternativeFirstName,
			AlternativeLastName:  clientData.AlternativeLastName,
			AlternativePhone:     clientData.AlternativePhone,
			AlternativeEmail:     clientData.AlternativeEmail,
			StreetAddress:        clientData.StreetAddress,
			City:                 clientData.City,
			// Use the new direct fields from JSON if available, otherwise fallback to mapped fields
			Zip:            getStringValue(clientData.Zip, clientData.PostalCode),
			Email:          getStringValue(clientData.Email, clientData.ContactEmail),
			Phone:          clientData.Phone,
			TherapyTitle:   clientData.TherapyTitle, // Use therapy_title directly from JSON
			UnitPrice:      &clientData.SessionRate, // Map session_rate to unit_price
			Status:         clientData.Status,
			Notes:          clientData.Notes,
			CostProviderID: costProviderID,
		}

		if err := db.Create(&client).Error; err != nil {
			log.Printf("⚠️  Failed to create client '%s %s': %v", client.FirstName, client.LastName, err)
			continue
		}

		createdClients = append(createdClients, client)

		providerInfo := "no cost provider"
		if costProviderID != nil {
			providerInfo = fmt.Sprintf("cost provider ID %d", *costProviderID)
		}
		log.Printf("✅ Created client #%d: %s %s (%s) - %s", i+1, client.FirstName, client.LastName, client.Status, providerInfo)
	}

	log.Printf("✅ Seeded %d cost providers and %d clients", len(createdProviders), len(createdClients))
	return nil
}

// getStringValue returns the first non-empty string from the provided values
func getStringValue(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
