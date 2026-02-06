package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/ae-base-server/internal/models"
	"gorm.io/gorm"
)

// TestUser is an alias for models.User
type TestUser = models.User

// TestTenant is an alias for models.Tenant
type TestTenant = models.Tenant

// TestInvoice represents a test invoice for fixtures
type TestInvoice struct {
	ID             uint
	InvoiceNumber  string
	TenantID       uint
	UserID         uint
	OrganizationID uint
	Status         string
	Total          float64
	WithItems      bool
}

// CreateTestTenant creates a test tenant in the database
func CreateTestTenant(t *testing.T, db *gorm.DB, name string) *TestTenant {
	tenant := &TestTenant{
		Name:       name,
		Slug:       fmt.Sprintf("test-%d", time.Now().UnixNano()),
		CustomerID: 1, // Default test customer
	}

	result := db.Create(tenant)
	if result.Error != nil {
		t.Fatalf("Failed to create test tenant: %v", result.Error)
	}

	return tenant
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *gorm.DB, email, passwordHash string, tenantID uint) *TestUser {
	user := &TestUser{
		Email:          email,
		PasswordHash:   passwordHash,
		Username:       email, // Use email as username by default
		TenantID:       tenantID,
		OrganizationID: 1, // Default test organization
		Active:         true,
		FirstName:      "Test",
		LastName:       "User",
		Role:           "user",
	}

	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create test user: %v", result.Error)
	}

	return user
}

// CreateTestInvoiceData generates test invoice data
func CreateTestInvoiceData(tenantID, userID, orgID uint, status string, withItems bool) map[string]interface{} {
	data := map[string]interface{}{
		"tenant_id":       tenantID,
		"user_id":         userID,
		"organization_id": orgID,
		"status":          status,
		"total":           100.00,
		"subtotal":        84.03,
		"vat_total":       15.97,
		"created_at":      time.Now(),
		"updated_at":      time.Now(),
	}

	if status == "finalized" || status == "sent" || status == "paid" {
		now := time.Now()
		data["invoice_number"] = fmt.Sprintf("TEST-%d-%05d", now.Year(), tenantID)
		data["finalized_at"] = now
	}

	return data
}

// SeedMinimalTestData seeds minimal test data for basic tests
func SeedMinimalTestData(t *testing.T, db *gorm.DB) {
	// Create test tenant
	CreateTestTenant(t, db, "Test Tenant")
}

// GenerateTestEmail generates a unique test email
func GenerateTestEmail(prefix string, index int) string {
	return fmt.Sprintf("%s.%d@test.example.com", prefix, index)
}

// GenerateTestInvoiceNumber generates a test invoice number
func GenerateTestInvoiceNumber(year int, sequence int) string {
	return fmt.Sprintf("TEST-%d-%05d", year, sequence)
}

// Ptr returns a pointer to the given value (helper for optional fields)
func Ptr[T any](v T) *T {
	return &v
}

// TimePtr returns a pointer to a time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// NowPtr returns a pointer to the current time
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

// PastTimePtr returns a pointer to a time in the past
func PastTimePtr(duration time.Duration) *time.Time {
	past := time.Now().Add(-duration)
	return &past
}

// FutureTimePtr returns a pointer to a time in the future
func FutureTimePtr(duration time.Duration) *time.Time {
	future := time.Now().Add(duration)
	return &future
}
