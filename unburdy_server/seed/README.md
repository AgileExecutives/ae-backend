# Unburdy Server Database Seeding

This directory contains consolidated seeding functionality for the Unburdy Server application. It handles base-server entities, application-specific data, documents module (templates, invoices, PDFs), and calendar data.

## Overview

The seeding system automatically:
- **Auto-migrates** all database entities (base-server + modules)
- **Seeds base-server data** (tenants, users, plans) from `seed-data.json`
- **Seeds application data** (cost providers, clients) from `seed_app_data.json`
- **Seeds documents module** (templates, invoice numbers, sample PDFs)
- **Seeds calendar data** (calendars, appointments, recurring events)
- **Provides statistics** and verification of seeded data

## What Gets Seeded

### Base Server Data
- **Tenants**: Test tenant organization
- **Users**: Admin and test users with authentication
- **Plans**: Subscription plans and pricing

### Application Data
- **Cost Providers**: German Jugendämter (youth welfare offices)
- **Clients**: Test therapy clients with contact information

### Documents Module (NEW)
- **Templates**: 
  - Standard therapy invoice template (German)
  - Email template for client communication
- **Invoice Numbers**: Sequential invoice number tracking (INV-YYYY-MM-XXXX format)
- **Sample Documents**: 3 actual invoice PDFs stored in MinIO with:
  - Realistic client data from seeded clients
  - Professional German invoice formatting
  - PDF storage in MinIO with database tracking

### Calendar Module
- **Calendars**: Personal calendars for each user
- **Calendar Entries**: Sample appointments and therapy sessions
- **Recurring Series**: Repeating events
- **Holiday Data**: German public and school holidays

### Client Management Module (NEW)
- **Sessions**: Therapy sessions linked to calendar entries with:
  - **Conducted sessions**: Completed therapy sessions with documentation
  - **Scheduled sessions**: Upcoming therapy appointments
  - **Canceled sessions**: Appointments that were canceled
  - **Date range**: Spans from 5 weeks ago to 10 weeks in the future
  - **~150 sessions** created with realistic distribution
- **Invoices**: Sample invoices with different payment statuses:
  - **Paid invoice**: Invoice marked as paid with payment date
  - **Sent invoice**: Invoice sent but not yet paid
  - **Reminder invoice**: Invoice with payment reminder sent
- **Invoice Items**: Line items linking invoices to therapy sessions
- **Organization**: Default organization created if none exists

## Prerequisites

Before running the seed script, ensure the following services are running:

1. **PostgreSQL** (default: localhost:5432)
2. **Redis** (default: localhost:6379) - Required for invoice number generation
3. **MinIO** (default: localhost:9000) - Required for template and document storage

### Quick Start with Docker Compose

```bash
cd /Users/alex/src/ae/backend/environments/dev
docker-compose up -d postgres redis minio
```

## Files

- `manage_db.sh` - Main management script with multiple commands
- `run_seed.sh` - Simple wrapper script for basic seeding
- `seed_database.go` - Consolidated Go seeding application
- `seed_app_data.json` - Application-specific seed data (1600+ lines)
- `README.md` - This documentation file

## Configuration

### Database Configuration
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=pass
export DB_NAME=ae_saas_basic_test
export DB_SSL_MODE=disable
```

### Redis Configuration
```bash
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=redis123
```

### MinIO Configuration
```bash
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=minioadmin
export MINIO_SECRET_KEY=minioadmin123
export MINIO_USE_SSL=false
export MINIO_REGION=us-east-1
```

## Usage

### Management Script (Recommended)

```bash
# Full seeding (base + application data + documents + calendar)
./manage_db.sh seed

# Test/check database contents
./manage_db.sh test

# Clear all data and re-seed
./manage_db.sh reset

# Show all available commands
./manage_db.sh
```

### Direct Execution

```bash
# Full seeding
go run seed_database.go

