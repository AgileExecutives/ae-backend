// Package api provides public access to ae-saas-basic models
package api

import (
	"github.com/ae-saas-basic/ae-saas-basic/internal/models"
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
