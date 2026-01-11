package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	// Settings system
	"github.com/ae-base-server/pkg/settings"
)

func main() {
	// Initialize database connection
	db, err := initializeDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize settings system
	settingsSystem, err := settings.NewSettingsSystem(db)
	if err != nil {
		log.Fatal("Failed to initialize settings system:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Setup middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Register settings routes
	settingsSystem.RegisterRoutes(r)

	// Start server
	port := ":8080"
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📊 Settings API available at http://localhost%s/api/v1/settings", port)
	log.Println("✅ Settings system initialized successfully")

	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// initializeDatabase initializes the database connection
func initializeDatabase() (*gorm.DB, error) {
	// Replace with your actual database configuration
	dsn := "host=localhost user=postgres password=password dbname=ae_backend port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("✅ Database initialized")
	return db, nil
}
