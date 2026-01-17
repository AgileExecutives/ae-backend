# Invoice Module Tests

This directory contains comprehensive test coverage for the invoice module's workflow, validation, and service operations.

## Test Files

### 1. **workflow_test.go**
Tests for invoice workflow methods:
- `TestFinalizeInvoice` - Draft to finalized state transitions
- `TestMarkAsSent` - Finalized to sent state transitions  
- `TestMarkAsPaid` - Sent/overdue to paid state transitions
- `TestSendReminder` - Reminder sending and overdue detection
- `TestCancelInvoice` - Cancellation workflows
- `TestCheckOverdueInvoices` - Batch overdue checking

**Coverage:**
- Status validation (can only finalize drafts, send finalized, etc.)
- Timestamp management (finalized_at, sent_at, payment_date)
- Error cases (invalid status transitions)
- Overdue status auto-detection

### 2. **validation_test.go**
Tests for business rule and VAT validation:
- `TestValidateVAT` - VAT rate validation, exemption text requirements
- `TestValidateInvoice` - Business rules (customer name, line items, amounts, dates)
- `TestValidateInvoiceItems` - Item-level validation (description, quantity, prices, calculations)
- `TestAmountCalculations` - Mathematical accuracy of discounts and VAT

**Coverage:**
- VAT exempt items must have exemption text
- VAT rates between 0-100%
- Negative amounts validation
- Due date > invoice date
- Discount rate 0-100%
- Performance period end >= start  
- Quantity * unit price = amount calculations
- Discount application accuracy

### 3. **invoice_number_test.go**
Tests for invoice number generation:
- `TestInvoiceNumberConcurrency` - Concurrent finalization (no duplicate numbers)
- `TestInvoiceNumberFormats` - Sequential, year-prefix, year-month formats
- `TestInvoiceNumberSequencing` - Sequential number generation
- `TestInvoiceNumberRollback` - Failed finalization doesn't consume numbers
- `TestInvoiceNumberByOrganization` - Separate sequences per organization

**Coverage:**
- Concurrency safety with PostgreSQL advisory locks
- Format-specific number generation
- Transaction rollback behavior
- Multi-tenant number isolation

**Benchmarks:**
- `BenchmarkFinalizeInvoice` - Single-threaded finalization performance
- `BenchmarkConcurrentFinalize` - Concurrent finalization throughput

### 4. **invoice_service_test.go**
Tests for CRUD operations and service methods:
- `TestCreateInvoice` - Invoice creation with all fields
- `TestUpdateInvoice` - Partial updates to draft invoices
- `TestDeleteInvoice` - Deletion rules (only drafts)
- `TestGetInvoice` - Retrieval with tenant isolation
- `TestListInvoices` - Filtering by status, pagination
- `TestInvoiceStatusTransitions` - Complete state machine validation

**Coverage:**
- Full field population (all 18+ business fields)
- VAT exempt invoices
- Optional fields (subject, references, dates)
- Tenant isolation enforcement
- Status-based filtering
- Complete workflow state transitions

### 5. **workflow_handlers_test.go**  
Tests for HTTP API endpoints:
- `TestFinalizeInvoiceHandler` - POST /invoices/:id/finalize
- `TestMarkInvoiceAsSentHandler` - POST /invoices/:id/send
- `TestMarkInvoiceAsPaidHandler` - POST /invoices/:id/pay
- `TestSendInvoiceReminderHandler` - POST /invoices/:id/remind
- `TestCancelInvoiceHandler` - POST /invoices/:id/cancel
- `TestGenerateInvoicePDFHandler` - POST /invoices/:id/generate-pdf

**Coverage:**
- HTTP status codes (200, 400, 403, 404, 422)
- Request parsing (JSON, query params, path params)
- Error response formats
- Successful response serialization

## Running Tests

### Run All Tests
```bash
cd /Users/alex/src/ae/backend/modules/invoice
go test ./... -v
```

### Run Specific Test Suites
```bash
# Workflow tests only
go test ./services -v -run TestFinalize
go test ./services -v -run TestMarkAs
go test ./services -v -run TestSend
go test ./services -v -run TestCancel

# Validation tests only
go test ./services -v -run TestValidate
go test ./services -v -run TestAmount

# Service CRUD tests
go test ./services -v -run TestCreate
go test ./services -v -run TestUpdate
go test ./services -v -run TestDelete
go test ./services -v -run TestGet
go test ./services -v -run TestList

# Handler tests
go test ./handlers -v -run TestFinalize
go test ./handlers -v -run TestMarkInvoice
go test ./handlers -v -run TestSend
go test ./handlers -v -run TestCancel
go test ./handlers -v -run TestGenerate
```

### Run with Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Run Benchmarks
```bash
go test ./services -bench=. -benchmem
```

## Test Dependencies

The tests use in-memory SQLite databases to avoid external dependencies:
- **gorm.io/gorm** - ORM
- **gorm.io/driver/sqlite** - In-memory test database
- **github.com/stretchr/testify** - Assertions and mocking
- **github.com/gin-gonic/gin** - HTTP testing

## Database Setup

Tests use `setupTestDB(t)` helper which:
1. Creates in-memory SQLite database
2. Auto-migrates Invoice and InvoiceItem schemas
3. Returns clean *gorm.DB instance per test

## Test Data Helpers

- `createTestInvoice(t, db, status, withItems)` - Creates test invoice with configurable status and line items
- `ptr[T](v T)` - Helper to create pointers for optional fields

## Known Limitations

### Concurrency Tests
`TestInvoiceNumberConcurrency` is skipped for SQLite because it doesn't support PostgreSQL advisory locks. To test properly:

1. Set up PostgreSQL test database
2. Update connection string in test
3. Remove `t.Skip()` line

### Integration vs Unit Tests

These tests are primarily **integration tests** using real database operations. For pure unit tests with mocked database, see handler tests which mock the service layer.

## Test Coverage Goals

| Package | Target Coverage |
|---------|----------------|
| services/workflow.go | 90%+ |
| services/validation.go | 95%+ |
| services/pdf.go | 80%+ |
| handlers/workflow_handlers.go | 85%+ |
| entities/invoice.go | 100% (DTOs) |

## Edge Cases Tested

1. **Status Transitions**: All invalid transitions rejected
2. **Concurrent Operations**: No duplicate invoice numbers
3. **VAT Validation**: Exempt items require text, rates 0-100%
4. **Amount Calculations**: Quantity * price - discount + VAT
5. **Date Logic**: Due >= invoice, period end >= start
6. **Tenant Isolation**: Can't access other tenant's invoices
7. **Null/Optional Fields**: All optional fields handled correctly
8. **Zero Amounts**: VAT exempt can have 0 total
9. **Negative Values**: Properly rejected
10. **Pagination**: List methods respect limits

## Performance Benchmarks

Expected performance targets (on M1 MacBook Pro):
- Finalize invoice: < 10ms (sequential)
- Concurrent finalize (10 invoices): < 50ms total
- List invoices (100 records): < 5ms
- Validate invoice: < 1ms

## Continuous Integration

To run tests in CI/CD:

```yaml
test:
  script:
    - cd modules/invoice
    - go test ./... -v -race -coverprofile=coverage.out
    - go tool cover -func=coverage.out
```

## Contributing

When adding new features:
1. Write tests FIRST (TDD)
2. Ensure all tests pass
3. Maintain coverage above 85%
4. Add benchmark tests for performance-critical code
5. Update this README with new test descriptions
