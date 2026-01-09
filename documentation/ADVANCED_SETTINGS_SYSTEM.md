# Advanced Settings System Design

## Overview
A self-registering, module-driven settings system where each module defines its own settings schema, types, defaults, and dependencies. Settings are stored granularly (one database row per setting) for easy database editing.

## Core Principles

### 1. Module Self-Registration
- Each module registers its settings schema during initialization
- Modules define their domain name, setting definitions, and defaults
- Settings system auto-discovers module requirements

### 2. Granular Storage
- One database row per individual setting
- Easy to edit with any database viewer/admin tool
- Type information stored with each setting
- Hierarchical setting keys (e.g., `billing.unit_price`, `billing.extra_efforts.mode`)

### 3. Type Safety & Validation
- Settings have defined types (string, int, float, bool, json, enum)
- Validation rules defined per setting
- Default values with proper typing

### 4. Cross-Module Dependencies
- Modules can declare dependencies on other domains
- Settings system resolves dependencies during registration
- Circular dependency detection

## Refined Requirements

### R1: Module Settings Schema Definition
```go
type ModuleSettingsSchema struct {
    Domain      SettingsDomain              // e.g., "billing", "invoicing"
    Settings    []SettingDefinition         // Individual setting definitions
    Dependencies []SettingsDomain           // Other domains this module needs
    Version     string                      // Schema version for migrations
}

type SettingDefinition struct {
    Key          string      // Setting key (e.g., "unit_price", "extra_efforts.mode")
    Type         SettingType // string, int, float, bool, json, enum
    DefaultValue interface{} // Default value
    Required     bool        // Whether setting is required
    Validation   *ValidationRules // Validation rules
    Description  string      // Human-readable description
    EnumValues   []string    // Valid values for enum type
    Group        string      // UI grouping (optional)
}

type SettingType string
const (
    TypeString SettingType = "string"
    TypeInt    SettingType = "int" 
    TypeFloat  SettingType = "float"
    TypeBool   SettingType = "bool"
    TypeJSON   SettingType = "json"
    TypeEnum   SettingType = "enum"
)
```

### R2: Database Schema (Granular)
```sql
CREATE TABLE settings (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    organization_id INTEGER NOT NULL,
    domain VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,           -- Always stored as string, parsed by type
    type VARCHAR(20) NOT NULL,     -- string, int, float, bool, json, enum
    default_value TEXT,            -- Default value as string
    required BOOLEAN DEFAULT false,
    description TEXT,
    enum_values JSONB,             -- For enum types
    version VARCHAR(20),           -- Schema version
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, organization_id, domain, key)
);

CREATE INDEX idx_settings_tenant_org_domain ON settings(tenant_id, organization_id, domain);
CREATE INDEX idx_settings_domain_key ON settings(domain, key);
CREATE INDEX idx_settings_tenant_org ON settings(tenant_id, organization_id);
```

### R3: Module Registration Process
1. Module implements `SettingsProvider` interface
2. During module initialization, registers settings schema
3. Settings system checks if tenant has values for these settings
4. If no values exist, defaults are automatically created
5. Module gets typed settings accessor

### R4: Settings Access Pattern
```go
// In module initialization
func (m *InvoiceModule) RegisterSettings(registry *SettingsRegistry) {
    schema := ModuleSettingsSchema{
        Domain: "invoicing",
        Settings: []SettingDefinition{
            {
                Key: "default_vat_rate", 
                Type: TypeFloat,
                DefaultValue: 19.0,
                Required: true,
                Validation: &ValidationRules{Min: 0, Max: 100},
                Description: "Default VAT rate in percentage",
            },
            {
                Key: "number_format",
                Type: TypeEnum, 
                DefaultValue: "sequential",
                EnumValues: []string{"sequential", "year_prefix", "year_month_prefix"},
                Description: "Invoice number format",
            },
        },
        Dependencies: []SettingsDomain{"company"}, // Needs company info
    }
    
    registry.RegisterModule(m.GetName(), schema)
}

// In module usage
vatRate := m.settingsAccessor.GetFloat("default_vat_rate")
numberFormat := m.settingsAccessor.GetString("number_format")
```

## Architecture Components

### 1. Settings Registry
Central registry that modules register with during initialization.

### 2. Settings Definition Store
Stores schema definitions for all modules, handles versioning and migrations.

### 3. Settings Value Store  
Manages actual setting values, handles tenant isolation, type conversion.

### 4. Settings Accessor
Provides type-safe access to settings for modules.

### 5. Settings Migration Engine
Handles schema version changes and data migrations.

## Implementation Plan

### Phase 1: Core Infrastructure
- [ ] Create settings database schema
- [ ] Implement SettingsRegistry
- [ ] Create SettingDefinition and validation framework
- [ ] Build basic SettingsAccessor with type conversion

