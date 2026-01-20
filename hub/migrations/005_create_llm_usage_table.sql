-- Migration: Create LLM usage tracking table
-- Purpose: Track LLM API usage for cost monitoring and quota management
-- Created: 2026-01-20

CREATE TABLE IF NOT EXISTS llm_usage (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    validation_id VARCHAR(255),
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    tokens_used INTEGER NOT NULL DEFAULT 0,
    estimated_cost DECIMAL(10, 6) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for project-based queries
CREATE INDEX IF NOT EXISTS idx_llm_usage_project_id ON llm_usage(project_id);

-- Create index for validation-based queries
CREATE INDEX IF NOT EXISTS idx_llm_usage_validation_id ON llm_usage(validation_id) WHERE validation_id IS NOT NULL;

-- Create index for time-based queries
CREATE INDEX IF NOT EXISTS idx_llm_usage_created_at ON llm_usage(created_at DESC);

-- Add comment to table
COMMENT ON TABLE llm_usage IS 'Tracks LLM API usage for cost monitoring and quota management';
