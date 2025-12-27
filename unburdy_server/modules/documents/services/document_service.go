package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/unburdy/unburdy-server-api/modules/documents/entities"
	"github.com/unburdy/unburdy-server-api/modules/documents/services/storage"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DocumentService handles business logic for document management
type DocumentService struct {
	db      *gorm.DB
	storage storage.DocumentStorage
}

// NewDocumentService creates a new document service
func NewDocumentService(db *gorm.DB, storage storage.DocumentStorage) *DocumentService {
	return &DocumentService{
		db:      db,
		storage: storage,
	}
}

// StoreDocument uploads a document and creates metadata record
func (s *DocumentService) StoreDocument(ctx context.Context, tenantID, userID uint, req entities.StoreDocumentRequest) (*entities.Document, error) {
	// Calculate checksum
	hash := sha256.Sum256(req.Content)
	checksum := hex.EncodeToString(hash[:])

	// Generate storage key with tenant isolation
	storageKey := fmt.Sprintf("tenants/%d/%s", tenantID, req.Path)

	// Upload to storage
	_, err := s.storage.Store(ctx, storage.StoreRequest{
		Bucket:      req.Bucket,
		Key:         storageKey,
		Data:        req.Content,
		ContentType: req.ContentType,
		Metadata:    req.Metadata,
		ACL:         "private",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	// Prepare metadata
	var metadata datatypes.JSON
	if req.Metadata != nil {
		metadataBytes, _ := datatypes.NewJSONType(req.Metadata).MarshalJSON()
		metadata = metadataBytes
	}

	var tags datatypes.JSON
	if req.Tags != nil {
		tagsBytes, _ := datatypes.NewJSONType(req.Tags).MarshalJSON()
		tags = tagsBytes
	}

	// Create document record
	document := &entities.Document{
		TenantID:       tenantID,
		OrganizationID: req.OrganizationID,
		UserID:         userID,
		DocumentType:   req.DocumentType,
		ReferenceType:  req.ReferenceType,
		ReferenceID:    req.ReferenceID,
		FileName:       req.FileName,
		StorageKey:     storageKey,
		StorageBucket:  req.Bucket,
		FileSizeBytes:  int64(len(req.Content)),
		ContentType:    req.ContentType,
		Checksum:       checksum,
		Metadata:       metadata,
		Tags:           tags,
	}

	if err := s.db.Create(document).Error; err != nil {
		// Rollback: delete from storage
		s.storage.Delete(ctx, req.Bucket, storageKey)
		return nil, fmt.Errorf("failed to create document record: %w", err)
	}

	return document, nil
}

// GetDocument retrieves document metadata by ID (tenant-scoped)
func (s *DocumentService) GetDocument(ctx context.Context, documentID, tenantID uint) (*entities.Document, error) {
	var doc entities.Document
	err := s.db.Where("id = ? AND tenant_id = ?", documentID, tenantID).First(&doc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to fetch document: %w", err)
	}

	return &doc, nil
}

// GetDocumentContent retrieves document bytes from storage
func (s *DocumentService) GetDocumentContent(ctx context.Context, documentID, tenantID uint) ([]byte, error) {
	// Verify tenant owns document
	doc, err := s.GetDocument(ctx, documentID, tenantID)
	if err != nil {
		return nil, err
	}

	// Fetch from storage
	content, err := s.storage.Retrieve(ctx, doc.StorageBucket, doc.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve document content: %w", err)
	}

	return content, nil
}

// GetDownloadURL generates a pre-signed URL for document download
func (s *DocumentService) GetDownloadURL(ctx context.Context, documentID, tenantID uint, expiresIn time.Duration) (string, error) {
	// Verify tenant owns document
	doc, err := s.GetDocument(ctx, documentID, tenantID)
	if err != nil {
		return "", err
	}

	// Generate pre-signed URL
	url, err := s.storage.GetURL(ctx, doc.StorageBucket, doc.StorageKey, expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

// ListDocuments retrieves paginated list of documents (tenant-scoped)
func (s *DocumentService) ListDocuments(ctx context.Context, tenantID uint, req entities.ListDocumentsRequest) ([]entities.Document, int64, error) {
	var documents []entities.Document
	var total int64

	// Build query with tenant isolation
	query := s.db.Where("tenant_id = ?", tenantID)

	// Apply filters
	if req.DocumentType != "" {
		query = query.Where("document_type = ?", req.DocumentType)
	}
	if req.ReferenceType != "" {
		query = query.Where("reference_type = ?", req.ReferenceType)
	}
	if req.ReferenceID > 0 {
		query = query.Where("reference_id = ?", req.ReferenceID)
	}
	if req.OrganizationID > 0 {
		query = query.Where("organization_id = ?", req.OrganizationID)
	}

	// Count total
	if err := query.Model(&entities.Document{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	offset := (req.Page - 1) * req.Limit

	// Fetch documents
	err := query.
		Order("created_at DESC").
		Limit(req.Limit).
		Offset(offset).
		Find(&documents).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch documents: %w", err)
	}

	return documents, total, nil
}

// DeleteDocument soft-deletes document metadata and removes from storage
func (s *DocumentService) DeleteDocument(ctx context.Context, documentID, tenantID uint) error {
	// Verify tenant owns document
	doc, err := s.GetDocument(ctx, documentID, tenantID)
	if err != nil {
		return err
	}

	// Delete from storage first
	if err := s.storage.Delete(ctx, doc.StorageBucket, doc.StorageKey); err != nil {
		// Log but don't fail if storage deletion fails
		fmt.Printf("Warning: failed to delete document from storage: %v\n", err)
	}

	// Soft delete from database
	if err := s.db.Delete(doc).Error; err != nil {
		return fmt.Errorf("failed to delete document record: %w", err)
	}

	return nil
}

// GetDocumentsByReference retrieves all documents for a specific reference
func (s *DocumentService) GetDocumentsByReference(ctx context.Context, tenantID uint, referenceType string, referenceID uint) ([]entities.Document, error) {
	var documents []entities.Document

	err := s.db.Where("tenant_id = ? AND reference_type = ? AND reference_id = ?",
		tenantID, referenceType, referenceID).
		Order("created_at DESC").
		Find(&documents).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch documents by reference: %w", err)
	}

	return documents, nil
}
