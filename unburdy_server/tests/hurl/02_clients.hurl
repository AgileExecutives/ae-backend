# Client Management Tests
# Test client CRUD operations and pagination with environment variables

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

# Test creating a client with full data
POST {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "John",
  "last_name": "Doe",
  "date_of_birth": "1990-01-15T00:00:00Z",
  "gender": "male",
  "primary_language": "English",
  "contact_first_name": "Jane",
  "contact_last_name": "Smith",
  "contact_email": "jane.smith@example.com",
  "contact_phone": "+1234567890",
  "street_address": "123 Main Street",
  "zip": "12345",
  "city": "New York",
  "email": "john.doe@example.com",
  "phone": "+1234567890",
  "therapy_title": "Cognitive Behavioral Therapy",
  "unit_price": 150.00,
  "status": "waiting",
  "admission_date": "2025-01-01T00:00:00Z",
  "referral_source": "Doctor Smith",
  "notes": "Test client for API testing"
}

HTTP 201
[Asserts]
jsonpath "$.first_name" == "John"
jsonpath "$.last_name" == "Doe"
jsonpath "$.email" == "john.doe@example.com"
jsonpath "$.status" == "waiting"
jsonpath "$.unit_price" == 150.00
jsonpath "$.tenant_id" exists
jsonpath "$.id" exists

[Captures]
client_id: jsonpath "$.id"

# Test creating a client with minimal data
POST {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "Jane",
  "last_name": "Smith"
}

HTTP 201
[Asserts]
jsonpath "$.first_name" == "Jane"
jsonpath "$.last_name" == "Smith"
jsonpath "$.status" == "waiting"
jsonpath "$.gender" == "undisclosed"

[Captures]
minimal_client_id: jsonpath "$.id"

# Test creating client with invalid data (missing required fields)
POST {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "email": "invalid@test.com"
}

HTTP 400
[Asserts]
jsonpath "$.error" contains "Field validation"

# Test getting all clients (should respect DEFAULT_PAGE_LIMIT=200)
GET {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.clients" isCollection
jsonpath "$.pagination.page" == 1
jsonpath "$.pagination.limit" == 200
jsonpath "$.pagination.total" exists

# Test getting clients with custom page size (within MAX_PAGE_LIMIT=1000)
GET {{host}}/api/v1/clients?page=1&limit=500
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.clients" isCollection
jsonpath "$.pagination.limit" == 500

# Test getting clients with page size exceeding MAX_PAGE_LIMIT (should use default)
GET {{host}}/api/v1/clients?page=1&limit=2000
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.clients" isCollection
jsonpath "$.pagination.limit" == 200

# Test getting clients with filters
GET {{host}}/api/v1/clients?status=waiting&limit=50
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.clients" isCollection
jsonpath "$.pagination.limit" == 50

# Test searching clients
GET {{host}}/api/v1/clients/search?q=John&page=1&limit=100
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.clients" isCollection
jsonpath "$.pagination.limit" == 100

# Test searching with invalid query (empty search term)
GET {{host}}/api/v1/clients/search
Authorization: Bearer {{auth_token}}

HTTP 400
[Asserts]
jsonpath "$.error" == "Search query is required"

# Test getting specific client
GET {{host}}/api/v1/clients/{{client_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.id" == {{client_id}}
jsonpath "$.first_name" == "John"
jsonpath "$.last_name" == "Doe"
jsonpath "$.email" == "john.doe@example.com"

# Test getting non-existent client
GET {{host}}/api/v1/clients/999999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test updating client
PUT {{host}}/api/v1/clients/{{client_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "Johnny",
  "email": "johnny.doe@example.com",
  "status": "active",
  "unit_price": 175.00,
  "notes": "Updated notes for Johnny"
}

HTTP 200
[Asserts]
jsonpath "$.first_name" == "Johnny"
jsonpath "$.email" == "johnny.doe@example.com"
jsonpath "$.status" == "active"
jsonpath "$.unit_price" == 175

# Test partial update (only one field)
PUT {{host}}/api/v1/clients/{{client_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "city": "Los Angeles"
}

HTTP 200
[Asserts]
jsonpath "$.city" == "Los Angeles"
# Other fields should remain unchanged
jsonpath "$.first_name" == "Johnny"

# Test updating client status to archived
PUT {{host}}/api/v1/clients/{{minimal_client_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "status": "archived"
}

HTTP 200
[Asserts]
jsonpath "$.status" == "archived"

# Test updating non-existent client
PUT {{host}}/api/v1/clients/999999
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "Non-existent"
}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test deleting client (soft delete)
DELETE {{host}}/api/v1/clients/{{client_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.message" contains "deleted"

# Test that deleted client is not accessible
GET {{host}}/api/v1/clients/{{client_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test deleting non-existent client
DELETE {{host}}/api/v1/clients/999999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test accessing client endpoints without authentication
# Test unauthorized access
GET {{host}}/api/v1/clients

HTTP 401
[Asserts]
jsonpath "$.error" == "Missing authorization header"

# Test client creation without authentication
# Test unauthorized POST request
POST {{host}}/api/v1/clients
Content-Type: application/json
{
  "first_name": "Unauthorized",
  "last_name": "User"
}

HTTP 401
[Asserts]
jsonpath "$.error" == "Missing authorization header"