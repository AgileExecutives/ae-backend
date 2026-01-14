package main

import (
	"context"
	"log"

	_ "github.com/ae-base-server/docs" // swagger docs
	"github.com/ae-base-server/modules/base"
	"github.com/ae-base-server/modules/customer"
	"github.com/ae-base-server/modules/email"
	invoicenumber "github.com/ae-base-server/modules/invoice_number"
	"github.com/ae-base-server/modules/organization"
	"github.com/ae-base-server/modules/pdf"
	"github.com/ae-base-server/modules/static"
	"github.com/ae-base-server/modules/templates"
	"github.com/ae-base-server/pkg/bootstrap"
	"github.com/ae-base-server/pkg/config"
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/startup"
	"github.com/joho/godotenv"
)

// @title AE Base Server - Modular API
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

// @tag.name static
// @tag.description [Static Module] Secure JSON file serving from statics directory with enhanced security validation

// @tag.name modules
// @tag.description [System] Module registry, discovery, and runtime information endpoints

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
		// In production, this will fail ValidateRequired() above
	}

	log.Println("✅ Configuration validated successfully")

	// Create application
	app := bootstrap.NewApplication(cfg)

	// Register modules in dependency order
	modules := []core.Module{
		base.NewBaseModule(),                   // Base authentication and user management
		customer.NewCustomerModule(),           // Customer and plan management
		organization.NewOrganizationModule(),   // Organization management within tenants
		invoicenumber.NewInvoiceNumberModule(), // Invoice number generation
		templates.NewCoreModule(),              // Template management (contracts and instances)
		email.NewEmailModule(),                 // Email management and notifications
		pdf.NewPDFModule(),                     // PDF generation services
		static.NewStaticModule(),               // Static JSON file serving
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
