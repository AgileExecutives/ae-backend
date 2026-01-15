// Package database provides base-server specific database seeding functionality.
//
// This package contains seeding logic specific to ae-base-server, including
// loading and creating initial tenant, plan, and user data from seed-data.json.
//
// For general database connection utilities, see the pkg/database package.
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/internal/services"
	"github.com/ae-base-server/modules/templates/entities"
	"github.com/ae-base-server/modules/templates/services/storage"
	pkgServices "github.com/ae-base-server/pkg/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SeedData represents the structure of seed data from JSON
type SeedData struct {
	Customers     []SeedCustomer     `json:"customers"`
	Tenants       []SeedTenant       `json:"tenants"`
	Organizations []SeedOrganization `json:"organizations"`
	Plans         []SeedPlan         `json:"plans"`
	Users         []SeedUser         `json:"users"`
}

// TemplateSeedData represents template seeding data structure
type TemplateSeedData struct {
	Templates []SeedTemplate `json:"templates"`
}

// SeedTemplate represents template seed data
type SeedTemplate struct {
	Module         string `json:"module"`
	TemplateKey    string `json:"template_key"`
	Channel        string `json:"channel"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Subject        string `json:"subject,omitempty"`
	FilePath       string `json:"file_path"`
	OrganizationID *uint  `json:"organization_id,omitempty"`
	TenantID       *uint  `json:"tenant_id,omitempty"`
}

// SeedCustomer represents customer seed data
type SeedCustomer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SeedTenant represents tenant seed data
type SeedTenant struct {
	CustomerID uint   `json:"customer_id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
}

// SeedOrganization represents organization seed data
type SeedOrganization struct {
	TenantID         uint     `json:"tenant_id"`
	Name             string   `json:"name"`
	OwnerName        string   `json:"owner_name"`
	OwnerTitle       string   `json:"owner_title"`
	StreetAddress    string   `json:"street_address"`
	Zip              string   `json:"zip"`
	City             string   `json:"city"`
	Email            string   `json:"email"`
	Phone            string   `json:"phone"`
	TaxID            string   `json:"tax_id"`
	TaxRate          *float64 `json:"tax_rate"`
	TaxUstID         string   `json:"tax_ustid"`
	UnitPrice        *float64 `json:"unit_price"`
	BankAccountOwner string   `json:"bank_account_owner"`
	BankAccountBank  string   `json:"bank_account_bank"`
	BankAccountBIC   string   `json:"bank_account_bic"`
	BankAccountIBAN  string   `json:"bank_account_iban"`
}

// SeedPlan represents plan seed data
type SeedPlan struct {
	Name          string                 `json:"name"`
	Slug          string                 `json:"slug"`
	Description   string                 `json:"description"`
	Price         float64                `json:"price"`
	Currency      string                 `json:"currency"`
	InvoicePeriod string                 `json:"invoice_period"`
	MaxUsers      int                    `json:"max_users"`
	MaxClients    int                    `json:"max_clients"`
	Features      map[string]interface{} `json:"features"`
	Active        bool                   `json:"active"`
}

// SeedUser represents user seed data
type SeedUser struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Role           string `json:"role"`
	Active         bool   `json:"active"`
	TenantSlug     string `json:"tenant_slug"`
	OrganizationID uint   `json:"organization_id"`
	EmailVerified  bool   `json:"email_verified"`
}

// loadSeedData loads seed data from JSON file in startupseed folder
func loadSeedData() (*SeedData, error) {
	// Get the current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for seed-data.json in startupseed folder
	seedDataPath := filepath.Join(pwd, "startupseed", "seed-data.json")
	if _, err := os.Stat(seedDataPath); os.IsNotExist(err) {
		// Try parent directory (in case running from subdirectory)
		seedDataPath = filepath.Join(filepath.Dir(pwd), "startupseed", "seed-data.json")
		if _, err := os.Stat(seedDataPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("seed-data.json not found in startupseed folder")
		}
	}

	log.Printf("Loading seed data from: %s", seedDataPath)

	// Read the JSON file
	data, err := os.ReadFile(seedDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed-data.json: %w", err)
	}

	// Parse JSON data
	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("failed to parse seed-data.json: %w", err)
	}

	return &seedData, nil
}

