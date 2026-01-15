# Module Lifecycle & Initialization Issues

## Problem Pattern

We have encountered the same initialization/seeding issue multiple times with different module features:

1. **Template Contracts**: Module needed to seed template contract definitions on startup
2. **Settings Definitions**: Module needed to seed setting definitions on startup

**Both failed silently** - the code was written, but never executed during application startup.

## Root Cause Analysis

### Dual Module System Problem

The application has **TWO parallel module initialization systems**:

#### 1. Legacy System (Currently Active in unburdy_server)
- Location: `main.go` uses old `NewModule()` constructor
- No lifecycle hooks (no `Initialize()`, `Start()`, `Stop()`)
- Module just registers routes via `RegisterRoutes()`
- **Problem**: No place to run startup code like seeding

#### 2. Bootstrap System (Designed but not used in unburdy_server)
- Location: `NewCoreModule()` constructor returns `core.Module` interface
- Has lifecycle methods: `Initialize()`, `Start()`, `Stop()`
- Used by base-server modules
- **Problem**: unburdy_server doesn't use this system

### Code Evidence

#### unburdy_server/modules/client_management/module.go has BOTH:

```go
// OLD LEGACY SYSTEM - Actually used
func NewModule(db *gorm.DB) baseAPI.ModuleRouteProvider {
    // ... creates services and handlers
    return &Module{routeProvider: routeProvider}
}

func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
    // Just registers routes - NO INITIALIZATION LOGIC
}

// NEW BOOTSTRAP SYSTEM - Never called!
func NewCoreModule() core.Module {
    return &CoreModule{}
}

func (m *CoreModule) Initialize(ctx core.ModuleContext) error {
    // ... this is where seeding code lives
    // Seed billing settings definitions
    settingsSeedService := clientServices.NewSettingsSeedService(ctx.DB)
    if err := settingsSeedService.SeedBillingSettings(); err != nil {
        // ...
    }
    // THIS CODE NEVER RUNS!
}
```

#### main.go registers module the old way:

```go
// main.go line ~100
clientManagementModule := client_management.NewModule(db)  // OLD SYSTEM
// ... later:
clientManagementModule.RegisterRoutes(apiV1)  // Just routes, no init
```

The `NewCoreModule()` constructor and its `Initialize()` method are **dead code** in unburdy_server!

## Historical Context

### Previous Occurrence: Template Contracts

Same pattern happened with template contract registration:
- Code existed in module to register contracts on startup
- Was placed in lifecycle methods that were never called
- Had to manually create contract files or run separate seeding scripts
- Silently failed - no errors, just missing data

### Current Occurrence: Settings Definitions

Exact same pattern:
- `SeedBillingSettings()` code exists in `CoreModule.Initialize()`
- Added debug logging with emojis to make it obvious
- Server starts successfully with no errors
- **Zero** of the debug messages appear in logs
- Settings definitions table remains empty
- Manual SQL insertion required as workaround

## Impact & Symptoms

### User-Facing Symptoms
1. Missing data that should be seeded automatically
2. 404 errors when trying to access features that depend on seeded data
3. "Definition not found" errors
4. Manual database intervention required

### Developer Symptoms
1. Code written but never executed
2. No errors or warnings - fails silently
3. Hard to debug because the code "looks right"
4. Confusion about why initialization isn't running
5. Time wasted writing code in the wrong lifecycle hooks

### Database Symptoms
```sql
-- Expected after startup:
SELECT COUNT(*) FROM setting_definitions WHERE domain = 'billing';
-- Result: 5 rows

-- Actual after startup:
SELECT COUNT(*) FROM setting_definitions WHERE domain = 'billing';
-- Result: 0 rows
```

## Current Workarounds

### Settings Definitions
Manual SQL seeding required after each fresh database:
```sql
INSERT INTO setting_definitions (domain, key, version, schema, data) 
VALUES 
  ('billing', 'mode', 1, '{}'::jsonb, '{}'::jsonb),
  ('billing', 'invoice_items', 1, '{}'::jsonb, '{}'::jsonb),
  ('billing', 'invoice_number', 1, '{}'::jsonb, '{}'::jsonb),
  ('billing', 'payment_terms', 1, '{}'::jsonb, '{}'::jsonb),
  ('billing', 'tax', 1, '{}'::jsonb, '{}'::jsonb)
ON CONFLICT (domain, key) DO NOTHING;
```

### Template Contracts
Manual file creation or separate seeding scripts.

## Why This Keeps Happening

### 1. Confusion About Which System to Use
- Developers see both `NewModule()` and `NewCoreModule()`
- Bootstrap system (with lifecycle) looks more "modern"
- Naturally write initialization code in `Initialize()`
- But unburdy_server uses legacy system!

### 2. No Clear Documentation
- No docs explaining which system to use when
- No migration guide from legacy to bootstrap
- Module interfaces are similar enough to be confusing

### 3. Silent Failures
- Application starts successfully
- No errors logged
- Missing functionality only discovered during testing
- Debug logging in dead code paths is never seen

### 4. Inconsistent Module Architecture
- base-server modules: Use bootstrap system
- unburdy_server modules: Use legacy system
- Same codebase, different patterns

## Architecture Comparison

### Base-Server Modules (Bootstrap System)
```
main.go
  └─> bootstrap.NewApplication()
      └─> app.Initialize()
          └─> For each module:
              ├─> module.Initialize(ctx)  ← Seeding happens here
              ├─> module.Start(ctx)
              └─> Routes registered automatically
```

### Unburdy-Server Modules (Legacy System)
```
main.go
  └─> NewModule(db)                      ← Just creates module
  └─> module.RegisterRoutes(router)      ← Just registers routes
  └─> [No initialization phase!]
```

## Proposed Solutions

