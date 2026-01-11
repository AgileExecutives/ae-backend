package handlers

import (
	"net/http"
	"strconv"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/ae-base-server/modules/templates/services"
	"github.com/gin-gonic/gin"
)

// ContractHandler handles template contract HTTP requests
type ContractHandler struct {
	service *services.ContractService
}

// NewContractHandler creates a new contract handler
func NewContractHandler(service *services.ContractService) *ContractHandler {
	return &ContractHandler{
		service: service,
	}
}

// RegisterContract registers or updates a template contract
// @Summary Register template contract
// @Description Register a new template contract or update existing one
// @Tags Template Contracts
// @ID registerContract
// @Accept json
// @Produce json
// @Param request body entities.RegisterContractRequest true "Contract data"
// @Success 201 {object} entities.ContractResponse
// @Success 200 {object} entities.ContractResponse "Contract updated"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/contracts [post]
// @Security BearerAuth
func (h *ContractHandler) RegisterContract(c *gin.Context) {
	var req entities.RegisterContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contract, err := h.service.RegisterContract(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case services.ErrInvalidModule, services.ErrInvalidTemplateKey, services.ErrInvalidChannelList:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Return 201 for new contracts, 200 for updates
	statusCode := http.StatusCreated
	c.JSON(statusCode, contract.ToResponse())
}

// GetContract retrieves a contract by module and template_key
// @Summary Get template contract
// @Description Get contract by module and template key
// @Tags Template Contracts
// @ID getContractByKey
// @Produce json
// @Param module path string true "Module name"
// @Param template_key path string true "Template key"
// @Success 200 {object} entities.ContractResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/contracts/by-key/{module}/{template_key} [get]
// @Security BearerAuth
func (h *ContractHandler) GetContract(c *gin.Context) {
	module := c.Param("module")
	templateKey := c.Param("template_key")

	contract, err := h.service.GetContract(c.Request.Context(), module, templateKey)
	if err != nil {
		if err == services.ErrContractNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, contract.ToResponse())
}

// GetContractByID retrieves a contract by ID
// @Summary Get template contract by ID
// @Description Get contract by ID
// @Tags Template Contracts
// @ID getContractByID
// @Produce json
// @Param id path int true "Contract ID"
// @Success 200 {object} entities.ContractResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/contracts/{id} [get]
// @Security BearerAuth
func (h *ContractHandler) GetContractByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contract ID"})
		return
	}

	contract, err := h.service.GetContractByID(c.Request.Context(), uint(id))
	if err != nil {
		if err == services.ErrContractNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, contract.ToResponse())
}

// ListContracts lists all contracts, optionally filtered by module
// @Summary List template contracts
// @Description List all contracts or filter by module
// @Tags Template Contracts
// @ID listContracts
// @Produce json
// @Param module query string false "Filter by module name"
// @Success 200 {array} entities.ContractResponse
// @Failure 500 {object} map[string]string
// @Router /templates/contracts [get]
// @Security BearerAuth
func (h *ContractHandler) ListContracts(c *gin.Context) {
	module := c.Query("module")

	contracts, err := h.service.ListContracts(c.Request.Context(), module)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]entities.ContractResponse, len(contracts))
	for i, contract := range contracts {
		responses[i] = contract.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// UpdateContract updates a contract
// @Summary Update template contract
// @Description Update an existing template contract
// @Tags Template Contracts
// @ID updateContract
// @Accept json
// @Produce json
// @Param id path int true "Contract ID"
// @Param request body entities.UpdateContractRequest true "Update data"
// @Success 200 {object} entities.ContractResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/contracts/{id} [put]
// @Security BearerAuth
func (h *ContractHandler) UpdateContract(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contract ID"})
		return
	}

	var req entities.UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contract, err := h.service.UpdateContract(c.Request.Context(), uint(id), &req)
	if err != nil {
		if err == services.ErrContractNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, contract.ToResponse())
}

// DeleteContract deletes a contract
// @Summary Delete template contract
// @Description Delete a template contract (only if not in use)
// @Tags Template Contracts
// @ID deleteContract
// @Param id path int true "Contract ID"
// @Success 204 "Contract deleted"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /templates/contracts/{id} [delete]
// @Security BearerAuth
func (h *ContractHandler) DeleteContract(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contract ID"})
		return
	}

	if err := h.service.DeleteContract(c.Request.Context(), uint(id)); err != nil {
		if err == services.ErrContractNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "contract not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
