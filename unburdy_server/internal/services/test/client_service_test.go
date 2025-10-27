package test
package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Client{})
	return db
}

func TestClientService_CreateClient(t *testing.T) {
	db := setupTestDB()
	service := services.NewClientService(db)

	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	req := models.CreateClientRequest{
		FirstName:   "John",
		LastName:    "Doe",
		DateOfBirth: &dob,
	}

	client, err := service.CreateClient(req)

	assert.NoError(t, err)
	assert.Equal(t, "John", client.FirstName)
	assert.Equal(t, "Doe", client.LastName)
	assert.Equal(t, &dob, client.DateOfBirth)
}

func TestClientService_GetClientByID(t *testing.T) {
	db := setupTestDB()
	service := services.NewClientService(db)

	// Create a client first
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	req := models.CreateClientRequest{
		FirstName:   "Jane",
		LastName:    "Smith",
		DateOfBirth: &dob,
	}

	created, err := service.CreateClient(req)
	assert.NoError(t, err)

	// Get the client by ID
	retrieved, err := service.GetClientByID(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, "Jane", retrieved.FirstName)
	assert.Equal(t, "Smith", retrieved.LastName)
}

func TestClientService_UpdateClient(t *testing.T) {
	db := setupTestDB()
	service := services.NewClientService(db)

	// Create a client first
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	req := models.CreateClientRequest{
		FirstName:   "John",
		LastName:    "Doe",
		DateOfBirth: &dob,
	}

	created, err := service.CreateClient(req)
	assert.NoError(t, err)

	// Update the client
	newFirstName := "Jane"
	updateReq := models.UpdateClientRequest{
		FirstName: &newFirstName,
	}

	updated, err := service.UpdateClient(created.ID, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, "Jane", updated.FirstName)
	assert.Equal(t, "Doe", updated.LastName) // Should remain unchanged
}

func TestClientService_SearchClients(t *testing.T) {
	db := setupTestDB()
	service := services.NewClientService(db)

	// Create test clients
	clients := []models.CreateClientRequest{
		{FirstName: "John", LastName: "Doe"},
		{FirstName: "Jane", LastName: "Smith"},
		{FirstName: "Bob", LastName: "Johnson"},
	}

	for _, req := range clients {
		_, err := service.CreateClient(req)
		assert.NoError(t, err)
	}

	// Search for "John"
	results, total, err := service.SearchClients("John", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total) // Should find "John Doe" and "Bob Johnson"
	assert.Len(t, results, 2)
}