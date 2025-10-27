// Package api provides public access to ae-saas-basic middleware
package api

import (
	"github.com/ae-saas-basic/ae-saas-basic/internal/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware returns the authentication middleware
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return middleware.AuthMiddleware(db)
}

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (uint, error) {
	return middleware.GetUserID(c)
}

// GetTenantID retrieves the tenant ID from the Gin context
func GetTenantID(c *gin.Context) (uint, error) {
	return middleware.GetTenantID(c)
}

// GetUser retrieves the user from the Gin context
func GetUser(c *gin.Context) (*User, error) {
	return middleware.GetUser(c)
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return middleware.RequireRole(requiredRoles...)
}

// RequireAdmin middleware checks if user is admin
func RequireAdmin() gin.HandlerFunc {
	return middleware.RequireAdmin()
}
