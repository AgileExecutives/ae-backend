package eventbus

import (
	"log"

	"github.com/ae-base-server/internal/eventbus/handlers"
) // InitializeEventHandlers sets up all event handlers
func InitializeEventHandlers() {
	log.Println("Initializing event handlers...")

	// Initialize calendar handler
	calendarHandler := handlers.NewCalendarHandler()
	if err := Subscribe(calendarHandler); err != nil {
		log.Printf("Failed to subscribe calendar handler: %v", err)
	} else {
		log.Printf("Successfully subscribed calendar handler for events: %v", calendarHandler.GetEventTypes())
	}

	// Add more handlers here as needed
	// Example:
	// emailHandler := handlers.NewEmailHandler()
	// if err := Subscribe(emailHandler); err != nil {
	//     log.Printf("Failed to subscribe email handler: %v", err)
	// }

	log.Println("Event handlers initialization complete")
}
