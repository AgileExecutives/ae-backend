# Client-Cost Provider Integration Tests
# Test the relationship between clients and cost providers

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

# Create a cost provider first
POST {{host}}/api/v1/cost-providers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "organization": "Integration Test Insurance",
  "department": "Mental Health Services",
  "contact_name": "Integration Contact",
  "street_address": "789 Integration Ave",
  "zip": "67890",
  "city": "Integration City"
}

HTTP 201
[Asserts]
jsonpath "$.organization" == "Integration Test Insurance"

[Captures]
integration_cost_provider_id: jsonpath "$.id"

# Create a client with the cost provider
POST {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "cost_provider_id": {{integration_cost_provider_id}},
  "first_name": "Integration",
  "last_name": "Client",
  "email": "integration.client@example.com",
  "phone": "+1555000111",
  "therapy_title": "Integration Therapy",
  "unit_price": 200.00,
  "status": "active"
}

HTTP 201
[Asserts]
jsonpath "$.first_name" == "Integration"
jsonpath "$.last_name" == "Client"
jsonpath "$.cost_provider_id" == {{integration_cost_provider_id}}

[Captures]
integration_client_id: jsonpath "$.id"

# Get the client and verify cost provider is included
GET {{host}}/api/v1/clients/{{integration_client_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.id" == {{integration_client_id}}
jsonpath "$.cost_provider_id" == {{integration_cost_provider_id}}
jsonpath "$.cost_provider" exists
jsonpath "$.cost_provider.organization" == "Integration Test Insurance"
jsonpath "$.cost_provider.department" == "Mental Health Services"

# Update client to remove cost provider association
PUT {{host}}/api/v1/clients/{{integration_client_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "cost_provider_id": 0
}

HTTP 200
[Asserts]
jsonpath "$.cost_provider_id" not exists

# Create another cost provider for testing association changes
POST {{host}}/api/v1/cost-providers
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "organization": "Secondary Insurance Co"
}

HTTP 201
[Captures]
secondary_cost_provider_id: jsonpath "$.id"

# Associate client with new cost provider
PUT {{host}}/api/v1/clients/{{integration_client_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "cost_provider_id": {{secondary_cost_provider_id}}
}

HTTP 200
[Asserts]
jsonpath "$.cost_provider_id" == {{secondary_cost_provider_id}}

# Verify the new association
GET {{host}}/api/v1/clients/{{integration_client_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.cost_provider.organization" == "Secondary Insurance Co"

# Test creating client with non-existent cost provider
POST {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "cost_provider_id": 999999,
  "first_name": "Invalid",
  "last_name": "Provider"
}

HTTP 500
[Asserts]
jsonpath "$.error" contains "foreign key constraint"

# Clean up - delete the integration test data
DELETE {{host}}/api/v1/clients/{{integration_client_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

DELETE {{host}}/api/v1/cost-providers/{{integration_cost_provider_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

DELETE {{host}}/api/v1/cost-providers/{{secondary_cost_provider_id}}
Authorization: Bearer {{auth_token}}

HTTP 200