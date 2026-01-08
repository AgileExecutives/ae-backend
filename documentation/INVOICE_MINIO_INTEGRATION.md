# Invoice PDF MinIO Storage Integration

## Overview

Invoice PDFs are automatically generated and stored in MinIO when invoices are finalized or credit notes are created. This provides persistent storage, version control, and secure access via pre-signed URLs.

## Architecture

### Components

1. **InvoicePDFService** - Generates PDFs and manages MinIO storage
2. **DocumentStorage Interface** - Abstraction layer from documents module
3. **MinIOStorage** - Implementation using MinIO/S3 SDK
4. **InvoiceService** - Orchestrates PDF generation during invoice lifecycle

### Service Integration

```
Documents Module
    └── MinIOStorage (exposed as "document-storage")
         ↓ (injected via service registry)
Client Management Module
    └── InvoiceService
         └── InvoicePDFService (initialized with storage)
```

## MinIO Configuration

### Connection Settings

```go
MinIOConfig{
    Endpoint:        "localhost:9000",
    AccessKeyID:     "minioadmin",
    SecretAccessKey: "minioadmin123",
    UseSSL:          false,
    Region:          "us-east-1",
}
```

**Environment Variables:**
- Production settings should use environment variables
- Configure via `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`

### Bucket Structure

**Bucket:** `invoices`

**Folder Structure:**
```
invoices/
├── drafts/
│   └── {tenant_id}/
│       └── {invoice_number}-draft.pdf
├── final/
│   └── {tenant_id}/
│       └── {invoice_number}.pdf
└── credit-notes/
    └── {tenant_id}/
        └── {invoice_number}.pdf
```

**Example Paths:**
- Draft: `invoices/drafts/1/INV-2025-001-draft.pdf`
- Final: `invoices/final/1/INV-2025-001.pdf`
- Credit Note: `invoices/credit-notes/1/CN-2025-001.pdf`

## Storage Methods

### 1. Store Draft PDF

```go
func (s *InvoicePDFService) StoreDraftPDFToMinIO(
    ctx context.Context, 
    invoice *entities.Invoice
) (string, error)
```

**Triggers:**
- When draft invoice is created (optional)
- When draft invoice is updated (optional)

**Storage Key Format:** `invoices/drafts/{tenant_id}/{invoice_number}-draft.pdf`

**Metadata:**
- `invoice_id`: Invoice database ID
- `tenant_id`: Tenant ID for isolation
- `invoice_number`: Invoice number
- `status`: "draft"

**Access Control:** `private`

### 2. Store Final PDF

```go
func (s *InvoicePDFService) StoreFinalPDFToMinIO(
    ctx context.Context, 
    invoice *entities.Invoice
) (string, error)
```

**Triggers:**
- Automatically when invoice is finalized
- Called in `FinalizeInvoice()` service method

**Storage Key Format:** `invoices/final/{tenant_id}/{invoice_number}.pdf`

**Metadata:**
- `invoice_id`: Invoice database ID
- `tenant_id`: Tenant ID for isolation
- `invoice_number`: Invoice number
- `status`: "final"
- `generated_at`: ISO 8601 timestamp

**Access Control:** `private`

**Immutability:** Final PDFs should not be modified after storage

### 3. Store Credit Note PDF

```go
func (s *InvoicePDFService) StoreCreditNotePDFToMinIO(
    ctx context.Context,
    creditNote *entities.Invoice,
    originalInvoiceNumber string,
    reason string
) (string, error)
```

**Triggers:**
- Automatically when credit note is created
- Called in `CreateCreditNote()` service method

**Storage Key Format:** `invoices/credit-notes/{tenant_id}/{invoice_number}.pdf`

**Metadata:**
- `invoice_id`: Credit note database ID
- `tenant_id`: Tenant ID for isolation
- `invoice_number`: Credit note number
- `status`: "credit_note"
- `original_invoice_number`: Reference to original invoice
- `reason`: Credit note reason
- `generated_at`: ISO 8601 timestamp

