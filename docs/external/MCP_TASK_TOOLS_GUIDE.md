# MCP Task Tools Guide

> **For Cursor Users**: Complete reference for using Sentinel task management tools via MCP in Cursor IDE.

## Overview

The Sentinel MCP Task Tools provide seamless integration between Cursor IDE and Sentinel's task dependency and verification system. These tools allow you to check task status, verify completion, and list tasks directly from Cursor.

## Prerequisites

- Cursor IDE with MCP support
- Sentinel Hub configured (`SENTINEL_HUB_URL` and `SENTINEL_API_KEY`)
- Phase 14E Task Dependency System enabled

## Tools Reference

### 1. sentinel_get_task_status

Get detailed status and information about a specific task.

#### Purpose
- Check if a task is complete
- View task details (title, description, priority, status)
- See verification confidence score
- View dependencies (blocked by, blocks)
- Check file location and line number

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `taskId` | string | Yes | UUID of the task |
| `codebasePath` | string | No | Optional codebase path for dependency analysis (defaults to current directory) |

#### Return Values

**Success Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "📋 Task: Implement user authentication\n\n🆔 ID: 550e8400-e29b-41d4-a716-446655440000\n📊 Status: in_progress\n⭐ Priority: high\n✅ Verification Confidence: 75%\n📁 File: src/auth.js:42\n\n📝 Description: Add JWT-based authentication for API endpoints\n\n🚫 Blocked by: [task-id-1]\n🔗 Blocks: [task-id-2]"
      }
    ],
    "data": {
      "task": { /* full task object */ },
      "dependencies": { /* dependency information */ }
    }
  }
}
```

**Error Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "taskId is required and must be a string"
  }
}
```

#### Examples

**Basic Usage**:
```json
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**With Codebase Path**:
```json
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000",
    "codebasePath": "/path/to/project"
  }
}
```

#### Error Handling

| Error Code | Meaning | Solution |
|------------|---------|----------|
| `-32602` | Invalid params | Ensure `taskId` is provided and is a valid UUID string |
| `-32002` | Config error | Set `SENTINEL_HUB_URL` and `SENTINEL_API_KEY` environment variables |
| `-32000` | Hub unavailable | Check Hub connectivity and network |
| `-32001` | Hub timeout | Request timed out (30s). Retry or check Hub status |

---

### 2. sentinel_verify_task

Verify task completion using multi-factor verification.

#### Purpose
- Verify if a task is actually completed
- Get detailed verification breakdown
- Force re-verification (bypass cache)
- Check verification confidence score

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `taskId` | string | Yes | UUID of the task |
| `force` | boolean | No | Force verification even if recently verified (default: false) |
| `codebasePath` | string | No | Optional codebase path for verification (defaults to current directory) |

#### Return Values

**Success Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "🔍 Task Verification Results\n\n🆔 Task ID: 550e8400-e29b-41d4-a716-446655440000\n✅ Overall Confidence: 85%\n📊 Status: completed\n\n📋 Verification Factors:\n  ✅ code_existence: passed (90%)\n  ✅ code_usage: passed (80%)\n  ✅ test_coverage: passed (85%)\n  ⏳ integration: pending (70%)"
      }
    ],
    "data": {
      "task_id": "550e8400-e29b-41d4-a716-446655440000",
      "overall_confidence": 0.85,
      "status": "completed",
      "verifications": [ /* verification details */ ]
    }
  }
}
```

#### Examples

