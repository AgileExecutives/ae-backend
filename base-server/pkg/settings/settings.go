package settings

import (
	"github.com/ae-base-server/pkg/settings/handlers"
	"github.com/ae-base-server/pkg/settings/repository"
	"github.com/ae-base-server/pkg/settings/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SettingsSystem holds all settings system components
type SettingsSystem struct {
	Repository *repository.SettingsRepository
	Service    *services.SettingsService
	Handler    *handlers.SettingsHandler
}

// NewSettingsSystem creates a new settings system
func NewSettingsSystem(db *gorm.DB) (*SettingsSystem, error) {
	// Initialize repository
	repo := repository.NewSettingsRepository(db)

	// Auto-migrate settings table
	if err := repo.AutoMigrate(); err != nil {
		return nil, err
	}

	// Initialize service
	service := services.NewSettingsService(repo)

	// Initialize handler
	handler := handlers.NewSettingsHandler(service)

	return &SettingsSystem{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}, nil
}

// RegisterRoutes registers all settings routes
func (s *SettingsSystem) RegisterRoutes(router *gin.Engine) {
	settingsGroup := router.Group("/api/v1/settings")
	{
		// System routes
		settingsGroup.GET("/health", s.Handler.HealthCheck)
		settingsGroup.GET("/modules", s.Handler.GetRegisteredModules)
		settingsGroup.GET("/version", s.Handler.GetVersion)

		// Organization routes
		orgGroup := settingsGroup.Group("/organizations/:organization_id")
		{
			// Basic operations
			orgGroup.GET("", s.Handler.GetOrganizationSettings)
			orgGroup.POST("", s.Handler.SetOrganizationSetting)

			// Individual setting operations
			orgGroup.PUT("/:domain/:key", s.Handler.UpdateOrganizationSetting)
			orgGroup.DELETE("/:domain/:key", s.Handler.DeleteOrganizationSetting)

			// Bulk operations
			orgGroup.POST("/bulk", s.Handler.BulkSetOrganizationSettings)

			// Domain operations
			orgGroup.GET("/domains", s.Handler.GetOrganizationDomains)
			orgGroup.GET("/domains/:domain", s.Handler.GetOrganizationDomainSettings)
			orgGroup.POST("/domains/:domain", s.Handler.SetOrganizationDomainSettings)
			orgGroup.DELETE("/domains/:domain", s.Handler.DeleteOrganizationDomainSettings)

			// Validation
			orgGroup.POST("/validate", s.Handler.ValidateOrganizationSettings)

			// Import/Export
			orgGroup.GET("/export", s.Handler.ExportOrganizationSettings)
			orgGroup.POST("/import", s.Handler.ImportOrganizationSettings)
		}
	}
}