### Phase 2: Module Integration
- [ ] Define SettingsProvider interface
- [ ] Implement module registration system  
- [ ] Create default value provisioning
- [ ] Add dependency resolution

### Phase 3: Advanced Features
- [ ] Schema versioning and migration system
- [ ] Admin UI for settings management
- [ ] Settings validation and error handling
- [ ] Cross-tenant settings inheritance

### Phase 4: Module Migration
- [ ] Migrate invoice module to new system
- [ ] Migrate billing module to new system
- [ ] Migrate booking module to new system
- [ ] Remove old organization-based settings

## Detailed Implementation TODOs

### TODO 1: Database Layer
**Priority: High**
**Estimated Time: 2-3 hours**

- [ ] Create settings table migration
- [ ] Implement SettingsRepository with CRUD operations
- [ ] Add type-safe value conversion (string <-> typed values)
- [ ] Create database indexes for performance
- [ ] Add tenant and organization isolation checks

```go
type SettingsRepository interface {
    GetSetting(tenantID, organizationID uint, domain, key string) (*Setting, error)
    SetSetting(tenantID, organizationID uint, setting *Setting) error
    GetDomainSettings(tenantID, organizationID uint, domain string) ([]*Setting, error)
    BulkSetSettings(tenantID, organizationID uint, settings []*Setting) error
    DeleteSetting(tenantID, organizationID uint, domain, key string) error
}
```

### TODO 2: Settings Registry & Schema Management
**Priority: High** 
**Estimated Time: 4-5 hours**

- [ ] Create ModuleSettingsSchema structs
- [ ] Implement SettingsRegistry with module registration
- [ ] Add schema validation and conflict detection
- [ ] Create dependency resolution algorithm
- [ ] Add circular dependency detection

```go
type SettingsRegistry interface {
    RegisterModule(moduleName string, schema ModuleSettingsSchema) error
    GetModuleSchema(moduleName string) (*ModuleSettingsSchema, error)
    GetDomainSchema(domain SettingsDomain) (*ModuleSettingsSchema, error)
    ValidateSchemas() error
    ResolveDependencies() ([]SettingsDomain, error)
}
```

### TODO 3: Type-Safe Settings Accessor
**Priority: High**
**Estimated Time: 3-4 hours**

- [ ] Create SettingsAccessor with type conversion
- [ ] Implement caching for performance
- [ ] Add change notification system
- [ ] Create batch setting operations
- [ ] Add validation on write operations

```go
type SettingsAccessor interface {
    GetString(key string) string
    GetInt(key string) int
    GetFloat(key string) float64
    GetBool(key string) bool
    GetJSON(key string, target interface{}) error
    SetString(key, value string) error
    SetInt(key string, value int) error
    SetFloat(key string, value float64) error
    SetBool(key string, value bool) error
    SetJSON(key string, value interface{}) error
    Reload() error // Refresh cache
}
```

### TODO 4: Default Value Provisioning
**Priority: Medium**
**Estimated Time: 2-3 hours**

- [ ] Create default value seeding on module registration
- [ ] Handle tenant and organization-specific vs global defaults
- [ ] Add default value override system
- [ ] Create bulk default provisioning for new tenants and organizations
- [ ] Add default value migration on schema changes

### TODO 5: Module Provider Interface
**Priority: Medium**
**Estimated Time: 2-3 hours**

- [ ] Define SettingsProvider interface for modules
- [ ] Create module initialization hooks
- [ ] Add settings accessor injection
- [ ] Implement module dependency injection
- [ ] Create module registration validation

```go
type SettingsProvider interface {
    GetSettingsSchema() ModuleSettingsSchema
    OnSettingsRegistered(accessor SettingsAccessor) error
    OnSettingsChanged(key string, oldValue, newValue interface{}) error
}
```

### TODO 6: Validation Framework
**Priority: Medium**
**Estimated Time: 3-4 hours**

- [ ] Create ValidationRules struct with common rules
- [ ] Implement validation engine
- [ ] Add custom validation functions
- [ ] Create validation error reporting
- [ ] Add real-time validation on settings change

```go
type ValidationRules struct {
    Min          *float64  // For numeric types
    Max          *float64  // For numeric types  
    MinLength    *int      // For string types
    MaxLength    *int      // For string types
    Pattern      *string   // Regex pattern for strings
    CustomFunc   func(interface{}) error // Custom validation
}
```

### TODO 7: Settings Migration & Versioning
**Priority: Low**
**Estimated Time: 4-5 hours**

- [ ] Add schema versioning system
- [ ] Create migration framework for schema changes
- [ ] Handle setting key renames and type changes
- [ ] Add rollback capabilities
- [ ] Create migration testing framework

