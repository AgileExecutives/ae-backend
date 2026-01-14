package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/ae-base-server/modules/templates/services/renderer"
	"github.com/ae-base-server/pkg/services"
	"gorm.io/gorm"
)

// TemplateType represents different template categories
type TemplateType string

const (
	TemplateTypeEmail    TemplateType = "email"
	TemplateTypePDF      TemplateType = "pdf"
	TemplateTypeInvoice  TemplateType = "invoice"
	TemplateTypeDocument TemplateType = "document"
)

// TemplateService handles template management operations
type TemplateService struct {
	db               *gorm.DB
	storageService   *services.StorageService
	dbRenderer       *renderer.DBRenderer
	contractProvider *DBContractProvider
	bucketService    *services.TenantBucketService
}

// NewTemplateService creates a new template service
func NewTemplateService(db *gorm.DB, storageService *services.StorageService, bucketService *services.TenantBucketService) *TemplateService {
	// Initialize database contract provider and renderer
	templatesDir := "statics/templates"
	contractProvider := NewDBContractProvider(db)
	dbRenderer, err := renderer.NewDBRenderer(templatesDir, contractProvider)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize database renderer: %v\n", err)
	}

	return &TemplateService{
		db:               db,
		storageService:   storageService,
		dbRenderer:       dbRenderer,
		contractProvider: contractProvider,
		bucketService:    bucketService,
	}
}

// getTenantBucketName returns the bucket name for a given tenant ID
func (s *TemplateService) getTenantBucketName(tenantID uint) string {
	if s.storageService != nil {
		return s.storageService.GetTenantBucketName(tenantID)
	}
	if s.bucketService != nil {
		return s.bucketService.GetTenantBucketName(tenantID)
	}
	// Fallback to legacy bucket if neither service is available
	return "templates"
}

// CreateTemplateRequest represents template creation request
type CreateTemplateRequest struct {
	TenantID       uint                   `json:"tenant_id"`
	OrganizationID *uint                  `json:"organization_id,omitempty"` // NULL = system default
	TemplateType   string                 `json:"template_type" binding:"required"`
	Module         string                 `json:"module,omitempty"`       // Module name for contract binding
	TemplateKey    string                 `json:"template_key,omitempty"` // Key derived from filename
	Channel        string                 `json:"channel,omitempty"`      // EMAIL or DOCUMENT
	Subject        *string                `json:"subject,omitempty"`      // Required for EMAIL templates
	Name           string                 `json:"name" binding:"required"`
	Description    string                 `json:"description"`
	Content        string                 `json:"content" binding:"required"` // HTML content
	Variables      []string               `json:"variables"`                  // List of variable names
	SampleData     map[string]interface{} `json:"sample_data"`                // Sample data for preview
	IsActive       bool                   `json:"is_active"`
	IsDefault      bool                   `json:"is_default"`
}

// UpdateTemplateRequest represents template update request
type UpdateTemplateRequest struct {
	OrganizationID *uint                   `json:"organization_id,omitempty"` // Set from middleware, not from client
	Name           *string                 `json:"name,omitempty"`
	Description    *string                 `json:"description,omitempty"`
	Content        *string                 `json:"content,omitempty"`
	Variables      *[]string               `json:"variables,omitempty"`
	SampleData     *map[string]interface{} `json:"sample_data,omitempty"`
	IsActive       *bool                   `json:"is_active,omitempty"`
	IsDefault      *bool                   `json:"is_default,omitempty"`
}

