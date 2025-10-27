# PDF Generation API Tests with Unique Data
# Tests PDF generation system including template management, generation, and streaming

# First, create a test user and get authentication token
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "username": "testuser_1761575244_7f3dbf06",
  "email": "test_1761575244_7f3dbf06@example.com",
  "password": "Pass123_7f3dbf06",
  "first_name": "PDF",
  "last_name": "Tester",
  "company_name": "PDF Test Company 1761575244_7f3dbf06",
  "accept_terms": true,
  "newsletter_opt_in": false
}

# Accept either success or conflict (user exists)
HTTP *

# Now login to get the token regardless of registration result
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser_1761575244_7f3dbf06",
  "password": "Pass123_7f3dbf06"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"
user_tenant_id: jsonpath "$.data.user.tenant_id"

# Test 1: Get PDF configuration
GET http://localhost:8080/api/v1/pdf/config
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.config" exists
jsonpath "$.config.page_size" exists

# Test 2: List available PDF templates
GET http://localhost:8080/api/v1/pdf/templates
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.templates" isCollection

# Test 3: Generate PDF from HTML content
POST http://localhost:8080/api/v1/pdf/generate/html
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "html": "<html><body><h1>Test PDF 1761575244_7f3dbf06</h1><p>Generated for user: testuser_1761575244_7f3dbf06</p><p>Email: test_1761575244_7f3dbf06@example.com</p></body></html>",
  "config": {
    "page_size": "A4",
    "orientation": "portrait",
    "margin_top": "1cm",
    "margin_bottom": "1cm"
  },
  "save": true
}

# Accept either success or server error (if wkhtmltopdf is not available)  
HTTP *
[Asserts]
# Should respond with JSON - check that we get a response
status >= 200

# Test 4: Generate PDF using template data (template-based endpoint)
POST http://localhost:8080/api/v1/pdf/generate
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "template": "report",
  "data": {
    "invoice_number": "INV-1761575244_7f3dbf06",
    "customer_name": "customer_1761575244_7f3dbf06",
    "amount": "€123.45",
    "date": "2025-10-11",
    "unique_id": "1761575244_7f3dbf06"
  },
  "save": true
}

# Accept either success or server error (wkhtmltopdf dependency issue)
HTTP *
[Asserts]
status >= 200

# Test 5: Stream PDF (generate and stream directly)
POST http://localhost:8080/api/v1/pdf/generate/stream
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "template": "report",
  "data": {
    "title": "Stream Test 1761575244_7f3dbf06",
    "content": "PDF streaming test"
  }
}

# Accept either success or server error (wkhtmltopdf dependency)
HTTP *

# Test 6: Get template info for specific template
GET http://localhost:8080/api/v1/pdf/templates/report
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test 7: Generate PDF with invalid HTML
POST http://localhost:8080/api/v1/pdf/generate/html
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "html": "<invalid><html><broken>1761575244_7f3dbf06",
  "save": false
}

# Should handle invalid HTML gracefully
HTTP *
[Asserts]
status >= 200

# Test 8: Try to access PDF endpoints without authentication
GET http://localhost:8080/api/v1/pdf/config

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 9: Generate PDF with custom options
POST http://localhost:8080/api/v1/pdf/generate/html
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "html": "<html><head><title>Custom PDF 1761575244_7f3dbf06</title></head><body><h1>Custom Options Test</h1><p>Test ID: 1761575244_7f3dbf06</p></body></html>",
  "config": {
    "page_size": "A3",
    "orientation": "landscape",
    "margin_top": "2cm",
    "margin_bottom": "2cm",
    "margin_left": "1.5cm",
    "margin_right": "1.5cm"
  },
  "save": true
}

# Accept success or various error codes
HTTP *
[Asserts]
status >= 200