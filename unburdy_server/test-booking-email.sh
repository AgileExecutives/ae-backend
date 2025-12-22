#!/bin/bash

# Test booking with token to trigger email confirmation

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing Booking Email Confirmation${NC}"
echo "======================================"

# First, we need to create a booking token
# You'll need to replace TOKEN_HERE with an actual token from your database
# or create one through your booking link creation endpoint

# Example booking request (adjust the token and dates as needed)
TOKEN="test-token-123"  # Replace with actual token

echo -e "\n${YELLOW}Attempting to book session with token: ${TOKEN}${NC}"

curl -X POST "http://localhost:8080/api/v1/sessions/book/${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "dates": [
      {
        "date": "2025-12-23",
        "start_time": "10:00",
        "end_time": "11:00"
      }
    ],
    "client_first_name": "John",
    "client_last_name": "Doe",
    "client_email": "john.doe@example.com",
    "client_phone": "+49123456789"
  }' \
  -v

echo -e "\n\n${YELLOW}Check the server console output above for:${NC}"
echo "1. Mock email output (should print the email to console)"
echo "2. Email subject: 'Terminbestätigung - [Title]'"
echo "3. Email body with appointment details"
echo ""
echo -e "${GREEN}If MOCK_EMAIL=true, the email should be printed to the console${NC}"
