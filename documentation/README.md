# Backend Documentation

**Last Updated:** January 29, 2026  
**Status:** ✅ Current & Validated

Multi-tenant SaaS platform with modular architecture for document management, invoice management, client management, booking, and audit logging.

> 📖 **Quick Start:** See [../DOCUMENTATION.md](../DOCUMENTATION.md) for complete system documentation

---

## 📚 Active Documentation

### Core System
- **[Architecture.md](Architecture.md)** - System architecture and design principles ✅
- **[DevPrinciples.md](DevPrinciples.md)** - Coding standards and best practices ✅
- **[MODULE_DEVELOPMENT_GUIDE.md](MODULE_DEVELOPMENT_GUIDE.md)** - How to create new modules ✅
- **[IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md)** - Current development status ⚠️ (May need update)

### Invoice System
- **[INVOICING.md](INVOICING.md)** - Complete invoice workflow ✅
- **[INVOICE_CANCELLATION.md](INVOICE_CANCELLATION.md)** - GoBD-compliant cancellation ✅
- **[INVOICE_CANCELLATION_GOBD_EN.md](INVOICE_CANCELLATION_GOBD_EN.md)** - Legal compliance (English) ✅
- **[INVOICE_CANCELLATION_GOBD_DE.md](INVOICE_CANCELLATION_GOBD_DE.md)** - Rechtliche Konformität (Deutsch) ✅
- **[INVOICE_CANCELLATION_QUICK_REF.md](INVOICE_CANCELLATION_QUICK_REF.md)** - Quick reference card ✅
- **[INVOICE_CANCELLATION_SWAGGER.md](INVOICE_CANCELLATION_SWAGGER.md)** - API documentation ✅
- **[INVOICE_CANCELLATION_RELEASE_NOTES.md](INVOICE_CANCELLATION_RELEASE_NOTES.md)** - Release notes ✅
- **[INVOICE_PDF_GENERATION.md](INVOICE_PDF_GENERATION.md)** - PDF generation system ⚠️ (Check current architecture)
- **[INVOICE_VAT_HANDLING.md](INVOICE_VAT_HANDLING.md)** - VAT/tax calculation ✅
- **[INVOICE_MINIO_INTEGRATION.md](INVOICE_MINIO_INTEGRATION.md)** - Document storage ✅
- **[XRECHNUNG_README.md](XRECHNUNG_README.md)** - German e-invoice standard ✅
- **[invoice_implementation.md](invoice_implementation.md)** - Implementation details ⚠️ (May be outdated)

### Features & Systems
- **[AUDIT_TRAIL_README.md](AUDIT_TRAIL_README.md)** - Audit logging system ✅ (Updated Jan 29)
- **[ADVANCED_SETTINGS_SYSTEM.md](ADVANCED_SETTINGS_SYSTEM.md)** - Advanced configuration ✅ (Fixed Jan 29)
- **[TEMPLATE_SYSTEM_ARCHITECTURE.md](TEMPLATE_SYSTEM_ARCHITECTURE.md)** - Document templates ✅
- **[TEMPLATE_MANAGEMENT.md](TEMPLATE_MANAGEMENT.md)** - Template operations ✅
- **[TEMPLATE_CONTRACT_SYSTEM.md](TEMPLATE_CONTRACT_SYSTEM.md)** - Contract-based templates ✅
- **[EXTRA_EFFORTS_DESIGN.md](EXTRA_EFFORTS_DESIGN.md)** - Billing for extra services ✅
- **[UNIFIED_TOKEN_SYSTEM.md](UNIFIED_TOKEN_SYSTEM.md)** - Authentication tokens ✅
- **[PDF_GENERATION.md](PDF_GENERATION.md)** - PDF generation with ChromeDP ✅
- **[CLIENT_REGISTRATION_WORKFLOW.md](CLIENT_REGISTRATION_WORKFLOW.md)** - Client registration ✅

### API & Integration
- **[SWAGGER_DOCUMENTATION.md](SWAGGER_DOCUMENTATION.md)** - OpenAPI/Swagger specs ✅
- **[SWAGGER_IMPLEMENTATION.md](SWAGGER_IMPLEMENTATION.md)** - Swagger setup ✅
- **[FRONTEND_INTEGRATION_GUIDE.md](FRONTEND_INTEGRATION_GUIDE.md)** - Frontend API usage ✅
- **[INVOICE-FRONTEND-IMPLEMENTATION.md](INVOICE-FRONTEND-IMPLEMENTATION.md)** - Invoice frontend guide ✅

### Organizational
- **[Module Lifecycle Phases.md](Module%20Lifecycle%20Phases.md)** - Module initialization phases ✅
- **[New_Template_Module_Req.md](New_Template_Module_Req.md)** - Template module requirements ✅

---

## 📦 Archived Documentation

Historical documentation moved to [archive/](archive/):
- Completed migrations (INVOICE_SCHEMA_MIGRATION, ORGANIZATION_FORMAT_MIGRATION, etc.)
- Old workflows (INVOICE_OLD_WORKFLOW)
- Refactoring plans (TEMPLATE_REFACTORING_PLAN, TEMPLATE_REFACTORING_COMPLETE)
- Phase completion summaries (PHASE_12, PRIORITY_3)
- Resolved issues (MODULE_LIFECYCLE_ISSUES)

These are kept for historical reference but no longer reflect current implementation.

---

## 🚀 Recent Changes (January 2026)

### January 29, 2026
- ✅ **Audit Module** - Moved to shared modules, created independent go.mod
- ✅ **Invoice Module** - Refactored to use PDF module instead of direct ChromeDP
- ✅ **Settings System** - Fixed API mismatches, updated to use JSONB Data field
- ✅ **Email Module** - Simplified to delivery-only (no template rendering)
- ✅ **Documentation** - Consolidated and validated against current implementation

### January 26, 2026
- ✅ **Invoice Cancellation** - GoBD-compliant storno feature with complete documentation

---

## 🔍 Quick Reference

### Starting the System
```bash
# Start services
cd environments/dev && docker-compose up -d

# Run base-server (port 8081)
cd base-server && go run main.go

# Run unburdy-server (port 8080)  
cd unburdy_server && go run main.go
```

### Module Structure
```
modules/your-module/
├── go.mod                    # Module definition
├── module.go                 # Module implementation
├── entities/                 # Data models
├── handlers/                 # HTTP handlers
├── services/                 # Business logic
└── routes/                   # Route definitions
```

### Creating a Module
1. Create directory in `modules/`
2. Initialize with `go mod init github.com/unburdy/your-module`
3. Implement `core.Module` interface
4. Add to application's module registry
5. Generate Swagger docs

### Shared Modules
Current shared modules in `/modules`:
- **audit** - Audit trail logging
- **booking** - Appointment scheduling
- **calendar** - Calendar management
- **documents** - Document storage (MinIO)
- **invoice** - Invoice generation
- **invoice_number** - Sequential numbering

---

## 📝 Documentation Standards

When creating/updating documentation:

1. **Include date** - Last updated timestamp
2. **Validate code** - Check against current implementation
3. **Mark status** - ✅ Current, ⚠️ Needs review, 🗄️ Archived
4. **Use examples** - Real code snippets from the repo
5. **Link related docs** - Cross-reference related documentation

---

## 🆘 Support

- **Questions:** Check [../DOCUMENTATION.md](../DOCUMENTATION.md) first
- **Issues:** GitHub Issues
- **API Docs:** `/swagger/index.html` when server running
- **Module Docs:** Each module has inline documentation
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
