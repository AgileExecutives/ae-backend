package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/unburdy-server-api/modules/client_management/entities"
	"github.com/unburdy/unburdy-server-api/modules/client_management/settings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupExtraEffortDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&entities.Client{},
		&entities.CostProvider{},
		&entities.Session{},
		&entities.ExtraEffort{},
		&entities.InvoiceItem{},
	))
	return db
}

func makeExtraEffortReq(clientID uint) *entities.CreateExtraEffortRequest {
	return &entities.CreateExtraEffortRequest{
		ClientID:    clientID,
		EffortType:  "preparation",
		EffortDate:  time.Now().Format("2006-01-02"),
		DurationMin: 30,
		Description: "test effort",
	}
}

func TestExtraEffortService_CreateExtraEffort(t *testing.T) {
	db := setupExtraEffortDB(t)
	svc := NewExtraEffortService(db)

	client := entities.Client{TenantID: 1, FirstName: "X", LastName: "Y"}
	require.NoError(t, db.Create(&client).Error)

	t.Run("creates effort successfully", func(t *testing.T) {
		effort, err := svc.CreateExtraEffort(makeExtraEffortReq(client.ID), 1, 10)
		require.NoError(t, err)
		assert.NotZero(t, effort.ID)
		assert.Equal(t, "unbilled", effort.BillingStatus)
		assert.Equal(t, "preparation", effort.EffortType)
	})

	t.Run("invalid date format returns error", func(t *testing.T) {
		req := makeExtraEffortReq(client.ID)
		req.EffortDate = "not-a-date"
		_, err := svc.CreateExtraEffort(req, 1, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "effort_date")
	})
}

func TestExtraEffortService_GetExtraEffort(t *testing.T) {
	db := setupExtraEffortDB(t)
	svc := NewExtraEffortService(db)

	client := entities.Client{TenantID: 1, FirstName: "A", LastName: "B"}
	require.NoError(t, db.Create(&client).Error)

	created, err := svc.CreateExtraEffort(makeExtraEffortReq(client.ID), 1, 10)
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		e, err := svc.GetExtraEffort(created.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, created.ID, e.ID)
	})

	t.Run("not found returns error", func(t *testing.T) {
		_, err := svc.GetExtraEffort(9999, 1)
		require.Error(t, err)
	})
}

func TestExtraEffortService_ListExtraEfforts(t *testing.T) {
	db := setupExtraEffortDB(t)
	svc := NewExtraEffortService(db)

	client := entities.Client{TenantID: 1, FirstName: "A", LastName: "B"}
	require.NoError(t, db.Create(&client).Error)

	for i := 0; i < 3; i++ {
		_, err := svc.CreateExtraEffort(makeExtraEffortReq(client.ID), 1, 10)
		require.NoError(t, err)
	}

	efforts, count, err := svc.ListExtraEfforts(1, map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	assert.Len(t, efforts, 3)
}

func TestExtraEffortService_UpdateExtraEffort(t *testing.T) {
	db := setupExtraEffortDB(t)
	svc := NewExtraEffortService(db)

	client := entities.Client{TenantID: 1, FirstName: "A", LastName: "B"}
	require.NoError(t, db.Create(&client).Error)

	created, err := svc.CreateExtraEffort(makeExtraEffortReq(client.ID), 1, 10)
	require.NoError(t, err)

	newType := "consultation"
	newDuration := 45
	newDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	t.Run("updates fields successfully", func(t *testing.T) {
		err := svc.UpdateExtraEffort(created.ID, 1, &entities.UpdateExtraEffortRequest{
			EffortType:  &newType,
			DurationMin: &newDuration,
			EffortDate:  &newDate,
		})
		require.NoError(t, err)
		updated, _ := svc.GetExtraEffort(created.ID, 1)
		assert.Equal(t, "consultation", updated.EffortType)
		assert.Equal(t, 45, updated.DurationMin)
	})

	t.Run("invalid date format returns error", func(t *testing.T) {
		bad := "not-a-date"
		err := svc.UpdateExtraEffort(created.ID, 1, &entities.UpdateExtraEffortRequest{EffortDate: &bad})
		require.Error(t, err)
	})

	t.Run("not found returns error", func(t *testing.T) {
		err := svc.UpdateExtraEffort(9999, 1, &entities.UpdateExtraEffortRequest{EffortType: &newType})
		require.Error(t, err)
	})

	t.Run("cannot update billed effort", func(t *testing.T) {
		// Directly set billing status to billed
		db.Model(created).Update("billing_status", "billed")
		err := svc.UpdateExtraEffort(created.ID, 1, &entities.UpdateExtraEffortRequest{EffortType: &newType})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "billed")
	})
}

func TestExtraEffortService_DeleteExtraEffort(t *testing.T) {
	db := setupExtraEffortDB(t)
	svc := NewExtraEffortService(db)

	client := entities.Client{TenantID: 1, FirstName: "A", LastName: "B"}
	require.NoError(t, db.Create(&client).Error)

	t.Run("deletes unbilled effort", func(t *testing.T) {
		created, err := svc.CreateExtraEffort(makeExtraEffortReq(client.ID), 1, 10)
		require.NoError(t, err)
		err = svc.DeleteExtraEffort(created.ID, 1)
		require.NoError(t, err)
		_, err2 := svc.GetExtraEffort(created.ID, 1)
		require.Error(t, err2)
	})

	t.Run("cannot delete billed effort", func(t *testing.T) {
		created, err := svc.CreateExtraEffort(makeExtraEffortReq(client.ID), 1, 10)
		require.NoError(t, err)
		db.Model(created).Update("billing_status", "billed")
		err = svc.DeleteExtraEffort(created.ID, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "billed")
	})

	t.Run("not found returns error", func(t *testing.T) {
		err := svc.DeleteExtraEffort(9999, 1)
		require.Error(t, err)
	})
}

