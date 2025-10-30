package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
)

type SeedData struct {
	CostProviders []map[string]interface{} `json:"cost_providers"`
	Clients       []map[string]interface{} `json:"clients"`
}

// Parent names for generating different parent combinations
var motherNames = []string{"Sarah", "Jennifer", "Amanda", "Patricia", "Lisa", "Michelle", "Elena", "Maria", "Katherine", "Rebecca", "Nicole", "Jessica", "Rachel", "Stephanie", "Angela", "Laura", "Melissa", "Christina", "Sandra", "Ashley"}
var fatherNames = []string{"Michael", "David", "Robert", "Steven", "James", "Daniel", "Carlos", "Luis", "Christopher", "Matthew", "Andrew", "Kevin", "Ryan", "Jason", "Brandon", "Tyler", "Eric", "Justin", "Scott", "Gregory"}

// Email domains for students
var emailDomains = []string{"student.edu", "school.org", "academy.edu", "highschool.edu"}

func main() {
	// Read the current seed data
	data, err := ioutil.ReadFile("seed_app_data.json")
	if err != nil {
		log.Fatalf("Failed to read seed data: %v", err)
	}

	var seedData SeedData
	err = json.Unmarshal(data, &seedData)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Update all clients
	for i := range seedData.Clients {
		client := seedData.Clients[i]

		// Generate age between 8-17 years old
		age := rand.Intn(10) + 8 // 8 to 17 years old
		birthYear := 2025 - age
		birthMonth := rand.Intn(12) + 1
		birthDay := rand.Intn(28) + 1

		// Create birth date
		birthDate := fmt.Sprintf("%d-%02d-%02dT00:00:00Z", birthYear, birthMonth, birthDay)

		// Get first and last name
		firstName := client["first_name"].(string)
		lastName := client["last_name"].(string)

		// Generate email and phone
		domain := emailDomains[rand.Intn(len(emailDomains))]
		email := fmt.Sprintf("%s.%s@%s",
			firstName,
			lastName,
			domain)
		phone := fmt.Sprintf("+1-555-%04d", rand.Intn(10000))

		// Generate parent names (different from each other and from client)
		motherName := motherNames[rand.Intn(len(motherNames))]
		fatherName := fatherNames[rand.Intn(len(fatherNames))]

		// Generate parent emails and phones
		motherEmail := fmt.Sprintf("%s.%s@email.com",
			motherName,
			lastName)
		motherPhone := fmt.Sprintf("+1-555-%04d", rand.Intn(10000))

		fatherEmail := fmt.Sprintf("%s.%s@work.com",
			fatherName,
			lastName)
		fatherPhone := fmt.Sprintf("+1-555-%04d", rand.Intn(10000))

		// Update client data
		client["email"] = email
		client["phone"] = phone
		client["date_of_birth"] = birthDate

		// Update parent information
		client["contact_first_name"] = motherName
		client["contact_last_name"] = lastName
		client["contact_email"] = motherEmail
		client["contact_phone"] = motherPhone

		client["alternative_first_name"] = fatherName
		client["alternative_last_name"] = lastName
		client["alternative_phone"] = fatherPhone
		client["alternative_email"] = fatherEmail

		// Clean up fields that don't match the model
		delete(client, "state")
		delete(client, "postal_code")
		delete(client, "country")
		delete(client, "insurance_id")
		delete(client, "session_rate")
		delete(client, "therapy_type")
		delete(client, "diagnosis")
		delete(client, "treatment_goals")

		// Ensure zip field exists (some might be postal_code)
		if _, exists := client["zip"]; !exists {
			if postalCode, exists := client["postal_code"]; exists {
				client["zip"] = postalCode
				delete(client, "postal_code")
			}
		}

		// Update notes to reflect minor status and parents
		notes := fmt.Sprintf("%d-year-old student. Parents: %s (mother) and %s (father)",
			age, motherName, fatherName)

		// Keep some of the original notes context if it exists
		if originalNotes, exists := client["notes"].(string); exists && originalNotes != "" {
			// Extract relevant parts of original notes
			if len(originalNotes) > 50 {
				notes += ". " + originalNotes[len(originalNotes)-50:]
			} else {
				notes += ". " + originalNotes
			}
		}

		client["notes"] = notes

		fmt.Printf("Updated client %s %s (age %d)\n", firstName, lastName, age)
	}

	// Write the updated data back
	updatedData, err := json.MarshalIndent(seedData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	err = ioutil.WriteFile("seed_app_data.json", updatedData, 0644)
	if err != nil {
		log.Fatalf("Failed to write updated seed data: %v", err)
	}

	fmt.Printf("Successfully updated %d clients\n", len(seedData.Clients))
}
