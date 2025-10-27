# Quick Integration Test
# Minimal test to verify system is working before running full test suite

# Test server availability
GET {{host}}/api/v1/health

HTTP 200
[Asserts]
jsonpath "$.status" == "healthy"
jsonpath "$.database" == "connected"

# Test basic authentication flow
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser",
  "password": "newpass123"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.token" isString

# Test static assets availability
GET {{host}}/statics/css/email.css

# Should return CSS or 404 if not set up
HTTP *

# Test PDF system availability (config endpoint)
GET {{host}}/api/v1/pdf/config

# Should return 401 (unauthorized) if system is working
HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "authorization"

# Test fuzzy search system availability
GET {{host}}/api/v1/search?q=test

# Should return 401 (unauthorized) if system is working
HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "authorization"