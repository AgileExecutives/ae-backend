# Health Check Tests
# Test basic API availability and health endpoints

# Test basic health endpoint
GET {{host}}/api/v1/health

HTTP 200
[Asserts]
jsonpath "$.status" == "healthy"
jsonpath "$.database" == "connected"
jsonpath "$.version" == "1.0.0"
jsonpath "$.timestamp" exists

# Test ping endpoint
GET {{host}}/api/v1/ping

HTTP 200
[Asserts]
jsonpath "$.message" == "pong"