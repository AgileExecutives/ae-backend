package routes

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/unburdy/unburdy-server-api/modules/client_management/handlers"
)

// RouteProvider provides routing functionality for client management
type RouteProvider struct {
	clientHandler         *handlers.ClientHandler
	costProviderHandler   *handlers.CostProviderHandler
	sessionHandler        *handlers.SessionHandler
	invoiceHandler        *handlers.InvoiceHandler
	invoiceAdapterHandler *handlers.InvoiceAdapterHandler
	extraEffortHandler    *handlers.ExtraEffortHandler
	db                    *gorm.DB
}

// NewRouteProvider creates a new route provider
func NewRouteProvider(clientHandler *handlers.ClientHandler, costProviderHandler *handlers.CostProviderHandler, sessionHandler *handlers.SessionHandler, invoiceHandler *handlers.InvoiceHandler, invoiceAdapterHandler *handlers.InvoiceAdapterHandler, extraEffortHandler *handlers.ExtraEffortHandler, db *gorm.DB) *RouteProvider {
	return &RouteProvider{
		clientHandler:         clientHandler,
		costProviderHandler:   costProviderHandler,
		sessionHandler:        sessionHandler,
		invoiceHandler:        invoiceHandler,
		invoiceAdapterHandler: invoiceAdapterHandler,
		extraEffortHandler:    extraEffortHandler,
		db:                    db,
	}
}

// RegisterRoutes registers the client management routes with the provided router group
func (rp *RouteProvider) RegisterRoutes(router *gin.RouterGroup, ctx *core.ModuleContext) {
	// Note: Auth middleware is already applied at the router group level

	// Register public token booking route on base router (bypass auth middleware)
	// The token itself is the authorization - no JWT authentication required
	if ctx != nil && ctx.Router != nil {
		ctx.Router.POST("/api/v1/sessions/book/:token", rp.sessionHandler.BookSessionsWithToken)
	}

	// Client management endpoints (authenticated)
	clients := router.Group("/clients")
	{
		clients.POST("", rp.clientHandler.CreateClient)
		clients.GET("", rp.clientHandler.GetAllClients)
		clients.GET("/search", rp.clientHandler.SearchClients)
		clients.GET("/:id/sessions", rp.sessionHandler.GetSessionsByClient)
		clients.GET("/:id", rp.clientHandler.GetClient)
		clients.PUT("/:id", rp.clientHandler.UpdateClient)
		clients.DELETE("/:id", rp.clientHandler.DeleteClient)
	}

	// Cost provider management endpoints (authenticated)
	costProviders := router.Group("/cost-providers")
	{
		costProviders.POST("", rp.costProviderHandler.CreateCostProvider)
		costProviders.GET("", rp.costProviderHandler.GetAllCostProviders)
		costProviders.GET("/search", rp.costProviderHandler.SearchCostProviders)
		costProviders.GET("/:id", rp.costProviderHandler.GetCostProvider)
		costProviders.PUT("/:id", rp.costProviderHandler.UpdateCostProvider)
		costProviders.DELETE("/:id", rp.costProviderHandler.DeleteCostProvider)
	}

	// Session management endpoints (authenticated)
	sessions := router.Group("/sessions")
	{
		sessions.POST("", rp.sessionHandler.CreateSession)
		sessions.POST("/book", rp.sessionHandler.BookSessions)
		sessions.GET("", rp.sessionHandler.GetAllSessions)
		sessions.GET("/detail", rp.sessionHandler.GetDetailedSessionsUpcoming)
		sessions.GET("/by_entry/:id", rp.sessionHandler.GetSessionByCalendarEntry)
		sessions.GET("/:id", rp.sessionHandler.GetSession)
		sessions.PUT("/:id", rp.sessionHandler.UpdateSession)
		sessions.DELETE("/:id", rp.sessionHandler.DeleteSession)
	}

	// Client invoice management endpoints (authenticated)
	clientInvoices := router.Group("/client-invoices")
	{
		clientInvoices.GET("/unbilled-sessions", rp.invoiceHandler.GetClientsWithUnbilledSessions)
		clientInvoices.GET("/vat-categories", rp.invoiceHandler.GetVATCategories)
		clientInvoices.POST("/from-sessions", rp.invoiceAdapterHandler.CreateInvoiceFromSessions)
		clientInvoices.POST("/draft", rp.invoiceHandler.CreateDraftInvoice)
		clientInvoices.POST("", rp.invoiceHandler.CreateInvoice)
		clientInvoices.GET("", rp.invoiceHandler.GetAllInvoices)
		clientInvoices.GET("/:id", rp.invoiceHandler.GetInvoice)
		clientInvoices.GET("/:id/pdf", rp.invoiceHandler.DownloadInvoicePDF)
		clientInvoices.GET("/:id/preview-pdf", rp.invoiceHandler.PreviewInvoicePDF)
		clientInvoices.GET("/:id/xrechnung", rp.invoiceHandler.ExportXRechnung)
		clientInvoices.PUT("/:id", rp.invoiceHandler.UpdateDraftInvoice)
		clientInvoices.DELETE("/:id", rp.invoiceHandler.CancelDraftInvoice)
		clientInvoices.POST("/:id/finalize", rp.invoiceHandler.FinalizeInvoice)
		clientInvoices.POST("/:id/mark-sent", rp.invoiceHandler.MarkInvoiceAsSent)
		clientInvoices.POST("/:id/send-email", rp.invoiceHandler.SendInvoiceEmail)
		clientInvoices.POST("/:id/mark-paid", rp.invoiceHandler.MarkInvoiceAsPaid)
		clientInvoices.POST("/:id/mark-overdue", rp.invoiceHandler.MarkInvoiceAsOverdue)
		clientInvoices.POST("/:id/reminder", rp.invoiceHandler.SendReminder)
		clientInvoices.POST("/:id/credit-note", rp.invoiceHandler.CreateCreditNote)
	}

	// Extra efforts management endpoints (authenticated)
	extraEfforts := router.Group("/extra-efforts")
	{
		extraEfforts.POST("", rp.extraEffortHandler.CreateExtraEffort)
		extraEfforts.GET("", rp.extraEffortHandler.ListExtraEfforts)
		extraEfforts.GET("/:id", rp.extraEffortHandler.GetExtraEffort)
		extraEfforts.PUT("/:id", rp.extraEffortHandler.UpdateExtraEffort)
		extraEfforts.DELETE("/:id", rp.extraEffortHandler.DeleteExtraEffort)
	}
}

// GetPrefix returns the route prefix for client management endpoints
func (rp *RouteProvider) GetPrefix() string {
	return ""
}

// GetMiddleware returns middleware to apply to all routes
func (rp *RouteProvider) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		middleware.AuthMiddleware(rp.db), // Require authentication for all client management routes
	}
}

// GetSwaggerTags returns swagger tags for the routes
func (rp *RouteProvider) GetSwaggerTags() []string {
	return []string{"clients", "cost-providers", "sessions", "invoices"}
}
