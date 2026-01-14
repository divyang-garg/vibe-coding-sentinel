# Task Dependency & Verification System

> **For AI Agents**: This document specifies the complete solution for task dependency tracking and verification in Sentinel. This addresses the critical gap where Cursor-generated tasks are not tracked, verified, or completed.

## Executive Summary

**Problem**: Cursor-generated tasks (TODO comments, task markers) are often left incomplete, creating technical debt and incomplete features. There's no systematic way to track task completion, verify implementation, or manage dependencies between tasks.

**Solution**: A comprehensive task dependency and verification system that:
- Detects Cursor-generated tasks automatically
- Verifies task completion using multi-factor verification
- Tracks dependencies between tasks (explicit, implicit, integration, feature-level)
- Auto-marks completed tasks or alerts on incomplete ones
- Integrates deeply with existing Sentinel systems

**Key Innovation**: Multi-factor verification with dependency management, integrated with comprehensive feature analysis for production-grade projects.

---

## 1. Problem Statement

### Current Limitations

1. **No Task Tracking**: Cursor generates tasks but they're not systematically tracked
2. **No Completion Verification**: No way to verify if tasks are actually completed
3. **No Dependency Management**: Tasks that depend on other tasks aren't identified
4. **No Integration Awareness**: Tasks requiring external integrations aren't tracked
5. **No Feature-Level Context**: Tasks part of larger features aren't linked
6. **Manual Process**: Developers must manually track and verify tasks

### Real-World Impact

- **Incomplete Features**: Tasks left unfinished create incomplete features
- **Technical Debt**: Uncompleted tasks accumulate over time
- **Production Issues**: Missing dependencies cause production failures
- **Integration Failures**: Integration tasks missed cause system failures
- **Feature Gaps**: Feature-level tasks incomplete create gaps in functionality

---

## 2. Solution Overview

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    TASK DEPENDENCY SYSTEM FLOW                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Codebase Scanning                                                │
│      │                                                            │
│      ▼                                                            │
│  Task Detection (TODO, markers, Cursor format)                    │
│      │                                                            │
│      ├── Extract task metadata (title, description, file, line)   │
│      ├── Identify task source (cursor, manual, change_request)   │
│      └── Store in database                                        │
│      │                                                            │
│      ▼                                                            │
│  Dependency Detection                                             │
│      │                                                            │
│      ├── Explicit (from task descriptions)                       │
│      ├── Implicit (code analysis)                                │
│      ├── Integration (external APIs/services)                    │
│      └── Feature-level (Phase 14A comprehensive analysis)         │
│      │                                                            │
│      ▼                                                            │
│  Task Verification (Multi-Factor)                                 │
│      │                                                            │
│      ├── Code Existence (AST search)                             │
│      ├── Code Usage (cross-file references)                       │
│      ├── Test Coverage (test file existence)                      │
│      └── Integration (external service verification)              │
│      │                                                            │
│      ▼                                                            │
│  Auto-Completion & Alerts                                         │
│      │                                                            │
│      ├── High confidence (>0.8) → Auto-complete                 │
│      ├── Critical tasks incomplete → Alert                        │
│      └── Dependency blocking → Alert                             │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Key Components

1. **Task Detection Engine**: Scans codebase for tasks (TODO comments, markers)
2. **Dependency Analyzer**: Detects explicit, implicit, integration, and feature-level dependencies
3. **Verification Engine**: Multi-factor verification (code, usage, tests, integration)
4. **Integration Layer**: Links tasks to change requests, knowledge items, comprehensive analysis
5. **Auto-Completion System**: Automatically marks completed tasks based on verification
6. **Alert System**: Notifies on incomplete critical tasks or dependency blocks

---

## 3. Task Detection

### Detection Methods

**1. TODO Comments**:
```javascript
// TODO: Implement user authentication
// FIXME: Fix memory leak in cache
// NOTE: Add error handling here
```

**2. Cursor Task Markers**:
```markdown
- [ ] Task: Implement payment processing
- [x] Task: Add user authentication (completed)
```

**3. Explicit Task Format**:
```javascript
// TASK: TASK-123 - Add order cancellation
// DEPENDS: TASK-122, TASK-121
```

**4. Change Request Tasks**:
Tasks automatically created from approved change requests (Phase 12)

**5. Comprehensive Analysis Tasks**:
Tasks discovered during comprehensive feature analysis (Phase 14A)

### Task Metadata

