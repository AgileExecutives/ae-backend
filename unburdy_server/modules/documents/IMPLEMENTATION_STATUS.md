# Documents Module - Implementation Status

## Overview
Multi-tenant document management module with MinIO storage, Redis caching, and PostgreSQL persistence.

## ✅ Completed (Priority 1 & 2)

### Infrastructure (Priority 1)
- [x] Docker services configured (PostgreSQL, MinIO, Redis)
- [x] Environment variables in `.env`
- [x] Development README with service documentation
- [x] Module structure created
- [x] Go module dependencies resolved

### Storage Layer (Priority 1)
- [x] `DocumentStorage` interface abstraction
- [x] MinIO S3-compatible implementation
  - Bucket auto-creation
  - Pre-signed URLs (1-hour default)
  - Error handling with MinIO error responses
  - Tenant-scoped storage keys: `tenants/{tenant_id}/documents/{filename}`

### Database Entities (Priority 1)
- [x] **Document entity** with:
  - Tenant isolation (indexed, required)
  - Organization association (optional)
  - Document metadata (type, reference, size, content type, checksum)
  - JSONB metadata and tags
  - Soft deletes
  - Validation hooks
  
- [x] **Template entity** with:
  - Organization-scoped templates (NULL = system default)
  - Version control
  - Active/default flags
  - JSONB variables and sample data
  - Template type classification

### Business Logic (Priority 2)
- [x] **DocumentService** with full CRUD:
  - `StoreDocument()` - Upload with SHA256 checksum
  - `GetDocument()` - Retrieve metadata
  - `GetDocumentContent()` - Get raw bytes
  - `GetDownloadURL()` - Generate pre-signed URL
  - `ListDocuments()` - Paginated with filters
  - `DeleteDocument()` - Soft delete
  - `GetDocumentsByReference()` - Query by reference type/ID

### API Layer (Priority 2)
- [x] **Document handlers** (5 REST endpoints):
  - `POST /documents` - Upload multipart file
  - `GET /documents` - List with pagination
  - `GET /documents/:id` - Get metadata
  - `GET /documents/:id/download` - Get download URL
  - `DELETE /documents/:id` - Soft delete

- [x] **Middleware**:
  - `EnsureTenantAccess()` - Verify document ownership
  - `EnsureTenantTemplateAccess()` - Verify template ownership

### Module Integration (Priority 2)
- [x] `CoreModule` implementing `core.Module` interface
- [x] Route provider for automatic registration
- [x] Lifecycle methods (Initialize, Start, Stop)
- [x] Entity registration for migrations
- [x] Service initialization with MinIO config

## 📊 Module Structure

```
modules/documents/
├── go.mod                          # Dependencies
├── module.go                       # Core module implementation
├── README.md                       # Module documentation
├── test-integration.sh             # Integration test script
├── entities/
│   ├── entities.go                # Entity registration
│   ├── document.go                # Document entity
│   └── template.go                # Template entity
├── services/
│   ├── document_service.go        # CRUD business logic
│   └── storage/
│       ├── interface.go           # Storage abstraction
│       └── minio_storage.go       # MinIO implementation
├── handlers/
│   └── document_handler.go        # HTTP request handlers
├── middleware/
│   └── tenant_isolation.go        # Access control
└── routes/
    └── routes.go                  # Route provider
```

## 🔧 Configuration

### Environment Variables (.env)
```bash
# MinIO Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
MINIO_REGION=us-east-1
MINIO_BUCKET_DOCUMENTS=documents
MINIO_BUCKET_TEMPLATES=templates
MINIO_BUCKET_INVOICES=invoices

# Redis Configuration
REDIS_URL=localhost:6379
REDIS_PASSWORD=redis123
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10
```

### Docker Services
- **PostgreSQL**: `localhost:5432` (user: postgres, pass: postgres, db: ae_dev)
- **MinIO**: `localhost:9000` (API), `localhost:9001` (Console)
- **Redis**: `localhost:6379` (pass: redis123)

## 🚀 API Endpoints

### Document Management

#### Upload Document
```http
POST /api/v1/documents
Content-Type: multipart/form-data

file: <binary>
document_type: invoice|template|attachment
reference_type: order|customer|invoice (optional)
reference_id: 123 (optional)
metadata: {"key": "value"} (optional)
tags: ["tag1", "tag2"] (optional)
```