// Seed adds initial data to the database
func Seed(db *gorm.DB, tenantService *services.TenantService) error {
	log.Println("Seeding database with initial data...")

	// Load seed data from JSON file
	seedData, err := loadSeedData()
	if err != nil {
		return fmt.Errorf("failed to load seed data: %w", err)
	}

	// Create customers first
	var customerCount int64
	db.Model(&models.Customer{}).Count(&customerCount)
	if customerCount == 0 {
		for _, customerData := range seedData.Customers {
			customer := models.Customer{
				Name:   customerData.Name,
				Email:  customerData.Email,
				Active: true,
				Status: "active",
			}
			if err := db.Create(&customer).Error; err != nil {
				return fmt.Errorf("failed to create customer %s: %w", customerData.Name, err)
			}
			log.Printf("Created customer: %s", customerData.Name)
		}
	}

	// Create tenants with MinIO buckets
	var tenantCount int64
	db.Model(&models.Tenant{}).Count(&tenantCount)
	if tenantCount == 0 {
		for _, tenantData := range seedData.Tenants {
			// First create tenant in database (needed for CustomerID foreign key)
			tenant := models.Tenant{
				CustomerID: tenantData.CustomerID,
				Name:       tenantData.Name,
				Slug:       tenantData.Slug,
			}
			if err := db.Create(&tenant).Error; err != nil {
				return fmt.Errorf("failed to create tenant %s: %w", tenantData.Name, err)
			}

			// Then ensure MinIO bucket exists for this tenant
			if err := tenantService.EnsureTenantBucket(context.Background(), tenant.ID); err != nil {
				log.Printf("⚠️ Warning: Failed to create MinIO bucket for tenant %s (ID: %d): %v", tenantData.Name, tenant.ID, err)
				// Don't fail the entire seed process if bucket creation fails
			}

			log.Printf("Created tenant with bucket: %s (ID: %d)", tenantData.Name, tenant.ID)
		}
	}

	// Create organizations
	var organizationCount int64
	db.Model(&models.Organization{}).Count(&organizationCount)
	if organizationCount == 0 {
		for _, orgData := range seedData.Organizations {
			organization := models.Organization{
				TenantID:         orgData.TenantID,
				Name:             orgData.Name,
				OwnerName:        orgData.OwnerName,
				OwnerTitle:       orgData.OwnerTitle,
				StreetAddress:    orgData.StreetAddress,
				Zip:              orgData.Zip,
				City:             orgData.City,
				Email:            orgData.Email,
				Phone:            orgData.Phone,
				TaxID:            orgData.TaxID,
				TaxRate:          orgData.TaxRate,
				TaxUstID:         orgData.TaxUstID,
				UnitPrice:        orgData.UnitPrice,
				BankAccountOwner: orgData.BankAccountOwner,
				BankAccountBank:  orgData.BankAccountBank,
				BankAccountBIC:   orgData.BankAccountBIC,
				BankAccountIBAN:  orgData.BankAccountIBAN,
			}
			if err := db.Create(&organization).Error; err != nil {
				return fmt.Errorf("failed to create organization %s: %w", orgData.Name, err)
			}
			log.Printf("Created organization: %s", orgData.Name)

			// Note: Templates will be seeded by the templates module after initialization
		}
	}

	// Create plans
	var planCount int64
	db.Model(&models.Plan{}).Count(&planCount)
	if planCount == 0 {
		for _, planData := range seedData.Plans {
			// Convert features map to JSON string
			featuresJSON, err := json.Marshal(planData.Features)
			if err != nil {
				return fmt.Errorf("failed to marshal features for plan %s: %w", planData.Name, err)
			}

			plan := models.Plan{
				Name:          planData.Name,
				Slug:          planData.Slug,
				Description:   planData.Description,
				Price:         planData.Price,
				Currency:      planData.Currency,
				InvoicePeriod: planData.InvoicePeriod,
				MaxUsers:      planData.MaxUsers,
				MaxClients:    planData.MaxClients,
				Features:      string(featuresJSON),
				Active:        planData.Active,
			}
			if err := db.Create(&plan).Error; err != nil {
				return fmt.Errorf("failed to create plan %s: %w", planData.Name, err)
			}
			log.Printf("Created plan: %s", planData.Name)
		}
	}

	// Create users
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		for _, userData := range seedData.Users {
			// Find the tenant by slug
			var tenant models.Tenant
			if err := db.Where("slug = ?", userData.TenantSlug).First(&tenant).Error; err != nil {
				return fmt.Errorf("failed to find tenant with slug %s: %w", userData.TenantSlug, err)
			}

			// Hash the password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password for user %s: %w", userData.Username, err)
			}

			user := models.User{
				Username:       userData.Username,
				Email:          userData.Email,
				PasswordHash:   string(hashedPassword),
				FirstName:      userData.FirstName,
				LastName:       userData.LastName,
				TenantID:       tenant.ID,
				OrganizationID: userData.OrganizationID,
				Role:           userData.Role,
				Active:         userData.Active,
				EmailVerified:  userData.EmailVerified,
			}

			// Set EmailVerifiedAt if email is verified
			if userData.EmailVerified {
				now := db.NowFunc()
				user.EmailVerifiedAt = &now
			}

			if err := db.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user %s: %w", userData.Username, err)
			}
			log.Printf("Created user: %s (email_verified: %v)", userData.Username, userData.EmailVerified)
		}
	}

	// Seed templates from startupseed directory
	err = seedTemplates(db, tenantService)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to seed templates: %v", err)
		// Don't fail the entire seed process if template seeding fails
	}

	log.Println("Database seeding completed successfully! 🎉")
	return nil
}