```json
{
  "id": "task_abc123",
  "project_id": "proj_xyz",
  "source": "cursor",
  "title": "Implement user authentication",
  "description": "Add JWT-based authentication middleware",
  "file_path": "src/auth/middleware.js",
  "line_number": 45,
  "status": "pending",
  "priority": "high",
  "created_at": "2024-12-10T10:00:00Z",
  "updated_at": "2024-12-10T10:00:00Z"
}
```

---

## 4. Dependency Detection

### Dependency Types

**1. Explicit Dependencies**:
Parsed from task descriptions:
```
TASK: Add payment processing
DEPENDS: TASK-123 (API setup), TASK-124 (Database schema)
```

**2. Implicit Dependencies**:
Detected through code analysis:
- Task A calls function from Task B → Task A depends on Task B
- Task A imports module from Task B → Task A depends on Task B

**3. Integration Dependencies**:
Detected through comprehensive analysis (Phase 14A):
- Task requires external API setup
- Task requires service configuration
- Task requires third-party library integration

**4. Feature-Level Dependencies**:
Detected through comprehensive feature analysis (Phase 14A):
- Task is part of larger feature
- Feature has multiple tasks with dependencies
- Feature completion requires all tasks complete

### Dependency Graph

```
TASK-001 (User Auth)
    │
    ├── TASK-002 (JWT Middleware) [explicit]
    ├── TASK-003 (Database Schema) [implicit]
    └── TASK-004 (API Endpoints) [feature-level]
        │
        └── TASK-005 (Integration Tests) [explicit]
```

---

## 5. Task Verification

### Multi-Factor Verification

**1. Code Existence Verification**:
- AST search for function/class/feature mentioned in task
- Pattern matching for task keywords
- Confidence score based on match quality

**2. Code Usage Verification**:
- Cross-file reference analysis
- Function call tracking
- Import/export analysis

**3. Test Coverage Verification**:
- Test file existence check
- Test coverage analysis (Phase 10)
- Test quality validation

**4. Integration Verification**:
- External API/service integration check
- Configuration file verification
- Service availability check

### Verification Scoring

```json
{
  "task_id": "task_abc123",
  "verifications": [
    {
      "type": "code_existence",
      "status": "verified",
      "confidence": 0.95,
      "evidence": {
        "files": ["src/auth/middleware.js"],
        "functions": ["authenticateUser"],
        "line_numbers": [45, 67, 89]
      }
    },
    {
      "type": "code_usage",
      "status": "verified",
      "confidence": 0.88,
      "evidence": {
        "call_sites": ["src/routes/users.js:23", "src/routes/orders.js:45"]
      }
    },
    {
      "type": "test_coverage",
      "status": "verified",
      "confidence": 0.92,
      "evidence": {
        "test_file": "tests/auth/middleware.test.js",
        "coverage": 0.95
      }
    },
    {
      "type": "integration",
      "status": "pending",
      "confidence": 0.0,
      "evidence": {}
    }
  ],
  "overall_confidence": 0.69,
  "status": "in_progress"
}
```

### Auto-Completion Threshold

- **High Confidence (>0.8)**: Auto-mark as completed
- **Medium Confidence (0.5-0.8)**: Mark as in_progress, alert developer
- **Low Confidence (<0.5)**: Keep as pending, require manual verification

---

## 6. Integration with Existing Systems

### Phase 11 (Doc-Sync)

- Reuse `detectBusinessRuleImplementation()` pattern for task verification
- Link tasks to documentation status markers
- Sync task status with doc-sync reports

### Phase 12 (Change Requests)

- Link tasks to change requests
- Use implementation tracking for task verification
- Sync task completion with change request status
- Auto-create tasks from approved change requests

### Phase 14A (Comprehensive Analysis)

- Use feature discovery for feature-level dependencies
- Use layer analysis for integration dependencies
- Link tasks to comprehensive analysis results
- Discover tasks during comprehensive feature analysis

### Phase 10 (Test Enforcement)

- Verify test-related tasks
- Link tasks to test requirements
- Use test coverage for verification
- Generate tasks from missing test requirements

### Phase 4 (Knowledge Base)

- Link tasks to business rules
- Use business rules for task context
- Verify tasks against business rule implementation

---

## 7. API Endpoints

### Task Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tasks` | Create or update task |
| GET | `/api/v1/tasks` | List tasks with filters |
| GET | `/api/v1/tasks/{id}` | Get task details |
| PUT | `/api/v1/tasks/{id}` | Update task |
| DELETE | `/api/v1/tasks/{id}` | Delete task |

