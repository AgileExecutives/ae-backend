// Consolidated seeding application for Unburdy Server
// Handles base-server entities and application-specific data
//
// Usage:
//   go run seed_database.go          # Full seeding

package main

import (
	"context"
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

	// Invoice numbers module
	invoiceNumberServices "github.com/ae-base-server/modules/invoice_number/services"

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

		// Step 2.5: Seed organizations from seed-data.json
		if err := seedOrganizations(db); err != nil {
			log.Printf("⚠️  Warning: Organizations seeding failed: %v", err)
		}

		// Step 3: Seed application-specific data (cost providers, clients)
		if err := seedAppData(db); err != nil {
			log.Fatal("Failed to seed application data:", err)
		}

		// Step 4: Seed invoice numbers (templates are now seeded during startup)
		if err := seedInvoiceNumbersOnly(db); err != nil {
			log.Printf("⚠️  Warning: Invoice numbers seeding failed: %v", err)
		}
	} else {
		log.Println("⏭️  Skipping base data and application data seeding (calendar-only mode)")
	}

	// Step 5: Seed calendar data for users
	if err := seedCalendarData(db); err != nil {
		if calendarOnly {
			log.Fatal("Failed to seed calendar data:", err)
		} else {
			log.Printf("⚠️  Warning: Calendar seeding failed (may be optional): %v", err)
		}
	}

	// Step 6: Seed sessions linked to calendar entries (only in full seeding mode)
	if !calendarOnly {
		if err := seedSessions(db); err != nil {
			log.Printf("⚠️  Warning: Session seeding failed: %v", err)
		}

		// Step 6.5: Seed extra efforts (unbilled work outside sessions)
		if err := seedExtraEfforts(db); err != nil {
			log.Printf("⚠️  Warning: Extra efforts seeding failed: %v", err)
		}

		// Step 7: Seed invoices with invoice items (only in full seeding mode)
		if err := seedInvoices(db); err != nil {
			log.Printf("⚠️  Warning: Invoice seeding failed: %v", err)
		}

		// Show final statistics including sessions and invoices
		showSessionAndInvoiceStatistics(db)
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
	if err := db.AutoMigrate(&entities.Client{}, &entities.CostProvider{}, &entities.ExtraEffort{}); err != nil {
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

// seedOrganizations seeds organizations from seed-data.json
func seedOrganizations(db *gorm.DB) error {
	log.Println("🏢 Seeding organizations...")

	// Load organizations from seed-data.json
	seedFile := findSeedDataFile()
	if seedFile == "" {
		return fmt.Errorf("seed-data.json not found")
	}

	data, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("failed to read seed-data.json: %w", err)
	}

	var seedData struct {
		Organizations []struct {
			TenantID uint   `json:"tenant_id"`
			Name     string `json:"name"`
		} `json:"organizations"`
	}

	if err := json.Unmarshal(data, &seedData); err != nil {
		return fmt.Errorf("failed to parse seed-data.json: %w", err)
	}

	// Import the organization module entities
	// We need to use a type assertion here since it's a different module
	type Organization struct {
		ID       uint
		TenantID uint
		Name     string
	}

	// Create organizations
	var orgCount int64
	db.Table("organizations").Count(&orgCount)
	if orgCount == 0 {
		for _, orgData := range seedData.Organizations {
			org := Organization{
				TenantID: orgData.TenantID,
				Name:     orgData.Name,
			}
			if err := db.Table("organizations").Create(&org).Error; err != nil {
				return fmt.Errorf("failed to create organization %s: %w", orgData.Name, err)
			}
			log.Printf("✅ Created organization: %s (Tenant ID: %d)", orgData.Name, orgData.TenantID)
		}
	} else {
		log.Println("⏭️  Organizations already exist, skipping...")
	}

	return nil
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

	// Try parent directory startupseed folder (project root)
	if _, err := os.Stat("../startupseed/seed-data.json"); err == nil {
		return "../startupseed/seed-data.json"
	}

	// Try startupseed folder from current directory
	if _, err := os.Stat("startupseed/seed-data.json"); err == nil {
		return "startupseed/seed-data.json"
	}

	// Try parent directory (legacy location)
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

/*
// DEPRECATED: seedDocumentsData - Templates are now seeded during application startup
// This function is kept for reference but is no longer used
// seedDocumentsData seeds documents module data (templates, invoice numbers, sample documents)
func seedDocumentsData(db *gorm.DB) error {
	log.Println("📄 Seeding documents module data (templates, invoice numbers, documents)...")

	// Get tenant for entity relationships
	var tenant baseAPI.Tenant
	if err := db.First(&tenant).Error; err != nil {
		return fmt.Errorf("no tenant found: %w", err)
	}

	// Get first user for document ownership
	var user baseAPI.User
	if err := db.First(&user).Error; err != nil {
		return fmt.Errorf("no user found: %w", err)
	}

	// Initialize MinIO storage for documents
	docMinioConfig := documentStorage.MinIOConfig{
		Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		UseSSL:          getEnv("MINIO_USE_SSL", "false") == "true",
		Region:          getEnv("MINIO_REGION", "us-east-1"),
	}

	docMinioStorage, err := documentStorage.NewMinIOStorage(docMinioConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize document MinIO storage: %w", err)
	}

	// Initialize MinIO storage for templates
	templateMinioConfig := templateStorage.MinIOConfig{
		Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		UseSSL:          getEnv("MINIO_USE_SSL", "false") == "true",
		Region:          getEnv("MINIO_REGION", "us-east-1"),
	}

	templateMinioStorage, err := templateStorage.NewMinIOStorage(templateMinioConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize template MinIO storage: %w", err)
	}

	// Initialize services
	templateService := templateServices.NewTemplateService(db, templateMinioStorage)
	invoiceNumberService := baseAPI.NewInvoiceNumberService(db)
	pdfService := documentServices.NewPDFService(db, docMinioStorage)

	// Create context for operations
	ctx := context.Background()

	// Step 1: Seed invoice template
	log.Println("📝 Creating invoice template...")
	template, err := seedInvoiceTemplate(ctx, templateService, tenant.ID)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create invoice template: %v", err)
	} else {
		log.Printf("✅ Created invoice template (ID: %d)", template.ID)
	}

	// Step 2: Seed email template
	log.Println("📧 Creating email template...")
	emailTemplate, err := seedEmailTemplate(ctx, templateService, tenant.ID)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create email template: %v", err)
	} else {
		log.Printf("✅ Created email template (ID: %d)", emailTemplate.ID)
	}

	// Step 3: Generate some invoice numbers
	log.Println("🔢 Generating sample invoice numbers...")
	invoiceNumbers, err := seedInvoiceNumbers(ctx, invoiceNumberService, tenant.ID)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to generate invoice numbers: %v", err)
	} else {
		log.Printf("✅ Generated %d invoice numbers", len(invoiceNumbers))
	}

	// Step 4: Generate sample invoice PDFs
	if template != nil && len(invoiceNumbers) > 0 {
		log.Println("📄 Generating sample invoice PDFs...")
		err := seedSampleInvoices(ctx, db, pdfService, template.ID, invoiceNumbers, tenant.ID, user.ID)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to generate sample invoices: %v", err)
		} else {
			log.Printf("✅ Generated sample invoice PDFs")
		}
	}

	// Show documents statistics
	showDocumentsStatistics(db)

	log.Println("✅ Documents module seeding completed")
	return nil
}
*/

// seedInvoiceNumbersOnly seeds only invoice numbers (templates are now seeded during startup)
func seedInvoiceNumbersOnly(db *gorm.DB) error {
	log.Println("🔢 Seeding invoice numbers...")

	// Get tenant for entity relationships
	var tenant baseAPI.Tenant
	if err := db.First(&tenant).Error; err != nil {
		return fmt.Errorf("no tenant found: %w", err)
	}

	// Initialize invoice number service
	invoiceNumberService := baseAPI.NewInvoiceNumberService(db)

	// Create context for operations
	ctx := context.Background()

	// Generate some invoice numbers
	log.Println("🔢 Generating sample invoice numbers...")
	invoiceNumbers, err := seedInvoiceNumbers(ctx, invoiceNumberService, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to generate invoice numbers: %w", err)
	}

	log.Printf("✅ Generated %d invoice numbers", len(invoiceNumbers))
	return nil
}

/*
// DEPRECATED TEMPLATE FUNCTIONS - Templates are now seeded during application startup
// seedInvoiceTemplate creates a sample invoice template
func seedInvoiceTemplate(ctx context.Context, service *templateServices.TemplateService, tenantID uint) (*templateEntities.Template, error) {
	// Check if template already exists
	templates, count, err := service.ListTemplates(ctx, tenantID, nil, "DOCUMENT", "invoice", nil, 1, 1)
	if err == nil && count > 0 {
		log.Println("📋 Invoice template already exists, skipping...")
		return &templates[0], nil
	}

	htmlContent := `<!DOCTYPE html>
<html lang="de">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Rechnung {{.invoice_number}}</title>
    <style>
        @page {
            margin: 2cm;
        }
        body {
            font-family: 'Helvetica Neue', Arial, sans-serif;
            color: #333;
            line-height: 1.6;
            margin: 0;
            padding: 0;
        }
        .header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 40px;
            border-bottom: 2px solid #2c3e50;
            padding-bottom: 20px;
        }
        .company-info {
            flex: 1;
        }
        .company-name {
            font-size: 24px;
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 5px;
        }
        .company-details {
            font-size: 12px;
            color: #7f8c8d;
        }
        .invoice-info {
            text-align: right;
        }
        .invoice-title {
            font-size: 28px;
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 10px;
        }
        .invoice-number {
            font-size: 14px;
            color: #7f8c8d;
        }
        .client-info {
            margin: 30px 0;
            padding: 20px;
            background-color: #f8f9fa;
            border-left: 4px solid #3498db;
        }
        .client-label {
            font-size: 12px;
            color: #7f8c8d;
            text-transform: uppercase;
            margin-bottom: 5px;
        }
        .client-name {
            font-size: 16px;
            font-weight: bold;
            color: #2c3e50;
        }
        .dates {
            display: flex;
            justify-content: space-between;
            margin: 20px 0;
            font-size: 14px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 30px 0;
        }
        thead {
            background-color: #2c3e50;
            color: white;
        }
        th {
            padding: 12px;
            text-align: left;
            font-weight: 600;
            font-size: 14px;
        }
        td {
            padding: 12px;
            border-bottom: 1px solid #ecf0f1;
        }
        tbody tr:hover {
            background-color: #f8f9fa;
        }
        .text-right {
            text-align: right;
        }
        .totals {
            margin-top: 30px;
            text-align: right;
        }
        .totals table {
            margin-left: auto;
            width: 300px;
        }
        .totals td {
            padding: 8px;
            border-bottom: none;
        }
        .total-row {
            font-size: 18px;
            font-weight: bold;
            border-top: 2px solid #2c3e50;
        }
        .footer {
            margin-top: 60px;
            padding-top: 20px;
            border-top: 1px solid #ecf0f1;
            font-size: 11px;
            color: #7f8c8d;
            text-align: center;
        }
        .payment-info {
            margin: 30px 0;
            padding: 15px;
            background-color: #e8f5e9;
            border-left: 4px solid #4caf50;
        }
        .payment-label {
            font-weight: bold;
            color: #2e7d32;
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="company-info">
            <div class="company-name">{{.company_name}}</div>
            <div class="company-details">
                {{.company_address}}<br>
                {{.company_city}}, {{.company_zip}}<br>
                Tel: {{.company_phone}} | Email: {{.company_email}}
            </div>
        </div>
        <div class="invoice-info">
            <div class="invoice-title">RECHNUNG</div>
            <div class="invoice-number">Nr. {{.invoice_number}}</div>
        </div>
    </div>

    <div class="client-info">
        <div class="client-label">Rechnungsempfänger</div>
        <div class="client-name">{{.client_name}}</div>
        <div>{{.client_address}}</div>
        <div>{{.client_zip}} {{.client_city}}</div>
    </div>

    <div class="dates">
        <div><strong>Rechnungsdatum:</strong> {{.invoice_date}}</div>
        <div><strong>Fälligkeitsdatum:</strong> {{.due_date}}</div>
    </div>

    <table>
        <thead>
            <tr>
                <th>Beschreibung</th>
                <th class="text-right">Menge</th>
                <th class="text-right">Einzelpreis</th>
                <th class="text-right">Gesamt</th>
            </tr>
        </thead>
        <tbody>
            {{range .items}}
            <tr>
                <td>{{.description}}</td>
                <td class="text-right">{{.quantity}}</td>
                <td class="text-right">€ {{.unit_price}}</td>
                <td class="text-right">€ {{.total}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>

    <div class="totals">
        <table>
            <tr>
                <td>Zwischensumme:</td>
                <td class="text-right">€ {{.subtotal}}</td>
            </tr>
            <tr>
                <td>MwSt. ({{.tax_rate}}%):</td>
                <td class="text-right">€ {{.tax_amount}}</td>
            </tr>
            <tr class="total-row">
                <td>Gesamtbetrag:</td>
                <td class="text-right">€ {{.total_amount}}</td>
            </tr>
        </table>
    </div>

    <div class="payment-info">
        <div class="payment-label">Zahlungsinformationen</div>
        <div>Bankverbindung: IBAN {{.bank_iban}}</div>
        <div>Verwendungszweck: {{.invoice_number}}</div>
    </div>

    <div class="footer">
        <p>{{.company_name}} | {{.company_address}} | {{.company_city}} {{.company_zip}}</p>
        <p>Steuernummer: {{.tax_number}} | Geschäftsführer: {{.ceo_name}}</p>
        <p>Vielen Dank für Ihr Vertrauen!</p>
    </div>
</body>
</html>`

	req := &templateServices.CreateTemplateRequest{
		TenantID:     tenantID,
		TemplateType: "invoice",
		Name:         "Standard Therapy Invoice",
		Description:  "Standard invoice template for therapy sessions",
		Content:      htmlContent,
		Variables: []string{
			"invoice_number", "company_name", "company_address", "company_city",
			"company_zip", "company_phone", "company_email", "client_name",
			"client_address", "client_city", "client_zip", "invoice_date",
			"due_date", "items", "subtotal", "tax_rate", "tax_amount",
			"total_amount", "bank_iban", "tax_number", "ceo_name",
		},
		SampleData: map[string]interface{}{
			"invoice_number":  "INV-2025-001",
			"company_name":    "Therapiepraxis Mustermann",
			"company_address": "Musterstraße 123",
			"company_city":    "Heidelberg",
			"company_zip":     "69115",
			"company_phone":   "+49 6221 123456",
			"company_email":   "info@therapie-mustermann.de",
			"client_name":     "Max Musterklient",
			"client_address":  "Beispielweg 45",
			"client_city":     "Mannheim",
			"client_zip":      "68159",
			"invoice_date":    "26.12.2025",
			"due_date":        "26.01.2026",
			"subtotal":        "600.00",
			"tax_rate":        "19",
			"tax_amount":      "114.00",
			"total_amount":    "714.00",
			"bank_iban":       "DE89 3704 0044 0532 0130 00",
			"tax_number":      "12345/67890",
			"ceo_name":        "Dr. Maria Mustermann",
		},
		IsActive:  true,
		IsDefault: true,
	}

	return service.CreateTemplate(ctx, req)
}

// seedEmailTemplate creates a sample email template
func seedEmailTemplate(ctx context.Context, service *templateServices.TemplateService, tenantID uint) (*templateEntities.Template, error) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2c3e50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { padding: 10px; text-align: center; font-size: 12px; color: #777; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{subject}}</h1>
        </div>
        <div class="content">
            <p>Sehr geehrte/r {{recipient_name}},</p>
            <p>{{message_body}}</p>
            <p>Mit freundlichen Grüßen,<br>{{sender_name}}</p>
        </div>
        <div class="footer">
            <p>{{.company_name}} | {{.company_email}}</p>
        </div>
    </div>
</body>
</html>`

	req := &templateServices.CreateTemplateRequest{
		TenantID:     tenantID,
		TemplateType: "email",
		Name:         "Standard Email Template",
		Description:  "Standard email template for client communication",
		Content:      htmlContent,
		Variables:    []string{"subject", "recipient_name", "message_body", "sender_name", "company_name", "company_email"},
		SampleData: map[string]interface{}{
			"subject":        "Terminbestätigung",
			"recipient_name": "Max Mustermann",
			"message_body":   "Hiermit bestätigen wir Ihren Termin am 15.01.2026 um 14:00 Uhr.",
			"sender_name":    "Dr. Maria Mustermann",
			"company_name":   "Therapiepraxis Mustermann",
			"company_email":  "info@therapie-mustermann.de",
		},
		IsActive:  true,
		IsDefault: false,
	}

	return service.CreateTemplate(ctx, req)
}

// seedInvoiceNumbers generates sample invoice numbers
*/

// seedInvoiceNumbers generates sample invoice numbers
func seedInvoiceNumbers(ctx context.Context, service *baseAPI.InvoiceNumberService, tenantID uint) ([]string, error) {
	var invoiceNumbers []string

	// Use default invoice config
	config := invoiceNumberServices.DefaultInvoiceConfig()

	// Generate 5 invoice numbers for current month
	for i := 0; i < 5; i++ {
		result, err := service.GenerateInvoiceNumber(ctx, tenantID, 1, config) // Assuming first organization
		if err != nil {
			return invoiceNumbers, fmt.Errorf("failed to generate invoice number: %w", err)
		}

		invoiceNumbers = append(invoiceNumbers, result.InvoiceNumber)
		log.Printf("  ✓ Generated: %s", result.InvoiceNumber)
	}

	return invoiceNumbers, nil
}

/*
// DEPRECATED: Sample invoice PDF generation - kept for reference
// seedSampleInvoices generates sample invoice PDFs
func seedSampleInvoices(ctx context.Context, db *gorm.DB, pdfService *documentServices.PDFService, templateID uint, invoiceNumbers []string, tenantID, userID uint) error {
	// Get some clients for realistic invoice data
	var clients []entities.Client
	db.Limit(3).Find(&clients)

	if len(clients) == 0 {
		log.Println("⚠️  No clients found, skipping sample invoice generation")
		return nil
	}

	successCount := 0
	for i, invoiceNum := range invoiceNumbers {
		if i >= len(clients) {
			break // Only generate as many invoices as we have clients
		}

		client := clients[i]

		// Create invoice data
		invoiceData := map[string]interface{}{
			"invoice_number":  invoiceNum,
			"company_name":    "Therapiepraxis Mustermann",
			"company_address": "Musterstraße 123",
			"company_city":    "Heidelberg",
			"company_zip":     "69115",
			"company_phone":   "+49 6221 123456",
			"company_email":   "info@therapie-mustermann.de",
			"client_name":     fmt.Sprintf("%s %s", client.FirstName, client.LastName),
			"client_address":  client.StreetAddress,
			"client_city":     client.City,
			"client_zip":      client.Zip,
			"invoice_date":    time.Now().Format("02.01.2006"),
			"due_date":        time.Now().AddDate(0, 1, 0).Format("02.01.2006"),
			"items": []map[string]interface{}{
				{
					"description": "Therapiesitzung - " + client.TherapyTitle,
					"quantity":    "4",
					"unit_price":  "150.00",
					"total":       "600.00",
				},
			},
			"subtotal":     "600.00",
			"tax_rate":     "19",
			"tax_amount":   "114.00",
			"total_amount": "714.00",
			"bank_iban":    "DE89 3704 0044 0532 0130 00",
			"tax_number":   "12345/67890",
			"ceo_name":     "Dr. Maria Mustermann",
		}

		// Generate PDF using the template
		req := &documentServices.GeneratePDFFromTemplateRequest{
			TenantID:     tenantID,
			UserID:       userID,
			TemplateID:   templateID,
			Data:         invoiceData,
			Filename:     fmt.Sprintf("invoice-%s.pdf", invoiceNum),
			DocumentType: "invoice",
			Metadata: map[string]interface{}{
				"client_id":      client.ID,
				"invoice_number": invoiceNum,
				"reference_type": "client",
			},
			SaveDocument: true,
		}

		result, err := pdfService.GeneratePDFFromTemplate(ctx, req)
		if err != nil {
			log.Printf("  ❌ Failed to generate invoice %s: %v", invoiceNum, err)
			continue
		}

		successCount++
		var documentID uint
		if result.Document != nil {
			documentID = result.Document.ID
		}
		log.Printf("  ✓ Generated invoice PDF: %s (Document ID: %d, Size: %d bytes)",
			invoiceNum, documentID, result.SizeBytes)
	}

	log.Printf("📊 Successfully generated %d out of %d invoice PDFs", successCount, len(invoiceNumbers))
	return nil
}

// showDocumentsStatistics displays documents seeding statistics
func showDocumentsStatistics(db *gorm.DB) {
	log.Println("\n📊 Documents Module Statistics")
	log.Println("==============================")

	var totalTemplates, totalDocuments int64
	db.Model(&templateEntities.Template{}).Count(&totalTemplates)
	db.Model(&documentEntities.Document{}).Count(&totalDocuments)

	log.Printf("📝 Total templates: %d", totalTemplates)
	log.Printf("📄 Total documents: %d", totalDocuments)

	// Template breakdown
	if totalTemplates > 0 {
		log.Println("\n📋 Template Breakdown:")
		var templateTypes []struct {
			TemplateType string
			Count        int64
		}
		db.Model(&templateEntities.Template{}).
			Select("template_type, COUNT(*) as count").
			Group("template_type").
			Find(&templateTypes)

		for _, tt := range templateTypes {
			log.Printf("   %s: %d templates", tt.TemplateType, tt.Count)
		}
	}

	// Document breakdown
	if totalDocuments > 0 {
		log.Println("\n📁 Document Breakdown:")
		var documentTypes []struct {
			DocumentType string
			Count        int64
		}
		db.Model(&documentEntities.Document{}).
			Select("document_type, COUNT(*) as count").
			Group("document_type").
			Find(&documentTypes)

		for _, dt := range documentTypes {
			log.Printf("   %s: %d documents", dt.DocumentType, dt.Count)
		}
	}
}
*/

// seedSessions creates sessions linked to calendar entries with various statuses
// Creates sessions from 5 weeks ago to 10 weeks ahead
func seedSessions(db *gorm.DB) error {
	log.Println("📅 Seeding client sessions...")

	// Calculate date range: 5 weeks back to 10 weeks ahead (use UTC)
	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -35) // 5 weeks ago
	endDate := now.AddDate(0, 0, 70)    // 10 weeks ahead

	log.Printf("  📆 Creating sessions from %s to %s (current time: %s UTC)",
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), now.Format("2006-01-02 15:04"))

	// Get calendar entries that could be therapy sessions in the date range
	type CalendarEntry struct {
		ID         uint `gorm:"primarykey"`
		TenantID   uint
		UserID     uint
		CalendarID uint
		Title      string
		Type       string
		StartTime  *time.Time
		EndTime    *time.Time
	}

	var calendarEntries []CalendarEntry
	// Get therapy-type calendar entries within date range
	// Order randomly to get a good mix of past and future
	if err := db.Table("calendar_entries").
		Where("type = ? AND start_time IS NOT NULL AND start_time >= ? AND start_time <= ?",
					"therapy", startDate, endDate).
		Order("RANDOM()"). // Random order to get mix of dates
		Limit(150).        // Increased limit to cover more sessions
		Find(&calendarEntries).Error; err != nil {
		return fmt.Errorf("failed to fetch calendar entries: %w", err)
	}

	if len(calendarEntries) == 0 {
		log.Println("⚠️  No calendar entries found for session creation in date range")
		return nil
	}

	log.Printf("  📋 Found %d calendar entries in date range", len(calendarEntries))

	// Get active clients for session creation
	var clients []entities.Client
	if err := db.Where("status = ?", "active").Limit(10).Find(&clients).Error; err != nil {
		return fmt.Errorf("failed to fetch clients: %w", err)
	}

	if len(clients) == 0 {
		log.Println("⚠️  No active clients found for session creation")
		return nil
	}

	createdSessions := 0
	conductedCount := 0
	scheduledCount := 0
	canceledCount := 0
	sessionsWithoutCalendar := 0

	// Create sessions for entries, determining status based on date
	// IMPORTANT: Each session MUST have a calendar entry
	for i, entry := range calendarEntries {
		// Validate that calendar entry has required fields
		if entry.ID == 0 || entry.StartTime == nil {
			log.Printf("  ⚠️  Skipping invalid calendar entry (ID: %d)", entry.ID)
			sessionsWithoutCalendar++
			continue
		}

		// Assign clients round-robin
		client := clients[i%len(clients)]

		// Determine status based on date
		var status string
		isPast := entry.StartTime.Before(now)

		if isPast {
			// Past sessions are mostly conducted, with some canceled
			if i%8 == 0 { // Every 8th session is canceled
				status = "canceled"
				canceledCount++
			} else {
				status = "conducted"
				conductedCount++
			}
		} else {
			// Future sessions are scheduled
			status = "scheduled"
			scheduledCount++
		}

		// Calculate duration
		durationMin := 60
		if entry.StartTime != nil && entry.EndTime != nil {
			durationMin = int(entry.EndTime.Sub(*entry.StartTime).Minutes())
		}

		// Create session linked to calendar entry
		// The CalendarEntryID is required and links this session to the appointment
		session := entities.Session{
			TenantID:          entry.TenantID,
			ClientID:          client.ID,
			CalendarEntryID:   &entry.ID, // REQUIRED: Links session to calendar appointment
			OriginalDate:      *entry.StartTime,
			OriginalStartTime: *entry.StartTime,
			DurationMin:       durationMin,
			Type:              "therapy",
			NumberUnits:       1,
			Status:            status,
			Documentation:     generateSessionDocumentation(status, client.TherapyTitle),
		}

		if err := db.Create(&session).Error; err != nil {
			log.Printf("  ❌ Failed to create session for calendar entry %d: %v", entry.ID, err)
			sessionsWithoutCalendar++
			continue
		}

		createdSessions++

		// Log progress every 10 sessions
		if createdSessions%10 == 0 {
			log.Printf("  📊 Progress: %d sessions created...", createdSessions)
		}
	}

	// Verify all sessions have calendar entries
	if sessionsWithoutCalendar > 0 {
		log.Printf("⚠️  Warning: %d sessions could not be linked to calendar entries", sessionsWithoutCalendar)
	}

	log.Printf("✅ Created %d sessions (all linked to calendar entries: %d conducted, %d scheduled, %d canceled)",
		createdSessions,
		conductedCount,
		scheduledCount,
		canceledCount)

	return nil
}

