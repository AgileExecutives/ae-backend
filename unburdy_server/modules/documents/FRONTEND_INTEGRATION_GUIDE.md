# Documents Module - Frontend Integration Guide

## Overview

The Documents Module provides a complete solution for template management, invoice generation, PDF creation, and document storage. This guide covers all API endpoints and integration workflows for frontend applications.

**Base URL**: `/api/v1`  
**Authentication**: All endpoints require Bearer token authentication

---

## Table of Contents

1. [Authentication](#authentication)
2. [Template Management](#template-management)
3. [Invoice Number Generation](#invoice-number-generation)
4. [Invoice Creation Workflow](#invoice-creation-workflow)
5. [Document Management](#document-management)
6. [PDF Generation](#pdf-generation)
7. [Complete Integration Examples](#complete-integration-examples)
8. [Error Handling](#error-handling)

---

## Authentication

All API requests require a Bearer token in the Authorization header:

```javascript
const headers = {
  'Authorization': `Bearer ${accessToken}`,
  'Content-Type': 'application/json'
};
```

---

## Template Management

Templates are reusable HTML/CSS designs for invoices, emails, and documents. They support variable substitution using `{{variable_name}}` syntax.

### 1. Create a Template

**Endpoint**: `POST /api/v1/templates`

**Request Body**:
```json
{
  "name": "Standard Invoice Template",
  "template_type": "invoice",
  "content": "<!DOCTYPE html><html><head><style>body { font-family: Arial; }</style></head><body><h1>Invoice #{{invoice_number}}</h1><p>Client: {{client_name}}</p><table><thead><tr><th>Item</th><th>Quantity</th><th>Price</th><th>Total</th></tr></thead><tbody>{{#items}}<tr><td>{{description}}</td><td>{{quantity}}</td><td>{{unit_price}}</td><td>{{total}}</td></tr>{{/items}}</tbody></table><p><strong>Total: {{total_amount}}</strong></p></body></html>",
  "description": "Standard invoice template with itemized billing",
  "default_variables": {
    "company_name": "Your Company",
    "company_address": "123 Main St",
    "currency": "EUR"
  },
  "is_active": true,
  "is_default": false
}
```

**Response** (201 Created):
```json
{
  "id": 1,
  "tenant_id": 123,
  "organization_id": 456,
  "name": "Standard Invoice Template",
  "template_type": "invoice",
  "description": "Standard invoice template with itemized billing",
  "is_active": true,
  "is_default": false,
  "version": 1,
  "created_at": "2025-12-26T10:00:00Z",
  "updated_at": "2025-12-26T10:00:00Z"
}
```

**Frontend Implementation**:
```javascript
async function createTemplate(templateData) {
  const response = await fetch('/api/v1/templates', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(templateData)
  });
  
  if (!response.ok) {
    throw new Error(`Failed to create template: ${response.statusText}`);
  }
  
  return await response.json();
}
```

### 2. List Templates

**Endpoint**: `GET /api/v1/templates`

**Query Parameters**:
- `organization_id` (optional): Filter by organization
- `template_type` (optional): Filter by type (invoice, email, pdf, document)
- `is_active` (optional): Filter by active status (true/false)
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20)

**Example Request**:
```
GET /api/v1/templates?template_type=invoice&is_active=true&page=1&page_size=10
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Standard Invoice Template",
      "template_type": "invoice",
      "is_active": true,
      "is_default": false,
      "version": 1,
      "created_at": "2025-12-26T10:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

**Frontend Implementation**:
```javascript
async function listTemplates(filters = {}) {
  const params = new URLSearchParams();
  
  if (filters.template_type) params.append('template_type', filters.template_type);
  if (filters.is_active !== undefined) params.append('is_active', filters.is_active);
  if (filters.page) params.append('page', filters.page);
  if (filters.page_size) params.append('page_size', filters.page_size);
  
  const response = await fetch(`/api/v1/templates?${params}`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  return await response.json();
}
```

### 3. Get Template with Content

**Endpoint**: `GET /api/v1/templates/{id}/content`

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Standard Invoice Template",
    "template_type": "invoice",
    "content": "<!DOCTYPE html>...",
    "default_variables": {
      "company_name": "Your Company",
      "currency": "EUR"
    },
    "is_active": true,
    "version": 1
  }
}
```

### 4. Preview Template with Sample Data

**Endpoint**: `GET /api/v1/templates/{id}/preview`

This endpoint renders the template with sample data to show how it will look.

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "rendered_html": "<!DOCTYPE html><html>...",
    "sample_data": {
      "invoice_number": "INV-2025-12-001",
      "client_name": "Sample Client",
      "total_amount": "1500.00"
    }
  }
}
```

### 5. Render Template with Custom Data

**Endpoint**: `POST /api/v1/templates/{id}/render`

**Request Body**:
```json
{
  "variables": {
    "invoice_number": "INV-2025-12-042",
    "client_name": "Acme Corporation",
    "client_address": "456 Business Ave",
    "invoice_date": "2025-12-26",
    "due_date": "2026-01-26",
    "items": [
      {
        "description": "Consulting Services",
        "quantity": 40,
        "unit_price": "150.00",
        "total": "6000.00"
      },
      {
        "description": "Software License",
        "quantity": 1,
        "unit_price": "2000.00",
        "total": "2000.00"
      }
    ],
    "total_amount": "8000.00",
    "currency": "EUR"
  }
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "rendered_html": "<!DOCTYPE html><html><body><h1>Invoice #INV-2025-12-042</h1>..."
  }
}
```

**Frontend Implementation**:
```javascript
async function renderTemplate(templateId, data) {
  const response = await fetch(`/api/v1/templates/${templateId}/render`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ variables: data })
  });
  
  return await response.json();
}
```

---

## Invoice Number Generation

The system provides sequential invoice number generation with Redis caching and PostgreSQL persistence.

### 1. Generate Invoice Number

**Endpoint**: `POST /api/v1/invoice-numbers/generate`

**Request Body**:
```json
{
  "organization_id": 456,
  "year": 2025,
  "month": 12,
  "prefix": "INV",
  "metadata": {
    "client_id": 789,
    "created_by": "user@example.com"
  }
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "invoice_number": "INV-2025-12-042",
    "organization_id": 456,
    "year": 2025,
    "month": 12,
    "sequence": 42,
    "prefix": "INV",
    "full_number": "INV-2025-12-042",
    "generated_at": "2025-12-26T10:30:00Z"
  }
}
```

**Frontend Implementation**:
```javascript
async function generateInvoiceNumber(organizationId, metadata = {}) {
  const now = new Date();
  const response = await fetch('/api/v1/invoice-numbers/generate', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      organization_id: organizationId,
      year: now.getFullYear(),
      month: now.getMonth() + 1,
      prefix: 'INV',
      metadata
    })
  });
  
  const result = await response.json();
  return result.data.full_number; // "INV-2025-12-042"
}
```

### 2. Get Current Sequence (Without Incrementing)

**Endpoint**: `GET /api/v1/invoice-numbers/current`

**Query Parameters**:
- `organization_id` (required): Organization ID
- `year` (optional): Year (defaults to current year)
- `month` (optional): Month (defaults to current month)

**Example**:
```
GET /api/v1/invoice-numbers/current?organization_id=456&year=2025&month=12
```

**Response** (200 OK):
```json
{
  "success": true,
  "organization_id": 456,
  "year": 2025,
  "month": 12,
  "current_sequence": 42
}
```

### 3. Void an Invoice Number

**Endpoint**: `DELETE /api/v1/invoice-numbers/void/{invoice_number}`

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Invoice number INV-2025-12-042 voided successfully"
}
```

