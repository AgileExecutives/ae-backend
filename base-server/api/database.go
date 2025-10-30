// Package api provides public access to ae-base-server database functions
package api

import (
	"log"

	"github.com/ae-base-server/internal/database"
	"github.com/ae-base-server/internal/models"
	"gorm.io/gorm"
)

// DatabaseConfig exports the database configuration type
type DatabaseConfig = database.Config

// ConnectDatabase connects to the database using the provided configuration
func ConnectDatabase(config DatabaseConfig) (*gorm.DB, error) {
	return database.Connect(config)
}

// ConnectDatabaseWithAutoCreate connects and creates the database if needed
func ConnectDatabaseWithAutoCreate(config DatabaseConfig) (*gorm.DB, error) {
	return database.ConnectWithAutoCreate(config)
}

// MigrateDatabase runs migrations for all base ae-saas models with proper migration logic
// This handles existing tables correctly and avoids migration conflicts (PostgreSQL)
func MigrateDatabase(db *gorm.DB) error {
	return database.Migrate(db)
}

// MigrateDatabaseSimple runs basic AutoMigrate for all base models
// Use this for SQLite or when you want simple migration without drop/recreate logic
func MigrateDatabaseSimple(db *gorm.DB) error {
	log.Println("Running simple AutoMigrate for all base models...")
	return db.AutoMigrate(
		&models.Tenant{},
		&models.Plan{},
		&models.User{},
		&models.Customer{},
		&models.Contact{},
		&models.Email{},
		&models.Newsletter{},
		&models.TokenBlacklist{},
		&models.UserSettings{},
	)
}

// SeedBaseData loads and seeds data from seed-data.json
func SeedBaseData(db *gorm.DB) error {
	return database.Seed(db)
}

// SetupDatabase is a convenience function that connects, migrates, and seeds
func SetupDatabase(config DatabaseConfig) (*gorm.DB, error) {
	// Connect with auto-create
	db, err := database.ConnectWithAutoCreate(config)
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		return nil, err
	}

	// Seed data
	if err := database.Seed(db); err != nil {
		return nil, err
	}

	return db, nil
}
