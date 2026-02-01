package handlers

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
)

// PublicEmailRoutes handles public (unauthenticated) email routes
type PublicEmailRoutes struct {
	emailHandler *EmailHandler
}

// NewPublicEmailRoutes creates a new public email routes handler
func NewPublicEmailRoutes(emailHandler *EmailHandler) *PublicEmailRoutes {
	return &PublicEmailRoutes{
		emailHandler: emailHandler,
	}
}

// RegisterRoutes registers public email routes
func (h *PublicEmailRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	// Only register if MOCK_EMAIL is enabled - checked in the handler itself
	router.GET("/emails/latest-emails", h.emailHandler.GetLatestEmails)
}

// GetPrefix returns the route prefix
func (h *PublicEmailRoutes) GetPrefix() string {
	return ""
}

// GetMiddleware returns route middleware (none for public routes)
func (h *PublicEmailRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

// GetSwaggerTags returns swagger tags for documentation
func (h *PublicEmailRoutes) GetSwaggerTags() []string {
	return []string{"emails"}
}
