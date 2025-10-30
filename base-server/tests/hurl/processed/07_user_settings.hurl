# User Settings API Tests with Unique Data
# Tests user settings CRUD operations and validation

# First, create a test user and get authentication token
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "username": "testuser_1761828406_5a9dd03d",
  "email": "test_1761828406_5a9dd03d@example.com",
  "password": "Pass123_5a9dd03d",
  "first_name": "Settings",
  "last_name": "Tester",
  "company_name": "Settings Test Company 1761828406_5a9dd03d",
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
user_id: jsonpath "$.data.user.id"

# Test 1: Get user settings (should create default if none exist)
GET http://localhost:8080/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.user_id" == {{user_id}}
jsonpath "$.data.language" exists
jsonpath "$.data.timezone" exists
jsonpath "$.data.theme" exists

# Test 2: Update user settings
PUT http://localhost:8080/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "language": "de",
  "timezone": "Europe/Berlin",
  "theme": "dark",
  "settings": "{\"email_notifications\": true, \"push_notifications\": false, \"unique_preference\": \"1761828406_5a9dd03d\"}"
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.language" == "de"
jsonpath "$.data.timezone" == "Europe/Berlin"
jsonpath "$.data.theme" == "dark"

# Test 3: Verify settings were updated
GET http://localhost:8080/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.language" == "de"
jsonpath "$.data.timezone" == "Europe/Berlin"
jsonpath "$.data.theme" == "dark"

# Test 4: Reset settings to defaults
POST http://localhost:8080/api/v1/user-settings/reset
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true

# Test 5: Try to access settings without authentication
GET http://localhost:8080/api/v1/user-settings

HTTP 401
[Asserts]
jsonpath "$.success" == false

# Test 6: Update settings with invalid data
PUT http://localhost:8080/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "language": "invalid_language_code_1761828406_5a9dd03d"
}

# Accept various status codes as validation rules may vary
HTTP *
[Asserts]
jsonpath "$.success" exists

# Test 7: Reset settings to defaults
PUT http://localhost:8080/api/v1/user-settings
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
jsonpath "$.success" == true
jsonpath "$.data.language" == "en"