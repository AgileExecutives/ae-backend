# Authentication Setup for Client Management Tests
# This file sets up authentication for use in other test files

# Test login to get authentication token for client management tests
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser@unburdy.de",
  "password": "newpass123"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Login successful"
jsonpath "$.data.token" isString

# Capture auth token for use in other tests
[Captures]
auth_token: jsonpath "$.data.token"
user_id: jsonpath "$.data.user.id"
tenant_id: jsonpath "$.data.user.tenant_id"