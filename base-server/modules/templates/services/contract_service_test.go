package services

import (
	"context"
	"testing"

	"github.com/ae-base-server/modules/templates/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupContractTestDB creates a SQLite DB with both TemplateContract and Template tables.
func setupContractTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entities.TemplateContract{}, &entities.Template{}))
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	})
	return db
}

func registerContractHelper(t *testing.T, svc *ContractService, module, key string) *entities.TemplateContract {
	t.Helper()
	contract, err := svc.RegisterContract(context.Background(), &entities.RegisterContractRequest{
		Module:            module,
		TemplateKey:       key,
		Description:       "Test contract",
		SupportedChannels: []string{"EMAIL", "DOCUMENT"},
		VariableSchema:    map[string]interface{}{"name": map[string]interface{}{"type": "string"}},
		DefaultSampleData: map[string]interface{}{"name": "Test User"},
	})
	require.NoError(t, err)
	return contract
}

func TestContractService_RegisterContract(t *testing.T) {
	db := setupContractTestDB(t)
	svc := NewContractService(db)
	ctx := context.Background()

	t.Run("registers new contract", func(t *testing.T) {
		contract, err := svc.RegisterContract(ctx, &entities.RegisterContractRequest{
			Module:            "email",
			TemplateKey:       "welcome",
			SupportedChannels: []string{"EMAIL"},
		})
		require.NoError(t, err)
		assert.Equal(t, "email", contract.Module)
		assert.Equal(t, "welcome", contract.TemplateKey)
	})

	t.Run("re-registration updates existing contract", func(t *testing.T) {
		registerContractHelper(t, svc, "email", "invoice")
		// Register again with a different description
		contract, err := svc.RegisterContract(ctx, &entities.RegisterContractRequest{
			Module:            "email",
			TemplateKey:       "invoice",
			Description:       "Updated description",
			SupportedChannels: []string{"EMAIL"},
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated description", contract.Description)
	})
}

func TestContractService_GetContract(t *testing.T) {
	db := setupContractTestDB(t)
	svc := NewContractService(db)
	ctx := context.Background()

	registerContractHelper(t, svc, "billing", "receipt")

	t.Run("gets existing contract", func(t *testing.T) {
		contract, err := svc.GetContract(ctx, "billing", "receipt")
		require.NoError(t, err)
		assert.Equal(t, "billing", contract.Module)
		assert.Equal(t, "receipt", contract.TemplateKey)
	})

	t.Run("returns error for non-existent contract", func(t *testing.T) {
		_, err := svc.GetContract(ctx, "billing", "non-existent")
		require.Error(t, err)
	})
}

func TestContractService_GetContractByID(t *testing.T) {
	db := setupContractTestDB(t)
	svc := NewContractService(db)
	ctx := context.Background()

	contract := registerContractHelper(t, svc, "notification", "reminder")

	t.Run("gets contract by ID", func(t *testing.T) {
		got, err := svc.GetContractByID(ctx, contract.ID)
		require.NoError(t, err)
		assert.Equal(t, contract.ID, got.ID)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := svc.GetContractByID(ctx, 9999)
		require.Error(t, err)
	})
}

func TestContractService_ListContracts(t *testing.T) {
	db := setupContractTestDB(t)
	svc := NewContractService(db)
	ctx := context.Background()

	t.Run("returns empty list for new module", func(t *testing.T) {
		contracts, err := svc.ListContracts(ctx, "empty-module")
		require.NoError(t, err)
		assert.Empty(t, contracts)
	})

	registerContractHelper(t, svc, "report", "monthly")
	registerContractHelper(t, svc, "report", "weekly")
	registerContractHelper(t, svc, "other", "daily")

	t.Run("lists contracts for specific module", func(t *testing.T) {
		contracts, err := svc.ListContracts(ctx, "report")
		require.NoError(t, err)
		assert.Len(t, contracts, 2)
	})

	t.Run("lists all contracts when module is empty", func(t *testing.T) {
		contracts, err := svc.ListContracts(ctx, "")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(contracts), 3)
	})
}

func TestContractService_ValidateChannel(t *testing.T) {
	db := setupContractTestDB(t)
	svc := NewContractService(db)
	ctx := context.Background()

	t.Run("non-existent contract returns error", func(t *testing.T) {
		err := svc.ValidateChannel(ctx, "no-mod", "no-key", "EMAIL")
		require.Error(t, err)
	})
}

func TestContractService_UpdateContract(t *testing.T) {
	db := setupContractTestDB(t)
	svc := NewContractService(db)
	ctx := context.Background()

	contract := registerContractHelper(t, svc, "doc", "cover")

	t.Run("updates description", func(t *testing.T) {
		desc := "New description"
		updated, err := svc.UpdateContract(ctx, contract.ID, &entities.UpdateContractRequest{
			Description: &desc,
		})
		require.NoError(t, err)
		assert.Equal(t, "New description", updated.Description)
	})

	t.Run("updates supported channels", func(t *testing.T) {
		channels := []string{"DOCUMENT"}
		updated, err := svc.UpdateContract(ctx, contract.ID, &entities.UpdateContractRequest{
			SupportedChannels: &channels,
		})
		require.NoError(t, err)
		assert.NotNil(t, updated)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		desc := "X"
		_, err := svc.UpdateContract(ctx, 9999, &entities.UpdateContractRequest{Description: &desc})
		require.Error(t, err)
	})
}

func TestContractService_DeleteContract(t *testing.T) {
	db := setupContractTestDB(t)
	svc := NewContractService(db)
	ctx := context.Background()

	t.Run("deletes existing contract", func(t *testing.T) {
		contract := registerContractHelper(t, svc, "del", "target")
		err := svc.DeleteContract(ctx, contract.ID)
		require.NoError(t, err)
		_, err = svc.GetContractByID(ctx, contract.ID)
		require.Error(t, err)
	})

	t.Run("silently succeeds for non-existent ID", func(t *testing.T) {
		err := svc.DeleteContract(ctx, 9999)
		assert.NoError(t, err) // GORM Delete with missing ID doesn't return an error
	})
}
