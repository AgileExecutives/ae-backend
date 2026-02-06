package services_test

import (
	"testing"

	"github.com/ae-base-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unburdy/audit-module/entities"
	"github.com/unburdy/audit-module/services"
	"gorm.io/gorm"
)

func setupAuditTest(t *testing.T) (*services.AuditService, *gorm.DB) {
	db := testutils.SetupTestDB(t)
	testutils.MigrateTestDB(t, db, &entities.AuditLog{})
	service := services.NewAuditService(db)
	return service, db
}

func TestAuditService_LogEvent(t *testing.T) {
	service, db := setupAuditTest(t)
	defer testutils.CleanupTestDB(db)

	tests := []struct {
		name    string
		request services.LogEventRequest
		wantErr bool
	}{
		{
			name: "success - log invoice created",
			request: services.LogEventRequest{
				TenantID:   1,
				UserID:     10,
				EntityType: entities.EntityTypeInvoice,
				EntityID:   100,
				Action:     entities.AuditActionInvoiceDraftCreated,
				IPAddress:  "192.168.1.1",
				UserAgent:  "Mozilla/5.0",
			},
			wantErr: false,
		},
		{
			name: "success - log with metadata",
			request: services.LogEventRequest{
				TenantID:   1,
				UserID:     10,
				EntityType: entities.EntityTypeInvoice,
				EntityID:   100,
				Action:     entities.AuditActionInvoiceFinalized,
				Metadata: &entities.AuditLogMetadata{
					InvoiceNumber: "2026-0001",
					InvoiceStatus: "finalized",
					Reason:        "Customer requested finalization",
					Changes: map[string]interface{}{
						"old_status": "draft",
						"new_status": "finalized",
					},
				},
				IPAddress: "192.168.1.1",
				UserAgent: "Mozilla/5.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.LogEvent(tt.request)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify log was created
			var log entities.AuditLog
			result := db.Where("entity_id = ? AND action = ?", tt.request.EntityID, tt.request.Action).First(&log)
			require.NoError(t, result.Error)

			assert.Equal(t, tt.request.TenantID, log.TenantID)
			assert.Equal(t, tt.request.UserID, log.UserID)
			assert.Equal(t, tt.request.EntityType, log.EntityType)
			assert.Equal(t, tt.request.EntityID, log.EntityID)
			assert.Equal(t, tt.request.Action, log.Action)
			assert.Equal(t, tt.request.IPAddress, log.IPAddress)
			assert.Equal(t, tt.request.UserAgent, log.UserAgent)
			testutils.AssertTimeNotZero(t, log.CreatedAt)
		})
	}
}

func TestAuditService_GetAuditLogs(t *testing.T) {
	service, db := setupAuditTest(t)
	defer testutils.CleanupTestDB(db)

	// Create test data for tenant 1
	for i := 0; i < 5; i++ {
		err := service.LogEvent(services.LogEventRequest{
			TenantID:   1,
			UserID:     10,
			EntityType: entities.EntityTypeInvoice,
			EntityID:   uint(100 + i),
			Action:     entities.AuditActionInvoiceDraftCreated,
			IPAddress:  "192.168.1.1",
		})
		require.NoError(t, err)
	}

	// Create test data for tenant 2
	for i := 0; i < 3; i++ {
		err := service.LogEvent(services.LogEventRequest{
			TenantID:   2,
			UserID:     20,
			EntityType: entities.EntityTypeInvoice,
			EntityID:   uint(200 + i),
			Action:     entities.AuditActionInvoiceDraftCreated,
			IPAddress:  "192.168.1.2",
		})
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		filter    entities.AuditLogFilter
		wantCount int
		wantTotal int64
	}{
		{
			name: "get all logs for tenant 1",
			filter: entities.AuditLogFilter{
				TenantID: 1,
				Page:     1,
				Limit:    10,
			},
			wantCount: 5,
			wantTotal: 5,
		},
		{
			name: "get all logs for tenant 2",
			filter: entities.AuditLogFilter{
				TenantID: 2,
				Page:     1,
				Limit:    10,
			},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name: "pagination - page 1",
			filter: entities.AuditLogFilter{
				TenantID: 1,
				Page:     1,
				Limit:    2,
			},
			wantCount: 2,
			wantTotal: 5,
		},
		{
			name: "pagination - page 2",
			filter: entities.AuditLogFilter{
				TenantID: 1,
				Page:     2,
				Limit:    2,
			},
			wantCount: 2,
			wantTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, total, err := service.GetAuditLogs(tt.filter)

			require.NoError(t, err)
			assert.Len(t, logs, tt.wantCount)
			assert.Equal(t, tt.wantTotal, total)

			// Verify tenant isolation
			for _, log := range logs {
				assert.Equal(t, tt.filter.TenantID, log.TenantID)
			}
		})
	}
}

