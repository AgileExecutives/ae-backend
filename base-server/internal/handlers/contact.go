package handlers

import (
	"net/http"
	"time"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/internal/services"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContactHandler struct {
	db           *gorm.DB
	emailService *services.EmailService
}

// NewContactHandler creates a new contact handler
func NewContactHandler(db *gorm.DB) *ContactHandler {
	return &ContactHandler{
		db:           db,
		emailService: services.NewEmailService(),
	}
}

// GetContacts retrieves all contacts with pagination
// DISABLED-SWAGGER: @Summary Get all contacts
// DISABLED-SWAGGER: @Description Get a paginated list of all contacts
// DISABLED-SWAGGER: @Tags contacts
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param page query int false "Page number" default(1)
// DISABLED-SWAGGER: @Param limit query int false "Items per page" default(10)
// DISABLED-SWAGGER: @Param active query bool false "Filter by active status"
// DISABLED-SWAGGER: @Param type query string false "Filter by contact type"
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=models.ListResponse}
// DISABLED-SWAGGER: @Failure 500 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /contacts [get]
func (h *ContactHandler) GetContacts(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)
	offset := utils.GetOffset(page, limit)

	var contacts []models.Contact
	var total int64

	query := h.db.Model(&models.Contact{})

	// Filter by active status if provided
	if activeStr := c.Query("active"); activeStr != "" {
		if activeStr == "true" {
			query = query.Where("active = ?", true)
		} else if activeStr == "false" {
			query = query.Where("active = ?", false)
		}
	}

	// Filter by type if provided
	if contactType := c.Query("type"); contactType != "" {
		query = query.Where("type = ?", contactType)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to count contacts", err.Error()))
		return
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve contacts", err.Error()))
		return
	}

	// Convert to response format
	var responses []models.ContactResponse
	for _, contact := range contacts {
		responses = append(responses, contact.ToResponse())
	}

	response := models.ListResponse{
		Data: responses,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: utils.CalculateTotalPages(int(total), limit),
		},
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Contacts retrieved successfully", response))
}

// GetContact retrieves a specific contact by ID
// DISABLED-SWAGGER: @Summary Get contact by ID
// DISABLED-SWAGGER: @Description Get a specific contact by its ID
// DISABLED-SWAGGER: @Tags contacts
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param id path int true "Contact ID"
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=models.ContactResponse}
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 404 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /contacts/{id} [get]
func (h *ContactHandler) GetContact(c *gin.Context) {
	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid contact ID", err.Error()))
		return
	}

	var contact models.Contact
	if err := h.db.First(&contact, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Contact not found", "Contact with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Contact retrieved successfully", contact.ToResponse()))
}

// CreateContact creates a new contact
// DISABLED-SWAGGER: @Summary Create a new contact
// DISABLED-SWAGGER: @Description Create a new contact
// DISABLED-SWAGGER: @Tags contacts
// DISABLED-SWAGGER: @Accept json
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param request body models.ContactCreateRequest true "Contact creation data"
// DISABLED-SWAGGER: @Success 201 {object} models.APIResponse{data=models.ContactResponse}
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req models.ContactCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	// Set default type if not provided
	contactType := req.Type
	if contactType == "" {
		contactType = "contact"
	}

	contact := models.Contact{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Mobile:    req.Mobile,
		Street:    req.Street,
		Zip:       req.Zip,
		City:      req.City,
		Country:   req.Country,
		Type:      contactType,
		Notes:     req.Notes,
		Active:    true,
	}

	if err := h.db.Create(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to create contact", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Contact created successfully", contact.ToResponse()))
}

// UpdateContact updates an existing contact
// DISABLED-SWAGGER: @Summary Update a contact
// DISABLED-SWAGGER: @Description Update an existing contact by ID
// DISABLED-SWAGGER: @Tags contacts
// DISABLED-SWAGGER: @Accept json
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param id path int true "Contact ID"
// DISABLED-SWAGGER: @Param request body models.ContactUpdateRequest true "Contact update data"
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse{data=models.ContactResponse}
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 404 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /contacts/{id} [put]
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid contact ID", err.Error()))
		return
	}

	var req models.ContactUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	var contact models.Contact
	if err := h.db.First(&contact, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Contact not found", "Contact with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve contact", err.Error()))
		return
	}

	// Update fields if provided
	if req.FirstName != "" {
		contact.FirstName = req.FirstName
	}
	if req.LastName != "" {
		contact.LastName = req.LastName
	}
	if req.Email != "" {
		contact.Email = req.Email
	}
	if req.Phone != "" {
		contact.Phone = req.Phone
	}
	if req.Mobile != "" {
		contact.Mobile = req.Mobile
	}
	if req.Street != "" {
		contact.Street = req.Street
	}
	if req.Zip != "" {
		contact.Zip = req.Zip
	}
	if req.City != "" {
		contact.City = req.City
	}
	if req.Country != "" {
		contact.Country = req.Country
	}
	if req.Type != "" {
		contact.Type = req.Type
	}
	if req.Notes != "" {
		contact.Notes = req.Notes
	}
	if req.Active != nil {
		contact.Active = *req.Active
	}

	if err := h.db.Save(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to update contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Contact updated successfully", contact.ToResponse()))
}

