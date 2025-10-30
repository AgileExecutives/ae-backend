package middleware

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	db     *gorm.DB
	logger core.Logger
}

// NewAuthMiddleware creates new auth middleware
func NewAuthMiddleware(db *gorm.DB, logger core.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		db:     db,
		logger: logger,
	}
}

// AuthMiddlewareProvider implements core.MiddlewareProvider for AuthMiddleware
type AuthMiddlewareProvider struct {
	middleware *AuthMiddleware
}

// NewAuthMiddlewareProvider creates a new auth middleware provider
func NewAuthMiddlewareProvider(middleware *AuthMiddleware) core.MiddlewareProvider {
	return &AuthMiddlewareProvider{
		middleware: middleware,
	}
}

func (p *AuthMiddlewareProvider) Name() string {
	return "auth"
}

func (p *AuthMiddlewareProvider) Handler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Implementation will be moved from internal/middleware/auth.go
		c.Next()
	})
}

func (p *AuthMiddlewareProvider) Priority() int {
	return 100
}

func (p *AuthMiddlewareProvider) ApplyTo() []string {
	return []string{} // Global middleware
}
