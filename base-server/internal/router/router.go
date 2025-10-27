package router

import (
	"github.com/ae-saas-basic/ae-saas-basic/internal/config"
	"github.com/ae-saas-basic/ae-saas-basic/internal/handlers"
	"github.com/ae-saas-basic/ae-saas-basic/internal/middleware"
	"github.com/ae-saas-basic/ae-saas-basic/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// SetupRouter sets up the Gin router with all routes
func SetupRouter(db *gorm.DB, cfg config.Config) *gin.Engine {
	router := gin.Default()

	// Add middleware
	router.Use(middleware.CORSMiddleware())

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve static files
	router.Static("/static", "./statics")
	router.StaticFile("/favicon.ico", "./statics/images/favicon.ico")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db)
	healthHandler := handlers.NewHealthHandler(db, cfg)
	planHandler := handlers.NewPlanHandler(db)
	customerHandler := handlers.NewCustomerHandler(db)
	contactHandler := handlers.NewContactHandler(db)
	emailHandler := handlers.NewEmailHandler(db)
	userSettingsHandler := handlers.NewUserSettingsHandler(db)
	staticHandler := handlers.NewStaticHandler("./statics")

	// Initialize PDF service and handler
	pdfServiceConfig := &services.PDFConfig{
		PageSize:     cfg.PDF.PageSize,
		Orientation:  cfg.PDF.Orientation,
		MarginTop:    cfg.PDF.MarginTop,
		MarginRight:  cfg.PDF.MarginRight,
		MarginBottom: cfg.PDF.MarginBottom,
		MarginLeft:   cfg.PDF.MarginLeft,
		Quality:      cfg.PDF.Quality,
		EnableJS:     cfg.PDF.EnableJS,
		LoadTimeout:  cfg.PDF.LoadTimeout,
		Headers:      make(map[string]string),
	}
	pdfService := services.NewPDFService(cfg.PDF.TemplateDir, cfg.PDF.OutputDir, pdfServiceConfig)
	pdfHandler := handlers.NewPDFHandler(pdfService)

	// Initialize fuzzy search service and handler
	fuzzySearchService := services.NewFuzzySearchService(db, nil)
	fuzzySearchHandler := handlers.NewFuzzySearchHandler(fuzzySearchService)

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		// Health checks
		public.GET("/health", healthHandler.Health)
		public.GET("/ping", healthHandler.Ping)

		// Authentication (with rate limiting)
		// Login: 5 attempts per 15 minutes per IP
		public.POST("/auth/login",
			middleware.NewRateLimiter(middleware.LoginRateLimiter),
			authHandler.Login,
		)

		// Register: 3 attempts per hour per IP
		public.POST("/auth/register",
			middleware.NewRateLimiter(middleware.RegisterRateLimiter),
			authHandler.Register,
		)

		// Password reset: 3 attempts per hour per IP
		public.POST("/auth/forgot-password",
			middleware.NewRateLimiter(middleware.PasswordResetRateLimiter),
			authHandler.ForgotPassword,
		)
		public.POST("/auth/new-password/:token",
			middleware.NewRateLimiter(middleware.PasswordResetRateLimiter),
			authHandler.ResetPassword,
		)

		// Email verification: 5 attempts per hour per IP
		public.GET("/auth/verify-email/:token",
			middleware.NewRateLimiter(middleware.EmailVerificationRateLimiter),
			authHandler.VerifyEmail,
		)

		// Other auth routes without strict limits
		public.GET("/auth/password-security", authHandler.GetPasswordSecurity)

		// Public plans (for signup pages)
		public.GET("/plans", planHandler.GetPlans)
		public.GET("/plans/:id", planHandler.GetPlan)

		// Static asset serving
		public.GET("/assets/*path", staticHandler.ServeAsset)
		public.GET("/logo", staticHandler.ServeLogo)

		// Public PDF routes (no public routes needed for PDF)
		// All PDF functionality requires authentication

		// Public fuzzy search routes (basic search only)
		search := public.Group("/search")
		{
			search.GET("/quick", fuzzySearchHandler.QuickSearch)
			search.GET("/types", fuzzySearchHandler.GetEntityTypes)
			search.GET("/suggestions", fuzzySearchHandler.SearchSuggestions)
			search.GET("/health", fuzzySearchHandler.HealthCheck)
		}

		// Public contact form route
		contact := public.Group("/contact")
		{
			contact.POST("/form", contactHandler.SubmitContactForm)
		}
	}

	// Protected routes (authentication required)
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(db))
	{
		// Auth routes for authenticated users
		auth := protected.Group("/auth")
		{
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/change-password", authHandler.ChangePassword)
			auth.GET("/me", authHandler.Me)
			auth.POST("/generate-signup-token", authHandler.GenerateSignupToken)
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

		// Newsletter management routes (protected)
		newsletter := protected.Group("/contact")
		{
			newsletter.GET("/newsletter", contactHandler.GetNewsletterSubscriptions)
			newsletter.DELETE("/newsletter/unsubscribe", contactHandler.UnsubscribeFromNewsletter)
		}

		// Email routes
		emails := protected.Group("/emails")
		{
			emails.GET("", emailHandler.GetEmails)
			emails.GET("/:id", emailHandler.GetEmail)
			emails.POST("/send", emailHandler.SendEmail)
			emails.GET("/stats", emailHandler.GetEmailStats)
		}

		// User settings routes
		userSettings := protected.Group("/user-settings")
		{
			userSettings.GET("", userSettingsHandler.GetUserSettings)
			userSettings.PUT("", userSettingsHandler.UpdateUserSettings)
			userSettings.POST("/reset", userSettingsHandler.ResetUserSettings)
		}

		// Static file management (authenticated)
		statics := protected.Group("/static")
		{
			statics.GET("/assets", staticHandler.ListAssets)
			statics.GET("/templates/:type/:template", staticHandler.ServeTemplate)
		}

		// PDF generation routes (authenticated)
		pdf := protected.Group("/pdf")
		{
			// Template management
			pdf.GET("/templates", pdfHandler.ListTemplates)
			pdf.GET("/templates/:template", pdfHandler.GetTemplateInfo)
			pdf.POST("/templates/:template/preview", pdfHandler.PreviewTemplate)

			// PDF generation
			pdf.POST("/generate", pdfHandler.GeneratePDF)
			pdf.POST("/generate/html", pdfHandler.GeneratePDFFromHTML)
			pdf.POST("/generate/stream", pdfHandler.StreamPDF)

			// Configuration
			pdf.GET("/config", pdfHandler.GetPDFConfig)
		}

		// Fuzzy search routes (authenticated)
		search := protected.Group("/search")
		{
			// Advanced search
			search.POST("", fuzzySearchHandler.Search)
			search.POST("/advanced", fuzzySearchHandler.Search)
			search.POST("/entities/:entity_type", fuzzySearchHandler.SearchInEntity)

			// Search management (authenticated only)
			search.GET("/config", fuzzySearchHandler.GetSearchConfig)
			search.GET("/stats", fuzzySearchHandler.SearchStats)

			// Authenticated search (GET)
			search.GET("", fuzzySearchHandler.Search)
		}
	}

	// Admin routes (admin authentication required)
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(db))
	admin.Use(middleware.RequireAdmin())
	{
		// Admin plan management
		adminPlans := admin.Group("/plans")
		{
			adminPlans.POST("", planHandler.CreatePlan)
			adminPlans.PUT("/:id", planHandler.UpdatePlan)
			adminPlans.DELETE("/:id", planHandler.DeletePlan)
		}

		// Admin search management
		adminSearch := admin.Group("/search")
		{
			adminSearch.PUT("/config", fuzzySearchHandler.UpdateSearchConfig)
			adminSearch.POST("/entities/:entity_type", fuzzySearchHandler.RegisterCustomEntity)
			adminSearch.GET("/stats", fuzzySearchHandler.SearchStats)
		}
	}

	return router
}
