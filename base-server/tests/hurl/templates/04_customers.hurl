# Customer Management API Tests with Unique Data
# Tests customer CRUD operations, filtering, and validation

# First, create a test user and get authentication token
POST {{HOST}}/api/v1/auth/register
Content-Type: application/json
{
  "email": "{{UNIQUE_EMAIL}}",
  "username": "user_{{UNIQUE_ID}}",
  "password": "{{UNIQUE_PASSWORD}}",
  "first_name": "Customer",
  "last_name": "Tester",
  "company_name": "Customer Test Company {{UNIQUE_ID}}",
  "tenant_name": "tenant_{{UNIQUE_ID}}",
  "accept_terms": true
}

# Accept either success or conflict (user exists)
HTTP *

# Now login to get the token regardless of registration result
POST {{HOST}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "{{UNIQUE_EMAIL}}",
  "password": "{{UNIQUE_PASSWORD}}"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test 1: Create a new customer
POST {{HOST}}/api/v1/customers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "{{UNIQUE_CUSTOMER}} Corp",
  "email": "{{UNIQUE_EMAIL}}",
  "username": "user_{{UNIQUE_ID}}",
  "phone": "+1-555-123456",
  "address": "123 Test Street, Test City, TC 12345",
  "tenant_id": 1,
  "plan_id": 1
}

HTTP *
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "{{UNIQUE_CUSTOMER}} Corp"
jsonpath "$.data.email" == "{{UNIQUE_EMAIL}}"

[Captures]
customer_id: jsonpath "$.data.id"

# Test 2: Get all customers
GET {{HOST}}/api/v1/customers
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection
jsonpath "$.data.data[*].name" contains "{{UNIQUE_CUSTOMER}} Corp"

# Test 3: Get customer by ID
GET {{HOST}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == {{customer_id}}
jsonpath "$.data.name" == "{{UNIQUE_CUSTOMER}} Corp"

# Test 4: Update customer
PUT {{HOST}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "{{UNIQUE_CUSTOMER}} Updated Corp",
  "email": "updated_{{UNIQUE_ID}}@example.com",
  "phone": "+1-555-{{UNIQUE_ID}}-UPD",
  "address": "456 Updated Street, Update City, UC 67890"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "{{UNIQUE_CUSTOMER}} Updated Corp"
jsonpath "$.data.email" == "updated_{{UNIQUE_ID}}@example.com"

# Test 5: Search customers by name
GET {{HOST}}/api/v1/customers?search={{UNIQUE_CUSTOMER}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 6: Filter customers by email
GET {{HOST}}/api/v1/customers?email=updated_{{UNIQUE_EMAIL}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 7: Create customer with missing required fields
POST {{HOST}}/api/v1/customers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "email": "incomplete_{{UNIQUE_ID}}@example.com"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

# Test 8: Try to access customers without authentication
GET {{HOST}}/api/v1/customers

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 9: Delete customer
DELETE {{HOST}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test 10: Try to get deleted customer
GET {{HOST}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false