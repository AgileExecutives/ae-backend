# Swagger API Documentation - Documents Module

## Overview
Complete Swagger/OpenAPI documentation for all 23 API endpoints in the documents module.

## Authentication
All endpoints require Bearer token authentication:
```yaml
securitySchemes:
  BearerAuth:
    type: http
    scheme: bearer
    bearerFormat: JWT
```

---

## Document Management API

### 1. Upload Document
```
POST /api/v1/documents
```

**Summary**: Upload a document  
**ID**: `uploadDocument`  
**Content-Type**: `multipart/form-data`

**Parameters**:
- `file` (formData, file, required): Document file to upload
- `document_type` (formData, string, required): Document type (invoice, contract, report, etc.)
- `bucket` (formData, string): Storage bucket (default: documents)
- `path` (formData, string): Storage path (default: filename)
- `reference_type` (formData, string): Reference type (invoice, client, session)
- `reference_id` (formData, int): Reference ID
- `organization_id` (formData, int): Organization ID

**Response 201** (`entities.DocumentResponse`):
```json
{
  "success": true,
  "message": "Document uploaded successfully",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "organization_id": 10,
    "user_id": 5,
    "document_type": "invoice",
    "reference_type": "invoice",
    "reference_id": 123,
    "file_name": "invoice-001.pdf",
    "file_size_bytes": 102400,
    "content_type": "application/pdf",
    "storage_key": "tenants/1/documents/invoice/invoice-001.pdf",
    "created_at": "2025-12-26T10:00:00Z"
  }
}
```

**Errors**: 400 (Bad Request), 401 (Unauthorized), 500 (Internal Server Error)

---

### 2. List Documents
```
GET /api/v1/documents
```

**Summary**: List documents  
**ID**: `listDocuments`

**Query Parameters**:
- `document_type` (string): Filter by document type
- `reference_type` (string): Filter by reference type
- `reference_id` (int): Filter by reference ID
- `organization_id` (int): Filter by organization ID
- `page` (int): Page number (default: 1)
- `page_size` (int): Page size (default: 20, max: 100)

