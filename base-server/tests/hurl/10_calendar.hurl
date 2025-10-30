# 10_calendar.hurl - Calendar Module API Tests
# Testing Calendar and RecurringEvent CRUD operations

# Test variables from auth flow
# Assuming JWT token is available from previous test or can be obtained

# First, let's login to get a JWT token
POST {{base_url}}/api/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "testpassword123"
}

HTTP/1.1 200
[Captures]
jwt_token: jsonpath "$.token"

# ==============================
# CALENDAR CRUD TESTS
# ==============================

# 1. Create a new calendar
POST {{base_url}}/api/calendars
Authorization: Bearer {{jwt_token}}
Content-Type: application/json

{
  "name": "My Work Calendar",
  "schedule_days_ahead": 30,
  "schedule_days_notice": 7,
  "schedule_unit": "days"
}

HTTP/1.1 201
[Captures]
calendar_id: jsonpath "$.id"
[Asserts]
jsonpath "$.name" == "My Work Calendar"
jsonpath "$.schedule_days_ahead" == 30
jsonpath "$.schedule_days_notice" == 7
jsonpath "$.schedule_unit" == "days"

# 2. Get all calendars
GET {{base_url}}/api/calendars
Authorization: Bearer {{jwt_token}}

HTTP/1.1 200
[Asserts]
jsonpath "$" isArray
jsonpath "$[0].name" exists
jsonpath "$[0].id" exists

# 3. Get specific calendar
GET {{base_url}}/api/calendars/{{calendar_id}}
Authorization: Bearer {{jwt_token}}

HTTP/1.1 200
[Asserts]
jsonpath "$.id" == {{calendar_id}}
jsonpath "$.name" == "My Work Calendar"

# 4. Update calendar
PUT {{base_url}}/api/calendars/{{calendar_id}}
Authorization: Bearer {{jwt_token}}
Content-Type: application/json

{
  "name": "Updated Work Calendar",
  "schedule_days_ahead": 45,
  "schedule_days_notice": 14,
  "schedule_unit": "days"
}

HTTP/1.1 200
[Asserts]
jsonpath "$.name" == "Updated Work Calendar"
jsonpath "$.schedule_days_ahead" == 45
jsonpath "$.schedule_days_notice" == 14

# ==============================
# RECURRING EVENT CRUD TESTS
# ==============================

# 5. Create a recurring event
POST {{base_url}}/api/recurring-events
Authorization: Bearer {{jwt_token}}
Content-Type: application/json

{
  "calendar_id": {{calendar_id}},
  "title": "Weekly Team Meeting",
  "participants": ["alice@company.com", "bob@company.com"],
  "weekday": "monday",
  "interval": "weekly",
  "start_time": "09:00",
  "end_time": "10:00",
  "description": "Regular team sync meeting",
  "location": "Conference Room A"
}

HTTP/1.1 201
[Captures]
recurring_event_id: jsonpath "$.id"
[Asserts]
jsonpath "$.title" == "Weekly Team Meeting"
jsonpath "$.weekday" == "monday"
jsonpath "$.interval" == "weekly"
jsonpath "$.start_time" == "09:00"
jsonpath "$.end_time" == "10:00"
jsonpath "$.calendar_id" == {{calendar_id}}

# 6. Get all recurring events
GET {{base_url}}/api/recurring-events
Authorization: Bearer {{jwt_token}}

HTTP/1.1 200
[Asserts]
jsonpath "$" isArray
jsonpath "$[0].title" exists
jsonpath "$[0].id" exists

# 7. Get specific recurring event with calendar preloaded
GET {{base_url}}/api/recurring-events/{{recurring_event_id}}
Authorization: Bearer {{jwt_token}}

HTTP/1.1 200
[Asserts]
jsonpath "$.id" == {{recurring_event_id}}
jsonpath "$.title" == "Weekly Team Meeting"
jsonpath "$.calendar.name" == "Updated Work Calendar"

# 8. Update recurring event
PUT {{base_url}}/api/recurring-events/{{recurring_event_id}}
Authorization: Bearer {{jwt_token}}
Content-Type: application/json

{
  "calendar_id": {{calendar_id}},
  "title": "Updated Team Meeting",
  "participants": ["alice@company.com", "bob@company.com", "charlie@company.com"],
  "weekday": "tuesday",
  "interval": "bi-weekly",
  "start_time": "10:00",
  "end_time": "11:00",
  "description": "Updated team sync meeting",
  "location": "Conference Room B"
}

HTTP/1.1 200
[Asserts]
jsonpath "$.title" == "Updated Team Meeting"
jsonpath "$.weekday" == "tuesday"
jsonpath "$.interval" == "bi-weekly"
jsonpath "$.start_time" == "10:00"
jsonpath "$.end_time" == "11:00"

# ==============================
# CLEANUP TESTS
# ==============================

# 9. Delete recurring event
DELETE {{base_url}}/api/recurring-events/{{recurring_event_id}}
Authorization: Bearer {{jwt_token}}

HTTP/1.1 200
[Asserts]
jsonpath "$.message" exists

# 10. Verify recurring event is deleted
GET {{base_url}}/api/recurring-events/{{recurring_event_id}}
Authorization: Bearer {{jwt_token}}

HTTP/1.1 404
[Asserts]
jsonpath "$.error" exists

# 11. Delete calendar
DELETE {{base_url}}/api/calendars/{{calendar_id}}
Authorization: Bearer {{jwt_token}}

HTTP/1.1 200
[Asserts]
jsonpath "$.message" exists

# 12. Verify calendar is deleted
GET {{base_url}}/api/calendars/{{calendar_id}}
Authorization: Bearer {{jwt_token}}

HTTP/1.1 404
[Asserts]
jsonpath "$.error" exists

# ==============================
# ERROR HANDLING TESTS
# ==============================

# 13. Try to create recurring event without calendar_id
POST {{base_url}}/api/recurring-events
Authorization: Bearer {{jwt_token}}
Content-Type: application/json

{
  "title": "Invalid Meeting",
  "weekday": "monday",
  "interval": "weekly",
  "start_time": "09:00",
  "end_time": "10:00"
}

HTTP/1.1 400
[Asserts]
jsonpath "$.error" exists

# 14. Try to access non-existent calendar
GET {{base_url}}/api/calendars/99999
Authorization: Bearer {{jwt_token}}

HTTP/1.1 404
[Asserts]
jsonpath "$.error" exists

# 15. Try to access without authentication
GET {{base_url}}/api/calendars

HTTP/1.1 401
[Asserts]
jsonpath "$.error" exists