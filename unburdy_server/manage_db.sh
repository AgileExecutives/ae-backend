#!/bin/bash

echo "🌱 Database Management for Unburdy Server"
echo "========================================"

case "$1" in
    "seed")
        echo "🌱 Seeding database with client data..."
        cd scripts && go run seed_database.go
        ;;
    "test")
        echo "🔍 Testing database contents..."
        cd scripts && go run test_clients.go
        ;;
    "reset")
        echo "🗑️  Clearing and re-seeding database..."
        cd scripts && go run seed_database.go
        ;;
    *)
        echo "Usage: $0 {seed|test|reset}"
        echo ""
        echo "Commands:"
        echo "  seed  - Populate database with 46 realistic clients"
        echo "  test  - Verify database contents and show sample data"
        echo "  reset - Clear existing clients and re-seed"
        echo ""
        echo "Examples:"
        echo "  ./manage_db.sh seed"
        echo "  ./manage_db.sh test"
        exit 1
        ;;
esac