**Response 200**:
```json
{
  "success": true,
  "data": [{...}],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

---

### 3. Get Document
```
GET /api/v1/documents/{id}
```

**Summary**: Get document metadata  
**ID**: `getDocument`

**Path Parameters**:
- `id` (int, required): Document ID

**Response 200** (`entities.DocumentResponse`)

**Errors**: 400, 401, 404, 500

---

### 4. Download Document
```
GET /api/v1/documents/{id}/download
```

**Summary**: Download document  
**ID**: `downloadDocument`

**Path Parameters**:
- `id` (int, required): Document ID

**Response 200**:
```json
{
  "success": true,
  "message": "Download URL generated successfully",
  "document_id": 1,
  "file_name": "invoice-001.pdf",
  "download_url": "https://minio.example.com/...",
  "expires_at": "2025-12-26T11:00:00Z"
}
```

---

### 5. Delete Document
```
DELETE /api/v1/documents/{id}
```

**Summary**: Delete document  
**ID**: `deleteDocument`

**Path Parameters**:
- `id` (int, required): Document ID

**Response 200**:
```json
{
  "success": true,
  "message": "Document deleted successfully"
}
```

---

## Invoice Number API

### 6. Generate Invoice Number
```
POST /api/v1/invoice-numbers/generate
```

**Summary**: Generate next invoice number  
**ID**: `generateInvoiceNumber`

**Request Body** (`handlers.GenerateInvoiceNumberRequest`):
```json
{
  "organization_id": 10,
  "prefix": "INV",
  "year_format": "YYYY",
  "month_format": "MM",
  "padding": 4,
  "separator": "-",
  "reset_monthly": false
}
```

**Response 200** (`handlers.InvoiceNumberResponse`):
```json
{
  "success": true,
  "invoice_number": "INV-2025-0001",
  "sequence": 1,
  "year": 2025,
  "month": 12
}
```

---

### 7. Get Current Sequence
```
GET /api/v1/invoice-numbers/current
```

**Summary**: Get current sequence  
**ID**: `getCurrentInvoiceSequence`

**Query Parameters**:
- `organization_id` (int, required): Organization ID
- `year` (int): Year (defaults to current year)
- `month` (int): Month (defaults to current month)

**Response 200**:
```json
{
  "success": true,
  "organization_id": 10,
  "year": 2025,
  "month": 12,
  "current_sequence": 5
}
```

---

### 8. Get Invoice Number History
```
GET /api/v1/invoice-numbers/history
```

**Summary**: Get invoice number history  
**ID**: `getInvoiceNumberHistory`

**Query Parameters**:
- `organization_id` (int, required): Organization ID
- `year` (int): Filter by year
- `month` (int): Filter by month
- `page` (int): Page number (default: 1)
- `page_size` (int): Page size (default: 20)

**Response 200**:
```json
{
  "success": true,
  "data": [{
    "id": 1,
    "invoice_number": "INV-2025-0001",
    "action": "generated",
    "created_at": "2025-12-26T10:00:00Z"
  }],
  "total": 10,
  "page": 1,
  "page_size": 20
}
```

---

### 9. Void Invoice Number
```
POST /api/v1/invoice-numbers/void
```

**Summary**: Void invoice number  
**ID**: `voidInvoiceNumber`

**Request Body**:
```json
{
  "invoice_number": "INV-2025-0001"
}
```

**Response 200**:
```json
{
  "success": true,
  "message": "Invoice number voided successfully"
}
```

---

## Template Management API

### 10. Create Template
```
POST /api/v1/templates
```

**Summary**: Create template  
**ID**: `createTemplate`

**Request Body** (`handlers.CreateTemplateRequest`):
```json
{
  "organization_id": 10,
  "template_type": "invoice",
  "name": "Standard Invoice",
  "description": "Default invoice template",
  "content": "<!DOCTYPE html>...",
  "variables": ["invoice_number", "customer_name"],
  "sample_data": {"invoice_number": "INV-001"},
  "is_active": true,
  "is_default": false
}
```

**Response 201** (`handlers.TemplateResponse`)

---

### 11. List Templates
```
GET /api/v1/templates
```

**Summary**: List templates  
**ID**: `listTemplates`

**Query Parameters**:
- `organization_id` (int): Filter by organization ID
- `template_type` (string): Filter by type (email, pdf, invoice, document)
- `is_active` (bool): Filter by active status
- `page` (int): Page number (default: 1)
- `page_size` (int): Page size (default: 20)

**Response 200** (`handlers.ListTemplatesResponse`)

---

### 12. Get Template
```
GET /api/v1/templates/{id}
```

**Summary**: Get template  
**ID**: `getTemplate`

**Path Parameters**:
- `id` (int, required): Template ID

**Response 200** (`handlers.TemplateResponse`)

---

### 13. Update Template
```
PUT /api/v1/templates/{id}
```

**Summary**: Update template  
**ID**: `updateTemplate`

**Path Parameters**:
- `id` (int, required): Template ID

**Request Body** (`handlers.UpdateTemplateRequest`):
```json
{
  "name": "Updated Invoice Template",
  "description": "Updated description",
  "content": "<!DOCTYPE html>...",
  "is_active": true,
  "is_default": false
}
```

**Response 200** (`handlers.TemplateResponse`)

**Note**: Content updates create a new version

---

### 14. Delete Template
```
DELETE /api/v1/templates/{id}
```

**Summary**: Delete template  
**ID**: `deleteTemplate`

**Path Parameters**:
- `id` (int, required): Template ID

**Response 200**:
```json
{
  "success": true,
  "message": "Template deleted successfully"
}
```

---

### 15. Get Template Content
```
GET /api/v1/templates/{id}/content
```

**Summary**: Get template content  
**ID**: `getTemplateContent`

**Path Parameters**:
- `id` (int, required): Template ID

**Response 200** (`handlers.TemplateWithContentResponse`):
```json
{
  "template": {...},
  "content": "<!DOCTYPE html>..."
}
```

---

### 16. Preview Template
```
GET /api/v1/templates/{id}/preview
```

**Summary**: Preview template  
**ID**: `previewTemplate`

**Path Parameters**:
- `id` (int, required): Template ID

**Response 200**: Rendered HTML (Content-Type: text/html)

---

### 17. Render Template
```
POST /api/v1/templates/{id}/render
```

**Summary**: Render template  
**ID**: `renderTemplate`

**Path Parameters**:
- `id` (int, required): Template ID

**Request Body** (`handlers.RenderTemplateRequest`):
```json
{
  "data": {
    "invoice_number": "INV-2025-001",
    "customer_name": "Acme Corp"
  }
}
```

**Response 200**: Rendered HTML (Content-Type: text/html)

---

### 18. Duplicate Template
```
POST /api/v1/templates/{id}/duplicate
```

**Summary**: Duplicate template  
**ID**: `duplicateTemplate`

**Path Parameters**:
- `id` (int, required): Template ID

**Request Body** (`handlers.DuplicateTemplateRequest`):
```json
{
  "name": "Copy of Standard Invoice"
}
```

**Response 201** (`handlers.TemplateResponse`)

---

### 19. Get Default Template
```
GET /api/v1/templates/default
```

**Summary**: Get default template  
**ID**: `getDefaultTemplate`

**Query Parameters**:
- `template_type` (string, required): Template type
- `organization_id` (int): Organization ID (fallback logic applies)

**Response 200** (`handlers.TemplateResponse`)

---

## PDF Generation API

### 20. Generate PDF from HTML
```
POST /api/v1/pdfs/generate
```

**Summary**: Generate PDF from HTML  
**ID**: `generatePDFFromHTML`

**Request Body** (`handlers.GeneratePDFRequest`):
```json
{
  "html": "<!DOCTYPE html>...",
  "filename": "document.pdf",
  "document_type": "report",
  "save_document": true,
  "metadata": {"author": "John Doe"}
}
```

**Response 200**: Binary PDF file

**Headers**:
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename="document.pdf"`

