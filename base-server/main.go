package main

import (
	"log"
	"net/http"

	_ "github.com/ae-saas-basic/ae-saas-basic/docs" // swagger docs
	"github.com/ae-saas-basic/ae-saas-basic/internal/config"
	"github.com/ae-saas-basic/ae-saas-basic/internal/database"
	"github.com/ae-saas-basic/ae-saas-basic/internal/eventbus"
	"github.com/ae-saas-basic/ae-saas-basic/internal/modules"
	"github.com/ae-saas-basic/ae-saas-basic/internal/router"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/auth"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title AE SaaS Basic API
// @version 1.0
// @description A comprehensive SaaS backend API built with Go and Gin framework, providing authentication, user management, customer management, email handling, PDF generation, search functionality, and more.
// @termsOfService https://ae-saas-basic.com/terms

// @contact.name API Support
// @contact.url https://ae-saas-basic.com/support
// @contact.email support@ae-saas-basic.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name authentication
// @tag.description Authentication and user management endpoints

// @tag.name users
// @tag.description User management operations

// @tag.name customers
// @tag.description Customer management operations

// @tag.name contacts
// @tag.description Contact management operations

// @tag.name emails
// @tag.description Email management and sending operations

// @tag.name plans
// @tag.description Subscription plan management

// @tag.name user-settings
// @tag.description User preferences and settings management

// @tag.name pdf
// @tag.description PDF generation and template management

// @tag.name search
// @tag.description Fuzzy search across all entities

// @tag.name static
// @tag.description Static file serving and asset management

// @tag.name health
// @tag.description System health and status endpoints

// @tag.name newsletter
// @tag.description Newsletter subscription management

// @tag.name contact-form
// @tag.description Public contact form endpoints

// @tag.name demo
// @tag.description Demo module endpoints for testing modular architecture

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Set JWT secret
	auth.SetJWTSecret(cfg.JWT.Secret)

	// Connect to database (create database if it doesn't exist)
	db, err := database.ConnectWithAutoCreate(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize module manager with all available modules
	moduleManager, err := modules.CreateModuleManager(db, cfg)
	if err != nil {
		log.Fatal("Failed to create module manager:", err)
	}

	// Get all module models for migration
	var allModuleModels []interface{}
	for _, module := range moduleManager.GetRegistry().GetEnabledModules() {
		allModuleModels = append(allModuleModels, module.GetModels()...)
	}

	// Run migrations including module models
	if err := database.MigrateWithModules(db, allModuleModels); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed database with initial data
	if err := database.Seed(db); err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	// Initialize event bus and handlers
	eventbus.InitializeEventHandlers()

	// Initialize modules (this will register event handlers)
	if err := moduleManager.InitializeModules(); err != nil {
		log.Fatal("Failed to initialize modules:", err)
	}

	// Setup router with module support
	r := router.SetupRouterWithModules(db, cfg, moduleManager)

	// Register module routes
	moduleManager.RegisterModuleRoutes(r)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting AE SaaS Basic server on %s", addr)
	log.Printf("Health check available at: http://%s/api/v1/health", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
