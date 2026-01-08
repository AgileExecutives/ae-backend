package routes

import (
	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/templates-module/handlers"
	"github.com/unburdy/templates-module/services"
	"gorm.io/gorm"
)

// TemplateRoutes implements RouteProvider for template management endpoints
type TemplateRoutes struct {
	templateService *services.TemplateService
	db              *gorm.DB
}

// NewTemplateRoutes creates a new TemplateRoutes instance
func NewTemplateRoutes(templateService *services.TemplateService, db *gorm.DB) *TemplateRoutes {
	return &TemplateRoutes{
		templateService: templateService,
		db:              db,
	}
}

// RegisterRoutes registers all template management routes
func (r *TemplateRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	handler := handlers.NewTemplateHandler(r.templateService, ctx.DB)

	// Template routes
	templates := router.Group("/templates")
	{
		// Get default template (must be before /:id to avoid matching)
		templates.GET("/default", handler.GetDefaultTemplate)

		// CRUD operations
		templates.POST("", handler.CreateTemplate)
		templates.GET("", handler.ListTemplates)
		templates.GET("/:id", handler.GetTemplate)
		templates.PUT("/:id", handler.UpdateTemplate)
		templates.DELETE("/:id", handler.DeleteTemplate)

		// Template operations
		templates.POST("/:id/render", handler.RenderTemplate)
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
