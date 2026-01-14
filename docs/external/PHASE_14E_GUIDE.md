# Phase 14E: Task Dependency & Verification System Guide

Complete guide to using the Task Dependency & Verification System for tracking and verifying Cursor-generated tasks.

## Overview

Phase 14E provides a comprehensive system for:
- **Task Detection**: Automatically scan codebase for TODO comments, Cursor task markers, and explicit task formats
- **Task Verification**: Multi-factor verification to confirm task completion
- **Dependency Management**: Track dependencies between tasks and detect blocking relationships
- **Auto-Completion**: Automatically mark tasks as complete when verification confidence is high
- **Integration**: Deep integration with other Sentinel features (Doc-Sync, Change Requests, Comprehensive Analysis, Test Requirements)

## Table of Contents

1. [Quick Start](#quick-start)
2. [Task Detection](#task-detection)
3. [Task Management](#task-management)
4. [Task Verification](#task-verification)
5. [Dependency Management](#dependency-management)
6. [Auto-Completion](#auto-completion)
7. [MCP Integration](#mcp-integration)
8. [Best Practices](#best-practices)
9. [Examples](#examples)

## Quick Start

### 1. Scan Your Codebase

```bash
# Scan for tasks in current directory
./sentinel tasks scan

# Scan specific directory
./sentinel tasks scan --path ./src

# Scan with custom patterns
./sentinel tasks scan --patterns "TODO,FIXME,HACK"
```

### 2. List Tasks

```bash
# List all tasks
./sentinel tasks list

# Filter by status
./sentinel tasks list --status pending

# Filter by priority
./sentinel tasks list --priority high

# Include archived tasks
./sentinel tasks list --include-archived
```

### 3. Verify Tasks

```bash
# Verify a specific task
./sentinel tasks verify TASK-123

# Verify all pending tasks
./sentinel tasks verify --all
```

## Task Detection

The system automatically detects tasks from:

### 1. TODO/FIXME Comments

```javascript
// TODO: Implement user authentication
// FIXME: Fix memory leak in cache
// HACK: Temporary solution, needs refactoring
```

### 2. Cursor Task Markers

```javascript
// Cursor Task: Implement JWT middleware
// Task: Add rate limiting
```

### 3. Explicit Task Format

```javascript
// TASK-001: Implement user authentication
// TASK-002: Add password reset functionality
```

### 4. Priority and Tags

```javascript
// TODO: CRITICAL - Fix security vulnerability
// TODO: HIGH - Add rate limiting
// TODO: LOW - Update documentation
// TODO: #auth #security Implement JWT middleware
```

## Task Management

### Creating Tasks

#### Via API

```bash
curl -X POST "http://localhost:8080/api/v1/tasks" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Implement user authentication",
    "description": "Add JWT-based authentication system",
    "file_path": "src/auth.js",
    "line_number": 42,
    "priority": "high",
    "tags": ["auth", "security"]
  }'
```

#### Via Agent Command

```bash
# Tasks are automatically created during scan
./sentinel tasks scan
```

### Updating Tasks

```bash
curl -X PUT "http://localhost:8080/api/v1/tasks/{id}" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_progress",
    "priority": "critical",
    "assigned_to": "developer@example.com"
  }'
```

### Task Statuses

- **pending**: Task detected but not started
- **in_progress**: Task is being worked on
- **completed**: Task is finished
- **blocked**: Task is blocked by dependencies

### Task Priorities

- **low**: Low priority task
- **medium**: Medium priority task
- **high**: High priority task
- **critical**: Critical priority task

## Task Verification

The system uses multi-factor verification to confirm task completion:

### Verification Factors

1. **Code Existence**: Checks if the code described in the task actually exists
2. **Code Usage**: Verifies that the code is being used (not orphaned)
3. **Test Coverage**: Ensures tests exist for the implementation
4. **Integration**: Verifies integration with other systems

### Verification Confidence

Confidence scores range from 0.0 to 1.0:
- **< 0.5**: Pending verification
- **0.5 - 0.7**: Low confidence (needs review)
- **0.7 - 0.9**: Medium confidence (likely complete)
- **> 0.9**: High confidence (auto-completed)

### Manual Verification

```bash
# Verify specific task
./sentinel tasks verify TASK-123

# Verify all pending tasks
./sentinel tasks verify --all
```

### Verification Response

```json
{
  "success": true,
  "verification": {
    "status": "verified",
    "confidence": 0.95,
    "factors": {
      "code_existence": true,
      "code_usage": true,
      "test_coverage": true,
      "integration": false
    },
    "evidence": {
      "code_files": ["src/auth.js"],
      "test_files": ["tests/auth.test.js"],
      "usage_locations": ["src/routes.js"]
    }
  }
}
```

## Dependency Management

### Dependency Types

1. **Explicit**: Task description mentions dependency (e.g., "Depends on: TASK-123")
2. **Implicit**: Code analysis shows Task A calls Task B's code
3. **Integration**: Task requires external API/service setup
4. **Feature**: Task is part of larger feature (from Phase 14A Comprehensive Analysis)

### Detecting Dependencies

```bash
# Detect dependencies for a task
./sentinel tasks dependencies TASK-123

# Or via API
curl -X POST "http://localhost:8080/api/v1/tasks/{id}/detect-dependencies" \
  -H "Authorization: Bearer $API_KEY"
```

### Dependency Graph

```bash
# Get dependency graph
curl -X GET "http://localhost:8080/api/v1/tasks/{id}/dependencies" \
  -H "Authorization: Bearer $API_KEY"
```

**Response:**
```json
{
  "success": true,
  "dependencies": {
    "depends_on": [
      {
        "task_id": "TASK-001",
        "dependency_type": "explicit",
        "confidence": 1.0
      }
    ],
    "required_by": [
      {
        "task_id": "TASK-003",
        "dependency_type": "implicit",
        "confidence": 0.8
      }
    ],
    "cycles": []
  }
}
```

### Cycle Detection

The system automatically detects circular dependencies:

```json
{
  "cycles": [
    {
      "tasks": ["TASK-001", "TASK-002", "TASK-001"],
      "severity": "error"
    }
  ]
}
```

## Auto-Completion

Tasks with high verification confidence (>0.9) are automatically marked as complete.

### Auto-Completion Criteria

- Verification confidence > 0.9
- All critical verification factors pass
- No blocking dependencies

### Manual Completion

```bash
# Mark task as complete
./sentinel tasks complete TASK-123
```

## MCP Integration

The system provides three MCP tools for Cursor integration:

### 1. sentinel_get_task_status

Get the status of a specific task.

**Parameters:**
- `taskId` (required): Task ID
- `codebasePath` (optional): Path to codebase for dependency analysis

**Example:**
```json
{
  "tool": "sentinel_get_task_status",
  "arguments": {
    "taskId": "TASK-123",
    "codebasePath": "."
  }
}
```

### 2. sentinel_verify_task

Verify task completion.

**Parameters:**
- `taskId` (required): Task ID
- `codebasePath` (optional): Path to codebase for verification

**Example:**
```json
{
  "tool": "sentinel_verify_task",
  "arguments": {
    "taskId": "TASK-123"
  }
}
```

### 3. sentinel_list_tasks

List tasks with optional filters.

**Parameters:**
- `status` (optional): Filter by status
- `priority` (optional): Filter by priority
- `source` (optional): Filter by source
- `assignedTo` (optional): Filter by assignee
- `tags` (optional): Filter by tags (comma-separated)
- `includeArchived` (optional): Include archived tasks
- `limit` (optional): Maximum results (default: 50)
- `offset` (optional): Pagination offset

**Example:**
```json
{
  "tool": "sentinel_list_tasks",
  "arguments": {
    "status": "pending",
    "priority": "high",
    "limit": 20
  }
}
```

See [MCP_TASK_TOOLS_GUIDE.md](./MCP_TASK_TOOLS_GUIDE.md) for complete MCP tool documentation.

## Best Practices

### 1. Use Descriptive Task Titles

**Good:**
```javascript
// TODO: Implement JWT-based authentication middleware
```

**Bad:**
```javascript
// TODO: Fix this
```

### 2. Include Context in Descriptions

When creating tasks manually, include:
- What needs to be done
- Why it's needed
- Any relevant context or constraints

### 3. Use Priority Appropriately

- **Critical**: Security issues, blocking bugs
- **High**: Important features, significant improvements
- **Medium**: Standard features, minor improvements
- **Low**: Nice-to-have features, documentation

### 4. Tag Tasks Consistently

Use consistent tags for filtering:
- Feature tags: `#auth`, `#payment`, `#api`
- Type tags: `#bug`, `#feature`, `#refactor`
- Component tags: `#frontend`, `#backend`, `#database`

### 5. Regular Verification

Run verification regularly to catch completed tasks:
```bash
# Daily verification
./sentinel tasks verify --all
```

### 6. Review Dependencies

Check dependencies before starting work:
```bash
./sentinel tasks dependencies TASK-123
```

## Examples

### Example 1: Complete Task Lifecycle

```bash
# 1. Scan codebase
./sentinel tasks scan

# 2. List pending tasks
./sentinel tasks list --status pending

# 3. Check dependencies
./sentinel tasks dependencies TASK-123

# 4. Work on task...

# 5. Verify completion
./sentinel tasks verify TASK-123

# 6. Task auto-completes if confidence > 0.9
```

### Example 2: Creating Task from Change Request

```bash
# Link task to change request
curl -X POST "http://localhost:8080/api/v1/tasks/{taskId}/links" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "link_type": "change_request",
    "linked_id": "CR-001"
  }'
```

### Example 3: Batch Verification

```bash
# Verify all pending tasks
./sentinel tasks verify --all

# Response shows:
# - Verified: 10 tasks
# - Failed: 2 tasks
# - Skipped: 5 tasks (blocked by dependencies)
```

## Integration with Other Features

### Phase 11: Doc-Sync

Tasks are automatically linked to documentation status markers. When documentation is updated, related tasks are notified.

### Phase 12: Change Requests

Tasks can be created from change requests and linked to track implementation status.

### Phase 14A: Comprehensive Analysis

Tasks are linked to comprehensive analysis results for feature-level dependency tracking.

### Phase 10: Test Requirements

Tasks are linked to test requirements to track test coverage.

## Troubleshooting

### Tasks Not Detected

1. Check if patterns match your comment style
2. Verify file paths are included in scan
3. Check task detection logs

### Verification Failing

1. Ensure code actually exists
2. Check test coverage
3. Verify integration points
4. Review verification evidence

### Dependencies Not Detected

1. Run dependency detection manually
2. Check code references
3. Verify task descriptions mention dependencies

## Additional Resources

- [TASK_DEPENDENCY_SYSTEM.md](./TASK_DEPENDENCY_SYSTEM.md) - Complete system documentation
- [MCP_TASK_TOOLS_GUIDE.md](./MCP_TASK_TOOLS_GUIDE.md) - MCP tools reference
- [HUB_API_REFERENCE.md](./HUB_API_REFERENCE.md) - API endpoint documentation









