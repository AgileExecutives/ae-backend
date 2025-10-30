package main
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	AlternativeEmail     string `json:"alternative_email"`
	AlternativeFirstName string `json:"alternative_first_name"`
	AlternativeLastName  string `json:"alternative_last_name"`
	AlternativePhone     string `json:"alternative_phone"`
	City                 string `json:"city"`
	ContactEmail         string `json:"contact_email"`
	ContactFirstName     string `json:"contact_first_name"`
	ContactLastName      string `json:"contact_last_name"`
	ContactPhone         string `json:"contact_phone"`
	DateOfBirth          string `json:"date_of_birth"`
	Email                string `json:"email"`
	FirstName            string `json:"first_name"`
	Gender               string `json:"gender"`
	LastName             string `json:"last_name"`
	Notes                string `json:"notes"`
	Phone                string `json:"phone"`
	PrimaryLanguage      string `json:"primary_language"`
	Status               string `json:"status"`
	StreetAddress        string `json:"street_address"`
	TherapyTitle         string `json:"therapy_title"`
	Zip                  string `json:"zip"`
	ProviderApprovalCode string `json:"provider_approval_code"` // New field
	ProviderApprovalDate string `json:"provider_approval_date"` // New field
}

type CostProvider struct {
	City         string `json:"city"`
	Department   string `json:"department"`
	Email        string `json:"email"`
	Organization string `json:"organization"`
	Phone        string `json:"phone"`
	State        string `json:"state"`
	Street       string `json:"street"`
	Zip          string `json:"zip"`
}

type SeedData struct {
	CostProviders []CostProvider `json:"cost_providers"`
	Clients       []Client       `json:"clients"`
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Read seed data
	seedPath := filepath.Join("..", "..", "seed_app_data.json")
	seedFile, err := os.ReadFile(seedPath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	var seedData SeedData
	if err := json.Unmarshal(seedFile, &seedData); err != nil {
		log.Fatalf("Failed to parse seed JSON: %v", err)
	}

	// Generate approval codes and dates
	prefixes := []string{"PROV", "APP", "AUTH", "CERT", "APPR"}
	
	for i := range seedData.Clients {
		client := &seedData.Clients[i]
		
		// Generate provider approval code
		prefix := prefixes[rand.Intn(len(prefixes))]
		code := fmt.Sprintf("%s-%d%04d", prefix, 2025, rand.Intn(10000))
		client.ProviderApprovalCode = code
		
		// Generate provider approval date (within last 2 years for active clients)
		if client.Status == "active" {
			// Random date within last 2 years
			days := rand.Intn(730) // 2 years in days
			approvalDate := time.Now().AddDate(0, 0, -days)
			client.ProviderApprovalDate = approvalDate.Format("2006-01-02T15:04:05Z")
		} else if client.Status == "waiting" {
			// No approval date for waiting clients (they're not approved yet)
			client.ProviderApprovalDate = ""
		} else {
			// Archived clients have older approval dates (2-5 years ago)
			days := rand.Intn(1095) + 730 // 2-5 years ago
			approvalDate := time.Now().AddDate(0, 0, -days)
			client.ProviderApprovalDate = approvalDate.Format("2006-01-02T15:04:05Z")
		}
		
		fmt.Printf("✅ Updated client %s %s: Code=%s, Date=%s\n", 
			client.FirstName, client.LastName, client.ProviderApprovalCode, client.ProviderApprovalDate)
	}

	fmt.Printf("\nUpdated %d clients with provider approval codes and dates\n", len(seedData.Clients))

	// Write updated seed data back to file
	updatedData, err := json.MarshalIndent(seedData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal updated data: %v", err)
	}

	if err := os.WriteFile(seedPath, updatedData, 0644); err != nil {
		log.Fatalf("Failed to write updated seed file: %v", err)
	}

	fmt.Println("Successfully updated seed_app_data.json with provider approval codes and dates!")
	
	// Show some examples
	fmt.Println("\nExample provider approval data assigned:")
	for i, client := range seedData.Clients[:5] {
		fmt.Printf("  %d. %s %s (%s): %s - %s\n", 
			i+1, client.FirstName, client.LastName, client.Status, 
			client.ProviderApprovalCode, client.ProviderApprovalDate)
	}
}