**Access Control:** `private`

### 4. Get PDF URL

```go
func (s *InvoicePDFService) GetPDFURL(
    ctx context.Context,
    storageKey string,
    expiresIn time.Duration
) (string, error)
```

**Purpose:** Generate pre-signed URL for secure PDF download

**Parameters:**
- `storageKey`: Full storage path (e.g., `invoices/final/1/INV-2025-001.pdf`)
- `expiresIn`: URL expiration duration (e.g., `7 * 24 * time.Hour`)

**Returns:** Pre-signed URL with expiration

**Example:**
```go
url, err := pdfService.GetPDFURL(
    ctx,
    "invoices/final/1/INV-2025-001.pdf",
    7 * 24 * time.Hour, // 7 days
)
```

**URL Format:**
```
http://localhost:9000/invoices/invoices/final/1/INV-2025-001.pdf?
X-Amz-Algorithm=AWS4-HMAC-SHA256&
X-Amz-Credential=minioadmin%2F20260108%2Fus-east-1%2Fs3%2Faws4_request&
X-Amz-Date=20260108T120000Z&
X-Amz-Expires=604800&
X-Amz-SignedHeaders=host&
X-Amz-Signature=...
```

### 5. Delete PDF

```go
func (s *InvoicePDFService) DeletePDF(
    ctx context.Context,
    storageKey string
) error
```

**Use Cases:**
- Remove draft PDFs when invoice is finalized
- Cleanup when invoice is cancelled
- Data retention compliance

## Workflow Integration

### Invoice Finalization

```go
// FinalizeInvoice (invoice_service.go)
func (s *InvoiceService) FinalizeInvoice(invoiceID, tenantID, userID uint) (*entities.Invoice, error) {
    // ... validation and finalization logic ...
    
    // Generate and store final PDF in MinIO
    if s.pdfService != nil {
        ctx := context.Background()
        storageKey, err := s.pdfService.StoreFinalPDFToMinIO(ctx, &invoice)
        if err != nil {
            // Log error but don't fail the finalization
            fmt.Printf("Warning: Failed to store final PDF in MinIO: %v\n", err)
        } else {
            fmt.Printf("Final PDF stored in MinIO: %s\n", storageKey)
        }
    }
    
    return &invoice, nil
}
```

**Key Points:**
- PDF storage is non-blocking
- Errors are logged but don't fail finalization
- Storage key is logged for debugging
- Future: Store storage key in invoice.document_storage_key field

### Credit Note Creation

```go
// CreateCreditNote (invoice_service.go)
func (s *InvoiceService) CreateCreditNote(originalInvoiceID, tenantID, userID uint, req entities.CreateCreditNoteRequest) (*entities.Invoice, error) {
    // ... credit note creation logic ...
    
    // Generate and store credit note PDF in MinIO
    if s.pdfService != nil {
        ctx := context.Background()
        storageKey, err := s.pdfService.StoreCreditNotePDFToMinIO(ctx, &creditNote, originalInvoice.InvoiceNumber, req.Reason)
        if err != nil {
            fmt.Printf("Warning: Failed to store credit note PDF in MinIO: %v\n", err)
        } else {
            fmt.Printf("Credit note PDF stored in MinIO: %s\n", storageKey)
        }
    }
    
    return &creditNote, nil
}
```

## Service Initialization

### Module Setup (module.go)

```go
func (m *CoreModule) Initialize(ctx core.ModuleContext) error {
    // Initialize invoice service
    invoiceService := clientServices.NewInvoiceService(ctx.DB)
    
    // Get document storage from registry
    if docStorageRaw, ok := ctx.Services.Get("document-storage"); ok {
        if docStorage, ok := docStorageRaw.(documentServices.DocumentStorage); ok {
            ctx.Logger.Info("✅ Document storage retrieved from registry")
            // Initialize PDF service and inject into invoice service
            invoiceService.SetDocumentStorage(docStorage)
            ctx.Logger.Info("✅ PDF service initialized with document storage")
        }
    }
    
    return nil
}
```

