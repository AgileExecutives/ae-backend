package services_test

import (
	"testing"
	"time"

	baseAuth "github.com/ae-base-server/pkg/auth"
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

func setupClientWithTokenDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entities.Client{}, &entities.RegistrationToken{}))
	return db
}

func TestClientService_GetAllClients(t *testing.T) {
	db := setupClientWithTokenDB(t)
	svc := services.NewClientService(db)

	dobTime := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	dob := entities.NullableDate{Time: &dobTime}
	tenantID := uint(1)

	for i, status := range []string{"active", "active", "inactive"} {
		_ = i
		req := entities.CreateClientRequest{FirstName: "Test", LastName: "User", DateOfBirth: dob, Status: status}
		_, err := svc.CreateClient(req, tenantID)
		require.NoError(t, err)
	}
	// Different tenant
	req2 := entities.CreateClientRequest{FirstName: "Other", LastName: "Tenant", DateOfBirth: dob}
	_, err := svc.CreateClient(req2, 2)
	require.NoError(t, err)

	t.Run("returns all for tenant", func(t *testing.T) {
		clients, total, err := svc.GetAllClients(1, 10, tenantID, "")
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, clients, 3)
	})

	t.Run("filter by status active", func(t *testing.T) {
		clients, total, err := svc.GetAllClients(1, 10, tenantID, "active")
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, clients, 2)
	})

	t.Run("filter by status inactive", func(t *testing.T) {
		clients, total, err := svc.GetAllClients(1, 10, tenantID, "inactive")
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, clients, 1)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		clients, total, err := svc.GetAllClients(1, 10, 2, "")
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, clients, 1)
	})

	t.Run("pagination", func(t *testing.T) {
		clients, total, err := svc.GetAllClients(1, 2, tenantID, "")
		require.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Len(t, clients, 2)
	})
}

func TestClientService_RegistrationToken(t *testing.T) {
	db := setupClientWithTokenDB(t)
	svc := services.NewClientService(db)

	t.Run("generate and validate token", func(t *testing.T) {
		token, err := svc.GenerateRegistrationToken(1, 1, 10, "test@example.com")
		require.NoError(t, err)
		assert.NotEmpty(t, token.Token)
		assert.Equal(t, uint(10), token.OrganizationID)
		assert.Equal(t, uint(1), token.TenantID)
		assert.Equal(t, "test@example.com", token.Email)
		assert.False(t, token.Blacklisted)

		validated, err := svc.ValidateRegistrationToken(token.Token)
		require.NoError(t, err)
		assert.Equal(t, token.ID, validated.ID)
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		_, err := svc.ValidateRegistrationToken("nonexistent-token")
		require.Error(t, err)
	})

	t.Run("blacklisted token returns error", func(t *testing.T) {
		token, err := svc.GenerateRegistrationToken(1, 2, 20, "blacklist@example.com")
		require.NoError(t, err)
		db.Model(token).Update("blacklisted", true)
		_, err = svc.ValidateRegistrationToken(token.Token)
		require.Error(t, err)
	})

	t.Run("generating new token blacklists old one for same org", func(t *testing.T) {
		old, err := svc.GenerateRegistrationToken(1, 3, 30, "old@example.com")
		require.NoError(t, err)
		_, err = svc.GenerateRegistrationToken(1, 3, 30, "new@example.com")
		require.NoError(t, err)
		_, err = svc.ValidateRegistrationToken(old.Token)
		require.Error(t, err)
	})

	t.Run("mark as used increments counter", func(t *testing.T) {
		token, err := svc.GenerateRegistrationToken(1, 4, 40, "used@example.com")
		require.NoError(t, err)
		assert.Equal(t, 0, token.UsedCount)
		err = svc.MarkRegistrationTokenAsUsed(token.ID)
		require.NoError(t, err)
		var updated entities.RegistrationToken
		db.First(&updated, token.ID)
		assert.Equal(t, 1, updated.UsedCount)
	})
}

