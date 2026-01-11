package services

import (
	"fmt"

	"github.com/ae-base-server/modules/templates/entities"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SeedService handles template seeding from static files
type SeedService struct {
	db *gorm.DB
}

// NewSeedService creates a new seed service
func NewSeedService(db *gorm.DB) *SeedService {
	return &SeedService{
		db: db,
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