// seedExtraEfforts creates sample extra efforts (unbilled work outside sessions)
func seedExtraEfforts(db *gorm.DB) error {
	log.Println("📝 Seeding extra efforts...")

	// Get active clients with conducted sessions
	var clients []entities.Client
	if err := db.Where("status = ?", "active").Limit(8).Find(&clients).Error; err != nil {
		return fmt.Errorf("failed to fetch clients: %w", err)
	}

	if len(clients) == 0 {
		log.Println("⚠️  No active clients found for extra effort creation")
		return nil
	}

	// Get some conducted sessions to link extra efforts to
	var sessions []entities.Session
	db.Where("status = ?", "conducted").Limit(20).Find(&sessions)

	// Extra effort types and sample descriptions
	effortTypes := []struct {
		Type        string
		Description []string
	}{
		{
			Type: "preparation",
			Description: []string{
				"Copied therapy materials and worksheets",
				"Prepared visual aids for next session",
				"Set up sensory integration equipment",
				"Organized therapy room for specialized activity",
			},
		},
		{
			Type: "consultation",
			Description: []string{
				"Phone consultation with school teacher",
				"Team meeting with other therapists",
				"Consultation with pediatrician regarding treatment plan",
				"Case discussion with supervisor",
			},
		},
		{
			Type: "parent_meeting",
			Description: []string{
				"Parent guidance session",
				"Progress report meeting with parents",
				"Home exercise instruction for parents",
				"Parent consultation about behavioral strategies",
			},
		},
		{
			Type: "documentation",
			Description: []string{
				"Updated progress notes and treatment plan",
				"Wrote detailed session report for cost provider",
				"Completed assessment documentation",
				"Prepared quarterly progress report",
			},
		},
		{
			Type: "other",
			Description: []string{
				"Attended training workshop on new therapy technique",
				"Administrative work for client file",
				"Coordinated with other service providers",
				"Research for specialized intervention",
			},
		},
	}

	now := time.Now().UTC()
	createdCount := 0
	unbilledCount := 0
	rand.Seed(now.UnixNano())

	// Create 2-5 extra efforts per client
	for _, client := range clients {
		effortsForClient := 2 + rand.Intn(4) // 2-5 efforts

		for i := 0; i < effortsForClient; i++ {
			// Random effort type
			effortTypeData := effortTypes[rand.Intn(len(effortTypes))]
			description := effortTypeData.Description[rand.Intn(len(effortTypeData.Description))]

			// Random date in the past 4 weeks
			daysAgo := rand.Intn(28)
			effortDate := now.AddDate(0, 0, -daysAgo)

			// Random duration: 15-60 minutes, favoring common durations
			durations := []int{15, 20, 30, 45, 60}
			durationMin := durations[rand.Intn(len(durations))]

			// Occasionally link to a session (30% chance)
			var sessionID *uint
			if len(sessions) > 0 && rand.Float32() < 0.3 {
				randomSession := sessions[rand.Intn(len(sessions))]
				sessionID = &randomSession.ID
			}

			// Most are billable and delivered (90%)
			billable := rand.Float32() < 0.9
			billingStatus := "delivered"
			if billable {
				unbilledCount++
			} else {
				billingStatus = "excluded"
			}

			extraEffort := entities.ExtraEffort{
				TenantID:      client.TenantID,
				ClientID:      client.ID,
				SessionID:     sessionID,
				EffortType:    effortTypeData.Type,
				EffortDate:    effortDate,
				DurationMin:   durationMin,
				Description:   description,
				Billable:      billable,
				BillingStatus: billingStatus,
				CreatedBy:     1, // Admin user
			}

			if err := db.Create(&extraEffort).Error; err != nil {
				log.Printf("  ❌ Failed to create extra effort for client %d: %v", client.ID, err)
				continue
			}

			createdCount++
		}
	}

	log.Printf("✅ Created %d extra efforts (%d delivered and billable)",
		createdCount,
		unbilledCount)

	return nil
}

