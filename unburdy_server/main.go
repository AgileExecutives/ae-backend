package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/unburdy/unburdy-server-api/docs" // Swagger docs
	"github.com/unburdy/unburdy-server-api/internal/database"
	"github.com/unburdy/unburdy-server-api/internal/router"
)

// @title Unburdy Extended API
// @version 1.0
// @description Extended SaaS backend API built on AE SaaS Basic, adding client management functionality
// @termsOfService https://unburdy.com/terms

// @contact.name API Support
// @contact.url https://unburdy.com/support
// @contact.email support@unburdy.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name authentication
// @tag.description Authentication and user management endpoints (from base)

// @tag.name users
// @tag.description User management operations (from base)

// @tag.name customers
// @tag.description Customer management operations (from base)

// @tag.name clients
// @tag.description Client management operations (extended functionality)

// @tag.name contacts
// @tag.description Contact management operations (from base)

// @tag.name emails
// @tag.description Email management and sending operations (from base)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Initialize extended database (includes base + client tables)
	db, err := database.SetupExtendedDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Set up extended router (includes all base routes + client routes)
	r := router.SetupExtendedRouter(db)

	// Start server
	log.Println("🚀 Unburdy Extended API Server starting on :8080")
	log.Println("📋 Swagger documentation available at http://localhost:8080/swagger/index.html")
	log.Println("🔧 Includes all AE Base Server endpoints plus client management")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
