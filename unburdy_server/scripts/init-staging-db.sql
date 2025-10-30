-- Initialize Unburdy Staging Database
-- This script runs automatically when the PostgreSQL container starts

-- Create extensions that might be needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set timezone
SET timezone TO 'UTC';

-- Create additional schemas if needed
-- CREATE SCHEMA IF NOT EXISTS analytics;
-- CREATE SCHEMA IF NOT EXISTS reporting;

-- Set connection limits and other database-level configurations
ALTER DATABASE unburdy_staging SET log_statement = 'all';
ALTER DATABASE unburdy_staging SET log_min_duration_statement = 1000;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE unburdy_staging TO unburdy_user;

-- Log initialization
INSERT INTO information_schema.sql_features (feature_id, feature_name, sub_feature_id, sub_feature_name, is_supported, comments) 
VALUES ('STAGING_INIT', 'Database initialized for staging environment', NULL, NULL, 'YES', 'Initialized at ' || CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;