package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	internalDB "github.com/ae-base-server/internal/database"
	internalHandlers "github.com/ae-base-server/internal/handlers"
	"github.com/ae-base-server/internal/middleware"
	internalMiddleware "github.com/ae-base-server/internal/middleware"
	internalServices "github.com/ae-base-server/internal/services"
	pdfServices "github.com/ae-base-server/modules/pdf/services"
	"github.com/ae-base-server/modules/templates/services/storage"
	"github.com/ae-base-server/pkg/auth"
	"github.com/ae-base-server/pkg/config"
	"github.com/ae-base-server/pkg/core"
	"github.com/ae-base-server/pkg/database"
	pkgServices "github.com/ae-base-server/pkg/services"
	"github.com/ae-base-server/pkg/startup"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// Application represents the main application
type Application struct {
	config   *config.Config
	registry core.ModuleRegistry
	context  core.ModuleContext
	server   *gin.Engine
	httpSrv  *http.Server
	logger   core.Logger
}

// NewApplication creates a new application instance
func NewApplication(cfg config.Config) *Application {
	return &Application{
		config:   &cfg,
		registry: core.NewModuleRegistry(),
		logger:   core.NewLogger(),
	}
}

// RegisterModule registers a module with the application
func (app *Application) RegisterModule(module core.Module) error {
	app.logger.Info("Registering module", "name", module.Name(), "version", module.Version())
	return app.registry.Register(module)
}

// Initialize initializes the application and all modules
func (app *Application) Initialize() error {
	app.logger.Info("Initializing application...")

	// 1. Initialize core services
	if err := app.initializeCoreServices(); err != nil {
		return fmt.Errorf("failed to initialize core services: %w", err)
	}

	// 2. Initialize modules
	app.logger.Info("Initializing modules...")
	if err := app.registry.InitializeAll(app.context); err != nil {
		return fmt.Errorf("failed to initialize modules: %w", err)
	}

	// 3. Run migrations (includes all module entities)
	app.logger.Info("Running database migrations...")
	if err := app.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// 3.5. Register contracts BEFORE seeding so contracts exist when templates are seeded
	app.logger.Info("Registering template contracts...")
	if err := app.registerContracts(); err != nil {
		app.logger.Warn("Failed to register contracts:", err)
		// Don't fail startup, contracts can be registered later
	}

	// 4. Seed database
	app.logger.Info("Seeding database...")
	if err := app.seedDatabase(); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	app.logger.Info("Application initialization completed")
	return nil
}

// Start starts the application and all modules
func (app *Application) Start(ctx context.Context) error {
	app.logger.Info("Starting application...")

	// Start event bus
	if err := app.context.EventBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// Start all modules
	if err := app.registry.StartAll(ctx); err != nil {
		return fmt.Errorf("failed to start modules: %w", err)
	}

	// Setup HTTP server
	addr := app.config.Server.Host + ":" + app.config.Server.Port
	app.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      app.server,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Setup graceful shutdown
	go app.setupGracefulShutdown()

	app.logger.Info("Server starting", "address", addr)
	app.logger.Info("Health check available", "url", fmt.Sprintf("http://%s/api/v1/health", addr))
	app.logger.Info("API documentation", "url", fmt.Sprintf("http://%s/swagger/index.html", addr))

	if err := app.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop stops the application and all modules
func (app *Application) Stop(ctx context.Context) error {
	app.logger.Info("Stopping application...")

	// Stop HTTP server
	if app.httpSrv != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := app.httpSrv.Shutdown(shutdownCtx); err != nil {
			app.logger.Error("HTTP server shutdown error", "error", err)
		}
	}

	// Stop modules
	if err := app.registry.StopAll(ctx); err != nil {
		app.logger.Error("Failed to stop modules", "error", err)
		return err
	}

	// Stop event bus
	if err := app.context.EventBus.Stop(ctx); err != nil {
		app.logger.Error("Failed to stop event bus", "error", err)
	}

	app.logger.Info("Application stopped")
	return nil
}

// GetModuleMetadata returns metadata for all registered modules
func (app *Application) GetModuleMetadata() []core.ModuleMetadata {
	return app.registry.GetMetadata()
}

// GetServiceRegistry returns the service registry
func (app *Application) GetServiceRegistry() core.ServiceRegistry {
	return app.context.Services
}

// DB returns the database connection
func (app *Application) DB() *gorm.DB {
	return app.context.DB
}