// seedInvoices creates sample invoices with different payment statuses using the new invoice system
func seedInvoices(db *gorm.DB) error {
	log.Println("💰 Seeding invoices...")

	// Get conducted sessions that aren't already invoiced
	var conductedSessions []entities.Session
	if err := db.Where("status = ?", "conducted").
		Limit(15). // Get enough for 3 invoices with ~5 sessions each
		Find(&conductedSessions).Error; err != nil {
		return fmt.Errorf("failed to fetch conducted sessions: %w", err)
	}

	if len(conductedSessions) < 3 {
		log.Println("⚠️  Not enough conducted sessions for invoice creation")
		return nil
	}

	// Get tenant and user
	var tenant baseAPI.Tenant
	if err := db.First(&tenant).Error; err != nil {
		return fmt.Errorf("no tenant found: %w", err)
	}

	var user baseAPI.User
	if err := db.First(&user).Error; err != nil {
		return fmt.Errorf("no user found: %w", err)
	}

	// Get the first organization
	var organization baseAPI.Organization
	if err := db.Where("tenant_id = ?", tenant.ID).First(&organization).Error; err != nil {
		return fmt.Errorf("no organization found: %w", err)
	}

	// Create 3 invoices with different statuses
	invoiceConfigs := []struct {
		status         entities.InvoiceStatus
		numReminders   int
		paymentDate    *time.Time
		paymentRef     *string
		latestReminder *time.Time
		sentAt         *time.Time
		finalizedAt    *time.Time
		daysAgo        int
	}{
		{
			status:       entities.InvoiceStatusPaid,
			numReminders: 0,
			paymentDate:  timePtr(time.Now().AddDate(0, 0, -5)),
			paymentRef:   strPtr("SEPA-2026-001"),
			sentAt:       timePtr(time.Now().AddDate(0, 0, -30)),
			finalizedAt:  timePtr(time.Now().AddDate(0, 0, -30)),
			daysAgo:      30,
		},
		{
			status:       entities.InvoiceStatusSent,
			numReminders: 0,
			paymentDate:  nil,
			sentAt:       timePtr(time.Now().AddDate(0, 0, -10)),
			finalizedAt:  timePtr(time.Now().AddDate(0, 0, -10)),
			daysAgo:      10,
		},
		{
			status:         entities.InvoiceStatusOverdue,
			numReminders:   1,
			paymentDate:    nil,
			latestReminder: timePtr(time.Now().AddDate(0, 0, -3)),
			sentAt:         timePtr(time.Now().AddDate(0, 0, -25)),
			finalizedAt:    timePtr(time.Now().AddDate(0, 0, -25)),
			daysAgo:        25,
		},
	}

	createdInvoices := 0
	sessionsPerInvoice := len(conductedSessions) / len(invoiceConfigs)

	for i, config := range invoiceConfigs {
		// Get sessions for this invoice
		startIdx := i * sessionsPerInvoice
		endIdx := startIdx + sessionsPerInvoice
		if i == len(invoiceConfigs)-1 {
			endIdx = len(conductedSessions) // Last invoice gets remaining sessions
		}

		if startIdx >= len(conductedSessions) {
			break
		}

		invoiceSessions := conductedSessions[startIdx:endIdx]
		if len(invoiceSessions) == 0 {
			continue
		}

		// Get the client for the first session
		firstSession := invoiceSessions[0]
		var client entities.Client
		if err := db.Preload("CostProvider").First(&client, firstSession.ClientID).Error; err != nil {
			log.Printf("  ❌ Failed to fetch client %d: %v", firstSession.ClientID, err)
			continue
		}

		// Calculate totals
		unitPrice := 120.0
		if client.UnitPrice != nil {
			unitPrice = *client.UnitPrice
		}

		var subtotal float64
		var invoiceItems []entities.InvoiceItem

		// Create invoice items for each session
		for _, session := range invoiceSessions {
			sessionTotal := float64(session.NumberUnits) * unitPrice
			subtotal += sessionTotal

			invoiceItems = append(invoiceItems, entities.InvoiceItem{
				ItemType:         "session",
				SessionID:        &session.ID,
				Description:      fmt.Sprintf("Therapiesitzung - %s", session.OriginalDate.Format("02.01.2006")),
				NumberUnits:      float64(session.NumberUnits),
				UnitPrice:        unitPrice,
				TotalAmount:      sessionTotal,
				VATRate:          0.00,
				VATExempt:        true,
				VATExemptionText: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
				UnitDurationMin:  &session.DurationMin,
				IsEditable:       false,
			})
		}

		// VAT exempt for healthcare services
		taxAmount := 0.0
		totalAmount := subtotal

		// Generate invoice number based on status
		var invoiceNumber string
		if config.status == entities.InvoiceStatusDraft {
			invoiceNumber = "DRAFT"
		} else {
			invoiceNumber = fmt.Sprintf("2026-%04d", 1+createdInvoices)
		}

		// Calculate due date (14 days after invoice date)
		invoiceDate := time.Now().AddDate(0, 0, -config.daysAgo)
		// Note: Due date functionality needs to be added to Invoice entity in future

		// Create invoice
		invoice := entities.Invoice{
			TenantID:       tenant.ID,
			UserID:         user.ID,
			OrganizationID: organization.ID,
			InvoiceDate:    invoiceDate,
			InvoiceNumber:  invoiceNumber,
			SumAmount:      subtotal,
			TaxAmount:      taxAmount,
			TotalAmount:    totalAmount,
			Status:         config.status,
			NumReminders:   config.numReminders,
			PayedDate:      config.paymentDate,
			LatestReminder: config.latestReminder,
			EmailSentAt:    config.sentAt,
			FinalizedAt:    config.finalizedAt,
			IsCreditNote:   false,
		}

		if err := db.Create(&invoice).Error; err != nil {
			log.Printf("  ❌ Failed to create invoice: %v", err)
			continue
		}

		// Create invoice items and link sessions
		for idx, item := range invoiceItems {
			item.InvoiceID = invoice.ID
			if err := db.Create(&item).Error; err != nil {
				log.Printf("  ❌ Failed to create invoice item: %v", err)
				continue
			}

			// Update session billing status
			session := invoiceSessions[idx]
			session.Status = "billed"
			if err := db.Save(&session).Error; err != nil {
				log.Printf("  ❌ Failed to update session status: %v", err)
			}

			// Create client invoice linking
			clientInvoice := entities.ClientInvoice{
				InvoiceID:      invoice.ID,
				ClientID:       client.ID,
				CostProviderID: ptrUint(client.CostProviderID),
				SessionID:      ptrUint(&session.ID),
				InvoiceItemID:  ptrUint(&item.ID),
			}
			if err := db.Create(&clientInvoice).Error; err != nil {
				log.Printf("  ❌ Failed to create client invoice link: %v", err)
			}
		}

		statusLabel := string(config.status)
		if config.paymentDate != nil {
			statusLabel = fmt.Sprintf("%s (paid %s)", statusLabel, config.paymentDate.Format("2006-01-02"))
		} else if config.status == entities.InvoiceStatusOverdue {
			statusLabel = fmt.Sprintf("%s (%d reminders)", statusLabel, config.numReminders)
		}

		log.Printf("  ✓ Created invoice %s - %s - %d sessions - €%.2f",
			invoiceNumber, statusLabel, len(invoiceSessions), totalAmount)
		createdInvoices++
	}

	// Create a draft invoice with some extra efforts
	if len(conductedSessions) > sessionsPerInvoice*3 {
		log.Println("  📝 Creating draft invoice with unbilled items...")

		var unbilledSessions []entities.Session
		if err := db.Where("status = ?", "conducted").
			Limit(3).
			Find(&unbilledSessions).Error; err == nil && len(unbilledSessions) > 0 {

			firstSession := unbilledSessions[0]
			var client entities.Client
			if err := db.First(&client, firstSession.ClientID).Error; err == nil {
				unitPrice := 120.0
				if client.UnitPrice != nil {
					unitPrice = *client.UnitPrice
				}

				var subtotal float64
				var invoiceItems []entities.InvoiceItem

				for _, session := range unbilledSessions {
					sessionTotal := float64(session.NumberUnits) * unitPrice
					subtotal += sessionTotal

					invoiceItems = append(invoiceItems, entities.InvoiceItem{
						ItemType:         "session",
						SessionID:        &session.ID,
						Description:      fmt.Sprintf("Therapiesitzung - %s", session.OriginalDate.Format("02.01.2006")),
						NumberUnits:      float64(session.NumberUnits),
						UnitPrice:        unitPrice,
						TotalAmount:      sessionTotal,
						VATRate:          0.00,
						VATExempt:        true,
						VATExemptionText: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
						IsEditable:       false,
					})
				}

				draftInvoice := entities.Invoice{
					TenantID:       tenant.ID,
					UserID:         user.ID,
					OrganizationID: organization.ID,
					InvoiceDate:    time.Now(),
					InvoiceNumber:  "DRAFT",
					SumAmount:      subtotal,
					TaxAmount:      0.0,
					TotalAmount:    subtotal,
					Status:         entities.InvoiceStatusDraft,
					IsCreditNote:   false,
				}

				if err := db.Create(&draftInvoice).Error; err == nil {
					for _, item := range invoiceItems {
						item.InvoiceID = draftInvoice.ID
						db.Create(&item)
					}
					log.Printf("  ✓ Created draft invoice - %d sessions - €%.2f", len(unbilledSessions), subtotal)
					createdInvoices++
				}
			}
		}
	}

	log.Printf("✅ Created %d invoices", createdInvoices)

	// Create additional conducted sessions that remain unbilled for testing unbilled sessions endpoint
	log.Println("  📝 Creating additional unbilled conducted sessions...")

	var availableClients []entities.Client
	if err := db.Where("tenant_id = ? AND status = ?", tenant.ID, "active").
		Limit(5).
		Find(&availableClients).Error; err == nil && len(availableClients) > 0 {

		conductedCount := 0
		for _, client := range availableClients {
			if conductedCount >= 8 { // Create 8 additional conducted sessions
				break
			}

			// Create 1-2 conducted sessions per client
			for j := 0; j < 2 && conductedCount < 8; j++ {
				conductedDate := time.Now().AddDate(0, 0, -(conductedCount+1)*2) // Spread over past days

				conductedSession := entities.Session{
					TenantID:          tenant.ID,
					ClientID:          client.ID,
					CalendarEntryID:   nil, // No calendar entry needed for this test data
					OriginalDate:      conductedDate,
					OriginalStartTime: conductedDate,
					DurationMin:       45,
					Type:              "therapy",
					NumberUnits:       1,
					Status:            "conducted", // Keep as conducted, not billed
					Documentation:     fmt.Sprintf("Unbilled therapy session for %s %s - %s", client.FirstName, client.LastName, conductedDate.Format("02.01.2006")),
				}

				if err := db.Create(&conductedSession).Error; err != nil {
					log.Printf("  ❌ Failed to create unbilled session: %v", err)
					continue
				}
				conductedCount++
			}
		}

		if conductedCount > 0 {
			log.Printf("  ✓ Created %d additional conducted sessions (unbilled)", conductedCount)
		}
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}

func ptrUint(p *uint) uint {
	if p != nil {
		return *p
	}
	return 0
}

func intPtr(i int) *int {
	return &i
}

// Helper functions
func generateSessionDocumentation(status, therapyTitle string) string {
	switch status {
	case "conducted":
		docs := []string{
			fmt.Sprintf("Therapiesitzung durchgeführt: %s. Gute Fortschritte.", therapyTitle),
			fmt.Sprintf("Reguläre Sitzung %s. Ziele erreicht.", therapyTitle),
			fmt.Sprintf("Behandlung %s erfolgreich durchgeführt.", therapyTitle),
		}
		return docs[rand.Intn(len(docs))]
	case "canceled":
		return "Termin vom Klienten abgesagt"
	case "scheduled":
		return ""
	default:
		return ""
	}
}

func countSessionsByStatus(statuses []string, status string) int {
	count := 0
	for _, s := range statuses {
		if s == status {
			count++
		}
	}
	return count
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// showSessionAndInvoiceStatistics displays session and invoice seeding statistics
func showSessionAndInvoiceStatistics(db *gorm.DB) {
	log.Println("\n📊 Session, Extra Effort & Invoice Statistics")
	log.Println("==============================================")

	// Session statistics
	var totalSessions int64
	db.Model(&entities.Session{}).Count(&totalSessions)
	log.Printf("📅 Total sessions: %d", totalSessions)

	if totalSessions > 0 {
		var sessionStatuses []struct {
			Status string
			Count  int64
		}
		db.Model(&entities.Session{}).
			Select("status, COUNT(*) as count").
			Group("status").
			Order("count DESC").
			Find(&sessionStatuses)

		log.Println("\n🎯 Session Status Breakdown:")
		for _, ss := range sessionStatuses {
			log.Printf("   %s: %d sessions", ss.Status, ss.Count)
		}
	}

	// Extra effort statistics
	var totalExtraEfforts int64
	db.Model(&entities.ExtraEffort{}).Count(&totalExtraEfforts)
	log.Printf("\n📝 Total extra efforts: %d", totalExtraEfforts)

	if totalExtraEfforts > 0 {
		var effortTypes []struct {
			EffortType string
			Count      int64
		}
		db.Model(&entities.ExtraEffort{}).
			Select("effort_type, COUNT(*) as count").
			Group("effort_type").
			Order("count DESC").
			Find(&effortTypes)

		log.Println("\n📋 Extra Effort Type Breakdown:")
		for _, et := range effortTypes {
			log.Printf("   %s: %d efforts", et.EffortType, et.Count)
		}

		var billingStatuses []struct {
			BillingStatus string
			Count         int64
		}
		db.Model(&entities.ExtraEffort{}).
			Select("billing_status, COUNT(*) as count").
			Group("billing_status").
			Order("count DESC").
			Find(&billingStatuses)

		log.Println("\n💵 Extra Effort Billing Status:")
		for _, bs := range billingStatuses {
			log.Printf("   %s: %d efforts", bs.BillingStatus, bs.Count)
		}
	}

	// Invoice statistics
	var totalInvoices int64
	db.Model(&entities.Invoice{}).Count(&totalInvoices)
	log.Printf("\n💰 Total invoices: %d", totalInvoices)

	if totalInvoices > 0 {
		var invoiceStatuses []struct {
			Status string
			Count  int64
		}
		db.Model(&entities.Invoice{}).
			Select("status, COUNT(*) as count").
			Group("status").
			Order("count DESC").
			Find(&invoiceStatuses)

		log.Println("\n💳 Invoice Status Breakdown:")
		for _, is := range invoiceStatuses {
			log.Printf("   %s: %d invoices", is.Status, is.Count)
		}

		// Invoice items
		var totalInvoiceItems int64
		db.Model(&entities.InvoiceItem{}).Count(&totalInvoiceItems)
		log.Printf("\n📋 Total invoice items: %d", totalInvoiceItems)
	}
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