---

## Invoice Creation Workflow

Here's a complete workflow for creating an invoice with template and document storage:

### Step-by-Step Process

```javascript
async function createInvoice(invoiceData) {
  try {
    // Step 1: Generate invoice number
    const invoiceNumber = await generateInvoiceNumber(
      invoiceData.organization_id,
      { client_id: invoiceData.client_id }
    );
    
    // Step 2: Prepare template data
    const templateData = {
      invoice_number: invoiceNumber,
      client_name: invoiceData.client_name,
      client_address: invoiceData.client_address,
      invoice_date: new Date().toISOString().split('T')[0],
      due_date: invoiceData.due_date,
      items: invoiceData.items.map(item => ({
        description: item.description,
        quantity: item.quantity,
        unit_price: item.unit_price.toFixed(2),
        total: (item.quantity * item.unit_price).toFixed(2)
      })),
      total_amount: invoiceData.total_amount.toFixed(2),
      currency: invoiceData.currency || 'EUR'
    };
    
    // Step 3: Generate PDF from template
    const pdfResult = await generateInvoicePDF({
      template_id: invoiceData.template_id,
      invoice_data: templateData
    });
    
    // Step 4: The PDF is automatically stored as a document
    // Return the document ID and download URL
    return {
      invoice_number: invoiceNumber,
      document_id: pdfResult.document_id,
      pdf_url: `/api/v1/documents/${pdfResult.document_id}/download`,
      created_at: new Date().toISOString()
    };
    
  } catch (error) {
    console.error('Invoice creation failed:', error);
    throw error;
  }
}
```

