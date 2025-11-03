package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Authorization required", "Missing authorization header"))
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Invalid authorization format", "Use Bearer <token> format"))
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate JWT token
		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Invalid token", err.Error()))
			c.Abort()
			return
		}

		// Check if token is blacklisted
		var blacklistedToken models.TokenBlacklist
		if err := db.Where("token_id = ? AND expires_at > NOW()", claims.ID).First(&blacklistedToken).Error; err == nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Token blacklisted", "Token has been revoked"))
			c.Abort()
			return
		}

		// Get user from database
		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User associated with token not found"))
			c.Abort()
			return
		}

		// Check if user is active
		if !user.Active {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Account disabled", "User account is not active"))
			c.Abort()
			return
		}

		// Verify tenant matches token
		if user.TenantID != claims.TenantID {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Tenant mismatch", "User tenant does not match token"))
			c.Abort()
			return
		}

		// Set user and token in context
		c.Set("user", &user)
		c.Set("userID", user.ID)
		c.Set("token", tokenString)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
			c.Abort()
			return
		}

		user := userInterface.(*models.User)

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range requiredRoles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Insufficient permissions", "User does not have required role"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware checks if user is admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin", "super-admin")
}

// TenantIsolation middleware ensures data access is limited to user's organization
func TenantIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
			c.Abort()
			return
		}

		user := userInterface.(*models.User)
		c.Set("tenant_id", user.TenantID)

		c.Next()
	}
}

// GetUser retrieves the authenticated user from context
func GetUser(c *gin.Context) (*models.User, error) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("user not found in context")
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid user type in context")
	}

	return user, nil
}

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c *gin.Context) (uint, error) {
	user, err := GetUser(c)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// GetTenantID retrieves the tenant ID from context
func GetTenantID(c *gin.Context) (uint, error) {
	user, err := GetUser(c)
	if err != nil {
		return 0, err
	}
	return user.TenantID, nil
}
