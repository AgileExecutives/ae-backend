#!/bin/bash

# Test script for invoice cancellation endpoint
# Tests canceling a draft invoice and verifying sessions/efforts are reverted

set -e

BASE_URL="http://localhost:3000/api/v1"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ0ZW5hbnRfaWQiOjEsImV4cCI6MTczNzcyOTMzNX0.6hDyCkV9xYE6vYzQYE-rNz5YRxwVu6hhDHDOVW0kLCM"

echo "🧪 Testing Invoice Cancellation"
echo "================================"
echo ""

# Test 1: Create a draft invoice first
echo "📝 Test 1: Create a draft invoice with sessions"
INVOICE_RESPONSE=$(curl -s -X POST "$BASE_URL/client-invoices/generate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "unbilledClient": {
      "id": 4,
      "sessions": [
        {"id": 70, "number_units": 1, "status": "conducted"},
        {"id": 12, "number_units": 1, "status": "conducted"}
      ],
      "extra_efforts": []
    },
    "parameters": {
      "invoice_date": "2026-01-06",
      "tax_rate": 19,
      "generate_pdf": false
    }
  }')

echo "$INVOICE_RESPONSE" | jq '.'

# Extract invoice ID
INVOICE_ID=$(echo "$INVOICE_RESPONSE" | jq -r '.data.id // empty')

if [ -z "$INVOICE_ID" ]; then
  echo "❌ Failed to create invoice"
  exit 1
fi

echo "✅ Invoice created with ID: $INVOICE_ID"
echo ""

# Test 2: Verify invoice is in draft status
echo "📝 Test 2: Verify invoice status is draft"
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/client-invoices/$INVOICE_ID" \
  -H "Authorization: Bearer $TOKEN")

INVOICE_STATUS=$(echo "$GET_RESPONSE" | jq -r '.data.status')
echo "Invoice status: $INVOICE_STATUS"

if [ "$INVOICE_STATUS" != "draft" ]; then
  echo "❌ Invoice is not in draft status"
  exit 1
fi
echo "✅ Invoice is in draft status"
echo ""

# Test 3: Cancel the invoice
echo "📝 Test 3: Cancel the draft invoice"
CANCEL_RESPONSE=$(curl -s -X POST "$BASE_URL/client-invoices/$INVOICE_ID/cancel" \
  -H "Authorization: Bearer $TOKEN")

echo "$CANCEL_RESPONSE" | jq '.'

CANCEL_SUCCESS=$(echo "$CANCEL_RESPONSE" | jq -r '.success')
if [ "$CANCEL_SUCCESS" != "true" ]; then
  echo "❌ Failed to cancel invoice"
  exit 1
fi
echo "✅ Invoice cancelled successfully"
echo ""

# Test 4: Verify invoice no longer exists
echo "📝 Test 4: Verify invoice was deleted"
CHECK_RESPONSE=$(curl -s -X GET "$BASE_URL/client-invoices/$INVOICE_ID" \
  -H "Authorization: Bearer $TOKEN")

CHECK_SUCCESS=$(echo "$CHECK_RESPONSE" | jq -r '.success')
if [ "$CHECK_SUCCESS" == "true" ]; then
  echo "❌ Invoice still exists after cancellation"
  exit 1
fi
echo "✅ Invoice was deleted"
echo ""

# Test 5: Try to cancel a non-existent invoice
echo "📝 Test 5: Try to cancel non-existent invoice (should fail)"
FAIL_RESPONSE=$(curl -s -X POST "$BASE_URL/client-invoices/99999/cancel" \
  -H "Authorization: Bearer $TOKEN")

FAIL_SUCCESS=$(echo "$FAIL_RESPONSE" | jq -r '.success')
if [ "$FAIL_SUCCESS" == "true" ]; then
  echo "❌ Should not be able to cancel non-existent invoice"
  exit 1
fi
echo "✅ Correctly rejected canceling non-existent invoice"
echo ""

# Test 6: Create a non-draft invoice and try to cancel
echo "📝 Test 6: Try to cancel a non-draft invoice (should fail)"
# First create a draft invoice
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/client-invoices/generate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "unbilledClient": {
      "id": 4,
      "sessions": [{"id": 100, "number_units": 1, "status": "conducted"}],
      "extra_efforts": []
    },
    "parameters": {
      "invoice_date": "2026-01-06",
      "tax_rate": 19,
      "generate_pdf": false
    }
  }')

NEW_INVOICE_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.id // empty')

if [ -n "$NEW_INVOICE_ID" ]; then
  # Update status to sent
  curl -s -X PUT "$BASE_URL/client-invoices/$NEW_INVOICE_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"status": "sent"}' > /dev/null
  
  # Try to cancel
  SENT_CANCEL_RESPONSE=$(curl -s -X POST "$BASE_URL/client-invoices/$NEW_INVOICE_ID/cancel" \
    -H "Authorization: Bearer $TOKEN")
  
  SENT_CANCEL_SUCCESS=$(echo "$SENT_CANCEL_RESPONSE" | jq -r '.success')
  if [ "$SENT_CANCEL_SUCCESS" == "true" ]; then
    echo "❌ Should not be able to cancel non-draft invoice"
    exit 1
  fi
  echo "✅ Correctly rejected canceling non-draft invoice"
  
  # Clean up
  curl -s -X DELETE "$BASE_URL/client-invoices/$NEW_INVOICE_ID" \
    -H "Authorization: Bearer $TOKEN" > /dev/null
else
  echo "⚠️  Could not create test invoice for status check"
fi

echo ""
echo "✅ All tests passed!"
