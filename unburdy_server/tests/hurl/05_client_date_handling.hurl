# Client Date Handling Tests
# Test empty string date conversion to null

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

# Test creating a client with empty date strings (should convert to null)
POST {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "DateTest",
  "last_name": "EmptyDates",
  "date_of_birth": "",
  "provider_approval_date": "",
  "admission_date": ""
}

HTTP 201
[Asserts]
jsonpath "$.first_name" == "DateTest"
jsonpath "$.last_name" == "EmptyDates"
jsonpath "$.date_of_birth" not exists
jsonpath "$.provider_approval_date" not exists
jsonpath "$.admission_date" not exists

[Captures]
empty_dates_client_id: jsonpath "$.id"

# Test updating a client with empty date strings (should convert to null)
PUT {{host}}/api/v1/clients/{{empty_dates_client_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "UpdatedDateTest",
  "date_of_birth": "",
  "provider_approval_date": "",
  "admission_date": ""
}

HTTP 200
[Asserts]
jsonpath "$.first_name" == "UpdatedDateTest"
jsonpath "$.date_of_birth" not exists
jsonpath "$.provider_approval_date" not exists
jsonpath "$.admission_date" not exists

# Test creating a client with mix of empty and valid dates
POST {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "MixedDates",
  "last_name": "TestClient", 
  "date_of_birth": "1990-05-15",
  "provider_approval_date": "",
  "admission_date": "2025-01-01"
}

HTTP 201
[Asserts]
jsonpath "$.first_name" == "MixedDates"
jsonpath "$.last_name" == "TestClient"
jsonpath "$.date_of_birth" == "1990-05-15T00:00:00Z"
jsonpath "$.provider_approval_date" not exists
jsonpath "$.admission_date" == "2025-01-01T00:00:00Z"

[Captures]
mixed_dates_client_id: jsonpath "$.id"

# Test updating with mix of empty and valid dates
PUT {{host}}/api/v1/clients/{{mixed_dates_client_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "date_of_birth": "",
  "provider_approval_date": "2025-01-15",
  "admission_date": ""
}

HTTP 200
[Asserts]
jsonpath "$.date_of_birth" not exists
jsonpath "$.provider_approval_date" == "2025-01-15T00:00:00Z"
jsonpath "$.admission_date" not exists

# Test with null values in JSON (should also work)
POST {{host}}/api/v1/clients
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "NullDates",
  "last_name": "TestClient",
  "date_of_birth": null,
  "provider_approval_date": null,
  "admission_date": null
}

HTTP 201
[Asserts]
jsonpath "$.first_name" == "NullDates"
jsonpath "$.last_name" == "TestClient"
jsonpath "$.date_of_birth" not exists
jsonpath "$.provider_approval_date" not exists
jsonpath "$.admission_date" not exists

# Clean up test clients
DELETE {{host}}/api/v1/clients/{{empty_dates_client_id}}
Authorization: Bearer {{auth_token}}

HTTP 200

DELETE {{host}}/api/v1/clients/{{mixed_dates_client_id}}
Authorization: Bearer {{auth_token}}

HTTP 200