// Consolidated seeding application for Unburdy Server
// Handles base-server entities and application-specific data
//
// Usage:
//   go run seed_database.go          # Full seeding

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	// Base server API for seeding base entities
	baseAPI "github.com/ae-base-server/api"

	// Application entities
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"

	// Calendar seeding
	calendarSeeding "github.com/unburdy/calendar-module/seeding"

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

// JugendamtSeedData represents a Jugendamt from the static JSON (same structure as CostProviderSeedData)
type JugendamtSeedData struct {
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
	// Check for calendar-only mode from environment or command line
	calendarOnly := getEnv("SEED_CALENDAR_ONLY", "false") == "true"
	
	if calendarOnly {
		log.Println("📅 Unburdy Server - Calendar-Only Seeding")
		log.Println("========================================")
	} else {
		log.Println("🌱 Unburdy Server - Complete Database Seeding")
		log.Println("===========================================")
	}

	// Initialize database connection
	db, err := connectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if !calendarOnly {
		// Step 1: Auto-migrate all entities (base-server + modules)
		if err := autoMigrateEntities(db); err != nil {
			log.Fatal("Failed to migrate entities:", err)
		}

		// Step 2: Seed base-server data (tenants, users, plans)
		if err := seedBaseData(db); err != nil {
			log.Printf("⚠️  Warning: Base data seeding failed (may already exist): %v", err)
		}

		// Step 3: Seed application-specific data (cost providers, clients)
		if err := seedAppData(db); err != nil {
			log.Fatal("Failed to seed application data:", err)
		}
	} else {
		log.Println("⏭️  Skipping base data and application data seeding (calendar-only mode)")
	}

	// Step 4: Seed calendar data for users
	if err := seedCalendarData(db); err != nil {
		if calendarOnly {
			log.Fatal("Failed to seed calendar data:", err)
		} else {
			log.Printf("⚠️  Warning: Calendar seeding failed (may be optional): %v", err)
		}
	}

	if calendarOnly {
		log.Println("\n📅 Calendar seeding finished successfully!")
	} else {
		log.Println("\n✨ Complete database seeding finished successfully!")
	}
}

// connectDatabase establishes database connection using environment variables
func connectDatabase() (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "pass")
	dbname := getEnv("DB_NAME", "ae_saas_basic_test")
	sslmode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	log.Printf("🔗 Connecting to PostgreSQL: %s:%s/%s", host, port, dbname)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// autoMigrateEntities performs auto-migration for all entities
func autoMigrateEntities(db *gorm.DB) error {
	log.Println("🏗️  Auto-migrating database entities...")

	// Migrate base-server entities
	if err := baseAPI.MigrateBaseEntities(db); err != nil {
		return fmt.Errorf("failed to migrate base entities: %w", err)
	}

	// Migrate application entities
	if err := db.AutoMigrate(&entities.Client{}, &entities.CostProvider{}); err != nil {
		return fmt.Errorf("failed to migrate app entities: %w", err)
	}

	log.Println("✅ Database migration completed")
	return nil
}

// seedBaseData seeds base-server data (tenants, users, plans)
func seedBaseData(db *gorm.DB) error {
	log.Println("🌱 Seeding base-server data (tenants, users, plans)...")

	// Find seed-data.json file (should be in project root)
	seedFile := findSeedDataFile()
	if seedFile == "" {
		return fmt.Errorf("seed-data.json not found")
	}

	log.Printf("📋 Using seed file: %s", seedFile)
	return baseAPI.SeedBaseData(db)
}

// seedAppData seeds application-specific data (cost providers, clients)
func seedAppData(db *gorm.DB) error {
	log.Println("🌱 Seeding application data (cost providers, clients)...")

	// Read seed data
	seedFile := filepath.Join(".", "seed_app_data.json")
	data, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("failed to read seed_app_data.json: %w", err)
	}

	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return fmt.Errorf("failed to parse seed_app_data.json: %w", err)
	}

	// Get tenant for entity relationships
	var tenant baseAPI.Tenant
	if err := db.First(&tenant).Error; err != nil {
		return fmt.Errorf("no tenant found after base seeding: %w", err)
	}

	// Load additional cost providers from jugendaemter.json
	jugendaemterProviders, err := loadJugendaemterData()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to load jugendaemter data: %v", err)
	} else {
		// Convert JugendamtSeedData to CostProviderSeedData format
		for _, jugendamt := range jugendaemterProviders {
			costProvider := CostProviderSeedData{
				Organization: jugendamt.Organization,
				Department:   jugendamt.Department,
				Street:       jugendamt.Street,
				Zip:          jugendamt.Zip,
				City:         jugendamt.City,
				Phone:        jugendamt.Phone,
				Email:        jugendamt.Email,
				State:        jugendamt.State,
			}
			seedData.CostProviders = append(seedData.CostProviders, costProvider)
		}
		log.Printf("📋 Added %d cost providers from jugendaemter.json", len(jugendaemterProviders))
	}

	// Seed cost providers (now includes both sources)
	createdProviders, err := seedCostProviders(db, seedData.CostProviders, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to seed cost providers: %w", err)
	}

	// Seed clients
	if err := seedClients(db, seedData.Clients, createdProviders, tenant.ID); err != nil {
		return fmt.Errorf("failed to seed clients: %w", err)
	}

	// Show statistics
	showSeedingStatistics(db)

	return nil
}

