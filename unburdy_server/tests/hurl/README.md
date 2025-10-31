# Client Management API Tests

This directory contains HURL tests for the Unburdy Server client management module.

## Overview

The tests cover:
- Client CRUD operations
- Cost Provider CRUD operations  
- Client-Cost Provider relationships
- Pagination with environment variable limits (MAX_PAGE_LIMIT=1000, DEFAULT_PAGE_LIMIT=200)
- Authentication and authorization
- Error handling and edge cases

## Test Files

1. `01_auth_setup.hurl` - Sets up authentication for other tests
2. `02_clients.hurl` - Tests all client management endpoints
3. `03_cost_providers.hurl` - Tests all cost provider endpoints  
4. `04_client_cost_provider_integration.hurl` - Tests client-cost provider relationships

## Running Tests

### Prerequisites

1. Make sure the server is running:
   ```bash
   cd /Users/alex/src/ae/backend/unburdy_server
   ./tmp/test
   # or
   air
   ```

2. Install HURL if not already installed:
   ```bash
   # macOS
   brew install hurl
   
   # Other platforms: https://hurl.dev/docs/installation.html
   ```

3. Install jq for JSON parsing (optional, for better error reporting):
   ```bash
   brew install jq
   ```

### Running All Tests

Run all tests in order:
```bash
./run-client-tests.sh
```

### Running Individual Tests

Run a specific test file:
```bash
hurl tests/hurl/02_clients.hurl --variable host=http://localhost:8080 --test
```

### Configuration

Edit `tests/hurl/hurl.config` to change the host or other settings.

## Environment Variables Tested

The tests verify that pagination respects these environment variables:
- `MAX_PAGE_LIMIT=1000` - Maximum allowed page size
- `DEFAULT_PAGE_LIMIT=200` - Default page size when not specified

## Test Features

### Authentication
- Tests JWT token-based authentication
- Verifies protected endpoints require valid tokens
- Tests unauthorized access scenarios

### Pagination
- Tests default pagination limits from environment variables
- Tests custom page sizes within limits
- Tests page size limits are enforced (MAX_PAGE_LIMIT)
- Tests pagination metadata in responses

### CRUD Operations
- Create, Read, Update, Delete for both clients and cost providers
- Tests both full and minimal data creation
- Tests partial updates
- Tests validation and error handling

### Relationships
- Tests client-cost provider associations
- Tests updating associations
- Tests removing associations
- Tests foreign key validation

### Error Handling
- Tests validation errors for missing required fields
- Tests 404 errors for non-existent resources
- Tests 401 errors for unauthorized access
- Tests 400 errors for invalid data

## Expected Server Response Format

All endpoints should return responses in this format:

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {...},
  "pagination": {
    "page": 1,
    "limit": 200,
    "total": 150,
    "total_pages": 1
  }
}
```

Error responses:
```json
{
  "success": false,
  "error": "Error message"
}
```

## Troubleshooting

### Server Not Running
If tests fail with connection errors, make sure the server is running on the configured host (default: http://localhost:8080).

### Authentication Failures  
The tests assume a user exists with credentials from the seed data. Check that the database is properly seeded.

### Pagination Test Failures
Ensure your `.env` file has:
```
MAX_PAGE_LIMIT=1000
DEFAULT_PAGE_LIMIT=200
```

### Verbose Output
Run with verbose mode for detailed error information:
```bash
VERBOSE=1 ./run-client-tests.sh
```