#!/bin/bash

# Test script for the minimal server with ping module

echo "=== Testing Minimal Server with Ping Module ==="
echo ""

# First, let's test if we can build the server
echo "1. Building the server..."
if make build; then
    echo "✅ Server built successfully"
else
    echo "❌ Server build failed"
    exit 1
fi

echo ""
echo "2. Testing ping endpoints (assuming server is running on port 8080)..."
echo ""

# Test basic ping endpoint
echo "Testing basic ping endpoint:"
echo "curl -X GET http://localhost:8080/api/v1/ping/ping"
echo ""

# Test protected ping endpoint
echo "Testing protected ping endpoint (will require auth):"
echo "curl -X GET http://localhost:8080/api/v1/ping/protected-ping"
echo ""

echo "=== Manual Testing Instructions ==="
echo ""
echo "1. Start the server:"
echo "   make run"
echo "   # or"
echo "   DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=password DB_NAME=ae_minimal DB_SSLMODE=disable ./bin/minimal-server"
echo ""
echo "2. Test the ping endpoint:"
echo "   curl http://localhost:8080/api/v1/ping/ping"
echo ""
echo "3. Register a user first:"
echo "   curl -X POST http://localhost:8080/api/v1/auth/register \\"
echo "     -H \"Content-Type: application/json\" \\"
echo "     -d '{\"email\": \"test@example.com\", \"password\": \"password123\", \"tenant_name\": \"My Company\"}'"
echo ""
echo "4. Login to get a token:"
echo "   curl -X POST http://localhost:8080/api/v1/auth/login \\"
echo "     -H \"Content-Type: application/json\" \\"
echo "     -d '{\"email\": \"test@example.com\", \"password\": \"password123\"}'"
echo ""
echo "5. Test protected ping with the token:"
echo "   curl -X GET http://localhost:8080/api/v1/ping/protected-ping \\"
echo "     -H \"Authorization: Bearer YOUR_TOKEN_HERE\""
echo ""
echo "Expected responses:"
echo "- ping endpoint: {\"message\": \"pong\", \"module\": \"ping\", \"version\": \"1.0.0\", ...}"
echo "- protected-ping: {\"message\": \"authenticated pong\", \"module\": \"ping\", \"user_authenticated\": true, ...}"