### React Component Example

```jsx
import React, { useState } from 'react';

function InvoiceCreator() {
  const [loading, setLoading] = useState(false);
  const [invoiceData, setInvoiceData] = useState({
    organization_id: 456,
    client_id: 789,
    client_name: '',
    client_address: '',
    due_date: '',
    template_id: 1,
    currency: 'EUR',
    items: [
      { description: '', quantity: 1, unit_price: 0 }
    ]
  });

  const calculateTotal = () => {
    return invoiceData.items.reduce(
      (sum, item) => sum + (item.quantity * item.unit_price),
      0
    );
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      const invoice = await createInvoice({
        ...invoiceData,
        total_amount: calculateTotal()
      });
      
      alert(`Invoice ${invoice.invoice_number} created successfully!`);
      
      // Open PDF in new tab
      window.open(invoice.pdf_url, '_blank');
      
    } catch (error) {
      alert(`Failed to create invoice: ${error.message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Form fields */}
      <button type="submit" disabled={loading}>
        {loading ? 'Creating Invoice...' : 'Create Invoice'}
      </button>
    </form>
  );
}
```

---

## Document Management

Documents are files stored in MinIO with metadata tracked in PostgreSQL.

### 1. Upload a Document

**Endpoint**: `POST /api/v1/documents`

**Request**: Multipart form data

**Form Fields**:
- `file` (required): The file to upload
- `document_type` (required): Type (invoice, contract, report, etc.)
- `bucket` (optional): Storage bucket (default: "documents")
- `path` (optional): Storage path (default: filename)
- `reference_type` (optional): Reference type (invoice, client, session)
- `reference_id` (optional): Reference ID
- `organization_id` (optional): Organization ID

**Frontend Implementation**:
```javascript
async function uploadDocument(file, metadata) {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('document_type', metadata.document_type);
  
  if (metadata.reference_type) {
    formData.append('reference_type', metadata.reference_type);
  }
  if (metadata.reference_id) {
    formData.append('reference_id', metadata.reference_id);
  }
  if (metadata.organization_id) {
    formData.append('organization_id', metadata.organization_id);
  }
  
  const response = await fetch('/api/v1/documents', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`
      // Don't set Content-Type - browser will set it with boundary
    },
    body: formData
  });
  
  return await response.json();
}

// Usage
const fileInput = document.getElementById('fileInput');
const file = fileInput.files[0];

const result = await uploadDocument(file, {
  document_type: 'invoice',
  reference_type: 'invoice',
  reference_id: 12345,
  organization_id: 456
});

console.log('Document uploaded:', result.id);
```

**Response** (201 Created):
```json
{
  "id": 100,
  "tenant_id": 123,
  "organization_id": 456,
  "filename": "invoice-2025-12-042.pdf",
  "original_filename": "invoice.pdf",
  "file_size": 45678,
  "content_type": "application/pdf",
  "document_type": "invoice",
  "bucket": "documents",
  "path": "invoices/2025/12/invoice-2025-12-042.pdf",
  "reference_type": "invoice",
  "reference_id": 12345,
  "created_at": "2025-12-26T11:00:00Z",
  "updated_at": "2025-12-26T11:00:00Z"
}
```

### 2. List Documents

**Endpoint**: `GET /api/v1/documents`

**Query Parameters**:
- `organization_id` (optional): Filter by organization
- `document_type` (optional): Filter by type
- `reference_type` (optional): Filter by reference type
- `reference_id` (optional): Filter by reference ID
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20)

**Example**:
```
GET /api/v1/documents?document_type=invoice&organization_id=456&page=1&page_size=10
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 100,
      "filename": "invoice-2025-12-042.pdf",
      "file_size": 45678,
      "content_type": "application/pdf",
      "document_type": "invoice",
      "reference_type": "invoice",
      "reference_id": 12345,
      "created_at": "2025-12-26T11:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

### 3. Get Document Metadata

**Endpoint**: `GET /api/v1/documents/{id}`

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Document retrieved successfully",
  "data": {
    "id": 100,
    "tenant_id": 123,
    "organization_id": 456,
    "filename": "invoice-2025-12-042.pdf",
    "original_filename": "invoice.pdf",
    "file_size": 45678,
    "content_type": "application/pdf",
    "document_type": "invoice",
    "bucket": "documents",
    "path": "invoices/2025/12/invoice-2025-12-042.pdf",
    "reference_type": "invoice",
    "reference_id": 12345,
    "created_at": "2025-12-26T11:00:00Z",
    "updated_at": "2025-12-26T11:00:00Z"
  }
}
```

### 4. Download Document

**Endpoint**: `GET /api/v1/documents/{id}/download`

**Response** (200 OK):
```json
{
  "success": true,
  "download_url": "https://minio.example.com/documents/invoices/2025/12/invoice-2025-12-042.pdf?X-Amz-Algorithm=...",
  "expires_in": 3600
}
```

**Frontend Implementation**:
```javascript
async function downloadDocument(documentId) {
  const response = await fetch(`/api/v1/documents/${documentId}/download`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const result = await response.json();
  
  // Open in new tab or download
  window.open(result.download_url, '_blank');
  
  // Or trigger download
  // const link = document.createElement('a');
  // link.href = result.download_url;
  // link.download = 'invoice.pdf';
  // link.click();
}
```

### 5. Delete Document

**Endpoint**: `DELETE /api/v1/documents/{id}`

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Document deleted successfully"
}
```

---

## PDF Generation

### 1. Generate PDF from HTML

**Endpoint**: `POST /api/v1/pdfs/generate`

**Request Body**:
```json
{
  "html": "<!DOCTYPE html><html><body><h1>Invoice #INV-001</h1><p>Total: €1,500.00</p></body></html>",
  "options": {
    "format": "A4",
    "margin": {
      "top": "1cm",
      "right": "1cm",
      "bottom": "1cm",
      "left": "1cm"
    }
  },
  "filename": "invoice-001.pdf",
  "store": true,
  "document_metadata": {
    "document_type": "invoice",
    "reference_type": "invoice",
    "reference_id": 12345,
    "organization_id": 456
  }
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "PDF generated and stored successfully",
  "document_id": 100,
  "file_size": 45678,
  "download_url": "/api/v1/documents/100/download"
}
```

### 2. Generate PDF from Template

**Endpoint**: `POST /api/v1/pdfs/from-template`

**Request Body**:
```json
{
  "template_id": 1,
  "variables": {
    "invoice_number": "INV-2025-12-042",
    "client_name": "Acme Corporation",
    "total_amount": "8000.00"
  },
  "options": {
    "format": "A4"
  },
  "filename": "invoice-2025-12-042.pdf",
  "store": true,
  "document_metadata": {
    "document_type": "invoice",
    "reference_type": "invoice",
    "reference_id": 12345,
    "organization_id": 456
  }
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "PDF generated from template and stored successfully",
  "document_id": 101,
  "file_size": 52341,
  "download_url": "/api/v1/documents/101/download"
}
```

### 3. Generate Invoice PDF (Simplified)

**Endpoint**: `POST /api/v1/pdfs/invoice`

This is a convenient endpoint that combines template rendering and PDF generation specifically for invoices.

**Request Body**:
```json
{
  "template_id": 1,
  "invoice_data": {
    "invoice_number": "INV-2025-12-042",
    "client_name": "Acme Corporation",
    "client_address": "456 Business Ave, Suite 200",
    "invoice_date": "2025-12-26",
    "due_date": "2026-01-26",
    "items": [
      {
        "description": "Consulting Services",
        "quantity": 40,
        "unit_price": "150.00",
        "total": "6000.00"
      },
      {
        "description": "Software License",
        "quantity": 1,
        "unit_price": "2000.00",
        "total": "2000.00"
      }
    ],
    "subtotal": "8000.00",
    "tax_rate": "19",
    "tax_amount": "1520.00",
    "total_amount": "9520.00",
    "currency": "EUR"
  },
  "organization_id": 456,
  "reference_id": 12345
}
```

**Response** (200 OK):
```json
{
  "success": true,
  "message": "Invoice PDF generated successfully",
  "document_id": 102,
  "invoice_number": "INV-2025-12-042",
  "file_size": 54321,
  "download_url": "/api/v1/documents/102/download"
}
```

**Frontend Implementation**:
```javascript
async function generateInvoicePDF(invoiceData) {
  const response = await fetch('/api/v1/pdfs/invoice', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(invoiceData)
  });
  
  if (!response.ok) {
    throw new Error('Failed to generate invoice PDF');
  }
  
  return await response.json();
}

