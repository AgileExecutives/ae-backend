package middleware

import (
	internalMiddleware "github.com/ae-base-server/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware returns authentication middleware that can be used by modules
// This is a wrapper around the internal auth middleware to make it accessible
// to modules without exposing internal packages
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return internalMiddleware.AuthMiddleware(db)
}

// RequireRole returns middleware that requires specific user roles
func RequireRole(db *gorm.DB, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First apply auth middleware
		internalMiddleware.AuthMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		// Then check roles
		// TODO: Implement role checking logic
		c.Next()
	}
}
