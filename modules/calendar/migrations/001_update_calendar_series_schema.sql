-- Migration: Update calendar_series schema
-- Remove old weekday and interval columns, add new interval_type, interval_value, and last_date columns

-- Step 1: Add new columns (nullable first to avoid constraint violations)
ALTER TABLE calendar_series 
ADD COLUMN IF NOT EXISTS interval_type VARCHAR(20),
ADD COLUMN IF NOT EXISTS interval_value INTEGER,
ADD COLUMN IF NOT EXISTS last_date TIMESTAMP WITH TIME ZONE;

-- Step 2: Set default values for existing records
UPDATE calendar_series 
SET interval_type = 'weekly',
    interval_value = 1
WHERE interval_type IS NULL;

-- Step 3: Drop old columns
ALTER TABLE calendar_series 
DROP COLUMN IF EXISTS weekday,
DROP COLUMN IF EXISTS interval;

-- Step 4: Add position_in_series to calendar_entries if it doesn't exist
ALTER TABLE calendar_entries
ADD COLUMN IF NOT EXISTS position_in_series INTEGER;

-- Verify changes
SELECT 'Migration completed successfully' AS status;