// seedTemplates seeds templates from startupseed directory
func seedTemplates(db *gorm.DB, tenantService *services.TenantService) error {
	log.Println("Seeding templates from startupseed directory...")

	// Load template seed data
	templateSeedData, err := loadTemplateSeedData()
	if err != nil {
		return fmt.Errorf("failed to load template seed data: %w", err)
	}

	// Initialize MinIO storage for template storage
	minioConfig := storage.MinIOConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin123",
		UseSSL:          false,
		Region:          "us-east-1",
	}
	minioStorage, err := storage.NewMinIOStorage(minioConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize MinIO storage: %w", err)
	}

	// Create unified storage service for template storage
	storageService := pkgServices.NewStorageService(minioStorage)

	// Get all tenants to seed templates for each
	var tenants []models.Tenant
	if err := db.Find(&tenants).Error; err != nil {
		return fmt.Errorf("failed to retrieve tenants: %w", err)
	}

	// Seed templates based on seed data (check per-template to allow re-seeding individual templates)
	log.Printf("Seeding templates from seed data...")

	for _, templateData := range templateSeedData.Templates {
		// Use tenant_id from template data, or skip if not specified
		if templateData.TenantID == nil {
			log.Printf("⚠️ Warning: Template %s has no tenant_id specified, skipping", templateData.Name)
			continue
		}

		tenantID := *templateData.TenantID

		// Verify that the tenant exists
		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			log.Printf("⚠️ Warning: Tenant ID %d not found for template %s, skipping", tenantID, templateData.Name)
			continue
		}

		// Check if this specific template already exists for this tenant
		var existingTemplate entities.Template
		err := db.Where("tenant_id = ? AND template_key = ? AND channel = ?", tenantID, templateData.TemplateKey, templateData.Channel).First(&existingTemplate).Error
		if err == nil {
			log.Printf("  ⏭️  Template %s/%s already exists for tenant %d, skipping", templateData.Channel, templateData.TemplateKey, tenantID)
			continue
		}

		err = createTemplateFromSeed(db, storageService, tenantID, templateData)
		if err != nil {
			log.Printf("⚠️ Warning: Failed to create template %s for tenant %d: %v",
				templateData.Name, tenantID, err)
			// Continue with other templates
			continue
		}
		log.Printf("✅ Created template: %s for tenant %s (ID: %d)", templateData.Name, tenant.Name, tenantID)
	}

	log.Printf("Template seeding completed for %d tenants", len(tenants))
	return nil
}

// loadTemplateSeedData loads template seed data from startupseed/templates_seed.json
func loadTemplateSeedData() (*TemplateSeedData, error) {
	// Get the base directory - from base-server directory, go to startupseed
	baseDir := "startupseed"
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		// Try relative path from different locations
		possiblePaths := []string{
			"./startupseed",
			"../startupseed",
			"../../base-server/startupseed",
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				baseDir = path
				break
			}
		}
	}

	seedFilePath := filepath.Join(baseDir, "templates_seed.json")
	data, err := ioutil.ReadFile(seedFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template seed file %s: %w", seedFilePath, err)
	}

	var templateSeedData TemplateSeedData
	if err := json.Unmarshal(data, &templateSeedData); err != nil {
		return nil, fmt.Errorf("failed to parse template seed data: %w", err)
	}

	return &templateSeedData, nil
}

