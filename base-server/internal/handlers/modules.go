package handlers

import (
"net/http"

"github.com/ae-base-server/internal/models"
"github.com/ae-base-server/internal/modules"
"github.com/gin-gonic/gin"
)

// ModuleHandler manages module-related operations
type ModuleHandler struct {
	manager *modules.Manager
}

// NewModuleHandler creates a new module handler
func NewModuleHandler(manager *modules.Manager) *ModuleHandler {
	return &ModuleHandler{
		manager: manager,
	}
}

// GetModules returns information about all registered modules
// @Summary Get registered modules
// @Description Get information about all registered modules in the system
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=[]modules.ModuleInfo}
// @Failure 401 {object} models.ErrorResponse
// @Router /admin/modules [get]
func (h *ModuleHandler) GetModules(c *gin.Context) {
	moduleInfo := h.manager.GetModuleInfo()
	c.JSON(http.StatusOK, models.SuccessResponse("Modules retrieved successfully", moduleInfo))
}

// GetModuleInfo returns detailed information about a specific module
// @Summary Get module details
// @Description Get detailed information about a specific module by name
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name path string true "Module name"
// @Success 200 {object} models.APIResponse{data=modules.ModuleInfo}
// @Failure 404 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /admin/modules/{name} [get]
func (h *ModuleHandler) GetModuleInfo(c *gin.Context) {
	name := c.Param("name")
	
	module, exists := h.manager.GetRegistry().GetModule(name)
	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponseFunc("Module not found", "Module does not exist"))
		return
	}
	
	moduleInfo := modules.ModuleInfo{
		Name:        module.GetName(),
		Version:     module.GetVersion(),
		Description: "Module: " + module.GetName(),
		Author:      "Base Server",
		Enabled:     true, // All registered modules are enabled by default
	}
	
	c.JSON(http.StatusOK, models.SuccessResponse("Module information retrieved", moduleInfo))
}
