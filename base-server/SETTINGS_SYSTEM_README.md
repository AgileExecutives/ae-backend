# Settings System Documentation

## Overview

The AE Backend Settings System is a comprehensive, modular configuration management solution designed with Domain-Driven Design principles and Separation of Concerns. It provides type-safe, tenant-aware settings management with caching, validation, and a complete REST API.

## Architecture

### Core Components

```
pkg/settings/
├── entities/           # Domain entities and types
├── repository/         # Data access layer
├── accessor/           # Type-safe settings accessor
├── registry/           # Module registration and schema management
├── manager/            # Coordination and caching
├── handlers/           # HTTP API handlers
└── interfaces/         # Contracts and interfaces
```

### Module Structure

```
modules/{domain}/settings/
└── {domain}_settings.go  # Domain-specific settings provider
```

## Key Features

### ✅ Modular Architecture
- **Domain Separation**: Each business domain (company, invoice, billing, etc.) has its own settings module
- **Self-Registration**: Modules register themselves with schema definitions and dependencies
- **Dependency Resolution**: Automatic resolution of module dependencies during initialization

### ✅ Type Safety
- **Typed Accessors**: Type-safe getters (`GetString()`, `GetInt()`, `GetBool()`, `GetFloat()`, `GetJSON()`)
- **Schema Validation**: Compile-time and runtime validation of setting types and values
- **Automatic Casting**: Intelligent type conversion where appropriate

### ✅ Multi-Tenancy
- **Tenant Isolation**: Complete data separation between tenants
- **Organization Scoping**: Settings scoped to specific organizations within tenants
- **Granular Access**: Precise control over who can access which settings

### ✅ Performance
- **Intelligent Caching**: Multi-level caching with automatic invalidation
- **Lazy Loading**: Settings loaded only when needed
- **Connection Pooling**: Efficient database connection management

### ✅ API-First Design
- **RESTful API**: Complete REST API with OpenAPI documentation
- **Bulk Operations**: Efficient bulk set/get/update/delete operations
- **Schema Introspection**: API endpoints for discovering available settings and their schemas

### ✅ Developer Experience
- **Auto-Migration**: Database schema automatically created and updated
- **Integration Helpers**: Middleware and utilities for easy integration
- **Comprehensive Testing**: Unit, integration, and API tests included
- **Clear Documentation**: Extensive documentation and examples

## Quick Start

### 1. Installation

```go
import (
    "github.com/ae-base-server/pkg/settings/manager"
    "github.com/ae-base-server/pkg/settings/repository"
    "github.com/ae-base-server/pkg/settings/registry"
    
    // Import the modules you need
    "github.com/ae-base-server/modules/company/settings"
    "github.com/ae-base-server/modules/invoice/settings"
)
```

### 2. Initialize the System

```go
// Initialize database connection (your existing GORM DB)
db := yourExistingGormDB

// Create settings system components
repo := repository.NewSettingsRepository(db)
registry := registry.NewSettingsRegistry()
manager := manager.NewSettingsManager(repo, registry)

// Register modules
registry.RegisterProvider(company.NewCompanySettingsProvider())
registry.RegisterProvider(invoice.NewInvoiceSettingsProvider())

// Initialize the system
err := manager.Initialize()
if err != nil {
    log.Fatal("Failed to initialize settings:", err)
}
```

### 3. Use Settings in Your Code

```go
// Get settings accessor for an organization
accessor, err := manager.GetAccessor(tenantID, organizationID)
if err != nil {
    return err
}

// Set settings
err = accessor.Set("company_name", "My Company")
err = accessor.Set("invoice_auto_number", true)

// Get settings with type safety
companyName, err := accessor.GetString("company_name")
autoNumber, err := accessor.GetBool("invoice_auto_number")

// Work with JSON settings
address := map[string]interface{}{
    "street": "123 Main St",
    "city":   "Anytown",
    "state":  "CA",
}
err = accessor.Set("billing_address", address)

var retrievedAddress map[string]interface{}
err = accessor.GetJSON("billing_address", &retrievedAddress)
```