// createTemplateFromSeed creates a template from seed data and stores it using storage service
func createTemplateFromSeed(db *gorm.DB, storageService *pkgServices.StorageService,
	tenantID uint, templateData SeedTemplate) error {

	// Tenant ID is already correctly passed from the calling function
	// Use organization_id from seed data if available

	// Read template content from file
	baseDir := "startupseed"
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		possiblePaths := []string{
			"./startupseed",
			"../startupseed",
			"../../base-server/startupseed",
		}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				baseDir = path
				break
			}
		}
	}

	templateFilePath := filepath.Join(baseDir, templateData.FilePath)
	content, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", templateFilePath, err)
	}

	// Generate storage key with tenant ID for uniqueness
	storageKey := fmt.Sprintf("templates/%s/%s_%s_%d_t%d.html",
		templateData.Channel,
		templateData.Module,
		templateData.TemplateKey,
		time.Now().Unix(),
		tenantID,
	)

	// Store template content using storage service
	err = storageService.StoreTemplateWithKey(context.Background(), tenantID, storageKey, string(content), map[string]string{
		"module":       templateData.Module,
		"template_key": templateData.TemplateKey,
		"channel":      templateData.Channel,
		"seeded":       "true",
	})
	if err != nil {
		return fmt.Errorf("failed to store template content: %w", err)
	}

	// Get contract data for variables and sample data
	log.Printf("🔍 Looking for contract: module=%s, template_key=%s", templateData.Module, templateData.TemplateKey)
	variablesJSON, sampleDataJSON, err := getContractData(db, templateData.Module, templateData.TemplateKey)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get contract data for %s.%s: %v", templateData.Module, templateData.TemplateKey, err)
		// Use empty defaults
		variablesJSON = datatypes.JSON([]byte("[]"))
		sampleDataJSON = datatypes.JSON([]byte("{}"))
	} else {
		log.Printf("✅ Retrieved contract data for %s.%s - variables: %d bytes, sample_data: %d bytes",
			templateData.Module, templateData.TemplateKey, len(variablesJSON), len(sampleDataJSON))
	}

	// Create template record in database
	template := entities.Template{
		TenantID:    tenantID,
		Module:      templateData.Module,
		TemplateKey: templateData.TemplateKey,
		Channel:     entities.Channel(templateData.Channel),
		Name:        templateData.Name,
		Description: templateData.Description,
		StorageKey:  storageKey,
		Version:     1,
		IsActive:    true,
		IsDefault:   true,           // Mark seeded templates as default
		Variables:   variablesJSON,  // From contract
		SampleData:  sampleDataJSON, // From contract
	}

	// Set organization_id from seed data if available
	if templateData.OrganizationID != nil {
		// Verify that the organization exists and belongs to the correct tenant
		var organization models.Organization
		if err := db.Where("id = ? AND tenant_id = ?", *templateData.OrganizationID, tenantID).First(&organization).Error; err != nil {
			return fmt.Errorf("organization ID %d not found for tenant %d: %w", *templateData.OrganizationID, tenantID, err)
		}
		template.OrganizationID = templateData.OrganizationID
		log.Printf("✅ Template %s assigned to organization: %s (ID: %d)", templateData.Name, organization.Name, *templateData.OrganizationID)
	}

	// Set subject for EMAIL templates
	if templateData.Channel == "EMAIL" && templateData.Subject != "" {
		template.Subject = &templateData.Subject
	}

	// Legacy template type mapping
	template.TemplateType = templateData.TemplateKey

	if err := db.Create(&template).Error; err != nil {
		return fmt.Errorf("failed to create template record: %w", err)
	}

	return nil
}

// getContractData retrieves variables and sample data from the template contract
func getContractData(db *gorm.DB, module, templateKey string) (datatypes.JSON, datatypes.JSON, error) {
	var contract entities.TemplateContract
	err := db.Where("module = ? AND template_key = ?", module, templateKey).First(&contract).Error
	if err != nil {
		return nil, nil, fmt.Errorf("contract not found: %w", err)
	}

	log.Printf("🔍 Contract found for %s.%s - VariableSchema: %d bytes, DefaultSampleData: %d bytes",
		module, templateKey, len(contract.VariableSchema), len(contract.DefaultSampleData))

	// Extract variable schema as variables (convert schema to variable list if needed)
	variablesJSON := contract.VariableSchema
	if len(variablesJSON) == 0 {
		variablesJSON = datatypes.JSON([]byte("[]"))
	}

	// Use default sample data from contract
	sampleDataJSON := contract.DefaultSampleData
	if len(sampleDataJSON) == 0 {
		sampleDataJSON = datatypes.JSON([]byte("{}"))
	}

	return variablesJSON, sampleDataJSON, nil
}
