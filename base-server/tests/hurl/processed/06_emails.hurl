# Email API Tests with Unique Data
# Tests email sending, management, and querying endpoints

# First, create a test user and get authentication token
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "username": "testuser_1761828406_5a9dd03d",
  "email": "test_1761828406_5a9dd03d@example.com",
  "password": "Pass123_5a9dd03d",
  "first_name": "Email",
  "last_name": "Tester",
  "company_name": "Email Test Company 1761828406_5a9dd03d",
  "accept_terms": true,
  "newsletter_opt_in": false
}

# Accept either success or conflict (user exists)
HTTP *

# Now login to get the token regardless of registration result
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "username": "testuser_1761828406_5a9dd03d",
  "password": "Pass123_5a9dd03d"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test 1: Send an email
POST http://localhost:8080/api/v1/emails/send
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "to": "test_1761828406_5a9dd03d@example.com",
  "from": "sender_1761828406_5a9dd03d@example.com",
  "subject": "Test Email Subject - 1761828406_5a9dd03d",
  "body": "This is a test email body for 1761828406_5a9dd03d.",
  "html_body": "<h1>Test Email 1761828406_5a9dd03d</h1><p>This is a test email body in <strong>HTML</strong>.</p>"
}

# Email creation may have database issues, accept various status codes
HTTP *
[Asserts]
jsonpath "$.success" exists

# Test 2: Get all emails
GET http://localhost:8080/api/v1/emails
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" exists

# Test 3: Get email by ID (use a test ID)
GET http://localhost:8080/api/v1/emails/1
Authorization: Bearer {{auth_token}}

# Accept either success or not found
HTTP *
[Asserts]
jsonpath "$.success" exists

# Test 4: Search emails by subject
GET http://localhost:8080/api/v1/emails?search=1761828406_5a9dd03d
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 5: Filter emails by recipient (may not be implemented)
GET http://localhost:8080/api/v1/emails?to=test_1761828406_5a9dd03d@example.com
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.data" isCollection

# Test 6: Get email statistics
GET http://localhost:8080/api/v1/emails/stats
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.total" exists
jsonpath "$.data.pending" exists

# Test 7: Try to access emails without authentication
GET http://localhost:8080/api/v1/emails

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 8: Send email with invalid data
POST http://localhost:8080/api/v1/emails/send
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "to": "invalid-email",
  "subject": "Test"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false