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

type DiagnosticRecord struct {
	ICD10        string `json:"icd10"`
	Category     string `json:"category,omitempty"`
	Kategorie    string `json:"kategorie,omitempty"`
	Title        string `json:"title,omitempty"`
	Titel        string `json:"titel,omitempty"`
	Abbreviation string `json:"abbreviation"`
	Description  string `json:"description,omitempty"`
	Beschreibung string `json:"beschreibung,omitempty"`
}

type DiagnosticData struct {
	De []DiagnosticRecord `json:"de"`
	En []DiagnosticRecord `json:"en"`
}

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
	Zip                  string `json:"zip"`
	TherapyTitle         string `json:"therapy_title"` // New field
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

	// Read diagnostic data
	diagnosticPath := filepath.Join("..", "statics", "json", "diagnostic_std.json")
	diagnosticFile, err := os.ReadFile(diagnosticPath)
	if err != nil {
		log.Fatalf("Failed to read diagnostic file: %v", err)
	}

	var diagnosticData DiagnosticData
	if err := json.Unmarshal(diagnosticFile, &diagnosticData); err != nil {
		log.Fatalf("Failed to parse diagnostic JSON: %v", err)
	}

	// Extract German therapy titles (using German records as they seem more appropriate)
	therapyTitles := make([]string, len(diagnosticData.De))
	for i, record := range diagnosticData.De {
		// Use German title (Titel) if available, otherwise fall back to English title
		if record.Titel != "" {
			therapyTitles[i] = record.Titel
		} else {
			therapyTitles[i] = record.Title
		}
	}

	fmt.Printf("Found %d therapy titles from diagnostic data\n", len(therapyTitles))

	// Read seed data
	seedPath := filepath.Join("..", "seed_app_data.json")
	seedFile, err := os.ReadFile(seedPath)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	var seedData SeedData
	if err := json.Unmarshal(seedFile, &seedData); err != nil {
		log.Fatalf("Failed to parse seed JSON: %v", err)
	}

	// Update clients with therapy titles
	for i := range seedData.Clients {
		// Randomly select a therapy title
		randomTitle := therapyTitles[rand.Intn(len(therapyTitles))]
		seedData.Clients[i].TherapyTitle = randomTitle
	}

	fmt.Printf("Updated %d clients with therapy titles\n", len(seedData.Clients))

	// Write updated seed data back to file
	updatedData, err := json.MarshalIndent(seedData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal updated data: %v", err)
	}

	if err := os.WriteFile(seedPath, updatedData, 0644); err != nil {
		log.Fatalf("Failed to write updated seed file: %v", err)
	}

	fmt.Println("Successfully updated seed_app_data.json with therapy titles!")

	// Show some examples
	fmt.Println("\nExample therapy titles assigned:")
	for i, client := range seedData.Clients[:5] {
		fmt.Printf("  %d. %s %s: %s\n", i+1, client.FirstName, client.LastName, client.TherapyTitle)
	}
}