func TestExtraEffortService_GetUnbilledEffortsByClient(t *testing.T) {
	db := setupExtraEffortDB(t)
	svc := NewExtraEffortService(db)

	client := entities.Client{TenantID: 1, FirstName: "A", LastName: "B"}
	require.NoError(t, db.Create(&client).Error)

	for i := 0; i < 3; i++ {
		_, err := svc.CreateExtraEffort(makeExtraEffortReq(client.ID), 1, 10)
		require.NoError(t, err)
	}
	// Mark one as billed
	var efforts []entities.ExtraEffort
	db.Where("client_id = ?", client.ID).Find(&efforts)
	db.Model(&efforts[0]).Update("billing_status", "billed")

	unbilled, err := svc.GetUnbilledEffortsByClient(client.ID, 1)
	require.NoError(t, err)
	assert.Len(t, unbilled, 2)
}

// BillingCalculator tests

func makeModeSettings(mode string) *settings.BillingModeSettings {
	return &settings.BillingModeSettings{
		ExtraEffortsBillingMode: mode,
		ExtraEffortsConfig: settings.ExtraEffortsConfig{
			ModeBThresholdMinutes: 45,
			ModeDPreparationRatio: 0.5,
		},
	}
}

func makeItemsSettings() *settings.InvoiceItemsSettings {
	return &settings.InvoiceItemsSettings{
		SingleUnitText: "1 Einheit",
		DoubleUnitText: "2 Einheiten",
	}
}

func makeTestSession(numberUnits int) *entities.Session {
	return &entities.Session{
		ID:          1,
		NumberUnits: numberUnits,
	}
}

func TestBillingCalculator_CalculateSessionUnits(t *testing.T) {
	t.Run("ignore mode returns session units", func(t *testing.T) {
		calc, err := NewBillingCalculator(makeModeSettings("ignore"), makeItemsSettings())
		require.NoError(t, err)
		result := calc.CalculateSessionUnits(makeTestSession(1), nil)
		assert.Equal(t, float64(1), result.Units)
	})

	t.Run("separate_items mode returns session units", func(t *testing.T) {
		calc, err := NewBillingCalculator(makeModeSettings("separate_items"), makeItemsSettings())
		require.NoError(t, err)
		result := calc.CalculateSessionUnits(makeTestSession(1), nil)
		assert.Equal(t, float64(1), result.Units)
	})

	t.Run("preparation_allowance mode returns session units", func(t *testing.T) {
		calc, err := NewBillingCalculator(makeModeSettings("preparation_allowance"), makeItemsSettings())
		require.NoError(t, err)
		result := calc.CalculateSessionUnits(makeTestSession(2), nil)
		assert.Equal(t, float64(2), result.Units)
	})

	t.Run("unknown mode returns session units as default", func(t *testing.T) {
		calc, err := NewBillingCalculator(makeModeSettings("unknown_mode"), makeItemsSettings())
		require.NoError(t, err)
		result := calc.CalculateSessionUnits(makeTestSession(1), nil)
		assert.Equal(t, float64(1), result.Units)
	})

	t.Run("bundle_double_units with no efforts returns single units", func(t *testing.T) {
		calc, err := NewBillingCalculator(makeModeSettings("bundle_double_units"), makeItemsSettings())
		require.NoError(t, err)
		result := calc.CalculateSessionUnits(makeTestSession(1), []entities.ExtraEffort{})
		assert.Equal(t, float64(1), result.Units)
	})

	t.Run("bundle_double_units with short effort stays single unit", func(t *testing.T) {
		calc, err := NewBillingCalculator(makeModeSettings("bundle_double_units"), makeItemsSettings())
		require.NoError(t, err)
		effort := entities.ExtraEffort{DurationMin: 20, EffortType: "preparation"}
		result := calc.CalculateSessionUnits(makeTestSession(1), []entities.ExtraEffort{effort})
		assert.Equal(t, float64(1), result.Units)
	})

	t.Run("bundle_double_units with long effort becomes double units", func(t *testing.T) {
		calc, err := NewBillingCalculator(makeModeSettings("bundle_double_units"), makeItemsSettings())
		require.NoError(t, err)
		session := &entities.Session{ID: 1, NumberUnits: 1, DurationMin: 50}
		effort := entities.ExtraEffort{DurationMin: 50, EffortType: "preparation"}
		result := calc.CalculateSessionUnits(session, []entities.ExtraEffort{effort})
		assert.Equal(t, float64(2), result.Units)
	})
}

func TestBillingCalculator_CalculateSeparateEffortUnits(t *testing.T) {
	calc, err := NewBillingCalculator(makeModeSettings("separate_items"), makeItemsSettings())
	require.NoError(t, err)

	t.Run("calculates units from duration", func(t *testing.T) {
		effort := &entities.ExtraEffort{DurationMin: 45, EffortType: "preparation"}
		result := calc.CalculateSeparateEffortUnits(effort)
		assert.NotNil(t, result)
		assert.Positive(t, result.Units)
	})
}

func TestBillingCalculator_CalculatePreparationAllowance(t *testing.T) {
	calc, err := NewBillingCalculator(makeModeSettings("preparation_allowance"), makeItemsSettings())
	require.NoError(t, err)

	result := calc.CalculatePreparationAllowance(2)
	assert.NotNil(t, result)
}
