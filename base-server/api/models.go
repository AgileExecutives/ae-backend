// Package api provides public access to ae-base-server models
package api

import (
	"github.com/ae-base-server/internal/models"
)

// Re-export commonly used models
type (
	User           = models.User
	UserResponse   = models.UserResponse
	Tenant         = models.Tenant
	TenantResponse = models.TenantResponse
	APIResponse    = models.APIResponse
	ErrorResponse  = models.ErrorResponse
	LoginRequest   = models.LoginRequest
	LoginResponse  = models.LoginResponse
)

// Helper functions
var (
	SuccessResponse   = models.SuccessResponse
	ErrorResponseFunc = models.ErrorResponseFunc
)
