#!/bin/bash

# Test script for Template Contract System
# Prerequisites: base-server must be running

BASE_URL="http://localhost:8080"
TENANT_ID="test-tenant"

echo "======================================"
echo "Template Contract System Test"
echo "======================================"
echo ""

# Get bearer token (assumes you have the bearer-tokens.json file)
if [ -f "../../bearer-tokens.json" ]; then
    TOKEN=$(jq -r ".\"$TENANT_ID\"" ../../bearer-tokens.json)
    if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
        echo "❌ Token not found for tenant: $TENANT_ID"
        exit 1
    fi
else
    echo "⚠️  bearer-tokens.json not found. Using default token."
    TOKEN="your-test-token-here"
fi

echo "Using tenant: $TENANT_ID"
echo "Token: ${TOKEN:0:20}..."
echo ""

# Test 1: Register a contract
echo "Test 1: Registering invoice contract..."
CONTRACT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/templates/contracts" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "module": "billing",
    "template_key": "invoice",
    "description": "Invoice template for client billing",
    "supported_channels": ["EMAIL", "DOCUMENT"],
    "variable_schema": {
      "invoice_number": {
        "type": "string",
        "required": true
      },
      "client": {
        "type": "object",
        "required": true,
        "properties": {
          "name": {
            "type": "string",
            "required": true
          },
          "email": {
            "type": "string",
            "required": true
          }
        }
      },
      "total_amount": {
        "type": "number",
        "required": true
      },
      "line_items": {
        "type": "array",
        "required": false,
        "items": {
          "type": "object"
        }
      }
    },
    "default_sample_data": {
      "invoice_number": "INV-2024-001",
      "client": {
        "name": "Sample Client GmbH",
        "email": "client@example.com"
      },
      "total_amount": 1500.00,
      "line_items": [
        {
          "description": "Consulting Services",
          "quantity": 10,
          "unit_price": 150.00,
          "total": 1500.00
        }
      ]
    }
  }')

if echo "$CONTRACT_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Contract registered successfully"
    CONTRACT_ID=$(echo "$CONTRACT_RESPONSE" | jq -r '.id')
    echo "   Contract ID: $CONTRACT_ID"
else
    echo "❌ Failed to register contract"
    echo "   Response: $CONTRACT_RESPONSE"
fi
echo ""

# Test 2: List contracts
echo "Test 2: Listing all contracts..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/api/templates/contracts" \
  -H "Authorization: Bearer $TOKEN")

CONTRACT_COUNT=$(echo "$LIST_RESPONSE" | jq 'length' 2>/dev/null || echo "0")
echo "   Found $CONTRACT_COUNT contract(s)"
if [ "$CONTRACT_COUNT" -gt "0" ]; then
    echo "✅ Contracts listed successfully"
else
    echo "⚠️  No contracts found"
fi
echo ""

# Test 3: Get contract by module and key
echo "Test 3: Getting contract by module and key..."
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/api/templates/contracts/billing/invoice" \
  -H "Authorization: Bearer $TOKEN")

if echo "$GET_RESPONSE" | jq -e '.module' > /dev/null 2>&1; then
    echo "✅ Contract retrieved successfully"
    echo "   Module: $(echo "$GET_RESPONSE" | jq -r '.module')"
    echo "   Key: $(echo "$GET_RESPONSE" | jq -r '.template_key')"
    echo "   Channels: $(echo "$GET_RESPONSE" | jq -r '.supported_channels | join(", ")')"
else
    echo "❌ Failed to retrieve contract"
    echo "   Response: $GET_RESPONSE"
fi
echo ""

# Test 4: Validate data against contract
echo "Test 4: Validating data against contract schema..."
VALIDATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/templates/validate?module=billing&template_key=invoice" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_number": "INV-2024-002",
    "client": {
      "name": "Test Client",
      "email": "test@example.com"
    },
    "total_amount": 2500.50
  }')

if echo "$VALIDATE_RESPONSE" | jq -e '.message' | grep -q "successful"; then
    echo "✅ Data validation passed"
else
    echo "⚠️  Validation response: $VALIDATE_RESPONSE"
fi
echo ""

# Test 5: Get required fields
echo "Test 5: Getting required fields for contract..."
FIELDS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/templates/required-fields?module=billing&template_key=invoice" \
  -H "Authorization: Bearer $TOKEN")

REQUIRED_FIELDS=$(echo "$FIELDS_RESPONSE" | jq -r '.required_fields | join(", ")')
echo "   Required fields: $REQUIRED_FIELDS"
if [ ! -z "$REQUIRED_FIELDS" ]; then
    echo "✅ Required fields retrieved"
else
    echo "⚠️  No required fields found"
fi
echo ""

echo "======================================"
echo "Test Summary"
echo "======================================"
echo "All contract system tests completed!"
echo ""
echo "Next steps:"
echo "1. Create a template instance using the contract"
echo "2. Test template rendering with sample data"
echo "3. Test public asset delivery"
echo ""
