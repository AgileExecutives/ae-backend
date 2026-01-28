package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	templateEntities "github.com/ae-base-server/modules/templates/entities"
	baseAuth "github.com/ae-base-server/pkg/auth"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/gorm"
)

type ClientService struct {
	db              *gorm.DB
	emailService    interface{} // Email service for sending verification emails
	templateService interface{} // Template service for rendering email templates
}

// NewClientService creates a new client service
func NewClientService(db *gorm.DB) *ClientService {
	return &ClientService{
		db:              db,
		emailService:    nil,
		templateService: nil,
	}
}

// SetEmailService sets the email service for dependency injection
func (s *ClientService) SetEmailService(emailService interface{}) {
	s.emailService = emailService
}

// SetTemplateService sets the template service for dependency injection
func (s *ClientService) SetTemplateService(templateService interface{}) {
	s.templateService = templateService
}

// CreateClient creates a new client for a specific tenant
func (s *ClientService) CreateClient(req entities.CreateClientRequest, tenantID uint) (*entities.Client, error) {
	client := entities.Client{
		TenantID:             tenantID,
		CostProviderID:       req.CostProviderID,
		FirstName:            req.FirstName,
		LastName:             req.LastName,
		DateOfBirth:          req.DateOfBirth.Time,
		Gender:               req.Gender,
		PrimaryLanguage:      req.PrimaryLanguage,
		ContactFirstName:     req.ContactFirstName,
		ContactLastName:      req.ContactLastName,
		ContactEmail:         req.ContactEmail,
		ContactPhone:         req.ContactPhone,
		AlternativeFirstName: req.AlternativeFirstName,
		AlternativeLastName:  req.AlternativeLastName,
		AlternativePhone:     req.AlternativePhone,
		AlternativeEmail:     req.AlternativeEmail,
		StreetAddress:        req.StreetAddress,
		Zip:                  req.Zip,
		City:                 req.City,
		Email:                req.Email,
		Phone:                req.Phone,
		TherapyTitle:         req.TherapyTitle,
		ProviderApprovalCode: req.ProviderApprovalCode,
		ProviderApprovalDate: req.ProviderApprovalDate.Time,
		UnitPrice:            req.UnitPrice,
		Status:               req.Status,
		AdmissionDate:        req.AdmissionDate.Time,
		ReferralSource:       req.ReferralSource,
		Notes:                req.Notes,
		Timezone:             req.Timezone,
	}

	// Set default values if not provided
	if client.Gender == "" {
		client.Gender = "undisclosed"
	}
	if client.Status == "" {
		client.Status = "waiting"
	}
	if client.Timezone == "" {
		client.Timezone = "Europe/Berlin"
	}
	if req.InvoicedIndividually != nil {
		client.InvoicedIndividually = *req.InvoicedIndividually
	}
	if req.IsSelfPayer != nil {
		client.IsSelfPayer = *req.IsSelfPayer
	}

	if err := s.db.Create(&client).Error; err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &client, nil
}

// GetClientByID returns a client by ID within a tenant with preloaded cost provider
func (s *ClientService) GetClientByID(id, tenantID uint) (*entities.Client, error) {
	var client entities.Client
	if err := s.db.Preload("CostProvider").Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch client: %w", err)
	}
	return &client, nil
}

// GetAllClients returns all clients with pagination for a tenant with preloaded cost providers
// Optionally filters by status if provided
func (s *ClientService) GetAllClients(page, limit int, tenantID uint, status string) ([]entities.Client, int, error) {
	var clients []entities.Client
	var total int64

	offset := (page - 1) * limit

	// Build query
	query := s.db.Model(&entities.Client{}).Where("tenant_id = ?", tenantID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records for the tenant
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count clients: %w", err)
	}

	// Get paginated records for the tenant with preloaded cost provider
	if err := query.Preload("CostProvider").Offset(offset).Limit(limit).Find(&clients).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch clients: %w", err)
	}

	return clients, int(total), nil
}

