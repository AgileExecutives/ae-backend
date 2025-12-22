-- Migration: Create booking token usage tracking table
-- This table tracks usage statistics for booking link tokens to enforce usage limits

CREATE TABLE IF NOT EXISTS booking_token_usage (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    token_id VARCHAR(64) NOT NULL UNIQUE,
    tenant_id BIGINT NOT NULL,
    template_id BIGINT NOT NULL,
    client_id BIGINT NOT NULL,
    use_count INTEGER NOT NULL DEFAULT 0,
    max_use_count INTEGER NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_booking_token_usage_token_id ON booking_token_usage(token_id);
CREATE INDEX IF NOT EXISTS idx_booking_token_usage_tenant_id ON booking_token_usage(tenant_id);
CREATE INDEX IF NOT EXISTS idx_booking_token_usage_template_id ON booking_token_usage(template_id);
CREATE INDEX IF NOT EXISTS idx_booking_token_usage_client_id ON booking_token_usage(client_id);
CREATE INDEX IF NOT EXISTS idx_booking_token_usage_deleted_at ON booking_token_usage(deleted_at);
CREATE INDEX IF NOT EXISTS idx_booking_token_usage_expires_at ON booking_token_usage(expires_at);

-- Add comment to the table
COMMENT ON TABLE booking_token_usage IS 'Tracks usage statistics for booking link tokens to enforce usage limits and expiration';
COMMENT ON COLUMN booking_token_usage.token_id IS 'SHA256 hash of the booking link token';
COMMENT ON COLUMN booking_token_usage.use_count IS 'Number of times the token has been used';
COMMENT ON COLUMN booking_token_usage.max_use_count IS 'Maximum allowed uses (0 means unlimited)';
COMMENT ON COLUMN booking_token_usage.last_used_at IS 'Timestamp of the last usage';
COMMENT ON COLUMN booking_token_usage.expires_at IS 'Token expiration timestamp (NULL means no expiration)';
