package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ae-base-server/modules/templates/entities"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SeedService handles template seeding from static files
type SeedService struct {
	db              *gorm.DB
	templateService *TemplateService
}

// NewSeedService creates a new seed service
func NewSeedService(db *gorm.DB) *SeedService {
	return &SeedService{
		db: db,
	}
}

// NewSeedServiceWithTemplateService creates a seed service with template service for instance creation
func NewSeedServiceWithTemplateService(db *gorm.DB, templateService *TemplateService) *SeedService {
	return &SeedService{
		db:              db,
		templateService: templateService,
	}
}

// TemplateDefinition defines a template contract to be seeded
type TemplateDefinition struct {
	Module            string
	TemplateKey       string
	Description       string
	SupportedChannels []string
	VariableSchema    map[string]interface{}
	DefaultSampleData map[string]interface{}
}

// getTemplateData loads sample data from file with fallback to default
func (s *SeedService) getTemplateData(fileName string, defaultData map[string]interface{}) map[string]interface{} {
	if fileSampleData, err := s.loadSampleDataFromFile(fileName); err == nil {
		return fileSampleData
	}
	return defaultData
}

// loadSampleDataFromFile loads sample data from a JSON file
func (s *SeedService) loadSampleDataFromFile(fileName string) (map[string]interface{}, error) {
	filePath := filepath.Join("startupseed", fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sample data file %s: %w", filePath, err)
	}

	var sampleData map[string]interface{}
	if err := json.Unmarshal(data, &sampleData); err != nil {
		return nil, fmt.Errorf("failed to parse sample data from %s: %w", filePath, err)
	}

	return sampleData, nil
}

