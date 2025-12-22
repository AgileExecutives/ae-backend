#!/bin/bash

# Complete test for booking confirmation email
# This script will create a booking link, then use it to book appointments

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}=== Testing Booking Confirmation Email ===${NC}\n"

# Step 1: Login to get auth token
echo -e "${YELLOW}Step 1: Getting auth token...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser@unburdy.de",
    "password": "newpass123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  # Try alternate extraction path
  TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"data":{"token":"[^"]*"' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
fi

if [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Failed to get auth token${NC}"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✅ Auth token obtained${NC}\n"

# Step 2: Get calendar ID (we need this for booking link)
echo -e "${YELLOW}Step 2: Getting calendar ID...${NC}"
CALENDARS=$(curl -s -X GET "${BASE_URL}/api/v1/calendars" \
  -H "Authorization: Bearer $TOKEN")

CALENDAR_ID=$(echo $CALENDARS | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$CALENDAR_ID" ]; then
  echo -e "${RED}❌ No calendar found${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Calendar ID: $CALENDAR_ID${NC}\n"

# Step 3: Create a booking link
echo -e "${YELLOW}Step 3: Creating booking link...${NC}"
LINK_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/booking/links" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "calendar_id": '$CALENDAR_ID',
    "title": "Test Booking Link for Email",
    "description": "Testing email confirmation",
    "duration_minutes": 60,
    "max_bookings": 1,
    "expires_at": "2025-12-31T23:59:59Z"
  }')

BOOKING_TOKEN=$(echo $LINK_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$BOOKING_TOKEN" ]; then
  echo -e "${RED}❌ Failed to create booking link${NC}"
  echo "Response: $LINK_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✅ Booking link created with token: $BOOKING_TOKEN${NC}\n"

# Step 4: Book an appointment using the token
echo -e "${YELLOW}Step 4: Booking appointment (this should trigger email)...${NC}"
echo -e "${YELLOW}Watch the server console for MOCK EMAIL output!${NC}\n"

BOOKING_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/sessions/book/${BOOKING_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "dates": [
      {
        "date": "2025-12-23",
        "start_time": "10:00",
        "end_time": "11:00"
      }
    ],
    "client_first_name": "Max",
    "client_last_name": "Mustermann",
    "client_email": "max.mustermann@example.com",
    "client_phone": "+49123456789"
  }')

echo -e "${GREEN}Booking Response:${NC}"
echo $BOOKING_RESPONSE | python3 -m json.tool 2>/dev/null || echo $BOOKING_RESPONSE

echo -e "\n${YELLOW}=== Check server console for email output ===${NC}"
echo -e "${YELLOW}You should see:${NC}"
echo "  🚀 MOCK EMAIL SERVICE - EMAIL DATA"
echo "  📧 To: max.mustermann@example.com"
echo "  📋 Subject: Terminbestätigung - Test Booking Link for Email"
echo "  📄 HTML CONTENT with appointment details"