func TestClientService_RegisterClientViaToken(t *testing.T) {
	db := setupClientWithTokenDB(t)
	svc := services.NewClientService(db)

	dobTime := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("registers client successfully", func(t *testing.T) {
		token, err := svc.GenerateRegistrationToken(1, 1, 10, "register@example.com")
		require.NoError(t, err)
		req := entities.ClientRegistrationRequest{
			FirstName:   "New",
			LastName:    "Client",
			Email:       "register@example.com",
			DateOfBirth: entities.NullableDate{Time: &dobTime},
		}
		client, err := svc.RegisterClientViaToken(token.Token, req)
		require.NoError(t, err)
		assert.Equal(t, "New", client.FirstName)
		assert.Equal(t, "waiting", client.Status)
		assert.Equal(t, uint(1), client.TenantID)
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		req := entities.ClientRegistrationRequest{FirstName: "X", LastName: "Y", Email: "x@example.com", DateOfBirth: entities.NullableDate{Time: &dobTime}}
		_, err := svc.RegisterClientViaToken("bad-token", req)
		require.Error(t, err)
	})

	t.Run("email mismatch with token email returns error", func(t *testing.T) {
		token, err := svc.GenerateRegistrationToken(1, 1, 50, "specific@example.com")
		require.NoError(t, err)
		req := entities.ClientRegistrationRequest{FirstName: "X", LastName: "Y", Email: "other@example.com", DateOfBirth: entities.NullableDate{Time: &dobTime}}
		_, err = svc.RegisterClientViaToken(token.Token, req)
		require.Error(t, err)
	})

	t.Run("duplicate email registration returns error", func(t *testing.T) {
		token, err := svc.GenerateRegistrationToken(1, 1, 60, "dup@example.com")
		require.NoError(t, err)
		req := entities.ClientRegistrationRequest{FirstName: "Dup", LastName: "User", Email: "dup@example.com", DateOfBirth: entities.NullableDate{Time: &dobTime}}
		_, err = svc.RegisterClientViaToken(token.Token, req)
		require.NoError(t, err)
		// Second token for same org, same email
		token2, err := svc.GenerateRegistrationToken(1, 1, 70, "dup@example.com")
		require.NoError(t, err)
		_, err = svc.RegisterClientViaToken(token2.Token, req)
		require.Error(t, err)
	})
}

func TestClientService_GenerateEmailVerificationToken(t *testing.T) {
	db := setupClientWithTokenDB(t)
	svc := services.NewClientService(db)

	t.Run("generates token for client with email", func(t *testing.T) {
		client := entities.Client{
			TenantID:  1,
			FirstName: "Email",
			LastName:  "Tester",
			Email:     "verify@test.com",
		}
		require.NoError(t, db.Create(&client).Error)

		token, err := svc.GenerateEmailVerificationToken(client.ID)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("returns error for client without email", func(t *testing.T) {
		client := entities.Client{
			TenantID:  1,
			FirstName: "No",
			LastName:  "Email",
		}
		require.NoError(t, db.Create(&client).Error)

		_, err := svc.GenerateEmailVerificationToken(client.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})

	t.Run("returns error for non-existent client", func(t *testing.T) {
		_, err := svc.GenerateEmailVerificationToken(9999)
		require.Error(t, err)
	})
}

func TestClientService_VerifyClientEmail(t *testing.T) {
	db := setupClientWithTokenDB(t)
	svc := services.NewClientService(db)

	t.Run("verifies email with valid token", func(t *testing.T) {
		client := entities.Client{
			TenantID:  1,
			FirstName: "Valid",
			LastName:  "Token",
			Email:     "valid-verify@test.com",
		}
		require.NoError(t, db.Create(&client).Error)

		// Generate a real token
		token, err := baseAuth.GenerateVerificationToken(client.Email, client.ID)
		require.NoError(t, err)

		result, err := svc.VerifyClientEmail(token)
		require.NoError(t, err)
		assert.True(t, result.EmailVerified)
	})

	t.Run("returns error for invalid token", func(t *testing.T) {
		_, err := svc.VerifyClientEmail("not-a-valid-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid verification token")
	})

	t.Run("returns error when client not found", func(t *testing.T) {
		// Generate token for non-existent client (email not in DB)
		token, err := baseAuth.GenerateVerificationToken("nonexistent@test.com", 99999)
		require.NoError(t, err)

		_, err = svc.VerifyClientEmail(token)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error when user ID mismatch", func(t *testing.T) {
		client := entities.Client{
			TenantID:  1,
			FirstName: "Mismatch",
			LastName:  "Test",
			Email:     "mismatch-verify@test.com",
		}
		require.NoError(t, db.Create(&client).Error)

		// Generate token for same email but different user ID
		token, err := baseAuth.GenerateVerificationToken(client.Email, 99999)
		require.NoError(t, err)

		_, err = svc.VerifyClientEmail(token)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "match")
	})
}

func TestClientService_SendVerificationEmail_NoEmailService(t *testing.T) {
	db := setupClientWithTokenDB(t)
	svc := services.NewClientService(db)

	t.Run("returns error when client not found", func(t *testing.T) {
		err := svc.SendVerificationEmail(9999, "some-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})

	t.Run("returns error when client has no email", func(t *testing.T) {
		client := entities.Client{
			TenantID:  1,
			FirstName: "No",
			LastName:  "Email",
		}
		require.NoError(t, db.Create(&client).Error)

		err := svc.SendVerificationEmail(client.ID, "some-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})

	t.Run("returns error when email service is nil", func(t *testing.T) {
		client := entities.Client{
			TenantID:  1,
			FirstName: "Has",
			LastName:  "Email",
			Email:     "has-email@test.com",
		}
		require.NoError(t, db.Create(&client).Error)

		err := svc.SendVerificationEmail(client.ID, "some-token")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email service")
	})
}