### Task Verification

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/tasks/{id}/verify` | Verify task completion |
| GET | `/api/v1/tasks/{id}/verifications` | Get verification results |
| POST | `/api/v1/tasks/verify-all` | Verify all pending tasks |

### Dependency Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tasks/{id}/dependencies` | Get task dependencies |
| POST | `/api/v1/tasks/{id}/dependencies` | Add dependency |
| DELETE | `/api/v1/tasks/{id}/dependencies/{dep_id}` | Remove dependency |
| GET | `/api/v1/tasks/{id}/dependency-graph` | Get dependency graph |

### Task Status

| Method | Endpoint | Description |
|--------|----------|-------------|
| PUT | `/api/v1/tasks/{id}/status` | Update task status |
| POST | `/api/v1/tasks/{id}/complete` | Mark task as complete |
| POST | `/api/v1/tasks/{id}/block` | Mark task as blocked |

---

## 8. Agent Commands

### Task Scanning

```bash
# Scan codebase for tasks
sentinel tasks scan

# Scan specific directory
sentinel tasks scan --dir src/

# Scan with filters
sentinel tasks scan --source cursor --status pending
```

### Task Listing

```bash
# List all tasks
sentinel tasks list

# List with filters
sentinel tasks list --status pending --priority high

# List dependencies
sentinel tasks list --show-dependencies
```

### Task Verification

```bash
# Verify specific task
sentinel tasks verify TASK-123

# Verify all pending tasks
sentinel tasks verify --all

# Verify with force (ignore cache)
sentinel tasks verify TASK-123 --force
```

### Dependency Management

```bash
# Show dependency graph
sentinel tasks dependencies

# Show dependencies for specific task
sentinel tasks dependencies TASK-123

# Export dependency graph
sentinel tasks dependencies --export graph.json
```

### Task Completion

```bash
# Manually mark task complete
sentinel tasks complete TASK-123

# Mark with reason
sentinel tasks complete TASK-123 --reason "Implemented manually"

# Auto-complete verified tasks
sentinel tasks complete --auto
```

---

## 9. MCP Integration

### MCP Tools

**1. `sentinel_get_task_status`**:
```json
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "task_id": "TASK-123"
  }
}
```

**Response**:
```json
{
  "task_id": "TASK-123",
  "status": "in_progress",
  "verification": {
    "overall_confidence": 0.75,
    "factors": {
      "code_existence": 0.95,
      "code_usage": 0.88,
      "test_coverage": 0.92,
      "integration": 0.0
    }
  },
  "dependencies": {
    "blocking": ["TASK-122"],
    "blocked_by": []
  }
}
```

**2. `sentinel_verify_task`**:
```json
{
  "name": "sentinel_verify_task",
  "arguments": {
    "task_id": "TASK-123",
    "force": false
  }
}
```

**3. `sentinel_list_tasks`**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "status": "pending",
    "priority": "high",
    "limit": 10
  }
}
```

---

## 10. Database Schema

### Tasks Table

```sql
CREATE TABLE tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  source VARCHAR(50) NOT NULL, -- 'cursor', 'manual', 'change_request', 'comprehensive_analysis'
  title TEXT NOT NULL,
  description TEXT,
  file_path VARCHAR(500),
  line_number INTEGER,
  status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'blocked'
  priority VARCHAR(10) DEFAULT 'medium', -- 'low', 'medium', 'high', 'critical'
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  completed_at TIMESTAMP,
  verified_at TIMESTAMP,
  verification_confidence FLOAT DEFAULT 0.0,
  CONSTRAINT valid_status CHECK (status IN ('pending', 'in_progress', 'completed', 'blocked')),
  CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'critical'))
);

CREATE INDEX idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX idx_tasks_project_priority ON tasks(project_id, priority);
CREATE INDEX idx_tasks_file_path ON tasks(file_path);
```

### Task Dependencies Table

```sql
CREATE TABLE task_dependencies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  depends_on_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  dependency_type VARCHAR(20) NOT NULL, -- 'explicit', 'implicit', 'integration', 'feature'
  confidence FLOAT DEFAULT 0.0, -- 0.0-1.0
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT valid_dependency_type CHECK (dependency_type IN ('explicit', 'implicit', 'integration', 'feature')),
  CONSTRAINT no_self_dependency CHECK (task_id != depends_on_task_id),
  CONSTRAINT unique_dependency UNIQUE (task_id, depends_on_task_id)
);

