package handlers

import (
	"net/http"
	"time"

	"github.com/ae-base-server/internal/models"
	_ "github.com/ae-base-server/modules/base/models"
	emailServices "github.com/ae-base-server/modules/email/services"
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContactHandlers struct {
	db           *gorm.DB
	logger       core.Logger
	emailService *emailServices.EmailService
}

func NewContactHandlers(db *gorm.DB, logger core.Logger) *ContactHandlers {
	return &ContactHandlers{
		db:           db,
		logger:       logger,
		emailService: emailServices.NewEmailService(),
	}
}

// @Summary Get all contacts
// @ID getContacts
// @Description Get paginated list of contacts with optional filters
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param active query boolean false "Filter by active status"
// @Param type query string false "Filter by contact type (business, personal, etc)"
// @Success 200 {object} models.APIResponse{data=models.ListResponse}
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts [get]
func (h *ContactHandlers) GetContacts(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)
	offset := utils.GetOffset(page, limit)
	var contacts []models.Contact
	var total int64
	query := h.db.Model(&models.Contact{})
	if activeStr := c.Query("active"); activeStr != "" {
		if activeStr == "true" {
			query = query.Where("active = ?", true)
		} else if activeStr == "false" {
			query = query.Where("active = ?", false)
		}
	}
	if contactType := c.Query("type"); contactType != "" {
		query = query.Where("type = ?", contactType)
	}
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to count contacts", err.Error()))
		return
	}
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve contacts", err.Error()))
		return
	}
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

// @Summary Get contact by ID
// @ID getContactById
// @Description Get a specific contact by ID
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contact ID (UUID)"
// @Success 200 {object} models.APIResponse{data=models.Contact}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts/{id} [get]
func (h *ContactHandlers) GetContact(c *gin.Context) {
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

// @Summary Create new contact
// @ID createContact
// @Description Create a new contact for the authenticated user
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contact body models.ContactCreateRequest true "Contact data"
// @Success 201 {object} models.APIResponse{data=models.ContactResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts [post]
func (h *ContactHandlers) CreateContact(c *gin.Context) {
	var req models.ContactCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}
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

// @Summary Update contact
// @ID updateContact
// @Description Update an existing contact by ID
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contact ID (UUID)"
// @Param contact body models.ContactUpdateRequest true "Updated contact data"
// @Success 200 {object} models.APIResponse{data=models.ContactResponse}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts/{id} [put]
func (h *ContactHandlers) UpdateContact(c *gin.Context) {
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

// @Summary Delete contact
// @ID deleteContact
// @Description Soft delete a contact by ID
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contact ID (UUID)"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts/{id} [delete]
func (h *ContactHandlers) DeleteContact(c *gin.Context) {
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

// @Summary Submit contact form
// @ID submitContactForm
// @Description Public endpoint to submit a contact form (no authentication required)
// @Tags contacts
// @Accept json
// @Produce json
// @Param form body models.ContactFormRequest true "Contact form data"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts/form [post]
func (h *ContactHandlers) SubmitContactForm(c *gin.Context) {
	var req models.ContactFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Timestamp == "" {
		req.Timestamp = time.Now().Format(time.RFC3339)
	}
	err := h.emailService.SendContactFormEmail(req.Name, req.Email, req.Subject, req.Message, req.Timestamp, req.Source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send contact form email: " + err.Error()})
		return
	}
	response := models.ContactFormResponse{Message: "Contact form submitted successfully"}
	if req.Newsletter {
		newsletter := models.Newsletter{Name: req.Name, Email: req.Email, Interest: req.Subject, Source: req.Source, LastContact: time.Now()}
		var count int64
		h.db.Model(&models.Newsletter{}).Where("email = ?", req.Email).Count(&count)
		if count == 0 {
			if err := h.db.Create(&newsletter).Error; err != nil {
				response.NewsletterAdded = false
				response.NewsletterMessage = "Contact form sent, but newsletter subscription failed"
			} else {
				response.NewsletterAdded = true
				response.NewsletterMessage = "Successfully subscribed to newsletter"
			}
		} else {
			var existingNewsletter models.Newsletter
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
				response.NewsletterAdded = false
				response.NewsletterMessage = "Contact form sent, but newsletter subscription failed"
			}
		}
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Get newsletter subscriptions
// @ID getNewsletterSubscriptions
// @Description Get all newsletter subscriptions
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Newsletter
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts/newsletter [get]
func (h *ContactHandlers) GetNewsletterSubscriptions(c *gin.Context) {
	var newsletters []models.Newsletter
	if err := h.db.Find(&newsletters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch newsletter subscriptions"})
		return
	}
	c.JSON(http.StatusOK, newsletters)
}

// @Summary Unsubscribe from newsletter
// @ID unsubscribeFromNewsletter
// @Description Unsubscribe an email address from the newsletter
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param email path string true "Email address to unsubscribe"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /contacts/newsletter/{email} [delete]
func (h *ContactHandlers) UnsubscribeFromNewsletter(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}
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
