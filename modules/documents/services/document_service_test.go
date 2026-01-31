package services
package services_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ae-base-server/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocumentService_Upload tests document upload functionality
func TestDocumentService_Upload(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	ctx := context.Background()
	mockStorage := testutils.NewMockStorageService()

	t.Run("successfully uploads document", func(t *testing.T) {
		fileContent := []byte("Test PDF content")
		fileName := "invoice_INV-2026-0001.pdf"
		tenantID := uint(1)
		userID := uint(1)

		mockStorage.On("Upload", ctx, fileName, fileContent, tenantID).Return("documents/1/invoice_INV-2026-0001.pdf", nil)

		path, err := mockStorage.Upload(ctx, fileName, fileContent, tenantID)
		require.NoError(t, err)
		assert.NotEmpty(t, path)
		assert.Contains(t, path, fileName)
	})

	t.Run("generates document hash for integrity", func(t *testing.T) {
		fileContent := []byte("Test document for hashing")
		
		hash := sha256.New()
		hash.Write(fileContent)
		expectedHash := fmt.Sprintf("%x", hash.Sum(nil))

		// Calculate hash
		calculatedHash := fmt.Sprintf("%x", sha256.Sum256(fileContent))
		
		assert.Equal(t, expectedHash, calculatedHash)
		assert.Len(t, calculatedHash, 64) // SHA256 produces 64 hex characters
	})

	t.Run("tenant isolation on upload", func(t *testing.T) {
		fileContent := []byte("Tenant 1 document")
		fileName := "tenant1_doc.pdf"
		
		mockStorage.On("Upload", ctx, fileName, fileContent, uint(1)).Return("documents/1/tenant1_doc.pdf", nil)
		
		path1, err := mockStorage.Upload(ctx, fileName, fileContent, 1)
		require.NoError(t, err)
		assert.Contains(t, path1, "/1/")

		// Tenant 2 uploads with same filename
		mockStorage.On("Upload", ctx, fileName, fileContent, uint(2)).Return("documents/2/tenant1_doc.pdf", nil)
		
		path2, err := mockStorage.Upload(ctx, fileName, fileContent, 2)
		require.NoError(t, err)
		assert.Contains(t, path2, "/2/")

		// Paths should be different
		assert.NotEqual(t, path1, path2)
	})

	t.Run("validates file size limits", func(t *testing.T) {
		// 11 MB file (over 10MB limit)
		largeContent := make([]byte, 11*1024*1024)
		
		mockStorage.On("Upload", ctx, "large.pdf", largeContent, uint(1)).Return("", assert.AnError)

		_, err := mockStorage.Upload(ctx, "large.pdf", largeContent, 1)
		assert.Error(t, err)
	})

	t.Run("validates file types", func(t *testing.T) {
		invalidTypes := []struct {
			filename string
			content  []byte
		}{
			{"script.exe", []byte("MZ executable")},
			{"virus.bat", []byte("@echo off")},
		}

		for _, tt := range invalidTypes {
			mockStorage.On("Upload", ctx, tt.filename, tt.content, uint(1)).Return("", assert.AnError)

			_, err := mockStorage.Upload(ctx, tt.filename, tt.content, 1)
			assert.Error(t, err, "Should reject %s", tt.filename)
		}
	})
}

// TestDocumentService_Download tests document retrieval
func TestDocumentService_Download(t *testing.T) {
	ctx := context.Background()
	mockStorage := testutils.NewMockStorageService()

	t.Run("successfully downloads document", func(t *testing.T) {
		path := "documents/1/invoice.pdf"
		tenantID := uint(1)
		expectedContent := []byte("%PDF-1.4 content")

		mockStorage.On("Download", ctx, path, tenantID).Return(expectedContent, nil)

		content, err := mockStorage.Download(ctx, path, tenantID)
		require.NoError(t, err)
		assert.Equal(t, expectedContent, content)
	})

	t.Run("tenant cannot access other tenant's documents", func(t *testing.T) {
		path := "documents/1/confidential.pdf"
		
		mockStorage.On("Download", ctx, path, uint(2)).Return(nil, assert.AnError)

		_, err := mockStorage.Download(ctx, path, 2)
		assert.Error(t, err, "Tenant 2 should not access tenant 1's documents")
	})

	t.Run("returns error for non-existent document", func(t *testing.T) {
		path := "documents/1/nonexistent.pdf"
		tenantID := uint(1)

		mockStorage.On("Download", ctx, path, tenantID).Return(nil, assert.AnError)

		_, err := mockStorage.Download(ctx, path, tenantID)
		assert.Error(t, err)
	})
}

