// @title Minimal Server - Base API Only
// @version 1.0
// @description A minimal example server using only the ae-base-server functionality. Demonstrates core SaaS features including authentication, user management, customers, contacts, and email services.
// @termsOfService https://ae-base-server.com/terms

// @contact.name API Support
// @contact.url https://ae-base-server.com/support
// @contact.email support@ae-base-server.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

// @tag.name auth
// @tag.description [Base] User authentication and session management

// @tag.name users
// @tag.description [Base] User account management and profiles

// @tag.name customers
// @tag.description [Base] Customer relationship management

// @tag.name contacts
// @tag.description [Base] Contact management and newsletter subscriptions

// @tag.name emails
// @tag.description [Base] Email sending and management

// @tag.name health
// @tag.description [System] Health checks and system status

package main

import (
	"log"
	"os"

	"github.com/ae-base-server/api"
)

func main() {
	// Database configuration from environment variables
	dbConfig := api.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "ae_minimal"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Initialize database with auto-create
	db, err := api.ConnectDatabaseWithAutoCreate(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Seed base data
	if err := api.SeedBaseData(db); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	// Define modules to include
	var modules []api.ModuleRouteProvider

	// Add ping module - a simple example module
	pingModule := NewPingModule(db)
	modules = append(modules, pingModule)

	// TODO: Add more modules here, for example:
	// calendarModule := calendar.NewModule(db)
	// modules = append(modules, calendarModule)

	// Setup modular router
	router := api.SetupModularRouter(db, modules)

	// Register public routes for modules that need them
	pingModule.RegisterPublicRoutes(router)

	// Start server
	port := getEnv("PORT", "8080")

	log.Printf("Minimal Server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("Base endpoints:")
	log.Printf("  - GET  /api/v1/health")
	log.Printf("  - POST /api/v1/auth/register")
	log.Printf("  - POST /api/v1/auth/login")
	log.Printf("  - GET  /api/v1/users/profile (protected)")
	log.Printf("  - GET  /api/v1/customers (protected)")
	log.Printf("  - GET  /api/v1/contacts (protected)")
	log.Printf("  - GET  /api/v1/emails (protected)")
	log.Printf("Ping module endpoints:")
	log.Printf("  - GET  /api/v1/ping/ping")
	log.Printf("  - GET  /api/v1/ping/protected-ping (protected)")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
