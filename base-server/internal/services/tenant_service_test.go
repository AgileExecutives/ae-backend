package services

import (
	"context"
	"testing"

	"github.com/ae/base-server/internal/models"
	"github.com/ae/base-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTenantTest(t *testing.T) (*TenantService, *gorm.DB) {
	t.Helper()
	db := testutils.SetupTestDB(t)
	// Tenant is already migrated by testutils.SetupTestDB
	// Pass nil for TenantBucketService – tests use CreateTenantWithoutBucket or
	// verify that the nil-bucket path does not panic.
	svc := NewTenantService(db, nil)
	return svc, db
}

// ─── CreateTenantWithoutBucket ────────────────────────────────────────────────

func TestTenantService_CreateWithoutBucket(t *testing.T) {
	svc, db := setupTenantTest(t)
	defer testutils.CleanupTestDB(db)

	req := models.TenantCreateRequest{
		Name: "Clinic A",
		Slug: "clinic-a",
	}
	tenant, err := svc.CreateTenantWithoutBucket(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, tenant)
	assert.NotZero(t, tenant.ID)
	assert.Equal(t, "Clinic A", tenant.Name)
	assert.Equal(t, "clinic-a", tenant.Slug)
}

func TestTenantService_CreateWithoutBucket_Fields(t *testing.T) {
	svc, db := setupTenantTest(t)
	defer testutils.CleanupTestDB(db)

	req := models.TenantCreateRequest{
		Name:       "Clinic B",
		Slug:       "clinic-b",
		CustomerID: 42,
	}
	tenant, err := svc.CreateTenantWithoutBucket(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, uint(42), tenant.CustomerID)
}

// ─── CreateTenant (nil bucket service) ───────────────────────────────────────

func TestTenantService_Create_NilBucketService(t *testing.T) {
	// When TenantBucketService is nil, CreateTenant should still succeed
	svc, db := setupTenantTest(t)
	defer testutils.CleanupTestDB(db)

	req := models.TenantCreateRequest{Name: "Clinic C", Slug: "clinic-c"}
	tenant, err := svc.CreateTenant(context.Background(), req)
	require.NoError(t, err)
	assert.NotZero(t, tenant.ID)
}

// ─── GetTenant ────────────────────────────────────────────────────────────────

func TestTenantService_GetTenant_Found(t *testing.T) {
	svc, db := setupTenantTest(t)
	defer testutils.CleanupTestDB(db)

	req := models.TenantCreateRequest{Name: "My Tenant", Slug: "my-tenant"}
	created, err := svc.CreateTenantWithoutBucket(context.Background(), req)
	require.NoError(t, err)

	fetched, err := svc.GetTenant(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "My Tenant", fetched.Name)
}

func TestTenantService_GetTenant_NotFound(t *testing.T) {
	svc, db := setupTenantTest(t)
	defer testutils.CleanupTestDB(db)

	_, err := svc.GetTenant(context.Background(), 9999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ─── GetAllTenants ────────────────────────────────────────────────────────────

func TestTenantService_GetAllTenants_Empty(t *testing.T) {
	svc, db := setupTenantTest(t)
	defer testutils.CleanupTestDB(db)

	tenants, err := svc.GetAllTenants(context.Background())
	require.NoError(t, err)
	assert.Empty(t, tenants)
}

func TestTenantService_GetAllTenants_Multiple(t *testing.T) {
	svc, db := setupTenantTest(t)
	defer testutils.CleanupTestDB(db)

	for _, name := range []string{"T1", "T2", "T3"} {
		req := models.TenantCreateRequest{Name: name, Slug: name}
		_, err := svc.CreateTenantWithoutBucket(context.Background(), req)
		require.NoError(t, err)
	}

	tenants, err := svc.GetAllTenants(context.Background())
	require.NoError(t, err)
	assert.Len(t, tenants, 3)
}

// ─── EnsureTenantBucket ───────────────────────────────────────────────────────

func TestTenantService_EnsureTenantBucket_NilService(t *testing.T) {
	svc, db := setupTenantTest(t)
	defer testutils.CleanupTestDB(db)

	// bucket service is nil – should return an error, not panic
	err := svc.EnsureTenantBucket(context.Background(), 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}
