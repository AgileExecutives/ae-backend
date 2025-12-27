package entities

import (
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Document represents a stored document with metadata
type Document struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	TenantID       uint           `gorm:"not null;index:idx_tenant_doc" json:"tenant_id"`
	OrganizationID *uint          `gorm:"index:idx_org_doc" json:"organization_id,omitempty"`
	UserID         uint           `gorm:"not null;index:idx_user_doc" json:"user_id"`

	// Document classification
	DocumentType  string `gorm:"size:50;not null;index" json:"document_type"` // "invoice", "contract", "report"
	ReferenceType string `gorm:"size:50" json:"reference_type"`               // "invoice", "client", "session"
	ReferenceID   *uint  `gorm:"index:idx_reference" json:"reference_id,omitempty"`

	// File information
	FileName      string `gorm:"size:255;not null" json:"file_name"`
	StorageKey    string `gorm:"size:500;not null;uniqueIndex" json:"storage_key"`
	StorageBucket string `gorm:"size:100;not null" json:"storage_bucket"`

	// File metadata
	FileSizeBytes int64  `gorm:"not null" json:"file_size_bytes"`
	ContentType   string `gorm:"size:100;not null" json:"content_type"`
	Checksum      string `gorm:"size:64" json:"checksum"` // SHA256

	// Additional metadata and tags
	Metadata datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	Tags     datatypes.JSON `gorm:"type:jsonb" json:"tags,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for the Document model
func (Document) TableName() string {
	return "documents"
}

// BeforeCreate validates tenant_id before inserting
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.TenantID == 0 {
		return errors.New("tenant_id is required")
	}
	if d.UserID == 0 {
		return errors.New("user_id is required")
	}
	return nil
}

// DocumentResponse represents the API response format
type DocumentResponse struct {
	ID             uint                   `json:"id"`
	TenantID       uint                   `json:"tenant_id"`
	OrganizationID *uint                  `json:"organization_id,omitempty"`
	DocumentType   string                 `json:"document_type"`
	ReferenceType  string                 `json:"reference_type,omitempty"`
	ReferenceID    *uint                  `json:"reference_id,omitempty"`
	FileName       string                 `json:"file_name"`
	StorageBucket  string                 `json:"storage_bucket"`
	FileSizeBytes  int64                  `json:"file_size_bytes"`
	ContentType    string                 `json:"content_type"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ToResponse converts Document to DocumentResponse
func (d *Document) ToResponse() DocumentResponse {
	resp := DocumentResponse{
		ID:             d.ID,
		TenantID:       d.TenantID,
		OrganizationID: d.OrganizationID,
		DocumentType:   d.DocumentType,
		ReferenceType:  d.ReferenceType,
		ReferenceID:    d.ReferenceID,
		FileName:       d.FileName,
		StorageBucket:  d.StorageBucket,
		FileSizeBytes:  d.FileSizeBytes,
		ContentType:    d.ContentType,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}

	// Parse metadata if present
	if len(d.Metadata) > 0 {
		var metadata map[string]interface{}
		_ = d.Metadata.Scan(&metadata)
		resp.Metadata = metadata
	}

	// Parse tags if present
	if len(d.Tags) > 0 {
		var tags []string
		_ = d.Tags.Scan(&tags)
		resp.Tags = tags
	}

	return resp
}

// StoreDocumentRequest represents a request to store a new document
type StoreDocumentRequest struct {
	OrganizationID *uint             `json:"organization_id,omitempty"`
	DocumentType   string            `json:"document_type" binding:"required"`
	ReferenceType  string            `json:"reference_type,omitempty"`
	ReferenceID    *uint             `json:"reference_id,omitempty"`
	FileName       string            `json:"file_name" binding:"required"`
	Content        []byte            `json:"-"` // Binary content, not in JSON
	Bucket         string            `json:"bucket" binding:"required"`
	Path           string            `json:"path" binding:"required"`
	ContentType    string            `json:"content_type"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
}

// ListDocumentsRequest represents pagination and filter parameters
type ListDocumentsRequest struct {
	Page           int    `form:"page" binding:"min=1"`
	Limit          int    `form:"limit" binding:"min=1,max=100"`
	DocumentType   string `form:"document_type"`
	ReferenceType  string `form:"reference_type"`
	ReferenceID    uint   `form:"reference_id"`
	OrganizationID uint   `form:"organization_id"`
}

// DocumentListResponse represents paginated document list
type DocumentListResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    []DocumentResponse `json:"data"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int64              `json:"total"`
}

// DocumentAPIResponse represents single document response
type DocumentAPIResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    DocumentResponse `json:"data"`
}

// DownloadURLResponse represents a document download URL
type DownloadURLResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	DocumentID  uint      `json:"document_id"`
	FileName    string    `json:"file_name"`
	DownloadURL string    `json:"download_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}
