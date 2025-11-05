# AE Backend Architecture Guide

## Architecture Overview

This monorepo uses a **modular architecture** where:
- **base-server**: Provides core SaaS functionality (auth, users, customers, plans, etc.)
- **modules/**: Contains reusable business logic modules
- *_server/**: Contains specific applications that combine base + modules

## Directory Structure

```
ae-backend/
├── go.work                   # Go workspace configuration  
├── base-server/             # Core SaaS foundation
│   ├── api/                 # Public API exports
│   ├── internal/            # Private implementation
│   └── go.mod
├── modules/                 # Reusable business modules
│   └── calendar/            # Calendar system 
├── unburdy_server/          # Unburdy application (main service)
│   ├── main.go             # Application entry point
│   ├── modules/            # Service-specific modules
│   │   └── client_management/ # Client and cost provider 
│
```

## Key Benefits of This Architecture

### 1. **Modular Design**
- Each module is independent and testable
- Modules can be mixed and matched per application
- Clear separation of concerns

### 2. **Authentication Integration**
- All modules automatically use base-server authentication
- Automatic tenant isolation through TenantID
- Consistent auth patterns across modules

### 3. **Easy Module Integration**
Just 3 lines to add any module to your application:
```go
// 1. Import the module
import "github.com/ae-backend/calendar-module"

// 2. Run migrations
calendar.MigrateCalendar(db)

// 3. Register routes
calendar.RegisterCalendarRoutes(protectedRoutes, db)
```

### 4. **Tenant Isolation Built-in**
Every module automatically respects tenant boundaries:
```go
// All queries automatically filtered by Tenant
user, _ := api.GetUser(c)
query := db.Where("tenant_id = ?", user.TenantID)
```

## Module Development Pattern

This project uses a **modular architecture** based on the `core.Module` interface.
Each feature (e.g., Calendar, Users, Tasks) is implemented as a self-contained module that registers its database entities, routes, and services automatically during initialization.

### 1. **Module Definition** (`module.go`)

Each module implements the `core.Module` interface (and optionally `core.ModuleWithEntities`) to integrate with the system lifecycle:

```go
type Module struct {
    db              *gorm.DB
    routeProvider   *routes.RouteProvider
    calendarService *services.CalendarService
    calendarHandler *handlers.CalendarHandler
}

func (m *Module) Name() string         { return "calendar" }
func (m *Module) Version() string      { return "1.0.0" }
func (m *Module) Dependencies() []string { return []string{"base"} }
```

Modules can be created manually or with **auto-migration** support:

```go
func NewModuleWithAutoMigration(db *gorm.DB) *Module {
    calendarService := services.NewCalendarService(db)
    calendarHandler := handlers.NewCalendarHandler(calendarService)
    routeProvider := routes.NewRouteProvider(calendarHandler, db)

    return &Module{
        db:              db,
        routeProvider:   routeProvider,
        calendarService: calendarService,
        calendarHandler: calendarHandler,
    }
}
```

### 2. **Lifecycle Methods**

Each module participates in the application lifecycle:

```go
func (m *Module) Initialize(ctx core.ModuleContext) error {
    m.db = ctx.DB
    m.calendarService = services.NewCalendarService(ctx.DB)
    m.calendarHandler = handlers.NewCalendarHandler(m.calendarService)
    m.routeProvider = routes.NewRouteProvider(m.calendarHandler, ctx.DB)
    return nil
}

func (m *Module) Start(ctx context.Context) error { return nil }
func (m *Module) Stop(ctx context.Context) error  { return nil }
```

### 3. **Entities (Database Models)**

Modules expose their database entities for **automatic migration**:

```go
func (m *Module) Entities() []core.Entity {
    return []core.Entity{
        entities.NewCalendarEntity(),
        entities.NewCalendarEntryEntity(),
        entities.NewCalendarSeriesEntity(),
        entities.NewExternalCalendarEntity(),
    }
}

func (m *Module) GetEntitiesForMigration() []interface{} {
    coreEntities := m.Entities()
    entities := make([]interface{}, len(coreEntities))
    for i, entity := range coreEntities {
        entities[i] = entity.GetModel()
    }
    return entities
}
```

### 4. **Handlers & Services**

Handlers depend on services, and services encapsulate business logic and database access.

```go
// services/calendar_service.go
type CalendarService struct {
    db *gorm.DB
}

// handlers/calendar_handler.go
type CalendarHandler struct {
    service *services.CalendarService
}
```

This separation keeps route logic thin and business logic reusable.

### 5. **Routes**

Each module exposes its own route provider that integrates with the base-server’s authentication and tenant isolation middleware:

```go
// routes/routes.go
func (p *RouteProvider) RegisterRoutes(router *gin.RouterGroup) {
    group := router.Group("/calendar")
    group.Use(api.AuthMiddleware(p.db))
    {
        group.POST("/", p.handler.CreateCalendar)
        group.GET("/", p.handler.GetCalendars)
    }
}
```

The module returns the provider to the core system:

```go
func (m *Module) Routes() []core.RouteProvider {
    return []core.RouteProvider{
        &calendarRouteAdapter{provider: m.routeProvider},
    }
}
```



### 6. **Swagger Integration**

Modules register their Swagger paths for automatic documentation:

```go
func (m *Module) SwaggerPaths() []string {
    return []string{
        "/calendar",
        "/calendar/{id}/import_holidays",
        "/calendar-entries",
        "/calendar-series",
        "/external-calendars",
    }
}
```



### Summary

| Layer        | Responsibility                                  |
|  | -- |
| **Entities** | Define GORM models for persistence              |
| **Services** | Contain business logic and DB access            |
| **Handlers** | Translate HTTP requests to service calls        |
| **Routes**   | Register endpoints and apply middleware         |
| **Module**   | Wires all parts, registers with the core system |

This pattern ensures **clear separation of concerns**, **automatic bootstrapping**, and **multi-tenant awareness** via the shared base module.


## How to Create a New Module

### Step 1: Create Module Structure
```bash
mkdir modules/your-module
cd modules/your-module
go mod init github.com/ae-backend/your-module
```

### Step 2: Add Base Dependencies
```go
// go.mod
require (
    github.com/ae-base-server v0.0.0
    github.com/gin-gonic/gin v1.10.1
    gorm.io/gorm v1.30.0
)
replace github.com/ae-base-server => ../../base-server
```

### Step 3: Implement Module
```go
// models.go - Define your data structures
// handlers.go - Implement business logic
// routes.go - Define API endpoints
// module.go - Export main functions
```

### Step 4: Use in Application
```go
// In your application's main.go or router
import "github.com/ae-backend/your-module"

// Migrate
yourmodule.Migrate(db)

// Register routes
yourmodule.RegisterRoutes(protectedGroup, db)
```

## Authentication & Authorization Patterns

### Automatic Authentication
```go
// Base middleware handles JWT validation
protected.Use(baseAPI.AuthMiddleware(db))

// Handlers automatically get authenticated user
user, err := api.GetUser(c)
if err != nil {
    // User not authenticated
    return
}
```

### Tenant Isolation
```go
// All data operations should include tenant filter
tenantID, _ := baseAPI.GetTenantID(c)
userID, _ := baseAPI.GetUserID(c)
db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Find(&records)
```

### Role-Based Access
```go
// Use base-server role middleware
adminGroup.Use(api.RequireAdmin())
managerGroup.Use(api.RequireRole("manager", "admin"))
```

## Testing Patterns

### Module Testing
```go
func TestCreateEvent(t *testing.T) {
    // Use base-server test database
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    // Create test user with Tenant
    user := createTestUser(db)
    
    // Test your module functionality
    handler := NewHandler(db)
    // ... test implementation
}
```

### Integration Testing
```go
func TestFullWorkflow(t *testing.T) {
    // Test base + module integration
    // Verify auth flows work
    // Verify tenant isolation works
}
```

## Building a SaaS App with the Minimal Server Architecture

This guide explains how to build a **modular SaaS application** using the `ae-base-server` framework and the **Minimal Server** as a starting point.
The minimal server demonstrates key SaaS features—**authentication**, **user and tenant management**, **contacts**, **email services**, and **health checks**—and provides the foundation for adding your own feature modules.



### 1. **Overview**

The Minimal Server provides:

* User and tenant authentication via JWT (`/auth/*`)
* Built-in base modules for **users**, **customers**, **contacts**, and **emails**
* Extensible routing and module registration
* Automatic database connection and schema creation
* Modular structure for additional functionality (e.g., calendar, billing, analytics)

Your app builds on this base by **adding modules** that implement new business logic while inheriting:

* Multi-tenant database scoping (`UserID`, `TenantID`)
* Common authentication and middleware
* Shared API conventions and Swagger documentation



### 2. **Project Structure**

A typical SaaS app built on the minimal server follows this layout:

```
/cmd/minimal-server/        → Main entry point
/pkg/api/                   → Base framework (auth, DB, router, middleware)
 /modules/
   ├── businisess-logic-1/
   │    ├── entities/
   │    ├── handlers/
   │    ├── routes/
   │    └── services/
   └── businisess-logic-2/
        ├── entities/
        ├── handlers/
        ├── routes/
        └── services/
```

Each module is self-contained, providing its own **models**, **handlers**, **services**, and **routes** (see: *Module Development Pattern*).



### 3. **Minimal Server Explained**

The minimal server (`main.go`) sets up everything needed for a multi-tenant SaaS backend:

```go
func main() {
    // 1. Configure and connect to the database
    dbConfig := api.DatabaseConfig{ ... }
    db, err := api.ConnectDatabaseWithAutoCreate(dbConfig)

    // 2. Seed base data (admin user, default tenant, etc.)
    api.SeedBaseData(db)

    // 3. Create a list of active modules
    var modules []api.ModuleRouteProvider

    // Example module registration
    pingModule := NewPingModule(db)
    modules = append(modules, pingModule)

    // 4. Setup the modular router
    router := api.SetupModularRouter(db, modules)

    // 5. Register any public routes
    pingModule.RegisterPublicRoutes(router)

    // 6. Start the HTTP server
    router.Run(":8080")
}
```



### 4. **Adding Your Own Module**

To extend the server with your own functionality:

1. **Create a new module package**, e.g. `/modules/calendar`.
2. Follow the [Module Development Pattern](#module-development-pattern-updated) to define:

   * `entities` (GORM models)
   * `services` (business logic)
   * `handlers` (HTTP endpoints)
   * `routes` (route registration)
   * `module.go` (wiring for initialization)
3. Add the module to the minimal server:

```go
calendarModule := calendar.NewModuleWithAutoMigration(db)
modules = append(modules, calendarModule)
```

4. Run the server and access your module’s routes automatically under `/api/v1/<module>`.



### 5. **Authentication and Tenant Isolation**

Every authenticated request is automatically scoped to a **tenant** and **user** through middleware provided by `ae-base-server`.
Modules can access the authenticated user like this:

```go
user, err := api.GetUser(c)
if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
}
entity.TenantID = user.TenantID
entity.UserID = user.ID
```

This ensures **strict tenant isolation** across all modules.



### 6. **Base Features**

The minimal server includes the following built-in routes:

| Tag         | Description                     | Example Endpoint            |
| -- | - |  |
| `auth`      | User authentication & session   | `POST /api/v1/auth/login`   |
| `users`     | User profiles and accounts      | `GET /api/v1/users/profile` |
| `customers` | Customer management             | `GET /api/v1/customers`     |
| `contacts`  | Contact and newsletter handling | `GET /api/v1/contacts`      |
| `emails`    | Email sending and tracking      | `GET /api/v1/emails`        |
| `health`    | Health checks                   | `GET /api/v1/health`        |

All routes follow the `/api/v1` base path and use the `Bearer` token for authentication.



### 7. **Environment Variables**

The minimal server is configured via environment variables:

| Variable      | Description       | Default      |
| - | -- |  |
| `DB_HOST`     | Database host     | `localhost`  |
| `DB_PORT`     | Database port     | `5432`       |
| `DB_USER`     | Database user     | `postgres`   |
| `DB_PASSWORD` | Database password | `password`   |
| `DB_NAME`     | Database name     | `ae_minimal` |
| `DB_SSLMODE`  | SSL mode          | `disable`    |
| `PORT`        | Server port       | `8080`       |


### 8. **Swagger**

Swagger annotations in each module (and the main server) provide automatic API documentation.

A server needs to declare the base-servers API paths, tags, and additional information. For example:

```go
// @title Minimal Server API
// @version 1.0
// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name auth
// @tag.description User authentication and session management  
...

Each module can define tags like:

```go
// @tag.name calendar
// @tag.description Calendar management and events
```

These tags appear in the generated Swagger UI under `/swagger/index.html`.



### 9. **Summary**

* The Minimal Server is a **blueprint** for building modular, multi-tenant SaaS systems.
* All functionality is organized into **modules** that register themselves automatically.
* Common concerns—auth, DB, tenants, logging, and Swagger—are handled centrally.
* Adding new features only requires writing a new module following the established pattern.

Start small, add modules as your product grows, and you’ll have a scalable, maintainable SaaS backend built on the `ae-base-server` foundation.