### Service Registry

**Service Name:** `document-storage`

**Provided By:** Documents Module

**Interface:** `storage.DocumentStorage`

**Factory Method:**
```go
func (p *documentStorageProvider) Factory(ctx core.ModuleContext) (interface{}, error) {
    if p.module.minioStorage == nil {
        minioConfig := storage.MinIOConfig{
            Endpoint:        "localhost:9000",
            AccessKeyID:     "minioadmin",
            SecretAccessKey: "minioadmin123",
            UseSSL:          false,
            Region:          "us-east-1",
        }
        
        minioStorage, err := storage.NewMinIOStorage(minioConfig)
        if err != nil {
            return nil, err
        }
        p.module.minioStorage = minioStorage
    }
    
    return p.module.minioStorage, nil
}
```

## Error Handling

### Storage Failures

**Strategy:** Graceful degradation

```go
storageKey, err := s.pdfService.StoreFinalPDFToMinIO(ctx, &invoice)
if err != nil {
    // Log error but don't fail invoice finalization
    fmt.Printf("Warning: Failed to store final PDF in MinIO: %v\n", err)
} else {
    fmt.Printf("Final PDF stored in MinIO: %s\n", storageKey)
}
```

**Rationale:**
- Invoice data is persisted in database
- PDF can be regenerated if needed
- Storage is supplementary to core invoice functionality

### Common Errors

| Error | Cause | Resolution |
|-------|-------|-----------|
| `storage not configured` | MinIO storage not injected | Verify documents module is loaded |
| `failed to create bucket` | Insufficient permissions | Check MinIO credentials |
| `failed to upload object` | Network/connection issue | Check MinIO service status |
| `failed to generate presigned URL` | Invalid storage key | Verify key format |

## Testing

### Manual Test Steps

1. **Verify MinIO is running:**
```bash
docker ps | grep minio
curl http://localhost:9000/minio/health/live
```

2. **Finalize an invoice:**
```bash
POST /invoices/{id}/finalize
```

3. **Check logs for storage confirmation:**
```
Final PDF stored in MinIO: invoices/final/1/INV-2025-001.pdf
```

4. **Verify in MinIO console:**
- Navigate to http://localhost:9001
- Login with minioadmin/minioadmin123
- Browse to `invoices` bucket
- Verify PDF exists in correct folder

5. **Test PDF download:**
```go
url, _ := pdfService.GetPDFURL(ctx, "invoices/final/1/INV-2025-001.pdf", 7*24*time.Hour)
// Open URL in browser to download PDF
```

### Unit Test Example

```go
func TestStoreFinalPDFToMinIO(t *testing.T) {
    // Setup
    mockStorage := &MockDocumentStorage{}
    pdfService := NewInvoicePDFService(db, mockStorage)
    
    invoice := &entities.Invoice{
        ID:            1,
        TenantID:      1,
        InvoiceNumber: "INV-2025-001",
        Status:        entities.InvoiceStatusSent,
    }
    
    // Execute
    storageKey, err := pdfService.StoreFinalPDFToMinIO(context.Background(), invoice)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, "invoices/final/1/INV-2025-001.pdf", storageKey)
    assert.Equal(t, 1, mockStorage.StoreCallCount)
    assert.Equal(t, "invoices", mockStorage.LastRequest.Bucket)
    assert.Equal(t, "application/pdf", mockStorage.LastRequest.ContentType)
}
```

## Future Enhancements

### 1. Document Storage Key Field

Add to invoices table:
```sql
ALTER TABLE invoices 
  ADD COLUMN document_storage_key VARCHAR(500);
```

Store key after PDF generation:
```go
if err := s.db.Model(&invoice).Update("document_storage_key", storageKey).Error; err != nil {
    return fmt.Errorf("failed to update storage key: %w", err)
}
```

### 2. Version Control

