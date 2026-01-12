package templates

import (
	"context"
	"fmt"
	"os"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/ae-base-server/modules/templates/routes"
	"github.com/ae-base-server/modules/templates/services"
	"github.com/ae-base-server/modules/templates/services/storage"
	"github.com/ae-base-server/pkg/core"
	"github.com/redis/go-redis/v9"
)

// CoreModule implements the core.Module interface for the templates module
type CoreModule struct {
	templateService *services.TemplateService
	templateRoutes  *routes.TemplateRoutes
	redisClient     *redis.Client
	minioStorage    storage.DocumentStorage
	seedService     *services.SeedService
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

	// Initialize contract service
	contractService := services.NewContractService(ctx.DB)

	// Initialize seed service with template service for instance creation
	m.seedService = services.NewSeedServiceWithTemplateService(ctx.DB, m.templateService)

	// Seed default contract definitions (only runs once)
	if err := m.seedService.SeedDefaultTemplates(); err != nil {
		ctx.Logger.Warn("Failed to seed template contracts:", err)
	}

	// Seed default email templates for tenant 1 (unburdy management tenant) if not already seeded
	if err := m.seedDefaultTemplatesForManagementTenant(ctx); err != nil {
		ctx.Logger.Warn("Failed to seed default templates for management tenant:", err)
	}

	// Cast minioStorage to *storage.MinIOStorage for routes
	minioStorageImpl, ok := m.minioStorage.(*storage.MinIOStorage)
	if !ok {
		ctx.Logger.Warn("MinIO storage is not *storage.MinIOStorage, routes may not work correctly")
	}

	// Initialize routes
	m.templateRoutes = routes.NewTemplateRoutes(m.templateService, contractService, minioStorageImpl, ctx.DB)

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
		entities.NewTemplateContractEntity(),
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

// SeedEmailTemplatesForTenant seeds email templates for a specific tenant
// Call this after creating a new tenant to set up their default email templates
func (m *CoreModule) SeedEmailTemplatesForTenant(tenantID uint, organizationID *uint, serverDir string) error {
	if m.seedService == nil {
		return fmt.Errorf("seed service not initialized")
	}
	return m.seedService.SeedEmailTemplatesForTenant(tenantID, organizationID, serverDir)
}

// seedDefaultTemplatesForManagementTenant seeds default email templates for tenant 1 (management tenant)
// This runs during module initialization to ensure unburdy management has default templates
func (m *CoreModule) seedDefaultTemplatesForManagementTenant(ctx core.ModuleContext) error {
	ctx.Logger.Info("🚀 Starting template seeding for management tenants...")

	// Get server directory - try to find the server root
	serverDir, err := os.Getwd()
	if err != nil {
		ctx.Logger.Error("Failed to get working directory:", err)
		return err
	}
	ctx.Logger.Info(fmt.Sprintf("📂 Working directory: %s", serverDir))

	// Seed templates for tenant 1, organization 1 (Unburdy Verwaltung - Management)
	if err := m.seedTemplatesForTenantOrg(ctx, 1, 1, "management tenant", serverDir); err != nil {
		ctx.Logger.Warn("Failed to seed templates for tenant 1 org 1:", err)
	}

	// Seed templates for tenant 2, organization 2 (Standard Organisation - Default)
	if err := m.seedTemplatesForTenantOrg(ctx, 2, 2, "default tenant", serverDir); err != nil {
		ctx.Logger.Warn("Failed to seed templates for tenant 2 org 2:", err)
	}

	return nil
}

// seedTemplatesForTenantOrg seeds both email and PDF templates for a specific tenant/organization
func (m *CoreModule) seedTemplatesForTenantOrg(ctx core.ModuleContext, tenantID uint, orgID uint, label string, serverDir string) error {
	// Check if tenant exists
	var tenantExists bool
	if err := ctx.DB.Raw("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = ?)", tenantID).Scan(&tenantExists).Error; err != nil {
		return fmt.Errorf("failed to check if tenant %d exists: %w", tenantID, err)
	}

	if !tenantExists {
		ctx.Logger.Info(fmt.Sprintf("Tenant %d does not exist, skipping template seeding", tenantID))
		return nil
	}

	// Check if tenant already has templates
	var templateCount int64
	if err := ctx.DB.Model(&entities.Template{}).Where("tenant_id = ? AND organization_id = ?", tenantID, orgID).Count(&templateCount).Error; err != nil {
		return fmt.Errorf("failed to count templates for tenant %d org %d: %w", tenantID, orgID, err)
	}

	if templateCount > 0 {
		ctx.Logger.Info(fmt.Sprintf("Tenant %d org %d already has %d templates, skipping seeding", tenantID, orgID, templateCount))
		return nil
	}

	// Seed email templates
	ctx.Logger.Info(fmt.Sprintf("🌱 Seeding email templates for %s (tenant %d, org %d)...", label, tenantID, orgID))
	if err := m.seedService.SeedEmailTemplatesForTenant(tenantID, &orgID, serverDir); err != nil {
		ctx.Logger.Error(fmt.Sprintf("Failed to seed email templates for tenant %d org %d:", tenantID, orgID), err)
		return err
	}
	ctx.Logger.Info(fmt.Sprintf("✅ Seeded email templates for %s (tenant %d, org %d)", label, tenantID, orgID))

	// Seed PDF templates
	ctx.Logger.Info(fmt.Sprintf("🌱 Seeding PDF templates for %s (tenant %d, org %d)...", label, tenantID, orgID))
	if err := m.seedService.SeedPDFTemplatesForTenant(tenantID, &orgID, serverDir); err != nil {
		ctx.Logger.Error(fmt.Sprintf("Failed to seed PDF templates for tenant %d org %d:", tenantID, orgID), err)
		return err
	}
	ctx.Logger.Info(fmt.Sprintf("✅ Seeded PDF templates for %s (tenant %d, org %d)", label, tenantID, orgID))

	return nil
}

// CopyTemplatesFromTenant2Org2 copies all templates from tenant 2, org 2 to a target tenant/org
func (m *CoreModule) CopyTemplatesFromTenant2Org2(ctx context.Context, targetTenantID, targetOrganizationID uint) error {
	if m.templateService == nil {
		return fmt.Errorf("template service not initialized")
	}
	return m.templateService.CopyTemplatesFromTenant2Org2(ctx, targetTenantID, targetOrganizationID)
}

// GetTemplateByKey retrieves a template by its key for a given tenant and channel
// This method exposes the TemplateService's GetTemplateByKey to satisfy the TemplateGetter interface
func (m *CoreModule) GetTemplateByKey(ctx context.Context, tenantID uint, channel, templateKey string) (string, error) {
	if m.templateService == nil {
		return "", fmt.Errorf("template service not initialized")
	}
	return m.templateService.GetTemplateByKey(ctx, tenantID, channel, templateKey)
}
