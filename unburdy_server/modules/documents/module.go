package documents

import (
	"context"

	"github.com/ae-base-server/pkg/core"
	"github.com/redis/go-redis/v9"
	"github.com/unburdy/unburdy-server-api/modules/documents/entities"
	"github.com/unburdy/unburdy-server-api/modules/documents/routes"
	"github.com/unburdy/unburdy-server-api/modules/documents/services"
	"github.com/unburdy/unburdy-server-api/modules/documents/services/storage"
)

// CoreModule implements the core.Module interface for the documents module
type CoreModule struct {
	documentService      *services.DocumentService
	invoiceNumberService *services.InvoiceNumberService
	templateService      *services.TemplateService
	pdfService           *services.PDFService
	documentRoutes       *routes.DocumentRoutes
	invoiceNumberRoutes  *routes.InvoiceNumberRoutes
	templateRoutes       *routes.TemplateRoutes
	pdfRoutes            *routes.PDFRoutes
	redisClient          *redis.Client
	minioStorage         storage.DocumentStorage
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
	return []string{"base"} // Depends on base module for auth and database
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
		ctx.Logger.Warn("Redis connection failed, invoice number caching will be limited:", err)
	}

	// Store storage for later use
	m.minioStorage = minioStorage

	// Initialize services
	m.documentService = services.NewDocumentService(ctx.DB, minioStorage)
	m.invoiceNumberService = services.NewInvoiceNumberService(ctx.DB, m.redisClient)
	m.templateService = services.NewTemplateService(ctx.DB, minioStorage)
	m.pdfService = services.NewPDFService(ctx.DB, minioStorage, m.templateService)

	// Initialize routes
	m.documentRoutes = routes.NewDocumentRoutes(m.documentService)
	m.invoiceNumberRoutes = routes.NewInvoiceNumberRoutes(m.invoiceNumberService)
	m.templateRoutes = routes.NewTemplateRoutes(m.templateService)
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
		entities.NewTemplateEntity(),
		entities.NewInvoiceNumberEntity(),
		entities.NewInvoiceNumberLogEntity(),
	}
}

// Routes returns the list of route handlers for this module
func (m *CoreModule) Routes() []core.RouteProvider {
	providers := []core.RouteProvider{}
	if m.documentRoutes != nil {
		providers = append(providers, m.documentRoutes)
	}
	if m.invoiceNumberRoutes != nil {
		providers = append(providers, m.invoiceNumberRoutes)
	}
	if m.templateRoutes != nil {
		providers = append(providers, m.templateRoutes)
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
