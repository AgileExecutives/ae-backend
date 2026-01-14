# Template Module Test Suite Documentation

## Overview

This directory contains comprehensive tests for the Template Module, covering unit tests, integration tests, and HTTP API tests using Hurl.

## Test Structure

```
tests/
├── hurl/                           # HTTP API tests using Hurl
│   ├── templates.hurl              # Main template CRUD API tests
│   ├── template_contracts.hurl     # Contract API tests
│   ├── template_rendering.hurl     # Template rendering tests
│   ├── template_crud.hurl          # Complete CRUD operations
│   ├── variables.hurl              # Test variables (auth tokens, etc.)
│   └── reports/                    # Generated test reports (HTML/JSON)
├── run-template-tests.sh           # Hurl test runner script
└── README.md                       # This file
```

## Unit Tests

Unit tests are located in the module source directories:

- `modules/templates/services/contract_registrar_test.go` - Contract registration tests
- `modules/templates/services/db_contract_provider_test.go` - Database contract provider tests
- `modules/templates/services/renderer/db_renderer_test.go` - Database renderer tests
- `modules/templates/services/template_service_test.go` - Template service tests

### Running Unit Tests

```bash
cd base-server
go test ./modules/templates/services/... -v
```

## Integration Tests

Integration tests validate end-to-end workflows:

- `modules/templates/services/integration_test.go` - Complete system integration tests

### Running Integration Tests

```bash
cd base-server
go test ./modules/templates/services/ -tags=integration -v
```

## HTTP API Tests (Hurl)

Hurl tests validate the REST API endpoints for template operations.

### Prerequisites

1. **Install Hurl**: https://hurl.dev/docs/installation.html
   ```bash
   # macOS
   brew install hurl
   
   # Linux
   curl -LO https://github.com/Orange-OpenSource/hurl/releases/latest/download/hurl-x.x.x-x86_64-linux.tar.gz
   ```

2. **Start the server**:
   ```bash
   cd base-server
   make run
   ```

3. **Update authentication token** in `tests/hurl/variables.hurl`:
   ```
   auth_token=your_actual_bearer_token
   ```

### Running Hurl Tests

**Run all template tests:**
```bash
cd base-server/tests
./run-template-tests.sh
```

**Run specific test file:**
```bash
./run-template-tests.sh -s template_contracts.hurl
```

**Run with verbose output:**
```bash
./run-template-tests.sh --verbose
```

### Hurl Test Files

#### 1. `templates.hurl` - Main Template API
Tests core template CRUD operations:
- GET /templates - List templates
- POST /templates - Create template
- GET /templates/{id} - Get template
- PUT /templates/{id} - Update template
- POST /templates/{id}/render - Render template
- POST /templates/{id}/duplicate - Duplicate template
- DELETE /templates/{id} - Delete template

#### 2. `template_contracts.hurl` - Contract API
Tests template contract endpoints:
- GET /templates/contracts - List contracts
- GET /templates/contracts/{key} - Get contract by key
- GET /templates/contracts/{key}/sample-data - Get sample data
- POST /templates/contracts/{key}/validate - Validate data

#### 3. `template_rendering.hurl` - Rendering Tests
Tests template rendering with various scenarios:
- Email template rendering (welcome, booking, password reset)
- PDF template rendering (invoices)
- Error handling (missing data, invalid templates)
- Data validation

#### 4. `template_crud.hurl` - CRUD Integration
Comprehensive CRUD operation tests:
- Template lifecycle (create → update → duplicate → delete)
- Status toggles (active/inactive)
- Bulk operations (filtering by type/channel)
- Cleanup verification

## Test Data and Authentication

### Authentication Setup

1. **Get Bearer Token** from:
   - `bearer-tokens.json` file in the project root
   - Login endpoint: POST /api/v1/auth/login
   - Existing user session

2. **Update variables.hurl**:
   ```
   auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   base_url=http://localhost:8081
   api_url=http://localhost:8081/api/v1
   ```

### Test Organization and User IDs

The tests use sample organization and user IDs that should exist in your test database. Update these in `variables.hurl` if needed:

```
test_organization_id=01JKZ200EQ08TAX0V0S0HM1J9D
test_user_id=01JKZ200EQ0CRQAQ7MKSFSHF6N
```

## Test Reports

Hurl generates detailed test reports:

- **HTML Report**: `tests/hurl/reports/index.html` - Interactive test results
- **JSON Report**: `tests/hurl/reports/results.json` - Machine-readable results

Open the HTML report in a browser to see detailed test results, response times, and failure details.

## Template Contract Testing

The tests validate the new contract-based template system:

### Registered Contracts

Tests verify these contracts are available:
- `welcome` - Welcome email template
- `booking_confirmation` - Booking confirmation email
- `password_reset` - Password reset email  
- `invoice` - PDF invoice template

### Contract Validation

Tests ensure:
- JSON schema validation works correctly
- Required fields are enforced
- Sample data matches schema
- Complex nested objects (invoice) validate properly

### Template Rendering

Tests verify:
- Templates render with valid data
- Missing required data triggers errors
- Extra fields are handled gracefully
- Different channels (EMAIL, PDF) work correctly

## Troubleshooting

### Common Issues

1. **Server not running**:
   ```bash
   cd base-server
   make run
   ```

2. **Invalid auth token**:
   - Update `auth_token` in `variables.hurl`
   - Check token expiration
   - Generate new token from bearer-tokens.json

3. **Database not seeded**:
   ```bash
   cd base-server
   make seed
   ```

4. **Hurl not installed**:
   ```bash
   brew install hurl  # macOS
   # or follow installation guide
   ```

### Debug Failed Tests

1. **Check HTML report**: Open `tests/hurl/reports/index.html`
2. **Run single test**: `./run-template-tests.sh -s failing_test.hurl`
3. **Enable verbose mode**: `./run-template-tests.sh --verbose`
4. **Check server logs**: Look at base-server console output

## Extending Tests

### Adding New Test Cases

1. **Create new .hurl file** in `tests/hurl/`
2. **Add to test runner** in `run-template-tests.sh`
3. **Use variables.hurl** for configuration
4. **Follow existing patterns** for consistency

### Test File Structure

```hurl
# Test Name and Description
GET http://localhost:8081/api/v1/endpoint
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" exists
jsonpath "$.data.field" equals "expected_value"
[Captures]
variable_name: jsonpath "$.data.id"
```

### Best Practices

- Use descriptive test names and comments
- Capture IDs for use in subsequent tests
- Clean up created resources
- Test both success and error scenarios
- Verify response structure and data
- Use meaningful assertions

## Template Module Architecture Testing

The tests validate the new template module architecture:

1. **Database-Driven Contracts**: Contracts stored in DB, not files
2. **Module Registration**: Modules register their own contracts
3. **Contract Provider Interface**: Abstracted contract access
4. **Schema Validation**: JSON schema validation with jsonschema/v5
5. **Template Storage**: Templates in Minio, metadata in DB

This comprehensive test suite ensures the refactored template system works correctly across all components and integration points.