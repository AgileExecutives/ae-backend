package handlers

import (
	"net/http"
	"strconv"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/internal/models"
	"github.com/unburdy/unburdy-server-api/internal/services"
)

type ClientHandler struct {
	clientService *services.ClientService
}

// NewClientHandler creates a new client handler
func NewClientHandler(clientService *services.ClientService) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
	}
}

// CreateClient godoc
// @Summary Create a new client
// @Description Create a new client with first name, last name, and date of birth
// @Tags clients
// @Accept json
// @Produce json
// @Param client body models.CreateClientRequest true "Client data"
// @Success 201 {object} models.ClientResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /clients [post]
func (h *ClientHandler) CreateClient(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	var req models.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := h.clientService.CreateClient(req, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, client.ToResponse())
}

// GetClient godoc
// @Summary Get a client by ID
// @Description Retrieve a specific client by their ID
// @Tags clients
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} models.ClientResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /clients/{id} [get]
func (h *ClientHandler) GetClient(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	client, err := h.clientService.GetClientByID(uint(id), tenantID)
	if err != nil {
		if err.Error() == "client not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client.ToResponse())
}

// GetAllClients godoc
// @Summary Get all clients
// @Description Get a paginated list of all clients
// @Tags clients
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "data: array of ClientResponse, total: int, page: int, limit: int"
// @Failure 500 {object} map[string]interface{} "error: string"
// @Security BearerAuth
// @Router /clients [get]
func (h *ClientHandler) GetAllClients(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	page, limit := utils.GetPaginationParams(c)

	clients, total, err := h.clientService.GetAllClients(page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	clientResponses := make([]models.ClientResponse, len(clients))
	for i, client := range clients {
		clientResponses[i] = client.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": clientResponses,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// UpdateClient godoc
// @Summary Update a client
// @Description Update an existing client's information
// @Tags clients
// @Accept json
// @Produce json
// @Param id path int true "Client ID"
// @Param client body models.UpdateClientRequest true "Client update data"
// @Success 200 {object} models.ClientResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /clients/{id} [put]
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req models.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := h.clientService.UpdateClient(uint(id), tenantID, req)
	if err != nil {
		if err.Error() == "client not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client.ToResponse())
}

// DeleteClient godoc
// @Summary Delete a client
// @Description Soft delete a client by ID
// @Tags clients
// @Produce json
// @Param id path int true "Client ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	err = h.clientService.DeleteClient(uint(id), tenantID)
	if err != nil {
		if err.Error() == "client not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchClients godoc
// @Summary Search clients
// @Description Search clients by first name or last name
// @Tags clients
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /clients/search [get]
func (h *ClientHandler) SearchClients(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Get authenticated tenant info
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant information required"})
		return
	}

	page, limit := utils.GetPaginationParams(c)

	clients, total, err := h.clientService.SearchClients(query, page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	clientResponses := make([]models.ClientResponse, len(clients))
	for i, client := range clients {
		clientResponses[i] = client.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": clientResponses,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"query":   query,
	})
}
