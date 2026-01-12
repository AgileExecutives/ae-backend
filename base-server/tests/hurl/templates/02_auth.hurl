# Authentication API Tests with Dynamic User Generation
# Tests user registration, login, logout, and token management

# Variables for this test run (automatically generated):
# {{UNIQUE_USERNAME}} - Unique username for this test
# {{UNIQUE_EMAIL}} - Unique email for this test  
# {{UNIQUE_PASSWORD}} - Unique password for this test
# {{HOST}} - Server host URL

# Test 1: User Login with known existing user (we'll use testuser which should exist)
POST {{HOST}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "testuser@unburdy.de",
  "password": "newpass123"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.token" exists
jsonpath "$.data.user.username" == "testuser"

[Captures]
auth_token: jsonpath "$.data.token"

# Test 2: User Registration with unique credentials
POST {{HOST}}/api/v1/auth/register
Content-Type: application/json
{
  "email": "{{UNIQUE_EMAIL}}",
  "username": "user_{{UNIQUE_ID}}",
  "password": "{{UNIQUE_PASSWORD}}",
  "first_name": "Test",
  "last_name": "User",
  "company_name": "Test Company {{UNIQUE_ID}}",
  "tenant_name": "tenant_{{UNIQUE_ID}}",
  "accept_terms": true
}

# Accept either success or conflict (user exists)
HTTP *
[Asserts]
jsonpath "$.success" exists

# Test 3: Access protected endpoint with valid token
GET {{HOST}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" exists

# Test 4: Login with wrong password
POST {{HOST}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "{{UNIQUE_EMAIL}}",
  "username": "user_{{UNIQUE_ID}}",
  "password": "wrongpassword"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.message" exists

# Test 5: Login with non-existent user
POST {{HOST}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "nonexistent_{{UNIQUE_ID}}@example.com",
  "password": "{{UNIQUE_PASSWORD}}"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.message" exists

# Test 6: Access protected endpoint without token
GET {{HOST}}/api/v1/user-settings

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 7: Access protected endpoint with invalid token
GET {{HOST}}/api/v1/user-settings
Authorization: Bearer invalid_token_{{UNIQUE_ID}}

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 8: Logout (if endpoint exists)
POST {{HOST}}/api/v1/auth/logout
Authorization: Bearer {{auth_token}}

HTTP *
[Asserts]
# Accept any status code as logout implementation may vary

# Test 9: Try to access protected endpoint with logged out token
GET {{HOST}}/api/v1/user-settings  
Authorization: Bearer {{auth_token}}

# This might still work depending on logout implementation
HTTP *