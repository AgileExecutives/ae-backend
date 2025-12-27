package middleware

import (
	"net/http"

	baseAPI "github.com/ae-base-server/api"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/documents-module/entities"
	"gorm.io/gorm"
)

// EnsureTenantAccess verifies that the requested document belongs to the authenticated tenant
func EnsureTenantAccess(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		documentID := c.Param("id")
		if documentID == "" {
			c.Next()
			return
		}

		// Get tenant ID from auth context
		tenantID, err := baseAPI.GetTenantID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized: failed to get tenant ID",
			})
			c.Abort()
			return
		}

		// Check if document belongs to tenant
		var doc entities.Document
		err = db.Where("id = ? AND tenant_id = ?", documentID, tenantID).First(&doc).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Access denied: document not found or does not belong to your organization",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to verify document access",
				})
			}
			c.Abort()
			return
		}

		// Store document in context for handlers to use
		c.Set("document", doc)
		c.Next()
	}
}
