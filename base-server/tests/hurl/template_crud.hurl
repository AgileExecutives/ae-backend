# Template CRUD Operations Integration Tests

# Setup - Create test templates for various scenarios

# Test GET /templates - Initial state (should have system default templates)
GET http://localhost:8081/api/v1/templates
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" exists
jsonpath "$.data" isCollection

# Test POST /templates - Create Email Template
POST http://localhost:8081/api/v1/templates
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_type": "email",
  "template_key": "integration_test_email",
  "channel": "EMAIL",
  "subject": "Integration Test Email",
  "name": "Integration Test Email Template",
  "description": "Template for integration testing",
  "content": "<h1>Hello \\{\\{.Name\\}\\}!</h1><p>This is a test email for \\{\\{.Purpose\\}\\}.</p><p>Contact: \\{\\{.ContactEmail\\}\\}</p>",
  "variables": ["Name", "Purpose", "ContactEmail"],
  "sample_data": {
    "Name": "Integration Tester",
    "Purpose": "API Testing",
    "ContactEmail": "test@integration.com"
  },
  "is_active": true,
  "is_default": false
}

HTTP 201
[Asserts]
jsonpath "$.data.id" exists
jsonpath "$.data.name" == "Integration Test Email Template"
jsonpath "$.data.template_type" == "email"
jsonpath "$.data.channel" == "EMAIL"
jsonpath "$.data.is_active" == true
[Captures]
email_template_id: jsonpath "$.data.id"

# Test POST /templates - Create PDF Template
POST http://localhost:8081/api/v1/templates
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "template_type": "document",
  "template_key": "integration_test_document",
  "channel": "PDF",
  "name": "Integration Test Document Template",
  "description": "PDF template for integration testing",
  "content": "<html><head><title>\\{\\{.Title\\}\\}</title></head><body><h1>\\{\\{.Title\\}\\}</h1><p>\\{\\{.Content\\}\\}</p><p>Generated on: \\{\\{.Date\\}\\}</p></body></html>",
  "variables": ["Title", "Content", "Date"],
  "sample_data": {
    "Title": "Integration Test Document",
    "Content": "This is test content for integration testing.",
    "Date": "2024-01-15"
  },
  "is_active": true,
  "is_default": false
}

HTTP 201
[Asserts]
jsonpath "$.data.id" exists
jsonpath "$.data.name" == "Integration Test Document Template"
jsonpath "$.data.channel" == "PDF"
[Captures]
pdf_template_id: jsonpath "$.data.id"

# Test GET /templates - Verify templates were created
GET http://localhost:8081/api/v1/templates
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data[?(@.name == 'Integration Test Email Template')]" exists
jsonpath "$.data[?(@.name == 'Integration Test Document Template')]" exists

# Test GET /templates/{id} - Get specific email template
GET http://localhost:8081/api/v1/templates/{{email_template_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data.id" == {{email_template_id}}
jsonpath "$.data.name" == "Integration Test Email Template"
jsonpath "$.data.content" contains "Hello \\{\\{.Name\\}\\}"
jsonpath "$.data.variables" includes "Name"
jsonpath "$.data.variables" includes "Purpose"
jsonpath "$.data.variables" includes "ContactEmail"

# Test PUT /templates/{id} - Update email template
PUT http://localhost:8081/api/v1/templates/{{email_template_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "name": "Updated Integration Test Email",
  "description": "Updated description for integration testing",
  "content": "<h1>Hi \\{\\{.Name\\}\\}!</h1><p>Updated content for \\{\\{.Purpose\\}\\}.</p><p>Updated contact: \\{\\{.ContactEmail\\}\\}</p><p>Version: \\{\\{.Version\\}\\}</p>",
  "variables": ["Name", "Purpose", "ContactEmail", "Version"],
  "sample_data": {
    "Name": "Updated Tester",
    "Purpose": "Updated API Testing",
    "ContactEmail": "updated@integration.com",
    "Version": "2.0"
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.name" == "Updated Integration Test Email"
jsonpath "$.data.description" == "Updated description for integration testing"
jsonpath "$.data.content" contains "Updated content"
jsonpath "$.data.variables" includes "Version"

# Test POST /templates/{id}/render - Render updated template
POST http://localhost:8081/api/v1/templates/{{email_template_id}}/render
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "data": {
    "Name": "John Doe",
    "Purpose": "CRUD Testing",
    "ContactEmail": "john@crudtest.com",
    "Version": "3.0"
  }
}

HTTP 200
[Asserts]
jsonpath "$.data.content" contains "John Doe"
jsonpath "$.data.content" contains "CRUD Testing"
jsonpath "$.data.content" contains "john@crudtest.com"
jsonpath "$.data.content" contains "3.0"

# Test POST /templates/{id}/duplicate - Duplicate email template
POST http://localhost:8081/api/v1/templates/{{email_template_id}}/duplicate
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "name": "Duplicated Integration Email",
  "template_key": "duplicated_integration_email"
}

HTTP 201
[Asserts]
jsonpath "$.data.name" == "Duplicated Integration Email"
jsonpath "$.data.template_key" == "duplicated_integration_email"
jsonpath "$.data.template_type" == "email"
jsonpath "$.data.content" contains "Updated content"
[Captures]
duplicated_email_id: jsonpath "$.data.id"

# Test GET /templates - Verify duplication
GET http://localhost:8081/api/v1/templates
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data[?(@.name == 'Updated Integration Test Email')]" exists
jsonpath "$.data[?(@.name == 'Duplicated Integration Email')]" exists

# Test Template Status Toggle - Deactivate template
PUT http://localhost:8081/api/v1/templates/{{email_template_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "is_active": false
}

HTTP 200
[Asserts]
jsonpath "$.data.is_active" == false

# Test Template Status Toggle - Reactivate template
PUT http://localhost:8081/api/v1/templates/{{email_template_id}}
Authorization: Bearer {{auth_token}}
Content-Type: application/json

{
  "is_active": true
}

HTTP 200
[Asserts]
jsonpath "$.data.is_active" == true

# Test Bulk Operations - Get templates by type
GET http://localhost:8081/api/v1/templates?template_type=email
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" isCollection
jsonpath "$.data[*].template_type" == "email"

# Test Bulk Operations - Get templates by channel
GET http://localhost:8081/api/v1/templates?channel=EMAIL
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.data" isCollection
jsonpath "$.data[*].channel" == "EMAIL"

# Cleanup - Delete duplicated template
DELETE http://localhost:8081/api/v1/templates/{{duplicated_email_id}}
Authorization: Bearer {{auth_token}}

HTTP 204

# Cleanup - Delete PDF template
DELETE http://localhost:8081/api/v1/templates/{{pdf_template_id}}
Authorization: Bearer {{auth_token}}

HTTP 204

# Cleanup - Delete email template
DELETE http://localhost:8081/api/v1/templates/{{email_template_id}}
Authorization: Bearer {{auth_token}}

HTTP 204

# Verify cleanup - Templates should be gone
GET http://localhost:8081/api/v1/templates/{{email_template_id}}
Authorization: Bearer {{auth_token}}

HTTP 404

GET http://localhost:8081/api/v1/templates/{{pdf_template_id}}
Authorization: Bearer {{auth_token}}

HTTP 404

GET http://localhost:8081/api/v1/templates/{{duplicated_email_id}}
Authorization: Bearer {{auth_token}}

HTTP 404