# Template Rendering Tests - Focus on different template types and rendering scenarios

# Test Email Template Rendering - Welcome Email
POST http://localhost:8081/api/v1/templates/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_key": "welcome",
  "channel": "EMAIL",
  "data": {
    "FirstName": "Sarah",
    "LastName": "Johnson",
    "OrganizationName": "Hurl Testing Corp",
    "ActivationLink": "https://app.example.com/activate?token=abc123"
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.content" contains "Sarah Johnson"
jsonpath "$.data.content" contains "Hurl Testing Corp"
jsonpath "$.data.content" contains "activate?token=abc123"
jsonpath "$.data.subject" exists

# Test Booking Confirmation Template Rendering
POST http://localhost:8081/api/v1/templates/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_key": "booking_confirmation",
  "channel": "EMAIL",
  "data": {
    "CustomerFirstName": "Michael",
    "CustomerLastName": "Brown",
    "ServiceName": "Professional Consultation",
    "BookingDate": "2024-02-15T14:30:00Z",
    "Duration": 60,
    "Location": "Conference Room A",
    "ProviderName": "Dr. Smith",
    "BookingReference": "BK-2024-001",
    "TotalAmount": 200.00,
    "Currency": "EUR"
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.content" contains "Michael Brown"
jsonpath "$.data.content" contains "Professional Consultation"
jsonpath "$.data.content" contains "BK-2024-001"
jsonpath "$.data.content" contains "200.00"

# Test Password Reset Template Rendering
POST http://localhost:8081/api/v1/templates/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_key": "password_reset",
  "channel": "EMAIL",
  "data": {
    "FirstName": "Alice",
    "LastName": "Wilson",
    "ResetLink": "https://app.example.com/reset-password?token=xyz789",
    "ExpirationTime": "24 hours"
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.content" contains "Alice Wilson"
jsonpath "$.data.content" contains "reset-password?token=xyz789"
jsonpath "$.data.content" contains "24 hours"

# Test Invoice PDF Template Rendering
POST http://localhost:8081/api/v1/templates/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_key": "invoice",
  "channel": "PDF",
  "data": {
    "Customer": {
      "Name": "ACME Corporation Ltd.",
      "Email": "billing@acme.com",
      "Address": {
        "Street": "456 Business Avenue",
        "City": "Commerce City",
        "PostalCode": "54321",
        "Country": "Germany"
      },
      "TaxId": "DE111222333"
    },
    "InvoiceData": {
      "InvoiceNumber": "INV-2024-002",
      "IssueDate": "2024-01-20T00:00:00Z",
      "DueDate": "2024-02-19T00:00:00Z",
      "Items": [
        {
          "Description": "Software License",
          "Quantity": 5,
          "UnitPrice": 99.99,
          "Total": 499.95,
          "VATRate": 19.0,
          "VATAmount": 94.99
        },
        {
          "Description": "Support Package",
          "Quantity": 1,
          "UnitPrice": 299.99,
          "Total": 299.99,
          "VATRate": 19.0,
          "VATAmount": 57.00
        }
      ],
      "Subtotal": 799.94,
      "TotalVAT": 151.99,
      "Total": 951.93,
      "Currency": "EUR",
      "Language": "en"
    },
    "OrganizationData": {
      "Name": "Hurl Test Services GmbH",
      "Address": {
        "Street": "Test Street 123",
        "City": "Berlin",
        "PostalCode": "10115",
        "Country": "Germany"
      },
      "TaxId": "DE999888777",
      "Email": "invoices@hurltest.com",
      "Phone": "+49 30 9876543"
    }
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.content" contains "ACME Corporation"
jsonpath "$.data.content" contains "INV-2024-002"
jsonpath "$.data.content" contains "951.93"
jsonpath "$.data.content" contains "Software License"
jsonpath "$.data.content" contains "Support Package"

# Test Template Rendering with Missing Required Data
POST http://localhost:8081/api/v1/templates/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_key": "welcome",
  "channel": "EMAIL",
  "data": {
    "FirstName": "John"
    // Missing required LastName and OrganizationName
  }
}

HTTP 400
[Asserts]
jsonpath "$.error" contains "validation"

# Test Template Rendering with Invalid Template Key
POST http://localhost:8081/api/v1/templates/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_key": "non_existent_template",
  "channel": "EMAIL",
  "data": {
    "SomeData": "value"
  }
}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test Template Rendering with Wrong Channel
POST http://localhost:8081/api/v1/templates/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_key": "welcome",
  "channel": "SMS", // Welcome template might not support SMS
  "data": {
    "FirstName": "Test",
    "LastName": "User",
    "OrganizationName": "Test Org"
  }
}

# This might return 404 if SMS channel is not supported for welcome template
HTTP [404, 400]

# Test Template Rendering with Extra Data (should be allowed)
POST http://localhost:8081/api/v1/templates/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_key": "welcome",
  "channel": "EMAIL",
  "data": {
    "FirstName": "Test",
    "LastName": "User",
    "OrganizationName": "Test Org",
    "ExtraField": "This should be ignored",
    "AnotherField": 12345
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.content" contains "Test User"
jsonpath "$.data.content" contains "Test Org"