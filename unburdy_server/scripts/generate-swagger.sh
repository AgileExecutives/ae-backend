#!/bin/bash

# Regenerate Swagger Documentation
# Usage: ./scripts/generate-swagger.sh

set -e

echo "🔄 Regenerating Swagger documentation..."

# Ensure we're in the project root
cd "$(dirname "$0")/.."

# Generate swagger docs, including all module directories (excluding test directories)
swag init -g main.go --parseDependency --parseInternal --exclude ../modules/invoice/tests,../modules/booking/tests,../modules/calendar/tests --dir ./,../modules/booking,../modules/calendar,../modules/documents,../modules/invoice,./modules/client_management,../base-server/modules/base,../base-server/modules/customer,../base-server/modules/email,../base-server/modules/pdf,../base-server/modules/static,../base-server/modules/templates,../base-server/modules/organization,../base-server/internal/organizations/handlers,../base-server/internal/organizations/routes,../base-server/internal/organizations/services,../base-server/modules/settings_api

echo "✅ Swagger documentation updated!"
echo "📋 View at: http://localhost:8080/swagger/index.html"
echo "📄 JSON spec: http://localhost:8080/swagger/doc.json"

# Optional: Format the generated files
if command -v jq > /dev/null; then
    echo "🎨 Formatting swagger.json..."
    jq . docs/swagger.json > docs/swagger.json.tmp && mv docs/swagger.json.tmp docs/swagger.json
fi

echo "✨ Done!"