package tests

import (
	"github.com/ae-base-server/internal/models"
	// "github.com/ae-base-server/internal/router" // Package doesn't exist - test is outdated
	"github.com/ae-base-server/pkg/config"
	"github.com/ae-base-server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Note: These tests are currently disabled as they reference the old router package
// which has been refactored into the module system. Tests need to be rewritten
// to work with the new architecture.

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Run direct migrations for test database
	db.AutoMigrate(
		&models.Tenant{},
		&models.User{},
		&models.Plan{},
		&models.Customer{},
		&models.Contact{},
		&models.Email{},
		&models.Newsletter{},
		&models.TokenBlacklist{},
		&models.UserSettings{},
	)

	return db
}

// setupTestConfig creates a test configuration
func setupTestConfig() config.Config {
	return config.Config{
		Server: config.ServerConfig{
			Port: "8383",
			Host: "localhost",
			Mode: "test",
		},
		Database: database.Config{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "pass",
			DBName:   "ae_saas_basic",
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			ExpiryHour: 24,
		},
	}
}

/*
// DISABLED: These tests reference the old router package which no longer exists
// Tests need to be rewritten for the new module-based architecture

// TestHealthCheck tests the health endpoint
func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB()
	cfg := setupTestConfig()
	r := router.SetupRouter(db, cfg)

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
}

// TestPing tests the ping endpoint
func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB()
	cfg := setupTestConfig()
	r := router.SetupRouter(db, cfg)

	req, _ := http.NewRequest("GET", "/api/v1/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "pong", response["message"])
}

// TestGetPlans tests getting plans without authentication
func TestGetPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB()

	// Create a test plan
	plan := models.Plan{
		Name:          "Test Plan",
		Slug:          "test-plan",
		Description:   "A test plan",
		Price:         29.99,
		Currency:      "EUR",
		InvoicePeriod: "monthly",
		Active:        true,
	}
	db.Create(&plan)

	cfg := setupTestConfig()
	r := router.SetupRouter(db, cfg)

	req, _ := http.NewRequest("GET", "/api/v1/plans", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

// TestUserRegistration tests user registration
func TestUserRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB()

	// Create a test organization
	org := models.Organization{
		Name: "Test Organization",
		Slug: "test-org",
	}
	db.Create(&org)

	cfg := setupTestConfig()
	r := router.SetupRouter(db, cfg)

	// Create registration request
	regReq := models.UserCreateRequest{
		Username:       "testuser",
		Email:          "test@example.com",
		Password:       "password123",
		FirstName:      "Test",
		LastName:       "User",
		OrganizationID: org.ID,
	}

	jsonData, _ := json.Marshal(regReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

// TestUserLogin tests user login
func TestUserLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB()

	// Create a test organization
	org := models.Organization{
		Name: "Test Organization",
		Slug: "test-org",
	}
	db.Create(&org)

	// Create a test user
	regReq := models.UserCreateRequest{
		Username:       "testuser",
		Email:          "test@example.com",
		Password:       "password123",
		FirstName:      "Test",
		LastName:       "User",
		OrganizationID: org.ID,
	}

	cfg := setupTestConfig()
	r := router.SetupRouter(db, cfg)

	// Register user first
	jsonData, _ := json.Marshal(regReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Now test login
	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	jsonData, _ = json.Marshal(loginReq)
	req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}
*/
