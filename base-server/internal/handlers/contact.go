package handlers

import (
	"net/http"
	"time"

	"github.com/ae-saas-basic/ae-saas-basic/internal/models"
	"github.com/ae-saas-basic/ae-saas-basic/internal/services"
	"github.com/ae-saas-basic/ae-saas-basic/pkg/utils"
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
// @Summary Get all contacts
// @Description Get a paginated list of all contacts
// @Tags contacts
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param active query bool false "Filter by active status"
// @Param type query string false "Filter by contact type"
// @Success 200 {object} models.APIResponse{data=models.ListResponse}
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts [get]
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
// @Summary Get contact by ID
// @Description Get a specific contact by its ID
// @Tags contacts
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contact ID"
// @Success 200 {object} models.APIResponse{data=models.ContactResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /contacts/{id} [get]
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
// @Summary Create a new contact
// @Description Create a new contact
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ContactCreateRequest true "Contact creation data"
// @Success 201 {object} models.APIResponse{data=models.ContactResponse}
// @Failure 400 {object} models.ErrorResponse
// @Router /contacts [post]
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
// @Summary Update a contact
// @Description Update an existing contact by ID
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contact ID"
// @Param request body models.ContactUpdateRequest true "Contact update data"
// @Success 200 {object} models.APIResponse{data=models.ContactResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /contacts/{id} [put]
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
// @Summary Delete a contact
// @Description Soft delete a contact by ID
// @Tags contacts
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contact ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /contacts/{id} [delete]
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
// @Summary Submit contact form
// @Description Submit a contact form and optionally subscribe to newsletter
// @Tags contact
// @Accept json
// @Produce json
// @Param contactForm body models.ContactFormRequest true "Contact form data"
// @Success 200 {object} models.ContactFormResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /contact/form [post]
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
// @Summary Get newsletter subscriptions
// @Description Get all newsletter subscriptions for admin users
// @Tags contact
// @Accept json
// @Produce json
// @Success 200 {array} models.Newsletter
// @Failure 500 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /contact/newsletter [get]
func (h *ContactHandler) GetNewsletterSubscriptions(c *gin.Context) {
	var newsletters []models.Newsletter

	if err := h.db.Find(&newsletters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch newsletter subscriptions"})
		return
	}

	c.JSON(http.StatusOK, newsletters)
}

// UnsubscribeFromNewsletter handles newsletter unsubscription
// @Summary Unsubscribe from newsletter
// @Description Unsubscribe an email from the newsletter
// @Tags contact
// @Accept json
// @Produce json
// @Param email query string true "Email to unsubscribe"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /contact/newsletter/unsubscribe [delete]
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
