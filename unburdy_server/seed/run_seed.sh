#!/bin/bash

# Simple wrapper for consolidated seeding
# Can be run from any directory within the project

echo "🌱 Starting database seeding process..."
echo "📍 Working directory: $(pwd)"

# Find project root (directory with go.mod)
PROJECT_ROOT=""
CURRENT_DIR="$(pwd)"

while [ "$CURRENT_DIR" != "/" ]; do
    if [ -f "$CURRENT_DIR/go.mod" ]; then
        PROJECT_ROOT="$CURRENT_DIR"
        break
    fi
    CURRENT_DIR="$(dirname "$CURRENT_DIR")"
done

if [ -z "$PROJECT_ROOT" ]; then
    echo "❌ Error: Could not find project root (go.mod file)"
    exit 1
fi

echo "📍 Project root: $PROJECT_ROOT"
cd "$PROJECT_ROOT"

# Check if seed data files exist
if [ ! -f "startupseed/seed-data.json" ]; then
    echo "❌ Error: startupseed/seed-data.json file not found"
    exit 1
fi

if [ ! -f "seed/seed_app_data.json" ]; then
    echo "❌ Error: seed/seed_app_data.json file not found"
    exit 1
fi

echo "📋 Found seed data files"

# Use the consolidated seed script
echo "🚀 Running consolidated seeding script..."
cd seed && go run seed_database.go

echo "✅ Database seeding process completed!"