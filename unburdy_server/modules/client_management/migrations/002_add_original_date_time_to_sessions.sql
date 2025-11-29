-- Migration: Add original date and start time columns to sessions table
-- Purpose: Preserve original appointment timing even after calendar entry deletion
-- Allows historical reporting and tracking of when sessions were originally scheduled

-- Add original_date column (stores UTC date of the calendar entry)
ALTER TABLE sessions 
ADD COLUMN IF NOT EXISTS original_date TIMESTAMP WITH TIME ZONE;

-- Add original_start_time column (stores UTC start time from the calendar entry)
ALTER TABLE sessions 
ADD COLUMN IF NOT EXISTS original_start_time TIMESTAMP WITH TIME ZONE;

-- Backfill existing sessions with data from their calendar entries
-- This populates the new columns for sessions that still have a valid calendar entry
UPDATE sessions s
SET 
    original_date = DATE_TRUNC('day', ce.start_time AT TIME ZONE 'UTC'),
    original_start_time = ce.start_time AT TIME ZONE 'UTC'
FROM calendar_entries ce
WHERE s.calendar_entry_id = ce.id 
  AND s.calendar_entry_id > 0
  AND s.original_date IS NULL;

-- For sessions with deleted or missing calendar entries, use created_at as fallback
-- This handles edge cases where calendar_entry_id = 0 or entry was already deleted
UPDATE sessions
SET 
    original_date = DATE_TRUNC('day', created_at AT TIME ZONE 'UTC'),
    original_start_time = created_at AT TIME ZONE 'UTC'
WHERE original_date IS NULL;

-- Make columns NOT NULL after backfilling all existing data
ALTER TABLE sessions 
ALTER COLUMN original_date SET NOT NULL;

ALTER TABLE sessions 
ALTER COLUMN original_start_time SET NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN sessions.original_date IS 'UTC date of the original calendar entry (preserved even after entry deletion)';
COMMENT ON COLUMN sessions.original_start_time IS 'UTC start time from the original calendar entry (preserved even after entry deletion)';
