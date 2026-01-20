-- Migration 002: Create core tables for organizations, projects, tasks, documents, knowledge items, and LLM configurations
-- This migration adds the foundational tables required by the application

-- Organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(255) PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL REFERENCES organizations(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    repository_url TEXT,
    default_branch VARCHAR(255) DEFAULT 'main',
    api_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tasks table (matching helpers.go query structure)
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    source VARCHAR(100),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    file_path TEXT,
    line_number INTEGER,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    priority VARCHAR(50) DEFAULT 'medium',
    assigned_to VARCHAR(255),
    estimated_effort INTEGER,
    actual_effort INTEGER,
    verification_confidence FLOAT DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1
);

-- Documents table (matching models/document.go)
CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'uploaded',
    progress INTEGER NOT NULL DEFAULT 0,
    extracted_text TEXT,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP WITH TIME ZONE
);

-- Knowledge items table
CREATE TABLE IF NOT EXISTS knowledge_items (
    id VARCHAR(255) PRIMARY KEY,
    document_id VARCHAR(255) REFERENCES documents(id),
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 0.0,
    source_page INTEGER,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    structured_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- LLM configurations table (matching llm/config.go)
CREATE TABLE IF NOT EXISTS llm_configurations (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    provider VARCHAR(50) NOT NULL,
    api_key_encrypted BYTEA NOT NULL,
    model VARCHAR(100) NOT NULL,
    key_type VARCHAR(50),
    cost_optimization JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for foreign keys and frequently queried columns
CREATE INDEX IF NOT EXISTS idx_projects_org ON projects(organization_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_documents_project ON documents(project_id);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_knowledge_project ON knowledge_items(project_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_document ON knowledge_items(document_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_type ON knowledge_items(type);
CREATE INDEX IF NOT EXISTS idx_llm_configs_project ON llm_configurations(project_id);