**Response:**
```json
{
  "id": 1,
  "tenant_id": 1,
  "document_type": "invoice",
  "filename": "invoice_001.pdf",
  "content_type": "application/pdf",
  "file_size_bytes": 45678,
  "checksum": "sha256:abc123...",
  "metadata": {"key": "value"},
  "tags": ["tag1", "tag2"],
  "created_at": "2025-01-29T10:00:00Z"
}
```

#### List Documents
```http
GET /api/v1/documents?page=1&page_size=20&document_type=invoice&reference_type=order&reference_id=123
```

#### Get Document Metadata
```http
GET /api/v1/documents/:id
```

#### Get Download URL
```http
GET /api/v1/documents/:id/download
```

**Response:**
```json
{
  "download_url": "https://minio:9000/documents/...",
  "expires_at": "2025-01-29T11:00:00Z"
}
```

#### Delete Document
```http
DELETE /api/v1/documents/:id
```

## 🔒 Security Features

1. **Tenant Isolation**
   - All queries filtered by `tenant_id`
   - Storage keys prefixed with `tenants/{tenant_id}/`
   - Middleware enforces access control

2. **Data Integrity**
   - SHA256 checksums for uploaded files
   - Database rollback on storage failure
   - Soft deletes preserve audit trail

3. **Access Control**
   - Middleware verifies ownership before access
   - Pre-signed URLs for time-limited downloads
   - Organization-scoped templates

## 📋 Next Steps (Priorities 3-5)

### Priority 3: Invoice Number Service (Day 5-6)
- [ ] Redis-backed invoice number sequence
- [ ] PostgreSQL persistence for audit trail
- [ ] Concurrent request handling
- [ ] Number format configuration (prefix, padding)
- [ ] Reset on year/month boundaries

### Priority 4: Template Management (Day 7-8)
- [ ] Template CRUD endpoints
- [ ] Version management
- [ ] Organization fallback logic
- [ ] Template variable validation
- [ ] Default template selection

### Priority 5: Invoice PDF Generation (Day 9-11)
- [ ] Chromedp-based HTML to PDF
- [ ] Template rendering with variables
- [ ] Invoice metadata integration
- [ ] Automatic document storage
- [ ] Error handling and retries

## 🧪 Testing

Run integration tests:
```bash
./test-integration.sh
```

Manual endpoint testing (requires server running):
```bash
# Upload document
curl -X POST http://localhost:8080/api/v1/documents \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@invoice.pdf" \
  -F "document_type=invoice" \
  -F "reference_type=order" \
  -F "reference_id=123"

# List documents
curl http://localhost:8080/api/v1/documents \
  -H "Authorization: Bearer $TOKEN"

# Get download URL
curl http://localhost:8080/api/v1/documents/1/download \
  -H "Authorization: Bearer $TOKEN"
```

## 📦 Dependencies

- `github.com/minio/minio-go/v7` - S3-compatible storage
- `github.com/redis/go-redis/v9` - Redis client (for invoice numbers)
- `gorm.io/gorm` - ORM and migrations
- `gorm.io/datatypes` - JSONB support
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/ae-base-server/pkg/core` - Base module interface

## 🐛 Known Issues & Fixes

### Fixed Issues
1. ✅ **Duplicate package declarations** - Files were corrupted by formatter
   - Solution: Recreated files cleanly

2. ✅ **JSON parsing errors** - `datatypes.JSON.AssignTo()` doesn't exist
   - Solution: Changed to `Scan()` method

3. ✅ **Module interface errors** - `core.RouteHandler` vs `core.RouteProvider`
   - Solution: Updated to correct `RouteProvider` interface

4. ✅ **JSON value assignment** - Type assertion needed for `driver.Value`
   - Solution: Used `MarshalJSON()` to get byte slices

## 📝 Notes

- All storage operations are scoped to tenants
- Pre-signed URLs expire after 1 hour (configurable)
- Soft deletes preserve documents for audit
- Templates support organization-level and system-level defaults
- SHA256 checksums prevent duplicate uploads (if implemented)
- Module follows ae-base-server plugin architecture
