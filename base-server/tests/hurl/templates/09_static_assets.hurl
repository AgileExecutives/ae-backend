# Static JSON API Tests
# Test secure JSON file serving with enhanced security validation

# Test listing available JSON files
GET {{HOST}}/api/v1/static

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.available_files" exists
jsonpath "$.available_files[*]" includes "diagnostic_std"
jsonpath "$.available_files[*]" includes "bundeslaender"
jsonpath "$.available_files[*]" includes "enumerations"
jsonpath "$.security_note" exists

# Test serving legitimate JSON file - diagnostic_std
GET {{HOST}}/api/v1/static/diagnostic_std

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
body contains "icd10"
body contains "kategorie"

# Test serving legitimate JSON file - bundeslaender
GET {{HOST}}/api/v1/static/bundeslaender

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - enumerations
GET {{HOST}}/api/v1/static/enumerations

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - ferien-feiertage
GET {{HOST}}/api/v1/static/ferien-feiertage

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - jugendaemter
GET {{HOST}}/api/v1/static/jugendaemter

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test serving legitimate JSON file - potential_customers
GET {{HOST}}/api/v1/static/potential_customers

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Security Tests - All should return 404 with security error

# Test path traversal protection - parent directory access
GET {{HOST}}/api/v1/static/../email_templates/welcome

HTTP 404
[Asserts]
body contains "404 page not found"

# Test path traversal with URL encoding
GET {{HOST}}/api/v1/static/..%2Femail_templates%2Fwelcome

HTTP 404
[Asserts]
body contains "404 page not found"

# Test accessing non-existent JSON file
GET {{HOST}}/api/v1/static/nonexistent_{{UNIQUE_ID}}

HTTP 404
[Asserts]
body contains "File not found"

# Test invalid characters - dot in filename
GET {{HOST}}/api/v1/static/test.json

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - special characters
GET {{HOST}}/api/v1/static/test@file

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - slash in filename
GET {{HOST}}/api/v1/static/test/file

HTTP 404
[Asserts]
body contains "404 page not found"

# Test invalid characters - space in filename
GET {{HOST}}/api/v1/static/test%20file

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test invalid characters - backslash
GET {{HOST}}/api/v1/static/test%5Cfile

HTTP 400
[Asserts]
body contains "Invalid file name"

# Test case sensitivity - uppercase should work for valid files
GET {{HOST}}/api/v1/static/DIAGNOSTIC_STD

HTTP 404
[Asserts]
body contains "File not found"

# Test system file access attempts
GET {{HOST}}/api/v1/static/.env

HTTP 404
[Asserts]
body contains "File not found"

GET {{HOST}}/api/v1/static/.git

HTTP 404
[Asserts]
body contains "File not found"

# Test double path traversal
GET {{HOST}}/api/v1/static/../../etc/passwd

HTTP 404
[Asserts]
body contains "404 page not found"

# Test null byte injection
GET {{HOST}}/api/v1/static/diagnostic_std%00.html

HTTP 404
[Asserts]
body contains "File not found"

# Test legitimate filenames with allowed characters
# Only alphanumeric, hyphens, and underscores are allowed

# Test hyphen (should work)
GET {{HOST}}/api/v1/static/ferien-feiertage

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test underscore (should work)
GET {{HOST}}/api/v1/static/potential_customers

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"

# Test mixed alphanumeric (should work)
GET {{HOST}}/api/v1/static/diagnostic_std

HTTP 200
[Asserts]
header "Content-Type" contains "application/json"