package routes

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/handlers"
	"github.com/unburdy/documents-module/middleware"
	"github.com/unburdy/documents-module/services"
)

// DocumentRoutes implements RouteProvider for document management endpoints
type DocumentRoutes struct {
	documentService *services.DocumentService
}

// NewDocumentRoutes creates a new DocumentRoutes instance
func NewDocumentRoutes(documentService *services.DocumentService) *DocumentRoutes {
	return &DocumentRoutes{
		documentService: documentService,
	}
}

// RegisterRoutes registers all document management routes
func (r *DocumentRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	documentHandler := handlers.NewDocumentHandler(r.documentService, ctx.DB)

	// Document routes with tenant isolation
	documents := router.Group("/documents")
	{
		// Upload a document
		documents.POST("", documentHandler.UploadDocument)

		// List documents with pagination and filters
		documents.GET("", documentHandler.ListDocuments)

		// Get document metadata (with tenant access check)
		documents.GET("/:id",
			middleware.EnsureTenantAccess(ctx.DB),
			documentHandler.GetDocument,
		)

		// Get document download URL (with tenant access check)
		documents.GET("/:id/download",
			middleware.EnsureTenantAccess(ctx.DB),
			documentHandler.DownloadDocument,
		)

		// Delete document (with tenant access check)
		documents.DELETE("/:id",
			middleware.EnsureTenantAccess(ctx.DB),
			documentHandler.DeleteDocument,
		)
	}
}

// GetPrefix returns the base path for document routes
func (r *DocumentRoutes) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware to apply to all document routes
func (r *DocumentRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// Auth middleware is typically applied globally or at router level
		// Add any document-specific middleware here
	}
}

// GetSwaggerTags returns Swagger tags for documentation
func (r *DocumentRoutes) GetSwaggerTags() []string {
	return []string{"Documents"}
}