// Usage
const result = await generateInvoicePDF({
  template_id: 1,
  invoice_data: {
    invoice_number: 'INV-2025-12-042',
    client_name: 'Acme Corporation',
    items: [...],
    total_amount: '9520.00',
    currency: 'EUR'
  },
  organization_id: 456,
  reference_id: 12345
});

// Download the generated PDF
window.open(result.download_url, '_blank');
```

---

## Complete Integration Examples

### Example 1: Invoice Creation Flow

```javascript
class InvoiceService {
  constructor(apiBaseUrl, authToken) {
    this.baseUrl = apiBaseUrl;
    this.token = authToken;
  }

  async createCompleteInvoice(invoiceData) {
    try {
      // 1. Generate invoice number
      const invoiceNumber = await this.generateInvoiceNumber(
        invoiceData.organization_id
      );

      // 2. Prepare invoice data
      const completeInvoiceData = {
        ...invoiceData,
        invoice_number: invoiceNumber,
        invoice_date: new Date().toISOString().split('T')[0]
      };

      // 3. Generate PDF from template
      const pdfResult = await this.generateInvoicePDF(completeInvoiceData);

      // 4. Return complete invoice information
      return {
        invoice_number: invoiceNumber,
        document_id: pdfResult.document_id,
        download_url: pdfResult.download_url,
        created_at: new Date().toISOString()
      };

    } catch (error) {
      console.error('Invoice creation failed:', error);
      throw error;
    }
  }

