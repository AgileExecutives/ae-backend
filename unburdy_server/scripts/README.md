# Database Seeding

This directory contains scripts to populate the unburdy database with realistic client data.

## Files

- `seed_app_data.json` - Contains 50 realistic client records with comprehensive therapy information
- `seed_database.go` - Go script that reads the JSON and populates the database
- `run_seed.sh` - Shell script wrapper for easy execution

## Usage

### Option 1: Using the shell script (recommended)
```bash
# From unburdy_server directory
./scripts/run_seed.sh
```

### Option 2: Running Go script directly
```bash
# From scripts directory
cd scripts
go run seed_database.go
```

## What it does

The seeding script:
1. Connects to the SQLite database (`unburdy.db`)
2. Auto-migrates the Client model schema
3. Clears existing client data (optional - can be commented out)
4. Creates 46 realistic client records
5. Shows statistics about:
   - Total clients created
   - Breakdown by therapy type
   - Breakdown by status

## Data Overview

The seed data includes diverse clients with:
- **Demographics**: Various ages, genders, locations (primarily Illinois/Wisconsin)
- **Contact Info**: Multiple contact methods, alternative contacts
- **Therapy Details**: 30+ different therapy types including CBT, EMDR, DBT, etc.
- **Insurance**: Various insurance providers
- **Session Rates**: Range from $110-$150
- **Comprehensive Fields**: All 25+ client model fields populated

## Therapy Types Included

- CBT (Cognitive Behavioral Therapy)
- EMDR (Eye Movement Desensitization and Reprocessing)  
- DBT (Dialectical Behavior Therapy)
- Solution-Focused Therapy
- Executive Coaching
- Family Therapy
- Couples Therapy
- Trauma-Informed Care
- And many more...