# Seed Script Update - Documents Module Integration

## Summary

The database seed script has been successfully updated to include comprehensive test data for the Documents Module, including templates, invoice numbers, and actual PDF documents.

## Changes Made

### 1. Updated `seed/seed_database.go`

**Added Imports:**
- `context` package for async operations
- Documents module entities and services
- Redis client for invoice number caching
- MinIO storage client

**New Functions:**

1. **`seedDocumentsData()`** - Main orchestration function that:
   - Initializes MinIO storage and Redis clients
   - Creates template, invoice number, and PDF services
   - Seeds templates, invoice numbers, and sample documents
   - Displays seeding statistics

2. **`seedInvoiceTemplate()`** - Creates a professional German invoice template with:
   - Complete HTML/CSS formatting
   - Variable substitution support
   - Sample data for previewing
   - Professional styling (responsive, print-optimized)

3. **`seedEmailTemplate()`** - Creates a basic email template for client communication

4. **`seedInvoiceNumbers()`** - Generates 5 sequential invoice numbers using:
   - Redis caching for performance
   - PostgreSQL persistence for reliability
   - Format: `INV-YYYY-MM-XXXX`

5. **`seedSampleInvoices()`** - Generates 3 actual PDF invoices:
   - Uses real client data from seeded clients
   - Renders templates with realistic therapy session data
   - Stores PDFs in MinIO document storage
   - Creates database records with metadata

6. **`showDocumentsStatistics()`** - Displays comprehensive statistics:
   - Template counts by type
   - Document counts by type
   - Invoice number tracking

### 2. Updated `seed/README.md`

Enhanced documentation with:
- Prerequisites (PostgreSQL, Redis, MinIO)
- Environment variable configuration
- New seeding output examples
- Instructions for accessing seeded data in MinIO
- Troubleshooting guide

## What Gets Seeded

### Templates (Stored in MinIO `templates` bucket)

1. **Standard Therapy Invoice Template**
   - Type: `invoice`
   - Language: German
   - Features: Professional layout, itemized billing, tax calculations
   - Storage: MinIO with HTML content
   - Variables: 20+ template variables

2. **Standard Email Template**
   - Type: `email`
   - Language: German
   - Features: Header, footer, variable substitution
   - Use case: Client communication

### Invoice Numbers (Tracked in PostgreSQL + Redis)

- 5 sequential invoice numbers
- Format: `INV-2025-12-0001` through `INV-2025-12-0005`
- Logged in `invoice_number_logs` table
- Cached in Redis for performance

### Documents (Stored in MinIO `documents` bucket)

- 3 PDF invoices generated from the template
- Linked to actual seeded clients
- Includes realistic therapy session billing
- Metadata stored in `documents` table:
  - Document type: `invoice`
  - Reference type: `client`
  - File size, checksum, content type

## Technical Details

### MinIO Integration
- **Endpoint**: localhost:9000 (configurable via env)
- **Buckets**: `templates`, `documents`
- **Access**: Private (requires authentication)
- **Organization**: Tenant-isolated storage paths

### Redis Integration
- **Purpose**: Invoice number sequence caching
- **Fallback**: PostgreSQL if Redis unavailable
- **Key format**: `invoice:seq:{tenantID}:{orgID}:{year}:{month}`

### PDF Generation
- **Engine**: chromedp (headless Chrome)
- **Format**: A4, landscape/portrait support
- **Quality**: Print-optimized with embedded styles
- **Features**: Background graphics, custom fonts

## Testing

The seed script compiles successfully and is ready for testing.

### To Test:

```bash
# 1. Ensure services are running
cd /Users/alex/src/ae/backend/environments/dev
docker-compose up -d postgres redis minio

# 2. Run the seed script
cd /Users/alex/src/ae/backend/unburdy_server/seed
go run seed_database.go
```

### Expected Behavior:

1. ✅ Database auto-migration completes
2. ✅ Base data seeded (tenants, users)
3. ✅ Application data seeded (clients, providers)
4. ✅ Templates created in MinIO
5. ✅ Invoice numbers generated
6. ✅ 3 PDF invoices created and stored
7. ✅ Statistics displayed

### Verification:

**Database:**
```sql
-- Check templates
SELECT id, name, template_type, is_active FROM templates;

-- Check invoice numbers
SELECT invoice_number, year, month, sequence FROM invoice_number_logs;

-- Check documents
SELECT id, file_name, document_type, file_size_bytes FROM documents;
```

**MinIO Console:**
1. Open http://localhost:9001
2. Login (minioadmin / minioadmin123)
3. Browse `templates` bucket for HTML templates
4. Browse `documents` bucket for PDF files

**API:**
```bash
# Get authentication token first
TOKEN="your-token-here"

# List all documents
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/documents

# Download a document
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/documents/1/download
```

## Benefits

### For Development
- Realistic test data for all document features
- Pre-configured templates ready for testing
- Sample PDFs to verify rendering pipeline

### For Testing
- Complete invoice workflow coverage
- Template management scenarios
- Document storage and retrieval testing

### For Demos
- Professional-looking invoice PDFs
- German-language templates (market-appropriate)
- Real-world therapy billing examples

## Next Steps

1. **Test the seed script** with all services running
2. **Verify PDFs** are generated correctly in MinIO
3. **Test API endpoints** for template and document management
4. **Consider adding** more template types (contracts, reports, etc.)
5. **Add custom templates** for specific therapy types

## Files Changed

- ✅ `/Users/alex/src/ae/backend/unburdy_server/seed/seed_database.go` - Main seeding logic
- ✅ `/Users/alex/src/ae/backend/unburdy_server/seed/README.md` - Updated documentation

## Notes

- Invoice template is production-ready with professional German formatting
- PDF generation requires Chrome/Chromium (provided by chromedp)
- All data is tenant-isolated for multi-tenancy support
- Invoice numbers are atomic and thread-safe via Redis locking
- Templates support Mustache-style variable substitution

---

**Status**: ✅ Complete and ready for testing
**Date**: December 27, 2025
