#!/bin/bash

# Example script demonstrating token validation with limited use
# This script shows how to create different types of booking links

BASE_URL="http://localhost:8080/api/v1"
TOKEN="your-auth-token-here"

echo "========================================="
echo "Booking Token Validation Examples"
echo "========================================="
echo ""

# Example 1: One-time booking link (default 24 hours)
echo "1. Creating a one-time booking link (expires in 24 hours)..."
curl -X POST "$BASE_URL/booking/link" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 1,
    "client_id": 123,
    "token_purpose": "one-time-booking-link"
  }'
echo -e "\n"

# Example 2: One-time booking link with custom expiration (7 days)
echo "2. Creating a one-time booking link (expires in 7 days)..."
curl -X POST "$BASE_URL/booking/link" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 1,
    "client_id": 123,
    "token_purpose": "one-time-booking-link",
    "validity_days": 7
  }'
echo -e "\n"

# Example 3: Limited use link (5 times, 30 days)
echo "3. Creating a limited use link (5 uses, 30 days)..."
curl -X POST "$BASE_URL/booking/link" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 1,
    "client_id": 123,
    "token_purpose": "timed-booking-link",
    "max_use_count": 5,
    "validity_days": 30
  }'
echo -e "\n"

# Example 4: Time-limited reusable link (unlimited uses, 7 days)
echo "4. Creating a time-limited reusable link (unlimited uses, 7 days)..."
curl -X POST "$BASE_URL/booking/link" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 1,
    "client_id": 123,
    "token_purpose": "timed-booking-link",
    "max_use_count": 0,
    "validity_days": 7
  }'
echo -e "\n"

# Example 5: Permanent unlimited link
echo "5. Creating a permanent unlimited link..."
curl -X POST "$BASE_URL/booking/link" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 1,
    "client_id": 123,
    "token_purpose": "timed-booking-link"
  }'
echo -e "\n"

# Example 6: Package deal link (10 uses, 90 days)
echo "6. Creating a package deal link (10 uses, 90 days)..."
curl -X POST "$BASE_URL/booking/link" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": 1,
    "client_id": 123,
    "token_purpose": "timed-booking-link",
    "max_use_count": 10,
    "validity_days": 90
  }'
echo -e "\n"

echo "========================================="
echo "Using a Booking Link"
echo "========================================="
echo ""

# Replace BOOKING_TOKEN with an actual token from the responses above
BOOKING_TOKEN="your-booking-token-here"

echo "7. Getting free slots with the booking token..."
curl -X GET "$BASE_URL/booking/freeslots/$BOOKING_TOKEN?start=2025-12-22&end=2025-12-31"
echo -e "\n"

echo "8. Getting client information with the booking token..."
curl -X GET "$BASE_URL/client/$BOOKING_TOKEN"
echo -e "\n"

echo "========================================="
echo "Error Scenarios"
echo "========================================="
echo ""

# These would show error messages after token expires or reaches limit
echo "9. Using an expired token (will fail with 'Token expired' error)..."
echo "   Run the same request after the token expires"
echo ""

echo "10. Using a token that reached its limit (will fail with 'Token limit reached' error)..."
echo "    Use a one-time token twice, or a limited token more than max_use_count times"
echo ""

echo "11. Using a revoked token (will fail with 'Token revoked' error)..."
echo "    Revoke the token via blacklist, then try to use it"
echo ""

echo "========================================="
echo "Examples complete!"
echo "========================================="