// CreateTemplate creates a new template
func (s *TemplateService) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*entities.Template, error) {
	// Check if setting as default, unset other defaults
	if req.IsDefault {
		if err := s.unsetDefaultTemplates(req.TenantID, req.OrganizationID, req.TemplateType); err != nil {
			return nil, fmt.Errorf("failed to unset default templates: %w", err)
		}
	}

	// Store template content using storage service
	storageKey, err := s.storageService.StoreTemplate(ctx, req.TenantID, req.TemplateType, req.Name, req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to store template content: %w", err)
	}

	// Convert variables and sample data to JSONB
	var variablesJSON []byte
	var sampleDataJSON []byte

	if len(req.Variables) > 0 {
		variablesJSON, _ = entities.MarshalJSON(req.Variables)
	}
	if len(req.SampleData) > 0 {
		sampleDataJSON, _ = entities.MarshalJSON(req.SampleData)
	}

	// If variables or sample data not provided, try to get from contract
	// Note: len(variablesJSON) == 0 means no variables were provided (not even empty slice)
	if (len(variablesJSON) == 0 || len(req.Variables) == 0) && req.Module != "" && req.TemplateKey != "" {
		if contractVariables, contractSampleData, err := s.getContractData(req.Module, req.TemplateKey); err == nil {
			if len(variablesJSON) == 0 || len(req.Variables) == 0 {
				variablesJSON = contractVariables
				fmt.Printf("📋 Using contract variable schema (%d bytes)\n", len(contractVariables))
			}
			if len(sampleDataJSON) == 0 || len(req.SampleData) == 0 {
				sampleDataJSON = contractSampleData
				fmt.Printf("📄 Using contract sample data (%d bytes)\n", len(contractSampleData))
			}
		} else {
			fmt.Printf("⚠️  Could not get contract data for %s.%s: %v\n", req.Module, req.TemplateKey, err)
		}
	}

	// Use empty defaults if still empty
	if len(variablesJSON) == 0 {
		variablesJSON = []byte("[]")
	}
	if len(sampleDataJSON) == 0 {
		sampleDataJSON = []byte("{}")
	}

	// Create template record
	tmpl := &entities.Template{
		TenantID:       req.TenantID,
		OrganizationID: req.OrganizationID,
		TemplateType:   req.TemplateType,
		Module:         req.Module,                    // Set module for contract binding
		TemplateKey:    req.TemplateKey,               // Set template key from filename
		Channel:        entities.Channel(req.Channel), // Set channel from request
		Subject:        req.Subject,                   // Set subject for EMAIL templates
		Name:           req.Name,
		Description:    req.Description,
		Version:        1,
		IsActive:       req.IsActive,
		IsDefault:      req.IsDefault,
		StorageKey:     storageKey,
		Variables:      variablesJSON,
		SampleData:     sampleDataJSON,
	}

	fmt.Printf("🔍 DEBUG: Creating template with TemplateKey='%s' (req.TemplateKey='%s')\n", tmpl.TemplateKey, req.TemplateKey)

	if err := s.db.Create(tmpl).Error; err != nil {
		// Rollback storage if DB insert fails
		s.storageService.DeleteTemplate(ctx, req.TenantID, storageKey)
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Printf("🔍 DEBUG: After Create - TemplateKey='%s', ID=%d\n", tmpl.TemplateKey, tmpl.ID)

	return tmpl, nil
}

// GetTemplate retrieves a template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, tenantID uint, templateID uint) (*entities.Template, error) {
	var tmpl entities.Template
	if err := s.db.Where("id = ? AND tenant_id = ?", templateID, tenantID).First(&tmpl).Error; err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// GetTemplateWithContent retrieves template with content from storage
func (s *TemplateService) GetTemplateWithContent(ctx context.Context, tenantID uint, templateID uint) (*entities.Template, string, error) {
	tmpl, err := s.GetTemplate(ctx, tenantID, templateID)
	if err != nil {
		return nil, "", err
	}

	// Retrieve content from storage using storage service
	content, err := s.storageService.RetrieveTemplate(ctx, tmpl.TenantID, tmpl.StorageKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve template content: %w", err)
	}

	return tmpl, string(content), nil
}

// ListTemplates lists templates with filters and pagination
func (s *TemplateService) ListTemplates(
	ctx context.Context,
	tenantID uint,
	organizationID *uint,
	channel string,
	templateKey string,
	isActive *bool,
	page int,
	pageSize int,
) ([]entities.Template, int64, error) {
	var templates []entities.Template
	var total int64

	query := s.db.Model(&entities.Template{}).Where("tenant_id = ?", tenantID)

	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	}
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	if templateKey != "" {
		query = query.Where("template_key = ?", templateKey)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// UpdateTemplate updates an existing template
func (s *TemplateService) UpdateTemplate(
	ctx context.Context,
	tenantID uint,
	templateID uint,
	req *UpdateTemplateRequest,
) (*entities.Template, error) {
	tmpl, err := s.GetTemplate(ctx, tenantID, templateID)
	if err != nil {
		return nil, err
	}

	// If content is being updated, create new version
	if req.Content != nil {
		newVersion := tmpl.Version + 1
		storageKey := fmt.Sprintf("templates/%s/%s_v%d_%d.html",
			tmpl.TemplateType,
			tmpl.Name,
			newVersion,
			time.Now().Unix(),
		)

		err := s.storageService.StoreTemplateWithKey(ctx, tmpl.TenantID, storageKey, *req.Content, map[string]string{
			"template_type": tmpl.TemplateType,
			"template_name": tmpl.Name,
			"version":       fmt.Sprintf("%d", newVersion),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to store new template version: %w", err)
		}

		tmpl.StorageKey = storageKey
		tmpl.Version = newVersion
	}

	// Update fields
	if req.Name != nil {
		tmpl.Name = *req.Name
	}
	if req.Description != nil {
		tmpl.Description = *req.Description
	}
	if req.Variables != nil {
		variablesJSON, _ := entities.MarshalJSON(*req.Variables)
		tmpl.Variables = variablesJSON
	} else if len(tmpl.Variables) == 0 {
		// If no variables stored and none provided, try to get from contract
		if contractVariables, _, err := s.getContractData(tmpl.Module, tmpl.TemplateKey); err == nil {
			tmpl.Variables = contractVariables
		}
	}
	if req.SampleData != nil {
		sampleDataJSON, _ := entities.MarshalJSON(*req.SampleData)
		tmpl.SampleData = sampleDataJSON
	} else if len(tmpl.SampleData) == 0 {
		// If no sample data stored and none provided, try to get from contract
		if _, contractSampleData, err := s.getContractData(tmpl.Module, tmpl.TemplateKey); err == nil {
			tmpl.SampleData = contractSampleData
		}
	}
	if req.IsActive != nil {
		tmpl.IsActive = *req.IsActive
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			// Unset other defaults
			if err := s.unsetDefaultTemplates(tenantID, tmpl.OrganizationID, tmpl.TemplateType); err != nil {
				return nil, err
			}
		}
		tmpl.IsDefault = *req.IsDefault
	}

	if err := s.db.Save(tmpl).Error; err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return tmpl, nil
}

// DeleteTemplate soft deletes a template
func (s *TemplateService) DeleteTemplate(ctx context.Context, tenantID uint, templateID uint) error {
	return s.db.Where("id = ? AND tenant_id = ?", templateID, tenantID).
		Delete(&entities.Template{}).Error
}

// GetDefaultTemplate gets the default template for a type
func (s *TemplateService) GetDefaultTemplate(
	ctx context.Context,
	tenantID uint,
	organizationID *uint,
	templateType string,
) (*entities.Template, error) {
	var tmpl entities.Template

	// First try organization-specific default
	if organizationID != nil {
		err := s.db.Where(
			"tenant_id = ? AND organization_id = ? AND template_type = ? AND is_default = ? AND is_active = ?",
			tenantID, *organizationID, templateType, true, true,
		).First(&tmpl).Error

		if err == nil {
			return &tmpl, nil
		}
	}

	// Fall back to system default (organization_id = NULL)
	err := s.db.Where(
		"tenant_id = ? AND organization_id IS NULL AND template_type = ? AND is_default = ? AND is_active = ?",
		tenantID, templateType, true, true,
	).First(&tmpl).Error

	if err != nil {
		return nil, err
	}

	return &tmpl, nil
}

// RenderTemplate renders a template with provided data
func (s *TemplateService) RenderTemplate(
	ctx context.Context,
	tenantID uint,
	templateID uint,
	data map[string]interface{},
) (string, error) {
	tmpl, content, err := s.GetTemplateWithContent(ctx, tenantID, templateID)
	if err != nil {
		return "", err
	}

	// Use new renderer for templates with contract binding
	if s.dbRenderer != nil && tmpl.Module != "" && tmpl.TemplateKey != "" {
		// Use contract from template's Module.TemplateKey binding
		contractName := tmpl.TemplateKey

		fmt.Printf("🎨 Rendering template with contract: %s.%s\n", tmpl.Module, tmpl.TemplateKey)

		// Convert data to JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal template data: %w", err)
		}

		// Use the database renderer with contract validation
		result, err := s.dbRenderer.RenderTemplateFromContent(content, contractName, jsonData)
		if err != nil {
			// Fall back to legacy rendering if contract validation fails
			fmt.Printf("⚠️  Warning: Contract validation failed for %s, falling back to legacy rendering: %v\n", contractName, err)
		} else {
			fmt.Printf("✅ Successfully rendered template with contract validation\n")
			return string(result), nil
		}
	}

	// Legacy rendering for unsupported templates or fallback
	templateData := s.prepareTemplateData(data)

	// Parse template
	htmlTemplate, err := template.New(tmpl.Name).Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := htmlTemplate.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// PreviewTemplate previews a template with sample data
func (s *TemplateService) PreviewTemplate(ctx context.Context, tenantID uint, templateID uint) (string, error) {
	tmpl, content, err := s.GetTemplateWithContent(ctx, tenantID, templateID)
	if err != nil {
		return "", err
	}

	// Get sample data
	var sampleData map[string]interface{}
	if len(tmpl.SampleData) > 0 {
		if err := entities.UnmarshalJSON(tmpl.SampleData, &sampleData); err != nil {
			return "", fmt.Errorf("failed to parse sample data: %w", err)
		}
	} else {
		sampleData = map[string]interface{}{}
	}

	// Use database renderer for templates with contract binding
	if s.dbRenderer != nil && tmpl.Module != "" && tmpl.TemplateKey != "" {
		contractName := tmpl.TemplateKey

		fmt.Printf("🎨 Previewing template with contract: %s.%s\n", tmpl.Module, tmpl.TemplateKey)

		// Convert sample data to JSON
		jsonData, err := json.Marshal(sampleData)
		if err != nil {
			return "", fmt.Errorf("failed to marshal sample data: %w", err)
		}

		// Use the database renderer with contract validation
		result, err := s.dbRenderer.RenderTemplateFromContent(content, contractName, jsonData)
		if err != nil {
			// Fall back to legacy rendering if contract validation fails
			fmt.Printf("⚠️  Warning: Contract validation failed for %s preview, falling back to legacy rendering: %v\n", contractName, err)
		} else {
			fmt.Printf("✅ Successfully previewed template with contract validation\n")
			return string(result), nil
		}
	}

	// Legacy rendering for unsupported templates or fallback
	htmlTemplate, err := template.New(tmpl.Name).Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := htmlTemplate.Execute(&buf, sampleData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// unsetDefaultTemplates removes default flag from other templates
func (s *TemplateService) unsetDefaultTemplates(tenantID uint, organizationID *uint, templateType string) error {
	query := s.db.Model(&entities.Template{}).
		Where("tenant_id = ? AND template_type = ?", tenantID, templateType)

	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	return query.Update("is_default", false).Error
}

// DuplicateTemplate creates a copy of an existing template
func (s *TemplateService) DuplicateTemplate(
	ctx context.Context,
	tenantID uint,
	templateID uint,
	newName string,
) (*entities.Template, error) {
	// Get source template with content
	source, content, err := s.GetTemplateWithContent(ctx, tenantID, templateID)
	if err != nil {
		return nil, err
	}

	// Get variables and sample data
	var variables []string
	var sampleData map[string]interface{}

	if len(source.Variables) > 0 {
		entities.UnmarshalJSON(source.Variables, &variables)
	}
	if len(source.SampleData) > 0 {
		entities.UnmarshalJSON(source.SampleData, &sampleData)
	}

	// Create new template
	req := &CreateTemplateRequest{
		TenantID:       tenantID,
		OrganizationID: source.OrganizationID,
		TemplateType:   source.TemplateType,
		TemplateKey:    source.TemplateKey, // Preserve template key when copying
		Name:           newName,
		Description:    fmt.Sprintf("Copy of %s", source.Name),
		Content:        content,
		Variables:      variables,
		SampleData:     sampleData,
		IsActive:       false, // New copy starts as inactive
		IsDefault:      false,
	}

	return s.CreateTemplate(ctx, req)
}

// CopyTemplatesFromTenant2Org2 copies all templates from tenant 2, org 2 to a new organization
// This is used when creating new tenants/organizations to provide default templates
func (s *TemplateService) CopyTemplatesFromTenant2Org2(ctx context.Context, targetTenantID, targetOrganizationID uint) error {
	// Get all templates from tenant 2, organization 2
	var sourceTemplates []entities.Template
	if err := s.db.Where("tenant_id = ? AND organization_id = ?", 2, 2).Find(&sourceTemplates).Error; err != nil {
		return fmt.Errorf("failed to get source templates from tenant 2 org 2: %w", err)
	}

	if len(sourceTemplates) == 0 {
		return fmt.Errorf("no templates found in tenant 2 org 2 to copy")
	}

	fmt.Printf("📋 Copying %d templates from tenant 2 org 2 to tenant %d org %d...\n", len(sourceTemplates), targetTenantID, targetOrganizationID)

	// Copy each template
	copiedCount := 0
	for _, source := range sourceTemplates {
		// Get the content from storage
		content, err := s.storageService.RetrieveTemplate(ctx, source.TenantID, source.StorageKey)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not read template content for %s: %v\n", source.Name, err)
			continue
		}

		// Get variables and sample data
		var variables []string
		var sampleData map[string]interface{}

		if len(source.Variables) > 0 {
			entities.UnmarshalJSON(source.Variables, &variables)
		}
		if len(source.SampleData) > 0 {
			entities.UnmarshalJSON(source.SampleData, &sampleData)
		}

		// Create new template for target tenant/org
		req := &CreateTemplateRequest{
			TenantID:       targetTenantID,
			OrganizationID: &targetOrganizationID,
			TemplateType:   source.TemplateType,
			TemplateKey:    source.TemplateKey, // Preserve template key when copying
			Channel:        string(source.Channel),
			Name:           source.Name,
			Description:    source.Description,
			Content:        string(content),
			Variables:      variables,
			SampleData:     sampleData,
			IsActive:       source.IsActive,
			IsDefault:      source.IsDefault,
		}

		if _, err := s.CreateTemplate(ctx, req); err != nil {
			fmt.Printf("⚠️  Warning: Failed to copy template %s: %v\\n", source.Name, err)
			continue
		}

		copiedCount++
	}

	fmt.Printf("✅ Successfully copied %d/%d templates from tenant 2 org 2\\n", copiedCount, len(sourceTemplates))
	return nil
}

// GetTemplateByKey retrieves a template by its key for a specific tenant and channel
func (s *TemplateService) GetTemplateByKey(ctx context.Context, tenantID uint, channel, templateKey string) (string, error) {
	var tmpl entities.Template

	query := s.db.Where("tenant_id = ? AND template_key = ? AND is_active = ?", tenantID, templateKey, true)

	if channel != "" {
		query = query.Where("channel = ?", channel)
	}

	// Order by is_default DESC to prefer default templates, then by created_at DESC for newest
	if err := query.Order("is_default DESC, created_at DESC").First(&tmpl).Error; err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}

	// Retrieve content from storage
	content, err := s.storageService.RetrieveTemplate(ctx, tmpl.TenantID, tmpl.StorageKey)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve template content: %w", err)
	}

	return string(content), nil
}

// prepareTemplateData maps input data to the expected template structure
func (s *TemplateService) prepareTemplateData(data map[string]interface{}) map[string]interface{} {
	templateData := make(map[string]interface{})

	// Add standard template fields
	templateData["AppName"] = getStringValue(data, "app_name", "Unburdy")

	// Map common booking template variables to CustomData structure
	customData := make(map[string]interface{})

	// Map user/client information
	if userName := getStringValue(data, "user_name", ""); userName != "" {
		customData["ClientName"] = userName
	}
	if userFirstName := getStringValue(data, "user_first_name", ""); userFirstName != "" {
		customData["ClientName"] = userFirstName
	}

	// Map booking/appointment information
	if bookingDate := getStringValue(data, "booking_date", ""); bookingDate != "" {
		customData["AppointmentDate"] = bookingDate
	}
	if bookingId := getStringValue(data, "booking_id", ""); bookingId != "" {
		customData["BookingId"] = bookingId
	}

	// Map time information
	customData["TimeFrom"] = getStringValue(data, "time_from", getStringValue(data, "start_time", ""))
	customData["TimeTo"] = getStringValue(data, "time_to", getStringValue(data, "end_time", ""))

	// Map additional common fields
	customData["Title"] = getStringValue(data, "title", getStringValue(data, "service_name", ""))
	customData["Description"] = getStringValue(data, "description", "")
	customData["Duration"] = getStringValue(data, "duration", "")
	customData["Location"] = getStringValue(data, "location", "")

	// Handle series/multiple appointments
	customData["IsSeries"] = getBoolValue(data, "is_series", false)
	customData["AppointmentCount"] = getIntValue(data, "appointment_count", 1)

	// Add appointments array for series
	if appointments, exists := data["appointments"]; exists {
		customData["Appointments"] = appointments
	}

	templateData["CustomData"] = customData

	// Also include all original data as fallback
	for key, value := range data {
		if _, exists := templateData[key]; !exists {
			templateData[key] = value
		}
	}

	return templateData
}

// Helper functions for safe type conversion
func getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getBoolValue(data map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getIntValue(data map[string]interface{}, key string, defaultValue int) int {
	if val, ok := data[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return defaultValue
}

// getContractData retrieves variables and sample data from the template contract
func (s *TemplateService) getContractData(module, templateKey string) ([]byte, []byte, error) {
	var contract entities.TemplateContract
	err := s.db.Where("module = ? AND template_key = ?", module, templateKey).First(&contract).Error
	if err != nil {
		return nil, nil, fmt.Errorf("contract not found: %w", err)
	}

	// Extract variable names from schema
	var variablesJSON []byte
	if len(contract.VariableSchema) > 0 {
		var schema map[string]interface{}
		if err := json.Unmarshal(contract.VariableSchema, &schema); err == nil {
			// Extract top-level keys as variable names
			variableNames := make([]string, 0, len(schema))
			for key := range schema {
				variableNames = append(variableNames, key)
			}
			variablesJSON, _ = json.Marshal(variableNames)
		} else {
			variablesJSON = []byte("[]")
		}
	} else {
		variablesJSON = []byte("[]")
	}

	// Use default sample data from contract
	sampleDataJSON := []byte(contract.DefaultSampleData)
	if len(sampleDataJSON) == 0 {
		sampleDataJSON = []byte("{}")
	}

	return variablesJSON, sampleDataJSON, nil
}