// UpdateClient updates an existing client within a tenant
func (s *ClientService) UpdateClient(id, tenantID uint, req entities.UpdateClientRequest) (*entities.Client, error) {
	var client entities.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("client with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Update fields if provided
	// Handle CostProviderID specially to support explicit null values
	if req.CostProviderID != nil {
		if *req.CostProviderID == 0 {
			// Setting to 0 means we want to clear the association
			client.CostProviderID = nil
		} else {
			client.CostProviderID = req.CostProviderID
		}
	}
	if req.FirstName != nil {
		client.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		client.LastName = *req.LastName
	}
	if req.DateOfBirth != nil {
		client.DateOfBirth = req.DateOfBirth.Time
	}
	if req.Gender != nil {
		client.Gender = *req.Gender
	}
	if req.PrimaryLanguage != nil {
		client.PrimaryLanguage = *req.PrimaryLanguage
	}
	if req.ContactFirstName != nil {
		client.ContactFirstName = *req.ContactFirstName
	}
	if req.ContactLastName != nil {
		client.ContactLastName = *req.ContactLastName
	}
	if req.ContactEmail != nil {
		client.ContactEmail = *req.ContactEmail
	}
	if req.ContactPhone != nil {
		client.ContactPhone = *req.ContactPhone
	}
	if req.AlternativeFirstName != nil {
		client.AlternativeFirstName = *req.AlternativeFirstName
	}
	if req.AlternativeLastName != nil {
		client.AlternativeLastName = *req.AlternativeLastName
	}
	if req.AlternativePhone != nil {
		client.AlternativePhone = *req.AlternativePhone
	}
	if req.AlternativeEmail != nil {
		client.AlternativeEmail = *req.AlternativeEmail
	}
	if req.StreetAddress != nil {
		client.StreetAddress = *req.StreetAddress
	}
	if req.Zip != nil {
		client.Zip = *req.Zip
	}
	if req.City != nil {
		client.City = *req.City
	}
	if req.Email != nil {
		client.Email = *req.Email
	}
	if req.Phone != nil {
		client.Phone = *req.Phone
	}
	if req.InvoicedIndividually != nil {
		client.InvoicedIndividually = *req.InvoicedIndividually
	}
	if req.IsSelfPayer != nil {
		client.IsSelfPayer = *req.IsSelfPayer
	}
	if req.TherapyTitle != nil {
		client.TherapyTitle = *req.TherapyTitle
	}
	if req.ProviderApprovalCode != nil {
		client.ProviderApprovalCode = *req.ProviderApprovalCode
	}
	if req.ProviderApprovalDate != nil {
		client.ProviderApprovalDate = req.ProviderApprovalDate.Time
	}
	if req.UnitPrice != nil {
		client.UnitPrice = req.UnitPrice
	}
	if req.Status != nil {
		client.Status = *req.Status
	}
	if req.AdmissionDate != nil {
		client.AdmissionDate = req.AdmissionDate.Time
	}
	if req.ReferralSource != nil {
		client.ReferralSource = *req.ReferralSource
	}
	if req.Notes != nil {
		client.Notes = *req.Notes
	}
	if req.Timezone != nil {
		client.Timezone = *req.Timezone
	}

	if err := s.db.Save(&client).Error; err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	// Reload client to get updated data with cost provider
	if err := s.db.Preload("CostProvider").First(&client, client.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload client: %w", err)
	}

	return &client, nil
}

// DeleteClient soft deletes a client within a tenant
func (s *ClientService) DeleteClient(id, tenantID uint) error {
	var client entities.Client
	if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("client with ID %d not found", id)
		}
		return fmt.Errorf("failed to get client: %w", err)
	}

	if err := s.db.Delete(&client).Error; err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	return nil
}

// SearchClients searches clients by first name or last name within a tenant
func (s *ClientService) SearchClients(query string, page, limit int, tenantID uint) ([]entities.Client, int64, error) {
	var clients []entities.Client
	var total int64

	// Build search query - use LIKE for SQLite compatibility, ILIKE for PostgreSQL
	searchPattern := "%" + query + "%"
	searchQuery := s.db.Model(&entities.Client{}).Where(
		"tenant_id = ? AND (LOWER(first_name) LIKE LOWER(?) OR LOWER(last_name) LIKE LOWER(?))",
		tenantID, searchPattern, searchPattern,
	)

	// Count total matching records
	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count clients: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get paginated search results with cost provider preload
	if err := searchQuery.Preload("CostProvider").Offset(offset).Limit(limit).Find(&clients).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search clients: %w", err)
	}

	return clients, total, nil
}

// GenerateRegistrationToken generates a new registration token for client registration
// and blacklists any existing tokens for the same organization
func (s *ClientService) GenerateRegistrationToken(tenantID, userID, organizationID uint, email string) (*entities.RegistrationToken, error) {
	// Blacklist all existing active tokens for this organization (if any exist)
	// Check if table exists by trying to count records first
	var count int64
	if err := s.db.Model(&entities.RegistrationToken{}).
		Where("organization_id = ? AND blacklisted = ?", organizationID, false).
		Count(&count).Error; err != nil {
		// If table doesn't exist or other error, log but continue (table will be created on first insert)
		fmt.Printf("Warning: Could not check existing tokens (table may not exist yet): %v\n", err)
	} else if count > 0 {
		// Only update if there are records to blacklist
		if err := s.db.Model(&entities.RegistrationToken{}).
			Where("organization_id = ? AND blacklisted = ?", organizationID, false).
			Update("blacklisted", true).Error; err != nil {
			return nil, fmt.Errorf("failed to blacklist old tokens: %w", err)
		}
	}

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	// Create registration token (no expiration)
	regToken := &entities.RegistrationToken{
		OrganizationID: organizationID,
		TenantID:       tenantID,
		Token:          tokenString,
		Email:          email,
		CreatedBy:      userID,
	}

	if err := s.db.Create(regToken).Error; err != nil {
		return nil, fmt.Errorf("failed to create registration token: %w", err)
	}

	return regToken, nil
}

