package database

import (
	"fmt"
	"log"
	"os"

	baseAPI "github.com/ae-saas-basic/ae-saas-basic/api"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"gorm.io/gorm"
)

// SetupExtendedDatabase initializes the database with ae-saas-basic models + client tables
// Uses the same PostgreSQL database and configuration as ae-saas
func SetupExtendedDatabase() (*gorm.DB, error) {
	log.Println("🔧 Setting up database with ae-saas-basic configuration...")

	// Use ae-saas database configuration (same database as ae-saas)
	config := baseAPI.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "ae_saas_basic_test"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	log.Printf("📡 Connecting to PostgreSQL: %s:%s/%s", config.Host, config.Port, config.DBName)

	// Connect with auto-create (creates database if it doesn't exist)
	db, err := baseAPI.ConnectDatabaseWithAutoCreate(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Run ae-saas migrations (handles existing tables correctly)
	log.Println("📦 Running ae-saas-basic migrations...")
	err = baseAPI.MigrateDatabase(db)
	if err != nil {
		return nil, fmt.Errorf("failed to run ae-saas migrations: %w", err)
	}

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

	log.Println("✅ Database initialized: ae-saas-basic + Client + CostProvider models")
	log.Printf("✅ Database: %s@%s:%s/%s", config.User, config.Host, config.Port, config.DBName)
	return db, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