// seedCostProviders seeds cost provider data
func seedCostProviders(db *gorm.DB, providerData []CostProviderSeedData, tenantID uint) ([]entities.CostProvider, error) {
	log.Printf("📊 Seeding %d cost providers...", len(providerData))

	var createdProviders []entities.CostProvider
	successCount := 0

	for i, providerData := range providerData {
		// Check if already exists
		var existing entities.CostProvider
		if err := db.Where("organization = ? AND tenant_id = ?", providerData.Organization, tenantID).First(&existing).Error; err == nil {
			createdProviders = append(createdProviders, existing)
			continue
		}

		provider := entities.CostProvider{
			TenantID:      tenantID,
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
			successCount++
			createdProviders = append(createdProviders, provider)
			log.Printf("✅ Created cost provider %d: %s (ID: %d)", i+1, providerData.Organization, provider.ID)
		}
	}

	log.Printf("🎉 Cost provider seeding completed! Successfully created %d out of %d.", successCount, len(providerData))
	return createdProviders, nil
}

// seedClients seeds client data
func seedClients(db *gorm.DB, clientData []ClientSeedData, providers []entities.CostProvider, tenantID uint) error {
	log.Printf("👥 Seeding %d clients...", len(clientData))

	rand.Seed(time.Now().UnixNano())
	successCount := 0

	for i, clientData := range clientData {
		// Check if already exists
		var existing entities.Client
		if err := db.Where("first_name = ? AND last_name = ? AND tenant_id = ?", clientData.FirstName, clientData.LastName, tenantID).First(&existing).Error; err == nil {
			continue
		}

		// Randomly assign a cost provider to active clients
		var costProviderID *uint
		if clientData.Status == "active" && len(providers) > 0 {
			randomProvider := providers[rand.Intn(len(providers))]
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

		client := entities.Client{
			TenantID:             tenantID,
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
			successCount++
			costProviderInfo := "no cost provider"
			if costProviderID != nil {
				costProviderInfo = fmt.Sprintf("cost provider ID %d", *costProviderID)
			}
			log.Printf("✅ Created client %d: %s %s (%s) - %s (ID: %d)",
				i+1, clientData.FirstName, clientData.LastName, clientData.Status, costProviderInfo, client.ID)
		}
	}

	log.Printf("🎉 Client seeding completed! Successfully created %d out of %d clients.", successCount, len(clientData))
	return nil
}

// showSeedingStatistics displays seeding results
func showSeedingStatistics(db *gorm.DB) {
	log.Println("\n📊 Seeding Statistics")
	log.Println("====================")

	// Total counts
	var totalClients, totalProviders int64
	db.Model(&entities.Client{}).Count(&totalClients)
	db.Model(&entities.CostProvider{}).Count(&totalProviders)

	log.Printf("📈 Total clients: %d", totalClients)
	log.Printf("🏢 Total cost providers: %d", totalProviders)

	// Therapy type breakdown
	log.Println("\n🎯 Therapy Type Breakdown:")
	var therapyTypes []struct {
		TherapyTitle string
		Count        int64
	}

	db.Model(&entities.Client{}).
		Select("therapy_title, COUNT(*) as count").
		Where("therapy_title IS NOT NULL AND therapy_title != ''").
		Group("therapy_title").
		Order("count DESC").
		Find(&therapyTypes)

	for _, tt := range therapyTypes {
		log.Printf("   %s: %d clients", tt.TherapyTitle, tt.Count)
	}

	// Status breakdown
	log.Println("\n📋 Status Breakdown:")
	var statuses []struct {
		Status string
		Count  int64
	}

	db.Model(&entities.Client{}).
		Select("status, COUNT(*) as count").
		Where("status IS NOT NULL").
		Group("status").
		Order("count DESC").
		Find(&statuses)

	for _, s := range statuses {
		log.Printf("   %s: %d clients", s.Status, s.Count)
	}
}

// findSeedDataFile locates the seed-data.json file
func findSeedDataFile() string {
	// Try current directory
	if _, err := os.Stat("seed-data.json"); err == nil {
		return "seed-data.json"
	}

	// Try parent directory (project root)
	if _, err := os.Stat("../seed-data.json"); err == nil {
		return "../seed-data.json"
	}

	return ""
}

// loadJugendaemterData loads cost provider data from the jugendaemter.json file
func loadJugendaemterData() ([]JugendamtSeedData, error) {
	// Try different possible locations for the jugendaemter.json file
	possiblePaths := []string{
		"../statics/json/jugendaemter.json",                // From seed directory
		"./statics/json/jugendaemter.json",                 // From project root
		"../unburdy_server/statics/json/jugendaemter.json", // From base-server
	}

	var data []byte
	var err error
	var usedPath string

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			usedPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("jugendaemter.json not found in any expected location: %w", err)
	}

	log.Printf("📋 Loading jugendaemter data from: %s", usedPath)

	var jugendaemter []JugendamtSeedData
	if err := json.Unmarshal(data, &jugendaemter); err != nil {
		return nil, fmt.Errorf("failed to parse jugendaemter.json: %w", err)
	}

	return jugendaemter, nil
}

