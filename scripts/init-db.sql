-- Sentinel Hub API Database Initialization
-- Complies with CODING_STANDARDS.md: Database schema standards

-- Create database (if not exists)
-- Note: This is handled by docker-compose environment variables

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    source VARCHAR(255) NOT NULL DEFAULT 'manual',
    title VARCHAR(500) NOT NULL,
    description TEXT,
    file_path VARCHAR(1000),
    line_number INTEGER,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'blocked', 'cancelled')),
    priority VARCHAR(50) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    assigned_to VARCHAR(255),
    estimated_effort INTEGER,
    actual_effort INTEGER,
    tags TEXT[], -- PostgreSQL array type
    verification_confidence FLOAT NOT NULL DEFAULT 0.0 CHECK (verification_confidence >= 0.0 AND verification_confidence <= 1.0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_verification_confidence ON tasks(verification_confidence);

-- LLM Configurations table (for future use)
CREATE TABLE IF NOT EXISTS llm_configurations (
    id SERIAL PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('openai', 'anthropic', 'azure')),
    api_key_encrypted TEXT NOT NULL,
    model VARCHAR(255) NOT NULL,
    key_type VARCHAR(50) NOT NULL DEFAULT 'personal',
    cost_optimization JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id)
);

-- Audit log table (for compliance and debugging)
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for audit logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;
CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_llm_configurations_updated_at ON llm_configurations;
CREATE TRIGGER update_llm_configurations_updated_at BEFORE UPDATE ON llm_configurations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security (RLS) policies (optional, for multi-tenant scenarios)
-- ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
-- CREATE POLICY tasks_user_access ON tasks FOR ALL USING (assigned_to = current_user OR current_user IN ('admin'));

-- Insert default admin user (for development)
-- Password: admin123 (hashed with bcrypt cost 12)
-- In production, this should be created through the API
INSERT INTO users (email, name, password, role, is_active)
VALUES (
    'admin@sentinel.com',
    'System Administrator',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/Le0Kd1sEJYzHfWjK6',
    'admin',
    true
)
ON CONFLICT (email) DO NOTHING;

-- Comments for documentation
COMMENT ON TABLE users IS 'System users with authentication and authorization';
COMMENT ON TABLE tasks IS 'Tracked development tasks with status and assignment';
COMMENT ON TABLE llm_configurations IS 'LLM service configurations per project';
COMMENT ON TABLE audit_logs IS 'Audit trail for compliance and debugging';

-- Grant permissions (adjust as needed for your deployment)
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO sentinel;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO sentinel;