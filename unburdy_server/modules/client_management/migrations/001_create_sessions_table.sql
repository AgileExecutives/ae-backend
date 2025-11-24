-- Migration: Create sessions table
-- Purpose: Store therapy sessions linked to calendar entries and clients

-- Drop table if exists (for idempotent migrations)
DROP TABLE IF EXISTS sessions CASCADE;

-- Create sessions table
CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    client_id BIGINT NOT NULL,
    calendar_entry_id BIGINT NOT NULL,
    duration_min INTEGER NOT NULL,
    type VARCHAR(255) NOT NULL,
    number_units INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'scheduled',
    documentation TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for performance
CREATE INDEX idx_session_tenant ON sessions(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_session_client ON sessions(client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_session_calendar ON sessions(calendar_entry_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_session_status ON sessions(status) WHERE deleted_at IS NULL;

-- Add foreign key constraints
ALTER TABLE sessions 
    ADD CONSTRAINT fk_sessions_calendar_entry 
    FOREIGN KEY (calendar_entry_id) 
    REFERENCES calendar_entries(id) 
    ON DELETE CASCADE;

-- Add check constraint for status values
ALTER TABLE sessions 
    ADD CONSTRAINT check_session_status 
    CHECK (status IN ('scheduled', 'canceled', 'conducted'));

-- Add comment
COMMENT ON TABLE sessions IS 'Therapy sessions linked to calendar entries and clients';
