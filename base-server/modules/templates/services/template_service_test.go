package services

import (
	"testing"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/ae-base-server/pkg/testutils"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTemplateService(t *testing.T) (*TemplateService, *gorm.DB) {
	db := testutils.SetupTestDB(t)

	err := db.AutoMigrate(&entities.Template{})
	require.NoError(t, err)

	// Template service tests will be skipped as they require MinIO storage
	// For now, return nil service to avoid compilation errors
	return nil, db
}

func TestTemplateService_CreateTemplate(t *testing.T) {
	t.Skip("Template service tests require MinIO storage - skipping")
}

func TestTemplateService_GetTemplate(t *testing.T) {
	t.Skip("Template service tests require MinIO storage - skipping")
}

func TestTemplateService_ListTemplates(t *testing.T) {
	t.Skip("Template service tests require MinIO storage - skipping")
}

func TestTemplateService_DeleteTemplate(t *testing.T) {
	t.Skip("Template service tests require MinIO storage - skipping")
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
