# Contact Form and Newsletter Tests
# Test the public contact form and newsletter management functionality
# This includes comprehensive testing of contact form submission, newsletter subscription,
# and newsletter management with proper authentication

# Test 1: Basic contact form submission (no auth required, no newsletter)
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Test User",
  "email": "test@example.com",
  "subject": "Test Contact Form",
  "message": "This is a test message from the contact form.",
  "newsletter": false,
  "source": "website",
  "timestamp": "2025-10-12T10:00:00Z"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"

# Test 2: Contact form with newsletter subscription enabled
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Newsletter Subscriber",
  "email": "newsletter@example.com", 
  "subject": "Interest in Services",
  "message": "I'm interested in your services and would like to subscribe to the newsletter.",
  "newsletter": true,
  "source": "website", 
  "timestamp": "2025-10-12T10:05:00Z"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"
# Note: Newsletter functionality may show partial success in response

# Test 3: Contact form with different newsletter subscriber
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Another Newsletter User",
  "email": "newsletter2@example.com",
  "subject": "Newsletter Signup", 
  "message": "Please add me to your newsletter.",
  "newsletter": true,
  "source": "website",
  "timestamp": "2025-10-12T10:10:00Z"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"

# Test contact form with missing required fields
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Incomplete User",
  "email": "incomplete@example.com"
}

HTTP 400
[Asserts]
jsonpath "$.error" exists

# Test contact form with invalid email
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Invalid Email User",
  "email": "not-an-email",
  "subject": "Test",
  "message": "Test message",
  "source": "website"
}

HTTP 400
[Asserts]
jsonpath "$.error" exists

# Test 6: Contact form with empty required fields (edge case)
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "",
  "email": "",
  "subject": "",
  "message": ""
}

HTTP 400
[Asserts]
jsonpath "$.error" exists

# Setup auth for newsletter management tests
# Try to register a test user (may fail if user already exists)
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

# Accept either 201 (created) or 409 (already exists)
HTTP *
[Asserts]
header "Content-Type" contains "application/json"

# Login with the test user (should work whether user was just created or already existed)
POST {{HOST}}/api/v1/auth/login
Content-Type: application/json
{
  "email": "newsletter_test@example.com",
  "password": "TestPass123!"
}

HTTP 200
[Captures]
auth_token: jsonpath "$.data.token"

# Newsletter Management Tests (require authentication)

# Test 7: Getting newsletter subscriptions (requires auth)
GET {{HOST}}/api/v1/contact/newsletter
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$" isCollection

# Test 8: Verify newsletter subscriptions are accessible without auth failure
GET {{HOST}}/api/v1/contact/newsletter

HTTP 401
[Asserts]
jsonpath "$.message" contains "Authorization required"

# Test 9: Newsletter unsubscribe with valid email
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe?email=newsletter@example.com
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.message" contains "unsubscribed"

# Test 10: Newsletter unsubscribe with another valid email
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe?email=newsletter2@example.com
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.message" contains "unsubscribed"

# Test 11: Unsubscribing non-existent email (error handling)
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe?email=nonexistent@example.com
Authorization: Bearer {{auth_token}}

HTTP 404
[Asserts]
jsonpath "$.error" contains "not found"

# Test 12: Unsubscribe without email parameter (validation error)
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe
Authorization: Bearer {{auth_token}}

HTTP 400
[Asserts]
jsonpath "$.error" contains "Email parameter is required"

# Test 13: Unsubscribe without authentication (security test)
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe?email=test@example.com

HTTP 401
[Asserts]
jsonpath "$.message" contains "Authorization required"

# Additional Authentication Tests (adapted from backend contact-form-auth)

# Test 14: Contact form with invalid email format (validation test)
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Invalid Email User",
  "email": "not-an-email",
  "subject": "Invalid Email Test",
  "message": "This should fail with invalid email format",
  "newsletter": false,
  "source": "website"
}

HTTP 400
[Asserts]
jsonpath "$.error" exists

# Test 15: Contact form with missing required fields (validation test)
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Incomplete User",
  "newsletter": false,
  "source": "website"
}

HTTP 400
[Asserts]
jsonpath "$.error" exists

# Test 16: Contact form with newsletter subscription for same email (update scenario)
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Newsletter Updated User",
  "email": "newsletter@example.com",
  "subject": "Updated Interest",
  "message": "This should update the existing newsletter subscription",
  "newsletter": true,
  "source": "website"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"

# Test 17: Multiple newsletter subscriptions with Bearer auth testing
POST {{HOST}}/api/v1/contact/form
Content-Type: application/json
{
  "name": "Auth Test User",
  "email": "auth_test_newsletter@example.com",
  "subject": "Newsletter with Auth Context",
  "message": "Testing newsletter functionality in authenticated context",
  "newsletter": true,
  "source": "website"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"

# Test 18: Verify the newly created newsletter subscription exists
GET {{HOST}}/api/v1/contact/newsletter
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$" isCollection
# Verify that our test email is in the newsletter list
jsonpath "$[*].email" includes "auth_test_newsletter@example.com"

# Test 19: Unsubscribe the test email we just created
DELETE {{HOST}}/api/v1/contact/newsletter/unsubscribe?email=auth_test_newsletter@example.com
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.message" contains "unsubscribed"