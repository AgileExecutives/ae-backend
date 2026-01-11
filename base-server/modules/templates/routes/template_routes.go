package routes

import (
	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/modules/templates/handlers"
	"github.com/ae-base-server/modules/templates/services"
	"github.com/ae-base-server/modules/templates/services/storage"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TemplateRoutes implements RouteProvider for template management endpoints
type TemplateRoutes struct {
	templateService *services.TemplateService
	contractService *services.ContractService
	minioStorage    *storage.MinIOStorage
	db              *gorm.DB
}

// NewTemplateRoutes creates a new TemplateRoutes instance
func NewTemplateRoutes(templateService *services.TemplateService, contractService *services.ContractService, minioStorage *storage.MinIOStorage, db *gorm.DB) *TemplateRoutes {
	return &TemplateRoutes{
		templateService: templateService,
		contractService: contractService,
		minioStorage:    minioStorage,
		db:              db,
	}
}

// RegisterRoutes registers all template management routes
func (r *TemplateRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	handler := handlers.NewTemplateHandler(r.templateService, ctx.DB)
	contractHandler := handlers.NewContractHandler(r.contractService)
	publicAssetHandler := handlers.NewPublicAssetHandler(r.minioStorage)

	// Public asset routes (no authentication required)
	public := router.Group("/public/templates")
	{
		public.GET("/assets/:tenant/:template/:file", publicAssetHandler.GetAsset)
	}

	// Contract routes
	contracts := router.Group("/templates/contracts")
	{
		contracts.POST("", contractHandler.RegisterContract)
		contracts.GET("", contractHandler.ListContracts)
		contracts.GET("/by-key/:module/:template_key", contractHandler.GetContract)
		contracts.GET("/:id", contractHandler.GetContractByID)
		contracts.PUT("/:id", contractHandler.UpdateContract)
		contracts.DELETE("/:id", contractHandler.DeleteContract)
	}

	// Template routes
	templates := router.Group("/templates")
	{
		// Get default template (must be before /:id to avoid matching)
		templates.GET("/default", handler.GetDefaultTemplate)

		// Preview handlers removed - render service not yet implemented

		// CRUD operations
		templates.POST("", handler.CreateTemplate)
		templates.GET("", handler.ListTemplates)
		templates.GET("/:id", handler.GetTemplate)
		templates.PUT("/:id", handler.UpdateTemplate)
		templates.DELETE("/:id", handler.DeleteTemplate)

		// Template operations
		templates.POST("/:id/render", handler.RenderTemplate)
		templates.POST("/:id/duplicate", handler.DuplicateTemplate)
	}
}

// GetPrefix returns the base path for template routes
func (r *TemplateRoutes) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware to apply to all template routes
func (r *TemplateRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		baseAPI.AuthMiddleware(r.db), // Require authentication for tenant ID extraction
	}
}

// GetSwaggerTags returns Swagger tags for documentation
func (r *TemplateRoutes) GetSwaggerTags() []string {
	return []string{"Templates"}
}
