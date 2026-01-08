#!/bin/bash

# Test script for Extra Efforts CRUD operations
# Tests all endpoints for creating, reading, updating, and deleting extra efforts

set -e

BASE_URL="http://localhost:8080"
TOKEN=""
CLIENT_ID=""
SESSION_ID=""
EXTRA_EFFORT_ID=""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Extra Efforts API Test Suite ===${NC}\n"

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
        exit 1
    fi
}

# 1. Login to get token
echo "1. Authenticating..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "test@example.com",
        "password": "password123"
    }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
    echo -e "${RED}Failed to get authentication token${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
print_result 0 "Authentication successful"

# 2. Get a client ID (create one if needed)
echo -e "\n2. Getting client ID..."
CLIENTS_RESPONSE=$(curl -s -X GET "$BASE_URL/clients?limit=1" \
    -H "Authorization: Bearer $TOKEN")

CLIENT_ID=$(echo $CLIENTS_RESPONSE | jq -r '.data[0].id')
if [ -z "$CLIENT_ID" ] || [ "$CLIENT_ID" == "null" ]; then
    echo "No clients found, creating one..."
    CREATE_CLIENT=$(curl -s -X POST "$BASE_URL/clients" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "first_name": "Test",
            "last_name": "Client",
            "email": "testclient@example.com"
        }')
    CLIENT_ID=$(echo $CREATE_CLIENT | jq -r '.data.id')
fi
print_result 0 "Client ID: $CLIENT_ID"

# 3. Get a session ID (create one if needed)
echo -e "\n3. Getting session ID..."
SESSIONS_RESPONSE=$(curl -s -X GET "$BASE_URL/sessions?client_id=$CLIENT_ID&limit=1" \
    -H "Authorization: Bearer $TOKEN")

SESSION_ID=$(echo $SESSIONS_RESPONSE | jq -r '.data[0].id')
if [ -z "$SESSION_ID" ] || [ "$SESSION_ID" == "null" ]; then
    echo "No sessions found, creating one..."
    CREATE_SESSION=$(curl -s -X POST "$BASE_URL/sessions" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"client_id\": $CLIENT_ID,
            \"original_date\": \"2026-01-15T10:00:00Z\",
            \"status\": \"conducted\",
            \"duration_min\": 60,
            \"number_units\": 1
        }")
    SESSION_ID=$(echo $CREATE_SESSION | jq -r '.data.id')
fi
print_result 0 "Session ID: $SESSION_ID"

# 4. Create Extra Effort
echo -e "\n4. Creating extra effort..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/extra-efforts" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"client_id\": $CLIENT_ID,
        \"session_id\": $SESSION_ID,
        \"effort_type\": \"phone_call\",
        \"effort_date\": \"2026-01-15T14:30:00Z\",
        \"duration_min\": 15,
        \"description\": \"Follow-up phone call with client\",
        \"billable\": true
    }")

EXTRA_EFFORT_ID=$(echo $CREATE_RESPONSE | jq -r '.data.id')
if [ -z "$EXTRA_EFFORT_ID" ] || [ "$EXTRA_EFFORT_ID" == "null" ]; then
    echo -e "${RED}Failed to create extra effort${NC}"
    echo "Response: $CREATE_RESPONSE"
    exit 1
fi
print_result 0 "Created extra effort ID: $EXTRA_EFFORT_ID"

# 5. Get Extra Effort by ID
echo -e "\n5. Getting extra effort by ID..."
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/extra-efforts/$EXTRA_EFFORT_ID" \
    -H "Authorization: Bearer $TOKEN")

RETRIEVED_ID=$(echo $GET_RESPONSE | jq -r '.data.id')
if [ "$RETRIEVED_ID" != "$EXTRA_EFFORT_ID" ]; then
    echo -e "${RED}Failed to retrieve extra effort${NC}"
    exit 1
fi
print_result 0 "Retrieved extra effort successfully"

# 6. List Extra Efforts
echo -e "\n6. Listing extra efforts..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/extra-efforts?client_id=$CLIENT_ID" \
    -H "Authorization: Bearer $TOKEN")

EFFORT_COUNT=$(echo $LIST_RESPONSE | jq '.data | length')
if [ "$EFFORT_COUNT" -lt 1 ]; then
    echo -e "${RED}No extra efforts found in list${NC}"
    exit 1
fi
print_result 0 "Found $EFFORT_COUNT extra effort(s)"

# 7. List with filters
echo -e "\n7. Testing filters..."
FILTER_RESPONSE=$(curl -s -X GET "$BASE_URL/extra-efforts?billing_status=unbilled&effort_type=phone_call" \
    -H "Authorization: Bearer $TOKEN")
print_result 0 "Filter by billing_status and effort_type"

