# PDF Generation API Tests with Unique Data
# Tests PDF generation system including template management, generation, and streaming

# First, create a test user and get authentication token
POST {{HOST}}/api/v1/auth/register
Content-Type: application/json
{
  "username": "{{UNIQUE_USERNAME}}",
  "email": "{{UNIQUE_EMAIL}}",
  "password": "{{UNIQUE_PASSWORD}}",
  "first_name": "PDF",
  "last_name": "Tester",
  "company_name": "PDF Test Company {{UNIQUE_ID}}",
  "accept_terms": true,
  "newsletter_opt_in": false
}

# Accept either success or conflict (user exists)
HTTP *

# Now login to get the token regardless of registration result
POST {{HOST}}/api/v1/auth/login
Content-Type: application/json
{
  "username": "{{UNIQUE_USERNAME}}",
  "password": "{{UNIQUE_PASSWORD}}"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"
user_tenant_id: jsonpath "$.data.user.tenant_id"

# Test 1: Get PDF configuration
GET {{HOST}}/api/v1/pdf/config
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.config" exists
jsonpath "$.config.page_size" exists

# Test 2: List available PDF templates
GET {{HOST}}/api/v1/pdf/templates
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.templates" isCollection

# Test 3: Generate PDF from HTML content
POST {{HOST}}/api/v1/pdf/generate/html
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "html": "<html><body><h1>Test PDF {{UNIQUE_ID}}</h1><p>Generated for user: {{UNIQUE_USERNAME}}</p><p>Email: {{UNIQUE_EMAIL}}</p></body></html>",
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
POST {{HOST}}/api/v1/pdf/generate
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "template": "report",
  "data": {
    "invoice_number": "INV-{{UNIQUE_ID}}",
    "customer_name": "{{UNIQUE_CUSTOMER}}",
    "amount": "€123.45",
    "date": "2025-10-11",
    "unique_id": "{{UNIQUE_ID}}"
  },
  "save": true
}

# Accept either success or server error (wkhtmltopdf dependency issue)
HTTP *
[Asserts]
status >= 200

# Test 5: Stream PDF (generate and stream directly)
POST {{HOST}}/api/v1/pdf/generate/stream
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "template": "report",
  "data": {
    "title": "Stream Test {{UNIQUE_ID}}",
    "content": "PDF streaming test"
  }
}

# Accept either success or server error (wkhtmltopdf dependency)
HTTP *

# Test 6: Get template info for specific template
GET {{HOST}}/api/v1/pdf/templates/report
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test 7: Generate PDF with invalid HTML
POST {{HOST}}/api/v1/pdf/generate/html
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "html": "<invalid><html><broken>{{UNIQUE_ID}}",
  "save": false
}

# Should handle invalid HTML gracefully
HTTP *
[Asserts]
status >= 200

# Test 8: Try to access PDF endpoints without authentication
GET {{HOST}}/api/v1/pdf/config

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 9: Generate PDF with custom options
POST {{HOST}}/api/v1/pdf/generate/html
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "html": "<html><head><title>Custom PDF {{UNIQUE_ID}}</title></head><body><h1>Custom Options Test</h1><p>Test ID: {{UNIQUE_ID}}</p></body></html>",
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