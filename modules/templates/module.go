package templates

import (
	"context"

	"github.com/ae-base-server/pkg/core"
	"github.com/redis/go-redis/v9"
	"github.com/unburdy/templates-module/entities"
	"github.com/unburdy/templates-module/routes"
	"github.com/unburdy/templates-module/services"
	"github.com/unburdy/templates-module/services/storage"
)

// CoreModule implements the core.Module interface for the templates module
type CoreModule struct {
	templateService *services.TemplateService
	templateRoutes  *routes.TemplateRoutes
	redisClient     *redis.Client
	minioStorage    storage.DocumentStorage
}

// NewCoreModule creates a new templates module instance
func NewCoreModule() *CoreModule {
	return &CoreModule{}
}

// Name returns the module name
func (m *CoreModule) Name() string {
	return "templates"
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
	// MinIO storage and template service may already be created by service provider
	if m.minioStorage == nil {
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

	// Initialize services if not already created
	if m.templateService == nil {
		m.templateService = services.NewTemplateService(ctx.DB, m.minioStorage)
	}

	// Initialize routes
	m.templateRoutes = routes.NewTemplateRoutes(m.templateService, ctx.DB)

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
		entities.NewTemplateEntity(),
	}
}

// Routes returns the list of route handlers for this module
func (m *CoreModule) Routes() []core.RouteProvider {
	providers := []core.RouteProvider{}
	if m.templateRoutes != nil {
		providers = append(providers, m.templateRoutes)
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
		&templateServiceProvider{module: m},
	}
}

// templateServiceProvider implements core.ServiceProvider
type templateServiceProvider struct {
	module *CoreModule
}

func (p *templateServiceProvider) ServiceName() string {
	return "template_service"
}

func (p *templateServiceProvider) ServiceInterface() interface{} {
	return (*services.TemplateService)(nil)
}

func (p *templateServiceProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
	// Create the service here since Factory is called before Initialize
	// This means we need to set up storage here too
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

	// Create and store the service in the module
	p.module.templateService = services.NewTemplateService(ctx.DB, minioStorage)
	p.module.minioStorage = minioStorage

	return p.module.templateService, nil
}

// SwaggerPaths returns Swagger documentation paths
func (m *CoreModule) SwaggerPaths() []string {
	return []string{}
}

// GetTemplateService returns the template service instance
func (m *CoreModule) GetTemplateService() *services.TemplateService {
	return m.templateService
}
