# Contacts API Tests with Unique Data
# Tests contact CRUD operations, filtering, and validation

# First, create a test user and get authentication token
POST {{HOST}}/api/v1/auth/register
Content-Type: application/json
{
  "email": "{{UNIQUE_EMAIL}}",
  "username": "user_{{UNIQUE_ID}}",
  "password": "{{UNIQUE_PASSWORD}}",
  "first_name": "Contact",
  "last_name": "Tester",
  "company_name": "Contact Test Company {{UNIQUE_ID}}",
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

# Test 1: Create a new contact
POST {{HOST}}/api/v1/contacts
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "{{UNIQUE_ID}}_John",
  "last_name": "{{UNIQUE_ID}}_Doe", 
  "email": "{{UNIQUE_EMAIL}}",
  "username": "user_{{UNIQUE_ID}}",  "phone": "+1-555-0100",
  "mobile": "+1-555-0101",
  "street": "456 Contact Street",
  "zip": "54321",
  "city": "Contact City",
  "country": "US",
  "type": "business",
  "tenant_id": 1
}

HTTP *
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.first_name" == "{{UNIQUE_ID}}_John"
jsonpath "$.data.email" == "{{UNIQUE_EMAIL}}"

[Captures]
contact_id: jsonpath "$.data.id"

# Test 2: Get all contacts
GET {{HOST}}/api/v1/contacts
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 3: Get contact by ID
GET {{HOST}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == {{contact_id}}
jsonpath "$.data.first_name" == "{{UNIQUE_ID}}_John"

# Test 4: Update contact
PUT {{HOST}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "{{UNIQUE_ID}}_John_Updated",
  "last_name": "{{UNIQUE_ID}}_Doe_Updated",
  "email": "updated_{{UNIQUE_ID}}@example.com",
  "phone": "+1-555-0200",
  "type": "personal"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.first_name" == "{{UNIQUE_ID}}_John_Updated"

# Test 5: Search contacts
GET {{HOST}}/api/v1/contacts?search={{UNIQUE_ID}}_John
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 6: Filter contacts by type
GET {{HOST}}/api/v1/contacts?type=personal
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 7: Try to access contacts without authentication
GET {{HOST}}/api/v1/contacts

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 8: Delete contact
DELETE {{HOST}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test 9: Try to get deleted contact
GET {{HOST}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false