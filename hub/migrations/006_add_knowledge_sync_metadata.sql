-- Migration 006: Add sync metadata columns to knowledge_items table
-- This migration adds columns for tracking sync status, timestamps, and approval metadata
-- Complies with CODING_STANDARDS.md: Database migrations are additive only

-- Add missing columns to knowledge_items table
ALTER TABLE knowledge_items 
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS approved_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS sync_version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS sync_status VARCHAR(50) DEFAULT 'pending';

-- Create indexes for performance (compliant with CODING_STANDARDS.md performance requirements)
CREATE INDEX IF NOT EXISTS idx_knowledge_status ON knowledge_items(status);
CREATE INDEX IF NOT EXISTS idx_knowledge_type_status ON knowledge_items(item_type, status);
CREATE INDEX IF NOT EXISTS idx_knowledge_sync_status ON knowledge_items(sync_status);
CREATE INDEX IF NOT EXISTS idx_knowledge_last_synced ON knowledge_items(last_synced_at);

-- Update existing rows to set updated_at to created_at if not set
UPDATE knowledge_items 
SET updated_at = created_at 
WHERE updated_at IS NULL;

-- Set default sync_status for existing rows
UPDATE knowledge_items 
SET sync_status = 'pending' 
WHERE sync_status IS NULL;
