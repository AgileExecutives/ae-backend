# Template Management System

## Overview
Comprehensive template management system integrated with base-server email templates, supporting multi-tenant email/PDF templates with MinIO storage, version control, and organization-based fallback.

## ✅ Features Implemented

### Template Types
- **Email Templates**: HTML email templates for transactional emails
- **PDF Templates**: Document generation templates
- **Invoice Templates**: Specialized invoice document templates
- **Document Templates**: General document templates

### Core Capabilities
- ✅ Create, read, update, delete (CRUD) templates
- ✅ Version control (auto-increment on content updates)
- ✅ Organization-specific templates with system-level fallbacks
- ✅ Template content stored in MinIO (S3-compatible)
- ✅ Variable definitions (JSONB array)
- ✅ Sample data for preview (JSONB object)
- ✅ Active/inactive status
- ✅ Default template per type/organization
- ✅ Template preview with sample data
- ✅ Template rendering with custom data
- ✅ Template duplication
- ✅ Tenant isolation

## Architecture

```
┌──────────────────┐
│   HTTP Handler   │ ← Template CRUD, Preview, Render
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Template Service │ ← Business Logic, Version Control
└────────┬─────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌─────────┐ ┌──────────┐
│PostgreSQL│ │  MinIO   │ ← Template Content Storage
│Metadata │ │ (S3)     │
└─────────┘ └──────────┘
```

## Database Schema

### Templates Table
```sql
CREATE TABLE templates (
  id BIGSERIAL PRIMARY KEY,
  tenant_id BIGINT NOT NULL,
  organization_id BIGINT,           -- NULL = system default
  template_type VARCHAR(50) NOT NULL,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  version INT DEFAULT 1,
  is_active BOOLEAN DEFAULT TRUE,
  is_default BOOLEAN DEFAULT FALSE,
  storage_key VARCHAR(500) UNIQUE NOT NULL,
  variables JSONB,                  -- ["recipientName", "orderNumber"]
  sample_data JSONB,                -- {"recipientName": "John", "orderNumber": "123"}
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  deleted_at TIMESTAMP,
  
  INDEX idx_tenant_org (tenant_id, organization_id),
  INDEX idx_template_type (template_type),
  INDEX idx_active_default (is_active, is_default)
);
```

## API Endpoints

### 1. Create Template
**POST** `/api/v1/templates`

Create a new template with content stored in MinIO.

**Request:**
```json
{
  "organization_id": 10,
  "template_type": "email",
  "name": "Welcome Email v2",
  "description": "Updated welcome email template",
  "content": "<!DOCTYPE html><html>...</html>",
  "variables": ["recipientName", "companyName", "loginUrl"],
  "sample_data": {
    "recipientName": "John Doe",
    "companyName": "Acme Corp",
    "loginUrl": "https://app.example.com/login"
  },
  "is_active": true,
  "is_default": false
}
```

**Response:**
```json
{
  "id": 1,
  "tenant_id": 1,
  "organization_id": 10,
  "template_type": "email",
  "name": "Welcome Email v2",
  "description": "Updated welcome email template",
  "version": 1,
  "is_active": true,
  "is_default": false,
  "storage_key": "tenants/1/templates/email/Welcome_Email_v2_1735214400.html",
  "variables": ["recipientName", "companyName", "loginUrl"],
  "sample_data": {...},
  "created_at": "2025-12-26T10:00:00Z"
}
```

### 2. List Templates
**GET** `/api/v1/templates?organization_id=10&template_type=email&is_active=true&page=1&page_size=20`

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "template_type": "email",
      "name": "Welcome Email",
      "version": 3,
      "is_active": true,
      "is_default": true,
      "created_at": "2025-12-26T10:00:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

### 3. Get Template
**GET** `/api/v1/templates/:id`

Retrieve template metadata (without content).

### 4. Get Template with Content
**GET** `/api/v1/templates/:id/content`

**Response:**
```json
{
  "template": {
    "id": 1,
    "name": "Welcome Email",
    "version": 3
  },
  "content": "<!DOCTYPE html><html>...</html>"
}
```

### 5. Update Template
**PUT** `/api/v1/templates/:id`

Updates metadata or content. **Content updates create a new version.**

**Request:**
```json
{
  "name": "Welcome Email - Updated",
  "content": "<!DOCTYPE html><html>...NEW...</html>",
  "is_active": true
}
```

**Response:**
```json
{
  "id": 1,
  "name": "Welcome Email - Updated",
  "version": 4,
  "storage_key": "tenants/1/templates/email/Welcome_Email_v4_1735214500.html"
}
```

### 6. Delete Template
**DELETE** `/api/v1/templates/:id`

Soft deletes the template.

**Response:** `204 No Content`

### 7. Preview Template
**GET** `/api/v1/templates/:id/preview`

Renders template with sample data.

**Response:** `text/html` - Rendered HTML

### 8. Render Template
**POST** `/api/v1/templates/:id/render`

Renders template with custom data.

