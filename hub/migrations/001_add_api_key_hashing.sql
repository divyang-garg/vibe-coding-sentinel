-- Migration: Add API Key Hashing Support
-- Description: Adds api_key_hash and api_key_prefix columns to projects table
-- Date: 2026-01-20
-- Compliance: Security Remediation Plan Phase 1.4

-- Add new columns for secure API key storage
ALTER TABLE projects 
ADD COLUMN IF NOT EXISTS api_key_hash VARCHAR(64),
ADD COLUMN IF NOT EXISTS api_key_prefix VARCHAR(8);

-- Migrate existing data: hash existing plaintext API keys
-- This uses PostgreSQL's pgcrypto extension for SHA-256 hashing
-- Note: Enable pgcrypto extension first if not already enabled: CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- For now, this will only run if pgcrypto is enabled and there's existing data to migrate
DO $$
BEGIN
    -- Only attempt migration if pgcrypto extension exists and there's data to migrate
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto') THEN
        UPDATE projects 
        SET api_key_hash = encode(digest(api_key, 'sha256'), 'hex'),
            api_key_prefix = LEFT(api_key, 8)
        WHERE api_key IS NOT NULL 
          AND api_key != '' 
          AND (api_key_hash IS NULL OR api_key_hash = '');
    END IF;
END $$;

-- Create index for faster hash lookups (critical for authentication performance)
CREATE INDEX IF NOT EXISTS idx_projects_api_key_hash ON projects(api_key_hash);

-- Create index for prefix lookups (for identification/filtering)
CREATE INDEX IF NOT EXISTS idx_projects_api_key_prefix ON projects(api_key_prefix);

-- Add comments for documentation
COMMENT ON COLUMN projects.api_key_hash IS 'SHA-256 hash of API key (hex-encoded) for secure storage';
COMMENT ON COLUMN projects.api_key_prefix IS 'First 8 characters of API key for identification without compromising security';

-- Note: api_key column kept for backward compatibility during migration period
-- After all keys are migrated and verified, consider:
-- ALTER TABLE projects DROP COLUMN api_key;
-- This should only be done after confirming all systems use hash-based validation