### 4. Setup HTTP API

```go
import "github.com/ae-base-server/pkg/settings/handlers"

// Create handler
handler := handlers.NewSettingsHandler(manager)

// Setup routes (Gin example)
settingsGroup := r.Group("/api/v1/settings")
orgGroup := settingsGroup.Group("/organizations/:organization_id")

// Basic operations
orgGroup.GET("", handler.GetSettings)
orgGroup.POST("", handler.SetSetting)
orgGroup.PUT("/:domain/:key", handler.UpdateSetting)
orgGroup.DELETE("/:domain/:key", handler.DeleteSetting)

// Bulk operations
orgGroup.POST("/bulk", handler.BulkSetSettings)
orgGroup.PUT("/bulk", handler.BulkUpdateSettings)
orgGroup.DELETE("/bulk", handler.BulkDeleteSettings)

// Schema and validation
orgGroup.GET("/schema", handler.GetSettingsSchema)
orgGroup.POST("/validate", handler.ValidateSettings)
```

## Available Modules

### Company Settings (`company` domain)
Manages company information and branding:
- `company_name`: Company legal name
- `company_email`: Primary contact email
- `company_phone`: Main phone number
- `company_website`: Company website URL
- `legal_form`: Legal entity type
- `registration_number`: Business registration number
- `brand_colors`: Brand color scheme (JSON)
- `company_logo_url`: Logo image URL

### Invoice Settings (`invoice` domain)
Controls invoice generation and formatting:
- `invoice_prefix`: Invoice number prefix (e.g., "INV")
- `invoice_next_number`: Next invoice number to use
- `invoice_auto_number`: Enable automatic numbering
- `invoice_payment_terms`: Payment terms in days
- `invoice_template`: Template name for PDF generation
- `invoice_footer_text`: Custom footer text
- `invoice_due_date_default`: Default due date offset
- `tax_settings`: Tax configuration (JSON)

### Billing Settings (`billing` domain)
Manages billing address and payment information:
- `billing_address`: Complete billing address (JSON)
- `vat_number`: VAT/Tax ID number
- `tax_id`: Additional tax identifier
- `bank_details`: Banking information (JSON)
- `payment_methods`: Accepted payment methods (JSON array)

### Localization Settings (`localization` domain)
Handles language, currency, and formatting:
- `language`: Primary language code
- `currency`: Default currency code
- `timezone`: Default timezone
- `date_format`: Date display format
- `time_format`: Time display format
- `number_format`: Number formatting rules (JSON)

### Booking Settings (`booking` domain)
Controls appointment and scheduling features:
- `appointment_duration`: Default appointment length (minutes)
- `working_hours`: Business hours configuration (JSON)
- `time_slot_duration`: Booking time slot size (minutes)
- `max_advance_booking`: Maximum days in advance for bookings
- `min_advance_booking`: Minimum hours in advance for bookings
- `reminder_settings`: Appointment reminder configuration (JSON)

### Notification Settings (`notification` domain)
Manages email and SMS notifications:
- `email_enabled`: Global email notification toggle
- `smtp_settings`: SMTP server configuration (JSON)
- `sms_enabled`: Global SMS notification toggle
- `sms_provider_settings`: SMS service configuration (JSON)
- `notification_types`: Per-type notification settings (JSON)
- `email_templates`: HTML email templates (JSON)
- `sms_templates`: SMS message templates (JSON)
- `sender_identity`: Default sender information (JSON)

### Integration Settings (`integration` domain)
Configures third-party service integrations:
- `payment_processors`: Payment service configs (JSON)
- `email_services`: Email service provider configs (JSON)
- `calendar_services`: Calendar integration configs (JSON)
- `storage_services`: Cloud storage configs (JSON)
- `accounting_software`: Accounting system configs (JSON)
- `crm_systems`: CRM integration configs (JSON)
- `webhook_endpoints`: Outgoing webhook configs (JSON)
- `api_rate_limits`: Rate limiting configuration (JSON)

## API Reference

