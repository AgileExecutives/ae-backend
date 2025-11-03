package handlers

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
)

// AuthRoutes provides authentication routes
type AuthRoutes struct {
	handlers *AuthHandlers
}

// NewAuthRoutes creates new auth route provider
func NewAuthRoutes(handlers *AuthHandlers) core.RouteProvider {
	return &AuthRoutes{
		handlers: handlers,
	}
}

func (r *AuthRoutes) GetPrefix() string {
	return "/auth"
}

func (r *AuthRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (r *AuthRoutes) GetSwaggerTags() []string {
	return []string{"authentication"}
}

func (r *AuthRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	router.POST("/login", r.handlers.Login)
	router.POST("/register", r.handlers.Register)
	router.POST("/logout", ctx.Auth.RequireAuth(), r.handlers.Logout)
	router.POST("/refresh", ctx.Auth.RequireAuth(), r.handlers.RefreshToken)
	router.GET("/me", ctx.Auth.RequireAuth(), r.handlers.Me)
	router.POST("/change-password", ctx.Auth.RequireAuth(), r.handlers.ChangePassword)
	router.GET("/verify-email/:token", r.handlers.VerifyEmail)
	router.POST("/forgot-password", r.handlers.ForgotPassword)
	router.POST("/new-password/:token", r.handlers.ResetPassword)
}

// ContactRoutes provides contact and newsletter routes
type ContactRoutes struct {
	handlers *ContactHandlers
}

// NewContactRoutes creates new contact route provider
func NewContactRoutes(handlers *ContactHandlers) core.RouteProvider {
	return &ContactRoutes{
		handlers: handlers,
	}
}

func (r *ContactRoutes) GetPrefix() string {
	return "/contacts"
}

func (r *ContactRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (r *ContactRoutes) GetSwaggerTags() []string {
	return []string{"contact-form", "newsletter"}
}

func (r *ContactRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	// Public contact form - no auth required
	router.POST("/form", r.handlers.SubmitContactForm)

	// Contact CRUD endpoints - require auth
	router.GET("", ctx.Auth.RequireAuth(), r.handlers.GetContacts)
	router.GET("/:id", ctx.Auth.RequireAuth(), r.handlers.GetContact)
	router.POST("", ctx.Auth.RequireAuth(), r.handlers.CreateContact)
	router.PUT("/:id", ctx.Auth.RequireAuth(), r.handlers.UpdateContact)
	router.DELETE("/:id", ctx.Auth.RequireAuth(), r.handlers.DeleteContact)

	// Newsletter management - requires auth
	router.GET("/newsletter", ctx.Auth.RequireAuth(), r.handlers.GetNewsletterSubscriptions)
	router.DELETE("/newsletter/unsubscribe", ctx.Auth.RequireAuth(), r.handlers.UnsubscribeFromNewsletter)
}

// UserSettingsRoutes provides user settings routes
type UserSettingsRoutes struct {
	handlers *UserSettingsHandlers
}

// NewUserSettingsRoutes creates new user settings route provider
func NewUserSettingsRoutes(handlers *UserSettingsHandlers) core.RouteProvider {
	return &UserSettingsRoutes{
		handlers: handlers,
	}
}

func (r *UserSettingsRoutes) GetPrefix() string {
	return "/user-settings"
}

func (r *UserSettingsRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (r *UserSettingsRoutes) GetSwaggerTags() []string {
	return []string{"user-settings"}
}

func (r *UserSettingsRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	// All user settings endpoints require authentication
	router.GET("", ctx.Auth.RequireAuth(), r.handlers.GetUserSettings)
	router.PUT("", ctx.Auth.RequireAuth(), r.handlers.UpdateUserSettings)
	router.POST("/reset", ctx.Auth.RequireAuth(), r.handlers.ResetUserSettings)
}
