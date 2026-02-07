package handlers

import (
	cmhandlers "github.com/unburdy/unburdy-server-api/modules/client_management/handlers"
	cmservices "github.com/unburdy/unburdy-server-api/modules/client_management/services"
)

// ClientHandler is kept for backward compatibility; it re-exports the module handler.
type ClientHandler = cmhandlers.ClientHandler

// NewClientHandler re-exports modules/client_management/handlers.NewClientHandler.
func NewClientHandler(clientService *cmservices.ClientService) *ClientHandler {
	return cmhandlers.NewClientHandler(clientService)
}