  async generateInvoiceNumber(organizationId) {
    const now = new Date();
    const response = await fetch(`${this.baseUrl}/invoice-numbers/generate`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        organization_id: organizationId,
        year: now.getFullYear(),
        month: now.getMonth() + 1,
        prefix: 'INV'
      })
    });

    const result = await response.json();
    return result.data.full_number;
  }

  async generateInvoicePDF(invoiceData) {
    const response = await fetch(`${this.baseUrl}/pdfs/invoice`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        template_id: invoiceData.template_id || 1,
        invoice_data: {
          invoice_number: invoiceData.invoice_number,
          client_name: invoiceData.client_name,
          client_address: invoiceData.client_address,
          invoice_date: invoiceData.invoice_date,
          due_date: invoiceData.due_date,
          items: invoiceData.items,
          total_amount: invoiceData.total_amount,
          currency: invoiceData.currency || 'EUR'
        },
        organization_id: invoiceData.organization_id,
        reference_id: invoiceData.reference_id
      })
    });

    return await response.json();
  }

  async downloadInvoice(documentId) {
    const response = await fetch(
      `${this.baseUrl}/documents/${documentId}/download`,
      {
        headers: { 'Authorization': `Bearer ${this.token}` }
      }
    );

    const result = await response.json();
    window.open(result.download_url, '_blank');
  }

  async listInvoices(organizationId, page = 1, pageSize = 20) {
    const params = new URLSearchParams({
      document_type: 'invoice',
      organization_id: organizationId,
      page,
      page_size: pageSize
    });

    const response = await fetch(
      `${this.baseUrl}/documents?${params}`,
      {
        headers: { 'Authorization': `Bearer ${this.token}` }
      }
    );

    return await response.json();
  }
}