**Basic Verification**:
```json
{
  "name": "sentinel_verify_task",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Force Re-verification**:
```json
{
  "name": "sentinel_verify_task",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000",
    "force": true
  }
}
```

**With Codebase Path**:
```json
{
  "name": "sentinel_verify_task",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000",
    "codebasePath": "/path/to/project",
    "force": false
  }
}
```

#### Verification Factors

The verification engine checks four factors:

1. **Code Existence**: Is the code present in the codebase?
2. **Code Usage**: Is the code actually used/referenced?
3. **Test Coverage**: Are there tests for this code?
4. **Integration**: Is the code integrated with the system?

Each factor returns:
- **Status**: `passed`, `failed`, or `pending`
- **Confidence**: 0.0 to 1.0 (0% to 100%)

#### Error Handling

| Error Code | Meaning | Solution |
|------------|---------|----------|
| `-32602` | Invalid params | Ensure `taskId` is provided |
| `-32002` | Config error | Configure Hub URL and API key |
| `-32001` | Hub timeout | Verification timed out (60s). Retry or check Hub |

---

### 3. sentinel_list_tasks

List tasks with advanced filtering and pagination.

#### Purpose
- Find tasks by status, priority, source
- Filter by assignee or tags
- Include/exclude archived tasks
- Paginate through large task lists

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `status` | string | No | Filter by status: `pending`, `in_progress`, `completed`, `blocked` |
| `priority` | string | No | Filter by priority: `low`, `medium`, `high`, `critical` |
| `source` | string | No | Filter by source: `cursor`, `manual`, `change_request`, `comprehensive_analysis` |
| `assigned_to` | string | No | Filter by assignee email |
| `tags` | array[string] | No | Filter by tags (array of strings) |
| `include_archived` | boolean | No | Include archived tasks (default: false) |
| `limit` | integer | No | Maximum tasks to return (1-100, default: 50) |
| `offset` | integer | No | Offset for pagination (default: 0) |

#### Return Values

**Success Response**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "📋 Tasks (3 total)\n\n1. ✅ 🔴 [in_progress] Implement user authentication\n   🆔 550e8400-e29b-41d4-a716-446655440000 | ✅ 75% confidence\n\n2. ⏳ 🟡 [pending] Add error handling\n   🆔 550e8400-e29b-41d4-a716-446655440001 | ✅ 0% confidence\n\n3. ✅ ⚪ [completed] Write unit tests\n   🆔 550e8400-e29b-41d4-a716-446655440002 | ✅ 95% confidence"
      }
    ],
    "data": {
      "tasks": [ /* array of task objects */ ],
      "total": 3,
      "limit": 50,
      "offset": 0,
      "has_next": false,
      "has_previous": false
    }
  }
}
```

#### Examples

**List All Tasks**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {}
}
```

**Filter by Status**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "status": "pending"
  }
}
```

**Filter by Priority**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "priority": "high"
  }
}
```

**Multiple Filters**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "status": "in_progress",
    "priority": "high",
    "source": "cursor",
    "limit": 25,
    "offset": 0
  }
}
```

**Filter by Tags**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "tags": ["bug", "urgent"]
  }
}
```

**Include Archived**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "include_archived": true
  }
}
```

**Pagination**:
```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "limit": 20,
    "offset": 40
  }
}
```

#### Error Handling

| Error Code | Meaning | Solution |
|------------|---------|----------|
| `-32602` | Invalid params | Check filter values match allowed enums (status, priority, source) |
| `-32002` | Config error | Configure Hub URL and API key |
| `-32000` | Hub unavailable | Check Hub connectivity |

---

## Common Workflows

### Workflow 1: Check Task Status

**Scenario**: You want to check if a task is complete and see its details.

