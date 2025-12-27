# Documents Module

Multi-tenant document management system with MinIO storage, Redis caching, and organization-based template management.

## Status: ✅ Priority 1-2 Complete

**Completed:** Foundation, storage layer, entities, CRUD services, API endpoints  
**Next:** Priority 3 - Invoice Number Service

See [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md) for detailed progress.

## Features

### Document Management
- ✅ Multi-tenant document storage with isolation
- ✅ S3-compatible storage via MinIO
- ✅ SHA256 checksum validation
- ✅ Pre-signed download URLs (1-hour expiry)
- ✅ Document metadata and tagging (JSONB)
- ✅ Soft deletes for audit trail
- ✅ Reference linking (orders, customers, invoices)

### Template Management
- ✅ Organization-scoped templates
- ✅ System-level default templates
- ✅ Version control
- ✅ Active/default flags
- ✅ Variable definitions (JSONB)
- ⏳ Template rendering (Priority 5)

### Security
- ✅ Tenant isolation at DB and storage levels
- ✅ Middleware access control
- ✅ Time-limited download URLs
- ✅ Organization-based permissions

## Quick Start

### 1. Start Docker Services
```bash
cd /Users/alex/src/ae/backend/environments/dev
docker-compose up -d
```

### 2. Verify Services
```bash
# PostgreSQL
psql -h localhost -U postgres -d ae_dev

# MinIO Console
open http://localhost:9001  # minioadmin / minioadmin123

# Redis
redis-cli -h localhost -p 6379 -a redis123 ping
```

### 3. Build Module
```bash
cd /Users/alex/src/ae/backend/modules/documents
go build ./...
```

## API Endpoints

### Upload Document
```bash
POST /api/v1/documents
Content-Type: multipart/form-data

file: <binary>
document_type: invoice|template|attachment
reference_type: order|customer|invoice (optional)
reference_id: 123 (optional)
metadata: {"key": "value"} (optional)
tags: ["tag1", "tag2"] (optional)
```

### List Documents
```bash
GET /api/v1/documents?page=1&page_size=20&document_type=invoice
```

### Get Document
```bash
GET /api/v1/documents/:id
```

### Download Document
```bash
GET /api/v1/documents/:id/download
```

### Delete Document
```bash
DELETE /api/v1/documents/:id
```

## Configuration

Environment variables in `/environments/dev/.env`:

```bash
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin123
MINIO_USE_SSL=false
MINIO_BUCKET_DOCUMENTS=documents

REDIS_URL=localhost:6379
REDIS_PASSWORD=redis123
REDIS_DB=0
```

## Architecture

```
modules/documents/
├── entities/           # Database models
├── services/          # Business logic
│   └── storage/      # Storage abstraction
├── handlers/         # HTTP handlers
├── middleware/       # Request middleware
├── routes/           # Route registration
└── module.go         # Module interface
```

## Next Steps

1. **Priority 3**: Invoice Number Service (Redis-backed sequences)
2. **Priority 4**: Template Management (CRUD + rendering)
3. **Priority 5**: PDF Generation (Chromedp integration)

For details, see [IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md).