// ValidateRegistrationToken validates a registration token and returns the associated data
func (s *ClientService) ValidateRegistrationToken(tokenString string) (*entities.RegistrationToken, error) {
	var regToken entities.RegistrationToken
	if err := s.db.Where("token = ?", tokenString).First(&regToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid registration token")
		}
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	// Check if token is blacklisted
	if regToken.Blacklisted {
		return nil, errors.New("registration token has been revoked")
	}

	return &regToken, nil
}

// MarkRegistrationTokenAsUsed increments the usage count for a registration token
func (s *ClientService) MarkRegistrationTokenAsUsed(tokenID uint) error {
	return s.db.Model(&entities.RegistrationToken{}).
		Where("id = ?", tokenID).
		Update("used_count", gorm.Expr("used_count + ?", 1)).Error
}

// RegisterClientViaToken registers a new client using a valid registration token
func (s *ClientService) RegisterClientViaToken(tokenString string, req entities.ClientRegistrationRequest) (*entities.Client, error) {
	// Validate token
	regToken, err := s.ValidateRegistrationToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if email from token matches email in request (if token has email)
	if regToken.Email != "" && regToken.Email != req.Email {
		return nil, errors.New("email does not match registration token")
	}

	// Check if client with this email already exists
	var existingClient entities.Client
	if err := s.db.Where("email = ? AND tenant_id = ?", req.Email, regToken.TenantID).First(&existingClient).Error; err == nil {
		return nil, errors.New("client with this email already exists")
	}

	// Create client in waiting status
	client := entities.Client{
		TenantID:         regToken.TenantID,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		EmailVerified:    false, // Will be verified via separate endpoint
		Phone:            req.Phone,
		DateOfBirth:      req.DateOfBirth.Time,
		Gender:           req.Gender,
		StreetAddress:    req.StreetAddress,
		Zip:              req.Zip,
		City:             req.City,
		ContactFirstName: req.ContactFirstName,
		ContactLastName:  req.ContactLastName,
		ContactEmail:     req.ContactEmail,
		ContactPhone:     req.ContactPhone,
		Notes:            req.Notes,
		Timezone:         req.Timezone,
		Status:           "waiting", // Waiting list status
	}

	// Set default timezone if not provided
	if client.Timezone == "" {
		client.Timezone = "Europe/Berlin"
	}

	// Create client in database
	if err := s.db.Create(&client).Error; err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Mark token as used
	if err := s.MarkRegistrationTokenAsUsed(regToken.ID); err != nil {
		// Log warning but don't fail the registration
		fmt.Printf("Warning: failed to mark registration token as used: %v\n", err)
	}

	// Load cost provider if exists
	if client.CostProviderID != nil {
		s.db.Preload("CostProvider").First(&client, client.ID)
	}

	return &client, nil
}

// VerifyClientEmail verifies a client's email using a verification token
func (s *ClientService) VerifyClientEmail(verificationToken string) (*entities.Client, error) {
	// Validate verification token using base auth service
	userID, email, err := baseAuth.ValidateVerificationToken(verificationToken)
	if err != nil {
		return nil, fmt.Errorf("invalid verification token: %w", err)
	}

	// Find client by email (verification tokens use email as subject)
	var client entities.Client
	if err := s.db.Where("email = ?", email).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, fmt.Errorf("failed to find client: %w", err)
	}

	// Verify the userID matches client ID (extra security check)
	if uint(userID) != client.ID {
		return nil, errors.New("token does not match client")
	}

	// Mark email as verified
	client.EmailVerified = true
	if err := s.db.Save(&client).Error; err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	return &client, nil
}

// GenerateEmailVerificationToken generates a verification token for a client's email
func (s *ClientService) GenerateEmailVerificationToken(clientID uint) (string, error) {
	// Get client
	var client entities.Client
	if err := s.db.First(&client, clientID).Error; err != nil {
		return "", fmt.Errorf("client not found: %w", err)
	}

	if client.Email == "" {
		return "", errors.New("client does not have an email address")
	}

	// Generate verification token using base auth service
	token, err := baseAuth.GenerateVerificationToken(client.Email, clientID)
	if err != nil {
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}

	return token, nil
}

