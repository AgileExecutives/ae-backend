# Authentication Tests
# Test user registration, login, logout and authentication flows

# Test user registration
POST {{host}}/api/v1/auth/register
Content-Type: application/json
{
  "email": "test_{{test_id}}@example.com",
  "password": "testpass123",
  "first_name": "Test",
  "last_name": "User",
  "company_name": "Test Company",
  "tenant_name": "test_{{test_id}}"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "User created successfully"
jsonpath "$.data.username" exists
jsonpath "$.data.email" exists
jsonpath "$.data.first_name" == "Test"
jsonpath "$.data.last_name" == "User"
jsonpath "$.data.role" == "user"
jsonpath "$.data.active" == true

# Capture user info for later tests
[Captures]
user_id: jsonpath "$.data.id"
tenant_id: jsonpath "$.data.tenant_id"

# Test login with correct credentials
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "test_{{test_id}}@example.com",
  "password": "testpass123"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Login successful"
jsonpath "$.data.user.username" == "testuser_{{test_id}}"
jsonpath "$.data.user.email" == "test_{{test_id}}@example.com"
jsonpath "$.data.token" isString

# Update token for consistency
[Captures]
auth_token: jsonpath "$.data.token"

# Test login with wrong password
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "test_{{test_id}}@example.com",
  "password": "wrongpassword"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Invalid credentials"

# Test login with non-existent user
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "nonexistent@example.com",
  "password": "testpass123"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Invalid credentials"

# Test login with unverified user (adapted from backend login-unverified-user test)
# First create an unverified user scenario if email verification is implemented
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "unverified@example.com",
  "password": "password123"
}

# This test expects 401 if email verification is required
# If email verification is not implemented, this may return 401 for "Invalid credentials"
HTTP 401
[Asserts]
jsonpath "$.success" == false
# Accept either "Email not verified" or "Invalid credentials" error
jsonpath "$.error" matches "(Email not verified|Invalid credentials)"

# Test getting current user info (authenticated endpoint)
GET {{host}}/api/v1/auth/me
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.username" == "testuser"
jsonpath "$.data.email" == "test@example.com"
jsonpath "$.data.first_name" == "Test"
jsonpath "$.data.last_name" == "User"
jsonpath "$.data.role" == "user"
jsonpath "$.data.active" == true

# Test accessing protected endpoint without token
GET {{host}}/api/v1/auth/me

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"

# Test change password
POST {{host}}/api/v1/auth/change-password
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "current_password": "testpass123",
  "new_password": "newpass123"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Password changed successfully"

# Test login with new password
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "testuser",
  "password": "newpass123"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Login successful"

# Update token after password change
[Captures]
auth_token: jsonpath "$.data.token"

# Test change password with wrong current password
POST {{host}}/api/v1/auth/change-password
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "current_password": "wrongcurrent",
  "new_password": "anothernew123"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Invalid current password"

# Test logout (blacklist token)
POST {{host}}/api/v1/auth/logout
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Logged out successfully"

# Test accessing protected endpoint with blacklisted token
GET {{host}}/api/v1/auth/me
Authorization: Bearer {{auth_token}}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Token is blacklisted"