### TODO 8: Admin Interface & API
**Priority: Low**
**Estimated Time: 3-4 hours**

- [ ] Create REST API for settings management
- [ ] Add settings UI for database editing
- [ ] Create bulk import/export functionality  
- [ ] Add settings history and auditing
- [ ] Create settings comparison tools

### TODO 9: Module Migrations
**Priority: Medium**
**Estimated Time: 6-8 hours**

#### Invoice Module Migration
- [ ] Define invoice settings schema
- [ ] Create InvoiceSettingsProvider
- [ ] Update invoice service to use new accessor
- [ ] Migrate existing organization data
- [ ] Remove invoice settings from organization

#### Billing Module Migration  
- [ ] Define billing settings schema
- [ ] Create BillingSettingsProvider
- [ ] Update billing logic to use new accessor
- [ ] Handle extra efforts configuration migration
- [ ] Test billing mode calculations

#### Booking Module Migration
- [ ] Define booking settings schema
- [ ] Migrate existing BookingTemplate to new system
- [ ] Update booking service integration
- [ ] Handle weekly availability configuration
- [ ] Test booking template functionality

### TODO 10: Performance & Optimization
**Priority: Low**
**Estimated Time: 2-3 hours**

- [ ] Implement settings caching strategy
- [ ] Add cache invalidation on updates
- [ ] Create settings preloading for modules
- [ ] Add database connection pooling
- [ ] Performance testing and optimization

### TODO 11: Testing & Quality Assurance
**Priority: High**
**Estimated Time: 4-6 hours**

- [ ] Unit tests for all settings components
- [ ] Integration tests for module registration
- [ ] Performance tests for settings access
- [ ] Migration testing with real data
- [ ] Error handling and edge case testing

## Example Module Implementation

### Invoice Module Settings Schema
```go
func (m *InvoiceModule) GetSettingsSchema() ModuleSettingsSchema {
    return ModuleSettingsSchema{
        Domain: "invoicing",
        Version: "1.0.0",
        Dependencies: []SettingsDomain{"company", "localization"},
        Settings: []SettingDefinition{
            {
                Key: "default_vat_rate",
                Type: TypeFloat,
                DefaultValue: 19.0,
                Required: true,
                Validation: &ValidationRules{Min: 0, Max: 100},
                Description: "Default VAT rate in percentage",
                Group: "tax",
            },
            {
                Key: "default_vat_exempt", 
                Type: TypeBool,
                DefaultValue: false,
                Description: "Default VAT exemption status",
                Group: "tax",
            },
            {
                Key: "vat_exemption_text",
                Type: TypeString,
                DefaultValue: "Umsatzsteuerfrei gemäß §4 Nr. 14 UStG",
                Description: "Default VAT exemption text",
                Group: "tax",
            },
            {
                Key: "number_format",
                Type: TypeEnum,
                DefaultValue: "sequential", 
                EnumValues: []string{"sequential", "year_prefix", "year_month_prefix"},
                Description: "Invoice number format",
                Group: "numbering",
            },
            {
                Key: "number_prefix",
                Type: TypeString,
                DefaultValue: "INV-",
                Validation: &ValidationRules{MaxLength: 20},
                Description: "Invoice number prefix",
                Group: "numbering",
            },
            {
                Key: "payment_due_days",
                Type: TypeInt,
                DefaultValue: 14,
                Validation: &ValidationRules{Min: 1, Max: 365},
                Description: "Payment due days",
                Group: "payment",
            },
            {
                Key: "first_reminder_days",
                Type: TypeInt, 
                DefaultValue: 7,
                Validation: &ValidationRules{Min: 1, Max: 365},
                Description: "First reminder days after due date",
                Group: "payment",
            },
            {
                Key: "second_reminder_days",
                Type: TypeInt,
                DefaultValue: 14, 
                Validation: &ValidationRules{Min: 1, Max: 365},
                Description: "Second reminder days after first reminder",
                Group: "payment",
            },
        },
    }
}
```

## Benefits of This Approach

### 1. **Module Autonomy**
- Each module owns its settings definition
- Modules can evolve settings independently
- Clear boundaries and responsibilities

### 2. **Database Transparency** 
- One row per setting = easy DB editing
- No JSON parsing needed for simple edits
- Standard SQL operations work

### 3. **Type Safety**
- Compile-time type checking
- Runtime validation
- Automatic type conversion

### 4. **Self-Documenting**
- Settings include descriptions
- Schema acts as documentation
- Clear dependency relationships

### 5. **Maintainability**
- Centralized settings management
- Automatic default provisioning
- Schema versioning for migrations

This approach provides the flexibility and maintainability you're looking for while keeping the system easy to understand and extend.