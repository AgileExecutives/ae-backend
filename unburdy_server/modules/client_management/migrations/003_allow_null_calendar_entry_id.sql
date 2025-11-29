-- Migration: Allow NULL for calendar_entry_id in sessions table
-- Purpose: When a calendar entry is deleted, sessions should remain with NULL calendar_entry_id
-- This preserves the session history while unlinking from deleted calendar entries

-- Drop the foreign key constraint first (if it exists)
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS fk_sessions_calendar_entry;

-- Make calendar_entry_id nullable
ALTER TABLE sessions ALTER COLUMN calendar_entry_id DROP NOT NULL;

-- Re-add the foreign key constraint with ON DELETE SET NULL
-- This automatically sets calendar_entry_id to NULL when the calendar entry is deleted
ALTER TABLE sessions 
ADD CONSTRAINT fk_sessions_calendar_entry 
FOREIGN KEY (calendar_entry_id) 
REFERENCES calendar_entries(id) 
ON DELETE SET NULL;

-- Add comment
COMMENT ON COLUMN sessions.calendar_entry_id IS 'Reference to calendar entry (nullable - set to NULL when entry is deleted)';
