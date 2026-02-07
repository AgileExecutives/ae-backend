package handlers

import (
	cmhandlers "github.com/unburdy/unburdy-server-api/modules/client_management/handlers"
	cmservices "github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// CostProviderHandler is kept for backward compatibility; it re-exports the module handler.
type CostProviderHandler = cmhandlers.CostProviderHandler

// NewCostProviderHandler re-exports modules/client_management/handlers.NewCostProviderHandler.
func NewCostProviderHandler(costProviderService *cmservices.CostProviderService, clientService *cmservices.ClientService) *CostProviderHandler {
	return cmhandlers.NewCostProviderHandler(costProviderService, clientService)
}
