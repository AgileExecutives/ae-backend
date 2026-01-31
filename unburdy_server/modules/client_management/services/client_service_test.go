package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the Client schema
	err = db.AutoMigrate(&entities.Client{})
	require.NoError(t, err)

	return db
}

// TestClientService_CreateClient tests creating a new client
func TestClientService_CreateClient(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewClientService(db)

	tenantID := uint(1)
	dobTime := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	dateOfBirth := entities.NullableDate{Time: &dobTime}

	req := entities.CreateClientRequest{
		FirstName:   "John",
		LastName:    "Doe",
		DateOfBirth: dateOfBirth,
		Gender:      "male",
		Email:       "john.doe@example.com",
		Phone:       "+49123456789",
	}

	client, err := service.CreateClient(req, tenantID)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "John", client.FirstName)
	assert.Equal(t, "Doe", client.LastName)
	assert.Equal(t, "male", client.Gender)
	assert.Equal(t, "john.doe@example.com", client.Email)
	assert.Equal(t, tenantID, client.TenantID)
}

// TestClientService_UpdateClient tests updating an existing client
func TestClientService_UpdateClient(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewClientService(db)

	tenantID := uint(1)
	dobTime := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	dateOfBirth := entities.NullableDate{Time: &dobTime}

	// Create initial client
	createReq := entities.CreateClientRequest{
		FirstName:   "John",
		LastName:    "Doe",
		DateOfBirth: dateOfBirth,
		Email:       "john.doe@example.com",
	}

	client, err := service.CreateClient(createReq, tenantID)
	require.NoError(t, err)

	// Update the client
	newEmail := "john.updated@example.com"
	updateReq := entities.UpdateClientRequest{
		Email: &newEmail,
	}

	updated, err := service.UpdateClient(client.ID, tenantID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, newEmail, updated.Email)
	assert.Equal(t, "John", updated.FirstName) // Unchanged
}

// TestClientService_DeleteClient tests soft-deleting a client
func TestClientService_DeleteClient(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewClientService(db)

	tenantID := uint(1)
	dobTime := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	dateOfBirth := entities.NullableDate{Time: &dobTime}

	// Create a client
	req := entities.CreateClientRequest{
		FirstName:   "John",
		LastName:    "Doe",
		DateOfBirth: dateOfBirth,
	}

	client, err := service.CreateClient(req, tenantID)
	require.NoError(t, err)

	// Delete the client
	err = service.DeleteClient(client.ID, tenantID)
	require.NoError(t, err)

	// Try to get the deleted client
	_, err = service.GetClientByID(client.ID, tenantID)
	assert.Error(t, err) // Should not be found
}

// TestClientService_GetClientByID tests retrieving a client by ID
func TestClientService_GetClientByID(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewClientService(db)

	tenantID := uint(1)
	dobTime := time.Date(1985, 7, 10, 0, 0, 0, 0, time.UTC)
	dateOfBirth := entities.NullableDate{Time: &dobTime}

	// Create a client
	req := entities.CreateClientRequest{
		FirstName:   "Alice",
		LastName:    "Johnson",
		DateOfBirth: dateOfBirth,
		Gender:      "female",
	}

	created, err := service.CreateClient(req, tenantID)
	require.NoError(t, err)

	// Retrieve the client
	retrieved, err := service.GetClientByID(created.ID, tenantID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, "Alice", retrieved.FirstName)
}

// TestClientService_SearchClients tests searching for clients
func TestClientService_SearchClients(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewClientService(db)

	tenantID := uint(1)
	dobTime := time.Date(1992, 1, 1, 0, 0, 0, 0, time.UTC)
	dateOfBirth := entities.NullableDate{Time: &dobTime}

	// Create multiple clients
	clients := []entities.CreateClientRequest{
		{FirstName: "Alice", LastName: "Anderson", DateOfBirth: dateOfBirth},
		{FirstName: "Bob", LastName: "Brown", DateOfBirth: dateOfBirth},
		{FirstName: "Charlie", LastName: "Clark", DateOfBirth: dateOfBirth},
	}

	for _, req := range clients {
		_, err := service.CreateClient(req, tenantID)
		require.NoError(t, err)
	}

	// Search for "Alice"
	results, total, err := service.SearchClients("Alice", 1, 10, tenantID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1)
	assert.GreaterOrEqual(t, total, int64(1))

	found := false
	for _, client := range results {
		if client.FirstName == "Alice" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// TestClientService_ConcurrentCreates tests concurrent client creation
// Note: This test demonstrates SQLite locking behavior with concurrent writes
func TestClientService_ConcurrentCreates(t *testing.T) {
	t.Skip("SQLite has limited concurrency support - this test validates production DB (PostgreSQL) behavior")

	// This test would pass with PostgreSQL but SQLite locks on concurrent writes
	// Keeping the implementation for when integration tests use real DB

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.Client{})
	require.NoError(t, err)

	service := services.NewClientService(db)
	tenantID := uint(1)
	concurrency := 10

	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			dobTime := time.Date(1990+index, 1, 1, 0, 0, 0, 0, time.UTC)
			dateOfBirth := entities.NullableDate{Time: &dobTime}

			req := entities.CreateClientRequest{
				FirstName:   "User",
				LastName:    "Concurrent",
				DateOfBirth: dateOfBirth,
			}

			service.CreateClient(req, tenantID)
			done <- true
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}

	results, total, _ := service.SearchClients("Concurrent", 1, 100, tenantID)
	t.Logf("Created %d/%d concurrent clients (SQLite locking expected)", len(results), concurrency)
	assert.GreaterOrEqual(t, total, int64(1)) // At least some should succeed
}

// TestClientService_TenantIsolation tests tenant data isolation
func TestClientService_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewClientService(db)

	dobTime := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	dateOfBirth := entities.NullableDate{Time: &dobTime}

	// Create client for tenant 1
	req1 := entities.CreateClientRequest{
		FirstName:   "Tenant1",
		LastName:    "Client",
		DateOfBirth: dateOfBirth,
	}
	client1, err := service.CreateClient(req1, 1)
	require.NoError(t, err)

	// Create client for tenant 2
	req2 := entities.CreateClientRequest{
		FirstName:   "Tenant2",
		LastName:    "Client",
		DateOfBirth: dateOfBirth,
	}
	client2, err := service.CreateClient(req2, 2)
	require.NoError(t, err)

	// Tenant 1 should not see tenant 2's client
	_, err = service.GetClientByID(client2.ID, 1)
	assert.Error(t, err)

	// Tenant 2 should not see tenant 1's client
	_, err = service.GetClientByID(client1.ID, 2)
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkClientService_CreateClient(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&entities.Client{})
	service := services.NewClientService(db)

	tenantID := uint(1)
	dobTime := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	dateOfBirth := entities.NullableDate{Time: &dobTime}

	req := entities.CreateClientRequest{
		FirstName:   "Benchmark",
		LastName:    "User",
		DateOfBirth: dateOfBirth,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateClient(req, tenantID)
	}
}

func BenchmarkClientService_SearchClients(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&entities.Client{})
	service := services.NewClientService(db)

	tenantID := uint(1)
	dobTime := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	dateOfBirth := entities.NullableDate{Time: &dobTime}

	// Create 100 test clients
	for i := 0; i < 100; i++ {
		req := entities.CreateClientRequest{
			FirstName:   "User",
			LastName:    "Test",
			DateOfBirth: dateOfBirth,
		}
		service.CreateClient(req, tenantID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SearchClients("User", 1, 10, tenantID)
	}
}
