#!/bin/bash

echo "🌱 Database Management for Unburdy Server"
echo "========================================"

# Get the script's directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Database environment variables with defaults
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432}
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-pass}
export DB_NAME=${DB_NAME:-ae_saas_basic_test}
export DB_SSL_MODE=${DB_SSL_MODE:-disable}

echo "📍 Project root: $PROJECT_ROOT"
echo "🔗 Database: $DB_HOST:$DB_PORT/$DB_NAME"

case "$1" in
    "seed")
        echo "🌱 Seeding database with complete application data..."
        echo "   - Base server entities (users, tenants, plans)"
        echo "   - Cost providers and clients from seed data"
        cd "$PROJECT_ROOT/seed" && go run seed_database.go
        ;;
    "calendars")
        echo "📅 Seeding only calendar data..."
        echo "   - Calendar entries with appointments"
        echo "   - German holidays (public and school holidays)"
        echo "   - Recurring series and events"
        export SEED_CALENDAR_ONLY=true
        cd "$PROJECT_ROOT/seed" && go run seed_database.go
        ;;
    "base")
        echo "🌱 Seeding only base server data (users, tenants, plans)..."
        cd "$PROJECT_ROOT" && go run -c "
            package main
            import (
                \"log\"
                baseAPI \"github.com/ae-base-server/api\"
                \"gorm.io/driver/postgres\"
                \"gorm.io/gorm\"
            )
            func main() {
                dsn := \"host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=$DB_SSL_MODE\"
                db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
                if err != nil { log.Fatal(err) }
                if err := baseAPI.SeedBaseData(db); err != nil { log.Fatal(err) }
                log.Println(\"✅ Base seeding completed\")
            }
        " -
        ;;
    "clients")
        echo "🌱 Seeding only client data..."
        cd "$PROJECT_ROOT/seed" && go run -tags clients seed_database.go
        ;;
    "clear")
        echo "🗑️  Clearing all seeded data..."
        cd "$PROJECT_ROOT/seed" && go run -tags clear seed_database.go
        ;;
    "test"|"check")
        echo "🔍 Testing database contents..."
        cd "$PROJECT_ROOT/seed" && go run -tags test seed_database.go
        ;;
    "reset")
        echo "🗑️  Clearing and re-seeding database..."
        echo "   Step 1: Clearing existing data..."
        cd "$PROJECT_ROOT/seed" && go run -tags clear seed_database.go
        echo "   Step 2: Seeding fresh data..."
        cd "$PROJECT_ROOT/seed" && go run seed_database.go
        ;;
    *)
        echo "Usage: $0 {seed|calendars|base|clients|clear|test|reset}"
        echo ""
        echo "Commands:"
        echo "  seed     - Complete seeding (base + clients + cost providers + calendars)"
        echo "  calendars- Seed only calendar data (appointments + holidays)"
        echo "  base     - Seed only base server data (users, tenants, plans)"
        echo "  clients  - Seed only client data (requires base data)"
        echo "  clear    - Clear all seeded data"
        echo "  test     - Verify database contents and show statistics"
        echo "  reset    - Clear and re-seed everything"
        echo ""
        echo "Environment Variables:"
        echo "  DB_HOST=$DB_HOST"
        echo "  DB_PORT=$DB_PORT"
        echo "  DB_USER=$DB_USER"
        echo "  DB_NAME=$DB_NAME"
        echo ""
        echo "Examples:"
        echo "  ./manage_db.sh seed"
        echo "  ./manage_db.sh calendars"
        echo "  ./manage_db.sh test"
        echo "  DB_NAME=my_test_db ./manage_db.sh reset"
        exit 1
        ;;
esac