CREATE INDEX idx_task_dependencies_task ON task_dependencies(task_id);
CREATE INDEX idx_task_dependencies_depends_on ON task_dependencies(depends_on_task_id);
```

### Task Verifications Table

```sql
CREATE TABLE task_verifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  verification_type VARCHAR(20) NOT NULL, -- 'code_existence', 'code_usage', 'test_coverage', 'integration'
  status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'verified', 'failed'
  confidence FLOAT DEFAULT 0.0,
  evidence JSONB,
  verified_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT valid_verification_type CHECK (verification_type IN ('code_existence', 'code_usage', 'test_coverage', 'integration')),
  CONSTRAINT valid_verification_status CHECK (status IN ('pending', 'verified', 'failed'))
);

CREATE INDEX idx_task_verifications_task ON task_verifications(task_id);
CREATE INDEX idx_task_verifications_status ON task_verifications(status);
```

### Task Links Table (for integration with other systems)

```sql
CREATE TABLE task_links (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  link_type VARCHAR(50) NOT NULL, -- 'change_request', 'knowledge_item', 'comprehensive_analysis', 'test_requirement'
  linked_id UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT valid_link_type CHECK (link_type IN ('change_request', 'knowledge_item', 'comprehensive_analysis', 'test_requirement'))
);

CREATE INDEX idx_task_links_task ON task_links(task_id);
CREATE INDEX idx_task_links_linked ON task_links(link_type, linked_id);
```

---

## 11. Real-World Scenarios

### Scenario 1: Cursor Generates Task

**Context**: Developer asks Cursor to "add user authentication"

**Flow**:
1. Cursor generates code with TODO comment: `// TODO: Add JWT token refresh`
2. Task detection scans codebase, finds TODO
3. Task created: `TASK-123: Add JWT token refresh`
4. Dependency detection finds it depends on `TASK-122: Implement JWT middleware`
5. Verification runs: Code exists (0.95), Usage found (0.88), Tests missing (0.0)
6. Overall confidence: 0.61 → Marked as `in_progress`
7. Alert sent: "Task TASK-123 needs tests"

### Scenario 2: Integration Dependency

**Context**: Task requires external API integration

**Flow**:
1. Comprehensive analysis (Phase 14A) discovers feature needs payment gateway
2. Task created: `TASK-124: Integrate payment gateway`
3. Dependency detection identifies integration dependency
4. Verification checks: Code exists (0.90), Integration config missing (0.0)
5. Overall confidence: 0.45 → Marked as `pending`
6. Alert sent: "Task TASK-124 requires payment gateway configuration"

### Scenario 3: Feature-Level Dependency

**Context**: Multiple tasks part of larger feature

**Flow**:
1. Comprehensive analysis identifies "Order Cancellation" feature
2. Tasks discovered: TASK-125 (API), TASK-126 (UI), TASK-127 (Tests)
3. Feature-level dependencies created
4. TASK-125 depends on TASK-126 (UI needs API)
5. TASK-127 depends on TASK-125 and TASK-126 (tests need both)
6. Dependency graph built and visualized

### Scenario 4: Auto-Completion

**Context**: High-confidence verification

**Flow**:
1. Task TASK-128 verified: Code (0.95), Usage (0.92), Tests (0.88), Integration (0.85)
2. Overall confidence: 0.90 (>0.8 threshold)
3. Auto-completed: Status changed to `completed`, `completed_at` set
4. Notification: "Task TASK-128 auto-completed (confidence: 90%)"

---

## 12. Edge Cases and Failure Modes

### Edge Case 1: Circular Dependencies

**Problem**: Task A depends on Task B, Task B depends on Task A

**Solution**: 
- Cycle detection algorithm identifies circular dependencies
- Alert sent: "Circular dependency detected: TASK-A ↔ TASK-B"
- Tasks marked as `blocked` until cycle resolved

### Edge Case 2: False Positive Verification

**Problem**: Verification incorrectly marks task as complete

**Solution**:
- Multi-factor verification reduces false positives
- Manual override available: `sentinel tasks verify TASK-123 --manual`
- Confidence thresholds configurable per project

### Edge Case 3: Missing Dependencies

**Problem**: Task depends on non-existent task

**Solution**:
- Dependency validation checks task existence
- Alert sent: "Task TASK-123 depends on non-existent task TASK-999"
- Option to create missing task or remove dependency

### Edge Case 4: Partial Implementation

