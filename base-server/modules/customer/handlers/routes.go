package handlers

import (
	"github.com/ae-base-server/pkg/core"
	"github.com/gin-gonic/gin"
)

// CustomerRoutes provides customer management routes
type CustomerRoutes struct {
	handlers *CustomerHandlers
}

// NewCustomerRoutes creates new customer route provider
func NewCustomerRoutes(handlers *CustomerHandlers) core.RouteProvider {
	return &CustomerRoutes{
		handlers: handlers,
	}
}

func (r *CustomerRoutes) GetPrefix() string {
	return "/customers"
}

func (r *CustomerRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (r *CustomerRoutes) GetSwaggerTags() []string {
	return []string{"customers"}
}

func (r *CustomerRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	// All customer routes require authentication
	authRouter := router.Use(ctx.Auth.RequireAuth())

	authRouter.GET("", r.handlers.GetCustomers)
	authRouter.POST("", r.handlers.CreateCustomer)
	authRouter.GET("/:id", r.handlers.GetCustomer)
	authRouter.PUT("/:id", r.handlers.UpdateCustomer)
	authRouter.DELETE("/:id", r.handlers.DeleteCustomer)
}

// PlanRoutes provides subscription plan routes
type PlanRoutes struct {
	handlers *PlanHandlers
}

// NewPlanRoutes creates new plan route provider
func NewPlanRoutes(handlers *PlanHandlers) core.RouteProvider {
	return &PlanRoutes{
		handlers: handlers,
	}
}

func (r *PlanRoutes) GetPrefix() string {
	return "/plans"
}

func (r *PlanRoutes) GetMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (r *PlanRoutes) GetSwaggerTags() []string {
	return []string{"plans"}
}

func (r *PlanRoutes) RegisterRoutes(router *gin.RouterGroup, ctx core.ModuleContext) {
	// Public routes - no auth required for viewing plans
	router.GET("", r.handlers.GetPlans)
	router.GET("/:id", r.handlers.GetPlan)

	// Admin only routes - require auth and admin role
	adminRouter := router.Use(ctx.Auth.RequireAuth(), ctx.Auth.RequireRole("admin", "super-admin"))
	adminRouter.POST("", r.handlers.CreatePlan)
	adminRouter.PUT("/:id", r.handlers.UpdatePlan)
	adminRouter.DELETE("/:id", r.handlers.DeletePlan)
}
