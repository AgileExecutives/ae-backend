package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/unburdy/invoice-number-module/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newInvoiceNumberTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&entities.InvoiceNumber{}, &entities.InvoiceNumberLog{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestInvoiceNumberService_GenerateInvoiceNumber_IncrementsAndLogs(t *testing.T) {
	db := newInvoiceNumberTestDB(t)
	svc := NewInvoiceNumberService(db)

	ctx := context.Background()
	tenantID := uint(1)
	orgID := uint(2)

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	cfg := DefaultInvoiceConfig()

	resp1, err := svc.GenerateInvoiceNumber(ctx, tenantID, orgID, cfg)
	if err != nil {
		t.Fatalf("GenerateInvoiceNumber #1: %v", err)
	}
	if resp1.Sequence != 1 {
		t.Fatalf("expected sequence 1, got %d", resp1.Sequence)
	}
	wantPrefix := fmt.Sprintf("%s%s%04d%s%02d%s", cfg.Prefix, cfg.Separator, year, cfg.Separator, month, cfg.Separator)
	if !strings.HasPrefix(resp1.InvoiceNumber, wantPrefix) {
		t.Fatalf("expected invoice number prefix %q, got %q", wantPrefix, resp1.InvoiceNumber)
	}
	if !strings.HasSuffix(resp1.InvoiceNumber, fmt.Sprintf("%0*d", cfg.Padding, 1)) {
		t.Fatalf("expected invoice number to end with %q, got %q", fmt.Sprintf("%0*d", cfg.Padding, 1), resp1.InvoiceNumber)
	}

	resp2, err := svc.GenerateInvoiceNumber(ctx, tenantID, orgID, cfg)
	if err != nil {
		t.Fatalf("GenerateInvoiceNumber #2: %v", err)
	}
	if resp2.Sequence != 2 {
		t.Fatalf("expected sequence 2, got %d", resp2.Sequence)
	}
	if !strings.HasSuffix(resp2.InvoiceNumber, fmt.Sprintf("%0*d", cfg.Padding, 2)) {
		t.Fatalf("expected invoice number to end with %q, got %q", fmt.Sprintf("%0*d", cfg.Padding, 2), resp2.InvoiceNumber)
	}

	seq, err := svc.GetCurrentSequence(ctx, tenantID, orgID, year, month)
	if err != nil {
		t.Fatalf("GetCurrentSequence: %v", err)
	}
	if seq != 2 {
		t.Fatalf("expected stored sequence 2, got %d", seq)
	}

	var logs int64
	if err := db.Model(&entities.InvoiceNumberLog{}).Where("tenant_id = ? AND organization_id = ?", tenantID, orgID).Count(&logs).Error; err != nil {
		t.Fatalf("count logs: %v", err)
	}
	if logs != 2 {
		t.Fatalf("expected 2 logs, got %d", logs)
	}
}