// SeedDefaultTemplates seeds default template contracts if none exist
func (s *SeedService) SeedDefaultTemplates() error { // First ensure table exists by running auto-migration
	if err := s.db.AutoMigrate(&entities.TemplateContract{}); err != nil {
		return fmt.Errorf("failed to migrate template contracts table: %w", err)
	}
	// Check if any template contracts exist
	var count int64
	if err := s.db.Model(&entities.TemplateContract{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count template contracts: %w", err)
	}

	if count > 0 {
		fmt.Printf("Template contracts already exist (%d found), skipping seeding\n", count)
		return nil
	}

	fmt.Println("No template contracts found, seeding defaults...")

	definitions := s.getTemplateDefinitions()

	for _, def := range definitions {
		if err := s.seedTemplate(def); err != nil {
			return fmt.Errorf("failed to seed template contract %s.%s: %w", def.Module, def.TemplateKey, err)
		}
		fmt.Printf("✓ Seeded template contract: %s.%s\n", def.Module, def.TemplateKey)
	}

	fmt.Printf("Successfully seeded %d template contracts\n", len(definitions))
	return nil
}

// getTemplateDefinitions returns all template definitions to seed
func (s *SeedService) getTemplateDefinitions() []TemplateDefinition {
	// Load sample data from files, fallback to hardcoded data
	invoiceSampleData := s.getTemplateData("invoice.json", map[string]interface{}{
		"invoice_number": "INV-2024-001",
		"invoice_date":   "2024-01-15",
		"organization": map[string]interface{}{
			"name": "Sample Organization GmbH",
		},
		"client": map[string]interface{}{
			"first_name": "Max",
			"last_name":  "Mustermann",
		},
		"totals": map[string]interface{}{
			"net_total":   1500.00,
			"tax_rate":    19.0,
			"gross_total": 1785.00,
		},
	})

	passwordResetSampleData := s.getTemplateData("password_reset.json", map[string]interface{}{
		"AppName":       "AE SaaS Platform",
		"RecipientName": "John Doe",
		"CompanyName":   "AE Technology Solutions",
		"ResetURL":      "https://app.example.com/reset-password?token=abc123",
		"SupportEmail":  "support@example.com",
	})

	welcomeSampleData := s.getTemplateData("welcome.json", map[string]interface{}{
		"FirstName":        "John",
		"LastName":         "Doe",
		"OrganizationName": "AE Technology Solutions",
		"Email":            "john.doe@example.com",
		"Username":         "johndoe",
		"PlanName":         "Professional Plan",
		"LoginURL":         "https://app.example.com/login",
	})

	emailVerificationSampleData := s.getTemplateData("email_verification.json", map[string]interface{}{
		"FirstName":        "Jane",
		"OrganizationName": "AE Technology Solutions",
		"Email":            "jane.smith@example.com",
		"Username":         "janesmith",
		"VerificationURL":  "https://app.example.com/verify-email?token=xyz789",
		"VerificationCode": "481592",
	})

	bookingConfirmationSampleData := s.getTemplateData("booking_confirmation.json", map[string]interface{}{
		"AppName":        "Unburdy Therapy Center",
		"Date":           "15. Februar 2026",
		"TimeFrom":       "14:00",
		"TimeTo":         "15:00",
		"CompanyName":    "Unburdy Therapiezentrum GmbH",
		"SupportEmail":   "info@unburdy.de",
		"CompanyAddress": "Musterstraße 123, 12345 Berlin",
	})

	return []TemplateDefinition{
		// Billing: Invoice
		{
			Module:            "billing",
			TemplateKey:       "invoice",
			Description:       "Invoice template for client billing",
			SupportedChannels: []string{"DOCUMENT"},
			VariableSchema: map[string]interface{}{
				"invoice_number": map[string]interface{}{"type": "string", "required": true},
				"invoice_date":   map[string]interface{}{"type": "string", "required": true},
				"organization":   map[string]interface{}{"type": "object", "required": true},
				"client":         map[string]interface{}{"type": "object", "required": true},
				"cost_provider":  map[string]interface{}{"type": "object", "required": true},
				"invoice_items":  map[string]interface{}{"type": "array", "required": true},
				"totals":         map[string]interface{}{"type": "object", "required": true},
			},
			DefaultSampleData: invoiceSampleData,
		},
		// Identity: Password Reset
		{
			Module:            "identity",
			TemplateKey:       "password_reset",
			Description:       "Password reset email template",
			SupportedChannels: []string{"EMAIL"},
			VariableSchema: map[string]interface{}{
				"AppName":       map[string]interface{}{"type": "string", "required": true},
				"RecipientName": map[string]interface{}{"type": "string", "required": true},
				"CompanyName":   map[string]interface{}{"type": "string", "required": true},
				"ResetURL":      map[string]interface{}{"type": "string", "required": true},
				"SupportEmail":  map[string]interface{}{"type": "string", "required": true},
			},
			DefaultSampleData: passwordResetSampleData,
		},
		// Identity: Email Verification
		{
			Module:            "identity",
			TemplateKey:       "email_verification",
			Description:       "Email verification template",
			SupportedChannels: []string{"EMAIL"},
			VariableSchema: map[string]interface{}{
				"FirstName":        map[string]interface{}{"type": "string", "required": true},
				"OrganizationName": map[string]interface{}{"type": "string", "required": true},
				"Email":            map[string]interface{}{"type": "string", "required": true},
				"Username":         map[string]interface{}{"type": "string", "required": true},
				"VerificationURL":  map[string]interface{}{"type": "string", "required": true},
				"VerificationCode": map[string]interface{}{"type": "string", "required": true},
			},
			DefaultSampleData: emailVerificationSampleData,
		},
		// Identity: Welcome Email
		{
			Module:            "identity",
			TemplateKey:       "welcome",
			Description:       "Welcome email for new users",
			SupportedChannels: []string{"EMAIL"},
			VariableSchema: map[string]interface{}{
				"FirstName":        map[string]interface{}{"type": "string", "required": true},
				"LastName":         map[string]interface{}{"type": "string", "required": true},
				"OrganizationName": map[string]interface{}{"type": "string", "required": true},
				"Email":            map[string]interface{}{"type": "string", "required": true},
				"Username":         map[string]interface{}{"type": "string", "required": true},
				"PlanName":         map[string]interface{}{"type": "string", "required": true},
				"LoginURL":         map[string]interface{}{"type": "string", "required": true},
			},
			DefaultSampleData: welcomeSampleData,
		},
		// Notification: Booking Confirmation
		{
			Module:            "notification",
			TemplateKey:       "booking_confirmation",
			Description:       "Booking confirmation email",
			SupportedChannels: []string{"EMAIL"},
			VariableSchema: map[string]interface{}{
				"AppName":        map[string]interface{}{"type": "string", "required": true},
				"Date":           map[string]interface{}{"type": "string", "required": true},
				"TimeFrom":       map[string]interface{}{"type": "string", "required": true},
				"TimeTo":         map[string]interface{}{"type": "string", "required": true},
				"CompanyName":    map[string]interface{}{"type": "string", "required": true},
				"SupportEmail":   map[string]interface{}{"type": "string", "required": true},
				"CompanyAddress": map[string]interface{}{"type": "string", "required": true},
			},
			DefaultSampleData: bookingConfirmationSampleData,
		},
	}
}

// seedTemplate seeds a single template definition (contract only for now)
func (s *SeedService) seedTemplate(def TemplateDefinition) error {
	// 1. Register contract
	schemaJSON, err := datatypes.NewJSONType(def.VariableSchema).MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	sampleDataJSON, err := datatypes.NewJSONType(def.DefaultSampleData).MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal sample data: %w", err)
	}

	channelsJSON, err := datatypes.NewJSONType(def.SupportedChannels).MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal channels: %w", err)
	}

	contract := &entities.TemplateContract{
		Module:            def.Module,
		TemplateKey:       def.TemplateKey,
		Description:       def.Description,
		SupportedChannels: channelsJSON,
		VariableSchema:    schemaJSON,
		DefaultSampleData: sampleDataJSON,
	}

	if err := s.db.Create(contract).Error; err != nil {
		return fmt.Errorf("failed to create contract: %w", err)
	}

	// TODO: Create actual template instances once the Template entity and service are updated
	// For now, templates will continue to be loaded from static files

	return nil
}

