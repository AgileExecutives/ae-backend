package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ae-base-server/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// SeedData represents the structure of seed data from JSON
type SeedData struct {
	Tenants []SeedTenant `json:"tenants"`
	Plans   []SeedPlan   `json:"plans"`
	Users   []SeedUser   `json:"users"`
}

// SeedTenant represents tenant seed data
type SeedTenant struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// SeedPlan represents plan seed data
type SeedPlan struct {
	Name          string                 `json:"name"`
	Slug          string                 `json:"slug"`
	Description   string                 `json:"description"`
	Price         float64                `json:"price"`
	Currency      string                 `json:"currency"`
	InvoicePeriod string                 `json:"invoice_period"`
	MaxUsers      int                    `json:"max_users"`
	MaxClients    int                    `json:"max_clients"`
	Features      map[string]interface{} `json:"features"`
	Active        bool                   `json:"active"`
}

// SeedUser represents user seed data
type SeedUser struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Role          string `json:"role"`
	Active        bool   `json:"active"`
	TenantSlug    string `json:"tenant_slug"`
	EmailVerified bool   `json:"email_verified"`
}

// CreateDatabaseIfNotExists creates the database if it doesn't exist
func CreateDatabaseIfNotExists(config Config) error {
	// Connect to PostgreSQL without specifying database (connect to 'postgres' database)
	adminDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent for admin operations
	})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}

	// Get the underlying SQL database
	sqlDB, err := adminDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database: %w", err)
	}
	defer sqlDB.Close()

	// Check if database exists
	var exists bool
	query := `SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)`
	err = adminDB.Raw(query, config.DBName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		log.Printf("Creating database '%s'...", config.DBName)
		createQuery := fmt.Sprintf("CREATE DATABASE %s", config.DBName)
		err = adminDB.Exec(createQuery).Error
		if err != nil {
			return fmt.Errorf("failed to create database '%s': %w", config.DBName, err)
		}
		log.Printf("Database '%s' created successfully", config.DBName)
	} else {
		log.Printf("Database '%s' already exists", config.DBName)
	}

	return nil
}

// Connect creates a database connection
func Connect(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure the connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

// ConnectWithAutoCreate creates the database if it doesn't exist, then connects to it
func ConnectWithAutoCreate(config Config) (*gorm.DB, error) {
	// First, ensure the database exists
	if err := CreateDatabaseIfNotExists(config); err != nil {
		return nil, err
	}

	// Then connect to the database
	return Connect(config)
}

// Migrate runs database migrations with table existence checking
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Check if we need to run migrations at all
	log.Println("Checking migration status...")

	var tableCount int64
	err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name IN ('tenants', 'users', 'plans', 'customers', 'contacts', 'newsletters', 'emails', 'token_blacklist', 'user_settings')").Scan(&tableCount).Error
	if err != nil {
		return fmt.Errorf("failed to check existing tables: %w", err)
	}

	if tableCount == 9 {
		log.Println("All core tables already exist, skipping core migrations")
		return nil
	}

	log.Printf("Found %d existing tables, running fresh migrations...", tableCount)

	// Drop all tables to avoid conflicts and recreate them
	log.Println("Dropping existing tables to avoid conflicts...")
	dropTables := []string{"emails", "contacts", "newsletters", "customers", "users", "plans", "tenants", "token_blacklist", "user_settings"}
	for _, table := range dropTables {
		err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error
		if err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		}
	}

	// Now run clean migrations
	log.Println("Running clean migrations...")
	models := []interface{}{
		&models.Tenant{},
		&models.Plan{},
		&models.Newsletter{},
		&models.Email{},
		&models.Contact{},
		&models.User{},
		&models.Customer{},
		&models.TokenBlacklist{},
		&models.UserSettings{},
	}

	for i, model := range models {
		log.Printf("Migrating model %d: %T", i+1, model)
		err := db.AutoMigrate(model)
		if err != nil {
			log.Printf("ERROR: Migration failed for model %T: %v", model, err)
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
		log.Printf("Successfully migrated model %T", model)
	}

	log.Println("Core database migrations completed successfully")
	return nil
}

