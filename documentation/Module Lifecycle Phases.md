# Module Lifecycle Phases
## Phase 1: Registration (in main.go)
Modules are registered in dependency order:

What happens: Module constructors create empty module structs. No initialization yet.

## Phase 2: Initialize() - Service Creation & Registration
Modules are initialized in topological (dependency) order. This is where most setup happens:

When: After all modules registered, before migrations
Order: Dependencies first (e.g., base → email → client-management)

What to do in Initialize():

✅ Create services (inject DB from context)
✅ Retrieve dependencies from service registry (ctx.Services.Get("email-service"))
✅ Register your own services (via Services() method or manual registry)
✅ Initialize handlers (with services)
✅ Setup route providers
✅ Seed data (like your billing settings seed)
❌ Don't start background workers yet

```go
func (m *CoreModule) Initialize(ctx core.ModuleContext) error {
    // Get dependencies from registry
    emailSvc, _ := ctx.Services.Get("email-service")
    
    // Create your services
    invoiceService := services.NewInvoiceService(ctx.DB)
    
    // Initialize handlers
    m.invoiceHandler = handlers.NewInvoiceHandler(invoiceService)
    
    // Seed definitions (NOT data)
    settingsSeedService.SeedBillingSettings()
    
    return nil
}
```

## Phase 3: Migration - Database Schema
When: After all modules initialized
What happens: System calls Entities() on each module and auto-migrates

## Phase 4: Contract Registration (Template System)
When: After migrations, before seeding
What: Template contracts are registered (invoice-contract.json)

## Phase 5: Database Seeding
When: After contracts registered
What: Seed initial data (users, templates, etc.)

## Phase 6: Start() - Runtime Services
When: After all initialization complete
What to do:

✅ Start background workers
✅ Start schedulers/cron jobs
✅ Final cross-module wiring (see base module example)

```go
func (m *BaseModule) Start(ctx context.Context) error {
    // Example: Set module registry in auth handlers
    // This needs to be in Start() because it happens 
    // after all Initialize() calls
    m.authHandlers.SetModuleRegistry(m.moduleContext.ModuleRegistry)
    return nil
}
```

## Phase 7: Routes Registration
When: During Initialize(), but routes are registered AFTER Initialize() completes
How: Automatic via Routes() method

## Phase 8: Event Handlers Registration
When: During Initialize(), handlers auto-subscribe to event bus

## Phase 9: Stop() - Graceful Shutdown
When: Application shutdown (reverse dependency order)
What to do:

✅ Stop background workers
✅ Close connections
✅ Cleanup resources

## Key Rules
✅ DO in Initialize()
- Create services
- Get dependencies from ctx.Services
- Register services
- Initialize handlers
- Setup route providers
- Seed definitions (settings schemas, not data)
- Setup event handlers

❌ DON'T in Initialize()
- Start background goroutines
- Use other modules' services (they might not be initialized yet)
- Start schedulers/timers
- Make HTTP requests

✅ DO in Start()
- Start background workers
- Start schedulers
- Final cross-module wiring
- Launch monitoring services

🔄 Dependency Resolution
Services registered during Initialize() are available immediately for modules initialized after you. Use Dependencies() to enforce order:

```go
func (m *CoreModule) Dependencies() []string {
    return []string{"base", "email", "booking", "audit"}
}
```

## Summary
main.go
  ↓
1. Register modules (NewCoreModule())
  ↓
2. Initialize core services (DB, Router, EventBus)
  ↓
3. Initialize modules in dependency order
   ├─ base.Initialize()
   ├─ email.Initialize() 
   ├─ client-management.Initialize()
   │   ├─ Create services
   │   ├─ Get email-service from registry
   │   ├─ Setup handlers
   │   └─ Seed billing settings
   ↓
4. Run migrations (all Entities())
  ↓
5. Register contracts
  ↓
6. Seed database
  ↓
7. Start modules
   ├─ base.Start()
   ├─ email.Start()
   ├─ client-management.Start()
  ↓
8. Start HTTP server
  ↓
[Running]
  ↓
9. Graceful shutdown
   └─ Stop modules (reverse order)