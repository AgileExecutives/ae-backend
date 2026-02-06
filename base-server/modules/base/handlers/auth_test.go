package handlers_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/ae-base-server/modules/base/handlers"
	"github.com/ae-base-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestAuthHandler_Login tests the login handler
func TestAuthHandler_Login(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing")
	defer os.Unsetenv("JWT_SECRET")

	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Setup
	logger := testutils.NewMockLogger()
	authHandlers := handlers.NewAuthHandlers(db, logger)
	router := testutils.SetupTestRouter()
	router.POST("/auth/login", authHandlers.Login)

	// Create test tenant and user with hashed password
	password := "SecurePass123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	tenant := testutils.CreateTestTenant(t, db, "Test Company")
	user := testutils.CreateTestUser(t, db, "test@example.com", string(hashedPassword), tenant.ID)

	// Update user to be verified and active
	db.Model(&user).Updates(map[string]interface{}{
		"active":         true,
		"email_verified": true,
	})

	t.Run("successful login with valid credentials", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    user.Email,
			"password": password,
		}

		w := testutils.MakeJSONRequest(t, router, "POST", "/auth/login", payload)
		assert.Equal(t, 200, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Check response structure
		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "token")
		assert.NotEmpty(t, data["token"])

		// JWT should have 3 parts separated by dots
		token := data["token"].(string)
		parts := bytes.Split([]byte(token), []byte("."))
		assert.Len(t, parts, 3, "JWT should have header.payload.signature")
		assert.NotEmpty(t, parts[0], "JWT header should not be empty")
		assert.NotEmpty(t, parts[1], "JWT payload should not be empty")
		assert.NotEmpty(t, parts[2], "JWT signature should not be empty")

		// Check user data in response
		assert.Contains(t, data, "user")
		userData := data["user"].(map[string]interface{})
		assert.Equal(t, user.Email, userData["email"])
		assert.Equal(t, user.Username, userData["username"])
	})

	t.Run("login fails with incorrect password", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    user.Email,
			"password": "WrongPassword123!",
		}

		w := testutils.MakeJSONRequest(t, router, "POST", "/auth/login", payload)
		assert.Equal(t, 401, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
	})

	t.Run("login fails for non-existent user", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    "nonexistent@example.com",
			"password": password,
		}

		w := testutils.MakeJSONRequest(t, router, "POST", "/auth/login", payload)
		assert.Equal(t, 401, w.Code)
	})

	t.Run("login fails for inactive user", func(t *testing.T) {
		// Create inactive user
		inactiveUser := testutils.CreateTestUser(t, db, "inactive@example.com", string(hashedPassword), tenant.ID)
		db.Model(&inactiveUser).Updates(map[string]interface{}{
			"active":         false,
			"email_verified": true,
		})

		payload := map[string]interface{}{
			"email":    inactiveUser.Email,
			"password": password,
		}

		w := testutils.MakeJSONRequest(t, router, "POST", "/auth/login", payload)
		assert.Equal(t, 401, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
	})

	t.Run("login fails with invalid email format", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    "not-an-email",
			"password": password,
		}

		w := testutils.MakeJSONRequest(t, router, "POST", "/auth/login", payload)
		assert.Equal(t, 400, w.Code)
	})

	t.Run("login fails with missing password", func(t *testing.T) {
		payload := map[string]interface{}{
			"email": user.Email,
		}

		w := testutils.MakeJSONRequest(t, router, "POST", "/auth/login", payload)
		assert.Equal(t, 400, w.Code)
	})
}
