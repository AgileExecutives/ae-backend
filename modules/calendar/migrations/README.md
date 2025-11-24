# Calendar Module Database Migration

## Migration: Update Calendar Series Schema

This migration updates the `calendar_series` and `calendar_entries` tables to support the new flexible recurrence system.

### Changes

**calendar_series table:**
- ❌ Removes: `weekday` (int) - old weekday field
- ❌ Removes: `interval` (int) - old interval field  
- ✅ Adds: `interval_type` (varchar) - type of recurrence: "none", "weekly", "monthly-date", "monthly-day", "yearly"
- ✅ Adds: `interval_value` (int) - number of intervals between occurrences
- ✅ Adds: `last_date` (timestamp) - end condition for recurring events

**calendar_entries table:**
- ✅ Adds: `position_in_series` (int) - position of entry in series (1, 2, 3, ...)

### Running the Migration

#### Option 1: Using the script (recommended)

```bash
cd /Users/alex/src/ae/backend/modules/calendar/migrations
./run_migration.sh
```

#### Option 2: Manual execution

```bash
psql -h localhost -p 5432 -U postgres -d unburdy_db -f 001_update_calendar_series_schema.sql
```

#### Option 3: Copy/paste SQL

If you prefer to run the migration through a database client, the SQL is in:
```
modules/calendar/migrations/001_update_calendar_series_schema.sql
```

### Notes

- Existing `calendar_series` records will be set to `interval_type='weekly'` and `interval_value=1` as defaults
- The migration is idempotent (safe to run multiple times)
- All `IF EXISTS` / `IF NOT EXISTS` clauses ensure compatibility with existing schemas
