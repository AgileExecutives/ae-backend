# Template Module API Tests

# Test GET /templates - List all templates
GET http://localhost:8081/api/v1/templates
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" exists
jsonpath "$.data" isCollection

# Test POST /templates - Create new template
POST http://localhost:8081/api/v1/templates
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_type": "email",
  "template_key": "test_welcome",
  "channel": "EMAIL",
  "subject": "Welcome to Test Platform",
  "name": "Test Welcome Template",
  "description": "A test welcome email template",
  "content": "<h1>Welcome \\{\\{.FirstName\\}\\} \\{\\{.LastName\\}\\}!</h1><p>Thank you for joining \\{\\{.OrganizationName\\}\\}.</p>",
  "variables": ["FirstName", "LastName", "OrganizationName"],
  "sample_data": {
    "FirstName": "John",
    "LastName": "Doe", 
    "OrganizationName": "Test Company"
  },
  "is_active": true,
  "is_default": false
}

HTTP 201
[Asserts]
jsonpath "$.data.id" exists
jsonpath "$.data.name" == "Test Welcome Template"
jsonpath "$.data.template_type" == "email"
jsonpath "$.data.is_active" == true
[Captures]
template_id: jsonpath "$.data.id"

# Test GET /templates/{id} - Get specific template
GET http://localhost:8081/api/v1/templates/{{template_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.id" == {{template_id}}
jsonpath "$.data.name" == "Test Welcome Template"
jsonpath "$.data.content" contains "Welcome"

# Test PUT /templates/{id} - Update template
PUT http://localhost:8081/api/v1/templates/{{template_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "name": "Updated Welcome Template",
  "description": "Updated description for welcome template",
  "content": "<h1>Hello \\{\\{.FirstName\\}\\} \\{\\{.LastName\\}\\}!</h1><p>Welcome to \\{\\{.OrganizationName\\}\\}!</p>"
}

HTTP 200
[Asserts]
jsonpath "$.data.name" == "Updated Welcome Template"
jsonpath "$.data.description" == "Updated description for welcome template"

# Test POST /templates/{id}/render - Render template
POST http://localhost:8081/api/v1/templates/{{template_id}}/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "data": {
    "FirstName": "Jane",
    "LastName": "Smith",
    "OrganizationName": "Hurl Test Company"
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.content" contains "Jane Smith"
jsonpath "$.data.content" contains "Hurl Test Company"

# Test POST /templates/{id}/duplicate - Duplicate template
POST http://localhost:8081/api/v1/templates/{{template_id}}/duplicate
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "name": "Duplicated Welcome Template"
}

HTTP 201
[Asserts]
jsonpath "$.data.name" == "Duplicated Welcome Template"
jsonpath "$.data.template_type" == "email"
[Captures]
duplicated_template_id: jsonpath "$.data.id"

# Test GET /templates/default - Get default template
GET http://localhost:8081/api/v1/templates/default?template_type=email&channel=EMAIL
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" exists
jsonpath "$.data.is_default" == true

# Test GET /templates/contracts - List all template contracts
GET http://localhost:8081/api/v1/templates/contracts
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" exists
jsonpath "$.data" isCollection

# Test GET /templates/contracts/by-key/{template_key} - Get contract by key
GET http://localhost:8081/api/v1/templates/contracts/by-key/welcome
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.template_key" == "welcome"
jsonpath "$.data.variable_schema" exists

# Test DELETE /templates/{id} - Delete template
DELETE http://localhost:8081/api/v1/templates/{{template_id}}
Authorization: Bearer {{auth_token}}

HTTP 204

# Test DELETE /templates/{id} - Delete duplicated template
DELETE http://localhost:8081/api/v1/templates/{{duplicated_template_id}}
Authorization: Bearer {{auth_token}}

HTTP 204

# Verify template was deleted
GET http://localhost:8081/api/v1/templates/{{template_id}}
Authorization: Bearer {{auth_token}}

HTTP 404