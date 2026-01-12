# Email Tests
# Test email sending and management endpoints

# Setup fresh auth session
POST {{host}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "testuser",
  "password": "newpass123"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Test sending an email
POST {{host}}/api/v1/emails/send
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "to_email": "recipient@example.com",
  "from_email": "sender@example.com",
  "subject": "Test Email Subject",
  "body": "This is a test email body in plain text.",
  "html_body": "<h1>Test Email</h1><p>This is a test email body in <strong>HTML</strong>.</p>",
  "metadata": {
    "campaign": "test_campaign",
    "source": "api_test"
  }
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "Email queued successfully"
jsonpath "$.data.to_email" == "recipient@example.com"
jsonpath "$.data.from_email" == "sender@example.com"
jsonpath "$.data.subject" == "Test Email Subject"
jsonpath "$.data.status" == "pending"

[Captures]
email_id: jsonpath "$.data.id"

# Test sending email with minimal required fields
POST {{host}}/api/v1/emails/send
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "to_email": "minimal@example.com",
  "subject": "Minimal Email",
  "body": "Minimal email body"
}

HTTP 201
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.to_email" == "minimal@example.com"
jsonpath "$.data.subject" == "Minimal Email"
jsonpath "$.data.status" == "pending"

[Captures]
minimal_email_id: jsonpath "$.data.id"

# Test sending email with invalid data
POST {{host}}/api/v1/emails/send
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "to_email": "invalid-email",
  "subject": "",
  "body": ""
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "validation"

# Test sending email without required fields
POST {{host}}/api/v1/emails/send
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "subject": "Missing recipient"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false

# Test getting all emails
GET {{host}}/api/v1/emails
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
jsonpath "$.pagination.page" == 1
jsonpath "$.pagination.limit" == 20

# Test getting emails with filters
GET {{host}}/api/v1/emails?status=pending&limit=5
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection
jsonpath "$.pagination.limit" == 5

# Test getting emails with date range filter
GET {{host}}/api/v1/emails?from_date=2025-01-01&to_date=2025-12-31
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection

# Test searching emails by recipient
GET {{host}}/api/v1/emails?search=recipient@example.com
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data" isCollection

# Test getting specific email
GET {{host}}/api/v1/emails/{{email_id}}
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.id" == {{email_id}}
jsonpath "$.data.to_email" == "recipient@example.com"
jsonpath "$.data.subject" == "Test Email Subject"
jsonpath "$.data.status" == "pending"

# Test getting non-existent email
GET {{host}}/api/v1/emails/999
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Email not found"

# Test getting email statistics
GET {{host}}/api/v1/emails/stats
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.total_emails" isNumber
jsonpath "$.data.pending_emails" isNumber
jsonpath "$.data.sent_emails" isNumber
jsonpath "$.data.failed_emails" isNumber

# Test getting email statistics with date filter
GET {{host}}/api/v1/emails/stats?from_date=2025-01-01&to_date=2025-12-31
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.total_emails" isNumber

# Test accessing email endpoints without authentication
GET {{host}}/api/v1/emails

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"

# Test sending email without authentication
POST {{host}}/api/v1/emails/send
Content-Type: application/json
{
  "to_email": "test@example.com",
  "subject": "Unauthorized email",
  "body": "This should fail"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"

# Test email statistics without authentication
GET {{host}}/api/v1/emails/stats

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"

# Test bulk email sending (if endpoint exists)
POST {{host}}/api/v1/emails/send
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "to_email": "bulk1@example.com,bulk2@example.com,bulk3@example.com",
  "subject": "Bulk Email Test",
  "body": "This is a bulk email test",
  "metadata": {
    "type": "bulk",
    "campaign": "test_bulk"
  }
}

HTTP 201
[Asserts]
jsonpath "$.success" == true