**Request:**
```json
{
  "recipientName": "Jane Smith",
  "companyName": "Tech Inc",
  "loginUrl": "https://portal.tech.com"
}
```

**Response:** `text/html` - Rendered HTML

### 9. Duplicate Template
**POST** `/api/v1/templates/:id/duplicate`

Creates a copy of an existing template.

**Request:**
```json
{
  "name": "Welcome Email - Copy"
}
```

**Response:**
```json
{
  "id": 2,
  "name": "Welcome Email - Copy",
  "version": 1,
  "is_active": false,
  "is_default": false
}
```

### 10. Get Default Template
**GET** `/api/v1/templates/default?template_type=email&organization_id=10`

Retrieves the default template for a type. Falls back to system default if organization-specific not found.

## Integration with Base-Server Email Service

### Current Email Service (base-server)
```go
// base-server/services/email.go
func (e *EmailService) loadTemplates() {
    templatesDir := utils.GetEnv("EMAIL_TEMPLATES_DIR", "./statics/email_templates")
    templates := map[EmailTemplate]string{
        TemplateVerification:  "verification.html",
        TemplatePasswordReset: "password_reset.html",
        TemplateWelcome:       "welcome.html",
    }
    // Loads from filesystem
}
```

### Integration Strategy

#### Option 1: Database-First (Recommended)
Modify email service to check database templates first, fallback to filesystem.

```go
// In base-server/modules/email/services/email_service.go
func (e *EmailService) SendTemplateEmail(to string, templateType EmailTemplate, data EmailData) error {
    // Try to get template from database
    tmpl, err := e.templateService.GetDefaultTemplate(ctx, tenantID, orgID, string(templateType))
    if err == nil {
        // Render from database template
        html, err := e.templateService.RenderTemplate(ctx, tenantID, tmpl.ID, data)
        if err == nil {
            return e.SendEmail(to, data.Subject, html, textBody)
        }
    }
    
    // Fallback to filesystem template
    return e.sendFileSystemTemplate(to, templateType, data)
}
```

#### Option 2: Migration Tool
Create a tool to import filesystem templates into the database:

```bash
# Import all email templates
./scripts/import-email-templates.sh
```

```go
// cmd/import-templates/main.go
func ImportEmailTemplates(templatesDir string, tenantID uint) error {
    files, _ := ioutil.ReadDir(templatesDir)
    for _, file := range files {
        content, _ := ioutil.ReadFile(filepath.Join(templatesDir, file.Name()))
        
        req := &services.CreateTemplateRequest{
            TenantID:     tenantID,
            TemplateType: "email",
            Name:         strings.TrimSuffix(file.Name(), ".html"),
            Content:      string(content),
            IsActive:     true,
            IsDefault:    false,
        }
        
        templateService.CreateTemplate(ctx, req)
    }
}
```

## Template Variable System

### Defining Variables
```json
{
  "variables": [
    "recipientName",
    "companyName",
    "orderNumber",
    "orderTotal",
    "orderDate"
  ]
}
```

### Using Variables in Templates
```html
<!DOCTYPE html>
<html>
<body>
    <h1>Hello {{.recipientName}}!</h1>
    <p>Your order #{{.orderNumber}} from {{.companyName}}</p>
    <p>Total: ${{.orderTotal}}</p>
    <p>Date: {{.orderDate}}</p>
</body>
</html>
```

### Sample Data for Preview
```json
{
  "sample_data": {
    "recipientName": "John Doe",
    "companyName": "Acme Corp",
    "orderNumber": "ORD-2025-001",
    "orderTotal": "149.99",
    "orderDate": "2025-12-26"
  }
}
```

## Organization Fallback Logic

### Hierarchy
1. **Organization-specific template** (organization_id = 10, is_default = true)
2. **System default template** (organization_id = NULL, is_default = true)
3. **Error: No template found**

### Example
```go
// Get default welcome email for org 10
tmpl, err := templateService.GetDefaultTemplate(ctx, tenantID, &orgID, "email")

// Query order:
// 1. WHERE tenant_id=1 AND organization_id=10 AND template_type='email' AND is_default=true
// 2. WHERE tenant_id=1 AND organization_id IS NULL AND template_type='email' AND is_default=true
```

## Version Control

### Auto-Increment on Content Update
```go
// Version 1: Initial creation
POST /api/v1/templates
{
  "name": "Invoice Template",
  "content": "<html>Version 1</html>"
}
→ version: 1, storage_key: "tenants/1/templates/invoice/Invoice_Template_v1.html"

// Version 2: Update content
PUT /api/v1/templates/1
{
  "content": "<html>Version 2 - Updated</html>"
}
→ version: 2, storage_key: "tenants/1/templates/invoice/Invoice_Template_v2.html"

// Version 3: Another update
PUT /api/v1/templates/1
{
  "content": "<html>Version 3 - New Design</html>"
}
→ version: 3, storage_key: "tenants/1/templates/invoice/Invoice_Template_v3.html"
```

