package routes

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/modules/documents/handlers"
	"github.com/unburdy/unburdy-server-api/modules/documents/services"
)

// TemplateRoutes implements RouteProvider for template management endpoints
type TemplateRoutes struct {
	templateService *services.TemplateService
}

// NewTemplateRoutes creates a new TemplateRoutes instance
func NewTemplateRoutes(templateService *services.TemplateService) *TemplateRoutes {
	return &TemplateRoutes{
		templateService: templateService,
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
		templates.GET("/:id/content", handler.GetTemplateContent)
		templates.GET("/:id/preview", handler.PreviewTemplate)
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
	// Auth middleware is typically applied globally
	return []gin.HandlerFunc{}
}

// GetSwaggerTags returns Swagger tags for documentation
func (r *TemplateRoutes) GetSwaggerTags() []string {
	return []string{"Templates"}
}
