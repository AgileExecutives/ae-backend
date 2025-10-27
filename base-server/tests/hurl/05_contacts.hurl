# Contact Tests
# Test contact CRUD operations

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

# Test creating a contact
POST {{host}}/api/v1/contacts
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john.doe@example.com",
  "phone": "+1-555-0100",
  "mobile": "+1-555-0101",
  "street": "456 Contact Street",
  "zip": "54321",
  "city": "Contact City",
  "country": "US",
  "type": "business",
  "notes": "Important business contact"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Contact created successfully"
jsonpath "$.data.first_name" == "John"
jsonpath "$.data.last_name" == "Doe"
jsonpath "$.data.email" == "john.doe@example.com"
jsonpath "$.data.phone" == "+1-555-0100"
jsonpath "$.data.type" == "business"
jsonpath "$.data.active" == true

[Captures]
contact_id: jsonpath "$.data.id"

# Test creating contact with minimal data
POST {{host}}/api/v1/contacts
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "Jane",
  "last_name": "Smith"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.first_name" == "Jane"
jsonpath "$.data.last_name" == "Smith"
jsonpath "$.data.type" == "contact"
jsonpath "$.data.active" == true

[Captures]
minimal_contact_id: jsonpath "$.data.id"

# Test creating contact with invalid data
POST {{host}}/api/v1/contacts
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "email": "invalid-email"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "validation"

# Test getting all contacts
GET {{host}}/api/v1/contacts
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
jsonpath "$.pagination.page" == 1
jsonpath "$.pagination.limit" == 20

# Test getting contacts with filters
GET {{host}}/api/v1/contacts?type=business&limit=5
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
jsonpath "$.pagination.limit" == 5

# Test getting contacts with search
GET {{host}}/api/v1/contacts?search=John
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection

# Test getting contacts by active status
GET {{host}}/api/v1/contacts?active=true
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection

# Test getting specific contact
GET {{host}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == {{contact_id}}
jsonpath "$.data.first_name" == "John"
jsonpath "$.data.last_name" == "Doe"
jsonpath "$.data.email" == "john.doe@example.com"

# Test getting non-existent contact
GET {{host}}/api/v1/contacts/999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Contact not found"

# Test updating contact
PUT {{host}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "Johnny",
  "email": "johnny.doe@example.com",
  "mobile": "+1-555-9999",
  "type": "personal",
  "notes": "Updated notes for Johnny"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Contact updated successfully"
jsonpath "$.data.first_name" == "Johnny"
jsonpath "$.data.email" == "johnny.doe@example.com"
jsonpath "$.data.mobile" == "+1-555-9999"
jsonpath "$.data.type" == "personal"

# Test partial update
PUT {{host}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "city": "New City"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.city" == "New City"
# Other fields should remain unchanged
jsonpath "$.data.first_name" == "Johnny"

# Test updating contact to inactive
PUT {{host}}/api/v1/contacts/{{minimal_contact_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "active": false
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.active" == false

# Test that inactive contacts are filtered when requested
GET {{host}}/api/v1/contacts?active=true
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test updating non-existent contact
PUT {{host}}/api/v1/contacts/999
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "first_name": "Non-existent"
}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Contact not found"

# Test deleting contact (soft delete)
DELETE {{host}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Contact deleted successfully"

# Test that deleted contact is not accessible
GET {{host}}/api/v1/contacts/{{contact_id}}
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Contact not found"

# Test deleting non-existent contact
DELETE {{host}}/api/v1/contacts/999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Contact not found"

# Test accessing contact endpoints without authentication
GET {{host}}/api/v1/contacts

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"

# Test contact creation without authentication
POST {{host}}/api/v1/contacts
Content-Type: application/json
{
  "first_name": "Unauthorized",
  "last_name": "User"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"