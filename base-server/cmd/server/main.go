package main

import (
	"log"
	"net/http"

	internalDB "github.com/ae-base-server/internal/database"
	"github.com/ae-base-server/internal/router"
	"github.com/ae-base-server/pkg/auth"
	"github.com/ae-base-server/pkg/config"
	"github.com/ae-base-server/pkg/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Set JWT secret
	auth.SetJWTSecret(cfg.JWT.Secret)

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Note: Migrations are now handled by the bootstrap system
	// This legacy server doesn't use the bootstrap system, so no migrations are run
	log.Println("⚠️  Warning: This legacy server doesn't run migrations. Use the main server with bootstrap system.")

	// Seed database with initial data
	if err := internalDB.Seed(db); err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	// Setup router
	r := router.SetupRouter(db, cfg)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting AE SaaS Basic server on %s", addr)
	log.Printf("Health check available at: http://%s/api/v1/health", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
