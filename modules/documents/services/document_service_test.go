package services

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ae/shared-modules/documents/entities"
)

func setupDocumentDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entities.Document{}))
	return db
}

func newDocumentService(db *gorm.DB) *DocumentService {
	return NewDocumentService(db, nil)
}

func createTestDocument(t *testing.T, db *gorm.DB, tenantID uint, opts ...func(*entities.Document)) *entities.Document {
	t.Helper()
	doc := &entities.Document{
		TenantID:      tenantID,
		UserID:        1,
		DocumentType:  "invoice",
		ReferenceType: "invoice",
		FileName:      "test.pdf",
		StorageKey:    generateUniqueKey(t),
		StorageBucket: "test-bucket",
		FileSizeBytes: 1024,
		ContentType:   "application/pdf",
	}
	for _, opt := range opts {
		opt(doc)
	}
	require.NoError(t, db.Create(doc).Error)
	return doc
}

var docKeyCounter int64

func generateUniqueKey(_ *testing.T) string {
	n := atomic.AddInt64(&docKeyCounter, 1)
	return fmt.Sprintf("tenants/1/doc-%d.pdf", n)
}

func TestDocumentService_GetDocument(t *testing.T) {
	db := setupDocumentDB(t)
	svc := newDocumentService(db)
	ctx := context.Background()

	doc := createTestDocument(t, db, 1)

	t.Run("returns document for correct tenant", func(t *testing.T) {
		result, err := svc.GetDocument(ctx, doc.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, doc.ID, result.ID)
		assert.Equal(t, "invoice", result.DocumentType)
	})

	t.Run("returns error for wrong tenant", func(t *testing.T) {
		_, err := svc.GetDocument(ctx, doc.ID, 99)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "document not found")
	})

	t.Run("returns error for non-existent document", func(t *testing.T) {
		_, err := svc.GetDocument(ctx, 9999, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "document not found")
	})
}

func TestDocumentService_ListDocuments(t *testing.T) {
	db := setupDocumentDB(t)
	svc := newDocumentService(db)
	ctx := context.Background()

	// Create documents for tenant 1
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.DocumentType = "invoice"
		d.ReferenceType = "client"
		d.StorageKey = "tenants/1/inv1.pdf"
	})
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.DocumentType = "contract"
		d.ReferenceType = "client"
		d.StorageKey = "tenants/1/contract1.pdf"
	})
	// Document for tenant 2 (should be isolated)
	createTestDocument(t, db, 2, func(d *entities.Document) {
		d.StorageKey = "tenants/2/inv1.pdf"
	})

	t.Run("returns all documents for tenant", func(t *testing.T) {
		docs, total, err := svc.ListDocuments(ctx, 1, entities.ListDocumentsRequest{Page: 1, Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, docs, 2)
	})

	t.Run("filters by document type", func(t *testing.T) {
		docs, total, err := svc.ListDocuments(ctx, 1, entities.ListDocumentsRequest{Page: 1, Limit: 20, DocumentType: "invoice"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "invoice", docs[0].DocumentType)
	})

	t.Run("filters by reference type", func(t *testing.T) {
		_, total, err := svc.ListDocuments(ctx, 1, entities.ListDocumentsRequest{Page: 1, Limit: 20, ReferenceType: "client"})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		docs, total, err := svc.ListDocuments(ctx, 2, entities.ListDocumentsRequest{Page: 1, Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, uint(2), docs[0].TenantID)
	})

	t.Run("pagination works", func(t *testing.T) {
		docs, total, err := svc.ListDocuments(ctx, 1, entities.ListDocumentsRequest{Page: 2, Limit: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, docs, 1)
	})

	t.Run("defaults to page 1 limit 20 when zero", func(t *testing.T) {
		docs, total, err := svc.ListDocuments(ctx, 1, entities.ListDocumentsRequest{})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, docs, 2)
	})

	t.Run("clamps limit over 100", func(t *testing.T) {
		docs, total, err := svc.ListDocuments(ctx, 1, entities.ListDocumentsRequest{Page: 1, Limit: 999})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, docs, 2)
	})
}

func TestDocumentService_GetDocumentsByReference(t *testing.T) {
	db := setupDocumentDB(t)
	svc := newDocumentService(db)
	ctx := context.Background()

	refID := uint(42)
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.ReferenceType = "invoice"
		d.ReferenceID = &refID
		d.StorageKey = "tenants/1/ref-inv1.pdf"
	})
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.ReferenceType = "invoice"
		d.ReferenceID = &refID
		d.StorageKey = "tenants/1/ref-inv2.pdf"
	})
	// Different reference
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.ReferenceType = "session"
		d.StorageKey = "tenants/1/sess1.pdf"
	})

	t.Run("returns documents for matching reference", func(t *testing.T) {
		docs, err := svc.GetDocumentsByReference(ctx, 1, "invoice", refID)
		require.NoError(t, err)
		assert.Len(t, docs, 2)
	})

	t.Run("returns empty for non-matching reference type", func(t *testing.T) {
		docs, err := svc.GetDocumentsByReference(ctx, 1, "contract", refID)
		require.NoError(t, err)
		assert.Empty(t, docs)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		docs, err := svc.GetDocumentsByReference(ctx, 99, "invoice", refID)
		require.NoError(t, err)
		assert.Empty(t, docs)
	})
}

func TestDocumentService_ListDocuments_OrganizationFilter(t *testing.T) {
	db := setupDocumentDB(t)
	svc := newDocumentService(db)
	ctx := context.Background()

	orgID := uint(5)
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.OrganizationID = &orgID
		d.StorageKey = "tenants/1/org5-doc1.pdf"
	})
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.StorageKey = "tenants/1/no-org.pdf"
	})

	docs, total, err := svc.ListDocuments(ctx, 1, entities.ListDocumentsRequest{Page: 1, Limit: 20, OrganizationID: 5})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, &orgID, docs[0].OrganizationID)
}

func TestDocumentService_ListDocuments_ReferenceIDFilter(t *testing.T) {
	db := setupDocumentDB(t)
	svc := newDocumentService(db)
	ctx := context.Background()

	refID := uint(7)
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.ReferenceID = &refID
		d.StorageKey = "tenants/1/refid-doc1.pdf"
	})
	createTestDocument(t, db, 1, func(d *entities.Document) {
		d.StorageKey = "tenants/1/no-refid.pdf"
	})

	_, total, err := svc.ListDocuments(ctx, 1, entities.ListDocumentsRequest{Page: 1, Limit: 20, ReferenceID: 7})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
}

func TestDocumentService_GetDocumentContent_NotFound(t *testing.T) {
	db := setupDocumentDB(t)
	svc := newDocumentService(db)
	ctx := context.Background()

	_, err := svc.GetDocumentContent(ctx, 9999, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
}

func TestDocumentService_GetDownloadURL_NotFound(t *testing.T) {
	db := setupDocumentDB(t)
	svc := newDocumentService(db)
	ctx := context.Background()

	_, err := svc.GetDownloadURL(ctx, 9999, 1, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
}

func TestDocumentService_DeleteDocument_NotFound(t *testing.T) {
	db := setupDocumentDB(t)
	svc := newDocumentService(db)
	ctx := context.Background()

	err := svc.DeleteDocument(ctx, 9999, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
}