**Problem**: Task partially implemented but not complete

**Solution**:
- Verification factors show partial completion
- Status set to `in_progress` with confidence score
- Alert sent with missing factors

---

## 13. Performance Considerations

### Optimization Strategies

1. **Incremental Scanning**: Only scan changed files
2. **Caching**: Cache verification results (1 hour TTL)
3. **Batch Processing**: Verify multiple tasks in parallel
4. **Lazy Loading**: Load dependencies on-demand
5. **Indexing**: Database indexes on frequently queried fields

### Performance Targets

- Task detection: < 5 seconds for 1000 files
- Verification: < 2 seconds per task
- Dependency graph: < 1 second for 100 tasks
- API response: < 500ms for single task query

---

## 14. Security Considerations

### Data Privacy

- Tasks stored per-project (project isolation)
- No code content stored (only metadata)
- Verification evidence sanitized (no secrets)

### Access Control

- Project-level access control
- Role-based permissions (admin, developer, viewer)
- Audit logging for task modifications

---

## 15. Success Metrics

- **Task Detection Accuracy**: >90% of tasks detected
- **Verification Accuracy**: >85% correct verification
- **Dependency Detection Accuracy**: >80% correct dependencies
- **Auto-Completion Rate**: >70% for high-confidence tasks
- **False Positive Rate**: <10% for auto-completion
- **Performance**: <5 seconds for full scan, <2 seconds per verification

---

## 16. Future Enhancements

### Phase 14E.2 (Future)

- **Task Templates**: Pre-defined task templates for common patterns
- **Task Estimation**: Effort estimation based on historical data
- **Task Prioritization**: AI-powered task prioritization
- **Task Scheduling**: Automatic task scheduling based on dependencies
- **Task Analytics**: Task completion trends and insights

---

## 7. MCP Integration

### Using Task Tools in Cursor

The task dependency system is accessible via MCP tools in Cursor IDE, providing seamless integration for task management and verification.

#### Available Tools

1. **sentinel_get_task_status**: Get detailed status and information about a specific task
2. **sentinel_verify_task**: Verify task completion using multi-factor verification
3. **sentinel_list_tasks**: List tasks with advanced filtering and pagination

#### Quick Start

**Check Task Status**:
```json
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Verify Task**:
```json
{
  "name": "sentinel_verify_task",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000",
    "force": true
  }
}
```

**List Tasks**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "status": "pending",
    "priority": "high"
  }
}
```

#### Example Workflow

1. **Find High-Priority Tasks**:
   ```json
   {
     "name": "sentinel_list_tasks",
     "arguments": {
       "status": "in_progress",
       "priority": "high"
     }
   }
   ```

2. **Check Task Details**:
   ```json
   {
     "name": "sentinel_get_task_status",
     "arguments": {
       "taskId": "<task-id-from-step-1>"
     }
   }
   ```

3. **Verify Completion**:
   ```json
   {
     "name": "sentinel_verify_task",
     "arguments": {
       "taskId": "<task-id>",
       "force": true
     }
   }
   ```

#### Error Handling

Common errors and solutions:

- **Config Error (-32002)**: Set `SENTINEL_HUB_URL` and `SENTINEL_API_KEY` environment variables
- **Invalid Params (-32602)**: Check parameter types and enum values
- **Hub Unavailable (-32000)**: Verify Hub connectivity and network
- **Hub Timeout (-32001)**: Retry request or check Hub performance

#### Performance Tips

- **Caching**: `get_task_status` is cached for 30s, `list_tasks` for 10s
- **Pagination**: Use `limit` and `offset` for large task lists
- **Filtering**: Combine filters to narrow results efficiently

#### Complete Documentation

For complete MCP tool reference, examples, and troubleshooting, see:
- **[MCP_TASK_TOOLS_GUIDE.md](./MCP_TASK_TOOLS_GUIDE.md)** - Complete MCP tools reference

---

## References

- [IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md) - Phase 14E specification
- [MCP_TASK_TOOLS_GUIDE.md](./MCP_TASK_TOOLS_GUIDE.md) - MCP tools reference and examples
- [COMPREHENSIVE_ANALYSIS_SOLUTION.md](./COMPREHENSIVE_ANALYSIS_SOLUTION.md) - Feature-level dependency detection
- [ARCHITECTURE.md](./ARCHITECTURE.md) - System architecture
- [TECHNICAL_SPEC.md](./TECHNICAL_SPEC.md) - Technical specifications


