#!/bin/bash

# Simple Base Module Update Script
# Usage: ./scripts/update-base.sh

set -e

echo "🔄 Updating base ae-saas module..."

# Ensure we're in the project root
cd "$(dirname "$0")/.."

echo "📦 Updating dependencies..."
go mod tidy

echo "🏗️ Testing build..."
if go build -o /tmp/test-build .; then
    echo "✅ Build successful!"
    rm -f /tmp/test-build
else
    echo "❌ Build failed! Check for breaking changes."
    exit 1
fi

echo "📚 Updating Swagger documentation..."
if command -v swag > /dev/null; then
    swag init --parseDependency --parseInternal
    echo "✅ Swagger docs updated!"
fi

echo "🧪 Running tests..."
if go test ./...; then
    echo "✅ All tests pass!"
fi

echo "✨ Update complete!"
echo ""
echo "🎯 **Simple Update Process:**"
echo "  1. Pull latest changes in ae-saas/server-api"
echo "  2. Run: go mod tidy"
echo "  3. Test: go build && go test ./..."
echo "  4. Update docs: swag init"
echo ""
echo "🔧 **We now use ae-saas public packages:**"
echo "  • github.com/ae-base-server/pkg/auth (JWT functions)"
echo "  • github.com/ae-base-server/pkg/utils (helper functions)"
echo "  • No more manual auth middleware or JWT handling!"