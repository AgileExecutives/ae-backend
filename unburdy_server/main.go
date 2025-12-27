package main

import (
	"context"
	"log"

	"github.com/ae-base-server/modules/base"
	"github.com/ae-base-server/modules/customer"
	"github.com/ae-base-server/modules/email"
	"github.com/ae-base-server/modules/pdf"
	"github.com/ae-base-server/modules/static"
	"github.com/ae-base-server/pkg/bootstrap"
	"github.com/ae-base-server/pkg/config"
	"github.com/ae-base-server/pkg/core"
	"github.com/joho/godotenv"
	booking "github.com/unburdy/booking-module"
	calendar "github.com/unburdy/calendar-module"
	organization "github.com/unburdy/organization-module"
	_ "github.com/unburdy/unburdy-server-api/docs" // swagger docs - unburdy-specific
	"github.com/unburdy/unburdy-server-api/modules/client_management"
	documents "github.com/unburdy/unburdy-server-api/modules/documents"
)

// @title Unburdy Server API
// @version 1.0
// @description A modular SaaS backend API built with Go and Gin framework. Features a plugin-based architecture with four core modules: Base (authentication, users, tenants, contacts, newsletter), Customer (plans, customer management), Email (SMTP services, notifications), and PDF (document generation). Supports dependency injection, event-driven communication, and automatic module discovery.

// @contact.name API Support
// @contact.url https://ae-base-server.com/support
// @contact.email support@ae-base-server.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name authentication
// @tag.description [Base Module] Authentication and user management endpoints including login, registration, password reset, and token management

// @tag.name clients
// @tag.description [Client Management Module] Client information management, therapy tracking, and client-specific operations

// @tag.name cost-providers
// @tag.description [Client Management Module] Cost provider (insurance) management and approval tracking

// @tag.name sessions
// @tag.description [Client Management Module] Session management, booking, and therapy session tracking

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

// @tag.name calendar-utilities
// @tag.description [Calendar Module] Calendar utilities including holiday import and data management

// @tag.name booking
// @tag.description [Booking Module] Booking template/configuration management for appointment scheduling

// @tag.name organizations
// @tag.description [Organization Module] Organization management including billing, tax, and bank account information

// @tag.name Documents
// @tag.description [Documents Module] Document storage and management with MinIO integration

// @tag.name Invoice Numbers
// @tag.description [Documents Module] Sequential invoice number generation with Redis caching

// @tag.name Templates
// @tag.description [Documents Module] Template management for emails and PDFs with versioning

// @tag.name PDFs
// @tag.description [Documents Module] PDF generation from HTML and templates using Chromedp

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Validate required environment variables
	if err := config.ValidateRequired(); err != nil {
		log.Fatalf("❌ Configuration validation failed:\n%v", err)
	}

	// Load configuration
	cfg := config.Load()

	// Validate loaded configuration
	if err := cfg.Validate(); err != nil {
		log.Printf("⚠️  Configuration warnings:\n%v", err)
		// Don't fatal in development, just warn
	}

	log.Println("✅ Configuration validated successfully")

	// Create application
	app := bootstrap.NewApplication(cfg)

	// Register modules in dependency order
	modules := []core.Module{
		base.NewBaseModule(),              // Base authentication and user management
		customer.NewCustomerModule(),      // Customer and plan management
		email.NewEmailModule(),            // Email management and notifications
		pdf.NewPDFModule(),                // PDF generation services
		static.NewStaticModule(),          // Static JSON file serving
		documents.NewCoreModule(),         // Documents, templates, invoices, PDF generation
		calendar.NewCoreModule(),          // Calendar management (unburdy-specific)
		booking.NewCoreModule(),           // Booking management (unburdy-specific)
		organization.NewCoreModule(),      // Organization management (unburdy-specific)
		client_management.NewCoreModule(), // Client management (unburdy-specific) - depends on booking
	}

	for _, module := range modules {
		if err := app.RegisterModule(module); err != nil {
			log.Fatalf("Failed to register module %s: %v", module.Name(), err)
		}
	}

	// Initialize application
	if err := app.Initialize(); err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// Start application
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		log.Fatal("Failed to start application:", err)
	}
}
