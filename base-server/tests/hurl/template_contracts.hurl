# Template Contract API Tests

# Test GET /templates/contracts - List all contracts
GET http://localhost:8081/api/v1/templates/contracts
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" exists
jsonpath "$.data" isCollection

# Test GET /templates/contracts/welcome - Get specific contract  
GET http://localhost:8081/api/v1/templates/contracts/welcome
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.template_key" == "welcome"
jsonpath "$.data.variable_schema" exists
jsonpath "$.data.variable_schema.type" == "object"

# Test GET /templates/contracts/booking_confirmation - Get booking contract
GET http://localhost:8081/api/v1/templates/contracts/booking_confirmation
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.template_key" == "booking_confirmation" 
jsonpath "$.data.variable_schema" exists
jsonpath "$.data.variable_schema.properties" exists

# Test GET /templates/contracts/password_reset - Get password reset contract
GET http://localhost:8081/api/v1/templates/contracts/password_reset
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.template_key" == "password_reset"
jsonpath "$.data.variable_schema" exists

# Test GET /templates/contracts/invoice - Get invoice contract
GET http://localhost:8081/api/v1/templates/contracts/invoice
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.template_key" == "invoice"
jsonpath "$.data.variable_schema" exists
jsonpath "$.data.variable_schema.properties.Customer" exists
jsonpath "$.data.variable_schema.properties.InvoiceData" exists

# Test GET /templates/contracts/{key}/sample-data - Get sample data for contract
GET http://localhost:8081/api/v1/templates/contracts/welcome/sample-data
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" exists
jsonpath "$.data.FirstName" exists
jsonpath "$.data.LastName" exists

# Test POST /templates/contracts/{key}/validate - Validate data against contract
POST http://localhost:8081/api/v1/templates/contracts/welcome/validate
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "FirstName": "John",
  "LastName": "Doe",
  "OrganizationName": "Test Company"
}

HTTP 200
[Asserts]
jsonpath "$.data.valid" == true
jsonpath "$.data.errors" isEmpty

# Test POST /templates/contracts/welcome/validate - Invalid data validation
POST http://localhost:8081/api/v1/templates/contracts/welcome/validate
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "FirstName": "John"
  // Missing required LastName
}

HTTP 200
[Asserts]
jsonpath "$.data.valid" == false
jsonpath "$.data.errors" isCollection
jsonpath "$.data.errors[0]" contains "required"

# Test POST /templates/contracts/invoice/validate - Complex invoice validation
POST http://localhost:8081/api/v1/templates/contracts/invoice/validate
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "Customer": {
    "Name": "Test Customer Ltd.",
    "Email": "customer@test.com",
    "Address": {
      "Street": "123 Test St",
      "City": "Test City",
      "PostalCode": "12345",
      "Country": "Germany"
    },
    "TaxId": "DE123456789"
  },
  "InvoiceData": {
    "InvoiceNumber": "INV-2024-001",
    "IssueDate": "2024-01-15T00:00:00Z",
    "DueDate": "2024-02-14T00:00:00Z",
    "Items": [
      {
        "Description": "Consulting Service",
        "Quantity": 10,
        "UnitPrice": 150.0,
        "Total": 1500.0,
        "VATRate": 19.0,
        "VATAmount": 285.0
      }
    ],
    "Subtotal": 1500.0,
    "TotalVAT": 285.0,
    "Total": 1785.0,
    "Currency": "EUR",
    "Language": "en"
  },
  "OrganizationData": {
    "Name": "Test Organization GmbH",
    "Address": {
      "Street": "Organization St 1",
      "City": "Berlin",
      "PostalCode": "10115",
      "Country": "Germany"
    },
    "TaxId": "DE987654321",
    "Email": "billing@testorg.com",
    "Phone": "+49 30 12345678"
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.valid" == true
jsonpath "$.data.errors" isEmpty

# Test non-existent contract
GET http://localhost:8081/api/v1/templates/contracts/non_existent_contract
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"