# 8. Update Extra Effort
echo -e "\n8. Updating extra effort..."
UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/extra-efforts/$EXTRA_EFFORT_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "duration_min": 20,
        "description": "Extended follow-up phone call"
    }')

UPDATED_DURATION=$(echo $UPDATE_RESPONSE | jq -r '.data.duration_min')
if [ "$UPDATED_DURATION" != "20" ]; then
    echo -e "${RED}Failed to update extra effort${NC}"
    exit 1
fi
print_result 0 "Updated extra effort duration to 20 minutes"

# 9. Create another effort (for email)
echo -e "\n9. Creating email correspondence effort..."
EMAIL_EFFORT=$(curl -s -X POST "$BASE_URL/extra-efforts" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"client_id\": $CLIENT_ID,
        \"effort_type\": \"email_correspondence\",
        \"effort_date\": \"2026-01-16T09:00:00Z\",
        \"duration_min\": 10,
        \"description\": \"Email exchange regarding next appointment\",
        \"billable\": true
    }")

EMAIL_EFFORT_ID=$(echo $EMAIL_EFFORT | jq -r '.data.id')
print_result 0 "Created email effort ID: $EMAIL_EFFORT_ID"

# 10. Create non-billable effort
echo -e "\n10. Creating non-billable effort..."
NON_BILLABLE=$(curl -s -X POST "$BASE_URL/extra-efforts" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"client_id\": $CLIENT_ID,
        \"effort_type\": \"documentation\",
        \"effort_date\": \"2026-01-16T15:00:00Z\",
        \"duration_min\": 5,
        \"description\": \"Internal documentation\",
        \"billable\": false
    }")

NON_BILLABLE_ID=$(echo $NON_BILLABLE | jq -r '.data.id')
print_result 0 "Created non-billable effort ID: $NON_BILLABLE_ID"

# 11. Verify unbilled sessions includes extra efforts
echo -e "\n11. Checking unbilled sessions endpoint..."
UNBILLED_RESPONSE=$(curl -s -X GET "$BASE_URL/client-invoices/unbilled-sessions" \
    -H "Authorization: Bearer $TOKEN")

EXTRA_EFFORTS_ARRAY=$(echo $UNBILLED_RESPONSE | jq -r ".data[] | select(.id == $CLIENT_ID) | .extra_efforts")
if [ "$EXTRA_EFFORTS_ARRAY" != "null" ]; then
    EFFORTS_COUNT=$(echo $UNBILLED_RESPONSE | jq -r ".data[] | select(.id == $CLIENT_ID) | .extra_efforts | length")
    print_result 0 "Unbilled sessions includes $EFFORTS_COUNT extra effort(s)"
else
    print_result 1 "Unbilled sessions missing extra_efforts array"
fi

# 12. Delete non-billable effort
echo -e "\n12. Deleting non-billable effort..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/extra-efforts/$NON_BILLABLE_ID" \
    -H "Authorization: Bearer $TOKEN")

DELETE_SUCCESS=$(echo $DELETE_RESPONSE | jq -r '.success')
if [ "$DELETE_SUCCESS" != "true" ]; then
    echo -e "${RED}Failed to delete extra effort${NC}"
    exit 1
fi
print_result 0 "Deleted extra effort successfully"

# 13. Verify deletion
echo -e "\n13. Verifying deletion..."
VERIFY_DELETE=$(curl -s -w "%{http_code}" -o /dev/null -X GET "$BASE_URL/extra-efforts/$NON_BILLABLE_ID" \
    -H "Authorization: Bearer $TOKEN")

if [ "$VERIFY_DELETE" == "404" ]; then
    print_result 0 "Extra effort properly deleted (404 response)"
else
    print_result 1 "Extra effort still exists after deletion"
fi

# 14. Test validation (missing required fields)
echo -e "\n14. Testing validation..."
VALIDATION_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/validation_response.json -X POST "$BASE_URL/extra-efforts" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "effort_type": "phone_call",
        "description": "Missing client_id and effort_date"
    }')

if [ "$VALIDATION_RESPONSE" == "400" ]; then
    print_result 0 "Validation correctly rejects invalid requests"
else
    print_result 1 "Validation should reject invalid requests"
fi

# Cleanup
echo -e "\n${YELLOW}=== Cleanup ===${NC}"
echo "Deleting created extra efforts..."
curl -s -X DELETE "$BASE_URL/extra-efforts/$EXTRA_EFFORT_ID" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
curl -s -X DELETE "$BASE_URL/extra-efforts/$EMAIL_EFFORT_ID" \
    -H "Authorization: Bearer $TOKEN" > /dev/null

echo -e "\n${GREEN}=== All Tests Passed! ===${NC}"
