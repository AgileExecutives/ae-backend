# Template Contract System - Architecture Diagram

## System Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                        Template Contract System                  │
└─────────────────────────────────────────────────────────────────┘

┌──────────────┐
│   Module A   │ (e.g., billing)
│  (Owner)     │
└──────┬───────┘
       │ Registers
       ▼
┌──────────────────────────────────────────────────────────────────┐
│  TemplateContract                                                 │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ Module: "billing"                                          │  │
│  │ TemplateKey: "invoice"                                     │  │
│  │ SupportedChannels: ["EMAIL", "DOCUMENT"]                   │  │
│  │ VariableSchema: {                                          │  │
│  │   "invoice_number": {"type": "string", "required": true},  │  │
│  │   "client": {"type": "object", ...}                        │  │
│  │ }                                                           │  │
│  │ DefaultSampleData: {...}                                   │  │
│  └────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────┘
       │
       │ Defines
       ▼
┌──────────────────────────────────────────────────────────────────┐
│  Template Instances                                               │
│                                                                    │
│  ┌─────────────────────────┐  ┌──────────────────────────┐       │
│  │ Template #1             │  │ Template #2              │       │
│  │ Module: "billing"       │  │ Module: "billing"        │       │
│  │ TemplateKey: "invoice"  │  │ TemplateKey: "invoice"   │       │
│  │ Channel: EMAIL          │  │ Channel: DOCUMENT        │       │
│  │ Subject: "Invoice {{..}}" │  │ Subject: null          │       │
│  │ Content: HTML template  │  │ Content: HTML template   │       │
│  └─────────────────────────┘  └──────────────────────────┘       │
└──────────────────────────────────────────────────────────────────┘
       │                              │
       │ Render                       │ Render
       ▼                              ▼
┌──────────────────┐           ┌─────────────────┐
│  Email Output    │           │  PDF Output     │
│  - HTML Body     │           │  - HTML for PDF │
│  - Subject Line  │           │                 │
└──────────────────┘           └─────────────────┘
```

## Component Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Request                           │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│  HTTP Layer (Handlers)                                           │
│  ┌──────────────┐ ┌──────────────┐ ┌────────────────────────┐   │
│  │ Contract     │ │ Preview      │ │ PublicAsset            │   │
│  │ Handler      │ │ Handler      │ │ Handler                │   │
│  └──────┬───────┘ └──────┬───────┘ └─────────┬──────────────┘   │
└─────────┼────────────────┼──────────────────┼──────────────────┘
          │                │                  │
          ▼                ▼                  ▼
┌─────────────────────────────────────────────────────────────────┐
│  Service Layer                                                   │
│  ┌──────────────┐ ┌──────────────┐ ┌────────────────────────┐   │
│  │ Contract     │ │ Render       │ │ Template               │   │
│  │ Service      │ │ Service      │ │ Service                │   │
│  │              │ │  │           │ │                        │   │
│  │ - Register   │ │  ├─Validate  │ │ - CRUD                 │   │
│  │ - Get        │ │  ├─Render    │ │ - MinIO                │   │
│  │ - Update     │ │  └─Subject   │ │                        │   │
│  │ - Delete     │ │              │ │                        │   │
│  └──────┬───────┘ └──────┬───────┘ └────────┬───────────────┘   │
└─────────┼────────────────┼──────────────────┼──────────────────┘
          │                │                  │
          │                ▼                  │
          │    ┌──────────────────┐           │
          │    │ Template         │           │
          │    │ Validator        │           │
          │    │  - Type Check    │           │
          │    │  - Required      │           │
          │    │  - Nested        │           │
          │    └──────────────────┘           │
          │                                   │
          ▼                                   ▼
┌─────────────────────────────────────────────────────────────────┐
│  Data Layer                                                      │
│  ┌──────────────┐ ┌──────────────┐ ┌────────────────────────┐   │
│  │ PostgreSQL   │ │ PostgreSQL   │ │ MinIO                  │   │
│  │              │ │              │ │                        │   │
│  │ template_    │ │ templates    │ │ templates/             │   │
│  │ contracts    │ │              │ │  └─{tenant}/           │   │
│  │              │ │              │ │     └─{template}/      │   │
│  │              │ │              │ │        └─assets/       │   │
│  └──────────────┘ └──────────────┘ └────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow: Render Template

```
1. Client Request
   │
   ├─ POST /templates/preview
   │  {
   │    "module": "billing",
   │    "template_key": "invoice",
   │    "channel": "EMAIL",
   │    "data": {...}
   │  }
   │
2. PreviewHandler
   │
   ├─ Extract tenant_id from JWT
   ├─ Parse request body
   │
