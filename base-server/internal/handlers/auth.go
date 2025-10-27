package handlers

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ae-saas-basic/ae-saas-basic/internal/config"
	"github.com/ae-saas-basic/ae-saas-basic/internal/models"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/auth"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/utils"
	"github.com/ae-saas-basic/ae-saas-basic/services"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		db:  db,
		cfg: config.Load(),
	}
}

// Login authenticates a user and returns a JWT token
// @Summary Login user
// @Description Authenticate user with username/email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.APIResponse{data=models.LoginResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	var user models.User
	// Find user by username or email
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Invalid credentials", "User not found"))
		return
	}

	// Check if user is active
	if !user.Active {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Account disabled", "User account is not active"))
		return
	}

	// Check if email is verified (only if email verification is required)
	if h.cfg.Email.RequireEmailVerification && !user.EmailVerified {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Email not verified", "Please verify your email address before logging in"))
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Invalid credentials", "Password mismatch"))
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID, user.TenantID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to generate token", err.Error()))
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Login successful", response))
}

// Logout blacklists the current JWT token
// @Summary Logout user
// @Description Blacklist the current JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from context (set by auth middleware)
	tokenString, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("No token found", "Token not provided"))
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}

	user := userInterface.(*models.User)

	// Parse token to get JTI and expiration
	tokenID, expiresAt, err := auth.ParseTokenClaims(tokenString.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid token", err.Error()))
		return
	}

	// Add token to blacklist
	blacklistEntry := models.TokenBlacklist{
		TokenID:   tokenID,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		Reason:    "User logout",
	}

	if err := h.db.Create(&blacklistEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to logout", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Logout successful", nil))
}