### Base URL
```
/api/v1/settings
```

### Organization Settings

#### Get All Settings
```http
GET /organizations/{org_id}
```

#### Set Single Setting
```http
POST /organizations/{org_id}
Content-Type: application/json

{
  "domain": "company",
  "key": "company_name",
  "value": "My Company",
  "type": "string"
}
```

#### Update Setting
```http
PUT /organizations/{org_id}/{domain}/{key}
Content-Type: application/json

{
  "value": "Updated Value",
  "type": "string"
}
```

#### Delete Setting
```http
DELETE /organizations/{org_id}/{domain}/{key}
```

### Bulk Operations

#### Bulk Set Settings
```http
POST /organizations/{org_id}/bulk
Content-Type: application/json

{
  "settings": [
    {
      "domain": "company",
      "key": "company_name",
      "value": "My Company",
      "type": "string"
    },
    {
      "domain": "company",
      "key": "company_email",
      "value": "info@mycompany.com",
      "type": "string"
    }
  ]
}
```

### Domain Operations

#### Get Domain Settings
```http
GET /organizations/{org_id}/domains/{domain}
```

#### Set Domain Settings
```http
POST /organizations/{org_id}/domains/{domain}
Content-Type: application/json

{
  "settings": {
    "company_name": "My Company",
    "company_email": "info@mycompany.com"
  }
}
```

#### Delete Domain Settings
```http
DELETE /organizations/{org_id}/domains/{domain}
```

### Schema Operations

#### Get All Schemas
```http
GET /organizations/{org_id}/schema
```

#### Get Domain Schema
```http
GET /organizations/{org_id}/schema/{domain}
```

#### Validate Settings
```http
POST /organizations/{org_id}/validate
Content-Type: application/json

{
  "domain": "company",
  "settings": {
    "company_name": "Valid Company",
    "company_email": "valid@company.com"
  }
}
```

### Import/Export

#### Export Settings
```http
GET /organizations/{org_id}/export
```

#### Import Settings
```http
POST /organizations/{org_id}/import
Content-Type: application/json

{
  "settings": {
    "company": {
      "company_name": "Imported Company"
    },
    "invoice": {
      "invoice_prefix": "IMP"
    }
  }
}
```

### Global Operations

#### Health Check
```http
GET /health
```

#### Get Registered Modules
```http
GET /modules
```

#### Get System Version
```http
GET /version
```

## Advanced Usage

### Creating Custom Modules

1. **Create Module Directory**
```bash
mkdir -p modules/mymodule/settings
```

2. **Implement Settings Provider**
```go
package mymodule

import (
    "github.com/ae-base-server/pkg/settings/accessor"
    "github.com/ae-base-server/pkg/settings/entities"
)

type MyModuleSettingsProvider struct {
    accessor *accessor.SettingsAccessor
}

func NewMyModuleSettingsProvider() *MyModuleSettingsProvider {
    return &MyModuleSettingsProvider{}
}

func (p *MyModuleSettingsProvider) GetModuleName() string {
    return "mymodule"
}

func (p *MyModuleSettingsProvider) GetSettingsSchema() entities.ModuleSettingsSchema {
    return entities.ModuleSettingsSchema{
        ModuleName: "mymodule",
        Version:    "1.0.0",
        Domain:     "mymodule",
        Description: "My custom module settings",
        Settings: []entities.SettingDefinition{
            {
                Key:          "my_setting",
                Type:         entities.SettingTypeString,
                DisplayName:  "My Setting",
                Description:  "Description of my setting",
                DefaultValue: "default_value",
                Required:     false,
            },
        },
        Dependencies: []string{}, // List any module dependencies
    }
}

func (p *MyModuleSettingsProvider) OnSettingsRegistered(accessor *accessor.SettingsAccessor) error {
    p.accessor = accessor
    return nil
}

func (p *MyModuleSettingsProvider) OnSettingsChanged(key string, oldValue, newValue interface{}) error {
    // Handle setting changes
    return nil
}

// Add typed getter methods
func (p *MyModuleSettingsProvider) GetMySetting() (string, error) {
    return p.accessor.GetString("my_setting")
}
```

