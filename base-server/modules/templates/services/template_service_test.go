package services

import (
	"context"
	"testing"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/ae-base-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTemplateService(t *testing.T) (*TemplateService, *gorm.DB) {
	db := testutils.SetupTestDB(t)
	require.NoError(t, db.AutoMigrate(&entities.Template{}, &entities.TemplateContract{}))
	// Pass nil for storage services - only test DB operations
	svc := NewTemplateService(db, nil, nil)
	return svc, db
}

func seedTemplateSvc(t *testing.T, db *gorm.DB, tenantID uint, templateType string, isDefault, isActive bool, storageKeySuffix string) *entities.Template {
	t.Helper()
	tmpl := entities.Template{
		TenantID:     tenantID,
		Name:         "Test " + templateType,
		TemplateType: templateType,
		StorageKey:   "test/" + storageKeySuffix,
		Channel:      "DOCUMENT",
		IsActive:     isActive,
		IsDefault:    isDefault,
		Version:      1,
	}
	require.NoError(t, db.Create(&tmpl).Error)
	return &tmpl
}

func TestTemplateService_CreateTemplate(t *testing.T) {
	t.Skip("Template service CreateTemplate requires MinIO storage - skipping")
}

func TestTemplateService_GetTemplate(t *testing.T) {
	svc, db := setupTemplateService(t)
	ctx := context.Background()

	tmpl := seedTemplateSvc(t, db, 1, "invoice", false, true, "get-1")

	t.Run("found", func(t *testing.T) {
		result, err := svc.GetTemplate(ctx, 1, tmpl.ID)
		require.NoError(t, err)
		assert.Equal(t, tmpl.ID, result.ID)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetTemplate(ctx, 1, 9999)
		require.Error(t, err)
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		_, err := svc.GetTemplate(ctx, 2, tmpl.ID)
		require.Error(t, err)
	})
}

func TestTemplateService_ListTemplates(t *testing.T) {
	svc, db := setupTemplateService(t)
	ctx := context.Background()

	for i, ttype := range []string{"invoice", "invoice", "contract"} {
		tmpl := entities.Template{
			TenantID:     1,
			Name:         "T",
			TemplateType: ttype,
			StorageKey:   "list-key-" + string(rune('0'+i)),
			Channel:      "DOCUMENT",
			IsActive:     true,
			Version:      1,
		}
		require.NoError(t, db.Create(&tmpl).Error)
	}
	other := entities.Template{TenantID: 2, Name: "O", TemplateType: "invoice", StorageKey: "other", Channel: "DOCUMENT", IsActive: true, Version: 1}
	require.NoError(t, db.Create(&other).Error)

	t.Run("returns all for tenant", func(t *testing.T) {
		_, total, err := svc.ListTemplates(ctx, 1, nil, "", "", nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
	})

	t.Run("filter by channel", func(t *testing.T) {
		_, total, err := svc.ListTemplates(ctx, 1, nil, "DOCUMENT", "", nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
	})

	t.Run("filter by isActive false", func(t *testing.T) {
		isActive := false
		_, total, err := svc.ListTemplates(ctx, 1, nil, "", "", &isActive, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		_, total, err := svc.ListTemplates(ctx, 2, nil, "", "", nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := svc.ListTemplates(ctx, 1, nil, "", "", nil, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 2)
	})
}

func TestTemplateService_DeleteTemplate(t *testing.T) {
	svc, db := setupTemplateService(t)
	ctx := context.Background()

	tmpl := seedTemplateSvc(t, db, 1, "invoice", false, true, "delete-1")

	err := svc.DeleteTemplate(ctx, 1, tmpl.ID)
	require.NoError(t, err)
	_, err2 := svc.GetTemplate(ctx, 1, tmpl.ID)
	require.Error(t, err2)
}

func TestTemplateService_GetDefaultTemplate(t *testing.T) {
	svc, db := setupTemplateService(t)
	ctx := context.Background()

	tmpl := seedTemplateSvc(t, db, 1, "invoice", true, true, "default-1")

	t.Run("returns system default", func(t *testing.T) {
		result, err := svc.GetDefaultTemplate(ctx, 1, nil, "invoice")
		require.NoError(t, err)
		assert.Equal(t, tmpl.ID, result.ID)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetDefaultTemplate(ctx, 1, nil, "contract")
		require.Error(t, err)
	})

	t.Run("org-specific default takes priority", func(t *testing.T) {
		orgID := uint(5)
		orgTmpl := entities.Template{
			TenantID: 1, OrganizationID: &orgID, Name: "OrgDefault",
			TemplateType: "invoice", StorageKey: "org-default-invoice", Channel: "DOCUMENT",
			IsActive: true, IsDefault: true, Version: 1,
		}
		require.NoError(t, db.Create(&orgTmpl).Error)
		result, err := svc.GetDefaultTemplate(ctx, 1, &orgID, "invoice")
		require.NoError(t, err)
		assert.Equal(t, orgTmpl.ID, result.ID)
	})
}

func TestTemplateService_UpdateTemplate_MetadataOnly(t *testing.T) {
	svc, db := setupTemplateService(t)
	ctx := context.Background()

	tmpl := seedTemplateSvc(t, db, 1, "invoice", false, true, "update-1")

	t.Run("update name and active flag", func(t *testing.T) {
		newName := "Updated Name"
		isActive := false
		updated, err := svc.UpdateTemplate(ctx, 1, tmpl.ID, &UpdateTemplateRequest{
			Name:     &newName,
			IsActive: &isActive,
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.False(t, updated.IsActive)
	})

	t.Run("set as default clears other defaults", func(t *testing.T) {
		tmpl2 := seedTemplateSvc(t, db, 1, "invoice", true, true, "update-2")
		isDefault := true
		_, err := svc.UpdateTemplate(ctx, 1, tmpl.ID, &UpdateTemplateRequest{IsDefault: &isDefault})
		require.NoError(t, err)
		var check entities.Template
		db.First(&check, tmpl2.ID)
		assert.False(t, check.IsDefault)
	})

	t.Run("not found returns error", func(t *testing.T) {
		newName := "X"
		_, err := svc.UpdateTemplate(ctx, 1, 9999, &UpdateTemplateRequest{Name: &newName})
		require.Error(t, err)
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
