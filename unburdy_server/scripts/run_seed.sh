#!/bin/bash

# Script to seed the database with realistic client data
# Run from the unburdy_server directory

echo "🌱 Starting database seeding process..."
echo "📍 Working directory: $(pwd)"

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the unburdy_server directory (where go.mod is located)"
    exit 1
fi

# Check if seed data file exists
if [ ! -f "seed_app_data.json" ]; then
    echo "❌ Error: seed_app_data.json file not found in current directory"
    exit 1
fi

echo "📋 Found seed data file: seed_app_data.json"

# Run the seeding script
echo "🚀 Running database seeding script..."
go run ./scripts/seed_database.go

echo "✅ Database seeding process completed!"