```json
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Result**: See task status, confidence, dependencies, and file location.

---

### Workflow 2: Verify Task Completion

**Scenario**: You've completed a task and want to verify it's actually done.

```json
{
  "name": "sentinel_verify_task",
  "arguments": {
    "taskId": "550e8400-e29b-41d4-a716-446655440000",
    "force": true
  }
}
```

**Result**: Get detailed verification breakdown showing what's verified and what's missing.

---

### Workflow 3: Find Blocked Tasks

**Scenario**: You want to find tasks that are blocked by incomplete dependencies.

```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "status": "blocked"
  }
}
```

Then for each blocked task:
```json
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "taskId": "<task-id>"
  }
}
```

**Result**: See which tasks are blocking each task.

---

### Workflow 4: Filter High-Priority Tasks

**Scenario**: You want to see all high-priority tasks that are in progress.

```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "status": "in_progress",
    "priority": "high"
  }
}
```

**Result**: List of high-priority in-progress tasks.

---

### Workflow 5: Find Tasks by Tag

**Scenario**: You want to find all tasks tagged with "security".

```json
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "tags": ["security"]
  }
}
```

**Result**: All tasks with the "security" tag.

---

## Best Practices

### When to Use Each Tool

1. **sentinel_get_task_status**: 
   - Quick status check
   - View task details
   - Check dependencies
   - **Cached for 30 seconds** - fast for repeated checks

2. **sentinel_verify_task**:
   - After completing a task
   - When you need detailed verification breakdown
   - To force re-verification
   - **Not cached** - always fresh results

3. **sentinel_list_tasks**:
   - Finding tasks by criteria
   - Browsing task list
   - Filtering by attributes
   - **Cached for 10 seconds** - efficient for repeated queries

### Performance Considerations

- **Caching**: `get_task_status` and `list_tasks` are cached. Use `verify_task` with `force: true` for fresh results.
- **Pagination**: Use `limit` and `offset` for large task lists to avoid loading everything.
- **Filtering**: Combine filters to narrow results and reduce response size.

### Error Handling Strategies

1. **Config Errors**: Always check `SENTINEL_HUB_URL` and `SENTINEL_API_KEY` are set.
2. **Network Errors**: Retry failed requests. The system has built-in retry logic.
3. **Invalid Params**: Check parameter types and enum values match documentation.
4. **Timeouts**: Increase timeout or check Hub status if requests timeout frequently.

### Caching Behavior

- **get_task_status**: Cached for 30 seconds. Same task ID returns cached result.
- **list_tasks**: Cached for 10 seconds. Cache key includes all filter parameters.
- **verify_task**: Never cached. Always makes fresh verification request.

---

## Troubleshooting

### Common Errors

#### "Hub not configured"
**Cause**: `SENTINEL_HUB_URL` or `SENTINEL_API_KEY` not set.

**Solution**:
```bash
export SENTINEL_HUB_URL="http://localhost:8080"
export SENTINEL_API_KEY="your-api-key"
```

#### "Invalid params"
**Cause**: Parameter validation failed (wrong type, invalid enum value, etc.).

**Solution**: Check parameter types and values match the documentation.

#### "Hub unavailable"
**Cause**: Cannot connect to Hub (network issue, Hub down).

**Solution**: 
- Check Hub is running
- Verify network connectivity
- Check `SENTINEL_HUB_URL` is correct

#### "Hub timeout"
**Cause**: Request took too long (>30s for GET, >60s for POST).

**Solution**:
- Check Hub performance
- Retry the request
- Verify Hub is not overloaded

### Debugging Tips

1. **Check Response Structure**: Always check the `error` field in responses for details.
2. **Verify Task IDs**: Ensure task IDs are valid UUIDs.
3. **Test Connectivity**: Use `curl` to test Hub endpoints directly.
4. **Check Logs**: Look at Sentinel logs for detailed error information.

### Performance Issues

1. **Slow Responses**: Check Hub performance and network latency.
2. **Cache Not Working**: Verify cache TTL hasn't expired.
3. **Large Task Lists**: Use pagination (`limit` and `offset`).

---

## Integration with Other Sentinel Features

### Phase 14A: Comprehensive Feature Analysis
Tasks can be linked to comprehensive feature analysis results for feature-level dependency detection.

### Phase 12: Change Request Management
Tasks can be linked to change requests for tracking implementation status.

### Phase 11: Doc-Sync
Task status can be synced with documentation status markers.

### Phase 10: Test Enforcement
Task verification includes test coverage checking.

---

## Examples in Context

### Example 1: Daily Task Review

```json
// 1. List all in-progress tasks
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "status": "in_progress"
  }
}

// 2. For each task, check status
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "taskId": "<task-id>"
  }
}

// 3. Verify completed tasks
{
  "name": "sentinel_verify_task",
  "arguments": {
    "taskId": "<task-id>",
    "force": true
  }
}
```

### Example 2: Finding Blocked Work

```json
// 1. Find blocked tasks
{
  "name": "sentinel_list_tasks",
  "arguments": {
    "status": "blocked"
  }
}

// 2. Check what's blocking each task
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "taskId": "<blocked-task-id>"
  }
}
// Response shows "Blocked by: [task-ids]"

// 3. Check status of blocking tasks
{
  "name": "sentinel_get_task_status",
  "arguments": {
    "taskId": "<blocking-task-id>"
  }
}
```

---

## References

- [TASK_DEPENDENCY_SYSTEM.md](./TASK_DEPENDENCY_SYSTEM.md) - Complete task dependency system documentation
- [IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md) - Phase 14E implementation details
- [ARCHITECTURE.md](./ARCHITECTURE.md) - System architecture

---

## Support

For issues or questions:
1. Check this guide for common solutions
2. Review error messages for specific guidance
3. Check Sentinel logs for detailed error information
4. Refer to [TASK_DEPENDENCY_SYSTEM.md](./TASK_DEPENDENCY_SYSTEM.md) for system details









