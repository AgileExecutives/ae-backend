# Development Principles

This document outlines the core development principles and standards for the AE Backend system.

## Table of Contents
- [Tenant Separation](#tenant-separation)
- [Date and Time Handling](#date-and-time-handling)
- [Response Wrappers](#response-wrappers)

---

## Tenant Separation

### Overview
The system implements strict tenant isolation to ensure data security and privacy in the multi-tenant SaaS architecture. All database queries MUST enforce tenant boundaries.

### Core Principles

#### 1. Tenant Context Requirement
- Every request MUST include tenant context (tenant_id and user_id)
- Tenant context is extracted from JWT authentication tokens
- Never trust client-provided tenant IDs; always use authenticated context

#### 2. Database Query Isolation
All GORM queries MUST include tenant filtering:

```go
// ✅ CORRECT - Always filter by tenant_id
db.Where("tenant_id = ?", tenantID).Find(&calendars)

// ❌ WRONG - Missing tenant isolation
db.Find(&calendars)
```

#### 3. Service Layer Enforcement
Services must accept and enforce tenant context:

```go
func (s *CalendarService) GetAllCalendars(tenantID, userID uint) ([]entities.Calendar, error) {
    var calendars []entities.Calendar
    
    // Always include tenant_id in WHERE clause
    result := s.db.Where("tenant_id = ?", tenantID).Find(&calendars)
    
    if result.Error != nil {
        return nil, result.Error
    }
    
    return calendars, nil
}
```

#### 4. Creation Operations
When creating new records, always set the tenant_id:

```go
calendar := entities.Calendar{
    TenantID:  tenantID,
    CreatedBy: userID,
    Title:     req.Title,
    // ... other fields
}

result := s.db.Create(&calendar)
```

#### 5. Update and Delete Operations
Always include tenant_id in update/delete conditions:

```go
// ✅ CORRECT
result := s.db.Where("id = ? AND tenant_id = ?", calendarID, tenantID).Delete(&entities.Calendar{})

// ❌ WRONG - Could delete another tenant's data
result := s.db.Where("id = ?", calendarID).Delete(&entities.Calendar{})
```

#### 6. Relationship Loading
When using GORM's Preload, ensure all related tables also enforce tenant isolation:

```go
db.Where("tenant_id = ?", tenantID).
   Preload("Entries", "tenant_id = ?", tenantID).
   Preload("Entries.Series", "tenant_id = ?", tenantID).
   Find(&calendars)
```

### Testing Tenant Isolation

Always include tenant isolation tests in your integration test suite:

```go
func TestTenantIsolation(t *testing.T) {
    db := setupTestDB(t)
    service := services.NewCalendarService(db)
    
    // Create data for tenant 1
    cal1, _ := service.CreateCalendar(req, 1, 1)
    
    // Create data for tenant 2
    cal2, _ := service.CreateCalendar(req, 2, 2)
    
    // Verify tenant 1 cannot access tenant 2's data
    calendars, _ := service.GetAllCalendars(1, 1)
    assert.Len(t, calendars, 1)
    assert.Equal(t, cal1.ID, calendars[0].ID)
}
```

### Security Checklist

Before merging any code that touches the database:
- [ ] All queries include `WHERE tenant_id = ?`
- [ ] Create operations set `TenantID` field
- [ ] Update/delete operations filter by tenant_id
- [ ] Preload operations include tenant filtering
- [ ] Integration tests verify tenant isolation
- [ ] No raw SQL queries bypass tenant filtering

---
## Date and Time Handling

### Overview

The system uses **UTC as the single source of truth** for all stored timestamps, ensuring consistency across services and environments.
However, some use cases (such as recurring meetings that must always occur at the same **local clock time**) require **local time + timezone** support.

### Core Principles

#### 1. UTC Storage for Instant-Based Events

* All one-time or global events MUST be stored in **UTC**.
* Use Go’s `time.Time` type (which uses UTC internally).
* Never store raw local times or offsets in the database.

```go
type CalendarEntry struct {
    ID        uint      `gorm:"primaryKey"`
    StartTime time.Time `gorm:"not null"` // Always UTC
    EndTime   time.Time `gorm:"not null"` // Always UTC
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
```

#### 2. Local-Time Storage for Recurring Local Events

For events that must always occur at the same **local clock time** (e.g., 10:00 every Monday Berlin time),
store the *local time together with the timezone identifier*.

```go
type LocalRecurringEvent struct {
    ID        uint      `gorm:"primaryKey"`
    StartTime time.Time `gorm:"not null"` // Local clock time (no offset)
    TimeZone  string    `gorm:"not null"` // e.g. "Europe/Berlin"
}
```

When generating occurrences, use the timezone to convert each local instance to UTC for that date. This ensures correct daylight saving transitions (10:00 remains 10:00 local time all year).


### 3. API Response Format

All event timestamps are returned in UTC:

```go
type CalendarEntryResponse struct {
    ID        uint      `json:"id"`
    StartTime time.Time `json:"start_time"` // "2025-11-05T14:30:00Z"
    EndTime   time.Time `json:"end_time"`
    TimeZone  string    `gorm:"not null"` // e.g. "Europe/Berlin"
    Title     string    `json:"title"`
}
```

the same logic is applied to recuing events

### 4. Database Queries

Always query in UTC unless explicitly computing local recurrences.

```go
year := 2025
startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
endOfYear := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)

db.Where("start_time BETWEEN ? AND ?", startOfYear, endOfYear).Find(&entries)
```



### 5. Current Time and Parsing

Always use UTC internally:

```go
// ✅ Correct
now := time.Now().UTC()

// ❌ Wrong
now := time.Now()
```

Parsing:

```go
// ✅ Correct
t, err := time.Parse(time.RFC3339, "2025-11-05T14:30:00Z")
```



### Best Practices

```go
// ✅ UTC event
entry := CalendarEntry{
    StartTime: time.Date(2025, 11, 5, 14, 30, 0, 0, time.UTC),
    EndTime:   time.Date(2025, 11, 5, 15, 30, 0, 0, time.UTC),
}

// ✅ Local recurring event
event := LocalRecurringEvent{
    StartTime: time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
    TimeZone:  "Europe/Berlin",
    Recurrence: "WEEKLY",
}

// ❌ Wrong — store local without zone context
entry.StartTime = time.Now()

// ❌ Wrong — convert to fixed offset manually
entry.StartTime = time.Now().In(time.FixedZone("EST", -5*3600))
```



### 6. Client Responsibilities

* Send all *instant-based* timestamps in UTC (ISO 8601 with Z suffix).
* For recurring events, include both the local time and the user’s timezone.
* Handle all timezone conversions for display.
* Server never performs arbitrary timezone conversions — it only uses explicit zone data.



### Summary

| Use Case                                     | Recommended Model        | Stored As                     | Example                                    |
| -- |  | -- |  |
| Global event (same moment for everyone)      | UTC-based                | UTC timestamp                 | `2025-11-05T14:30:00Z`                     |
| Local recurring meeting (10 AM local always) | Local-time-with-timezone | `"10:00"` + `"Europe/Berlin"` | Recomputed as `09:00Z`/`08:00Z` seasonally |


## Response Wrappers

### Overview
All API endpoints MUST use standardized response wrappers to ensure consistent API responses across the entire system.

### Core Principles

#### 1. Base API Package
Use the `base-server/api` package for all response formatting:

```go
import baseAPI "github.com/ae-base-server/api"
```

#### 2. Success Response Types

**Single Entity Response:**
```go
func (h *CalendarHandler) GetCalendar(c *gin.Context) {
    calendar, err := h.service.GetCalendarByID(calendarID, tenantID, userID)
    if err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    baseAPI.SuccessResponse("Calendar retrieved successfully", calendar)(c)
}
```

Response format:
```json
{
    "message": "Calendar retrieved successfully",
    "data": {
        "id": 1,
        "title": "My Calendar",
        "color": "#FF5733"
    }
}
```

**List Response (Paginated):**
```go
func (h *CalendarHandler) GetAllCalendars(c *gin.Context) {
    calendars, total, err := h.service.GetAllCalendars(tenantID, userID, page, limit)
    if err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    baseAPI.SuccessListResponse(calendars, page, limit, total)(c)
}
```

Response format:
```json
{
    "data": [
        {"id": 1, "title": "Calendar 1"},
        {"id": 2, "title": "Calendar 2"}
    ],
    "pagination": {
        "page": 1,
        "limit": 10,
        "total": 25,
        "total_pages": 3
    }
}
```

#### 3. Error Response

**Standardized Error Handling:**
```go
func (h *CalendarHandler) CreateCalendar(c *gin.Context) {
    var req entities.CreateCalendarRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    calendar, err := h.service.CreateCalendar(req, tenantID, userID)
    if err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    baseAPI.SuccessResponse("Calendar created successfully", calendar)(c)
}
```

Error response format:
```json
{
    "error": {
        "category": "calendar",
        "message": "Invalid request format",
        "details": "Title field is required"
    }
}
```

#### 4. Swagger Documentation

Annotate handlers with correct response types:

```go
// @Summary Get all calendars
// @Tags Calendar
// @Accept json
// @Produce json
// @Param from_year query int false "Start year for filtering entries (default: current year)"
// @Param to_year query int false "End year for filtering entries (default: next year)"
// @Success 200 {object} baseAPI.APIResponse{data=[]entities.CalendarResponse}
// @Failure 500 {object} baseAPI.APIResponse
// @Router /api/v1/calendars [get]
func (h *CalendarHandler) GetAllCalendars(c *gin.Context) {
    // ...
}
```

**Key Swagger Annotations:**
- `@Success 200 {object} baseAPI.APIResponse{data=EntityType}` - Single entity
- `@Success 200 {object} baseAPI.APIResponse{data=[]EntityType}` - List of entities
- `@Failure 500 {object} baseAPI.APIResponse` - Error response

#### 5. Response Wrapper Functions

The base-server/api package provides three core functions:

```go
// SuccessResponse - For single entity responses
func SuccessResponse(message string, data interface{}) gin.HandlerFunc

// SuccessListResponse - For paginated list responses
func SuccessListResponse(data interface{}, page, limit int, total int64) gin.HandlerFunc

// ErrorResponseFunc - For error responses
func ErrorResponseFunc(category string, err error) gin.HandlerFunc
```

### Implementation Examples

**Create Operation:**
```go
func (h *CalendarHandler) CreateCalendar(c *gin.Context) {
    var req entities.CreateCalendarRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    tenantID := c.GetUint("tenant_id")
    userID := c.GetUint("user_id")
    
    calendar, err := h.service.CreateCalendar(req, tenantID, userID)
    if err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    baseAPI.SuccessResponse("Calendar created successfully", calendar)(c)
}
```

**Get Single Entity:**
```go
func (h *CalendarHandler) GetCalendar(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    tenantID := c.GetUint("tenant_id")
    userID := c.GetUint("user_id")
    
    calendar, err := h.service.GetCalendarByID(uint(id), tenantID, userID)
    if err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    baseAPI.SuccessResponse("Calendar retrieved successfully", calendar)(c)
}
```

**Get List (Paginated):**
```go
func (h *CalendarHandler) GetAllCalendars(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    tenantID := c.GetUint("tenant_id")
    userID := c.GetUint("user_id")
    
    calendars, total, err := h.service.GetAllCalendars(tenantID, userID, page, limit)
    if err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    baseAPI.SuccessListResponse(calendars, page, limit, total)(c)
}
```

**Update Operation:**
```go
func (h *CalendarHandler) UpdateCalendar(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    
    var req entities.UpdateCalendarRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    tenantID := c.GetUint("tenant_id")
    userID := c.GetUint("user_id")
    
    calendar, err := h.service.UpdateCalendar(uint(id), req, tenantID, userID)
    if err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    baseAPI.SuccessResponse("Calendar updated successfully", calendar)(c)
}
```

**Delete Operation:**
```go
func (h *CalendarHandler) DeleteCalendar(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    tenantID := c.GetUint("tenant_id")
    userID := c.GetUint("user_id")
    
    if err := h.service.DeleteCalendar(uint(id), tenantID, userID); err != nil {
        baseAPI.ErrorResponseFunc("calendar", err)(c)
        return
    }
    
    baseAPI.SuccessResponse("Calendar deleted successfully", nil)(c)
}
```

### Best Practices

**DO:**
- ✅ Always use `baseAPI.SuccessResponse()` for single entities
- ✅ Always use `baseAPI.SuccessListResponse()` for paginated lists
- ✅ Always use `baseAPI.ErrorResponseFunc()` for errors
- ✅ Document all responses in Swagger annotations
- ✅ Use descriptive success messages
- ✅ Include category in error responses

**DON'T:**
- ❌ Never use `c.JSON()` directly in handlers
- ❌ Never create custom response formats
- ❌ Never return inconsistent response structures
- ❌ Never omit Swagger documentation

### Migration Checklist

When creating new endpoints or updating existing ones:
- [ ] Import `baseAPI "github.com/ae-base-server/api"`
- [ ] Replace `c.JSON()` calls with appropriate wrapper functions
- [ ] Add/update Swagger annotations with correct response types
- [ ] Test response format matches expected structure
- [ ] Regenerate Swagger documentation: `make swagger`
- [ ] Verify Swagger UI shows correct response schema

---

## Conclusion

These three principles form the foundation of our backend architecture:

1. **Tenant Separation** ensures data security and privacy
2. **UTC-Only Time Handling** ensures consistency and simplifies time management
3. **Response Wrappers** ensure API consistency and developer experience

All new code MUST adhere to these principles. Code reviews should verify compliance with these standards.
