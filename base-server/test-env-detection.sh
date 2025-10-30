#!/bin/bash

# Test Environment Variable Detection from Health Endpoint
# Demonstrates how test scripts can detect server configuration

set -e

HOST="http://localhost:8081"

echo "🔍 Testing Environment Variable Detection from Health Endpoint"
echo "============================================================="

# Get health response
HEALTH_RESPONSE=$(curl -s "${HOST}/api/v1/health")

echo "📡 Health Endpoint Response:"
echo "$HEALTH_RESPONSE" | jq .

# Extract environment variables
MOCK_EMAIL=$(echo "$HEALTH_RESPONSE" | jq -r '.environment.mock_email')
RATE_LIMIT_ENABLED=$(echo "$HEALTH_RESPONSE" | jq -r '.environment.rate_limit_enabled')
EMAIL_VERIFICATION=$(echo "$HEALTH_RESPONSE" | jq -r '.environment.email_verification')
GIN_MODE=$(echo "$HEALTH_RESPONSE" | jq -r '.environment.gin_mode')

echo ""
echo "🔧 Detected Environment Configuration:"
echo "======================================"
echo "Mock Email: $MOCK_EMAIL"
echo "Rate Limiting: $RATE_LIMIT_ENABLED"
echo "Email Verification: $EMAIL_VERIFICATION"
echo "Gin Mode: $GIN_MODE"

echo ""
echo "📋 Test Configuration Recommendations:"
echo "====================================="

if [ "$MOCK_EMAIL" = "true" ]; then
    echo "✅ Email mocking is ENABLED - emails will be logged, not sent"
    echo "   → Safe for testing email functionality"
else
    echo "⚠️  Email mocking is DISABLED - emails will be sent via SMTP"
    echo "   → Use test email addresses only"
fi

if [ "$RATE_LIMIT_ENABLED" = "true" ]; then
    echo "⚠️  Rate limiting is ENABLED - tests may be throttled"
    echo "   → Consider spacing out authentication tests"
    echo "   → Or disable rate limiting for testing: RATE_LIMIT_ENABLED=false"
else
    echo "✅ Rate limiting is DISABLED - no throttling during tests"
    echo "   → Tests can run at full speed"
fi

if [ "$EMAIL_VERIFICATION" = "true" ]; then
    echo "⚠️  Email verification is ENABLED - new users need verification"
    echo "   → Tests may need to handle email verification flow"
else
    echo "✅ Email verification is DISABLED - new users are immediately active"
    echo "   → Tests can skip email verification steps"
fi

echo ""
echo "🎯 Test Script Adaptation:"
echo "========================="

# Example of how tests can adapt based on environment
if [ "$RATE_LIMIT_ENABLED" = "true" ]; then
    echo "   → Add delays between authentication attempts"
    echo "   → Reduce concurrent test execution"
    echo "   → Use existing test users instead of creating new ones"
fi

if [ "$MOCK_EMAIL" = "false" ]; then
    echo "   → Use disposable email services for test accounts"
    echo "   → Monitor email delivery in test environment"
fi

if [ "$EMAIL_VERIFICATION" = "true" ]; then
    echo "   → Include email verification steps in user registration tests"
    echo "   → Test both verified and unverified user scenarios"
fi

echo ""
echo "✅ Environment detection test completed!"