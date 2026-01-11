# Migration Guide: From Organization Settings to Advanced Settings System

This guide shows how to migrate existing code from using the organization entity for settings to the new Advanced Settings System.

## Overview of Changes

### Before: Organization-Based Settings ❌
```go
type Organization struct {
    ID               uint   `gorm:"primaryKey"`
    CompanyName      string `gorm:"size:255"`
    InvoicePrefix    string `gorm:"size:10"`
    NextInvoiceNumber int   `gorm:"default:1000"`
    PaymentTermsdays int   `gorm:"default:30"`
    VATNumber        string `gorm:"size:50"`
    DefaultCurrency  string `gorm:"size:3;default:EUR"`
    // ... mixed concerns in one entity
}
```

### After: Modular Settings System ✅
```go
// Settings stored as individual, typed records
// Domains: company, invoicing, billing, localization, etc.
// Type-safe access with caching
```

## Step-by-Step Migration

### 1. Update Service Dependencies

#### Before:
```go
type InvoiceService struct {
    db *gorm.DB
}

func NewInvoiceService(db *gorm.DB) *InvoiceService {
    return &InvoiceService{db: db}
}
```

#### After:
```go
type InvoiceService struct {
    db              *gorm.DB
    settingsManager *manager.SettingsManager
}

func NewInvoiceService(db *gorm.DB, settingsManager *manager.SettingsManager) *InvoiceService {
    return &InvoiceService{
        db:              db,
        settingsManager: settingsManager,
    }
}
```

### 2. Replace Organization Queries with Settings Access

#### Before:
```go
func (s *InvoiceService) generateInvoiceNumber(orgID uint) (string, error) {
    var org Organization
    err := s.db.Where("id = ?", orgID).First(&org).Error
    if err != nil {
        return "", err
    }
    
    prefix := org.InvoicePrefix
    if prefix == "" {
        prefix = "INV"
    }
    
    nextNumber := org.NextInvoiceNumber
    org.NextInvoiceNumber++
    s.db.Save(&org)
    
    return fmt.Sprintf("%s-%04d", prefix, nextNumber), nil
}
```

#### After:
```go
func (s *InvoiceService) generateInvoiceNumber(ctx context.Context, tenantID, organizationID uint) (string, error) {
    // Get invoice settings accessor
    invoiceAccessor, err := s.settingsManager.GetModuleAccessor("invoice")
    if err != nil {
        return "", fmt.Errorf("failed to get invoice settings: %w", err)
    }

    // Get settings with fallback defaults
    prefix, err := invoiceAccessor.GetString("invoice_prefix")
    if err != nil {
        prefix = "INV" // Fallback
    }

    nextNumber, err := invoiceAccessor.GetInt("next_invoice_number")
    if err != nil {
        nextNumber = 1000 // Fallback
    }

    // Atomic increment
    if err := invoiceAccessor.SetInt("next_invoice_number", nextNumber+1); err != nil {
        return "", fmt.Errorf("failed to increment: %w", err)
    }

    return fmt.Sprintf("%s-%04d", prefix, nextNumber), nil
}
```

### 3. Update Handler Dependencies

#### Before:
```go
type InvoiceHandler struct {
    service *InvoiceService
}

func NewInvoiceHandler(service *InvoiceService) *InvoiceHandler {
    return &InvoiceHandler{service: service}
}
```

#### After:
```go
type InvoiceHandler struct {
    service         *InvoiceService
    settingsManager *manager.SettingsManager
}

func NewInvoiceHandler(service *InvoiceService, settingsManager *manager.SettingsManager) *InvoiceHandler {
    return &InvoiceHandler{
        service:         service,
        settingsManager: settingsManager,
    }
}
```

### 4. Create Settings Modules

Create dedicated settings providers for each domain:

```go
// modules/invoice/settings/invoice_settings.go
package invoice

import (
    "github.com/ae-base-server/pkg/settings/accessor"
    "github.com/ae-base-server/pkg/settings/entities"
)

type InvoiceSettingsProvider struct {
    accessor *accessor.SettingsAccessor
}

func (p *InvoiceSettingsProvider) GetSettingsSchema() entities.ModuleSettingsSchema {
    return entities.ModuleSettingsSchema{
        ModuleName:  "invoice",
        Domain:      entities.DomainInvoicing,
        Settings: []entities.SettingDefinition{
            {
                Key:          "invoice_prefix",
                Type:         entities.SettingTypeString,
                DefaultValue: "INV",
                Required:     true,
            },
            {
                Key:          "next_invoice_number",
                Type:         entities.SettingTypeInt,
                DefaultValue: 1000,
                Required:     true,
            },
            // ... more settings
        },
    }
}

// Typed getters
func (p *InvoiceSettingsProvider) GetInvoicePrefix() (string, error) {
    return p.accessor.GetString("invoice_prefix")
}

func (p *InvoiceSettingsProvider) GetNextInvoiceNumber() (int, error) {
    return p.accessor.GetInt("next_invoice_number")
}
```

### 5. Update Main Application

#### Before:
```go
func main() {
    db := initDatabase()
    
    invoiceService := services.NewInvoiceService(db)
    invoiceHandler := handlers.NewInvoiceHandler(invoiceService)
    
    router := setupRoutes(invoiceHandler)
    router.Run(":8080")
}
```

