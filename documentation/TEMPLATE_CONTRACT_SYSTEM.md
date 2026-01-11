# Template Contract System - Implementation Complete

## Overview

The Template Contract System has been successfully implemented in base-server. This system provides a robust, module-driven approach to template management with schema validation, multi-channel support, and public asset delivery.

## Implementation Status

### ✅ Phase 1: Data Model Changes
- **TemplateContract Entity** (`entities/contract.go`)
  - Unique constraint on `(module, template_key)`
  - JSONB fields for flexible schema storage
  - Support for EMAIL and DOCUMENT channels
  - Default sample data for testing

- **Template Entity Updates** (`entities/template.go`)
  - Added `Channel` enum (EMAIL, DOCUMENT)
  - Added `Module`, `TemplateKey`, `Channel` for contract binding
  - Added `Subject` field for EMAIL templates
  - Maintained backward compatibility with `TemplateType`

### ✅ Phase 2: Contract Management
- **ContractService** (`services/contract_service.go`)
  - RegisterContract (upsert logic)
  - GetContract, GetContractByID, ListContracts
  - ValidateChannel
  - UpdateContract, DeleteContract (with protection)
  
- **ContractHandler** (`handlers/contract_handler.go`)
  - Full REST API with Swagger documentation
  - Endpoints:
    - `POST /templates/contracts` - Register contract
    - `GET /templates/contracts` - List contracts
    - `GET /templates/contracts/:id` - Get by ID
    - `GET /templates/contracts/:module/:template_key` - Get by module/key
    - `PUT /templates/contracts/:id` - Update contract
    - `DELETE /templates/contracts/:id` - Delete contract

### ✅ Phase 3: Schema Validator
- **TemplateValidator** (`services/validator.go`)
  - Type validation (string, number, boolean, object, array)
  - Required field checking
  - Nested object validation
  - Array item validation
  - Detailed error reporting with field paths

### ✅ Phase 4: Rendering Engine
- **RenderService** (`services/render_service.go`)
  - Contract-based rendering with validation
  - Template lookup by module/key/channel
  - Template lookup by ID (for saved templates)
  - Subject rendering for EMAIL templates
  - HTML template compilation and execution
  - Fallback to sample data on validation failure

### ✅ Phase 5: Public Asset Endpoint
- **PublicAssetHandler** (`handlers/public_asset_handler.go`)
  - Public endpoint: `GET /public/templates/assets/:tenant/:template/:file`
  - No authentication required
  - Content-Type detection
  - Cache headers for performance
  - Support for images (JPEG, PNG, GIF, SVG), CSS, JS

### ✅ Phase 6: Preview API
- **PreviewHandler** (`handlers/preview_handler.go`)
  - Template preview endpoints:
    - `POST /templates/:id/preview` - Preview by ID (returns JSON)
    - `POST /templates/:id/preview/html` - Preview as HTML (for browser)
    - `POST /templates/preview` - Preview by contract (without saving)
    - `POST /templates/validate` - Validate data against schema
    - `GET /templates/required-fields` - Get required fields

## API Endpoints

### Contract Management

#### Register Contract
```bash
POST /api/templates/contracts
Authorization: Bearer {token}
Content-Type: application/json

{
  "module": "billing",
  "template_key": "invoice",
  "description": "Invoice template",
  "supported_channels": ["EMAIL", "DOCUMENT"],
  "variable_schema": {
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
    }
  },
  "default_sample_data": {
    "invoice_number": "INV-001",
    "client": {"name": "Sample", "email": "sample@example.com"}
  }
}
```

#### List Contracts
```bash
GET /api/templates/contracts?module=billing
Authorization: Bearer {token}
```

#### Get Contract
```bash
GET /api/templates/contracts/billing/invoice
Authorization: Bearer {token}
```

### Template Rendering

#### Render Template
```bash
POST /api/templates/preview
Authorization: Bearer {token}
Content-Type: application/json

{
  "module": "billing",
  "template_key": "invoice",
  "channel": "EMAIL",
  "data": {
    "invoice_number": "INV-2024-001",
    "client": {
      "name": "Client GmbH",
      "email": "client@example.com"
    }
  }
}
```

Response:
```json
{
  "html": "<html>...</html>",
  "subject": "Invoice INV-2024-001"
}
```

### Data Validation

#### Validate Data
```bash
POST /api/templates/validate?module=billing&template_key=invoice
Authorization: Bearer {token}
Content-Type: application/json

{
  "invoice_number": "INV-001",
  "client": {
    "name": "Test Client",
    "email": "test@example.com"
  }
}
```

#### Get Required Fields
```bash
GET /api/templates/required-fields?module=billing&template_key=invoice
Authorization: Bearer {token}
```

