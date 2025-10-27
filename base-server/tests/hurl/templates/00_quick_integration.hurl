# Quick Integration Test with Unique Users
# Minimal test to verify system is working before running full test suite

# Test server availability
GET {{HOST}}/api/v1/health

HTTP 200
[Asserts]
jsonpath "$.status" == "healthy"
jsonpath "$.database" == "connected"

# Test basic authentication flow with unique user
POST {{HOST}}/api/v1/auth/register
Content-Type: application/json
{
  "username": "{{UNIQUE_USERNAME}}",
  "email": "{{UNIQUE_EMAIL}}",
  "password": "{{UNIQUE_PASSWORD}}",
  "first_name": "Integration",
  "last_name": "Test",
  "company_name": "Test Company {{UNIQUE_ID}}",
  "accept_terms": true,
  "newsletter_opt_in": false
}

HTTP 201
[Asserts]
jsonpath "$.success" == true

# Test login with the created user
POST {{HOST}}/api/v1/auth/login
Content-Type: application/json
{
  "username": "{{UNIQUE_USERNAME}}",
  "password": "{{UNIQUE_PASSWORD}}"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.token" exists

[Captures]
integration_token: jsonpath "$.data.token"

# Test accessing protected endpoint
GET {{HOST}}/api/v1/user-settings
Authorization: Bearer {{integration_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true