// Register creates a new user account
// @Summary Register new user
// @Description Create a new user account. If a user-signup token is provided in Authorization header, user will join that tenant. Otherwise, company_name is required to create new tenant.
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer token for user signup invitation"
// @Param request body models.UserCreateRequest true "User registration data"
// @Success 201 {object} models.APIResponse{data=models.UserResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	// Validate required terms acceptance
	if !req.AcceptTerms {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Terms not accepted", "You must accept the terms and conditions to register"))
		return
	}

	// Validate password against requirements
	if err := utils.ValidatePassword(req.Password); err != nil {
		log.Printf("❌ Register: Password validation failed: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid password", err.Error()))
		return
	}

	// Check if username or email already exists
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponseFunc("User already exists", "Username or email already taken"))
		return
	}

	var tenantID uint

	// Check if there's a user-signup token in the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		signupToken := authHeader[7:]

		// Try to validate as user-signup token
		if signupTenantID, err := auth.ValidateUserSignupToken(signupToken); err == nil {
			// Valid user-signup token found, join the specified tenant
			var tenant models.Tenant
			if err := h.db.First(&tenant, signupTenantID).Error; err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Tenant not found", "Signup token references invalid tenant"))
				return
			}

			tenantID = tenant.ID

			// Set default role for joining existing tenant via invitation
			if req.Role == "" {
				req.Role = "user"
			}

			log.Printf("✅ User signup via invitation token to tenant: %s (ID: %d)", tenant.Name, tenant.ID)
		} else {
			// Not a valid user-signup token, proceed with normal registration
			if req.CompanyName == "" {
				c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Company name required", "Company name is required to create a new tenant"))
				return
			}

			// Create new tenant - logic will be handled below
		}
	} else {
		// No token provided, company_name is required to create new tenant
		if req.CompanyName == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Company name required", "Company name is required to create a new tenant"))
			return
		}
	}

	// Create new tenant if tenantID is still 0 (no valid signup token was found)
	if tenantID == 0 {
		// Check if company name already exists
		var existingTenant models.Tenant
		if err := h.db.Where("name = ?", req.CompanyName).First(&existingTenant).Error; err == nil {
			c.JSON(http.StatusConflict, models.ErrorResponseFunc("Company already exists", "A company with this name already exists"))
			return
		}

		// Generate slug from company name
		slug := utils.GenerateSlug(req.CompanyName)

		// Check if slug already exists and make it unique if necessary
		var existingSlugs []string
		var tenants []models.Tenant
		h.db.Select("slug").Find(&tenants)
		for _, t := range tenants {
			existingSlugs = append(existingSlugs, t.Slug)
		}
		slug = utils.EnsureUniqueSlug(slug, existingSlugs)

		// Create new tenant
		tenant := models.Tenant{
			Name: req.CompanyName,
			Slug: slug,
		}

		if err := h.db.Create(&tenant).Error; err != nil {
			log.Printf("❌ Register: Failed to create tenant: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create tenant", err.Error()))
			return
		}

		tenantID = tenant.ID

		// User is admin of their own tenant
		req.Role = "admin"

		log.Printf("✅ Created new tenant: %s (ID: %d, Slug: %s)", tenant.Name, tenant.ID, tenant.Slug)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to hash password", err.Error()))
		return
	}

	// Create user
	emailVerified := !h.cfg.Email.RequireEmailVerification
	var emailVerifiedAt *time.Time
	if emailVerified {
		now := time.Now()
		emailVerifiedAt = &now
	}

	user := models.User{
		Username:        req.Username,
		Email:           req.Email,
		PasswordHash:    string(hashedPassword),
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Role:            req.Role,
		TenantID:        tenantID,
		Active:          true,
		EmailVerified:   emailVerified,
		EmailVerifiedAt: emailVerifiedAt,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create user", err.Error()))
		return
	}

	// Handle newsletter subscription if opted in
	if req.NewsletterOptIn {
		newsletter := models.Newsletter{
			Name:     req.FirstName + " " + req.LastName,
			Email:    req.Email,
			Interest: "general",
			Source:   "registration",
		}

		if err := h.db.Create(&newsletter).Error; err != nil {
			log.Printf("⚠️ Register: Failed to create newsletter subscription: %v (continuing anyway)", err)
			// Don't fail registration if newsletter subscription fails
		} else {
			log.Printf("✅ Newsletter subscription created for %s", req.Email)
		}
	}

	// Send verification email only if email verification is required
	if h.cfg.Email.RequireEmailVerification {
		// Generate email verification token
		verificationToken, err := auth.GenerateVerificationToken(user.Email, user.ID)
		if err != nil {
			log.Printf("❌ Register: Failed to generate verification token: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to generate verification token", err.Error()))
			return
		}

		// Send verification email
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:5173" // Default for development
		}
		verificationURL := frontendURL + "/verify-email?token=" + verificationToken

		emailService := services.NewEmailService()
		err = emailService.SendVerificationEmail(user.Email, user.FirstName, verificationURL)
		if err != nil {
			log.Printf("⚠️ Register: Failed to send verification email: %v (continuing anyway)", err)
			// Don't fail registration if email fails, user can request resend
		}
	}

	// Generate onboarding token for limited access
	onboardingToken, err := auth.GenerateOnboardingToken(user.ID, user.TenantID, user.Role)
	if err != nil {
		log.Printf("❌ Register: Failed to generate onboarding token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to generate onboarding token", err.Error()))
		return
	}

	// Return user data with onboarding token
	response := models.LoginResponse{
		Token: onboardingToken,
		User:  user.ToResponse(),
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("User created successfully. Please check your email to verify your account.", response))
}

// ChangePassword changes the user's password
// @Summary Change password
// @Description Change user password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{current_password=string,new_password=string} true "Password change data"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ ChangePassword: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	log.Printf("🔍 ChangePassword: Received request - has_current: %v, has_new: %v", req.CurrentPassword != "", req.NewPassword != "")

	// Validate new password against requirements
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		log.Printf("❌ ChangePassword: Password validation failed: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid password", err.Error()))
		return
	}

	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}

	user := userInterface.(*models.User)

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid current password", "Current password is incorrect"))
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to hash password", err.Error()))
		return
	}

	// Update password
	if err := h.db.Model(user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to update password", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Password changed successfully", nil))
}

// Me returns the current user information
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.UserResponse}
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}

	user := userInterface.(*models.User)

	// Note: Tenant relation temporarily disabled due to GORM relation issues
	// h.db.Preload("Tenant").First(user, user.ID)

	c.JSON(http.StatusOK, models.SuccessResponse("User retrieved successfully", user.ToResponse()))
}

