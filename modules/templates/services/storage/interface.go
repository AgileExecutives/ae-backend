package storage

import (
	"context"
	"time"
)

// DocumentStorage defines the interface for document storage operations
type DocumentStorage interface {
	// Store saves document bytes and returns storage key
	Store(ctx context.Context, req StoreRequest) (string, error)

	// Retrieve gets document by storage key
	Retrieve(ctx context.Context, bucket, key string) ([]byte, error)

	// GetURL returns a pre-signed URL for direct access
	GetURL(ctx context.Context, bucket, key string, expiresIn time.Duration) (string, error)

	// Delete removes document
	Delete(ctx context.Context, bucket, key string) error

	// List returns documents matching prefix
	List(ctx context.Context, bucket, prefix string) ([]DocumentMeta, error)

	// Exists checks if document exists
	Exists(ctx context.Context, bucket, key string) (bool, error)
}

// StoreRequest represents a document upload request
type StoreRequest struct {
	Bucket      string
	Key         string
	Data        []byte
	ContentType string
	Metadata    map[string]string
	ACL         string // "private", "public-read"
}

// DocumentMeta represents document metadata from storage
type DocumentMeta struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
	ETag         string
	Metadata     map[string]string
}
