package email

import (
	"context"

	"github.com/ae-base-server/modules/email/entities"
	"github.com/ae-base-server/modules/email/events"
	"github.com/ae-base-server/modules/email/handlers"
	"github.com/ae-base-server/modules/email/services"
	"github.com/ae-base-server/pkg/core"
)

// EmailModule represents the email module
type EmailModule struct {
	emailEntity   *entities.EmailEntity
	emailHandler  *handlers.EmailHandler
	emailService  *services.EmailService
	eventHandlers []core.EventHandler
}

// NewEmailModule creates a new email module instance
func NewEmailModule() *EmailModule {
	return &EmailModule{}
}

// Name returns the module name
func (m *EmailModule) Name() string {
	return "email"
}

// Version returns the module version
func (m *EmailModule) Version() string {
	return "1.0.0"
}

// Description returns the module description
func (m *EmailModule) Description() string {
	return "Email management and notification system"
}

// Dependencies returns the module dependencies
func (m *EmailModule) Dependencies() []string {
	return []string{} // Email module has no dependencies
}

// Initialize initializes the email module
func (m *EmailModule) Initialize(ctx core.ModuleContext) error {
	ctx.Logger.Info("Initializing Email module...")

	// Initialize entities
	m.emailEntity = entities.NewEmailEntity()

	// Initialize services
	m.emailService = services.NewEmailService()

	// Initialize handlers
	m.emailHandler = handlers.NewEmailHandler(ctx.DB, m.emailService)

	// Initialize event handlers
	m.eventHandlers = []core.EventHandler{
		events.NewEmailSentHandler(ctx.DB),
		events.NewEmailFailedHandler(ctx.DB, ctx.Logger),
	}

	ctx.Logger.Info("Email module initialized successfully")
	return nil
}

// Start starts the email module
func (m *EmailModule) Start(ctx context.Context) error {
	return nil
}

// Stop stops the email module
func (m *EmailModule) Stop(ctx context.Context) error {
	return nil
}

// Entities returns the module entities
func (m *EmailModule) Entities() []core.Entity {
	return []core.Entity{
		m.emailEntity,
	}
}

// Routes returns the module route providers
func (m *EmailModule) Routes() []core.RouteProvider {
	return []core.RouteProvider{
		// Temporarily disabled - using internal handlers instead
		// m.emailHandler,
	}
}

// EventHandlers returns the module event handlers
func (m *EmailModule) EventHandlers() []core.EventHandler {
	return m.eventHandlers
}

// Services returns the module service providers
func (m *EmailModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{
		&EmailServiceProvider{
			emailService: m.emailService,
		},
	}
}

// Middleware returns the module middleware providers
func (m *EmailModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{
		// Email module has no middleware
	}
}

// SwaggerPaths returns the swagger documentation paths
func (m *EmailModule) SwaggerPaths() []string {
	return []string{
		// Swagger paths for email endpoints
	}
}

// EmailServiceProvider provides the email service for dependency injection
type EmailServiceProvider struct {
	emailService *services.EmailService
}

// ServiceName returns the service name
func (p *EmailServiceProvider) ServiceName() string {
	return "email-service"
}

// ServiceInterface returns the service interface
func (p *EmailServiceProvider) ServiceInterface() interface{} {
	return p.emailService
}

// Factory creates the service instance
func (p *EmailServiceProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	return p.emailService, nil
}
