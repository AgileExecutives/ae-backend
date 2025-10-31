package handlers

import (
	"net/http"
	"strconv"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"

	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// ClientHandler handles client-related HTTP requests
type ClientHandler struct {
	clientService *services.ClientService
}

// NewClientHandler creates a new client handler
func NewClientHandler(clientService *services.ClientService) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
	}
}

// CreateClient handles creating a new client
// @Summary Create a new client
// @Description Create a new client with the provided information
// @Tags clients
// @Accept json
// @Produce json
// @Param client body entities.CreateClientRequest true "Client information"
// @Success 201 {object} entities.ClientResponse "Created client"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients [post]
func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req entities.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get tenant ID: " + err.Error()})
		return
	}

	client, err := h.clientService.CreateClient(req, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, client.ToResponse())
}

// GetClient handles retrieving a client by ID
// @Summary Get a client by ID
// @Description Retrieve a specific client by their ID
// @Tags clients
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} entities.ClientResponse "Client found"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients/{id} [get]
func (h *ClientHandler) GetClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get tenant ID: " + err.Error()})
		return
	}

	client, err := h.clientService.GetClientByID(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client.ToResponse())
}

// GetAllClients handles retrieving all clients with pagination
// @Summary Get all clients
// @Description Retrieve all clients with optional pagination
// @Tags clients
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of clients per page (respects DEFAULT_PAGE_LIMIT and MAX_PAGE_LIMIT env vars)" default(200)
// @Success 200 {object} map[string]interface{} "Clients retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients [get]
func (h *ClientHandler) GetAllClients(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get tenant ID: " + err.Error()})
		return
	}

	clients, total, err := h.clientService.GetAllClients(page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]entities.ClientResponse, len(clients))
	for i, client := range clients {
		responses[i] = client.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": responses,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + limit - 1) / limit,
		},
	})
}

// UpdateClient handles updating a client
// @Summary Update a client
// @Description Update a client's information
// @Tags clients
// @Accept json
// @Produce json
// @Param id path int true "Client ID"
// @Param client body entities.UpdateClientRequest true "Updated client information"
// @Success 200 {object} entities.ClientResponse "Updated client"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients/{id} [put]
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req entities.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get tenant ID: " + err.Error()})
		return
	}

	client, err := h.clientService.UpdateClient(uint(id), tenantID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, client.ToResponse())
}

// DeleteClient handles deleting a client
// @Summary Delete a client
// @Description Soft delete a client by ID
// @Tags clients
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} map[string]string "Client deleted successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get tenant ID: " + err.Error()})
		return
	}

	err = h.clientService.DeleteClient(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Client deleted successfully"})
}

// SearchClients handles searching clients
// @Summary Search clients
// @Description Search clients by first name or last name
// @Tags clients
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of clients per page (respects DEFAULT_PAGE_LIMIT and MAX_PAGE_LIMIT env vars)" default(200)
// @Success 200 {object} map[string]interface{} "Search results"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients/search [get]
func (h *ClientHandler) SearchClients(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	page, limit := utils.GetPaginationParams(c)

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get tenant ID: " + err.Error()})
		return
	}

	clients, total, err := h.clientService.SearchClients(query, page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]entities.ClientResponse, len(clients))
	for i, client := range clients {
		responses[i] = client.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": responses,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (int(total) + limit - 1) / limit,
		},
		"query": query,
	})
}
