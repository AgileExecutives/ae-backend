# Booking Module Development Guide

This document outlines the principles, patterns, and techniques used in developing the booking module. These practices should be followed for creating new modules in the ae-backend ecosystem.

## Table of Contents

1. [Module Structure](#module-structure)
2. [Using Base Server Types](#using-base-server-types)
3. [Standardized API Responses](#standardized-api-responses)
4. [Swagger Documentation](#swagger-documentation)
5. [Database Patterns](#database-patterns)
6. [Module Integration](#module-integration)
7. [Testing](#testing)

---

## Module Structure

A well-organized module follows this directory structure:

```
booking/
├── entities/               # Domain models and data structures
│   ├── booking_template.go    # Main entity with GORM tags
│   ├── core_entities.go        # Core.Entity implementation
│   └── requests.go             # Request/Response DTOs
├── services/               # Business logic layer
│   └── booking_service.go
├── handlers/               # HTTP handlers (controllers)
│   └── booking_handler.go
├── routes/                 # Route registration
│   └── routes.go
├── documentation/          # Module-specific documentation
│   └── MODULE_DEVELOPMENT_GUIDE.md
├── module.go              # Module registration and lifecycle
└── go.mod                 # Module dependencies

```

### Key Principles

1. **Separation of Concerns**: Keep entities, business logic, and HTTP handling separate
2. **Dependency Injection**: Services are injected into handlers, DB into services
3. **Single Responsibility**: Each file/package has one clear purpose

---

## Using Base Server Types

### Importing Base Server Components

Always use the public API package for base-server types:

```go
import (
    baseAPI "github.com/ae-base-server/api"
)
```

**DO NOT** import from internal packages:
```go
// ❌ WRONG - Cannot import internal packages from external modules
import baseModels "github.com/ae-base-server/internal/models"
```

### Available Base Server Types

The `baseAPI` package re-exports commonly used types:

```go
// Response Types
type APIResponse = models.APIResponse
type ErrorResponse = models.ErrorResponse
type ListResponse = models.ListResponse

// Authentication Types
type User = models.User
type UserResponse = models.UserResponse
type Tenant = models.Tenant
type TenantResponse = models.TenantResponse

// Helper Functions
var (
    SuccessResponse        = models.SuccessResponse
    SuccessMessageResponse = models.SuccessMessageResponse
    SuccessListResponse    = models.SuccessListResponse
    ErrorResponseFunc      = models.ErrorResponseFunc
)
```

---

## Standardized API Responses

### Response Structure

All API endpoints MUST return this standardized structure:

```json
{
  "success": true|false,
  "message": "optional message",
  "data": <T>,
  "error": "error message if success=false"
}
```

### Using Response Helper Functions

#### Success Response with Data

```go
func (h *BookingHandler) CreateConfiguration(c *gin.Context) {
    // ... business logic ...
    
    c.JSON(http.StatusCreated, baseAPI.SuccessResponse(
        "Booking configuration created successfully",
        config.ToResponse(),
    ))
}
```

#### Success Response with Message Only

```go
func (h *BookingHandler) DeleteConfiguration(c *gin.Context) {
    // ... business logic ...
    
    c.JSON(http.StatusOK, baseAPI.SuccessMessageResponse(
        "Booking configuration deleted successfully",
    ))
}
```

#### Error Response

```go
func (h *BookingHandler) GetConfiguration(c *gin.Context) {
    config, err := h.service.GetConfiguration(id, tenantID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, baseAPI.ErrorResponseFunc(
            "",  // optional user-friendly message
            err.Error(),
        ))
        return
    }
}
```

#### Paginated List Response

```go
func (h *Handler) GetAll(c *gin.Context) {
    page, limit := utils.GetPaginationParams(c)
    items, total, err := h.service.GetAll(tenantID, page, limit)
    // ... error handling ...
    
    c.JSON(http.StatusOK, baseAPI.SuccessListResponse(
        responses,
        page,
        limit,
        int(total),
    ))
}
```

This returns:
```json
{
  "success": true,
  "data": {
    "data": [...],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "total_pages": 10
    }
  }
}
```

#### Simple Array Response (No Pagination)

```go
func (h *BookingHandler) GetAllConfigurations(c *gin.Context) {
    // Get all records without pagination (limit = -1)
    configs, _, err := h.service.GetAllConfigurations(tenantID, 1, -1)
    // ... error handling ...
    
    c.JSON(http.StatusOK, baseAPI.SuccessResponse("", responses))
}
```

---

## Swagger Documentation

### Handler-Level Documentation

Every handler function MUST include comprehensive Swagger annotations:

```go
// CreateConfiguration godoc
// @Summary Create a new booking configuration
// @Description Create a new booking configuration/template for a user's calendar
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param configuration body entities.CreateBookingTemplateRequest true "Booking configuration data"
// @Success 201 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates [post]
// @ID createBookingTemplate
func (h *BookingHandler) CreateConfiguration(c *gin.Context) {
    // implementation
}
```

### Swagger Annotation Components

| Annotation | Purpose | Required | Example |
|------------|---------|----------|---------|
| `@Summary` | Short description | **Yes** | `Create a new booking configuration` |
| `@Description` | Detailed description | **Yes** | `Create a new booking configuration/template...` |
| `@Tags` | Group endpoints | **Yes** | `booking-templates` |
| `@Accept` | Request content type | For POST/PUT | `json` |
| `@Produce` | Response content type | **Yes** | `json` |
| `@Param` | Parameters | For requests with params | `configuration body entities.CreateBookingTemplateRequest true "desc"` |
| `@Success` | Success response | **Yes** | `201 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}` |
| `@Failure` | Error responses | **Yes** | `400 {object} baseAPI.APIResponse` |
| `@Security` | Auth requirement | **Yes** | `BearerAuth` |
| `@Router` | Endpoint path | **Yes** | `/booking/templates [post]` |
| `@ID` | Operation ID | **REQUIRED** | `createBookingTemplate` |

### Operation IDs (@ID) - REQUIRED for API Clients

**Every endpoint MUST include an `@ID` annotation with a unique operation ID.**

#### Why Operation IDs Are Required

Operation IDs are critical for API client generation tools (like OpenAPI Generator, Swagger Codegen, etc.) because they:

1. **Generate Method Names**: The operationId becomes the function/method name in generated client libraries
   ```typescript
   // Generated TypeScript client
   await apiClient.createBookingTemplate(request);  // from @ID createBookingTemplate
   await apiClient.listBookingTemplates();          // from @ID listBookingTemplates
   ```

2. **Ensure Uniqueness**: Prevent conflicts when the same HTTP method is used on different paths
   ```go
   // Both are GET requests, but different operations
   // @ID listBookingTemplates
   GET /booking/templates
   
   // @ID listBookingTemplatesByUser
   GET /booking/templates/by-user
   ```

3. **Enable Stable APIs**: Changing the handler function name doesn't break generated clients if operationId remains consistent

4. **Improve Documentation**: Provides clear, semantic names in API documentation

#### Operation ID Naming Convention

Use **camelCase** and follow this pattern:

| HTTP Method | Pattern | Example |
|-------------|---------|---------|
| POST | `create{Resource}` | `createBookingTemplate` |
| GET (single) | `get{Resource}` | `getBookingTemplate` |
| GET (list) | `list{Resources}` | `listBookingTemplates` |
| GET (filtered) | `list{Resources}By{Filter}` | `listBookingTemplatesByUser` |
| PUT/PATCH | `update{Resource}` | `updateBookingTemplate` |
| DELETE | `delete{Resource}` | `deleteBookingTemplate` |

#### Complete Example

```go
// CreateConfiguration godoc
// @Summary Create a new booking configuration
// @Description Create a new booking configuration/template for a user's calendar
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param configuration body entities.CreateBookingTemplateRequest true "Booking configuration data"
// @Success 201 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates [post]
// @ID createBookingTemplate                    ← REQUIRED
func (h *BookingHandler) CreateConfiguration(c *gin.Context) {
    // implementation
}

// GetAllConfigurations godoc
// @Summary Get all booking configurations
// @Description Retrieve all booking configurations for the tenant
// @Tags booking-templates
// @Produce json
// @Success 200 {object} baseAPI.APIResponse{data=[]entities.BookingTemplateResponse}
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates [get]
// @ID listBookingTemplates                     ← REQUIRED
func (h *BookingHandler) GetAllConfigurations(c *gin.Context) {
    // implementation
}

// GetConfigurationsByUser godoc
// @Summary Get booking configurations by user ID
// @Description Retrieve all booking configurations for a specific user
// @Tags booking-templates
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} baseAPI.APIResponse{data=[]entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/by-user [get]
// @ID listBookingTemplatesByUser               ← REQUIRED (note the "ByUser" suffix)
func (h *BookingHandler) GetConfigurationsByUser(c *gin.Context) {
    // implementation
}
```

#### Booking Module Operation IDs

The booking module uses these operation IDs:

| Endpoint | Method | Operation ID |
|----------|--------|--------------|
| `/booking/templates` | POST | `createBookingTemplate` |
| `/booking/templates` | GET | `listBookingTemplates` |
| `/booking/templates/{id}` | GET | `getBookingTemplate` |
| `/booking/templates/{id}` | PUT | `updateBookingTemplate` |
| `/booking/templates/{id}` | DELETE | `deleteBookingTemplate` |
| `/booking/templates/by-user` | GET | `listBookingTemplatesByUser` |
| `/booking/templates/by-calendar` | GET | `listBookingTemplatesByCalendar` |

#### Verification

After adding `@ID` annotations, verify they appear in the generated Swagger:

```bash
cd unburdy_server
make swagger
jq '.paths."/booking/templates".post.operationId' docs/swagger.json
# Output: "createBookingTemplate"
```

```go
// CreateConfiguration godoc
// @Summary Create a new booking configuration
// @Description Create a new booking configuration/template for a user's calendar
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param configuration body entities.CreateBookingTemplateRequest true "Booking configuration data"
// @Success 201 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates [post]
func (h *BookingHandler) CreateConfiguration(c *gin.Context) {
    // implementation
}
```

### Swagger Annotation Components

| Annotation | Purpose | Example |
|------------|---------|---------|
| `@Summary` | Short description | `Create a new booking configuration` |
| `@Description` | Detailed description | `Create a new booking configuration/template...` |
| `@Tags` | Group endpoints | `booking-templates` |
| `@Accept` | Request content type | `json` |
| `@Produce` | Response content type | `json` |
| `@Param` | Parameters | `configuration body entities.CreateBookingTemplateRequest true "desc"` |
| `@Success` | Success response | `201 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}` |
| `@Failure` | Error responses | `400 {object} baseAPI.APIResponse` |
| `@Security` | Auth requirement | `BearerAuth` |
| `@Router` | Endpoint path | `/booking/templates [post]` |

### Typed Response Documentation

Use the `{object} Type{field=SubType}` syntax for typed responses:

```go
// Single object in data field
// @Success 200 {object} baseAPI.APIResponse{data=entities.BookingTemplateResponse}

// Array in data field
// @Success 200 {object} baseAPI.APIResponse{data=[]entities.BookingTemplateResponse}

// Paginated response
// @Success 200 {object} baseAPI.APIResponse{data=baseAPI.ListResponse}

// Simple success message
// @Success 200 {object} baseAPI.APIResponse
```

### Integration with Server Swagger

#### 1. Add Module to Main Server

In `unburdy_server/main.go`:

```go
// @tag.name booking-templates
// @tag.description Booking template management endpoints
func main() {
    // ... existing code ...
    
    // Register modules
    moduleManager.RegisterModule(booking.NewCoreModule())
}
```

#### 2. Update Makefile

In `unburdy_server/Makefile`:

```makefile
.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@echo "Scanning unburdy handlers, ae-saas-basic handlers, calendar module, and booking module..."
	swag init -g main.go -o ./docs \
		--parseDependency \
		--parseInternal \
		--parseDepth 2 \
		-d ./,../base-server,../modules/calendar,../modules/booking
	@echo "✓ Swagger docs generated with base API, calendar, and booking endpoints"
```

#### 3. Update go.mod

Add module dependency in `unburdy_server/go.mod`:

```go
require (
    github.com/unburdy/booking-module v0.0.0
)

replace github.com/unburdy/booking-module => ../modules/booking
```

#### 4. Generate Swagger

```bash
cd unburdy_server
make swagger
```

This generates:
- `docs/docs.go` - Go code with embedded Swagger spec
- `docs/swagger.json` - OpenAPI JSON specification
- `docs/swagger.yaml` - OpenAPI YAML specification

---

## Database Patterns

### Entity Definition

```go
type BookingTemplate struct {
    gorm.Model  // Includes ID, CreatedAt, UpdatedAt, DeletedAt
    
    // Foreign Keys
    UserID     uint `gorm:"not null;index" json:"user_id"`
    CalendarID uint `gorm:"not null;index" json:"calendar_id"`
    TenantID   uint `gorm:"not null;index" json:"tenant_id"`
    
    // Regular fields
    Name        string `gorm:"not null" json:"name"`
    Description string `json:"description"`
    
    // JSONB fields for complex types
    WeeklyAvailability WeeklyAvailability `gorm:"type:jsonb" json:"weekly_availability"`
    AllowedIntervals   IntervalArray      `gorm:"type:jsonb" json:"allowed_intervals"`
}

func (BookingTemplate) TableName() string {
    return "booking_templates"
}
```

### JSONB Custom Types

For complex nested structures, implement `driver.Valuer` and `sql.Scanner`:

```go
type WeeklyAvailability struct {
    Monday    []TimeRange `json:"monday"`
    Tuesday   []TimeRange `json:"tuesday"`
    // ... other days
}

func (w WeeklyAvailability) Value() (driver.Value, error) {
    return json.Marshal(w)
}

func (w *WeeklyAvailability) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("type assertion to []byte failed")
    }
    return json.Unmarshal(bytes, w)
}
```

### Service Layer Pattern

```go
type BookingService struct {
    db *gorm.DB
}

func NewBookingService(db *gorm.DB) *BookingService {
    return &BookingService{db: db}
}

// Always filter by tenant_id for multi-tenancy
func (s *BookingService) GetConfiguration(id, tenantID uint) (*entities.BookingTemplate, error) {
    var config entities.BookingTemplate
    
    if err := s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&config).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("booking configuration not found")
        }
        return nil, fmt.Errorf("failed to retrieve booking configuration: %w", err)
    }
    
    return &config, nil
}
```

### Pagination Pattern

```go
func (s *BookingService) GetAll(tenantID uint, page, limit int) ([]Entity, int64, error) {
    var entities []Entity
    var total int64
    
    query := s.db.Model(&Entity{}).Where("tenant_id = ?", tenantID)
    
    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count: %w", err)
    }
    
    // Get paginated results
    // Use limit = -1 to get all records without pagination
    offset := (page - 1) * limit
    if err := query.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to retrieve: %w", err)
    }
    
    return entities, total, nil
}
```

---

## Module Integration

### Module Interface Implementation

Every module must implement the `core.Module` interface:

```go
type Module struct {
    db *gorm.DB
}

func NewCoreModule() core.Module {
    return &Module{}
}

func (m *Module) Name() string {
    return "booking"
}

func (m *Module) Initialize(db *gorm.DB) error {
    m.db = db
    return nil
}

func (m *Module) GetEntities() []core.Entity {
    return []core.Entity{
        entities.NewBookingTemplateEntity(),
    }
}

func (m *Module) Dependencies() []string {
    return []string{"base", "calendar"}
}

func (m *Module) GetRouteProviders(router *gin.Engine) []core.RouteProvider {
    service := services.NewBookingService(m.db)
    handler := handlers.NewBookingHandler(service)
    
    return []core.RouteProvider{
        &bookingRouteAdapter{
            provider: routes.NewRouteProvider(handler, m.db),
        },
    }
}
```

### Route Provider Pattern

```go
type RouteProvider struct {
    bookingHandler *handlers.BookingHandler
    db             *gorm.DB
}

func NewRouteProvider(handler *handlers.BookingHandler, db *gorm.DB) *RouteProvider {
    return &RouteProvider{
        bookingHandler: handler,
        db:             db,
    }
}

func (rp *RouteProvider) RegisterRoutes(router *gin.RouterGroup) {
    templates := router.Group("/booking/templates")
    {
        templates.POST("", rp.bookingHandler.CreateConfiguration)
        templates.GET("", rp.bookingHandler.GetAllConfigurations)
        templates.GET("/:id", rp.bookingHandler.GetConfiguration)
        templates.PUT("/:id", rp.bookingHandler.UpdateConfiguration)
        templates.DELETE("/:id", rp.bookingHandler.DeleteConfiguration)
    }
}

func (rp *RouteProvider) GetMiddleware() []gin.HandlerFunc {
    return []gin.HandlerFunc{
        middleware.AuthMiddleware(rp.db),
    }
}
```

---

## Testing

### Service Layer Tests

```go
func TestCreateConfiguration(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    service := NewBookingService(db)
    
    req := entities.CreateBookingTemplateRequest{
        Name:     "Test Config",
        UserID:   1,
        TenantID: 1,
        // ... other fields
    }
    
    config, err := service.CreateConfiguration(req, 1)
    
    assert.NoError(t, err)
    assert.NotNil(t, config)
    assert.Equal(t, "Test Config", config.Name)
}
```

### Integration Tests

Use hurl files for API testing:

```hurl
# Create Booking Template
POST http://localhost:8080/api/v1/booking/templates
Authorization: Bearer {{token}}
Content-Type: application/json

{
  "name": "Morning Slots",
  "user_id": 1,
  "calendar_id": 1,
  "slot_duration": 30,
  "buffer_time": 5
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "Morning Slots"
```

---

## Best Practices Checklist

### Before Creating a New Module

- [ ] Define clear module boundaries and responsibilities
- [ ] Identify dependencies on other modules
- [ ] Design entity schema with proper indexes
- [ ] Plan API endpoints (CRUD + custom operations)

### During Development

- [ ] Use base-server types (`baseAPI.APIResponse`, etc.)
- [ ] Implement standardized response structure for all endpoints
- [ ] Add comprehensive Swagger annotations to all handlers
- [ ] Implement tenant isolation in all database queries
- [ ] Create ToResponse() methods for clean DTO conversion
- [ ] Handle errors consistently with proper HTTP status codes

### Before Integration

- [ ] Update server's `main.go` with module registration
- [ ] Add module path to Makefile swagger target
- [ ] Update server's `go.mod` with module dependency
- [ ] Test Swagger generation (`make swagger`)
- [ ] Verify all endpoints return standardized responses
- [ ] Verify all endpoints have unique operationIds
- [ ] Test multi-tenancy isolation

### Documentation

- [ ] Document module-specific patterns in this guide
- [ ] Add inline code comments for complex logic
- [ ] Update API documentation for consumers
- [ ] Create example requests/responses

---

## Common Pitfalls to Avoid

### ❌ Don't: Import Internal Packages

```go
// WRONG
import baseModels "github.com/ae-base-server/internal/models"
```

**Why?** Go's internal package visibility prevents cross-module imports.

**Solution:** Use the public API package:
```go
import baseAPI "github.com/ae-base-server/api"
```

### ❌ Don't: Create Custom Response Structures

```go
// WRONG
type CustomResponse struct {
    Data interface{} `json:"data"`
    OK   bool        `json:"ok"`
}
```

**Solution:** Always use `baseAPI.SuccessResponse()` or related helpers.

### ❌ Don't: Skip Tenant Filtering

```go
// WRONG - No tenant isolation
s.db.Where("id = ?", id).First(&entity)

// CORRECT - Always filter by tenant
s.db.Where("id = ? AND tenant_id = ?", id, tenantID).First(&entity)
```

### ❌ Don't: Return Direct Entities

```go
// WRONG
c.JSON(http.StatusOK, config)

// CORRECT - Use response DTOs
c.JSON(http.StatusOK, baseAPI.SuccessResponse("", config.ToResponse()))
```

### ❌ Don't: Forget Swagger Annotations

Every handler needs:
- Summary and Description
- Tags for grouping
- Parameter definitions
- Typed response definitions
- Error responses

---

## Summary

This module demonstrates:

1. **Clean Architecture**: Separation of entities, services, handlers, and routes
2. **Type Safety**: Using base-server types for consistency
3. **API Standards**: Standardized response structure across all endpoints
4. **Documentation**: Comprehensive Swagger annotations with typed responses
5. **Multi-tenancy**: Proper tenant isolation in all queries
6. **Integration**: Seamless module registration with the main server
7. **Maintainability**: Clear patterns that scale across modules

Follow these patterns for all new modules to ensure consistency and maintainability across the codebase.
