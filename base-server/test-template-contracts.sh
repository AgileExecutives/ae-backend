#!/bin/bash
# Test script to verify template contracts are registered

echo "🔍 Testing Template Contract Registration"
echo "========================================="
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Database connection details from .env or defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-ae_base}"

echo "📊 Checking template_contracts table..."
echo ""

# Query to check template contracts
QUERY="SELECT module, template_key, description, supported_channels FROM template_contracts ORDER BY module, template_key;"

# Execute query using psql
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$QUERY"

echo ""
echo "📈 Contract count by module:"
COUNT_QUERY="SELECT module, COUNT(*) as contract_count FROM template_contracts GROUP BY module ORDER BY module;"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$COUNT_QUERY"

echo ""
echo "✅ Test complete!"
