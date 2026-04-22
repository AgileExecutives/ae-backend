package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCostProviderDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entities.CostProvider{}))
	return db
}

func TestCostProviderService_CreateCostProvider(t *testing.T) {
	db := setupCostProviderDB(t)
	svc := NewCostProviderService(db)

	req := entities.CreateCostProviderRequest{
		Organization: "Health Insurance Corp",
		Department:   "Mental Health",
		ContactName:  "Jane Smith",
	}

	cp, err := svc.CreateCostProvider(req, 1)
	require.NoError(t, err)
	assert.NotZero(t, cp.ID)
	assert.Equal(t, "Health Insurance Corp", cp.Organization)
	assert.Equal(t, "Mental Health", cp.Department)
	assert.Equal(t, uint(1), cp.TenantID)
}

func TestCostProviderService_GetCostProviderByID(t *testing.T) {
	db := setupCostProviderDB(t)
	svc := NewCostProviderService(db)

	req := entities.CreateCostProviderRequest{Organization: "Insurer A"}
	created, err := svc.CreateCostProvider(req, 1)
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		cp, err := svc.GetCostProviderByID(created.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, "Insurer A", cp.Organization)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetCostProviderByID(9999, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		_, err := svc.GetCostProviderByID(created.ID, 2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCostProviderService_GetAllCostProviders(t *testing.T) {
	db := setupCostProviderDB(t)
	svc := NewCostProviderService(db)

	for _, name := range []string{"Alpha", "Beta", "Gamma"} {
		_, err := svc.CreateCostProvider(entities.CreateCostProviderRequest{Organization: name}, 1)
		require.NoError(t, err)
	}
	// One for a different tenant
	_, err := svc.CreateCostProvider(entities.CreateCostProviderRequest{Organization: "Other"}, 2)
	require.NoError(t, err)

	t.Run("first page", func(t *testing.T) {
		results, total, err := svc.GetAllCostProviders(1, 10, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := svc.GetAllCostProviders(1, 2, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 2)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		results, total, err := svc.GetAllCostProviders(1, 10, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
	})
}

func TestCostProviderService_UpdateCostProvider(t *testing.T) {
	db := setupCostProviderDB(t)
	svc := NewCostProviderService(db)

	created, err := svc.CreateCostProvider(entities.CreateCostProviderRequest{
		Organization: "Old Corp",
		ContactName:  "Old Name",
	}, 1)
	require.NoError(t, err)

	newOrg := "New Corp"
	newContact := "New Name"

	t.Run("update fields", func(t *testing.T) {
		updated, err := svc.UpdateCostProvider(created.ID, 1, entities.UpdateCostProviderRequest{
			Organization: &newOrg,
			ContactName:  &newContact,
		})
		require.NoError(t, err)
		assert.Equal(t, "New Corp", updated.Organization)
		assert.Equal(t, "New Name", updated.ContactName)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.UpdateCostProvider(9999, 1, entities.UpdateCostProviderRequest{Organization: &newOrg})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		_, err := svc.UpdateCostProvider(created.ID, 2, entities.UpdateCostProviderRequest{Organization: &newOrg})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCostProviderService_DeleteCostProvider(t *testing.T) {
	db := setupCostProviderDB(t)
	svc := NewCostProviderService(db)

	created, err := svc.CreateCostProvider(entities.CreateCostProviderRequest{Organization: "To Delete"}, 1)
	require.NoError(t, err)

	t.Run("delete succeeds", func(t *testing.T) {
		err := svc.DeleteCostProvider(created.ID, 1)
		require.NoError(t, err)
		_, err2 := svc.GetCostProviderByID(created.ID, 1)
		require.Error(t, err2)
	})

	t.Run("not found returns error", func(t *testing.T) {
		err := svc.DeleteCostProvider(9999, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCostProviderService_SearchCostProviders(t *testing.T) {
	db := setupCostProviderDB(t)
	svc := NewCostProviderService(db)

	names := []string{"Acme Insurance", "Beta Health", "Acme Dental"}
	for _, name := range names {
		_, err := svc.CreateCostProvider(entities.CreateCostProviderRequest{Organization: name}, 1)
		require.NoError(t, err)
	}

	t.Run("search by organization", func(t *testing.T) {
		results, total, err := svc.SearchCostProviders("Acme", 1, 10, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
	})

	t.Run("no matches", func(t *testing.T) {
		results, total, err := svc.SearchCostProviders("NONEXISTENT", 1, 10, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, results)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		results, total, err := svc.SearchCostProviders("Acme", 1, 10, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, results)
	})
}

func TestCostProviderService_GetAllCostProvidersForTenant(t *testing.T) {
	db := setupCostProviderDB(t)
	svc := NewCostProviderService(db)

	for _, name := range []string{"Corp A", "Corp B"} {
		_, err := svc.CreateCostProvider(entities.CreateCostProviderRequest{Organization: name}, 5)
		require.NoError(t, err)
	}
	_, err := svc.CreateCostProvider(entities.CreateCostProviderRequest{Organization: "Corp C"}, 6)
	require.NoError(t, err)

	results, err := svc.GetAllCostProvidersForTenant(5)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	results6, err := svc.GetAllCostProvidersForTenant(6)
	require.NoError(t, err)
	assert.Len(t, results6, 1)
}