// initializeCoreServices initializes core application services
func (app *Application) initializeCoreServices() error {
	// Set Gin mode
	gin.SetMode(app.config.Server.Mode)

	// Set JWT secret
	auth.SetJWTSecret(app.config.JWT.Secret)

	// Database
	db, err := database.ConnectWithAutoCreate(app.config.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Router with core middleware
	router := gin.New()
	router.Use(
		gin.Logger(),
		gin.Recovery(),
		app.corsMiddleware(),
		app.securityMiddleware(),
	)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup miscellaneous static file routes (favicon, robots.txt, etc.)
	app.setupStaticMiscRoutes(router)

	// ✅ Auth routes now handled by modular base module
	// app.setupInternalAuthRoutes(router, db, *app.config)

	// Event Bus
	eventBus := core.NewEventBus()

	// Auth Service
	authService := &authServiceAdapter{
		jwtSecret: app.config.JWT.Secret,
		db:        db,
	}

	// Token Service - shared JWT token service for all modules
	tokenService := auth.NewTokenService(app.config.JWT.Secret)

	// Service Registry
	services := core.NewServiceRegistry()

	// Create module context
	app.context = core.ModuleContext{
		DB:             db,
		Router:         router,
		EventBus:       eventBus,
		Config:         app.config,
		Logger:         app.logger,
		Services:       services,
		Auth:           authService,
		TokenService:   tokenService,
		ModuleRegistry: app.registry,
	}

	app.server = router
	return nil
}

// setupInternalAuthRoutes temporarily adds internal auth routes until modular auth is implemented
func (app *Application) setupInternalAuthRoutes(router *gin.Engine, db *gorm.DB, cfg config.Config) {
	// Import needed for internal handlers - these imports will need to be added at the top
	// For now, we'll add the most critical auth routes manually

	// Initialize MinIO storage for tenant buckets
	minioConfig := storage.MinIOConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin123",
		UseSSL:          false,
		Region:          "us-east-1",
	}
	minioStorage, err := storage.NewMinIOStorage(minioConfig)
	if err != nil {
		// Log error but don't fail entirely - some features may not work
		log.Printf("⚠️ Warning: Failed to initialize MinIO storage: %v", err)
	}

	// Initialize services for tenant bucket management
	tenantBucketService := pkgServices.NewTenantBucketService(minioStorage)
	tenantService := internalServices.NewTenantService(db, tenantBucketService)

	// Initialize internal auth handler with tenant service
	authHandler := internalHandlers.NewAuthHandler(db, tenantService)

	// Public routes group
	public := router.Group("/api/v1")

	// Authentication routes with rate limiting
	public.POST("/auth/login",
		internalMiddleware.NewRateLimiter(internalMiddleware.LoginRateLimiter),
		authHandler.Login,
	)
	public.POST("/auth/register",
		internalMiddleware.NewRateLimiter(internalMiddleware.RegisterRateLimiter),
		authHandler.Register,
	)
	public.POST("/auth/forgot-password",
		internalMiddleware.NewRateLimiter(internalMiddleware.PasswordResetRateLimiter),
		authHandler.ForgotPassword,
	)
	public.POST("/auth/new-password/:token",
		internalMiddleware.NewRateLimiter(internalMiddleware.PasswordResetRateLimiter),
		authHandler.ResetPassword,
	)
	public.GET("/auth/verify-email/:token",
		internalMiddleware.NewRateLimiter(internalMiddleware.EmailVerificationRateLimiter),
		authHandler.VerifyEmail,
	)

	// Initialize other internal handlers
	planHandler := internalHandlers.NewPlanHandler(db)
	customerHandler := internalHandlers.NewCustomerHandler(db)
	contactHandler := internalHandlers.NewContactHandler(db)
	userSettingsHandler := internalHandlers.NewUserSettingsHandler(db)

	// Initialize PDF service and handler
	pdfService := pdfServices.NewPDFGenerator()
	pdfHandler := internalHandlers.NewPDFHandler(pdfService)

	// Public routes for plans (used by signup pages)
	public.GET("/plans", planHandler.GetPlans)
	public.GET("/plans/:id", planHandler.GetPlan)

	// Public contact form
	public.POST("/contact/form", contactHandler.SubmitContactForm)

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(internalMiddleware.AuthMiddleware(db))

	// Auth routes for authenticated users
	auth := protected.Group("/auth")
	{
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/change-password", authHandler.ChangePassword)
		auth.GET("/me", authHandler.Me)
	}

	// Customer routes
	customers := protected.Group("/customers")
	{
		customers.GET("", customerHandler.GetCustomers)
		customers.GET("/:id", customerHandler.GetCustomer)
		customers.POST("", customerHandler.CreateCustomer)
		customers.PUT("/:id", customerHandler.UpdateCustomer)
		customers.DELETE("/:id", customerHandler.DeleteCustomer)
	}

	// Contact routes
	contacts := protected.Group("/contacts")
	{
		contacts.GET("", contactHandler.GetContacts)
		contacts.GET("/:id", contactHandler.GetContact)
		contacts.POST("", contactHandler.CreateContact)
		contacts.PUT("/:id", contactHandler.UpdateContact)
		contacts.DELETE("/:id", contactHandler.DeleteContact)
	}

	// User settings routes
	userSettings := protected.Group("/user-settings")
	{
		userSettings.GET("", userSettingsHandler.GetUserSettings)
		userSettings.PUT("", userSettingsHandler.UpdateUserSettings)
		userSettings.POST("/reset", userSettingsHandler.ResetUserSettings)
	}

	// PDF generation routes
	pdf := protected.Group("/pdf")
	{
		pdf.POST("/create", pdfHandler.GeneratePDFFromTemplate)
	}
}