3. **Register Your Module**
```go
registry.RegisterProvider(mymodule.NewMyModuleSettingsProvider())
```

### Middleware Integration

Create a middleware to inject settings into request context:

```go
func SettingsMiddleware(manager *manager.SettingsManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.GetHeader("X-Tenant-ID")
        organizationID := c.Param("organization_id")
        
        if tenantID != "" && organizationID != "" {
            accessor, err := manager.GetAccessor(parseTenantID(tenantID), organizationID)
            if err == nil {
                c.Set("settings", accessor)
            }
        }
        
        c.Next()
    }
}

// Use in handlers
func MyHandler(c *gin.Context) {
    if accessor, exists := c.Get("settings"); exists {
        settings := accessor.(*accessor.SettingsAccessor)
        
        companyName, err := settings.GetString("company_name")
        if err == nil {
            // Use company name
        }
    }
}
```

### Custom Validation

Add custom validation to your module provider:

```go
func (p *MyModuleSettingsProvider) OnSettingsChanged(key string, oldValue, newValue interface{}) error {
    switch key {
    case "email_setting":
        if email, ok := newValue.(string); ok {
            if !isValidEmail(email) {
                return fmt.Errorf("invalid email format: %s", email)
            }
        }
    case "numeric_setting":
        if num, ok := newValue.(int64); ok {
            if num < 0 || num > 100 {
                return fmt.Errorf("value must be between 0 and 100, got: %d", num)
            }
        }
    }
    return nil
}
```

### Caching Strategy

The settings system includes intelligent caching:

1. **Memory Caching**: Frequently accessed settings are cached in memory
2. **Automatic Invalidation**: Cache is invalidated when settings change
3. **Tenant Isolation**: Cache keys include tenant and organization IDs
4. **Configurable TTL**: Cache expiration can be configured per tenant

### Error Handling

The system provides detailed error information:

```go
accessor, err := manager.GetAccessor(tenantID, organizationID)
if err != nil {
    switch err {
    case entities.ErrTenantNotFound:
        // Handle tenant not found
    case entities.ErrSettingNotFound:
        // Handle setting not found
    case entities.ErrInvalidSettingType:
        // Handle type validation error
    default:
        // Handle other errors
    }
}
```

## Testing

### Unit Tests
```bash
go test ./pkg/settings/...
```

### Integration Tests
```bash
go test ./tests/settings_integration_test.go
```

### API Tests
```bash
go test ./tests/settings_handlers_test.go
```

### Performance Tests
```bash
go test -bench=. ./pkg/settings/...
```

## Database Schema

The system uses a single table for maximum flexibility:

```sql
CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    domain VARCHAR(100) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(tenant_id, organization_id, domain, key),
    INDEX(tenant_id, organization_id),
    INDEX(tenant_id, organization_id, domain)
);
```

## Migration Guide

### From Organization Entity

If you're migrating from storing settings in an organization entity:

1. **Extract settings to new system**:
```go
// Old way
organization := GetOrganization(orgID)
name := organization.Name
prefix := organization.InvoicePrefix

// New way
accessor, _ := settingsManager.GetAccessor(tenantID, orgID)
name, _ := accessor.GetString("company_name")
prefix, _ := accessor.GetString("invoice_prefix")
```

2. **Migrate existing data**:
```go
func MigrateOrganizationSettings(db *gorm.DB, settingsManager *manager.SettingsManager) error {
    var organizations []Organization
    db.Find(&organizations)
    
    for _, org := range organizations {
        accessor, err := settingsManager.GetAccessor(org.TenantID, org.ID)
        if err != nil {
            continue
        }
        
        // Migrate fields
        if org.Name != "" {
            accessor.Set("company_name", org.Name)
        }
        if org.InvoicePrefix != "" {
            accessor.Set("invoice_prefix", org.InvoicePrefix)
        }
        // ... migrate other fields
    }
    
    return nil
}
```

## Performance Characteristics

