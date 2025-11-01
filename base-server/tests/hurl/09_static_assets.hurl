# Static JSON API Tests
# Test secure JSON file serving with enhanced security validation

# Test listing available JSON files
GET {{host}}/api/v1/static

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

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
body contains "icd10"
body contains "category"

# Test serving legitimate JSON file - bundeslaender
GET {{host}}/api/v1/static/bundeslaender

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - enumerations
GET {{host}}/api/v1/static/enumerations

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - holidays
GET {{host}}/api/v1/static/holidays

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - jugendaemter
GET {{host}}/api/v1/static/jugendaemter

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - potential_customers
GET {{host}}/api/v1/static/potential_customers

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Security Tests - All should return 404 with security error

# Test path traversal protection - parent directory access
GET {{host}}/api/v1/static/../email_templates/welcome

HTTP 404
[Asserts]
body contains "404 page not found"

# Test path traversal with URL encoding
GET {{host}}/api/v1/static/..%2Femail_templates%2Fwelcome

HTTP 404
[Asserts]
body contains "404 page not found"

# Test accessing non-existent JSON file
GET {{host}}/api/v1/static/nonexistent_file

HTTP 404
[Asserts]
body contains "File not found"

# Test invalid characters - dot in filename
GET {{host}}/api/v1/static/test.json

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - special characters
GET {{host}}/api/v1/static/test@file

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - slash in filename
GET {{host}}/api/v1/static/test/file

HTTP 404
[Asserts]
body contains "404 page not found"

# Test invalid characters - space in filename
GET {{host}}/api/v1/static/test%20file

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - backslash
GET {{host}}/api/v1/static/test%5Cfile

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test case sensitivity - uppercase should work for valid files
GET {{host}}/api/v1/static/DIAGNOSTIC_STD

HTTP 404
[Asserts]
body contains "File not found"

# Test system file access attempts
GET {{host}}/api/v1/static/.env

HTTP 404
[Asserts]
body contains "File not found"

GET {{host}}/api/v1/static/.git

HTTP 404
[Asserts]
body contains "File not found"

# Test double path traversal
GET {{host}}/api/v1/static/../../etc/passwd

HTTP 404
[Asserts]
body contains "404 page not found"

# Test null byte injection
GET {{host}}/api/v1/static/diagnostic_std%00.html

HTTP 404
[Asserts]
body contains "File not found"

# Test legitimate filenames with allowed characters
# Only alphanumeric, hyphens, and underscores are allowed

# Test hyphen (should work)
GET {{host}}/api/v1/static/holidays

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test underscore (should work)
GET {{host}}/api/v1/static/potential_customers

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test mixed alphanumeric (should work)
GET {{host}}/api/v1/static/diagnostic_std

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"