// DeleteContact deletes a contact (soft delete)
// DISABLED-SWAGGER: @Summary Delete a contact
// DISABLED-SWAGGER: @Description Soft delete a contact by ID
// DISABLED-SWAGGER: @Tags contacts
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Param id path int true "Contact ID"
// DISABLED-SWAGGER: @Success 200 {object} models.APIResponse
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 404 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	id, err := utils.ValidateID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid contact ID", err.Error()))
		return
	}

	var contact models.Contact
	if err := h.db.First(&contact, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Contact not found", "Contact with specified ID does not exist"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve contact", err.Error()))
		return
	}

	if err := h.db.Delete(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to delete contact", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Contact deleted successfully", nil))
}

// SubmitContactForm handles contact form submissions
// DISABLED-SWAGGER: @Summary Submit contact form
// DISABLED-SWAGGER: @Description Submit a contact form and optionally subscribe to newsletter
// DISABLED-SWAGGER: @Tags contact
// DISABLED-SWAGGER: @Accept json
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Param contactForm body models.ContactFormRequest true "Contact form data"
// DISABLED-SWAGGER: @Success 200 {object} models.ContactFormResponse
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 500 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /contact/form [post]
func (h *ContactHandler) SubmitContactForm(c *gin.Context) {
	var req models.ContactFormRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamp if not provided
	if req.Timestamp == "" {
		req.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Send email to support
	err := h.emailService.SendContactFormEmail(
		req.Name,
		req.Email,
		req.Subject,
		req.Message,
		req.Timestamp,
		req.Source,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send contact form email: " + err.Error()})
		return
	}

	response := models.ContactFormResponse{
		Message: "Contact form submitted successfully",
	}

	// Handle newsletter subscription if requested
	if req.Newsletter {
		newsletter := models.Newsletter{
			Name:        req.Name,
			Email:       req.Email,
			Interest:    req.Subject, // Use subject as interest
			Source:      req.Source,
			LastContact: time.Now(),
		}

		// Check if email already exists in newsletter
		var existingNewsletter models.Newsletter
		var count int64
		h.db.Model(&models.Newsletter{}).Where("email = ?", req.Email).Count(&count)

		if count == 0 {
			// Create new newsletter subscription
			if err := h.db.Create(&newsletter).Error; err != nil {
				// Don't fail the whole request if newsletter signup fails
				response.NewsletterAdded = false
				response.NewsletterMessage = "Contact form sent, but newsletter subscription failed"
			} else {
				response.NewsletterAdded = true
				response.NewsletterMessage = "Successfully subscribed to newsletter"
			}
		} else {
			// Update existing subscription
			result := h.db.Where("email = ?", req.Email).First(&existingNewsletter)
			if result.Error == nil {
				existingNewsletter.Name = req.Name
				existingNewsletter.Interest = req.Subject
				existingNewsletter.Source = req.Source
				existingNewsletter.LastContact = time.Now()

				if err := h.db.Save(&existingNewsletter).Error; err != nil {
					response.NewsletterAdded = false
					response.NewsletterMessage = "Contact form sent, but newsletter update failed"
				} else {
					response.NewsletterAdded = true
					response.NewsletterMessage = "Newsletter subscription updated"
				}
			} else {
				// Database error
				response.NewsletterAdded = false
				response.NewsletterMessage = "Contact form sent, but newsletter subscription failed"
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetNewsletterSubscriptions gets all newsletter subscriptions (admin only)
// DISABLED-SWAGGER: @Summary Get newsletter subscriptions
// DISABLED-SWAGGER: @Description Get all newsletter subscriptions for admin users
// DISABLED-SWAGGER: @Tags contact
// DISABLED-SWAGGER: @Accept json
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Success 200 {array} models.Newsletter
// DISABLED-SWAGGER: @Failure 500 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Security BearerAuth
// DISABLED-SWAGGER: @Router /contact/newsletter [get]
func (h *ContactHandler) GetNewsletterSubscriptions(c *gin.Context) {
	var newsletters []models.Newsletter

	if err := h.db.Find(&newsletters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch newsletter subscriptions"})
		return
	}

	c.JSON(http.StatusOK, newsletters)
}

// UnsubscribeFromNewsletter handles newsletter unsubscription
// DISABLED-SWAGGER: @Summary Unsubscribe from newsletter
// DISABLED-SWAGGER: @Description Unsubscribe an email from the newsletter
// DISABLED-SWAGGER: @Tags contact
// DISABLED-SWAGGER: @Accept json
// DISABLED-SWAGGER: @Produce json
// DISABLED-SWAGGER: @Param email query string true "Email to unsubscribe"
// DISABLED-SWAGGER: @Success 200 {object} map[string]string
// DISABLED-SWAGGER: @Failure 400 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 404 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Failure 500 {object} models.ErrorResponse
// DISABLED-SWAGGER: @Router /contact/newsletter/unsubscribe [delete]
func (h *ContactHandler) UnsubscribeFromNewsletter(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	// Soft delete the newsletter subscription
	result := h.db.Where("email = ?", email).Delete(&models.Newsletter{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe from newsletter"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found in newsletter subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unsubscribed from newsletter"})
}
