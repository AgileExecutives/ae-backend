#!/bin/bash

# Migration script for client_management module
# Runs all pending migrations for sessions table

set -e

echo "🔄 Running client_management module migrations..."

# Database connection details - update these as needed
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-unburdy_db}"
DB_USER="${DB_USER:-postgres}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "⚠️  psql not found. Please install PostgreSQL client tools."
    echo ""
    echo "To run migrations manually, execute these SQL files in order:"
    for migration in "$SCRIPT_DIR"/*.sql; do
        if [ -f "$migration" ]; then
            echo "  - $(basename "$migration")"
        fi
    done
    echo ""
    echo "psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f <migration_file>"
    exit 1
fi

echo "🗄️  Database: $DB_NAME on $DB_HOST:$DB_PORT"
echo ""

# Run all migrations in order
for migration in "$SCRIPT_DIR"/*.sql; do
    if [ -f "$migration" ]; then
        filename=$(basename "$migration")
        echo "▶️  Running migration: $filename"
        
        if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration" > /dev/null 2>&1; then
            echo "   ✅ $filename completed"
        else
            echo "   ⚠️  $filename may have already been applied or encountered an error"
            echo "   (This is normal if the migration uses IF NOT EXISTS)"
        fi
        echo ""
    fi
done

echo "✅ All migrations completed!"
echo ""
echo "To verify the changes, you can run:"
echo "psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c '\d sessions'"
