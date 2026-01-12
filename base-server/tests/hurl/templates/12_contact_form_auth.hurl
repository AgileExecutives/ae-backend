# Contact Form Authentication Tests
# Advanced authentication scenarios for contact form and newsletter management
# This test focuses on security and authentication edge cases

# Test 1: Verify contact form is publicly accessible (no auth required)
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Public Access Test",
  "email": "public_access@example.com",
  "subject": "Public Contact Form Test",
  "message": "This should work without authentication",
  "newsletter": false,
  "source": "website"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"

# Test 2: Newsletter management requires authentication - test GET endpoint
GET {{HOST}}/api/v1/contact/newsletter

HTTP 401
[Asserts]
jsonpath "$.error" contains "Authorization"

# Test 3: Newsletter unsubscribe requires authentication
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe?email=test@example.com

HTTP 401
[Asserts]
jsonpath "$.error" contains "Authorization"

# Setup authentication for protected tests
# Use the same pattern as the existing working auth tests
POST {{HOST}}/api/v1/auth/register
Content-Type: application/json
{
  "username": "newsletter_test_user",
  "email": "newsletter_test@example.com",
  "password": "TestPass123!",
  "first_name": "Newsletter",
  "last_name": "Tester",
  "company_name": "Newsletter Test Corp",
  "tenant_name": "newsletter_test",
  "accept_terms": true
}

# Accept either created or already exists
HTTP *
[Asserts]
header "Content-Type" contains "application/json"

# Login to get auth token (using email field as required)
POST {{HOST}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "newsletter_test@example.com",
  "password": "TestPass123!"
}

HTTP 200
[Captures]
contact_auth_token: jsonpath "$.data.token"

# Test 4: Newsletter management with valid authentication
GET {{HOST}}/api/v1/contact/newsletter
Authorization: Bearer {{contact_auth_token}}

HTTP 200
[Asserts]
jsonpath "$" isCollection

# Test 5: Contact form with newsletter subscription to create test data
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Auth Newsletter Test",
  "email": "auth_newsletter_test@example.com",
  "subject": "Newsletter Auth Test",
  "message": "Creating newsletter subscription for auth testing",
  "newsletter": true,
  "source": "website"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"

# Test 6: Newsletter unsubscribe with valid authentication
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe?email=auth_newsletter_test@example.com
Authorization: Bearer {{contact_auth_token}}

HTTP 200
[Asserts]
jsonpath "$.message" contains "unsubscribed"

# Test 7: Newsletter unsubscribe with invalid token
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe?email=test@example.com
Authorization: Bearer invalid_token_12345

HTTP 401
[Asserts]
jsonpath "$.error" contains "Authorization"

# Test 8: Newsletter GET with malformed authorization header
GET {{HOST}}/api/v1/contact/newsletter
Authorization: InvalidFormat

HTTP 401
[Asserts]
jsonpath "$.error" contains "Authorization"

# Test 9: Test that contact form still works after auth tests
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Final Auth Test",
  "email": "final_auth_test@example.com",
  "subject": "Final Test After Auth",
  "message": "Ensuring contact form still works after authentication tests",
  "newsletter": false,
  "source": "website"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"