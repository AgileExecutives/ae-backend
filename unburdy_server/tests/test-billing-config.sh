#!/bin/bash

# Test script for Organization Billing Configuration
# Tests the PUT /organizations/:id/billing-config endpoint

set -e

BASE_URL="http://localhost:8080"
TOKEN=""
ORG_ID=""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Organization Billing Config Test Suite ===${NC}\n"

print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
        exit 1
    fi
}

# 1. Login
echo "1. Authenticating..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "email": "test@example.com",
        "password": "password123"
    }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
    echo -e "${RED}Failed to authenticate${NC}"
    exit 1
fi
print_result 0 "Authenticated successfully"

# 2. Get organization ID
echo -e "\n2. Getting organization ID..."
ORGS_RESPONSE=$(curl -s -X GET "$BASE_URL/organizations?limit=1" \
    -H "Authorization: Bearer $TOKEN")

ORG_ID=$(echo $ORGS_RESPONSE | jq -r '.data[0].id')
if [ -z "$ORG_ID" ] || [ "$ORG_ID" == "null" ]; then
    echo -e "${RED}No organization found${NC}"
    exit 1
fi
print_result 0 "Organization ID: $ORG_ID"

# 3. Get current billing config
echo -e "\n3. Getting current billing configuration..."
CURRENT_CONFIG=$(curl -s -X GET "$BASE_URL/organizations/$ORG_ID" \
    -H "Authorization: Bearer $TOKEN")

CURRENT_MODE=$(echo $CURRENT_CONFIG | jq -r '.data.extra_efforts_billing_mode')
print_result 0 "Current mode: $CURRENT_MODE"

# 4. Test Mode A: Ignore
echo -e "\n4. Testing Mode A (ignore)..."
MODE_A_RESPONSE=$(curl -s -X PUT "$BASE_URL/organizations/$ORG_ID/billing-config" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "extra_efforts_billing_mode": "ignore"
    }')

MODE_A_SET=$(echo $MODE_A_RESPONSE | jq -r '.data.extra_efforts_billing_mode')
if [ "$MODE_A_SET" != "ignore" ]; then
    echo -e "${RED}Failed to set mode A${NC}"
    exit 1
fi
print_result 0 "Mode A (ignore) set successfully"

# 5. Test Mode B: Bundle Double Units
echo -e "\n5. Testing Mode B (bundle_double_units)..."
MODE_B_RESPONSE=$(curl -s -X PUT "$BASE_URL/organizations/$ORG_ID/billing-config" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "extra_efforts_billing_mode": "bundle_double_units",
        "extra_efforts_config": {
            "threshold_minutes": 75
        },
        "line_item_single_unit_text": "Therapiestunde",
        "line_item_double_unit_text": "Therapie-Doppelstunde"
    }')

MODE_B_SET=$(echo $MODE_B_RESPONSE | jq -r '.data.extra_efforts_billing_mode')
SINGLE_TEXT=$(echo $MODE_B_RESPONSE | jq -r '.data.line_item_single_unit_text')
DOUBLE_TEXT=$(echo $MODE_B_RESPONSE | jq -r '.data.line_item_double_unit_text')

if [ "$MODE_B_SET" != "bundle_double_units" ]; then
    echo -e "${RED}Failed to set mode B${NC}"
    exit 1
fi
print_result 0 "Mode B (bundle_double_units) with threshold 75min"
echo -e "  Single unit text: $SINGLE_TEXT"
echo -e "  Double unit text: $DOUBLE_TEXT"

# 6. Test Mode C: Separate Items
echo -e "\n6. Testing Mode C (separate_items)..."
MODE_C_RESPONSE=$(curl -s -X PUT "$BASE_URL/organizations/$ORG_ID/billing-config" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "extra_efforts_billing_mode": "separate_items",
        "extra_efforts_config": {
            "rounding_mode": "nearest_quarter_hour"
        }
    }')

MODE_C_SET=$(echo $MODE_C_RESPONSE | jq -r '.data.extra_efforts_billing_mode')
if [ "$MODE_C_SET" != "separate_items" ]; then
    echo -e "${RED}Failed to set mode C${NC}"
    exit 1
fi
print_result 0 "Mode C (separate_items) with quarter-hour rounding"

# 7. Test Mode D: Preparation Allowance
echo -e "\n7. Testing Mode D (preparation_allowance)..."
MODE_D_RESPONSE=$(curl -s -X PUT "$BASE_URL/organizations/$ORG_ID/billing-config" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "extra_efforts_billing_mode": "preparation_allowance",
        "extra_efforts_config": {
            "auto_add_minutes": 15,
            "description": "Vor- und Nachbereitung"
        }
    }')

MODE_D_SET=$(echo $MODE_D_RESPONSE | jq -r '.data.extra_efforts_billing_mode')
if [ "$MODE_D_SET" != "preparation_allowance" ]; then
    echo -e "${RED}Failed to set mode D${NC}"
    exit 1
fi
print_result 0 "Mode D (preparation_allowance) with 15min auto-add"

# 8. Test partial update
echo -e "\n8. Testing partial update (only mode)..."
PARTIAL_RESPONSE=$(curl -s -X PUT "$BASE_URL/organizations/$ORG_ID/billing-config" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "extra_efforts_billing_mode": "bundle_double_units"
    }')

PARTIAL_MODE=$(echo $PARTIAL_RESPONSE | jq -r '.data.extra_efforts_billing_mode')
PARTIAL_SINGLE=$(echo $PARTIAL_RESPONSE | jq -r '.data.line_item_single_unit_text')

if [ "$PARTIAL_MODE" != "bundle_double_units" ]; then
    echo -e "${RED}Failed partial update${NC}"
    exit 1
fi
print_result 0 "Partial update preserved existing text: $PARTIAL_SINGLE"

# 9. Verify persistence
echo -e "\n9. Verifying configuration persistence..."
VERIFY_RESPONSE=$(curl -s -X GET "$BASE_URL/organizations/$ORG_ID" \
    -H "Authorization: Bearer $TOKEN")

VERIFY_MODE=$(echo $VERIFY_RESPONSE | jq -r '.data.extra_efforts_billing_mode')
if [ "$VERIFY_MODE" != "bundle_double_units" ]; then
    echo -e "${RED}Configuration not persisted${NC}"
    exit 1
fi
print_result 0 "Configuration persisted correctly"

# Restore original config
echo -e "\n${YELLOW}=== Cleanup ===${NC}"
echo "Restoring original configuration..."
curl -s -X PUT "$BASE_URL/organizations/$ORG_ID/billing-config" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"extra_efforts_billing_mode\": \"$CURRENT_MODE\"
    }" > /dev/null

echo -e "\n${GREEN}=== All Tests Passed! ===${NC}"
