package main

import (
	"log"
	"net/http"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/config"
	"github.com/ae-base-server/pkg/database"
	"github.com/joho/godotenv"
	calendar "github.com/unburdy/calendar-module"
	_ "github.com/unburdy/unburdy-server-api/docs" // swagger docs
	"github.com/unburdy/unburdy-server-api/modules/client_management"
)

// @title Unburdy Server - Modular API
// @version 2.0
// @description A modular SaaS backend API built with Go and Gin framework. Features a plugin-based architecture with four core modules: Base (authentication, users, tenants, contacts, newsletter), Customer (plans, customer management), Email (SMTP services, notifications), and PDF (document generation). Supports dependency injection, event-driven communication, and automatic module discovery.
// @termsOfService https://ae-base-server.com/terms

// @contact.name API Support
// @contact.url https://ae-base-server.com/support
// @contact.email support@ae-base-server.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name authentication
// @tag.description [Base Module] Authentication and user management endpoints including login, registration, password reset, and token management

// @tag.name users
// @tag.description [Base Module] User account management operations within tenant context

// @tag.name tenants
// @tag.description [Base Module] Multi-tenant organization management and configuration

// @tag.name contact-form
// @tag.description [Base Module] Public contact form submission and processing

// @tag.name newsletter
// @tag.description [Base Module] Newsletter subscription management and bulk operations

// @tag.name customers
// @tag.description [Customer Module] Customer relationship management and account operations

// @tag.name plans
// @tag.description [Customer Module] Subscription plan management and pricing configuration

// @tag.name emails
// @tag.description [Email Module] Email sending, tracking, and notification management with SMTP integration

// @tag.name pdf
// @tag.description [PDF Module] Document generation from templates with ChromeDP integration

// @tag.name health
// @tag.description [System] Application health checks, system status, and monitoring endpoints

// @tag.name modules
// @tag.description [System] Module registry, discovery, and runtime information endpoints

// @tag.name clients
// @tag.description [Client Management Module] Client information management, therapy tracking, and client-specific operations

// @tag.name cost-providers
// @tag.description [Client Management Module] Cost provider (insurance) management and approval tracking

// @tag.name calendar
// @tag.description [Calendar Module] Calendar management, scheduling, and event organization

// @tag.name calendar-entries
// @tag.description [Calendar Module] Individual calendar entries and event management

// @tag.name calendar-series
// @tag.description [Calendar Module] Recurring event series and pattern management

// @tag.name external-calendars
// @tag.description [Calendar Module] External calendar integration and synchronization

// @tag.name calendar-views
// @tag.description [Calendar Module] Calendar views including week, year, and custom period views

// @tag.name calendar-availability
// @tag.description [Calendar Module] Availability checking and free slot discovery

// @tag.name calendar-utilities
// @tag.description [Calendar Module] Calendar utilities including holiday import and data management

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration and connect to database
	cfg := config.Load()
	db, err := database.ConnectWithAutoCreate(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create external modules
	modules := []baseAPI.ModuleRouteProvider{
		client_management.NewModule(db), // Client management and cost provider tracking
		calendar.NewModule(db),          // Calendar management with events, series, and external calendars
	}

	// Setup modular router (includes base server + external modules)
	router := baseAPI.SetupModularRouter(db, modules)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	log.Printf("Health check available at http://%s/api/v1/health", addr)
	log.Printf("API documentation at http://%s/swagger/index.html", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
