# Integration Tests

This directory contains integration tests that test multiple components working together.

## Structure

```
integration/
├── invoice_workflow_test.go      # Complete invoice lifecycle
├── booking_workflow_test.go      # Booking and calendar integration
├── document_workflow_test.go     # Document upload/download
├── multi_tenant_test.go          # Tenant isolation verification
└── README.md
```

## Running Integration Tests

```bash
# Run all integration tests
go test -v ./base-server/tests/integration/...

# Run specific test
go test -v ./base-server/tests/integration/ -run TestInvoiceWorkflow

# Run with coverage
go test -v -coverprofile=coverage.out ./base-server/tests/integration/...
```

## Writing Integration Tests

Integration tests should:
1. Use real database (SQLite in-memory)
2. Test multiple services working together
3. Verify end-to-end workflows
4. Test cross-module interactions

Example:
```go
func TestInvoiceWorkflow_EndToEnd(t *testing.T) {
    db := testutils.SetupTestDB(t)
    defer testutils.CleanupTestDB(db)
    
    // Migrate all required entities
    testutils.MigrateTestDB(t, db, 
        &entities.Invoice{},
        &entities.InvoiceItem{},
        &entities.Document{},
    )
    
    // Initialize all services
    invoiceService := services.NewInvoiceService(db)
    pdfService := services.NewPDFService()
    docService := services.NewDocumentService(db, minioClient)
    
    // Execute full workflow
    // ... test code ...
}
```
