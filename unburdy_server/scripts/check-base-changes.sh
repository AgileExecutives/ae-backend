#!/bin/bash

# Monitor Base Module Changes
# Usage: ./scripts/check-base-changes.sh

set -e

BASE_MODULE_PATH="../../ae-saas/server-api"
TEMP_DIR="/tmp/ae-saas-diff"

echo "🔍 Checking for changes in base ae-saas module..."

# Ensure we're in the project root
cd "$(dirname "$0")/.."

if [ ! -d "$BASE_MODULE_PATH" ]; then
    echo "❌ Base module not found at: $BASE_MODULE_PATH"
    exit 1
fi

# Create temp directory for comparison
mkdir -p "$TEMP_DIR"

echo "📂 Analyzing key files for changes..."

# Files to monitor for changes
declare -a MONITOR_FILES=(
    "internal/models"
    "internal/handlers" 
    "internal/middleware"
    "internal/router/router.go"
    "go.mod"
    "main.go"
)

echo "🔍 Files being monitored:"
for file in "${MONITOR_FILES[@]}"; do
    if [ -e "$BASE_MODULE_PATH/$file" ]; then
        echo "  ✅ $file"
        # Get last modified date
        if [[ "$OSTYPE" == "darwin"* ]]; then
            mod_date=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" "$BASE_MODULE_PATH/$file")
        else
            mod_date=$(stat -c "%y" "$BASE_MODULE_PATH/$file")
        fi
        echo "    Last modified: $mod_date"
    else
        echo "  ❌ $file (not found)"
    fi
done

echo ""
echo "📝 Recent git commits in base module:"
cd "$BASE_MODULE_PATH"
if git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Last 5 commits:"
    git log --oneline -5
    echo ""
    echo "Changes since last week:"
    git log --since="1 week ago" --oneline
else
    echo "Not a git repository or no git history"
fi

cd - > /dev/null

echo ""
echo "🏗️ Current unburdy module status:"
echo "  Go module: $(grep "module " go.mod)"
echo "  Base dependency: $(grep "ae-saas-basic" go.mod || echo "Not found")"

echo ""
echo "💡 Next steps if changes detected:"
echo "  1. Run: ./scripts/update-base.sh"
echo "  2. Review: UPDATE_CHECKLIST.md"
echo "  3. Test: go test ./..."
echo "  4. Update docs: ./scripts/generate-swagger.sh"