3. RenderService.Render()
   │
   ├─ Step 1: Lookup Contract
   │  │
   │  └─ ContractService.GetContract("billing", "invoice")
   │     │
   │     └─ SELECT * FROM template_contracts
   │        WHERE module='billing' AND template_key='invoice'
   │
   ├─ Step 2: Validate Channel
   │  │
   │  └─ Check "EMAIL" in contract.SupportedChannels
   │
   ├─ Step 3: Validate Data
   │  │
   │  └─ TemplateValidator.Validate(schema, data)
   │     │
   │     ├─ Check required fields
   │     ├─ Check types (string, number, object, etc.)
   │     └─ Check nested properties
   │
   ├─ Step 4: Lookup Template
   │  │
   │  └─ SELECT * FROM templates
   │     WHERE module='billing'
   │       AND template_key='invoice'
   │       AND channel='EMAIL'
   │       AND tenant_id='...'
   │
   ├─ Step 5: Render Template
   │  │
   │  ├─ Parse HTML template
   │  ├─ Execute with data
   │  └─ Render subject (if EMAIL)
   │
   └─ Return Result
      │
      └─ {
           "html": "<html>...</html>",
           "subject": "Invoice INV-001"
         }
```

## Data Flow: Public Asset

```
1. Browser Request (NO AUTH)
   │
   ├─ GET /public/templates/assets/tenant-123/template-456/logo.png
   │
2. PublicAssetHandler
   │
   ├─ Extract: tenant=tenant-123, template=template-456, file=logo.png
   │
3. MinIO Lookup
   │
   ├─ Path: templates/tenant-123/template-456/assets/logo.png
   │
4. Stream Response
   │
   ├─ Set Content-Type: image/png
   ├─ Set Cache-Control: public, max-age=86400
   └─ Stream bytes
```

## Schema Validation Flow

```
Contract Schema:
{
  "invoice_number": {
    "type": "string",
    "required": true
  },
  "client": {
    "type": "object",
    "required": true,
    "properties": {
      "name": {"type": "string", "required": true},
      "email": {"type": "string", "required": true}
    }
  },
  "items": {
    "type": "array",
    "items": {
      "type": "object"
    }
  }
}

User Data:
{
  "invoice_number": "INV-001",     ✓ string, required → OK
  "client": {                      ✓ object, required → OK
    "name": "Test Client",         ✓ string, required → OK
    "email": "test@example.com"    ✓ string, required → OK
  },
  "items": [                       ✓ array → OK
    {"desc": "Item 1"}             ✓ object → OK
  ]
}

Validation Result: ✅ PASS
```

## Module Ownership Model

```
┌────────────────────────────────────────────────────────────┐
│  Module: billing                                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Contracts:                                          │  │
│  │  - invoice                                           │  │
│  │  - receipt                                           │  │
│  │  - statement                                         │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│  Module: identity                                           │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Contracts:                                          │  │
│  │  - password_reset                                    │  │
│  │  - email_verification                                │  │
│  │  - welcome_email                                     │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│  Module: notification                                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Contracts:                                          │  │
│  │  - generic_email                                     │  │
│  │  - sms                                               │  │
│  │  - push_notification                                 │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────┘
```

## Multi-Channel Support

```
Contract: billing.invoice
Channels: [EMAIL, DOCUMENT]

┌────────────────────────────────────────────────────────────┐
│  Template Instance 1                                        │
│  - Module: billing                                          │
│  - TemplateKey: invoice                                     │
│  - Channel: EMAIL                                           │
│  - Subject: "Your Invoice {{.invoice_number}}"             │
│  - Content: <html>Email-friendly layout...</html>          │
│  - Use Case: Send invoice via email                        │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│  Template Instance 2                                        │
│  - Module: billing                                          │
│  - TemplateKey: invoice                                     │
│  - Channel: DOCUMENT                                        │
│  - Subject: null                                            │
│  - Content: <html>PDF-optimized layout...</html>           │
│  - Use Case: Generate PDF invoice                          │
└────────────────────────────────────────────────────────────┘

Both share the same:
- Variable schema
- Validation rules
- Sample data
```

## Benefits Visualization

```
WITHOUT Contract System:
┌─────────────────┐
│ Template        │
│  - Content      │
│  - Type         │
└─────────────────┘
Problems:
❌ No validation
❌ No schema
❌ No reusability
❌ No type safety

WITH Contract System:
┌─────────────────────────────────────┐
│ Contract (Shared Definition)        │
│  - Module                            │
│  - TemplateKey                       │
│  - VariableSchema                    │
│  - SupportedChannels                 │
│  - DefaultSampleData                 │
└─────────────────────────────────────┘
         │
         ├─ Template (EMAIL)
         ├─ Template (DOCUMENT)
         └─ Template (SMS)

Benefits:
✅ Schema validation
✅ Type safety
✅ Reusable across channels
✅ Self-documenting
✅ Testable with sample data
```

---

## Quick Reference

### Contract → Template → Render Flow
```
1. Register Contract   (Module owns)
2. Create Template     (Implements contract)
3. Render with Data    (Validated & safe)
```

### Key Concepts
- **Contract**: What CAN be rendered (schema, channels)
- **Template**: HOW to render (HTML content)
- **Channel**: WHERE it's used (EMAIL, DOCUMENT)
- **Module**: WHO owns it (billing, identity, etc.)