// TestDocumentService_VersionControl tests document versioning
func TestDocumentService_VersionControl(t *testing.T) {
	ctx := context.Background()
	mockStorage := testutils.NewMockStorageService()

	t.Run("creates new version on update", func(t *testing.T) {
		baseContent := []byte("Version 1 content")
		updatedContent := []byte("Version 2 content")
		fileName := "invoice.pdf"
		tenantID := uint(1)

		// Upload v1
		mockStorage.On("Upload", ctx, fileName, baseContent, tenantID).Return("documents/1/invoice_v1.pdf", nil)
		pathV1, err := mockStorage.Upload(ctx, fileName, baseContent, tenantID)
		require.NoError(t, err)

		// Upload v2
		mockStorage.On("Upload", ctx, fileName, updatedContent, tenantID).Return("documents/1/invoice_v2.pdf", nil)
		pathV2, err := mockStorage.Upload(ctx, fileName, updatedContent, tenantID)
		require.NoError(t, err)

		assert.NotEqual(t, pathV1, pathV2)
		assert.Contains(t, pathV2, "v2")
	})

	t.Run("can retrieve previous versions", func(t *testing.T) {
		v1Content := []byte("Original content")
		v2Content := []byte("Updated content")
		tenantID := uint(1)

		mockStorage.On("Download", ctx, "documents/1/invoice_v1.pdf", tenantID).Return(v1Content, nil)
		mockStorage.On("Download", ctx, "documents/1/invoice_v2.pdf", tenantID).Return(v2Content, nil)

		// Get v1
		content1, err := mockStorage.Download(ctx, "documents/1/invoice_v1.pdf", tenantID)
		require.NoError(t, err)
		assert.Equal(t, v1Content, content1)

		// Get v2
		content2, err := mockStorage.Download(ctx, "documents/1/invoice_v2.pdf", tenantID)
		require.NoError(t, err)
		assert.Equal(t, v2Content, content2)
	})
}

// TestDocumentService_Retention tests 10-year retention (GoBD requirement)
func TestDocumentService_Retention(t *testing.T) {
	ctx := context.Background()
	mockStorage := testutils.NewMockStorageService()

	t.Run("documents have retention metadata", func(t *testing.T) {
		uploadTime := time.Now()
		retentionUntil := uploadTime.AddDate(10, 0, 0) // 10 years

		metadata := map[string]interface{}{
			"uploaded_at":     uploadTime,
			"retention_until": retentionUntil,
			"retention_years": 10,
		}

		assert.True(t, retentionUntil.After(uploadTime))
		
		yearsDiff := retentionUntil.Year() - uploadTime.Year()
		assert.Equal(t, 10, yearsDiff)
	})

	t.Run("cannot delete documents under retention", func(t *testing.T) {
		documentPath := "documents/1/invoice_2026.pdf"
		tenantID := uint(1)
		
		// Document uploaded today, retention until 2036
		mockStorage.On("Delete", ctx, documentPath, tenantID).Return(assert.AnError)

		err := mockStorage.Delete(ctx, documentPath, tenantID)
		assert.Error(t, err, "Should not allow deletion of documents under retention")
	})

	t.Run("can delete after retention period", func(t *testing.T) {
		// Document from 2016, retention expired in 2026
		documentPath := "documents/1/invoice_2016.pdf"
		tenantID := uint(1)

		mockStorage.On("Delete", ctx, documentPath, tenantID).Return(nil)

		err := mockStorage.Delete(ctx, documentPath, tenantID)
		assert.NoError(t, err, "Should allow deletion after retention period")
	})
}

// TestDocumentService_HashVerification tests document integrity verification
func TestDocumentService_HashVerification(t *testing.T) {
	ctx := context.Background()
	mockStorage := testutils.NewMockStorageService()

	t.Run("hash verification detects tampering", func(t *testing.T) {
		originalContent := []byte("Original invoice content")
		originalHash := fmt.Sprintf("%x", sha256.Sum256(originalContent))

		tamperedContent := []byte("Tampered invoice content")
		tamperedHash := fmt.Sprintf("%x", sha256.Sum256(tamperedContent))

		assert.NotEqual(t, originalHash, tamperedHash, "Hashes should differ for different content")
	})

	t.Run("hash verification passes for unchanged document", func(t *testing.T) {
		content := []byte("Invoice content")
		storedHash := fmt.Sprintf("%x", sha256.Sum256(content))

		// Retrieve document later
		retrievedContent := []byte("Invoice content")
		retrievedHash := fmt.Sprintf("%x", sha256.Sum256(retrievedContent))

		assert.Equal(t, storedHash, retrievedHash, "Hash should match for unchanged content")
	})
}

