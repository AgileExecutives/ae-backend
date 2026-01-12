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
	return []TemplateDefinition{
		// Billing: Invoice
		{
			Module:            "billing",
			TemplateKey:       "invoice",
			Description:       "Invoice template for client billing",
			SupportedChannels: []string{"DOCUMENT"},
			VariableSchema: map[string]interface{}{
				"Invoice": map[string]interface{}{
					"type":     "object",
					"required": true,
					"properties": map[string]interface{}{
						"InvoiceNumber": map[string]interface{}{"type": "string", "required": true},
						"IsDraft":       map[string]interface{}{"type": "boolean"},
						"InvoiceDate":   map[string]interface{}{"type": "string"},
						"DueDate":       map[string]interface{}{"type": "string"},
					},
				},
				"Organization": map[string]interface{}{
					"type":     "object",
					"required": true,
					"properties": map[string]interface{}{
						"Name":    map[string]interface{}{"type": "string", "required": true},
						"Address": map[string]interface{}{"type": "string"},
						"Email":   map[string]interface{}{"type": "string"},
					},
				},
				"Client": map[string]interface{}{
					"type":     "object",
					"required": true,
					"properties": map[string]interface{}{
						"Name":    map[string]interface{}{"type": "string", "required": true},
						"Address": map[string]interface{}{"type": "string"},
					},
				},
				"NetTotal":   map[string]interface{}{"type": "number", "required": true},
				"TaxRate":    map[string]interface{}{"type": "number", "required": true},
				"GrossTotal": map[string]interface{}{"type": "number", "required": true},
			},
			DefaultSampleData: map[string]interface{}{
				"Invoice": map[string]interface{}{
					"InvoiceNumber": "INV-2024-001",
					"IsDraft":       false,
					"InvoiceDate":   "2024-01-15",
					"DueDate":       "2024-02-15",
				},
				"Organization": map[string]interface{}{
					"Name":    "Sample Organization GmbH",
					"Address": "Musterstraße 1, 12345 Berlin",
					"Email":   "info@example.com",
				},
				"Client": map[string]interface{}{
					"Name":    "Sample Client GmbH",
					"Address": "Kundenstraße 2, 54321 Hamburg",
				},
				"NetTotal":   1500.00,
				"TaxRate":    19.0,
				"GrossTotal": 1785.00,
			},
		},
		// Identity: Password Reset
		{
			Module:            "identity",
			TemplateKey:       "password_reset",
			Description:       "Password reset email template",
			SupportedChannels: []string{"EMAIL"},
			VariableSchema: map[string]interface{}{
				"user_email": map[string]interface{}{"type": "string", "required": true},
				"reset_link": map[string]interface{}{"type": "string", "required": true},
				"expires_in": map[string]interface{}{"type": "string"},
			},
			DefaultSampleData: map[string]interface{}{
				"user_email": "user@example.com",
				"reset_link": "https://example.com/reset-password?token=abc123",
				"expires_in": "24 hours",
			},
		},
		// Identity: Email Verification
		{
			Module:            "identity",
			TemplateKey:       "email_verification",
			Description:       "Email verification template",
			SupportedChannels: []string{"EMAIL"},
			VariableSchema: map[string]interface{}{
				"user_name":         map[string]interface{}{"type": "string", "required": true},
				"verification_link": map[string]interface{}{"type": "string", "required": true},
			},
			DefaultSampleData: map[string]interface{}{
				"user_name":         "John Doe",
				"verification_link": "https://example.com/verify?token=xyz789",
			},
		},
		// Identity: Welcome Email
		{
			Module:            "identity",
			TemplateKey:       "welcome",
			Description:       "Welcome email for new users",
			SupportedChannels: []string{"EMAIL"},
			VariableSchema: map[string]interface{}{
				"user_name": map[string]interface{}{"type": "string", "required": true},
			},
			DefaultSampleData: map[string]interface{}{
				"user_name": "John Doe",
			},
		},
		// Notification: Booking Confirmation
		{
			Module:            "notification",
			TemplateKey:       "booking_confirmation",
			Description:       "Booking confirmation email",
			SupportedChannels: []string{"EMAIL"},
			VariableSchema: map[string]interface{}{
				"booking_id":   map[string]interface{}{"type": "string", "required": true},
				"user_name":    map[string]interface{}{"type": "string", "required": true},
				"booking_date": map[string]interface{}{"type": "string"},
			},
			DefaultSampleData: map[string]interface{}{
				"booking_id":   "BK-12345",
				"user_name":    "John Doe",
				"booking_date": "2024-02-15",
			},
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

		// Unmarshal variable schema and sample data from JSON
		var variableSchema map[string]interface{}
		if len(contract.VariableSchema) > 0 {
			if err := json.Unmarshal(contract.VariableSchema, &variableSchema); err != nil {
				fmt.Printf("⚠️  Warning: Failed to unmarshal variable schema: %v\n", err)
				variableSchema = nil
			}
		}

		var sampleData map[string]interface{}
		if len(contract.DefaultSampleData) > 0 {
			if err := json.Unmarshal(contract.DefaultSampleData, &sampleData); err != nil {
				fmt.Printf("⚠️  Warning: Failed to unmarshal sample data: %v\n", err)
				sampleData = nil
			}
		}

		// Extract variable names from schema (top-level keys)
		var variableNames []string
		for key := range variableSchema {
			variableNames = append(variableNames, key)
		}

		// Create template instance
		createReq := &CreateTemplateRequest{
			TenantID:       tenantID,
			OrganizationID: organizationID,
			TemplateType:   tmpl.TemplateKey,
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
