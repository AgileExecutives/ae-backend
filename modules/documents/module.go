package documents

import (
	"context"

	"github.com/ae-base-server/pkg/core"
	"github.com/redis/go-redis/v9"
	"github.com/unburdy/documents-module/entities"
	"github.com/unburdy/documents-module/routes"
	"github.com/unburdy/documents-module/services"
	"github.com/unburdy/documents-module/services/storage"
	templateServices "github.com/unburdy/templates-module/services"
)

// CoreModule implements the core.Module interface for the documents module
type CoreModule struct {
	documentService *services.DocumentService
	pdfService      *services.PDFService
	documentRoutes  *routes.DocumentRoutes
	pdfRoutes       *routes.PDFRoutes
	redisClient     *redis.Client
	minioStorage    storage.DocumentStorage
	templateService *templateServices.TemplateService // From templates module
}

// NewCoreModule creates a new documents module instance
func NewCoreModule() *CoreModule {
	return &CoreModule{}
}

// Name returns the module name
func (m *CoreModule) Name() string {
	return "documents"
}

// Version returns the module version
func (m *CoreModule) Version() string {
	return "1.0.0"
}

// Dependencies returns module dependencies
func (m *CoreModule) Dependencies() []string {
	return []string{"base", "templates"} // Depends on base module for auth and templates for template service
}

// Initialize sets up the module with dependencies
func (m *CoreModule) Initialize(ctx core.ModuleContext) error {
	// Initialize MinIO storage
	minioConfig := storage.MinIOConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin123",
		UseSSL:          false,
		Region:          "us-east-1",
	}

	minioStorage, err := storage.NewMinIOStorage(minioConfig)
	if err != nil {
		return err
	}

	// Initialize Redis client
	m.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "redis123",
		DB:       0,
	})

	// Test Redis connection
	if err := m.redisClient.Ping(context.Background()).Err(); err != nil {
		ctx.Logger.Warn("Redis connection failed:", err)
	}

	// Store storage for later use
	m.minioStorage = minioStorage

	// Initialize services
	m.documentService = services.NewDocumentService(ctx.DB, minioStorage)

	// Note: Template service should be injected from templates module
	// PDF service will work with nil template service for HTML-only generation
	m.pdfService = services.NewPDFService(ctx.DB, minioStorage, nil)

	// Initialize routes
	m.documentRoutes = routes.NewDocumentRoutes(m.documentService)
	m.pdfRoutes = routes.NewPDFRoutes(m.pdfService)

	return nil
}

// Start starts the module
func (m *CoreModule) Start(ctx context.Context) error {
	return nil
}

// Stop stops the module
func (m *CoreModule) Stop(ctx context.Context) error {
	if m.redisClient != nil {
		return m.redisClient.Close()
	}
	return nil
}

// Entities returns the list of entities for database migration
func (m *CoreModule) Entities() []core.Entity {
	return []core.Entity{
		entities.NewDocumentEntity(),
	}
}

// Routes returns the list of route handlers for this module
func (m *CoreModule) Routes() []core.RouteProvider {
	providers := []core.RouteProvider{}
	if m.documentRoutes != nil {
		providers = append(providers, m.documentRoutes)
	}
	if m.pdfRoutes != nil {
		providers = append(providers, m.pdfRoutes)
	}
	return providers
}

// EventHandlers returns event handlers for the module
func (m *CoreModule) EventHandlers() []core.EventHandler {
	return []core.EventHandler{}
}

// Middleware returns middleware providers
func (m *CoreModule) Middleware() []core.MiddlewareProvider {
	return []core.MiddlewareProvider{}
}

// Services returns service providers
func (m *CoreModule) Services() []core.ServiceProvider {
	return []core.ServiceProvider{}
}

// SwaggerPaths returns Swagger documentation paths
func (m *CoreModule) SwaggerPaths() []string {
	return []string{}
}
