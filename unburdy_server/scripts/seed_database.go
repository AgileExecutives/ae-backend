package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/unburdy/unburdy-server-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SeedData represents the structure of our seed data JSON
type SeedData struct {
	CostProviders []CostProviderSeedData `json:"cost_providers"`
	Clients       []ClientSeedData       `json:"clients"`
}

// CostProviderSeedData represents a cost provider from the seed data
type CostProviderSeedData struct {
	Organization string `json:"organization"`
	Department   string `json:"department"`
	Street       string `json:"street"`
	Zip          string `json:"zip"`
	City         string `json:"city"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	State        string `json:"state"`
}

// ClientSeedData represents a client from the seed data
type ClientSeedData struct {
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	DateOfBirth          string `json:"date_of_birth"`
	Gender               string `json:"gender"`
	PrimaryLanguage      string `json:"primary_language"`
	ContactFirstName     string `json:"contact_first_name"`
	ContactLastName      string `json:"contact_last_name"`
	ContactEmail         string `json:"contact_email"`
	ContactPhone         string `json:"contact_phone"`
	AlternativeFirstName string `json:"alternative_first_name"`
	AlternativeLastName  string `json:"alternative_last_name"`
	AlternativePhone     string `json:"alternative_phone"`
	AlternativeEmail     string `json:"alternative_email"`
	StreetAddress        string `json:"street_address"`
	City                 string `json:"city"`
	Zip                  string `json:"zip"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	TherapyTitle         string `json:"therapy_title"`
	Status               string `json:"status"`
	Notes                string `json:"notes"`
	ProviderApprovalCode string `json:"provider_approval_code"`
	ProviderApprovalDate string `json:"provider_approval_date"`
}

func main() {
	// Get the current directory (should be project root)
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current directory:", err)
	}

	// Read seed data
	seedFile := filepath.Join(projectRoot, "seed_app_data.json")
	data, err := os.ReadFile(seedFile)
	if err != nil {
		log.Fatal("Error reading seed data file:", err)
	}

	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		log.Fatal("Error parsing seed data:", err)
	}

	// Initialize PostgreSQL database connection (same as unburdy server)
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "pass")
	dbname := getEnv("DB_NAME", "ae_saas_basic_test")
	sslmode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	log.Printf("Connecting to PostgreSQL: %s:%s/%s", host, port, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&models.Client{}, &models.CostProvider{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Clear existing data (optional - comment out if you want to keep existing data)
	log.Println("Clearing existing clients and cost providers...")
	if err := db.Where("1 = 1").Delete(&models.Client{}).Error; err != nil {
		log.Printf("Warning: Could not clear existing clients: %v", err)
	}
	if err := db.Where("1 = 1").Delete(&models.CostProvider{}).Error; err != nil {
		log.Printf("Warning: Could not clear existing cost providers: %v", err)
	}

	// Seed cost providers first
	log.Printf("Seeding %d cost providers...", len(seedData.CostProviders))
	var createdProviders []models.CostProvider
	costProviderSuccessCount := 0

	for i, providerData := range seedData.CostProviders {
		provider := models.CostProvider{
			TenantID:      1, // Default tenant ID
			Organization:  providerData.Organization,
			Department:    providerData.Department,
			ContactName:   "", // Not provided in JSON
			StreetAddress: providerData.Street,
			Zip:           providerData.Zip,
			City:          providerData.City,
		}

		if err := db.Create(&provider).Error; err != nil {
			log.Printf("❌ Failed to create cost provider %d (%s): %v", i+1, providerData.Organization, err)
		} else {
			costProviderSuccessCount++
			createdProviders = append(createdProviders, provider)
			log.Printf("✅ Created cost provider %d: %s (ID: %d)", i+1, providerData.Organization, provider.ID)
		}
	}

	log.Printf("🎉 Cost provider seeding completed! Successfully created %d out of %d cost providers.", costProviderSuccessCount, len(seedData.CostProviders))

	// Seed clients
	log.Printf("Seeding %d clients...", len(seedData.Clients))

	// Initialize random seed for cost provider assignment
	rand.Seed(time.Now().UnixNano())

	clientSuccessCount := 0
	for i, clientData := range seedData.Clients {
		// Randomly assign a cost provider to active clients
		var costProviderID *uint
		if clientData.Status == "active" && len(createdProviders) > 0 {
			randomProvider := createdProviders[rand.Intn(len(createdProviders))]
			costProviderID = &randomProvider.ID
		}
		// Parse date of birth
		var dateOfBirth *time.Time
		if clientData.DateOfBirth != "" {
			if parsedDate, err := time.Parse(time.RFC3339, clientData.DateOfBirth); err == nil {
				dateOfBirth = &parsedDate
			}
		}

		// Parse provider approval date
		var providerApprovalDate *time.Time
		if clientData.ProviderApprovalDate != "" {
			if parsedDate, err := time.Parse(time.RFC3339, clientData.ProviderApprovalDate); err == nil {
				providerApprovalDate = &parsedDate
			}
		}

		client := models.Client{
			TenantID:             1, // Default tenant ID
			FirstName:            clientData.FirstName,
			LastName:             clientData.LastName,
			DateOfBirth:          dateOfBirth,
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
			Zip:                  clientData.Zip,
			Email:                clientData.Email,
			Phone:                clientData.Phone,
			TherapyTitle:         clientData.TherapyTitle,
			ProviderApprovalCode: clientData.ProviderApprovalCode,
			ProviderApprovalDate: providerApprovalDate,
			Status:               clientData.Status,
			Notes:                clientData.Notes,
			CostProviderID:       costProviderID,
			InvoicedIndividually: false,
		}

		if err := db.Create(&client).Error; err != nil {
			log.Printf("❌ Failed to create client %d (%s %s): %v", i+1, clientData.FirstName, clientData.LastName, err)
		} else {
			clientSuccessCount++
			costProviderInfo := "no cost provider"
			if costProviderID != nil {
				costProviderInfo = fmt.Sprintf("cost provider ID %d", *costProviderID)
			}
			log.Printf("✅ Created client %d: %s %s (%s) - %s (ID: %d)", i+1, clientData.FirstName, clientData.LastName, clientData.Status, costProviderInfo, client.ID)
		}
	}

	log.Printf("\n🎉 Client seeding completed! Successfully created %d out of %d clients.", clientSuccessCount, len(seedData.Clients))

	// Show some statistics
	var totalClients int64
	db.Model(&models.Client{}).Count(&totalClients)
	log.Printf("📊 Total clients in database: %d", totalClients)

	// Show breakdown by therapy type
	log.Println("\n📈 Therapy Type Breakdown:")
	var therapyTypes []struct {
		TherapyTitle string
		Count        int64
	}

	db.Model(&models.Client{}).
		Select("therapy_title, COUNT(*) as count").
		Where("therapy_title IS NOT NULL AND therapy_title != ''").
		Group("therapy_title").
		Order("count DESC").
		Find(&therapyTypes)

	for _, tt := range therapyTypes {
		log.Printf("   %s: %d clients", tt.TherapyTitle, tt.Count)
	}

	// Show breakdown by status
	log.Println("\n📊 Status Breakdown:")
	var statuses []struct {
		Status string
		Count  int64
	}

	db.Model(&models.Client{}).
		Select("status, COUNT(*) as count").
		Where("status IS NOT NULL").
		Group("status").
		Order("count DESC").
		Find(&statuses)

	for _, s := range statuses {
		log.Printf("   %s: %d clients", s.Status, s.Count)
	}

	log.Println("\n✨ Database seeding script completed successfully!")
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
