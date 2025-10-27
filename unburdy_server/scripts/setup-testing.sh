#!/bin/bash

# Comprehensive API Testing Setup Script
# This script sets up multiple testing approaches for your Swagger-based API

set -e

echo "🧪 Setting up comprehensive API testing..."

# 1. Install Go testing dependencies
echo "📦 Installing Go testing dependencies..."
go mod tidy

# 2. Install Newman for Postman collection testing
echo "📦 Installing Newman (Postman CLI)..."
if ! command -v newman &> /dev/null; then
    npm install -g newman
fi

# 3. Install HURL for HTTP testing
echo "📦 Installing HURL..."
if ! command -v hurl &> /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install hurl
    else
        echo "Please install HURL manually: https://hurl.dev/docs/installation.html"
    fi
fi

# 4. Install OpenAPI Generator
echo "📦 Installing OpenAPI Generator..."
if ! command -v openapi-generator &> /dev/null; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install openapi-generator
    else
        echo "Please install OpenAPI Generator manually"
    fi
fi

echo "✅ API Testing setup complete!"
echo ""
echo "Available testing approaches:"
echo "1. Go Native Tests: make test"
echo "2. HURL Integration Tests: make test-integration"  
echo "3. Postman Collections: make test-postman"
echo "4. All Tests: make test-all"