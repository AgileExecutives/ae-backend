# Organization Module

A Go module for managing organizations within a multi-tenant application.

## Features

- **CRUD Operations**: Create, Read, Update, Delete organizations
- **Multi-tenant Support**: Organizations are scoped to tenants and users
- **RESTful API**: Well-documented API endpoints with Swagger
- **JSON Fields**: Support for flexible additional payment methods and invoice content
- **Comprehensive Testing**: Unit tests for all operations

## Entity Fields

- `name`: Organization name (required)
- `owner_name`: Owner's full name
- `owner_title`: Owner's title/position
- `street_address`: Street address
- `zip`: Postal code
- `city`: City name
- `email`: Contact email
- `tax_id`: Tax identification number
- `tax_rate`: Tax rate (decimal)
- `tax_ustid`: VAT ID (USt-ID)
- `unit_price`: Default unit price
- `bankaccount_owner`: Bank account owner name
- `bankaccount_bank`: Bank name
- `bankaccount_bic`: BIC/SWIFT code
- `bankaccount_iban`: IBAN number
- `additional_payment_methods`: JSON field for additional payment options
- `invoice_content`: JSON field for invoice template content

## API Endpoints

All endpoints require authentication (BearerAuth).

### Create Organization
`POST /api/v1/organizations`

### Get All Organizations
`GET /api/v1/organizations?page=1&limit=10`

### Get Organization by ID
`GET /api/v1/organizations/{id}`

### Update Organization
`PUT /api/v1/organizations/{id}`

### Delete Organization
`DELETE /api/v1/organizations/{id}`

## Service Methods

### Public Methods (Exposed to Other Modules)

```go
// GetOrganizations returns all organizations with pagination
func (s *OrganizationService) GetOrganizations(page, limit int, tenantID, userID uint) ([]entities.Organization, int64, error)
```

## Usage in Other Modules

```go
import (
    orgModule "github.com/unburdy/organization-module"
)

// Get the organization service
module := orgModule.NewCoreModule()
service := module.GetService()

// Fetch organizations
organizations, total, err := service.GetOrganizations(1, 10, tenantID, userID)
```

## Testing

Run tests with:
```bash
cd /Users/alex/src/ae/backend/modules/organization
go test ./tests -v
```

## Dependencies

- `github.com/ae-base-server` - Base authentication and utilities
- `github.com/gin-gonic/gin` - HTTP web framework
- `gorm.io/gorm` - ORM library
- `gorm.io/datatypes` - JSON field support
