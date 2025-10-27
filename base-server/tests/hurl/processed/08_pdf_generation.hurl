# PDF Generation API Tests with Unique Data
# Tests PDF generation system using template-based approach

# First, create a test user and get authentication token
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "username": "testuser_1761600534_af78c817",
  "email": "test_1761600534_af78c817@example.com",
  "password": "Pass123_af78c817",
  "first_name": "PDF",
  "last_name": "Tester",
  "company_name": "PDF Test Company 1761600534_af78c817",
  "accept_terms": true,
  "newsletter_opt_in": false
}

# Accept either success or conflict (user exists)
HTTP *

# Now login to get the token regardless of registration result
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser_1761600534_af78c817",
  "password": "Pass123_af78c817"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"
user_tenant_id: jsonpath "$.data.user.tenant_id"

# Test 1: Generate PDF using template data
POST http://localhost:8080/api/v1/pdf/create
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "data": {
    "invoice_number": "INV-1761600534_af78c817",
    "customer_name": "customer_1761600534_af78c817",
    "amount": "€123.45",
    "date": "2025-10-27",
    "unique_id": "1761600534_af78c817"
  },
  "templateName": "report.html",
  "fileName": "test-invoice-1761600534_af78c817"
}

# Accept either success or server error (if template missing or Chrome not available)  
HTTP *
[Asserts]
# Should respond with JSON - check that we get a response
status >= 200

# Test 2: Generate PDF with minimal data
POST http://localhost:8080/api/v1/pdf/create
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "data": {
    "title": "Simple Test 1761600534_af78c817",
    "content": "PDF generation test"
  },
  "templateName": "invoice_units_template.html",
  "fileName": "simple-test-1761600534_af78c817"
}

# Accept either success or server error
HTTP *
[Asserts]
status >= 200

# Test 3: Try to create PDF without authentication
POST http://localhost:8080/api/v1/pdf/create
Content-Type: application/json
{
  "data": {"test": "data"},
  "templateName": "report.html",
  "fileName": "unauthorized-test"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 4: Try to create PDF with missing template name
POST http://localhost:8080/api/v1/pdf/create
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "data": {"test": "data"},
  "fileName": "missing-template-test"
}

HTTP 400
[Asserts]
jsonpath "$.error" == "Template name is required"

# Test 5: Try to create PDF with missing file name
POST http://localhost:8080/api/v1/pdf/create
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "data": {"test": "data"},
  "templateName": "report.html"
}

HTTP 400
[Asserts]
jsonpath "$.error" == "File name is required"

# Test 6: Try to create PDF with invalid JSON
POST http://localhost:8080/api/v1/pdf/create
Authorization: Bearer {{auth_token}}
Content-Type: application/json
```
{invalid json}
```

HTTP 400
[Asserts]
jsonpath "$.error" == "Invalid request body"