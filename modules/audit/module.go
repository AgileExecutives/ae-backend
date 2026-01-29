package audit

import (
	"github.com/gin-gonic/gin"
	"github.com/unburdy/audit-module/entities"
	"github.com/unburdy/audit-module/handlers"
	"github.com/unburdy/audit-module/routes"
	"github.com/unburdy/audit-module/services"
	"gorm.io/gorm"
)

type Module struct {
	db            *gorm.DB
	service       *services.AuditService
	handler       *handlers.AuditHandler
	routeProvider *routes.RouteProvider
}

func NewModule(db *gorm.DB) *Module {
	service := services.NewAuditService(db)
	handler := handlers.NewAuditHandler(service)
	routeProvider := routes.NewRouteProvider(handler)

	return &Module{
		db:            db,
		service:       service,
		handler:       handler,
		routeProvider: routeProvider,
	}
}

func (m *Module) GetService() *services.AuditService {
	return m.service
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	m.routeProvider.RegisterRoutes(router)
}

func (m *Module) AutoMigrate() error {
	return m.db.AutoMigrate(&entities.AuditLog{})
}

func (m *Module) GetName() string {
	return "audit"
}
