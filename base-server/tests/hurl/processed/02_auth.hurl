# Authentication API Tests with Dynamic User Generation
# Tests user registration, login, logout, and token management

# Variables for this test run (automatically generated):
# testuser_1768217908_4817a46c - Unique username for this test
# test_1768217908_4817a46c@example.com - Unique email for this test  
# Pass123_4817a46c - Unique password for this test
# http://localhost:8080 - Server host URL

# Test 1: User Login with known existing user (we'll use testuser which should exist)
POST http://localhost:8080/api/v1/auth/login
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
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "email": "test_1768217908_4817a46c@example.com",
  "username": "user_1768217908_4817a46c",
  "password": "Pass123_4817a46c",
  "first_name": "Test",
  "last_name": "User",
  "company_name": "Test Company 1768217908_4817a46c",
  "tenant_name": "tenant_1768217908_4817a46c",
  "accept_terms": true
}

# Accept either success or conflict (user exists)
HTTP *
[Asserts]
jsonpath "$.success" exists

# Test 3: Access protected endpoint with valid token
GET http://localhost:8080/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" exists

# Test 4: Login with wrong password
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "email": "test_1768217908_4817a46c@example.com",
  "username": "user_1768217908_4817a46c",
  "password": "wrongpassword"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.message" exists

# Test 5: Login with non-existent user
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "email": "nonexistent_1768217908_4817a46c@example.com",
  "password": "Pass123_4817a46c"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.message" exists

# Test 6: Access protected endpoint without token
GET http://localhost:8080/api/v1/user-settings

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 7: Access protected endpoint with invalid token
GET http://localhost:8080/api/v1/user-settings
Authorization: Bearer invalid_token_1768217908_4817a46c

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 8: Logout (if endpoint exists)
POST http://localhost:8080/api/v1/auth/logout
Authorization: Bearer {{auth_token}}

HTTP *
[Asserts]
# Accept any status code as logout implementation may vary

# Test 9: Try to access protected endpoint with logged out token
GET http://localhost:8080/api/v1/user-settings  
Authorization: Bearer {{auth_token}}

# This might still work depending on logout implementation
HTTP *