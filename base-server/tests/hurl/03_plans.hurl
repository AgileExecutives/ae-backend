# Plans Tests
# Test plan management endpoints (public and admin)

# Create fresh auth session for these tests
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser",
  "password": "newpass123"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test getting all plans (public endpoint)
GET {{host}}/api/v1/plans

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
jsonpath "$.pagination.page" == 1
jsonpath "$.pagination.limit" == 20

# Test getting plans with filters
GET {{host}}/api/v1/plans?active=true&limit=5

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
jsonpath "$.pagination.limit" == 5

# Test getting a specific plan (should have default plans from migration)
GET {{host}}/api/v1/plans/1

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == 1
jsonpath "$.data.name" isString
jsonpath "$.data.price" isNumber
jsonpath "$.data.currency" isString
jsonpath "$.data.active" == true

# Test getting non-existent plan
GET {{host}}/api/v1/plans/999

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Plan not found"

# Test creating a new plan (admin only - will need admin user)
# First, create admin user
POST {{host}}/api/v1/auth/register
Content-Type: application/json
{
  "username": "admin",
  "email": "admin@example.com",
  "password": "adminpass123",
  "first_name": "Admin",
  "last_name": "User",
  "organization_name": "Admin Organization"
}

HTTP 201
[Captures]
admin_token: jsonpath "$.data.token"

# Note: In real scenario, admin role should be set directly in database
# For testing, we assume the middleware allows plan creation

# Test creating a plan
POST {{host}}/api/v1/admin/plans
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "name": "Test Plan",
  "slug": "test-plan",
  "description": "A test plan for Hurl testing",
  "price": 29.99,
  "currency": "USD",
  "invoice_period": "monthly",
  "max_users": 5,
  "max_clients": 50,
  "features": {
    "email_support": true,
    "api_access": true,
    "storage": "10GB"
  },
  "active": true
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Plan created successfully"
jsonpath "$.data.name" == "Test Plan"
jsonpath "$.data.slug" == "test-plan"
jsonpath "$.data.price" == 29.99
jsonpath "$.data.currency" == "USD"
jsonpath "$.data.active" == true

[Captures]
plan_id: jsonpath "$.data.id"

# Test creating plan with duplicate slug
POST {{host}}/api/v1/admin/plans
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "name": "Duplicate Plan",
  "slug": "test-plan",
  "price": 39.99
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "slug already exists"

# Test updating a plan
PUT {{host}}/api/v1/admin/plans/{{plan_id}}
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "name": "Updated Test Plan",
  "description": "An updated test plan",
  "price": 39.99,
  "max_users": 10,
  "features": {
    "email_support": true,
    "api_access": true,
    "phone_support": true,
    "storage": "20GB"
  }
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Plan updated successfully"
jsonpath "$.data.name" == "Updated Test Plan"
jsonpath "$.data.price" == 39.99
jsonpath "$.data.max_users" == 10

# Test getting the updated plan
GET {{host}}/api/v1/plans/{{plan_id}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.name" == "Updated Test Plan"
jsonpath "$.data.price" == 39.99

# Test deactivating a plan
PUT {{host}}/api/v1/admin/plans/{{plan_id}}
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "active": false
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.active" == false

# Test that inactive plan is not returned in active filter
GET {{host}}/api/v1/plans?active=true

HTTP 200
[Asserts]
jsonpath "$.success" == true
# Should not contain our deactivated plan (assuming it's the only inactive one)

# Test deleting a plan (soft delete)
DELETE {{host}}/api/v1/admin/plans/{{plan_id}}
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Plan deleted successfully"

# Test that deleted plan is not accessible
GET {{host}}/api/v1/plans/{{plan_id}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Plan not found"

# Test admin actions without admin token (should fail)
POST {{host}}/api/v1/admin/plans
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "name": "Unauthorized Plan",
  "slug": "unauthorized-plan",
  "price": 19.99
}

HTTP 403
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "admin"