---

### 21. Generate PDF from Template
```
POST /api/v1/pdfs/from-template
```

**Summary**: Generate PDF from template  
**ID**: `generatePDFFromTemplate`

**Request Body** (`handlers.GeneratePDFFromTemplateRequest`):
```json
{
  "template_id": 1,
  "organization_id": 10,
  "data": {
    "invoice_number": "INV-2025-001",
    "customer_name": "Acme Corp"
  },
  "filename": "invoice.pdf",
  "document_type": "invoice",
  "save_document": true
}
```

**Response 200**: Binary PDF file

---

### 22. Generate Invoice PDF
```
POST /api/v1/pdfs/invoice?organization_id=10
```

**Summary**: Generate invoice PDF  
**ID**: `generateInvoicePDF`

**Query Parameters**:
- `organization_id` (int): Organization ID for template selection

**Request Body** (`handlers.InvoicePDFRequest`):
```json
{
  "invoice_number": "INV-2025-001",
  "invoice_date": "2025-12-26",
  "customer_name": "Acme Corp",
  "customer_address": "123 Main St",
  "items": [
    {
      "description": "Consulting Services",
      "quantity": 10,
      "unit_price": 100.00,
      "amount": 1000.00
    }
  ],
  "subtotal": 1000.00,
  "tax_rate": 0.10,
  "tax_amount": 100.00,
  "total": 1100.00,
  "payment_terms": "Net 30",
  "notes": "Thank you for your business"
}
```

**Response 200**: Binary PDF file

---

## Common Response Types

### Success Response
```json
{
  "success": true,
  "message": "Operation successful"
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error message"
}
```

### Pagination Response
```json
{
  "success": true,
  "data": [...],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

---

## Swagger Tags

- **Documents**: Document management endpoints (6 endpoints)
- **Invoice Numbers**: Invoice number generation and tracking (4 endpoints)
- **Templates**: Template management and rendering (10 endpoints)
- **PDFs**: PDF generation from HTML and templates (3 endpoints)

---

## Total API Endpoints: 23

**Swagger Files Generated**:
- `handlers/swagger_types.go` - Request/response type definitions
- All handlers include complete `@Summary`, `@Description`, `@Tags`, `@ID`, `@Param`, `@Success`, `@Failure`, `@Router`, and `@Security` annotations

**Documentation Status**: ✅ Complete
