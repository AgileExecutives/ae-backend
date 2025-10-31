# Health Check API Tests
# Tests system health and database connectivity

# Test 1: Basic health check
GET http://localhost:8080/api/v1/health

HTTP 200
[Asserts]
jsonpath "$.status" == "healthy"
jsonpath "$.timestamp" exists
jsonpath "$.version" exists
jsonpath "$.database" == "connected"

# Test 2: Health check response format validation
GET http://localhost:8080/api/v1/health

HTTP 200
[Asserts]
jsonpath "$.status" isString
jsonpath "$.timestamp" matches "^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}"
jsonpath "$.version" isString
jsonpath "$.database" isString

# Test 3: Health endpoint should not require authentication
GET http://localhost:8080/api/v1/health
Authorization: Bearer invalid_token

HTTP 200
[Asserts]
jsonpath "$.status" == "healthy"