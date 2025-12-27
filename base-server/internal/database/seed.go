// Package database provides base-server specific database seeding functionality.
//
// This package contains seeding logic specific to ae-base-server, including
// loading and creating initial tenant, plan, and user data from seed-data.json.
//
// For general database connection utilities, see the pkg/database package.
package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ae-base-server/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedData represents the structure of seed data from JSON
type SeedData struct {
	Customers     []SeedCustomer     `json:"customers"`
	Tenants       []SeedTenant       `json:"tenants"`
	Organizations []SeedOrganization `json:"organizations"`
	Plans         []SeedPlan         `json:"plans"`
	Users         []SeedUser         `json:"users"`
}

// SeedCustomer represents customer seed data
type SeedCustomer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SeedTenant represents tenant seed data
type SeedTenant struct {
	CustomerID uint   `json:"customer_id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
}

// SeedOrganization represents organization seed data
type SeedOrganization struct {
	TenantID uint   `json:"tenant_id"`
	Name     string `json:"name"`
}

// SeedPlan represents plan seed data
type SeedPlan struct {
	Name          string                 `json:"name"`
	Slug          string                 `json:"slug"`
	Description   string                 `json:"description"`
	Price         float64                `json:"price"`
	Currency      string                 `json:"currency"`
	InvoicePeriod string                 `json:"invoice_period"`
	MaxUsers      int                    `json:"max_users"`
	MaxClients    int                    `json:"max_clients"`
	Features      map[string]interface{} `json:"features"`
	Active        bool                   `json:"active"`
}

// SeedUser represents user seed data
type SeedUser struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Role           string `json:"role"`
	Active         bool   `json:"active"`
	TenantSlug     string `json:"tenant_slug"`
	OrganizationID uint   `json:"organization_id"`
	EmailVerified  bool   `json:"email_verified"`
}

// loadSeedData loads seed data from JSON file
func loadSeedData() (*SeedData, error) {
	// Get the current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for seed-data.json in current directory or parent directories
	seedDataPath := filepath.Join(pwd, "seed-data.json")
	if _, err := os.Stat(seedDataPath); os.IsNotExist(err) {
		// Try parent directory (in case running from subdirectory)
		seedDataPath = filepath.Join(filepath.Dir(pwd), "seed-data.json")
		if _, err := os.Stat(seedDataPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("seed-data.json not found in current or parent directory")
		}
	}

	// Read the JSON file
	data, err := os.ReadFile(seedDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed-data.json: %w", err)
	}

	// Parse JSON data
	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("failed to parse seed-data.json: %w", err)
	}

	return &seedData, nil
}

// Seed adds initial data to the database
func Seed(db *gorm.DB) error {
	log.Println("Seeding database with initial data...")

	// Load seed data from JSON file
	seedData, err := loadSeedData()
	if err != nil {
		return fmt.Errorf("failed to load seed data: %w", err)
	}

	// Create customers first
	var customerCount int64
	db.Model(&models.Customer{}).Count(&customerCount)
	if customerCount == 0 {
		for _, customerData := range seedData.Customers {
			customer := models.Customer{
				Name:   customerData.Name,
				Email:  customerData.Email,
				Active: true,
				Status: "active",
			}
			if err := db.Create(&customer).Error; err != nil {
				return fmt.Errorf("failed to create customer %s: %w", customerData.Name, err)
			}
			log.Printf("Created customer: %s", customerData.Name)
		}
	}

	// Create tenants
	var tenantCount int64
	db.Model(&models.Tenant{}).Count(&tenantCount)
	if tenantCount == 0 {
		for _, tenantData := range seedData.Tenants {
			tenant := models.Tenant{
				CustomerID: tenantData.CustomerID,
				Name:       tenantData.Name,
				Slug:       tenantData.Slug,
			}
			if err := db.Create(&tenant).Error; err != nil {
				return fmt.Errorf("failed to create tenant %s: %w", tenantData.Name, err)
			}
			log.Printf("Created tenant: %s", tenantData.Name)
		}
	}

	// Create organizations
	var organizationCount int64
	db.Model(&models.Organization{}).Count(&organizationCount)
	if organizationCount == 0 {
		for _, orgData := range seedData.Organizations {
			organization := models.Organization{
				TenantID: orgData.TenantID,
				Name:     orgData.Name,
			}
			if err := db.Create(&organization).Error; err != nil {
				return fmt.Errorf("failed to create organization %s: %w", orgData.Name, err)
			}
			log.Printf("Created organization: %s", orgData.Name)
		}
	}

	// Create plans
	var planCount int64
	db.Model(&models.Plan{}).Count(&planCount)
	if planCount == 0 {
		for _, planData := range seedData.Plans {
			// Convert features map to JSON string
			featuresJSON, err := json.Marshal(planData.Features)
			if err != nil {
				return fmt.Errorf("failed to marshal features for plan %s: %w", planData.Name, err)
			}

			plan := models.Plan{
				Name:          planData.Name,
				Slug:          planData.Slug,
				Description:   planData.Description,
				Price:         planData.Price,
				Currency:      planData.Currency,
				InvoicePeriod: planData.InvoicePeriod,
				MaxUsers:      planData.MaxUsers,
				MaxClients:    planData.MaxClients,
				Features:      string(featuresJSON),
				Active:        planData.Active,
			}
			if err := db.Create(&plan).Error; err != nil {
				return fmt.Errorf("failed to create plan %s: %w", planData.Name, err)
			}
			log.Printf("Created plan: %s", planData.Name)
		}
	}

	// Create users
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		for _, userData := range seedData.Users {
			// Find the tenant by slug
			var tenant models.Tenant
			if err := db.Where("slug = ?", userData.TenantSlug).First(&tenant).Error; err != nil {
				return fmt.Errorf("failed to find tenant with slug %s: %w", userData.TenantSlug, err)
			}

			// Hash the password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password for user %s: %w", userData.Username, err)
			}

			user := models.User{
				Username:       userData.Username,
				Email:          userData.Email,
				PasswordHash:   string(hashedPassword),
				FirstName:      userData.FirstName,
				LastName:       userData.LastName,
				TenantID:       tenant.ID,
				OrganizationID: userData.OrganizationID,
				Role:           userData.Role,
				Active:         userData.Active,
				EmailVerified:  userData.EmailVerified,
			}

			// Set EmailVerifiedAt if email is verified
			if userData.EmailVerified {
				now := db.NowFunc()
				user.EmailVerifiedAt = &now
			}

			if err := db.Create(&user).Error; err != nil {
				return fmt.Errorf("failed to create user %s: %w", userData.Username, err)
			}
			log.Printf("Created user: %s (email_verified: %v)", userData.Username, userData.EmailVerified)
		}
	}

	log.Println("Database seeding completed successfully! 🎉")
	return nil
}
