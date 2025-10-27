# Plans API Tests with Authentication
# Tests plan management endpoints (public and admin)

# First, authenticate to get access token
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser",
  "password": "newpass123"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test 1: Get all plans (public endpoint)
GET http://localhost:8080/api/v1/plans

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection
jsonpath "$.data.pagination.page" == 1
jsonpath "$.data.pagination.limit" >= 1

# Test 2: Get plans with filters
GET http://localhost:8080/api/v1/plans?active=true&limit=5

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 3: Get plans with pagination
GET http://localhost:8080/api/v1/plans?page=1&limit=10

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.pagination.page" == 1

# Test 4: Get specific plan by ID (if available)
GET http://localhost:8080/api/v1/plans/1

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == 1
jsonpath "$.data.name" exists
jsonpath "$.data.price" exists

# Test 5: Test plan endpoint without authentication (should still work for public access)
GET http://localhost:8080/api/v1/plans

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test 6: Test invalid plan ID
GET http://localhost:8080/api/v1/plans/999999

HTTP 404
[Asserts]
jsonpath "$.success" == false