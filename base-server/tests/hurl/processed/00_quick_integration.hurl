# Quick Integration Test with Unique Users
# Minimal test to verify system is working before running full test suite

# Test server availability
GET http://localhost:8080/api/v1/health

HTTP 200
[Asserts]
jsonpath "$.status" == "healthy"
jsonpath "$.database" == "connected"

# Test basic authentication flow with unique user
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "username": "testuser_1761828406_5a9dd03d",
  "email": "test_1761828406_5a9dd03d@example.com",
  "password": "Pass123_5a9dd03d",
  "first_name": "Integration",
  "last_name": "Test",
  "company_name": "Test Company 1761828406_5a9dd03d",
  "accept_terms": true,
  "newsletter_opt_in": false
}

HTTP 201
[Asserts]
jsonpath "$.success" == true

# Test login with the created user
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser_1761828406_5a9dd03d",
  "password": "Pass123_5a9dd03d"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.token" exists

[Captures]
integration_token: jsonpath "$.data.token"

# Test accessing protected endpoint
GET http://localhost:8080/api/v1/user-settings
Authorization: Bearer {{integration_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true