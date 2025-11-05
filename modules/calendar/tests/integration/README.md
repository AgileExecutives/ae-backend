# Calendar Module Integration Tests

## Overview

These integration tests verify the calendar service functionality using a real database (SQLite in-memory) instead of mocks. This approach provides more reliable testing of database operations, GORM queries, and data persistence.

## Test Coverage

### Implemented Tests

1. **TestCalendarService_CreateCalendar**
   - Verifies calendar creation with all required fields
   - Tests UUID generation
   - Validates tenant and user assignment

2. **TestCalendarService_GetCalendarByID**
   - Tests retrieval of existing calendars
   - Verifies data persistence

3. **TestCalendarService_UpdateCalendar**
   - Tests updating calendar properties
   - Validates partial updates

4. **TestCalendarService_DeleteCalendar**
   - Tests soft delete functionality
   - Verifies deleted records cannot be retrieved

5. **TestCalendarService_CreateCalendarEntry**
   - Tests creating calendar entries
   - Validates relationships between calendars and entries
   - Tests JSON field handling (participants)

6. **TestCalendarService_GetCalendarsWithDeepPreload**
   - Tests GORM preloading functionality
   - Verifies nested relationships are loaded
   - Tests calendar entries preloading

7. **TestCalendarService_TenantIsolation**
   - Critical security test
   - Ensures tenant data isolation
   - Verifies cross-tenant access is prevented

## Running Tests

### Run all integration tests:
```bash
cd /Users/alex/src/ae/backend/modules/calendar
./run_tests.sh
```

### Run integration tests directly:
```bash
cd /Users/alex/src/ae/backend/modules/calendar/tests
GOWORK=off go test -v ./integration/...
```

### Run specific test:
```bash
cd /Users/alex/src/ae/backend/modules/calendar/tests
GOWORK=off go test -v ./integration/... -run TestCalendarService_CreateCalendar
```

## Test Database

- **Engine**: SQLite (in-memory)
- **Lifecycle**: Fresh database for each test
- **Schema**: Auto-migrated using GORM
- **Data**: No shared state between tests

## Benefits of Integration Tests

✅ **Real Database Operations**: Tests actual SQL queries and GORM behavior  
✅ **No Mock Complexity**: Eliminates brittle mock setups  
✅ **Relationship Testing**: Validates foreign keys and joins  
✅ **Migration Verification**: Ensures schema is correct  
✅ **Isolation Testing**: Confirms tenant boundaries work correctly  

## Future Enhancements

- [ ] Add calendar series tests
- [ ] Add external calendar tests  
- [ ] Add week/year view tests
- [ ] Add free slots calculation tests
- [ ] Add coverage reporting
- [ ] Add benchmark tests

## Dependencies

- `gorm.io/gorm` - ORM
- `gorm.io/driver/sqlite` - SQLite driver for testing
- `github.com/stretchr/testify` - Testing utilities
