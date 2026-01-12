# Customer Management API Tests with Unique Data
# Tests customer CRUD operations, filtering, and validation

# First, create a test user and get authentication token
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "email": "test_1768217908_4817a46c@example.com",
  "username": "user_1768217908_4817a46c",
  "password": "Pass123_4817a46c",
  "first_name": "Customer",
  "last_name": "Tester",
  "company_name": "Customer Test Company 1768217908_4817a46c",
  "tenant_name": "tenant_1768217908_4817a46c",
  "accept_terms": true
}

# Accept either success or conflict (user exists)
HTTP *

# Now login to get the token regardless of registration result
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "email": "test_1768217908_4817a46c@example.com",
  "password": "Pass123_4817a46c"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test 1: Create a new customer
POST http://localhost:8080/api/v1/customers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "customer_1768217908_4817a46c Corp",
  "email": "test_1768217908_4817a46c@example.com",
  "username": "user_1768217908_4817a46c",  "phone": "+1-555-123456",
  "address": "123 Test Street, Test City, TC 12345",
  "tenant_id": 1,
  "plan_id": 1
}

HTTP *
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "customer_1768217908_4817a46c Corp"
jsonpath "$.data.email" == "test_1768217908_4817a46c@example.com"

[Captures]
customer_id: jsonpath "$.data.id"

# Test 2: Get all customers
GET http://localhost:8080/api/v1/customers
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection
jsonpath "$.data.data[*].name" contains "customer_1768217908_4817a46c Corp"

# Test 3: Get customer by ID
GET http://localhost:8080/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == {{customer_id}}
jsonpath "$.data.name" == "customer_1768217908_4817a46c Corp"

# Test 4: Update customer
PUT http://localhost:8080/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "customer_1768217908_4817a46c Updated Corp",
  "email": "updated_1768217908_4817a46c@example.com",
  "phone": "+1-555-1768217908_4817a46c-UPD",
  "address": "456 Updated Street, Update City, UC 67890"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "customer_1768217908_4817a46c Updated Corp"
jsonpath "$.data.email" == "updated_1768217908_4817a46c@example.com"

# Test 5: Search customers by name
GET http://localhost:8080/api/v1/customers?search=customer_1768217908_4817a46c
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 6: Filter customers by email
GET http://localhost:8080/api/v1/customers?email=updated_test_1768217908_4817a46c@example.com
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 7: Create customer with missing required fields
POST http://localhost:8080/api/v1/customers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "email": "incomplete_1768217908_4817a46c@example.com"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

# Test 8: Try to access customers without authentication
GET http://localhost:8080/api/v1/customers

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 9: Delete customer
DELETE http://localhost:8080/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test 10: Try to get deleted customer
GET http://localhost:8080/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false