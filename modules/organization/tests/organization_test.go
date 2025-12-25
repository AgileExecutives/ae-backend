package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unburdy/organization-module/entities"
	"github.com/unburdy/organization-module/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&entities.Organization{})
	assert.NoError(t, err)

	return db
}

func TestCreateOrganization(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewOrganizationService(db)

	req := entities.CreateOrganizationRequest{
		Name:       "Test Organization",
		OwnerName:  "John Doe",
		OwnerTitle: "CEO",
		Email:      "test@example.com",
	}

	org, err := service.CreateOrganization(req, 1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, "Test Organization", org.Name)
	assert.Equal(t, "John Doe", org.OwnerName)
	assert.Equal(t, uint(1), org.TenantID)
	assert.Equal(t, uint(1), org.UserID)
}

func TestGetOrganization(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewOrganizationService(db)

	// Create organization first
	req := entities.CreateOrganizationRequest{
		Name:       "Test Organization",
		OwnerName:  "John Doe",
		OwnerTitle: "CEO",
		Email:      "test@example.com",
	}
	created, err := service.CreateOrganization(req, 1, 1)
	assert.NoError(t, err)

	// Get organization
	org, err := service.GetOrganizationByID(created.ID, 1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, created.ID, org.ID)
	assert.Equal(t, "Test Organization", org.Name)
}

func TestGetAllOrganizations(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewOrganizationService(db)

	// Create multiple organizations
	for i := 1; i <= 3; i++ {
		req := entities.CreateOrganizationRequest{
			Name:       "Test Organization",
			OwnerName:  "John Doe",
			OwnerTitle: "CEO",
		}
		_, err := service.CreateOrganization(req, 1, 1)
		assert.NoError(t, err)
	}

	// Get all organizations
	orgs, total, err := service.GetOrganizations(1, 10, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, orgs, 3)
}

func TestUpdateOrganization(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewOrganizationService(db)

	// Create organization
	req := entities.CreateOrganizationRequest{
		Name:       "Test Organization",
		OwnerName:  "John Doe",
		OwnerTitle: "CEO",
		Email:      "test@example.com",
	}
	created, err := service.CreateOrganization(req, 1, 1)
	assert.NoError(t, err)

	// Update organization
	newName := "Updated Organization"
	updateReq := entities.UpdateOrganizationRequest{
		Name: &newName,
	}
	updated, err := service.UpdateOrganization(created.ID, 1, 1, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Organization", updated.Name)
	assert.Equal(t, "John Doe", updated.OwnerName) // Unchanged
}

func TestDeleteOrganization(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewOrganizationService(db)

	// Create organization
	req := entities.CreateOrganizationRequest{
		Name:       "Test Organization",
		OwnerName:  "John Doe",
		OwnerTitle: "CEO",
		Email:      "test@example.com",
	}
	created, err := service.CreateOrganization(req, 1, 1)
	assert.NoError(t, err)

	// Delete organization
	err = service.DeleteOrganization(created.ID, 1, 1)
	assert.NoError(t, err)

	// Verify organization is deleted
	_, err = service.GetOrganizationByID(created.ID, 1, 1)
	assert.Error(t, err)
}

func TestGetOrganizationsWithPagination(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewOrganizationService(db)

	// Create 15 organizations
	for i := 1; i <= 15; i++ {
		req := entities.CreateOrganizationRequest{
			Name:       "Test Organization",
			OwnerName:  "John Doe",
			OwnerTitle: "CEO",
		}
		_, err := service.CreateOrganization(req, 1, 1)
		assert.NoError(t, err)
	}

	// Get first page (10 items)
	orgs, total, err := service.GetOrganizations(1, 10, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, orgs, 10)

	// Get second page (5 items)
	orgs, total, err = service.GetOrganizations(2, 10, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, orgs, 5)
}

func TestMultiTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewOrganizationService(db)

	// Create organizations for tenant 1
	req1 := entities.CreateOrganizationRequest{
		Name: "Tenant 1 Org",
	}
	_, err := service.CreateOrganization(req1, 1, 1)
	assert.NoError(t, err)

	// Create organizations for tenant 2
	req2 := entities.CreateOrganizationRequest{
		Name: "Tenant 2 Org",
	}
	_, err = service.CreateOrganization(req2, 2, 2)
	assert.NoError(t, err)

	// Tenant 1 should only see their organizations
	orgs, total, err := service.GetOrganizations(1, 10, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, orgs, 1)
	assert.Equal(t, "Tenant 1 Org", orgs[0].Name)

	// Tenant 2 should only see their organizations
	orgs, total, err = service.GetOrganizations(1, 10, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, orgs, 1)
	assert.Equal(t, "Tenant 2 Org", orgs[0].Name)
}