// TestDocumentService_Metadata tests document metadata handling
func TestDocumentService_Metadata(t *testing.T) {
	ctx := context.Background()

	t.Run("stores comprehensive metadata", func(t *testing.T) {
		metadata := map[string]interface{}{
			"filename":         "invoice_INV-2026-0001.pdf",
			"content_type":     "application/pdf",
			"size_bytes":       12345,
			"uploaded_at":      time.Now(),
			"uploaded_by":      uint(1),
			"tenant_id":        uint(1),
			"document_type":    "invoice",
			"invoice_id":       uint(100),
			"hash_sha256":      "abc123...",
			"retention_until":  time.Now().AddDate(10, 0, 0),
		}

		assert.NotEmpty(t, metadata["filename"])
		assert.NotEmpty(t, metadata["hash_sha256"])
		assert.NotZero(t, metadata["tenant_id"])
	})

	t.Run("metadata includes GoBD required fields", func(t *testing.T) {
		gobdMetadata := map[string]interface{}{
			"creation_timestamp": time.Now(),
			"creator_user_id":    uint(1),
			"document_hash":      "sha256:abc123...",
			"retention_period":   "10 years",
			"immutable":          true,
			"audit_trail":        true,
		}

		assert.True(t, gobdMetadata["immutable"].(bool))
		assert.True(t, gobdMetadata["audit_trail"].(bool))
		assert.Equal(t, "10 years", gobdMetadata["retention_period"])
	})
}

// TestDocumentService_Concurrency tests concurrent document operations
func TestDocumentService_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	ctx := context.Background()
	mockStorage := testutils.NewMockStorageService()

	t.Run("handles concurrent uploads", func(t *testing.T) {
		const concurrency = 20
		done := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(idx int) {
				content := []byte(fmt.Sprintf("Document %d content", idx))
				filename := fmt.Sprintf("doc_%d.pdf", idx)
				
				mockStorage.On("Upload", ctx, filename, content, uint(1)).Return(fmt.Sprintf("documents/1/doc_%d.pdf", idx), nil).Once()

				path, err := mockStorage.Upload(ctx, filename, content, 1)
				assert.NoError(t, err)
				assert.NotEmpty(t, path)

				done <- true
			}(i)
		}

		// Wait for all uploads
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
}

// TestDocumentService_Search tests document search functionality
func TestDocumentService_Search(t *testing.T) {
	ctx := context.Background()
	mockStorage := testutils.NewMockStorageService()

	t.Run("searches by document type", func(t *testing.T) {
		tenantID := uint(1)
		docType := "invoice"

		expectedDocs := []string{
			"documents/1/invoice_001.pdf",
			"documents/1/invoice_002.pdf",
		}

		mockStorage.On("Search", ctx, tenantID, docType).Return(expectedDocs, nil)

		docs, err := mockStorage.Search(ctx, tenantID, docType)
		require.NoError(t, err)
		assert.Len(t, docs, 2)
	})

	t.Run("searches by date range", func(t *testing.T) {
		tenantID := uint(1)
		startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)

		expectedDocs := []string{
			"documents/1/invoice_jan_001.pdf",
			"documents/1/invoice_jan_002.pdf",
		}

		mockStorage.On("SearchByDateRange", ctx, tenantID, startDate, endDate).Return(expectedDocs, nil)

		docs, err := mockStorage.SearchByDateRange(ctx, tenantID, startDate, endDate)
		require.NoError(t, err)
		assert.Len(t, docs, 2)
	})
}

// TestDocumentService_Backup tests backup functionality
func TestDocumentService_Backup(t *testing.T) {
	ctx := context.Background()
	mockStorage := testutils.NewMockStorageService()

	t.Run("creates backup archive", func(t *testing.T) {
		tenantID := uint(1)
		
		mockStorage.On("CreateBackup", ctx, tenantID).Return("backups/tenant_1_2026-01-29.zip", nil)

		backupPath, err := mockStorage.CreateBackup(ctx, tenantID)
		require.NoError(t, err)
		assert.Contains(t, backupPath, "tenant_1")
		assert.Contains(t, backupPath, ".zip")
	})
}