// seedCalendarData seeds calendar data for all users
func seedCalendarData(db *gorm.DB) error {
	log.Println("🗓️  Seeding calendar data...")

	// Get all users to seed calendars for
	var users []baseAPI.User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) == 0 {
		log.Println("⚠️  No users found for calendar seeding")
		return nil
	}

	// Create calendar seeder
	seeder := calendarSeeding.NewCalendarSeeder(db)

	successCount := 0
	for _, user := range users {
		log.Printf("🗓️  Seeding calendar data for user %d: %s", user.ID, user.Email)
		
		if err := seeder.SeedCalendarData(user.TenantID, user.ID); err != nil {
			log.Printf("❌ Failed to seed calendar for user %d (%s): %v", user.ID, user.Email, err)
			continue
		}
		
		successCount++
		log.Printf("✅ Successfully seeded calendar data for user %d (%s)", user.ID, user.Email)
	}

	log.Printf("🎉 Calendar seeding completed! Successfully seeded %d out of %d users.", successCount, len(users))
	
	// Show calendar seeding statistics
	showCalendarStatistics(db)
	return nil
}

// showCalendarStatistics displays calendar seeding results  
func showCalendarStatistics(db *gorm.DB) {
	log.Println("\n📅 Calendar Seeding Statistics")
	log.Println("===============================")

	// Import calendar entities to count them
	type Calendar struct {
		ID       uint `gorm:"primarykey"`
		TenantID uint
		UserID   uint
		Title    string
	}
	
	type CalendarEntry struct {
		ID         uint `gorm:"primarykey"`
		TenantID   uint
		UserID     uint
		CalendarID uint
		Title      string
		Type       string
	}
	
	type CalendarSeries struct {
		ID         uint `gorm:"primarykey"`
		TenantID   uint
		UserID     uint
		CalendarID uint
		Title      string
	}

	// Total counts
	var totalCalendars, totalEntries, totalSeries int64
	db.Model(&Calendar{}).Count(&totalCalendars)
	db.Model(&CalendarEntry{}).Count(&totalEntries)
	db.Model(&CalendarSeries{}).Count(&totalSeries)

	log.Printf("📋 Total calendars: %d", totalCalendars)
	log.Printf("📅 Total calendar entries: %d", totalEntries)
	log.Printf("🔄 Total recurring series: %d", totalSeries)

	// Calendar breakdown by user
	log.Println("\n👤 Calendar Breakdown by User:")
	var calendarsByUser []struct {
		UserID uint
		Count  int64
	}
	
	db.Model(&Calendar{}).
		Select("user_id, COUNT(*) as count").
		Group("user_id").
		Find(&calendarsByUser)

	for _, cu := range calendarsByUser {
		log.Printf("   User %d: %d calendars", cu.UserID, cu.Count)
	}

	// Entry type breakdown
	log.Println("\n🎯 Calendar Entry Type Breakdown:")
	var entryTypes []struct {
		Type  string
		Count int64
	}

	db.Model(&CalendarEntry{}).
		Select("type, COUNT(*) as count").
		Where("type IS NOT NULL AND type != ''").
		Group("type").
		Order("count DESC").
		Find(&entryTypes)

	for _, et := range entryTypes {
		log.Printf("   %s: %d entries", et.Type, et.Count)
	}

	// Holiday-specific counts
	log.Println("\n🎉 Holiday Entry Details:")
	var holidayCounts []struct {
		Type  string
		Count int64
	}

	db.Model(&CalendarEntry{}).
		Select("type, COUNT(*) as count").
		Where("type IN (?, ?)", "public_holiday", "school_holiday").
		Group("type").
		Find(&holidayCounts)

	for _, hc := range holidayCounts {
		if hc.Type == "public_holiday" {
			log.Printf("   🎉 Public holidays: %d entries", hc.Count)
		} else if hc.Type == "school_holiday" {
			log.Printf("   🏫 School holidays: %d entries", hc.Count)
		}
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
