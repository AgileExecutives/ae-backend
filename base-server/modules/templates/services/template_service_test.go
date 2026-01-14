package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/ae-base-server/modules/templates/services/storage"
	"github.com/ae-base-server/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockStorage is a mock implementation of DocumentStorage
type MockStorage struct {
	storedFiles map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		storedFiles: make(map[string][]byte),
	}
}

func (m *MockStorage) Store(ctx context.Context, req storage.StoreRequest) (string, error) {
	storageKey := req.Key
	m.storedFiles[storageKey] = req.Data
	return storageKey, nil
}

func (m *MockStorage) Retrieve(ctx context.Context, bucket, key string) ([]byte, error) {
	if content, exists := m.storedFiles[key]; exists {
		return content, nil
	}
	return nil, errors.New("file not found")
}

func (m *MockStorage) GetURL(ctx context.Context, bucket, key string, expiresIn time.Duration) (string, error) {
	return "http://mock-storage/" + key, nil
}

func (m *MockStorage) Delete(ctx context.Context, bucket, key string) error {
	delete(m.storedFiles, key)
	return nil
}

func (m *MockStorage) List(ctx context.Context, bucket, prefix string) ([]storage.DocumentMeta, error) {
	var docs []storage.DocumentMeta
	for key := range m.storedFiles {
		if strings.HasPrefix(key, prefix) {
			docs = append(docs, storage.DocumentMeta{
				Key:  key,
				Size: int64(len(m.storedFiles[key])),
			})
		}
	}
	return docs, nil
}

func (m *MockStorage) Exists(ctx context.Context, bucket, key string) (bool, error) {
	_, exists := m.storedFiles[key]
	return exists, nil
}

func setupTemplateService(t *testing.T) (*TemplateService, *gorm.DB, *MockStorage) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.Template{})
	require.NoError(t, err)

	mockStorage := NewMockStorage()
	bucketService := &services.TenantBucketService{} // Mock bucket service for tests
	service := NewTemplateService(db, mockStorage, bucketService)

	return service, db, mockStorage
}

func TestTemplateService_CreateTemplate(t *testing.T) {
	service, db, mockStorage := setupTemplateService(t)

	req := &CreateTemplateRequest{
		TenantID:     1,
		TemplateType: "email",
		TemplateKey:  "test_template",
		Channel:      "EMAIL",
		Subject:      stringPtr("Test Subject"),
		Name:         "Test Template",
		Description:  "Test Description",
		Content:      "<h1>Test Content</h1>",
		Variables:    []string{"Name", "Email"},
		SampleData: map[string]interface{}{
			"Name":  "John Doe",
			"Email": "john@example.com",
		},
		IsActive:  true,
		IsDefault: false,
	}

	template, err := service.CreateTemplate(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "Test Template", template.Name)
	assert.Equal(t, "email", template.TemplateType)
	assert.Equal(t, entities.ChannelEmail, template.Channel)

	// Verify template was stored in database
	var dbTemplate entities.Template
	err = db.Where("id = ?", template.ID).First(&dbTemplate).Error
	assert.NoError(t, err)
	assert.Equal(t, template.Name, dbTemplate.Name)

	// Verify content was stored in mock storage
	assert.Contains(t, mockStorage.storedFiles, dbTemplate.StorageKey)
}

func TestTemplateService_GetTemplate(t *testing.T) {
	service, db, mockStorage := setupTemplateService(t)

	// Create a template in the database
	content := "<h1>Test Template</h1><p>Hello {{.Name}}!</p>"
	storageKey := "test-storage-key"

	template := &entities.Template{
		TenantID:     1,
		TemplateType: "email",
		TemplateKey:  "test_key",
		Channel:      "EMAIL",
		Subject:      stringPtr("Test Subject"),
		Name:         "Test Template",
		Description:  "Test Description",
		StorageKey:   storageKey,
		Variables:    datatypes.JSON(`["Name"]`),
		SampleData:   datatypes.JSON(`{"Name": "John"}`),
		IsActive:     true,
		IsDefault:    false,
	}

	err := db.Create(template).Error
	require.NoError(t, err)

	// Store content in mock storage
	mockStorage.storedFiles["test-storage-key"] = []byte(content)

	// Get the template
	result, err := service.GetTemplate(context.Background(), 1, template.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Template", result.Name)
}

func TestTemplateService_ListTemplates(t *testing.T) {
	service, db, _ := setupTemplateService(t)

	// Create multiple templates
	templates := []*entities.Template{
		{
			TenantID:     1,
			TemplateType: "email",
			TemplateKey:  "template1",
			Channel:      "EMAIL",
			Name:         "Template 1",
			StorageKey:   "key1",
			IsActive:     true,
		},
		{
			TenantID:     1,
			TemplateType: "document",
			TemplateKey:  "template2",
			Channel:      "PDF",
			Name:         "Template 2",
			StorageKey:   "key2",
			IsActive:     true,
		},
	}

	for _, template := range templates {
		err := db.Create(template).Error
		require.NoError(t, err)
	}

	// List all templates
	templatesList, total, err := service.ListTemplates(context.Background(), 1, nil, "", "", nil, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, templatesList, 2)
	assert.Equal(t, int64(2), total)

	// Filter by channel
	templatesList, total, err = service.ListTemplates(context.Background(), 1, nil, "EMAIL", "", nil, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, templatesList, 1)
	assert.Equal(t, "email", templatesList[0].TemplateType)
	assert.Equal(t, int64(1), total)
}

func TestTemplateService_DeleteTemplate(t *testing.T) {
	service, db, _ := setupTemplateService(t)

	// Create a template
	template := &entities.Template{
		TenantID:     1,
		TemplateType: "email",
		TemplateKey:  "test_key",
		Channel:      "EMAIL",
		Name:         "Test Template",
		StorageKey:   "test-key",
		IsActive:     true,
	}

	err := db.Create(template).Error
	require.NoError(t, err)

	// Delete the template
	err = service.DeleteTemplate(context.Background(), 1, template.ID)
	assert.NoError(t, err)

	// Verify template was deleted from database
	var count int64
	err = db.Model(&entities.Template{}).Where("id = ?", template.ID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
