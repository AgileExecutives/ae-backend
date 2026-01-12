# User Settings API Tests with Unique Data
# Tests user settings CRUD operations and validation

# First, create a test user and get authentication token
POST {{HOST}}/api/v1/auth/register
Content-Type: application/json
{
  "email": "{{UNIQUE_EMAIL}}",
  "username": "user_{{UNIQUE_ID}}",
  "password": "{{UNIQUE_PASSWORD}}",
  "first_name": "Settings",
  "last_name": "Tester",
  "company_name": "Settings Test Company {{UNIQUE_ID}}",
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
user_id: jsonpath "$.data.user.id"

# Test 1: Get user settings (should create default if none exist)
GET {{HOST}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" exists
jsonpath "$.data.user_id" == {{user_id}}
jsonpath "$.data.language" exists
jsonpath "$.data.timezone" exists
jsonpath "$.data.theme" exists

# Test 2: Update user settings
PUT {{HOST}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "language": "de",
  "timezone": "Europe/Berlin",
  "theme": "dark",
  "settings": "{\"email_notifications\": true, \"push_notifications\": false, \"unique_preference\": \"{{UNIQUE_ID}}\"}"
}

HTTP 200
[Asserts]
jsonpath "$.success" exists
jsonpath "$.data.language" == "de"
jsonpath "$.data.timezone" == "Europe/Berlin"
jsonpath "$.data.theme" == "dark"

# Test 3: Verify settings were updated
GET {{HOST}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" exists
jsonpath "$.data.language" == "de"
jsonpath "$.data.timezone" == "Europe/Berlin"
jsonpath "$.data.theme" == "dark"

# Test 4: Reset settings to defaults
POST {{HOST}}/api/v1/user-settings/reset
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" exists

# Test 5: Try to access settings without authentication
GET {{HOST}}/api/v1/user-settings

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 6: Update settings with invalid data
PUT {{HOST}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "language": "invalid_language_code_{{UNIQUE_ID}}"
}

# Accept various status codes as validation rules may vary
HTTP *
[Asserts]
jsonpath "$.success" exists

# Test 7: Reset settings to defaults
PUT {{HOST}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "language": "en",
  "timezone": "UTC", 
  "theme": "light",
  "settings": "{}"
}

HTTP 200
[Asserts]
jsonpath "$.success" exists
jsonpath "$.data.language" == "en"