-- Migration 004: Create task dependencies table
-- This migration adds the task_dependencies table for persisting task dependency relationships

CREATE TABLE IF NOT EXISTS task_dependencies (
    id VARCHAR(255) PRIMARY KEY,
    task_id VARCHAR(255) NOT NULL REFERENCES tasks(id),
    depends_on_task_id VARCHAR(255) NOT NULL REFERENCES tasks(id),
    dependency_type VARCHAR(50) NOT NULL DEFAULT 'blocks',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_task_dependency UNIQUE (task_id, depends_on_task_id)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_task_deps_task ON task_dependencies(task_id);
CREATE INDEX IF NOT EXISTS idx_task_deps_depends_on ON task_dependencies(depends_on_task_id);
CREATE INDEX IF NOT EXISTS idx_task_deps_type ON task_dependencies(dependency_type);