#### After:
```go
func main() {
    db := initDatabase()
    
    // Initialize settings system
    settingsRepo := repositories.NewGORMSettingsRepository(db)
    settingsManager := manager.NewSettingsManager(settingsRepo)
    
    // Register modules
    invoiceProvider := invoice.NewInvoiceSettingsProvider()
    settingsManager.RegisterModule(invoiceProvider, tenantID, orgID)
    
    // Initialize services with settings
    invoiceService := services.NewInvoiceService(db, settingsManager)
    invoiceHandler := handlers.NewInvoiceHandler(invoiceService, settingsManager)
    
    router := setupRoutes(invoiceHandler, settingsManager)
    router.Run(":8080")
}
```

### 6. Update API Endpoints

Add settings management endpoints:

```go
func setupRoutes(invoiceHandler *InvoiceHandler, settingsManager *manager.SettingsManager) *gin.Engine {
    router := gin.Default()
    
    api := router.Group("/api/v1")
    
    // Settings routes
    settingsHandler := handlers.NewSettingsHandler(settingsManager)
    settingsHandler.RegisterRoutes(api)
    
    // Invoice routes
    invoices := api.Group("/invoices")
    invoices.POST("/", invoiceHandler.CreateInvoice)
    invoices.POST("/auto-number", invoiceHandler.CreateInvoiceWithAutoNumber) // New endpoint
    
    return router
}
```

## Data Migration

### 1. Run Database Migration

The settings system includes automatic data migration:

```go
import "github.com/ae-base-server/pkg/settings/migrations"

// This migrates existing organization data to settings
migrations.Migration20241220002_MigrateOrganizationSettings(db)
```

### 2. Manual Data Migration (if needed)

```sql
-- Migrate invoice prefix
INSERT INTO settings (tenant_id, organization_id, domain, key, value_type, string_value, created_at, updated_at)
SELECT 
    o.tenant_id,
    o.id as organization_id,
    'invoicing' as domain,
    'invoice_prefix' as key,
    'string' as value_type,
    o.invoice_prefix as string_value,
    NOW() as created_at,
    NOW() as updated_at
FROM organizations o 
WHERE o.invoice_prefix IS NOT NULL
ON CONFLICT (tenant_id, organization_id, domain, key) DO NOTHING;

-- Migrate next invoice number
INSERT INTO settings (tenant_id, organization_id, domain, key, value_type, int_value, created_at, updated_at)
SELECT 
    o.tenant_id,
    o.id as organization_id,
    'invoicing' as domain,
    'next_invoice_number' as key,
    'int' as value_type,
    o.next_invoice_number as int_value,
    NOW() as created_at,
    NOW() as updated_at
FROM organizations o 
WHERE o.next_invoice_number IS NOT NULL
ON CONFLICT (tenant_id, organization_id, domain, key) DO NOTHING;
```

## Testing the Migration

### 1. Before Migration Test
```go
func TestOldOrganizationSettings(t *testing.T) {
    // Old way - querying organization table
    var org Organization
    db.Where("id = ?", 1).First(&org)
    
    assert.Equal(t, "INV", org.InvoicePrefix)
    assert.Equal(t, 1000, org.NextInvoiceNumber)
}
```

### 2. After Migration Test
```go
func TestNewSettingsSystem(t *testing.T) {
    // New way - using settings accessor
    accessor, _ := settingsManager.GetModuleAccessor("invoice")
    
    prefix, err := accessor.GetString("invoice_prefix")
    assert.NoError(t, err)
    assert.Equal(t, "INV", prefix)
    
    nextNum, err := accessor.GetInt("next_invoice_number")
    assert.NoError(t, err)
    assert.Equal(t, 1000, nextNum)
}
```

## Benefits After Migration

### ✅ Improved Architecture
- **Single Responsibility**: Each domain handles its own settings
- **Type Safety**: Compile-time checking of setting access
- **Performance**: Built-in caching and optimized queries
- **Extensibility**: Easy to add new modules and settings

### ✅ Better Developer Experience
```go
// Type-safe access
prefix, err := invoiceAccessor.GetString("invoice_prefix")
nextNumber, err := invoiceAccessor.GetInt("next_invoice_number") 
autoSend, err := invoiceAccessor.GetBool("auto_send_invoice")

// Structured settings
var taxSettings TaxSettings
err := invoiceAccessor.GetJSON("tax_settings", &taxSettings)
```

### ✅ Better Operations
- **API Management**: Full REST API for settings
- **Database Editing**: Direct database editing is simple
- **Auditing**: Built-in change tracking
- **Validation**: Schema-based validation

## Migration Checklist

- [ ] Update service constructors to accept `SettingsManager`
- [ ] Replace organization queries with settings accessor calls  
- [ ] Create settings providers for each domain
- [ ] Register modules in main application
- [ ] Run database migration
- [ ] Add settings API routes
- [ ] Update tests to use new settings system
- [ ] Verify data migration completed correctly
- [ ] Update documentation and API examples
- [ ] (Optional) Remove old organization setting fields

## Common Pitfalls

1. **Forgetting to register modules**: Settings won't work if modules aren't registered
2. **Not handling fallback defaults**: Always provide fallback values for critical settings
3. **Mixing old and new systems**: Don't access organization directly after migration
4. **Missing tenant/org context**: Ensure proper tenant/organization isolation

## Rollback Plan

If issues arise, you can temporarily revert by:

1. Keeping both systems running in parallel
2. Using feature flags to switch between old/new systems
3. Maintaining organization table until migration is fully verified
4. Having database backup before running migrations

The new system is designed to be backwards-compatible and can coexist with the old system during migration.