Store multiple versions of PDFs:
```
invoices/final/1/INV-2025-001/v1.pdf
invoices/final/1/INV-2025-001/v2.pdf
invoices/final/1/INV-2025-001/current.pdf
```

### 3. PDF Download Endpoint

```go
// GET /invoices/{id}/pdf
func (h *InvoiceHandler) DownloadPDF(c *gin.Context) {
    // Fetch invoice
    // Get storage key
    // Generate pre-signed URL
    // Redirect or return URL
}
```

### 4. Bulk PDF Export

Generate ZIP archive of multiple invoices:
```go
// GET /invoices/export?ids=1,2,3
func (h *InvoiceHandler) BulkExportPDF(c *gin.Context) {
    // Fetch PDFs from MinIO
    // Create ZIP archive
    // Return download
}
```

### 5. PDF Regeneration

Allow manual PDF regeneration:
```go
// POST /invoices/{id}/regenerate-pdf
func (h *InvoiceHandler) RegeneratePDF(c *gin.Context) {
    // Delete old PDF
    // Generate new PDF
    // Store in MinIO
    // Update storage key
}
```

### 6. Metrics and Monitoring

Track PDF storage metrics:
- Total storage size per tenant
- PDF generation success/failure rate
- Average PDF generation time
- Storage access frequency

## Security Considerations

### 1. Access Control

- All PDFs stored with `private` ACL
- Access only via pre-signed URLs
- URLs expire after 7 days (configurable)

### 2. Tenant Isolation

- Storage keys include tenant_id
- Prevents cross-tenant access
- Query validation ensures tenant ownership

### 3. Data Encryption

**At Rest:**
- MinIO supports server-side encryption (SSE-S3)
- Enable via MinIO configuration

**In Transit:**
- Use SSL for production (`UseSSL: true`)
- Configure with valid certificates

### 4. Audit Trail

Log all storage operations:
```go
ctx.Logger.Info("PDF stored", map[string]interface{}{
    "invoice_id":     invoice.ID,
    "tenant_id":      invoice.TenantID,
    "storage_key":    storageKey,
    "operation":      "store_final_pdf",
    "timestamp":      time.Now(),
})
```

## Performance Optimization

### 1. Async PDF Generation

Generate PDFs asynchronously:
```go
go func() {
    ctx := context.Background()
    _, err := s.pdfService.StoreFinalPDFToMinIO(ctx, &invoice)
    if err != nil {
        log.Printf("Async PDF storage failed: %v", err)
    }
}()
```

### 2. Caching

Cache pre-signed URLs:
```go
cacheKey := fmt.Sprintf("pdf_url:%s", storageKey)
if cachedURL, found := cache.Get(cacheKey); found {
    return cachedURL.(string), nil
}

url, err := s.storage.GetURL(ctx, "invoices", storageKey, expiresIn)
if err == nil {
    cache.Set(cacheKey, url, expiresIn)
}
return url, err
```

### 3. Compression

Compress PDFs before storage:
```go
compressedData, err := compressPDF(pdfData)
storeReq.Data = compressedData
storeReq.Metadata["compressed"] = "true"
```

## Troubleshooting

### MinIO Not Accessible

**Symptoms:**
```
Warning: Failed to store final PDF in MinIO: failed to create MinIO client: connection refused
```

**Resolution:**
1. Check MinIO service: `docker ps | grep minio`
2. Restart MinIO: `docker restart minio`
3. Verify configuration in documents module

### Storage Key Not Found

**Symptoms:**
```
failed to generate PDF URL: The specified key does not exist
```

**Resolution:**
1. Verify storage key format
2. Check MinIO console for file existence
3. Ensure PDF was successfully stored

### Permission Denied

**Symptoms:**
```
failed to upload object: Access Denied
```

**Resolution:**
1. Verify MinIO credentials
2. Check bucket permissions
3. Ensure ACL is correct

---

**Last Updated:** January 8, 2026  
**Version:** 1.0  
**Status:** Production Ready
