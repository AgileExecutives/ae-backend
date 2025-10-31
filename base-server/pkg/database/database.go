// Package database provides reusable database connection utilities for ae-base-server applications.
//
// This package contains the core database connection logic that can be used by any application
// built on ae-base-server. It provides configuration management, connection pooling, and
// database creation utilities.
//
// For seeding operations, see the internal/database package which contains base-server
// specific seeding logic.
package database

import (
	"fmt"
	"log"

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
