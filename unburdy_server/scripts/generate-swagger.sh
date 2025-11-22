#!/bin/bash

# Regenerate Swagger Documentation
# Usage: ./scripts/generate-swagger.sh

set -e

echo "🔄 Regenerating Swagger documentation..."

# Ensure we're in the project root
cd "$(dirname "$0")/.."

# Generate swagger docs, including module directories (booking, calendar)
swag init -g main.go --parseDependency --parseInternal --dir ./,../modules/booking,../modules/calendar

echo "✅ Swagger documentation updated!"
echo "📋 View at: http://localhost:8080/swagger/index.html"
echo "📄 JSON spec: http://localhost:8080/swagger/doc.json"

# Optional: Format the generated files
if command -v jq > /dev/null; then
    echo "🎨 Formatting swagger.json..."
    jq . docs/swagger.json > docs/swagger.json.tmp && mv docs/swagger.json.tmp docs/swagger.json
fi

echo "✨ Done!"