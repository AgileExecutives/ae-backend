# Contacts API Tests with Unique Data
# Tests contact CRUD operations, filtering, and validation

# First, create a test user and get authentication token
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "email": "test_1768230441_496de870@example.com",
  "username": "user_1768230441_496de870",
  "password": "Pass123_496de870",
  "first_name": "Contact",
  "last_name": "Tester",
  "company_name": "Contact Test Company 1768230441_496de870",
  "tenant_name": "tenant_1768230441_496de870",
  "accept_terms": true
}

# Accept either success or conflict (user exists)
HTTP *

# Now login to get the token regardless of registration result
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "email": "test_1768230441_496de870@example.com",
  "password": "Pass123_496de870"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test 1: Create a new contact
POST http://localhost:8080/api/v1/contacts
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "1768230441_496de870_John",
  "last_name": "1768230441_496de870_Doe", 
  "email": "test_1768230441_496de870@example.com",
  "username": "user_1768230441_496de870",  "phone": "+1-555-0100",
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
jsonpath "$.data.first_name" == "1768230441_496de870_John"
jsonpath "$.data.email" == "test_1768230441_496de870@example.com"

[Captures]
contact_id: jsonpath "$.data.id"

# Test 2: Get all contacts
GET http://localhost:8080/api/v1/contacts
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 3: Get contact by ID
GET http://localhost:8080/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == {{contact_id}}
jsonpath "$.data.first_name" == "1768230441_496de870_John"

# Test 4: Update contact
PUT http://localhost:8080/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "1768230441_496de870_John_Updated",
  "last_name": "1768230441_496de870_Doe_Updated",
  "email": "updated_1768230441_496de870@example.com",
  "phone": "+1-555-0200",
  "type": "personal"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.first_name" == "1768230441_496de870_John_Updated"

# Test 5: Search contacts
GET http://localhost:8080/api/v1/contacts?search=1768230441_496de870_John
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 6: Filter contacts by type
GET http://localhost:8080/api/v1/contacts?type=personal
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 7: Try to access contacts without authentication
GET http://localhost:8080/api/v1/contacts

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 8: Delete contact
DELETE http://localhost:8080/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test 9: Try to get deleted contact
GET http://localhost:8080/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false