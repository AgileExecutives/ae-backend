# Customer Tests
# Test customer CRUD operations

# Setup fresh auth session
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser",
  "password": "newpass123"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"
user_tenant_id: jsonpath "$.data.user.tenant_id"

# Test creating a customer
POST {{host}}/api/v1/customers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "Test Customer Inc",
  "email": "customer@testcorp.com",
  "phone": "+1-555-0123",
  "street": "123 Business Ave",
  "zip": "12345",
  "city": "Test City",
  "country": "US",
  "tax_id": "123456789",
  "vat": "US123456789",
  "plan_id": 1,
  "payment_method": "credit_card",
  "status": "active"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Customer created successfully"
jsonpath "$.data.name" == "Test Customer Inc"
jsonpath "$.data.email" == "customer@testcorp.com"
jsonpath "$.data.plan_id" == 1
jsonpath "$.data.tenant_id" == {{user_tenant_id}}
jsonpath "$.data.status" == "active"
jsonpath "$.data.active" == true

[Captures]
customer_id: jsonpath "$.data.id"

# Test creating customer with invalid plan
POST {{host}}/api/v1/customers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "Invalid Plan Customer",
  "email": "invalid@test.com",
  "plan_id": 999
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "plan"

# Test getting all customers
GET {{host}}/api/v1/customers
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
jsonpath "$.data[0].tenant_id" == {{user_tenant_id}}
jsonpath "$.pagination.page" == 1

# Test getting customers with filters
GET {{host}}/api/v1/customers?status=active&limit=5
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
jsonpath "$.pagination.limit" == 5

# Test getting customers with search
GET {{host}}/api/v1/customers?search=Test Customer
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection

# Test getting specific customer
GET {{host}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == {{customer_id}}
jsonpath "$.data.name" == "Test Customer Inc"
jsonpath "$.data.email" == "customer@testcorp.com"
jsonpath "$.data.tenant_id" == {{user_tenant_id}}

# Test getting non-existent customer
GET {{host}}/api/v1/customers/999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Customer not found"

# Test updating customer
PUT {{host}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "Updated Customer Inc",
  "email": "updated@testcorp.com",
  "phone": "+1-555-9999",
  "city": "Updated City",
  "status": "trial"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Customer updated successfully"
jsonpath "$.data.name" == "Updated Customer Inc"
jsonpath "$.data.email" == "updated@testcorp.com"
jsonpath "$.data.phone" == "+1-555-9999"
jsonpath "$.data.city" == "Updated City"
jsonpath "$.data.status" == "trial"

# Test partial update
PUT {{host}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "status": "active"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.status" == "active"
# Other fields should remain unchanged
jsonpath "$.data.name" == "Updated Customer Inc"

# Test updating non-existent customer
PUT {{host}}/api/v1/customers/999
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "Non-existent"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Customer not found"

# Create a second customer for testing multi-tenant isolation
POST {{host}}/api/v1/auth/register
Content-Type: application/json
{
  "username": "otheruser",
  "email": "other@example.com",
  "password": "otherpass123",
  "first_name": "Other",
  "last_name": "User",
  "organization_name": "Other Organization"
}

HTTP 201
[Captures]
other_token: jsonpath "$.data.token"

# Create customer in other organization
POST {{host}}/api/v1/customers
Authorization: Bearer {{other_token}}
Content-Type: application/json
{
  "name": "Other Org Customer",
  "email": "othercust@example.com",
  "plan_id": 1
}

HTTP 201
[Captures]
other_customer_id: jsonpath "$.data.id"

# Test that users cannot see customers from other organizations
GET {{host}}/api/v1/customers/{{other_customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Customer not found"

# Test that customer list is properly filtered by organization
GET {{host}}/api/v1/customers
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
# Should only contain customers from user's organization

# Test deleting customer (soft delete)
DELETE {{host}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Customer deleted successfully"

# Test that deleted customer is not accessible
GET {{host}}/api/v1/customers/{{customer_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Customer not found"

# Test deleting non-existent customer
DELETE {{host}}/api/v1/customers/999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Customer not found"

# Test accessing customer endpoints without authentication
GET {{host}}/api/v1/customers

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"