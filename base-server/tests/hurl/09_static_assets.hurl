# Static JSON API Tests
# Test secure JSON file serving with enhanced security validation
# Note: Static endpoints now require authentication

# Setup fresh auth session
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "testuser",
  "password": "newpass123"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test listing available JSON files
GET {{host}}/api/v1/static
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.available_files" exists
jsonpath "$.available_files[*]" includes "diagnostic_std"
jsonpath "$.available_files[*]" includes "bundeslaender"
jsonpath "$.available_files[*]" includes "enumerations"
jsonpath "$.security_note" exists

# Test serving legitimate JSON file - diagnostic_std
GET {{host}}/api/v1/static/diagnostic_std
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
body contains "icd10"
body contains "category"

# Test serving legitimate JSON file - bundeslaender
GET {{host}}/api/v1/static/bundeslaender
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - enumerations
GET {{host}}/api/v1/static/enumerations
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - holidays
GET {{host}}/api/v1/static/holidays
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - jugendaemter
GET {{host}}/api/v1/static/jugendaemter
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - potential_customers
GET {{host}}/api/v1/static/potential_customers
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Security Tests - All should return 404 with security error

# Test path traversal protection - parent directory access
GET {{host}}/api/v1/static/../email_templates/welcome
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "404 page not found"

# Test path traversal with URL encoding
GET {{host}}/api/v1/static/..%2Femail_templates%2Fwelcome
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "404 page not found"

# Test accessing non-existent JSON file
GET {{host}}/api/v1/static/nonexistent_file
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "File not found"

# Test invalid characters - dot in filename
GET {{host}}/api/v1/static/test.json
Authorization: Bearer {{auth_token}}

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - special characters
GET {{host}}/api/v1/static/test@file
Authorization: Bearer {{auth_token}}

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - slash in filename
GET {{host}}/api/v1/static/test/file
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "404 page not found"

# Test invalid characters - space in filename
GET {{host}}/api/v1/static/test%20file
Authorization: Bearer {{auth_token}}

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - backslash
GET {{host}}/api/v1/static/test%5Cfile
Authorization: Bearer {{auth_token}}

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test case sensitivity - uppercase should work for valid files
GET {{host}}/api/v1/static/DIAGNOSTIC_STD
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "File not found"

# Test system file access attempts
GET {{host}}/api/v1/static/.env
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "File not found"

GET {{host}}/api/v1/static/.git
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "File not found"

# Test double path traversal
GET {{host}}/api/v1/static/../../etc/passwd
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "404 page not found"

# Test null byte injection
GET {{host}}/api/v1/static/diagnostic_std%00.html
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
body contains "File not found"

# Test legitimate filenames with allowed characters
# Only alphanumeric, hyphens, and underscores are allowed

# Test hyphen (should work)
GET {{host}}/api/v1/static/holidays
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test underscore (should work)
GET {{host}}/api/v1/static/potential_customers
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test mixed alphanumeric (should work)
GET {{host}}/api/v1/static/diagnostic_std
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test unauthenticated access - should return 401
GET {{host}}/api/v1/static

HTTP 401
[Asserts]
body contains "Authorization required"

GET {{host}}/api/v1/static/diagnostic_std

HTTP 401
[Asserts]
body contains "Authorization required"