// Usage
const invoiceService = new InvoiceService('/api/v1', userToken);

const newInvoice = await invoiceService.createCompleteInvoice({
  organization_id: 456,
  client_name: 'Acme Corporation',
  client_address: '456 Business Ave',
  due_date: '2026-01-26',
  template_id: 1,
  items: [
    {
      description: 'Consulting',
      quantity: 40,
      unit_price: '150.00',
      total: '6000.00'
    }
  ],
  total_amount: '6000.00',
  currency: 'EUR',
  reference_id: 12345
});

console.log('Invoice created:', newInvoice.invoice_number);
invoiceService.downloadInvoice(newInvoice.document_id);
```

### Example 2: Template Management UI

```javascript
class TemplateManager {
  constructor(apiBaseUrl, authToken) {
    this.baseUrl = apiBaseUrl;
    this.token = authToken;
  }

  async loadTemplates(type = 'invoice') {
    const response = await fetch(
      `${this.baseUrl}/templates?template_type=${type}&is_active=true`,
      {
        headers: { 'Authorization': `Bearer ${this.token}` }
      }
    );

    const result = await response.json();
    return result.data;
  }

  async previewTemplate(templateId) {
    const response = await fetch(
      `${this.baseUrl}/templates/${templateId}/preview`,
      {
        headers: { 'Authorization': `Bearer ${this.token}` }
      }
    );

    const result = await response.json();
    return result.data.rendered_html;
  }

  async saveTemplate(templateData) {
    const response = await fetch(`${this.baseUrl}/templates`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(templateData)
    });

