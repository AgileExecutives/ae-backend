#!/bin/bash

# Test script for mock email latest-emails endpoint

BASE_URL="http://localhost:8080/api/v1"

echo "========================================="
echo "Testing Mock Email Latest-Emails Endpoint"
echo "========================================="
echo ""

# First, send a test email to populate the mock emails file
echo "1. Sending a test email..."
curl -X POST "$BASE_URL/emails/send" \
  -H "Content-Type: application/json" \
  -d '{
    "to_email": "test@example.com",
    "subject": "Test Email",
    "body": "This is a test email body.",
    "html_body": "<h1>Test</h1><p>This is a test email body.</p>"
  }' \
  -s -o /dev/null -w "HTTP Status: %{http_code}\n"

echo ""
echo "2. Retrieving latest mock emails..."
curl -X GET "$BASE_URL/emails/latest-emails" \
  -H "Content-Type: application/json" \
  -s | jq '.'

echo ""
echo "========================================="
echo "Test Complete!"
echo "========================================="
echo ""
echo "Note: This endpoint is only available when MOCK_EMAIL=true"
echo "If you get a 503 error, make sure MOCK_EMAIL=true in your .env file"
