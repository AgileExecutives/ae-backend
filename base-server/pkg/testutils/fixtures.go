package testutils

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TestUser represents a test user for fixtures
type TestUser struct {
	ID       uint
	Email    string
	Password string
	TenantID uint
}

// TestTenant represents a test tenant for fixtures
type TestTenant struct {
	ID   uint
	Name string
}

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
func CreateTestTenant(db *gorm.DB, name string) (*TestTenant, error) {
	tenant := &TestTenant{
		Name: name,
	}

	// Note: This is a simplified version. In real usage, you'd use your actual Tenant entity
	result := db.Table("tenants").Create(map[string]interface{}{
		"name":       name,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return nil, result.Error
	}

	// Get the ID
	var id uint
	db.Raw("SELECT id FROM tenants WHERE name = ? ORDER BY id DESC LIMIT 1", name).Scan(&id)
	tenant.ID = id

	return tenant, nil
}

// CreateTestUser creates a test user in the database
func CreateTestUser(db *gorm.DB, email string, tenantID uint) (*TestUser, error) {
	user := &TestUser{
		Email:    email,
		Password: "hashed_password_123",
		TenantID: tenantID,
	}

	// Note: This is a simplified version. In real usage, you'd use your actual User entity
	result := db.Table("users").Create(map[string]interface{}{
		"email":      email,
		"password":   "hashed_password_123",
		"tenant_id":  tenantID,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return nil, result.Error
	}

	// Get the ID
	var id uint
	db.Raw("SELECT id FROM users WHERE email = ? ORDER BY id DESC LIMIT 1", email).Scan(&id)
	user.ID = id

	return user, nil
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
func SeedMinimalTestData(db *gorm.DB) error {
	// Create test tenant
	_, err := CreateTestTenant(db, "Test Tenant")
	if err != nil {
		return fmt.Errorf("failed to create test tenant: %w", err)
	}

	return nil
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

// GenerateTestInvoiceNumber generates a test invoice number
// Format: YEAR-TENANT-SEQUENCE (e.g., "2024-1-0042")
func GenerateTestInvoiceNumber(year, tenantID int) string {
	sequence := time.Now().Nanosecond() % 10000
	return fmt.Sprintf("%d-%d-%04d", year, tenantID, sequence)
}
