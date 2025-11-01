package seeding

import (
	"fmt"
	"time"
)

// GetRecommendedSeedingApproach returns instructions for seeding calendar data
func GetRecommendedSeedingApproach() string {
	return `
🚀 Calendar Seeding Integration Approaches:

1️⃣  SIMPLE APPROACH - Direct Integration:
   Add this to your server's main.go or startup code:
   
   import calendarSeeding "github.com/unburdy/calendar-module/seeding"
   
   func seedCalendarData(db *gorm.DB) {
       seeder := calendarSeeding.NewCalendarSeeder(db)
       if err := seeder.SeedCalendarData(2, 2); err != nil { // tenant=2, user=2
           log.Printf("Warning: Calendar seeding failed: %v", err)
       } else {
           log.Printf("✅ Calendar seeding completed!")
       }
   }

2️⃣  COMMAND-LINE APPROACH - Separate Tool:
   Create a dedicated seeding script in your project root:
   
   // calendar_seed.go
   package main
   
   import (
       "log"
       "github.com/ae-base-server/pkg/config"
       "github.com/ae-base-server/pkg/database"
       calendarSeeding "github.com/unburdy/calendar-module/seeding"
   )
   
   func main() {
       cfg := config.Load()
       db, _ := database.ConnectWithAutoCreate(cfg.Database)
       
       seeder := calendarSeeding.NewCalendarSeeder(db)
       err := seeder.SeedCalendarData(2, 2) // tenant=2, user=2
       if err != nil {
           log.Fatal("Seeding failed:", err)
       }
       log.Println("✅ Calendar seeding completed!")
   }
   
   Run: go run calendar_seed.go

3️⃣  API ENDPOINT APPROACH - Development Only:
   Add a development endpoint to trigger seeding:
   
   // In your development routes
   router.POST("/dev/seed-calendar", func(c *gin.Context) {
       seeder := calendarSeeding.NewCalendarSeeder(db)
       err := seeder.SeedCalendarData(2, 2)
       if err != nil {
           c.JSON(500, gin.H{"error": err.Error()})
           return
       }
       c.JSON(200, gin.H{"message": "Calendar seeded successfully"})
   })
   
   Run: curl -X POST http://localhost:8080/dev/seed-calendar
`
}

// GetSeedingSummary returns a summary of what data will be seeded
func GetSeedingSummary() string {
	now := time.Now()
	startDate := now.AddDate(0, -2, 0)
	endDate := now.AddDate(0, 6, 0)

	return fmt.Sprintf(`Calendar Seeding Strategy Summary:

📅 Time Range: %s to %s (%d months total)

📋 Data Structure:
• 2 Calendars: Work Calendar (blue) + Personal Calendar (green)
• Each calendar has weekly availability rules
• Timezone: Europe/Berlin

🔄 Recurring Series:
• Weekly Team Meeting (Mondays 09:00-10:00)
• Regular patterns with 95%% attendance probability
• Auto-generates individual entries for each occurrence

📊 Event Density (Realistic busy schedule):
• Past Period (2 months back): 15 events total (less crowded)
• Current Period (±4 weeks): 60 events total (VERY BUSY)  
• Future Period (remaining): 35 events total (moderate)

🎯 Event Types:
Work Calendar:
• Project Reviews (1h meetings)
• 1:1 Meetings (30min)
• Code Reviews (45min) 
• Client Presentations (90min)
• Training Sessions (2h)

Personal Calendar:
• Doctor Appointments (30min)
• Grocery Shopping (1h)
• Lunch with Friends (90min)
• Hobby Time (2h)
• Exercise (1h)

🕐 Realistic Scheduling:
• Work events: 08:00-18:00 weekdays (10%% weekend probability)
• Personal events: 06:00-23:00 any day
• Avoids lunch break (12:00-13:00)
• 15min gaps between events
• Realistic locations and participants

This creates a calendar that looks like a real busy professional's schedule!`,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		8) // 2 + 6 months
}