Response:
```json
{
  "required_fields": [
    "invoice_number",
    "client",
    "client.name",
    "client.email"
  ]
}
```

### Public Assets

#### Get Asset (No Auth Required)
```bash
GET /api/public/templates/assets/{tenant}/{template}/{filename}
```

Example:
```bash
GET /api/public/templates/assets/tenant-123/template-456/logo.png
```

## Schema Format

The variable schema uses a simple JSON structure:

```json
{
  "field_name": {
    "type": "string|number|boolean|object|array",
    "required": true|false,
    "properties": {
      // For nested objects
      "nested_field": {
        "type": "string",
        "required": true
      }
    },
    "items": {
      // For arrays
      "type": "object"
    }
  }
}
```

### Supported Types
- `string` - Text values
- `number` - Numeric values (int or float)
- `boolean` - True/false values
- `object` - Nested objects with properties
- `array` - Lists with optional item schema

## Migration Guide

### From Old Template System

The new system maintains backward compatibility through the `TemplateType` field, but new templates should use contracts:

**Old Way:**
```json
{
  "template_type": "invoice",
  "content": "<html>...</html>"
}
```

**New Way:**
1. Register contract:
```json
{
  "module": "billing",
  "template_key": "invoice",
  "supported_channels": ["EMAIL", "DOCUMENT"],
  "variable_schema": {...}
}
```

2. Create template instance:
```json
{
  "module": "billing",
  "template_key": "invoice",
  "channel": "DOCUMENT",
  "content": "<html>...</html>"
}
```

### Contract Naming Conventions

- **Module**: Service/feature area (e.g., "billing", "identity", "notification")
- **Template Key**: Specific purpose (e.g., "invoice", "password_reset", "welcome_email")
- **Channel**: Delivery method (EMAIL or DOCUMENT)

Examples:
- `billing.invoice.DOCUMENT` - PDF invoice
- `billing.invoice.EMAIL` - Email invoice notification
- `identity.password_reset.EMAIL` - Password reset email
- `notification.welcome_email.EMAIL` - Welcome email

## Testing

Run the test suite:
```bash
cd /Users/alex/src/ae/backend/base-server
./tests/test-contract-system.sh
```

This will test:
1. Contract registration
2. Contract listing
3. Contract retrieval
4. Data validation
5. Required fields extraction

## Database Migrations

The system automatically creates:
- `template_contracts` table with unique index on `(module, template_key)`
- Updated `templates` table with `module`, `template_key`, `channel`, `subject` columns
- Index on `(module, template_key, channel)` for fast lookups

## Performance Considerations

### Caching
- Contract lookups are optimized with database indexes
- Public assets include Cache-Control headers (24h)
- Consider adding Redis caching for frequently-used contracts

### Asset Delivery
- Assets are streamed from MinIO
- Content-Type is automatically detected
- Supports range requests for large files

## Security

### Authentication
- All template/contract management requires authentication
- Public asset endpoint is intentionally unauthenticated for email rendering

### Validation
- All input data is validated against schemas before rendering
- SQL injection protection through GORM
- XSS protection in template rendering (use Go's html/template)

## Next Steps

### Phase 7: Testing (In Progress)
- Create integration tests
- Test concurrent contract updates
- Test validation edge cases
- Performance testing

### Phase 8: Migration Scripts
- Create migration tool for existing templates
- Map old TemplateType values to new contracts
- Bulk migration utilities

## Files Modified/Created

### New Files
- `modules/templates/entities/contract.go`
- `modules/templates/services/contract_service.go`
- `modules/templates/services/validator.go`
- `modules/templates/services/render_service.go`
- `modules/templates/handlers/contract_handler.go`
- `modules/templates/handlers/preview_handler.go`
- `modules/templates/handlers/public_asset_handler.go`
- `tests/test-contract-system.sh`

### Modified Files
- `modules/templates/entities/template.go` - Added contract fields
- `modules/templates/routes/template_routes.go` - Added new routes
- `modules/templates/module.go` - Added service initialization
- `modules/templates/handlers/template_handler.go` - Fixed Swagger annotations

## Support

For questions or issues with the template contract system:
1. Check the Swagger documentation at `/swagger/index.html`
2. Review the test script for usage examples
3. Check GORM logs for database-related issues
4. Enable debug logging in the base-server

## Changelog

### 2024-01-11
- ✅ Implemented Phase 1-6 of refactoring plan
- ✅ Added contract management system
- ✅ Added schema validation
- ✅ Added rendering engine
- ✅ Added public asset endpoint
- ✅ Added preview API
- ✅ Created test scripts
- ✅ Updated Swagger documentation