# Calendar-only mode (for reseeding calendar data)
SEED_CALENDAR_ONLY=true go run seed_database.go
```

### Custom Database Example

```bash
DB_NAME=my_custom_db ./manage_db.sh seed
```

## What Gets Seeded (Detailed)

### Base-Server Entities (from `../seed-data.json`)
- **Tenants** - Multi-tenancy support (admin, default-tenant)
- **Users** - System users (admin, testuser)
- **Plans** - Subscription plans (Basic, Pro)

### Application Entities (from `seed_app_data.json`)
- **Cost Providers** - 45+ German youth offices (Jugendämter)
- **Clients** - 46+ realistic client records with therapy information

### Documents Module Entities
- **Templates** - Invoice and email templates stored in MinIO
- **Invoice Numbers** - Sequential tracking with Redis caching
- **Documents** - Sample PDF invoices with metadata

### Calendar Module Entities
- **Calendars** - Personal calendars per user
- **Calendar Entries** - Appointments and therapy sessions
- **Recurring Series** - Repeating events
- **Holidays** - German public and school holidays

### Client Management Module Entities (NEW)
- **Sessions** - ~150 therapy sessions spanning 5 weeks back to 10 weeks ahead
  - Conducted sessions (past) with documentation
  - Scheduled sessions (future) for upcoming appointments
  - Some canceled sessions distributed throughout
  - All sessions linked to calendar entries and clients
- **Invoices** - 3 invoices with different statuses
  - 1 paid invoice (marked with payment date)
  - 1 sent invoice (awaiting payment)
  - 1 reminder invoice (payment reminder sent)
- **Invoice Items** - ~12 line items linking sessions to invoices
- **Organizations** - Default organization auto-created if needed

### Auto-Migration
- All base-server entities (users, tenants, customers, emails, etc.)
- All module entities (clients, cost_providers, calendar tables, documents, templates, etc.)

## Output Example

```
🌱 Unburdy Server - Complete Database Seeding
===========================================
🔗 Connecting to PostgreSQL: localhost:5432/ae_saas_basic_test
🏗️  Auto-migrating database entities...
✅ Database migration completed
🌱 Seeding base-server data (tenants, users, plans)...
✅ Base data seeding completed
🌱 Seeding application data (cost providers, clients)...
✅ Created 45 cost providers
✅ Created 46 clients

📄 Seeding documents module data (templates, invoice numbers, documents)...
📝 Creating invoice template...
✅ Created invoice template (ID: 1)
📧 Creating email template...
✅ Created email template (ID: 2)
🔢 Generating sample invoice numbers...
  ✓ Generated: INV-2025-12-0001
  ✓ Generated: INV-2025-12-0002
  ✓ Generated: INV-2025-12-0003
✅ Generated 5 invoice numbers
📄 Generating sample invoice PDFs...
  ✓ Generated invoice PDF: INV-2025-12-0001 (Document ID: 1, Size: 45678 bytes)
  ✓ Generated invoice PDF: INV-2025-12-0002 (Document ID: 2, Size: 46123 bytes)
  ✓ Generated invoice PDF: INV-2025-12-0003 (Document ID: 3, Size: 45891 bytes)
✅ Documents module seeding completed

📊 Documents Module Statistics
==============================
📝 Total templates: 2
📄 Total documents: 3
🔢 Total invoice numbers: 5

🗓️  Seeding calendar data...
✅ Successfully seeded calendar data for user 1

📅 Calendar Seeding Statistics
==============================
📋 Total calendars: 8
📅 Total calendar entries: 1280
🔄 Total recurring series: 8

📅 Seeding client sessions...
  📆 Creating sessions from 2025-11-22 to 2026-03-07 (current time: 2025-12-27 07:57 UTC)
  📋 Found 150 calendar entries in date range
  📊 Progress: 10 sessions created...
  📊 Progress: 20 sessions created...
  ... (progress continues)
  📊 Progress: 150 sessions created...
✅ Created 150 sessions (66 conducted, 73 scheduled, 11 canceled)

💰 Seeding invoices...
  📋 Creating default organization...
  ✅ Created default organization (ID: 1)
  ✓ Created invoice INV-1-2025-1000 - payed (paid 2025-12-22) - 4 sessions - €714.00
  ✓ Created invoice INV-1-2025-1001 - sent - 4 sessions - €714.00
  ✓ Created invoice INV-1-2025-1002 - reminder - 4 sessions - €714.00
✅ Created 3 invoices

📊 Session & Invoice Statistics
================================
📅 Total sessions: 150
🎯 Session Status Breakdown:
   scheduled: 73 sessions
   conducted: 66 sessions
   canceled: 11 sessions
💰 Total invoices: 3
💳 Invoice Status Breakdown:
   payed: 1 invoices
   sent: 1 invoices
   reminder: 1 invoices
📋 Total invoice items: 12

✨ Complete database seeding finished successfully!
```

## Accessing Seeded Data

### View Sessions

```bash
# List all sessions
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/sessions

# Get session details
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/sessions/1
```

### View Invoices

```bash
# List all invoices
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/invoices

# Get invoice details with items
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/invoices/1
```

### View Templates in MinIO

1. Open MinIO Console: http://localhost:9001
2. Login with credentials (minioadmin / minioadmin123)
3. Navigate to bucket: `templates`
4. You'll find invoice and email templates

### View Documents (PDFs) in MinIO

1. Open MinIO Console: http://localhost:9001
2. Navigate to bucket: `documents`
3. You'll find generated invoice PDFs organized by tenant

### Download Invoice PDFs via API

```bash
# Get all documents
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/documents

# Download a specific document
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/documents/1/download
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