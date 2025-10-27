# User Settings Tests
# Test user settings management endpoints

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
user_id: jsonpath "$.data.user.id"

# Test getting user settings (should create default if none exist)
GET {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.user_id" == {{user_id}}
jsonpath "$.data.language" == "en"
jsonpath "$.data.timezone" == "UTC"
jsonpath "$.data.theme" == "light"
jsonpath "$.data.settings" == {}

# Test updating user settings
PUT {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "language": "de",
  "timezone": "Europe/Berlin",
  "theme": "dark",
  "settings": {
    "notifications": {
      "email": true,
      "push": false,
      "sms": false
    },
    "display": {
      "compact_mode": true,
      "show_avatars": true
    },
    "preferences": {
      "auto_save": true,
      "default_view": "grid"
    }
  }
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "User settings updated successfully"
jsonpath "$.data.language" == "de"
jsonpath "$.data.timezone" == "Europe/Berlin"
jsonpath "$.data.theme" == "dark"
jsonpath "$.data.settings.notifications.email" == true
jsonpath "$.data.settings.display.compact_mode" == true
jsonpath "$.data.settings.preferences.auto_save" == true

# Test partial update of user settings
PUT {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "language": "fr",
  "settings": {
    "notifications": {
      "email": false
    }
  }
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.language" == "fr"
# Theme should remain unchanged
jsonpath "$.data.theme" == "dark"
# Settings should be merged/updated
jsonpath "$.data.settings.notifications.email" == false

# Test getting updated settings to verify persistence
GET {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.language" == "fr"
jsonpath "$.data.timezone" == "Europe/Berlin"
jsonpath "$.data.theme" == "dark"

# Test updating with invalid timezone
PUT {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "timezone": "Invalid/Timezone"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "timezone"

# Test updating with invalid language code
PUT {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "language": "invalid_lang"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "language"

# Test updating with invalid theme
PUT {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "theme": "invalid_theme"
}

HTTP 400
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" contains "theme"

# Test resetting user settings to defaults
POST {{host}}/api/v1/user-settings/reset
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.message" == "User settings reset to defaults"
jsonpath "$.data.language" == "en"
jsonpath "$.data.timezone" == "UTC"
jsonpath "$.data.theme" == "light"
jsonpath "$.data.settings" == {}

# Test getting settings after reset to verify
GET {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.language" == "en"
jsonpath "$.data.timezone" == "UTC"
jsonpath "$.data.theme" == "light"

# Test complex settings structure
PUT {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}
Content-Type: application/json
{
  "settings": {
    "dashboard": {
      "widgets": ["revenue", "customers", "plans"],
      "refresh_interval": 30,
      "show_welcome": false
    },
    "api": {
      "rate_limit_notifications": true,
      "webhook_url": "https://example.com/webhook"
    },
    "security": {
      "two_factor": false,
      "session_timeout": 3600,
      "ip_whitelist": ["192.168.1.1", "10.0.0.1"]
    }
  }
}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.settings.dashboard.widgets[0]" == "revenue"
jsonpath "$.data.settings.dashboard.refresh_interval" == 30
jsonpath "$.data.settings.api.rate_limit_notifications" == true
jsonpath "$.data.settings.security.two_factor" == false

# Test settings isolation between users
POST {{host}}/api/v1/auth/register
Content-Type: application/json
{
  "username": "settingsuser2",
  "email": "settings2@example.com",
  "password": "settings123",
  "first_name": "Settings",
  "last_name": "User2",
  "organization_name": "Settings Org 2"
}

HTTP 201
[Captures]
other_user_token: jsonpath "$.data.token"

# Get settings for new user (should be defaults)
GET {{host}}/api/v1/user-settings
Authorization: Bearer {{other_user_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.language" == "en"
jsonpath "$.data.timezone" == "UTC"
jsonpath "$.data.theme" == "light"
jsonpath "$.data.settings" == {}

# Verify first user's settings are unchanged
GET {{host}}/api/v1/user-settings
Authorization: Bearer {{auth_token}}

HTTP 200
[Asserts]
jsonpath "$.success" == true
jsonpath "$.data.settings.dashboard.widgets[0]" == "revenue"

# Test accessing user settings without authentication
GET {{host}}/api/v1/user-settings

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"

# Test updating user settings without authentication
PUT {{host}}/api/v1/user-settings
Content-Type: application/json
{
  "language": "es"
}

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"

# Test resetting user settings without authentication
POST {{host}}/api/v1/user-settings/reset

HTTP 401
[Asserts]
jsonpath "$.success" == false
jsonpath "$.error" == "Authorization header required"