// runMigrations runs database migrations for all registered module entities
func (app *Application) runMigrations() error {
	// Collect entities from all registered modules (including base module)
	entities := make([]interface{}, 0)
	moduleCount := 0

	for _, module := range app.registry.GetAll() {
		moduleEntities := module.Entities()
		if len(moduleEntities) > 0 {
			app.logger.Info("Collecting entities from module", "module", module.Name(), "count", len(moduleEntities))
			for _, entity := range moduleEntities {
				entities = append(entities, entity.GetModel())
			}
			moduleCount++
		}
	}

	if len(entities) > 0 {
		app.logger.Info("Running migrations", "modules", moduleCount, "entities", len(entities))
		if err := app.context.DB.AutoMigrate(entities...); err != nil {
			return fmt.Errorf("failed to migrate entities: %w", err)
		}
		app.logger.Info("All entity migrations completed successfully")
	} else {
		app.logger.Info("No entities to migrate")
	}

	return nil
}

// registerContracts registers all template contracts from modules
func (app *Application) registerContracts() error {
	app.logger.Info("Registering template contracts...")
	err := startup.RegisterAllContracts(app.context.DB)
	if err != nil {
		app.logger.Error("Failed to register contracts", "error", err)
		return err
	}
	app.logger.Info("Template contracts registered successfully")
	return nil
}

// seedDatabase seeds the database with initial data
func (app *Application) seedDatabase() error {
	// Initialize MinIO storage for tenant buckets
	minioConfig := storage.MinIOConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin123",
		UseSSL:          false,
		Region:          "us-east-1",
	}
	minioStorage, err := storage.NewMinIOStorage(minioConfig)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to initialize MinIO storage for seeding: %v", err)
	}

	// Create services for tenant bucket management
	tenantBucketService := pkgServices.NewTenantBucketService(minioStorage)
	tenantService := internalServices.NewTenantService(app.context.DB, tenantBucketService)

	// Use the enhanced seed function that creates MinIO buckets
	return internalDB.Seed(app.context.DB, tenantService)
}

// corsMiddleware adds CORS headers
func (app *Application) corsMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// securityMiddleware adds security headers
func (app *Application) securityMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	})
}

// setupStaticMiscRoutes sets up routes for static miscellaneous files (favicon, robots.txt, etc.)
func (app *Application) setupStaticMiscRoutes(router *gin.Engine) {
	// Serve favicon
	router.StaticFile("/favicon.ico", "./statics/images/favicon.ico")
}

// setupGracefulShutdown sets up graceful shutdown handling
func (app *Application) setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	app.logger.Info("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Stop(ctx); err != nil {
		app.logger.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}
}

// authServiceAdapter adapts the existing auth package to the core.AuthService interface
type authServiceAdapter struct {
	jwtSecret string
	db        *gorm.DB
}

func (a *authServiceAdapter) ValidateToken(token string) (interface{}, error) {
	return auth.ValidateJWT(token)
}

func (a *authServiceAdapter) GenerateToken(user interface{}) (string, error) {
	// This would need to be implemented based on your existing auth logic
	return "", fmt.Errorf("not implemented")
}

func (a *authServiceAdapter) GetCurrentUser(c *gin.Context) (interface{}, error) {
	// This would need to be implemented based on your existing auth logic
	return nil, fmt.Errorf("not implemented")
}

func (a *authServiceAdapter) RequireAuth() gin.HandlerFunc {
	return middleware.AuthMiddleware(a.db)
}

func (a *authServiceAdapter) RequireRole(roles ...string) gin.HandlerFunc {
	return middleware.RequireRole(roles...)
}

// GetDB returns the database instance
func (app *Application) GetDB() *gorm.DB {
	return app.context.DB
}

// GetRouter returns the Gin router instance
func (app *Application) GetRouter() *gin.Engine {
	return app.server
}
