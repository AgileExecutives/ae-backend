# AE Backend Architecture Guide

## Architecture Overview

This monorepo uses a **modular architecture** where:
- **base-server**: Provides core SaaS functionality (auth, users, customers, etc.)
- **modules/**: Contains reusable business logic modules
- **services/**: Contains specific applications that combine base + modules

## Directory Structure

```
ae-backend/
├── go.work                   # Go workspace configuration
├── base-server/             # Core SaaS foundation
│   ├── api/                 # Public API exports
│   ├── internal/            # Private implementation
│   └── go.mod
├── modules/                 # Reusable business modules
│   ├── calendar-module/     # Calendar system
│   ├── billing-module/      # Future: billing system
│   └── notifications-module/ # Future: notifications
└── services/                # Application-specific services
    ├── unburdy_server/      # Unburdy application
    └── other-services/      # Future applications
```

## Key Benefits of This Architecture

### 1. **Modular Design**
- Each module is independent and testable
- Modules can be mixed and matched per application
- Clear separation of concerns

### 2. **Authentication Integration**
- All modules automatically use base-server authentication
- Automatic tenant isolation through OrganizationID
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
// All queries automatically filtered by organization
user, _ := api.GetUser(c)
query := db.Where("organization_id = ?", user.OrganizationID)
```

## Module Development Pattern

### 1. **Models** (`models.go`)
```go
type Event struct {
    ID             uint `json:"id"`
    Title          string `json:"title"`
    // Base fields for tenant isolation
    UserID         uint `json:"user_id"`
    OrganizationID uint `json:"organization_id"`
}
```

### 2. **Handlers** (`handlers.go`)
```go
func (h *Handler) CreateEvent(c *gin.Context) {
    // Get authenticated user (automatic with base middleware)
    user, err := api.GetUser(c)
    
    // Create with tenant isolation
    event.OrganizationID = user.OrganizationID
}
```

### 3. **Routes** (`routes.go`)
```go
func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
    handler := NewHandler(db)
    
    // Use base-server auth automatically
    calendarGroup := router.Group("/calendar")
    calendarGroup.Use(api.AuthMiddleware(db))
    {
        calendarGroup.POST("/events", handler.CreateEvent)
        calendarGroup.GET("/events", handler.GetEvents)
    }
}
```

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
    github.com/ae-saas-basic/ae-saas-basic v0.0.0
    github.com/gin-gonic/gin v1.10.1
    gorm.io/gorm v1.30.0
)
replace github.com/ae-saas-basic/ae-saas-basic => ../../base-server
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
// All data operations should include organization filter
db.Where("organization_id = ?", user.OrganizationID).Find(&records)
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
    
    // Create test user with organization
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

## Deployment Options

### Option 1: Single Service
Deploy unburdy-server with all modules as one service:
```bash
cd services/unburdy-server
go build -o unburdy-server
./unburdy-server
```

### Option 2: Multiple Services
Deploy separate services for different purposes:
- Core API service (base-server only)
- Full-featured service (base + all modules)
- Specialized services (base + specific modules)

### Option 3: Microservices
Each module can become its own service if needed:
- Shared authentication via JWT
- Shared database or separate databases
- API gateway to route requests

## Recommended Technologies

### Current Stack (Good Choice!)
- **Go + Gin**: Excellent performance, simple deployment
- **GORM**: Good ORM with migrations
- **PostgreSQL**: Robust, ACID compliant
- **JWT**: Stateless authentication

### Additional Recommendations
- **Redis**: For caching and sessions
- **Docker**: For consistent deployments
- **nginx**: For reverse proxy and load balancing
- **Prometheus**: For monitoring
- **GitHub Actions**: For CI/CD

## Migration Path

### Current → Modular (Recommended Steps)

1. **Keep existing structure working** ✅ (You already have this)
2. **Create calendar module** ✅ (Just implemented)
3. **Test integration** (Next step)
4. **Gradually extract more modules** (billing, notifications, etc.)
5. **Refactor existing code into modules** (optional)

### Monorepo vs Separate Repos

**✅ Monorepo (Recommended for your case):**
- Easier dependency management
- Atomic changes across base + modules
- Simpler CI/CD
- Better for rapid development
- Easier to maintain consistency

**❌ Separate Repos (Not recommended):**
- Complex version synchronization
- Difficult to make breaking changes
- Multiple CI/CD pipelines
- Harder to test integrations

## Next Steps

1. **Test the calendar module integration**:
   ```bash
   cd unburdy_server
   go mod tidy
   go run main_with_modules.go
   ```

2. **Create more modules**:
   - Billing module
   - Notification module
   - Reporting module

3. **Enhance base-server**:
   - Add more middleware options
   - Improve API exports
   - Add more utility functions

4. **Set up CI/CD**:
   - Run tests for all modules
   - Build and deploy services
   - Version management

Would you like me to help you implement any of these next steps?