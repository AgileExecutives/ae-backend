#!/bin/bash

# Migration script for calendar module schema changes
# Updates calendar_series table structure

set -e

echo "🔄 Running calendar module migration..."

# Database connection details - update these as needed
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-unburdy_db}"
DB_USER="${DB_USER:-postgres}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATION_FILE="$SCRIPT_DIR/001_update_calendar_series_schema.sql"

# Check if migration file exists
if [ ! -f "$MIGRATION_FILE" ]; then
    echo "❌ Migration file not found: $MIGRATION_FILE"
    exit 1
fi

echo "📋 Migration file: $MIGRATION_FILE"
echo "🗄️  Database: $DB_NAME on $DB_HOST:$DB_PORT"
echo ""

# Run migration
if command -v psql &> /dev/null; then
    echo "▶️  Executing migration..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE"
    echo ""
    echo "✅ Migration completed successfully!"
else
    echo "⚠️  psql not found. Please run the migration manually:"
    echo ""
    echo "psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $MIGRATION_FILE"
    echo ""
    echo "Or copy and paste the SQL from:"
    echo "$MIGRATION_FILE"
fi
