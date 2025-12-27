package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/unburdy/unburdy-server-api/modules/documents/entities"
	"github.com/unburdy/unburdy-server-api/modules/documents/services/storage"
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
	db      *gorm.DB
	storage storage.DocumentStorage
}

// NewTemplateService creates a new template service
func NewTemplateService(db *gorm.DB, storage storage.DocumentStorage) *TemplateService {
	return &TemplateService{
		db:      db,
		storage: storage,
	}
}

// CreateTemplateRequest represents template creation request
type CreateTemplateRequest struct {
	TenantID       uint                   `json:"tenant_id"`
	OrganizationID *uint                  `json:"organization_id,omitempty"` // NULL = system default
	TemplateType   string                 `json:"template_type" binding:"required"`
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
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Content     *string                `json:"content,omitempty"`
	Variables   *[]string              `json:"variables,omitempty"`
	SampleData  *map[string]interface{} `json:"sample_data,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
	IsDefault   *bool                  `json:"is_default,omitempty"`
}

// CreateTemplate creates a new template
func (s *TemplateService) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*entities.Template, error) {
	// Check if setting as default, unset other defaults
	if req.IsDefault {
		if err := s.unsetDefaultTemplates(req.TenantID, req.OrganizationID, req.TemplateType); err != nil {
			return nil, fmt.Errorf("failed to unset default templates: %w", err)
		}
	}

	// Store template content in MinIO
	storageKey := fmt.Sprintf("tenants/%d/templates/%s/%s_%d.html",
		req.TenantID,
		req.TemplateType,
		req.Name,
		time.Now().Unix(),
	)

	_, err := s.storage.Store(ctx, storage.StoreRequest{
		Bucket:      "templates",
		Key:         storageKey,
		Data:        []byte(req.Content),
		ContentType: "text/html",
		Metadata: map[string]string{
			"template_type": req.TemplateType,
			"template_name": req.Name,
		},
	})
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

	// Create template record
	tmpl := &entities.Template{
		TenantID:       req.TenantID,
		OrganizationID: req.OrganizationID,
		TemplateType:   req.TemplateType,
		Name:           req.Name,
		Description:    req.Description,
		Version:        1,
		IsActive:       req.IsActive,
		IsDefault:      req.IsDefault,
		StorageKey:     storageKey,
		Variables:      variablesJSON,
		SampleData:     sampleDataJSON,
	}

	if err := s.db.Create(tmpl).Error; err != nil {
		// Rollback storage if DB insert fails
		s.storage.Delete(ctx, "templates", storageKey)
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

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

	// Retrieve content from storage
	content, err := s.storage.Retrieve(ctx, "templates", tmpl.StorageKey)
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
	templateType string,
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
	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
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
		storageKey := fmt.Sprintf("tenants/%d/templates/%s/%s_v%d_%d.html",
			tenantID,
			tmpl.TemplateType,
			tmpl.Name,
			newVersion,
			time.Now().Unix(),
		)

		_, err := s.storage.Store(ctx, storage.StoreRequest{
			Bucket:      "templates",
			Key:         storageKey,
			Data:        []byte(*req.Content),
			ContentType: "text/html",
			Metadata: map[string]string{
				"template_type": tmpl.TemplateType,
				"template_name": tmpl.Name,
				"version":       fmt.Sprintf("%d", newVersion),
			},
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
	}
	if req.SampleData != nil {
		sampleDataJSON, _ := entities.MarshalJSON(*req.SampleData)
		tmpl.SampleData = sampleDataJSON
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

	// Parse template
	htmlTemplate, err := template.New(tmpl.Name).Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := htmlTemplate.Execute(&buf, data); err != nil {
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

	// Parse and execute template
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