// ForgotPassword sends a password reset link to the user's email
// @Summary Request password reset
// @Description Send password reset link to user email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string} true "User email"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not for security
		c.JSON(http.StatusOK, models.SuccessResponse("If the email exists, a password reset link has been sent", nil))
		return
	}

	// Check if user is active
	if !user.Active {
		c.JSON(http.StatusOK, models.SuccessResponse("If the email exists, a password reset link has been sent", nil))
		return
	}

	// Get reset token expiry from env (default 2 hours)
	expiryStr := os.Getenv("RESET_TOKEN_EXPIRY")
	if expiryStr == "" {
		expiryStr = "2h"
	}
	expiryDuration, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiryDuration = 2 * time.Hour
	}

	// Generate reset token
	token, err := auth.GenerateResetToken(user.Email, expiryDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to generate reset token", err.Error()))
		return
	}

	// Build reset URL
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// Get reset route slug from env (default: /auth/new-password/)
	resetRouteSlug := os.Getenv("RESET_PASSWORD_ROUTE")
	if resetRouteSlug == "" {
		resetRouteSlug = "/auth/new-password/"
	}

	resetURL := frontendURL + resetRouteSlug + token

	// Send email
	emailService := services.NewEmailService()
	userName := user.FirstName
	if userName == "" {
		userName = user.Username
	}

	if err := emailService.SendPasswordResetEmail(user.Email, userName, resetURL); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to send reset email", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Password reset link has been sent to your email", nil))
}

// ResetPassword resets the user's password using a reset token
// @Summary Reset password
// @Description Reset user password with reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param token path string true "Reset token"
// @Param request body object{new_password=string} true "New password"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/new-password/{token} [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Reset token is required"))
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	// Validate password against requirements
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Password validation failed", err.Error()))
		return
	}

	// Validate reset token and get email
	email, err := auth.ValidateResetToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid or expired reset token", err.Error()))
		return
	}

	// Find user by email
	var user models.User
	if err := h.db.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("User not found", "Invalid reset token"))
		return
	}

	// Check if user is active
	if !user.Active {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Account disabled", "User account is not active"))
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to hash password", err.Error()))
		return
	}

	// Update password
	if err := h.db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to update password", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Password has been reset successfully", nil))
}

// GetPasswordSecurity godoc
// @Summary Get password security requirements
// @Description Returns the current password security requirements including minimum length and character requirements
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} utils.PasswordRequirements
// @Router /auth/password-security [get]
func (h *AuthHandler) GetPasswordSecurity(c *gin.Context) {
	requirements := utils.GetPasswordRequirements()
	c.JSON(http.StatusOK, requirements)
}

// VerifyEmail verifies a user's email address using the verification token
// @Summary Verify email address
// @Description Verify user email address with token from verification email
// @Tags auth
// @Accept json
// @Produce json
// @Param token path string true "Verification token"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /auth/verify-email/{token} [get]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Missing token", "Verification token is required"))
		return
	}

	// Validate verification token
	userID, email, err := auth.ValidateVerificationToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid or expired token", err.Error()))
		return
	}

	// Find user by ID and email (both must match)
	var user models.User
	if err := h.db.Where("id = ? AND email = ?", userID, email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("User not found", "Invalid verification token"))
		return
	}

	// Check if already verified
	if user.EmailVerified {
		c.JSON(http.StatusOK, models.SuccessResponse("Email already verified", nil))
		return
	}

	// Update user's email verification status
	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to verify email", err.Error()))
		return
	}

	log.Printf("✅ Email verified for user %d (%s)", user.ID, user.Email)
	c.JSON(http.StatusOK, models.SuccessResponse("Email verified successfully. You can now log in.", nil))
}

// GenerateSignupToken generates a signup token for inviting users to a tenant
// @Summary Generate user signup token
// @Description Generate a token for inviting users to join a specific tenant
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=object{token=string,expires_in=string}}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /auth/generate-signup-token [post]
func (h *AuthHandler) GenerateSignupToken(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("User not found", "User not authenticated"))
		return
	}

	user := userInterface.(*models.User)

	// Check if user has permission to invite others (admin, owner, or super-admin)
	if user.Role != "admin" && user.Role != "owner" && user.Role != "super-admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponseFunc("Permission denied", "Only admins and owners can generate signup tokens"))
		return
	}

	// Generate signup token for the user's tenant
	token, err := auth.GenerateUserSignupToken(user.TenantID, user.Email)
	if err != nil {
		log.Printf("❌ GenerateSignupToken: Failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to generate token", err.Error()))
		return
	}

	response := map[string]interface{}{
		"token":      token,
		"expires_in": "7 days",
		"tenant_id":  user.TenantID,
		"usage":      "Include this token in the Authorization header (Bearer token) when registering new users to add them to your tenant",
	}

	log.Printf("✅ Generated signup token for tenant %d by user %d (%s)", user.TenantID, user.ID, user.Email)
	c.JSON(http.StatusOK, models.SuccessResponse("Signup token generated successfully", response))
}