// SendVerificationEmail sends an email verification email to a client using template system
func (s *ClientService) SendVerificationEmail(clientID uint, verificationToken string) error {
	fmt.Printf("🔍 SendVerificationEmail called for client ID: %d\n", clientID)

	// Get client
	var client entities.Client
	if err := s.db.First(&client, clientID).Error; err != nil {
		return fmt.Errorf("client not found: %w", err)
	}
	fmt.Printf("🔍 Client found: %s %s (%s)\n", client.FirstName, client.LastName, client.Email)

	if client.Email == "" {
		return errors.New("client does not have an email address")
	}

	// Check if email service is available
	if s.emailService == nil {
		fmt.Printf("❌ Email service is nil!\n")
		return errors.New("email service not available")
	}
	fmt.Printf("✅ Email service is available\n")

	// Check if template service is available
	if s.templateService == nil {
		fmt.Printf("❌ Template service is nil!\n")
		return errors.New("template service not available")
	}
	fmt.Printf("✅ Template service is available\n")

	// Type assert services
	type EmailServiceInterface interface {
		SendEmail(to, subject, htmlBody, textBody string) error
	}

	type TemplateServiceInterface interface {
		RenderTemplate(ctx context.Context, tenantID uint, templateID uint, data map[string]interface{}) (string, error)
	}

	emailSvc, ok := s.emailService.(EmailServiceInterface)
	if !ok {
		fmt.Printf("❌ Email service does not implement SendEmail interface\n")
		return errors.New("email service does not implement required interface")
	}

	templateSvc, ok := s.templateService.(TemplateServiceInterface)
	if !ok {
		fmt.Printf("❌ Template service does not implement RenderTemplate interface\n")
		return errors.New("template service does not implement required interface")
	}
	fmt.Printf("✅ Both services interface OK\n")

	// Build client-specific verification URL (frontend URL)
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // Fallback to default
	}
	verificationURL := fmt.Sprintf("%s/client-email-verification/%s", frontendURL, verificationToken)
	fmt.Printf("🔗 Verification URL: %s\n", verificationURL)

	// Get recipient name
	recipientName := client.FirstName
	if recipientName == "" {
		recipientName = client.LastName
	}
	if recipientName == "" {
		recipientName = "Client"
	}

	// Get client's tenant ID
	tenantID := uint(1)
	if client.TenantID > 0 {
		tenantID = client.TenantID
	}

	// Find the ttemplateEmplate
	var template templateEntities.Template
	err := s.db.Where("tenant_id = ? AND module = ? AND template_key = ? AND channel = ? AND is_active = ?",
		tenantID, "client_management", "client_email_verification", "EMAIL", true).
		Order("is_default DESC, created_at DESC").
		First(&template).Error

	if err != nil {
		fmt.Printf("❌ Template not found: %v\n", err)
		return fmt.Errorf("client email verification template not found: %w", err)
	}
	fmt.Printf("✅ Found template ID: %d (Name: %s)\n", template.ID, template.Name)

	// Prepare template data
	templateData := map[string]interface{}{
		"FirstName":       client.FirstName,
		"LastName":        client.LastName,
		"Email":           client.Email,
		"VerificationURL": verificationURL,
		"PortalName":      "Client Portal",
	}

	fmt.Printf("📧 Rendering template for: %s (tenant: %d)\n", client.Email, tenantID)

	// Render the template
	htmlBody, err := templateSvc.RenderTemplate(context.Background(), tenantID, template.ID, templateData)
	if err != nil {
		fmt.Printf("❌ Template rendering failed: %v\n", err)
		return fmt.Errorf("failed to render email template: %w", err)
	}
	fmt.Printf("✅ Template rendered successfully (%d bytes)\n", len(htmlBody))

	// Create plain text version
	textBody := fmt.Sprintf(`Welcome to the Client Portal

Hello %s,

Thank you for registering with our client portal. To complete your registration and access your account, please verify your email address.

Verification URL:
%s

Note: This verification link will expire in 24 hours.

If you didn't register for a client account, please ignore this email.

---
This is an automated message, please do not reply to this email.`, recipientName, verificationURL)

	// Get subject from template or use default
	subject := "Verify Your Email - Client Portal Access"
	if template.Subject != nil && *template.Subject != "" {
		subject = *template.Subject
	}

	// Send email
	fmt.Printf("📨 Sending email to: %s with subject: %s\n", client.Email, subject)
	err = emailSvc.SendEmail(client.Email, subject, htmlBody, textBody)
	if err != nil {
		fmt.Printf("❌ SendEmail failed: %v\n", err)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	fmt.Printf("✉️  Client verification email sent to %s (Client ID: %d)\n", client.Email, client.ID)
	return nil
}
