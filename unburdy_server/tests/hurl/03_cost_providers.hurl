# Cost Provider Management Tests
# Test cost provider CRUD operations and pagination with environment variables

# Setup fresh auth session
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser@unburdy.de",
  "password": "newpass123"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test creating a cost provider with full data
POST {{host}}/api/v1/cost-providers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "organization": "Health Insurance Corp",
  "department": "Mental Health Division",
  "contact_name": "Jane Smith",
  "street_address": "456 Insurance Blvd",
  "zip": "12345",
  "city": "New York"
}

HTTP 201
[Asserts]
jsonpath "$.organization" == "Health Insurance Corp"
jsonpath "$.department" == "Mental Health Division"
jsonpath "$.contact_name" == "Jane Smith"
jsonpath "$.tenant_id" exists
jsonpath "$.id" exists

[Captures]
cost_provider_id: jsonpath "$.id"

# Test creating a cost provider with minimal data
POST {{host}}/api/v1/cost-providers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "organization": "Basic Insurance"
}

HTTP 201
[Asserts]
jsonpath "$.organization" == "Basic Insurance"

[Captures]
minimal_cost_provider_id: jsonpath "$.id"

# Test creating cost provider with invalid data (missing required fields)
POST {{host}}/api/v1/cost-providers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "department": "Some Department"
}

HTTP 400
[Asserts]
jsonpath "$.error" contains "Field validation"

# Test getting all cost providers (should respect DEFAULT_PAGE_LIMIT=200)
GET {{host}}/api/v1/cost-providers
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.cost_providers" isCollection
jsonpath "$.pagination.page" == 1
jsonpath "$.pagination.limit" == 200
jsonpath "$.pagination.total" exists

# Test getting cost providers with custom page size (within MAX_PAGE_LIMIT=1000)
GET {{host}}/api/v1/cost-providers?page=1&limit=500
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.cost_providers" isCollection
jsonpath "$.pagination.limit" == 500

# Test getting cost providers with page size exceeding MAX_PAGE_LIMIT (should use default)
GET {{host}}/api/v1/cost-providers?page=1&limit=2000
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.cost_providers" isCollection
jsonpath "$.pagination.limit" == 200

# Test getting cost providers with small limit
GET {{host}}/api/v1/cost-providers?limit=10
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.cost_providers" isCollection
jsonpath "$.pagination.limit" == 10

# Test searching cost providers
GET {{host}}/api/v1/cost-providers/search?q=Insurance&page=1&limit=100
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.cost_providers" isCollection
jsonpath "$.pagination.limit" == 100

# Test searching with invalid query (empty search term)
GET {{host}}/api/v1/cost-providers/search
Authorization: Bearer {{auth_token}}

HTTP 400
[Asserts]
jsonpath "$.error" == "Search query is required"

# Test getting specific cost provider
GET {{host}}/api/v1/cost-providers/{{cost_provider_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.id" == {{cost_provider_id}}
jsonpath "$.organization" == "Health Insurance Corp"
jsonpath "$.department" == "Mental Health Division"
jsonpath "$.contact_name" == "Jane Smith"

# Test getting non-existent cost provider
GET {{host}}/api/v1/cost-providers/999999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test updating cost provider
PUT {{host}}/api/v1/cost-providers/{{cost_provider_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "organization": "Updated Insurance Corp",
  "contact_name": "John Updated",
  "city": "Los Angeles"
}

HTTP 200
[Asserts]
jsonpath "$.organization" == "Updated Insurance Corp"
jsonpath "$.contact_name" == "John Updated"
jsonpath "$.city" == "Los Angeles"

# Test partial update (only one field)
PUT {{host}}/api/v1/cost-providers/{{cost_provider_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "zip": "54321"
}

HTTP 200
[Asserts]
jsonpath "$.zip" == "54321"
# Other fields should remain unchanged
jsonpath "$.organization" == "Updated Insurance Corp"

# Test updating non-existent cost provider
PUT {{host}}/api/v1/cost-providers/999999
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "organization": "Non-existent"
}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test deleting cost provider (soft delete)
DELETE {{host}}/api/v1/cost-providers/{{cost_provider_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.message" contains "deleted"

# Test that deleted cost provider is not accessible
GET {{host}}/api/v1/cost-providers/{{cost_provider_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test deleting non-existent cost provider
DELETE {{host}}/api/v1/cost-providers/999999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test accessing cost provider endpoints without authentication
# Test unauthorized access
GET {{host}}/api/v1/cost-providers

HTTP 401
[Asserts]
jsonpath "$.error" == "Missing authorization header"

# Test cost provider creation without authentication
# Test unauthorized POST request
POST {{host}}/api/v1/cost-providers
Content-Type: application/json
{
  "organization": "Unauthorized Provider"
}

HTTP 401
[Asserts]
jsonpath "$.error" == "Missing authorization header"