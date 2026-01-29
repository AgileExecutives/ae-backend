package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/unburdy/audit-module/handlers"
)

type RouteProvider struct {
	auditHandler *handlers.AuditHandler
}

func NewRouteProvider(auditHandler *handlers.AuditHandler) *RouteProvider {
	return &RouteProvider{
		auditHandler: auditHandler,
	}
}

func (rp *RouteProvider) RegisterRoutes(router *gin.RouterGroup) {
	audit := router.Group("/audit")
	{
		audit.GET("/logs", rp.auditHandler.GetAuditLogs)
		audit.GET("/entity/:entity_type/:entity_id", rp.auditHandler.GetEntityAuditLogs)
		audit.GET("/export", rp.auditHandler.ExportAuditLogs)
		audit.GET("/statistics", rp.auditHandler.GetAuditStatistics)
	}
}
