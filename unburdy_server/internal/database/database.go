package database

import (
	"fmt"
	"log"
	"os"
	"time"

	baseAPI "github.com/ae/base-server/api"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"gorm.io/gorm"
)

// SetupExtendedDatabase initializes the database with ae-base-server models + client tables
// Uses the same PostgreSQL database and configuration as ae-base-server
func SetupExtendedDatabase() (*gorm.DB, error) {
	log.Println("🔧 Setting up database with ae-base-server configuration...")

	// Use ae-base-server database configuration (same database as ae-base-server)
	dbConfig := baseAPI.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "pass"),
		DBName:   getEnv("DB_NAME", "ae_base_server_test"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	log.Printf("📡 Connecting to PostgreSQL: %s:%s/%s", dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	// Connect with auto-create (creates database if it doesn't exist)
	// Retry connection with exponential backoff for database startup
	var db *gorm.DB
	var err error
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		db, err = baseAPI.ConnectDatabaseWithAutoCreate(dbConfig)
		if err == nil {
			break
		}

		if i == maxRetries-1 {
			return nil, fmt.Errorf("failed to connect to PostgreSQL after %d retries: %w", maxRetries, err)
		}

		waitTime := time.Duration(i+1) * time.Second
		log.Printf("🔄 Database connection failed (attempt %d/%d), retrying in %v... Error: %v", i+1, maxRetries, waitTime, err)
		time.Sleep(waitTime)
	}

	// Note: ae-base-server migrations are now handled by the bootstrap system
	// This function only handles unburdy-specific models (Client and CostProvider)

	// Run migrations for unburdy-specific models (Client and CostProvider)
	log.Println("📦 Running unburdy-specific migrations (Client and CostProvider tables)...")

	// First, let AutoMigrate create the tables if they don't exist
	log.Println("Creating tables if they don't exist...")
	err = db.AutoMigrate(&models.CostProvider{}, &models.Client{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate Client and CostProvider models: %w", err)
	}

	log.Println("✅ Client and CostProvider tables migrated successfully")

	// Seed initial data
	log.Println("🌱 Seeding database from seed-data.json...")
	err = SeedDatabase(db)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to seed database: %v", err)
	}

	log.Println("✅ Database initialized: ae-base-server + Client + CostProvider models")
	log.Printf("✅ Database: %s@%s:%s/%s", dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.DBName)
	return db, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
