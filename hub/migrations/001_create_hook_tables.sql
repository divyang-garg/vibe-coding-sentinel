CREATE TABLE IF NOT EXISTS hook_executions (
    id UUID PRIMARY KEY,
    agent_id VARCHAR(64),
    org_id UUID,
    team_id UUID,
    hook_type VARCHAR(20) NOT NULL,
    result VARCHAR(20) NOT NULL,
    override_reason VARCHAR(255),
    findings_summary JSONB,
    user_actions JSONB,
    duration_ms BIGINT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hook_baselines (
    id UUID PRIMARY KEY,
    agent_id VARCHAR(64),
    org_id UUID,
    team_id UUID,
    hook_type VARCHAR(20) NOT NULL,
    file TEXT NOT NULL,
    line INTEGER NOT NULL,
    pattern TEXT,
    message TEXT,
    severity VARCHAR(20),
    source VARCHAR(20) NOT NULL,
    reviewed BOOLEAN DEFAULT false,
    reviewed_by UUID,
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hook_policies (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    policy_config JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hook_executions_agent ON hook_executions(agent_id);
CREATE INDEX IF NOT EXISTS idx_hook_executions_org ON hook_executions(org_id);
CREATE INDEX IF NOT EXISTS idx_hook_executions_created ON hook_executions(created_at);
CREATE INDEX IF NOT EXISTS idx_hook_baselines_org ON hook_baselines(org_id);
CREATE INDEX IF NOT EXISTS idx_hook_baselines_reviewed ON hook_baselines(reviewed);
CREATE INDEX IF NOT EXISTS idx_hook_policies_org ON hook_policies(org_id);