### Storage Keys
```
tenants/{tenant_id}/templates/{type}/{name}_{timestamp}.html        # New template
tenants/{tenant_id}/templates/{type}/{name}_v{version}_{timestamp}.html  # Updated template
```

## Storage Structure (MinIO)

```
templates-bucket/
└── tenants/
    └── {tenant_id}/
        └── templates/
            ├── email/
            │   ├── Welcome_Email_1735214400.html
            │   ├── Welcome_Email_v2_1735214500.html
            │   └── Welcome_Email_v3_1735214600.html
            ├── pdf/
            │   └── Invoice_Template_1735214700.html
            └── invoice/
                └── Standard_Invoice_1735214800.html
```

## Security & Tenant Isolation

### Tenant Scoping
- All queries filtered by `tenant_id`
- Storage keys prefixed with `tenants/{tenant_id}/`
- Middleware enforces tenant access control

### Organization Isolation
- Templates can be organization-specific or system-wide
- System templates (organization_id = NULL) available to all orgs in tenant
- Organization templates override system templates

## Use Cases

### 1. Multi-Brand Email Templates
```
Organization A (Acme Corp):
  - Welcome Email (branded)
  - Invoice Email (custom colors)
  
Organization B (Tech Inc):
  - Welcome Email (different branding)
  - Invoice Email (different colors)

System Defaults:
  - Password Reset (shared)
  - Verification Email (shared)
```

### 2. Dynamic Invoice Generation
```go
// Get invoice template
tmpl, _ := templateService.GetDefaultTemplate(ctx, tenantID, &orgID, "invoice")

// Render with invoice data
html, _ := templateService.RenderTemplate(ctx, tenantID, tmpl.ID, map[string]interface{}{
    "invoiceNumber": "INV-2025-001",
    "customerName": "John Doe",
    "items": []InvoiceItem{...},
    "total": 1499.99,
})

// Generate PDF from rendered HTML
pdfService.GeneratePDFFromHTML(html, "invoice.pdf")
```

### 3. A/B Testing Email Templates
```go
// Create variant A
CreateTemplate({name: "Welcome V1", content: "..."})

// Create variant B  
CreateTemplate({name: "Welcome V2", content: "..."})

// Test and set winner as default
UpdateTemplate(winnerID, {is_default: true})
```

## Module Integration

### Entities Registered
- `entities.NewDocumentEntity()`
- `entities.NewTemplateEntity()` ← **Email/PDF templates**
- `entities.NewInvoiceNumberEntity()`
- `entities.NewInvoiceNumberLogEntity()`

### Routes Registered
- `/api/v1/documents` (DocumentRoutes)
- `/api/v1/invoice-numbers` (InvoiceNumberRoutes)
- `/api/v1/templates` ← **Template management** (TemplateRoutes)

### Services Available
- `DocumentService` - Document storage
- `InvoiceNumberService` - Sequential number generation
- `TemplateService` ← **Template management**

## Testing

### Create Email Template
```bash
curl -X POST http://localhost:8080/api/v1/templates \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "organization_id": 10,
    "template_type": "email",
    "name": "Welcome Email",
    "content": "<!DOCTYPE html><html><body>Hello {{.name}}!</body></html>",
    "variables": ["name"],
    "sample_data": {"name": "John"},
    "is_active": true,
    "is_default": true
  }'
```

### Preview Template
```bash
curl http://localhost:8080/api/v1/templates/1/preview \
  -H "Authorization: Bearer $TOKEN"
```

### Render with Custom Data
```bash
curl -X POST http://localhost:8080/api/v1/templates/1/render \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Smith"}'
```

## Performance

### Caching Strategy (TODO)
- Cache rendered templates in Redis
- Cache key: `template:rendered:{template_id}:{data_hash}`
- TTL: 1 hour
- Invalidate on template update

### Optimization
- Template content lazy-loaded from MinIO
- Metadata queries use indexed fields
- Pagination for large template lists

## Next Steps

1. **Redis Caching**: Cache rendered templates
2. **Template Validation**: Validate HTML/Go template syntax
3. **Migration Tool**: Import filesystem templates to database
4. **Template Scheduler**: Schedule template activation/deactivation
5. **Template Analytics**: Track usage, render times, errors
6. **Template Inheritance**: Base templates with child overrides
7. **Template Marketplace**: Share templates across tenants
8. **Visual Editor**: WYSIWYG template editor UI

## Summary

✅ **Complete CRUD API** for template management  
✅ **Version control** with auto-increment  
✅ **Organization fallback** (org → system defaults)  
✅ **MinIO storage** for template content  
✅ **Preview & render** capabilities  
✅ **Tenant isolation** and security  
✅ **Base-server module** integration  
✅ **Email template** compatibility  

**Total Files**: 18 Go files in documents module  
**New Files**: 4 (template_service.go, template_handler.go, template_routes.go, helpers.go)  
**API Endpoints**: 10 template management endpoints  
**Build Status**: ✅ Success