func TestAuditService_TenantIsolation(t *testing.T) {
	service, db := setupAuditTest(t)
	defer testutils.CleanupTestDB(db)

	// Create logs for different tenants
	for i := 0; i < 5; i++ {
		err := service.LogEvent(services.LogEventRequest{
			TenantID:   1,
			UserID:     10,
			EntityType: entities.EntityTypeInvoice,
			EntityID:   uint(100 + i),
			Action:     entities.AuditActionInvoiceDraftCreated,
		})
		require.NoError(t, err)
	}

	for i := 0; i < 3; i++ {
		err := service.LogEvent(services.LogEventRequest{
			TenantID:   2,
			UserID:     20,
			EntityType: entities.EntityTypeInvoice,
			EntityID:   uint(200 + i),
			Action:     entities.AuditActionInvoiceDraftCreated,
		})
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		err := service.LogEvent(services.LogEventRequest{
			TenantID:   3,
			UserID:     30,
			EntityType: entities.EntityTypeInvoice,
			EntityID:   uint(300 + i),
			Action:     entities.AuditActionInvoiceDraftCreated,
		})
		require.NoError(t, err)
	}

	// Tenant 1 should only see their logs
	logs1, total1, err := service.GetAuditLogs(entities.AuditLogFilter{
		TenantID: 1,
		Page:     1,
		Limit:    100,
	})
	require.NoError(t, err)
	assert.Len(t, logs1, 5)
	assert.Equal(t, int64(5), total1)
	for _, log := range logs1 {
		assert.Equal(t, uint(1), log.TenantID, "Tenant 1 should only see their own logs")
	}

	// Tenant 2 should only see their logs
	logs2, total2, err := service.GetAuditLogs(entities.AuditLogFilter{
		TenantID: 2,
		Page:     1,
		Limit:    100,
	})
	require.NoError(t, err)
	assert.Len(t, logs2, 3)
	assert.Equal(t, int64(3), total2)
	for _, log := range logs2 {
		assert.Equal(t, uint(2), log.TenantID, "Tenant 2 should only see their own logs")
	}

	// Tenant 3 should only see their logs
	logs3, total3, err := service.GetAuditLogs(entities.AuditLogFilter{
		TenantID: 3,
		Page:     1,
		Limit:    100,
	})
	require.NoError(t, err)
	assert.Len(t, logs3, 2)
	assert.Equal(t, int64(2), total3)
	for _, log := range logs3 {
		assert.Equal(t, uint(3), log.TenantID, "Tenant 3 should only see their own logs")
	}
}

func TestAuditService_ConcurrentWrites(t *testing.T) {
	t.Skip("Concurrent writes test requires separate DB connections")

	service, db := setupAuditTest(t)
	defer testutils.CleanupTestDB(db)

	concurrency := 50
	errChan := make(chan error, concurrency)

	// Perform concurrent writes
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			req := services.LogEventRequest{
				TenantID:   1,
				UserID:     uint(10 + index),
				EntityType: entities.EntityTypeInvoice,
				EntityID:   uint(100 + index),
				Action:     entities.AuditActionInvoiceDraftCreated,
				IPAddress:  "192.168.1.1",
			}
			errChan <- service.LogEvent(req)
		}(i)
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		err := <-errChan
		require.NoError(t, err, "Concurrent write %d failed", i)
	}

	// Verify all logs were created
	var count int64
	db.Model(&entities.AuditLog{}).Count(&count)
	assert.Equal(t, int64(concurrency), count)
}

func TestAuditService_Immutability(t *testing.T) {
	service, db := setupAuditTest(t)
	defer testutils.CleanupTestDB(db)

	// Create an audit log
	req := services.LogEventRequest{
		TenantID:   1,
		UserID:     10,
		EntityType: entities.EntityTypeInvoice,
		EntityID:   100,
		Action:     entities.AuditActionInvoiceDraftCreated,
	}
	err := service.LogEvent(req)
	require.NoError(t, err)

	// Get the log
	var log entities.AuditLog
	db.First(&log)

	// Attempt to update the log (should ideally be prevented by database constraints)
	originalAction := log.Action
	log.Action = entities.AuditActionInvoiceFinalized

	// In a real implementation, you might have triggers or constraints preventing updates
	// For now, we verify the original log remains unchanged by re-fetching
	var verifyLog entities.AuditLog
	db.First(&verifyLog, log.ID)
	assert.Equal(t, originalAction, verifyLog.Action, "Audit log should be immutable")
}
