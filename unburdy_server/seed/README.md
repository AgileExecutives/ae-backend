# Unburdy Server Database Seeding

This directory contains consolidated seeding functionality for the Unburdy Server application. It handles both base-server entities and application-specific data with a unified approach.

## Overview

The seeding system automatically:
- **Auto-migrates** all database entities (base-server + modules)
- **Seeds base-server data** (tenants, users, plans) from `seed-data.json`
- **Seeds application data** (cost providers, clients) from `seed_app_data.json`
- **Provides statistics** and verification of seeded data

## Files

- `manage_db.sh` - Main management script with multiple commands
- `run_seed.sh` - Simple wrapper script for basic seeding
- `seed_database.go` - Consolidated Go seeding application
- `seed_app_data.json` - Application-specific seed data (1600+ lines)

## Usage

### Management Script (Recommended)

```bash
# Full seeding (base + application data)
./manage_db.sh seed

# Test/check database contents
./manage_db.sh test

# Clear all data and re-seed
./manage_db.sh reset

# Show all available commands
./manage_db.sh
```

### Environment Variables

The seeding system uses these database connection variables:

```bash
DB_HOST=localhost        # Default: localhost
DB_PORT=5432            # Default: 5432
DB_USER=postgres        # Default: postgres
DB_PASSWORD=pass        # Default: pass
DB_NAME=ae_saas_basic_test  # Default: ae_saas_basic_test
DB_SSL_MODE=disable     # Default: disable
```

### Custom Database Example

```bash
DB_NAME=my_custom_db ./manage_db.sh seed
```

## What Gets Seeded

### Base-Server Entities (from `../seed-data.json`)
- **Tenants** - Multi-tenancy support (admin, default-tenant)
- **Users** - System users (admin, testuser)
- **Plans** - Subscription plans (Basic, Pro)

### Application Entities (from `seed_app_data.json`)
- **Cost Providers** - 45 German youth offices (Jugendämter)
- **Clients** - 46 realistic client records with therapy information

### Auto-Migration
- All base-server entities (users, tenants, customers, emails, etc.)
- All module entities (clients, cost_providers, calendar tables, etc.)

## Statistics

After seeding, you'll see detailed statistics:

```
📊 Seeding Statistics
====================
📈 Total clients: 46
🏢 Total cost providers: 45

🎯 Therapy Type Breakdown:
   Leichte Intelligenzminderung: 7 clients
   Anpassungsstörung: 4 clients
   Expressive Sprachstörung: 4 clients
   [... and more]

📋 Status Breakdown:
   active: 40 clients
   archived: 3 clients
   waiting: 3 clients
```

## Features

### Smart Duplicate Prevention
- Checks for existing records before creating
- Re-running seeding is safe (won't create duplicates)

### Comprehensive Auto-Migration
- Uses base-server `MigrateBaseEntities()` for core entities
- Uses module auto-migration for application entities
- Works with the new unified migration system

### Realistic Test Data
- German therapy client data with proper diagnostic codes
- Real German youth office information
- Proper relationships between clients and cost providers

## Integration

This seeding system integrates with:
- **Base-Server API** - Uses `baseAPI.SeedBaseData()` and `baseAPI.MigrateBaseEntities()`
- **Module System** - Auto-migrates all module entities
- **Database Management** - Part of the overall database workflow

## Troubleshooting

### Common Issues

1. **"seed-data.json not found"**
   - Ensure you're running from the correct directory
   - The script will look for `../seed-data.json` (project root)

2. **Database connection failed**
   - Check PostgreSQL is running
   - Verify environment variables are correct
   - Ensure database exists

3. **Import errors**
   - Run `go mod tidy` in the project root
   - Ensure all modules are properly initialized

### Manual Commands

```bash
# Run from project root
cd /path/to/unburdy_server

# Full seeding
cd seed && go run seed_database.go

# With custom database
DB_NAME=test_db cd seed && go run seed_database.go
```

## Development

### Adding New Seed Data

1. **Base-server data**: Edit `../seed-data.json`
2. **Application data**: Edit `seed_app_data.json`
3. **New entities**: Update `seed_database.go` auto-migration section

### Extending Functionality

The `seed_database.go` file is designed to be extensible:
- Add new entity types in the auto-migration section
- Add new seed data sources
- Extend statistics reporting

## Success Indicators

✅ **Working correctly when you see:**
- "Database migration completed"
- "Successfully created X out of Y [entities]"
- Statistics showing expected counts
- No error messages

❌ **Issues if you see:**
- "Failed to connect to database"
- "Failed to migrate entities"
- Many "Failed to create" messages
- Missing statistics