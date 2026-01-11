package documents

import (
	"context"

	templateServices "github.com/ae-base-server/modules/templates/services"
	"github.com/ae-base-server/pkg/core"
	"github.com/redis/go-redis/v9"
	"github.com/unburdy/documents-module/entities"
	"github.com/unburdy/documents-module/routes"
	"github.com/unburdy/documents-module/services"
	"github.com/unburdy/documents-module/services/storage"
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
	if m.minioStorage == nil {
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

		m.minioStorage = minioStorage
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

	// Get template service from registry (injected by templates module)
	if service, ok := ctx.Services.Get("template_service"); ok {
		if templateSvc, ok := service.(*templateServices.TemplateService); ok {
			m.templateService = templateSvc
		}
	}

	// Initialize services
	m.documentService = services.NewDocumentService(ctx.DB, m.minioStorage)

	// PDF service no longer needs template service - orchestration happens in handler
	if m.pdfService == nil {
		m.pdfService = services.NewPDFService(ctx.DB, m.minioStorage)
	}

	// Initialize routes
	m.documentRoutes = routes.NewDocumentRoutes(m.documentService, ctx.DB)
	m.pdfRoutes = routes.NewPDFRoutes(m.pdfService, m.templateService, ctx.DB)

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
	return []core.ServiceProvider{
		&documentServiceProvider{module: m},
		&pdfServiceProvider{module: m},
		&documentStorageProvider{module: m},
	}
}

// documentServiceProvider exposes the Document service via the service registry
type documentServiceProvider struct {
	module *CoreModule
}

func (p *documentServiceProvider) ServiceName() string {
	return "document_service"
}

func (p *documentServiceProvider) ServiceInterface() interface{} {
	return (*services.DocumentService)(nil)
}

func (p *documentServiceProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	// Ensure storage is ready
	if p.module.minioStorage == nil {
		minioConfig := storage.MinIOConfig{
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin123",
			UseSSL:          false,
			Region:          "us-east-1",
		}

		minioStorage, err := storage.NewMinIOStorage(minioConfig)
		if err != nil {
			return nil, err
		}
		p.module.minioStorage = minioStorage
	}

	// Create Document service if not already present
	if p.module.documentService == nil {
		p.module.documentService = services.NewDocumentService(ctx.DB, p.module.minioStorage)
	}

	return p.module.documentService, nil
}

// pdfServiceProvider exposes the PDF service via the service registry
type pdfServiceProvider struct {
	module *CoreModule
}

func (p *pdfServiceProvider) ServiceName() string {
	return "pdf_service"
}

func (p *pdfServiceProvider) ServiceInterface() interface{} {
	return (*services.PDFService)(nil)
}

func (p *pdfServiceProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	// Ensure storage is ready
	if p.module.minioStorage == nil {
		minioConfig := storage.MinIOConfig{
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin123",
			UseSSL:          false,
			Region:          "us-east-1",
		}

		minioStorage, err := storage.NewMinIOStorage(minioConfig)
		if err != nil {
			return nil, err
		}
		p.module.minioStorage = minioStorage
	}

	// Create PDF service if not already present
	if p.module.pdfService == nil {
		p.module.pdfService = services.NewPDFService(ctx.DB, p.module.minioStorage)
	}

	return p.module.pdfService, nil
}

// documentStorageProvider exposes the MinIO storage via the service registry
type documentStorageProvider struct {
	module *CoreModule
}

func (p *documentStorageProvider) ServiceName() string {
	return "document-storage"
}

func (p *documentStorageProvider) ServiceInterface() interface{} {
	return (*storage.DocumentStorage)(nil)
}

func (p *documentStorageProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	// Ensure storage is ready
	if p.module.minioStorage == nil {
		minioConfig := storage.MinIOConfig{
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin123",
			UseSSL:          false,
			Region:          "us-east-1",
		}

		minioStorage, err := storage.NewMinIOStorage(minioConfig)
		if err != nil {
			return nil, err
		}
		p.module.minioStorage = minioStorage
	}

	return p.module.minioStorage, nil
}

// SwaggerPaths returns Swagger documentation paths
func (m *CoreModule) SwaggerPaths() []string {
	return []string{}
}
