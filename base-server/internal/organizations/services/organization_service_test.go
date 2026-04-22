package services

import (
	"context"
	"testing"

	"github.com/ae-base-server/internal/models"
	"github.com/ae-base-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupOrgTest(t *testing.T) (*OrganizationService, *gorm.DB) {
	t.Helper()
	db := testutils.SetupTestDB(t)
	// Organization is already migrated by testutils.SetupTestDB
	svc := NewOrganizationService(db)
	return svc, db
}

func makeCreateOrgReq() models.CreateOrganizationRequest {
	return models.CreateOrganizationRequest{
		Name:      "Acme Therapy",
		OwnerName: "Dr. Smith",
		Email:     "info@acme.example.com",
	}
}

// ─── CreateOrganization ───────────────────────────────────────────────────────

func TestOrganizationService_Create(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	org, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)
	require.NotNil(t, org)
	assert.NotZero(t, org.ID)
	assert.Equal(t, "Acme Therapy", org.Name)
	assert.Equal(t, uint(1), org.TenantID)
}

func TestOrganizationService_Create_AllFields(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	req := models.CreateOrganizationRequest{
		Name:             "Full Org",
		OwnerName:        "Owner",
		OwnerTitle:       "CEO",
		StreetAddress:    "Main St 1",
		Zip:              "12345",
		City:             "Berlin",
		Email:            "ceo@fullorg.de",
		Phone:            "+49 30 12345678",
		TaxID:            "123/456/789",
		TaxUstID:         "DE123456789",
		BankAccountOwner: "Owner",
		BankAccountBank:  "Sparkasse",
		BankAccountBIC:   "BELADEBEXXX",
		BankAccountIBAN:  "DE44 1234 5678 9012 3456 78",
	}
	org, err := svc.CreateOrganization(req, 2)
	require.NoError(t, err)
	assert.Equal(t, "Full Org", org.Name)
	assert.Equal(t, "Berlin", org.City)
	assert.Equal(t, "DE123456789", org.TaxUstID)
}

// ─── GetOrganizationByID ──────────────────────────────────────────────────────

func TestOrganizationService_GetByID_Found(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	created, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)

	fetched, err := svc.GetOrganizationByID(created.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Acme Therapy", fetched.Name)
}

func TestOrganizationService_GetByID_NotFound(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	_, err := svc.GetOrganizationByID(9999, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestOrganizationService_GetByID_WrongTenant(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	created, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)

	// tenant 2 must not see tenant 1's org
	_, err = svc.GetOrganizationByID(created.ID, 2)
	require.Error(t, err)
}

// ─── GetOrganizations ────────────────────────────────────────────────────────

func TestOrganizationService_GetOrganizations_Empty(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	orgs, total, err := svc.GetOrganizations(1, 10, 99)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, orgs)
}

func TestOrganizationService_GetOrganizations_Pagination(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	for i := 0; i < 5; i++ {
		req := makeCreateOrgReq()
		req.Name = "Org " + string(rune('A'+i))
		_, err := svc.CreateOrganization(req, 1)
		require.NoError(t, err)
	}

	// page 1, 3 per page
	page1, total, err := svc.GetOrganizations(1, 3, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, page1, 3)

	// page 2
	page2, _, err := svc.GetOrganizations(2, 3, 1)
	require.NoError(t, err)
	assert.Len(t, page2, 2)
}

func TestOrganizationService_GetOrganizations_TenantIsolation(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	_, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)

	// tenant 2 should see 0
	orgs, total, err := svc.GetOrganizations(1, 10, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, orgs)
}

// ─── UpdateOrganization ───────────────────────────────────────────────────────

func TestOrganizationService_Update_Name(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	org, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)

	newName := "Updated Name"
	updated, err := svc.UpdateOrganization(org.ID, 1, models.UpdateOrganizationRequest{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
}

func TestOrganizationService_Update_MultipleFields(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	org, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)

	city := "Munich"
	zip := "80331"
	phone := "+49 89 12345"
	updated, err := svc.UpdateOrganization(org.ID, 1, models.UpdateOrganizationRequest{
		City:  &city,
		Zip:   &zip,
		Phone: &phone,
	})
	require.NoError(t, err)
	assert.Equal(t, "Munich", updated.City)
	assert.Equal(t, "80331", updated.Zip)
	assert.Equal(t, "+49 89 12345", updated.Phone)
}

func TestOrganizationService_Update_NotFound(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	newName := "X"
	_, err := svc.UpdateOrganization(9999, 1, models.UpdateOrganizationRequest{Name: &newName})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestOrganizationService_Update_WrongTenant(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	org, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)

	newName := "X"
	_, err = svc.UpdateOrganization(org.ID, 2, models.UpdateOrganizationRequest{Name: &newName})
	require.Error(t, err)
}

// ─── DeleteOrganization ───────────────────────────────────────────────────────

func TestOrganizationService_Delete(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	org, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)

	err = svc.DeleteOrganization(org.ID, 1)
	require.NoError(t, err)

	_, err = svc.GetOrganizationByID(org.ID, 1)
	require.Error(t, err)
}

func TestOrganizationService_Delete_NotFound(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	err := svc.DeleteOrganization(9999, 1)
	require.Error(t, err)
}

func TestOrganizationService_Delete_WrongTenant(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)

	org, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)

	err = svc.DeleteOrganization(org.ID, 2)
	require.Error(t, err)
}

// ─── SetTemplateService (smoke) ───────────────────────────────────────────────

func TestOrganizationService_SetTemplateService_Nil(t *testing.T) {
	svc, db := setupOrgTest(t)
	defer testutils.CleanupTestDB(db)
	// should not panic
	svc.SetTemplateService(nil)
	_, err := svc.CreateOrganization(makeCreateOrgReq(), 1)
	require.NoError(t, err)
}

// ensure context import does not go unused when template service is nil
var _ = context.Background