// GenerateSampleDataFiles generates JSON files with sample data next to HTML templates
func (s *SeedService) GenerateSampleDataFiles(serverDir string) error {
	definitions := s.getTemplateDefinitions()
	templatesDir := filepath.Join(serverDir, "startupseed", "email_templates")

	// Ensure directory exists
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	for _, def := range definitions {
		// Generate JSON filename based on template key
		jsonFileName := fmt.Sprintf("%s.json", def.TemplateKey)
		jsonFilePath := filepath.Join(templatesDir, jsonFileName)

		// Marshall sample data to JSON
		jsonData, err := json.MarshalIndent(def.DefaultSampleData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal sample data for %s: %w", def.TemplateKey, err)
		}

		// Write JSON file
		if err := os.WriteFile(jsonFilePath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write sample data file %s: %w", jsonFilePath, err)
		}

		fmt.Printf("✓ Generated sample data: %s\n", jsonFilePath)
	}

	return nil
}

// TemplateSeedDefinition represents a template definition from JSON
type TemplateSeedDefinition struct {
	Module      string `json:"module"`
	TemplateKey string `json:"template_key"`
	Channel     string `json:"channel"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subject     string `json:"subject,omitempty"`
	FilePath    string `json:"file_path"`
}

// TemplateSeedConfig represents the JSON structure
type TemplateSeedConfig struct {
	Templates []TemplateSeedDefinition `json:"templates"`
}

// loadTemplateSeedConfig loads template definitions from JSON file
func loadTemplateSeedConfig(serverDir string) (*TemplateSeedConfig, error) {
	configPath := filepath.Join(serverDir, "startupseed", "templates_seed.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates_seed.json: %w", err)
	}

	var config TemplateSeedConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse templates_seed.json: %w", err)
	}

	return &config, nil
}

// SeedEmailTemplatesForTenant seeds email templates for a specific tenant and organization
func (s *SeedService) SeedEmailTemplatesForTenant(tenantID uint, organizationID *uint, serverDir string) error {
	return s.seedTemplatesForTenant(tenantID, organizationID, serverDir, "EMAIL")
}

// SeedPDFTemplatesForTenant seeds PDF/document templates for a specific tenant and organization
func (s *SeedService) SeedPDFTemplatesForTenant(tenantID uint, organizationID *uint, serverDir string) error {
	return s.seedTemplatesForTenant(tenantID, organizationID, serverDir, "DOCUMENT")
}

// seedTemplatesForTenant seeds templates for a specific tenant, organization and channel
func (s *SeedService) seedTemplatesForTenant(tenantID uint, organizationID *uint, serverDir string, channelFilter string) error {
	if s.templateService == nil {
		return fmt.Errorf("template service not initialized")
	}

	ctx := context.Background()

	// Load template definitions from JSON
	config, err := loadTemplateSeedConfig(serverDir)
	if err != nil {
		return fmt.Errorf("failed to load template config: %w", err)
	}

	// Filter templates by channel
	var templates []TemplateSeedDefinition
	for _, tmpl := range config.Templates {
		if tmpl.Channel == channelFilter {
			templates = append(templates, tmpl)
		}
	}

	for _, tmpl := range templates {
		// Check if template already exists for this tenant/org
		var existingCount int64
		query := s.db.Model(&entities.Template{}).
			Where("tenant_id = ? AND template_key = ? AND channel = ?", tenantID, tmpl.TemplateKey, tmpl.Channel)

		if organizationID != nil {
			query = query.Where("organization_id = ?", *organizationID)
		} else {
			query = query.Where("organization_id IS NULL")
		}

		if err := query.Count(&existingCount).Error; err != nil {
			return fmt.Errorf("failed to check existing templates: %w", err)
		}

		if existingCount > 0 {
			fmt.Printf("⏭️  Template %s.%s already exists for tenant %d, skipping\n", tmpl.Module, tmpl.TemplateKey, tenantID)
			continue
		}

		// Read template file from startupseed folder
		filePath := filepath.Join(serverDir, "startupseed", tmpl.FilePath)
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not read template file %s: %v\n", filePath, err)
			continue
		}

		// Get variable schema from contract
		var contract entities.TemplateContract
		if err := s.db.Where("module = ? AND template_key = ?", tmpl.Module, tmpl.TemplateKey).
			First(&contract).Error; err != nil {
			fmt.Printf("⚠️  Warning: Contract not found for %s.%s, skipping\n", tmpl.Module, tmpl.TemplateKey)
			continue
		}

		// Unmarshal variable schema from contract
		var variableSchema map[string]interface{}
		if len(contract.VariableSchema) > 0 {
			if err := json.Unmarshal(contract.VariableSchema, &variableSchema); err != nil {
				fmt.Printf("⚠️  Warning: Failed to unmarshal variable schema: %v\n", err)
				variableSchema = nil
			} else {
				fmt.Printf("📋 Contract variable schema has %d variables\n", len(variableSchema))
			}
		} else {
			fmt.Printf("⚠️  Warning: Contract has empty variable schema for %s.%s\n", tmpl.Module, tmpl.TemplateKey)
		}

		// Read sample data from JSON file in startupseed root directory
		var sampleData map[string]interface{}
		jsonFileName := fmt.Sprintf("%s.json", tmpl.TemplateKey)
		jsonFilePath := filepath.Join(serverDir, "startupseed", jsonFileName)

		if jsonContent, err := os.ReadFile(jsonFilePath); err == nil {
			if err := json.Unmarshal(jsonContent, &sampleData); err != nil {
				fmt.Printf("⚠️  Warning: Failed to parse sample data JSON %s: %v\n", jsonFilePath, err)
				// Fallback to contract sample data if JSON parsing fails
				if len(contract.DefaultSampleData) > 0 {
					if err := json.Unmarshal(contract.DefaultSampleData, &sampleData); err != nil {
						fmt.Printf("⚠️  Warning: Failed to unmarshal contract sample data: %v\n", err)
						sampleData = nil
					}
				}
			} else {
				fmt.Printf("📄 Loaded sample data from: %s\n", jsonFilePath)
			}
		} else {
			fmt.Printf("⚠️  Warning: Sample data JSON not found %s, using contract default\n", jsonFilePath)
			// Fallback to contract sample data
			if len(contract.DefaultSampleData) > 0 {
				if err := json.Unmarshal(contract.DefaultSampleData, &sampleData); err != nil {
					fmt.Printf("⚠️  Warning: Failed to unmarshal contract sample data: %v\n", err)
					sampleData = nil
				}
			}
		}

		// Extract variable names from schema (top-level keys)
		var variableNames []string
		for key := range variableSchema {
			variableNames = append(variableNames, key)
		}

		// If no variables from schema, try to extract from sample data
		if len(variableNames) == 0 && sampleData != nil {
			fmt.Printf("📋 No variables in schema, extracting from sample data\n")
			for key := range sampleData {
				variableNames = append(variableNames, key)
			}
		}

		fmt.Printf("📋 Template will have %d variables: %v\n", len(variableNames), variableNames)

		// Create template instance
		createReq := &CreateTemplateRequest{
			TenantID:       tenantID,
			OrganizationID: organizationID,
			TemplateType:   tmpl.TemplateKey,
			Module:         tmpl.Module, // Set module for contract binding
			TemplateKey:    tmpl.TemplateKey,
			Channel:        tmpl.Channel,
			Name:           tmpl.Name,
			Description:    tmpl.Description,
			Content:        string(content),
			Variables:      variableNames,
			SampleData:     sampleData,
			IsActive:       true,
			IsDefault:      true,
		}

		// Add subject for email templates
		if tmpl.Channel == "EMAIL" && tmpl.Subject != "" {
			createReq.Subject = &tmpl.Subject
		}

		template, err := s.templateService.CreateTemplate(ctx, createReq)
		if err != nil {
			return fmt.Errorf("failed to create template %s.%s: %w", tmpl.Module, tmpl.TemplateKey, err)
		}

		fmt.Printf("✓ Seeded template: %s.%s [%s] (ID: %d) for tenant %d\n",
			tmpl.Module, tmpl.TemplateKey, tmpl.Channel, template.ID, tenantID)
	}

	return nil
}
