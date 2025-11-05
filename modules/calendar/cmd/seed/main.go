package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ae-base-server/pkg/config"
	"github.com/ae-base-server/pkg/database"
	calendarSeeding "github.com/unburdy/calendar-module/seeding"
)

func main() {
	//Parse command line flags
	tenantID := flag.Uint("tenant", 1, "Tenant ID for seeding")
	userID := flag.Uint("user", 1, "User ID for seeding")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		fmt.Println("Calendar Data Seeder")
		fmt.Println("\nUsage:")
		fmt.Println("  go run cmd/seed/main.go [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment Variables:")
		fmt.Println("  CALENDAR_TENANT_ID - Tenant ID (default: 1)")
		fmt.Println("  CALENDAR_USER_ID   - User ID (default: 1)")
		fmt.Println("\nExample:")
		fmt.Println("  go run cmd/seed/main.go -tenant 2 -user 2")
		fmt.Println("  CALENDAR_TENANT_ID=2 CALENDAR_USER_ID=2 ./seed_calendar.sh")
		os.Exit(0)
	}

	log.Println("🌱 Starting Calendar Data Seeder...")
	log.Printf("📋 Configuration: Tenant=%d, User=%d\n", *tenantID, *userID)

	// Load configuration
	cfg := config.Load()

	// Connect to database
	log.Println("🔌 Connecting to database...")
	db, err := database.ConnectWithAutoCreate(cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connection established")

	// Create seeder
	seeder := calendarSeeding.NewCalendarSeeder(db)

	// Run seeding
	log.Printf("🚀 Seeding calendar data for tenant %d, user %d...\n", *tenantID, *userID)
	if err := seeder.SeedCalendarData(uint(*tenantID), uint(*userID)); err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✅ Calendar data seeded successfully!")
	log.Println("🎉 Done!")
}