    return await response.json();
  }

  async updateTemplate(templateId, updates) {
    const response = await fetch(
      `${this.baseUrl}/templates/${templateId}`,
      {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(updates)
      }
    );

    return await response.json();
  }
}
```

### Example 3: Vue.js Integration

```vue
<template>
  <div class="invoice-creator">
    <h2>Create Invoice</h2>
    
    <form @submit.prevent="createInvoice">
      <!-- Client Information -->
      <div class="form-group">
        <label>Client Name</label>
        <input v-model="invoice.client_name" required />
      </div>
      
      <div class="form-group">
        <label>Client Address</label>
        <textarea v-model="invoice.client_address" required></textarea>
      </div>
      
      <div class="form-group">
        <label>Due Date</label>
        <input type="date" v-model="invoice.due_date" required />
      </div>
      
      <!-- Invoice Items -->
      <div class="invoice-items">
        <h3>Items</h3>
        <div v-for="(item, index) in invoice.items" :key="index" class="item">
          <input v-model="item.description" placeholder="Description" />
          <input v-model.number="item.quantity" type="number" placeholder="Qty" />
          <input v-model.number="item.unit_price" type="number" step="0.01" placeholder="Price" />
          <span>Total: {{ (item.quantity * item.unit_price).toFixed(2) }}</span>
          <button type="button" @click="removeItem(index)">Remove</button>
        </div>
        <button type="button" @click="addItem">Add Item</button>
      </div>
      
      <!-- Total -->
      <div class="total">
        <strong>Total: € {{ calculateTotal().toFixed(2) }}</strong>
      </div>
      
      <!-- Submit -->
      <button type="submit" :disabled="loading">
        {{ loading ? 'Creating...' : 'Create Invoice' }}
      </button>
    </form>
    
    <!-- Success Message -->
    <div v-if="createdInvoice" class="success">
      <p>Invoice {{ createdInvoice.invoice_number }} created!</p>
      <button @click="downloadInvoice">Download PDF</button>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      loading: false,
      invoice: {
        organization_id: 456,
        client_name: '',
        client_address: '',
        due_date: '',
        template_id: 1,
        currency: 'EUR',
        items: [
          { description: '', quantity: 1, unit_price: 0 }
        ]
      },
      createdInvoice: null
    };
  },
  
  methods: {
    addItem() {
      this.invoice.items.push({ description: '', quantity: 1, unit_price: 0 });
    },
    
    removeItem(index) {
      this.invoice.items.splice(index, 1);
    },
    
    calculateTotal() {
      return this.invoice.items.reduce(
        (sum, item) => sum + (item.quantity * item.unit_price),
        0
      );
    },
    
    async createInvoice() {
      this.loading = true;
      
      try {
        // 1. Generate invoice number
        const invoiceNumberRes = await this.$api.post('/invoice-numbers/generate', {
          organization_id: this.invoice.organization_id,
          year: new Date().getFullYear(),
          month: new Date().getMonth() + 1,
          prefix: 'INV'
        });
        
        const invoiceNumber = invoiceNumberRes.data.data.full_number;
        
        // 2. Generate PDF
        const pdfRes = await this.$api.post('/pdfs/invoice', {
          template_id: this.invoice.template_id,
          invoice_data: {
            invoice_number: invoiceNumber,
            client_name: this.invoice.client_name,
            client_address: this.invoice.client_address,
            invoice_date: new Date().toISOString().split('T')[0],
            due_date: this.invoice.due_date,
            items: this.invoice.items.map(item => ({
              description: item.description,
              quantity: item.quantity,
              unit_price: item.unit_price.toFixed(2),
              total: (item.quantity * item.unit_price).toFixed(2)
            })),
            total_amount: this.calculateTotal().toFixed(2),
            currency: this.invoice.currency
          },
          organization_id: this.invoice.organization_id
        });
        
        this.createdInvoice = {
          invoice_number: invoiceNumber,
          document_id: pdfRes.data.document_id,
          download_url: pdfRes.data.download_url
        };
        
        this.$message.success(`Invoice ${invoiceNumber} created successfully!`);
        
      } catch (error) {
        this.$message.error(`Failed to create invoice: ${error.message}`);
      } finally {
        this.loading = false;
      }
    },
    
    async downloadInvoice() {
      const res = await this.$api.get(
        `/documents/${this.createdInvoice.document_id}/download`
      );
      window.open(res.data.download_url, '_blank');
    }
  }
};
</script>
```

---

## Error Handling

### Common Error Responses

All endpoints return consistent error responses:

**400 Bad Request**:
```json
{
  "error": "Invalid request",
  "message": "template_id is required"
}
```

**401 Unauthorized**:
```json
{
  "error": "Unauthorized",
  "message": "Invalid or missing authentication token"
}
```

**404 Not Found**:
```json
{
  "error": "Not found",
  "message": "Template with ID 999 not found"
}
```

**500 Internal Server Error**:
```json
{
  "error": "Internal server error",
  "message": "Failed to generate PDF"
}
```

### Error Handling Best Practices

```javascript
class APIClient {
  async request(url, options = {}) {
    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json',
          ...options.headers
        }
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || response.statusText);
      }

      return await response.json();

    } catch (error) {
      // Log error for debugging
      console.error('API Request failed:', error);

      // Handle specific error types
      if (error.message.includes('Unauthorized')) {
        // Redirect to login or refresh token
        this.handleUnauthorized();
      } else if (error.message.includes('not found')) {
        // Show user-friendly message
        this.showNotFoundError();
      } else {
        // Generic error handling
        this.showError(error.message);
      }

      throw error;
    }
  }

  handleUnauthorized() {
    // Redirect to login or attempt token refresh
    window.location.href = '/login';
  }

  showError(message) {
    // Display error to user (using your UI framework)
    alert(message);
  }

  showNotFoundError() {
    alert('The requested resource was not found');
  }
}
```

---

## Testing with Swagger UI

Access the complete API documentation and test endpoints interactively:

**Swagger UI URL**: `http://localhost:8080/swagger/index.html`

All endpoints are documented with:
- Request/response examples
- Parameter descriptions
- Authentication requirements
- Error responses

You can test endpoints directly from the Swagger UI before implementing them in your frontend.

---

## Support

For issues or questions:
1. Check the Swagger documentation at `/swagger/index.html`
2. Review the backend logs for detailed error messages
3. Ensure all required parameters are provided
4. Verify authentication token is valid and not expired