### Option 1: Migrate unburdy_server to Bootstrap System ⭐ RECOMMENDED

**Pros:**
- Unified architecture across all modules
- Proper lifecycle management
- Seeding happens automatically
- Consistent with base-server patterns
- Future-proof

**Cons:**
- Significant refactoring required
- Need to test all existing modules
- Migration effort

**Implementation:**
```go
// main.go
func main() {
    config := config.LoadConfig()
    app := bootstrap.NewApplication(config)
    
    // Register modules using bootstrap system
    app.RegisterModule(base.NewCoreModule())
    app.RegisterModule(client_management.NewCoreModule())
    // ... etc
    
    // This will call Initialize() on all modules
    if err := app.Initialize(); err != nil {
        log.Fatal(err)
    }
    
    app.Start()
}
```

### Option 2: Add Explicit Seeding Phase to Legacy System

**Pros:**
- Minimal code changes
- Works with existing architecture
- Quick to implement

**Cons:**
- Perpetuates dual-system problem
- Still requires manual tracking of what needs seeding
- Not scalable

**Implementation:**
```go
// main.go after module creation
func main() {
    // ... create modules ...
    
    // Explicit seeding phase
    log.Println("Running module seeders...")
    if err := clientManagementModule.SeedData(db); err != nil {
        log.Printf("Warning: failed to seed client management: %v", err)
    }
    // ... seed other modules ...
    
    // Register routes
    clientManagementModule.RegisterRoutes(apiV1)
}
```

### Option 3: Hybrid - Seeding Service Registry

**Pros:**
- Works with either system
- Centralized seeding management
- Easy to add new seeders
- Good for gradual migration

**Cons:**
- Adds another layer of abstraction
- Still doesn't solve dual-system confusion

**Implementation:**
```go
// pkg/seeding/registry.go
type Seeder interface {
    Seed(db *gorm.DB) error
    Name() string
}

type SeedRegistry struct {
    seeders []Seeder
}

func (r *SeedRegistry) Register(s Seeder) {
    r.seeders = append(r.seeders, s)
}

func (r *SeedRegistry) SeedAll(db *gorm.DB) error {
    for _, s := range r.seeders {
        log.Printf("Running seeder: %s", s.Name())
        if err := s.Seed(db); err != nil {
            return fmt.Errorf("seeder %s failed: %w", s.Name(), err)
        }
    }
    return nil
}

// main.go
func main() {
    registry := seeding.NewRegistry()
    registry.Register(client_management.NewSettingsSeeder())
    registry.Register(templates.NewTemplateSeeder())
    
    if err := registry.SeedAll(db); err != nil {
        log.Fatal(err)
    }
}
```

## Recommendations

### Immediate (Emergency Fix)
1. ✅ Document manual SQL seeding steps
2. ✅ Add to README.md or setup scripts
3. Create database migration files for definitions
4. Add health check endpoint that validates seeded data

### Short-term (This Sprint)
1. **Choose ONE system** - document the decision
2. Update all module templates/examples
3. Add linting or tests to catch dead code in unused lifecycle methods
4. Create seeding service registry (Option 3)

### Long-term (Next Quarter)
1. Migrate unburdy_server to bootstrap system (Option 1)
2. Deprecate legacy module system
3. Update documentation with clear patterns
4. Add automated tests for module initialization

## Detection & Prevention

### How to Detect This Problem

**1. Module has seeding code but data isn't seeded:**
```bash
# Check if Initialize() is actually called
grep -r "Initialize.*client" logs/ 
# If empty → method not called

# Check database
psql -c "SELECT COUNT(*) FROM setting_definitions;"
# If 0 → seeding didn't run
```

**2. Debug logging doesn't appear:**
```go
// Add obvious logging
fmt.Println("🌱🌱🌱 SEEDING STARTS HERE 🌱🌱🌱")
// If not in logs → dead code path
```

**3. Code exists in NewCoreModule but app uses NewModule:**
```bash
# Check main.go for which constructor is used
grep "NewModule\|NewCoreModule" main.go
```

### Prevention Measures

**1. Module Template Checklist:**
- [ ] Which initialization system does this app use?
- [ ] Where should seeding code go?
- [ ] Does the module's Initialize() method actually get called?
- [ ] Is there a test that validates seeding?

**2. Startup Validation:**
```go
// Add to application startup
func ValidateSeedingComplete(db *gorm.DB) error {
    var count int64
    db.Model(&entities.SettingDefinition{}).
        Where("domain = ?", "billing").
        Count(&count)
    
    if count == 0 {
        return errors.New("CRITICAL: Billing settings not seeded!")
    }
    return nil
}
```

**3. CI/CD Checks:**
```yaml
# .github/workflows/test.yml
- name: Verify seeding
  run: |
    ./app &
    sleep 5
    psql -c "SELECT COUNT(*) FROM setting_definitions" | grep -q "5"
```

**4. Development Environment Check:**
```go
// In development mode, panic if seeding incomplete
if config.IsDevelopment() {
    if err := ValidateSeedingComplete(db); err != nil {
        panic(err) // Force developer to notice
    }
}
```

## Related Documentation
- `SETTINGS_SYSTEM_README.md` - Settings architecture
- `TEMPLATE_CONTRACT_SYSTEM.md` - Template system
- `Architecture.md` - Overall system architecture
- `MODULE_DEVELOPMENT_GUIDE.md` - Module development patterns (needs update)

## Action Items
- [ ] Decide on Option 1, 2, or 3
- [ ] Update MODULE_DEVELOPMENT_GUIDE.md with clear guidance
- [ ] Create database migration for settings definitions
- [ ] Add startup validation for critical seeded data
- [ ] Document manual seeding steps in README
- [ ] Add tests for module initialization
- [ ] Schedule migration to unified bootstrap system
