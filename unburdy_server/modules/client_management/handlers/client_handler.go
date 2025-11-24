package handlers

import (
	"net/http"
	"strconv"

	baseAPI "github.com/ae-base-server/api"
	"github.com/ae-base-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/unburdy/unburdy-server-api/internal/models"
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
// @ID createClient
// @Accept json
// @Produce json
// @Param client body entities.CreateClientRequest true "Client information"
// @Success 201 {object} models.APIResponse{data=entities.ClientResponse} "Client created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients [post]
func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req entities.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	client, err := h.clientService.CreateClient(req, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Client created successfully", client.ToResponse()))
}

// GetClient handles retrieving a client by ID
// @Summary Get a client by ID
// @Description Retrieve a specific client by their ID
// @Tags clients
// @ID getClientById
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} models.APIResponse{data=entities.ClientResponse} "Client found"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients/{id} [get]
func (h *ClientHandler) GetClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid client ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	client, err := h.clientService.GetClientByID(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Client retrieved successfully", client.ToResponse()))
}

// GetAllClients handles retrieving all clients with pagination
// @Summary Get all clients
// @Description Retrieve all clients with optional pagination
// @Tags clients
// @ID getClients
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of clients per page (respects DEFAULT_PAGE_LIMIT and MAX_PAGE_LIMIT env vars)" default(200)
// @Success 200 {object} models.APIResponse{data=models.ListResponse} "Clients retrieved successfully"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /clients [get]
func (h *ClientHandler) GetAllClients(c *gin.Context) {
	page, limit := utils.GetPaginationParams(c)

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	clients, total, err := h.clientService.GetAllClients(page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Failed to retrieve clients", err.Error()))
		return
	}

	responses := make([]entities.ClientResponse, len(clients))
	for i, client := range clients {
		responses[i] = client.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessListResponse(responses, page, limit, int(total)))
}

// UpdateClient handles updating a client
// @Summary Update a client
// @Description Update a client's information
// @Tags clients
// @ID updateClient
// @Accept json
// @Produce json
// @Param id path int true "Client ID"
// @Param client body entities.UpdateClientRequest true "Updated client information"
// @Success 200 {object} models.APIResponse{data=entities.ClientResponse} "Updated client"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /clients/{id} [put]
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid client ID"))
		return
	}

	var req entities.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", err.Error()))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	client, err := h.clientService.UpdateClient(uint(id), tenantID, req)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Client updated successfully", client.ToResponse()))
}

// DeleteClient handles deleting a client
// @Summary Delete a client
// @Description Soft delete a client by ID
// @Tags clients
// @ID deleteClient
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} models.APIResponse "Client deleted successfully"
// @Failure 400 {object} models.APIResponse "Bad request"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 404 {object} models.APIResponse "Client not found"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Invalid client ID"))
		return
	}

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	err = h.clientService.DeleteClient(uint(id), tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessMessageResponse("Client deleted successfully"))
}

// SearchClients handles searching clients
// @Summary Search clients
// @Description Search clients by first name or last name
// @Tags clients
// @ID searchClients
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of clients per page (respects DEFAULT_PAGE_LIMIT and MAX_PAGE_LIMIT env vars)" default(200)
// @Success 200 {object} models.APIResponse{data=models.ListResponse} "Search results"
// @Failure 400 {object} models.APIResponse "Bad request"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /clients/search [get]
func (h *ClientHandler) SearchClients(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponseFunc("Invalid request", "Search query is required"))
		return
	}

	page, limit := utils.GetPaginationParams(c)

	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	clients, total, err := h.clientService.SearchClients(query, page, limit, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	responses := make([]entities.ClientResponse, len(clients))
	for i, client := range clients {
		responses[i] = client.ToResponse()
	}

	c.JSON(http.StatusOK, models.SuccessListResponse(responses, page, limit, int(total)))
}

// GetClientByToken handles retrieving client information from a token
// @Summary Get client by token
// @Description Retrieve client details from any valid JWT token containing client_id
// @Tags clients
// @ID getClientByToken
// @Produce json
// @Param token path string true "JWT token containing client_id"
// @Success 200 {object} models.APIResponse{data=entities.ClientResponse} "Client found"
// @Failure 401 {object} map[string]string "Unauthorized or invalid token"
// @Failure 404 {object} map[string]string "Client not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /client/{token} [get]
func (h *ClientHandler) GetClientByToken(c *gin.Context) {
	// Try to get client_id from context (set by any middleware that validates tokens)
	clientIDInterface, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "No client_id found in token"))
		return
	}

	clientID, ok := clientIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Invalid client_id format"))
		return
	}

	// Get tenant_id from context
	tenantID, err := baseAPI.GetTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseFunc("Unauthorized", "Failed to get tenant ID: "+err.Error()))
		return
	}

	// Retrieve client
	client, err := h.clientService.GetClientByID(clientID, tenantID)
	if err != nil {
		if err.Error() == "client not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Not found", "Client not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponseFunc("Internal server error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Client information retrieved successfully", client.ToResponse()))
}