// MigrateWithModules runs migrations including module models
func MigrateWithModules(db *gorm.DB, moduleModels []interface{}) error {
	// First run core migrations
	if err := Migrate(db); err != nil {
		return err
	}

	// Then run module migrations if any
	if len(moduleModels) > 0 {
		log.Printf("🗃️  Running module model migrations (%d models)...", len(moduleModels))
		for i, model := range moduleModels {
			log.Printf("   Migrating module model %d/%d: %T", i+1, len(moduleModels), model)
			if err := db.AutoMigrate(model); err != nil {
				return fmt.Errorf("failed to migrate module model %T: %w", model, err)
			}
		}
		log.Println("✅ Module model migrations completed successfully")
	}

	return nil
}

// loadSeedData loads seed data from JSON file
func loadSeedData() (*SeedData, error) {
	// Get the current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for seed-data.json in current directory or parent directories
	seedDataPath := filepath.Join(pwd, "seed-data.json")
	if _, err := os.Stat(seedDataPath); os.IsNotExist(err) {
		// Try parent directory (in case running from subdirectory)
		seedDataPath = filepath.Join(filepath.Dir(pwd), "seed-data.json")
		if _, err := os.Stat(seedDataPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("seed-data.json not found in current or parent directory")
		}
	}

	// Read the JSON file
	data, err := os.ReadFile(seedDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed-data.json: %w", err)
	}

	// Parse JSON data
	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("failed to parse seed-data.json: %w", err)
	}

	return &seedData, nil
}

// Seed adds initial data to the database
func Seed(db *gorm.DB) error {
	log.Println("Seeding database with initial data...")

	// Load seed data from JSON file
	seedData, err := loadSeedData()
	if err != nil {
		return fmt.Errorf("failed to load seed data: %w", err)
	}

	// Create tenants
	var tenantCount int64
	db.Model(&models.Tenant{}).Count(&tenantCount)
	if tenantCount == 0 {
		for _, tenantData := range seedData.Tenants {
			tenant := models.Tenant{
				Name: tenantData.Name,
				Slug: tenantData.Slug,
			}
			if err := db.Create(&tenant).Error; err != nil {
				return fmt.Errorf("failed to create tenant %s: %w", tenantData.Name, err)
			}
			log.Printf("Created tenant: %s", tenantData.Name)
		}
	}

	// Create plans
	var planCount int64
	db.Model(&models.Plan{}).Count(&planCount)
	if planCount == 0 {
		for _, planData := range seedData.Plans {
			// Convert features map to JSON string
			featuresJSON, err := json.Marshal(planData.Features)
			if err != nil {
				return fmt.Errorf("failed to marshal features for plan %s: %w", planData.Name, err)
			}

			plan := models.Plan{
				Name:          planData.Name,
				Slug:          planData.Slug,
				Description:   planData.Description,
				Price:         planData.Price,
				Currency:      planData.Currency,
				InvoicePeriod: planData.InvoicePeriod,
				MaxUsers:      planData.MaxUsers,
				MaxClients:    planData.MaxClients,
				Features:      string(featuresJSON),
				Active:        planData.Active,
			}
			if err := db.Create(&plan).Error; err != nil {
				return fmt.Errorf("failed to create plan %s: %w", planData.Name, err)
			}
			log.Printf("Created plan: %s", planData.Name)
		}
	}

	// Create users
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		for _, userData := range seedData.Users {
			// Find the tenant by slug
			var tenant models.Tenant
			if err := db.Where("slug = ?", userData.TenantSlug).First(&tenant).Error; err != nil {
				return fmt.Errorf("failed to find tenant with slug %s: %w", userData.TenantSlug, err)
			}

			// Hash the password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password for user %s: %w", userData.Username, err)
			}

			user := models.User{
				Username:      userData.Username,
				Email:         userData.Email,
				PasswordHash:  string(hashedPassword),
				FirstName:     userData.FirstName,
				LastName:      userData.LastName,
				TenantID:      tenant.ID,
				Role:          userData.Role,
				Active:        userData.Active,
				EmailVerified: userData.EmailVerified,
			}

			// Set EmailVerifiedAt if email is verified
			if userData.EmailVerified {
				now := db.NowFunc()
				user.EmailVerifiedAt = &now
			}

			if err := db.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user %s: %w", userData.Username, err)
			}
			log.Printf("Created user: %s (email_verified: %v)", userData.Username, userData.EmailVerified)
		}
	}

	log.Println("Database seeding completed successfully! 🎉")
	return nil
}

// GetDefaultConfig returns default database configuration
func GetDefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "password",
		DBName:   "ae_saas_basic",
		SSLMode:  "disable",
	}
}
