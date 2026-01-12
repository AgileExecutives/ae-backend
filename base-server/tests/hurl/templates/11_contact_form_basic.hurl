# Basic Contact Form Tests (No Newsletter Management)
# Test the public contact form functionality that we know is working

# Test 1: Basic contact form submission (no newsletter)
POST {{HOST}}/api/v1/contacts/form
Content-Type: application/json
{
  "name": "Test User Basic",
  "email": "test-basic@example.com",
  "subject": "Test Contact Form Basic",
  "message": "This is a basic test message from the contact form.",
  "newsletter": false,
  "source": "website",
  "timestamp": "2025-10-12T10:00:00Z"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"

# Test 2: Contact form with newsletter subscription (may show partial success)
POST {{HOST}}/api/v1/contacts/form
Content-Type: application/json
{
  "name": "Newsletter Test User",
  "email": "newsletter-test@example.com", 
  "subject": "Newsletter Interest",
  "message": "I would like to subscribe to the newsletter.",
  "newsletter": true,
  "source": "website", 
  "timestamp": "2025-10-12T10:05:00Z"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Contact form submitted successfully"

# Test 3: Contact form validation - missing required fields
POST {{HOST}}/api/v1/contacts/form
Content-Type: application/json
{
  "name": "Incomplete User",
  "email": "incomplete@example.com"
}

HTTP 400
[Asserts]
jsonpath "$.error" exists

# Test 4: Contact form validation - invalid email format
POST {{HOST}}/api/v1/contacts/form
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

# Test 5: Contact form validation - empty fields
POST {{HOST}}/api/v1/contacts/form
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

# Test 6: Verify newsletter endpoints require authentication (security test)
GET {{HOST}}/api/v1/contacts/newsletter

HTTP 401
[Asserts]
jsonpath "$.message" contains "Authorization required"

# Test 7: Verify unsubscribe endpoint requires authentication (security test)  
DELETE {{HOST}}/api/v1/contacts/newsletter/unsubscribe?email=test@example.com

HTTP 401
[Asserts]
jsonpath "$.message" contains "Authorization required"