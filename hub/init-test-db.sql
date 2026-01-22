-- Initialize test database with proper users and permissions
-- This script runs when the test database container starts

-- Create role if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'sentinel') THEN
        CREATE ROLE sentinel WITH LOGIN PASSWORD 'sentinel';
    END IF;
END
$$;

-- Grant privileges
ALTER ROLE sentinel CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE sentinel_test TO sentinel;

-- Ensure we're using the correct database
\c sentinel_test

-- Grant schema privileges
GRANT ALL ON SCHEMA public TO sentinel;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO sentinel;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO sentinel;