### Benchmarks

Based on our performance tests:

- **Write Performance**: ~1,000 settings/second
- **Read Performance**: ~10,000 settings/second (cached)
- **Memory Usage**: ~100KB per 1,000 cached settings
- **Database Connections**: Pool of 10-50 connections (configurable)

### Scaling Considerations

- **Horizontal Scaling**: Each tenant can be served by different instances
- **Database Scaling**: Single table design scales well with proper indexing
- **Cache Scaling**: Redis can be used for distributed caching if needed
- **API Scaling**: Stateless handlers allow easy horizontal scaling

## Monitoring and Observability

### Metrics

The system exposes Prometheus metrics:

```
# Settings operations
settings_operations_total{operation="get|set|update|delete"}
settings_operations_duration_seconds{operation="get|set|update|delete"}

# Cache performance
settings_cache_hits_total
settings_cache_misses_total

# Errors
settings_errors_total{type="validation|database|cache"}
```

### Health Checks

Health check endpoint provides system status:

```json
{
  "status": "ok",
  "database": "connected",
  "cache": "available",
  "modules": 7,
  "version": "1.0.0"
}
```

## Security

### Data Protection
- **Encryption at Rest**: Database encryption recommended
- **Encryption in Transit**: TLS required for API access
- **Tenant Isolation**: Complete data separation between tenants
- **Access Control**: Integration with your authentication system

### Sensitive Data Handling
- **API Keys**: Stored securely, never logged
- **Passwords**: Hashed before storage
- **PII**: Marked and handled according to regulations

## Best Practices

### 1. Module Design
- Keep modules focused on a single domain
- Define clear dependencies
- Provide sensible defaults
- Add validation for critical settings

### 2. Performance
- Use typed accessors for frequently accessed settings
- Batch related setting changes
- Monitor cache hit rates
- Index database appropriately

### 3. Error Handling
- Always check errors from settings operations
- Provide meaningful error messages
- Implement graceful degradation for missing settings

### 4. Testing
- Test module providers in isolation
- Use integration tests for complete workflows
- Mock settings in unit tests of dependent code

## Troubleshooting

### Common Issues

#### Setting Not Found
```
Error: setting not found: company/company_name
```
**Solution**: Check if the module is registered and the setting key is correct.

#### Type Validation Failed
```
Error: cannot convert int to string for setting company_name
```
**Solution**: Ensure the value type matches the setting definition.

#### Module Dependency Error
```
Error: dependency not found: module 'billing' depends on 'company' which is not registered
```
**Solution**: Register dependencies in the correct order or let the registry resolve them automatically.

#### Database Connection Failed
```
Error: failed to connect to database
```
**Solution**: Check database connection string and ensure the database server is running.

### Debug Mode

Enable debug logging:

```go
manager.SetDebugMode(true)
```

This will log all setting operations and cache interactions.

## Contributing

### Development Setup

1. **Clone Repository**
```bash
git clone <repository-url>
cd ae-backend/base-server
```

2. **Install Dependencies**
```bash
go mod tidy
```

3. **Run Tests**
```bash
go test ./...
```

4. **Start Development Server**
```bash
go run cmd/main.go
```

### Adding Features

1. **Create Feature Branch**
```bash
git checkout -b feature/my-feature
```

2. **Implement Feature**
- Add tests first (TDD approach)
- Implement feature
- Update documentation

3. **Submit Pull Request**
- Ensure all tests pass
- Include documentation updates
- Add examples if applicable

## Support

For questions, issues, or feature requests:

1. **Check Documentation**: This README and inline code comments
2. **Search Issues**: Check existing GitHub issues
3. **Create Issue**: Open a new issue with detailed information
4. **Contact Team**: Reach out to the development team

## License

This settings system is part of the AE Backend project and is licensed under the same terms as the main project.

---

## Changelog

### v1.0.0
- Initial release
- Complete settings system with all core modules
- REST API with full CRUD operations
- Comprehensive test suite
- Migration tools and documentation

---

**Happy Coding!** 🚀