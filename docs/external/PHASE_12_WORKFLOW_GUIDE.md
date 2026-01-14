# Phase 12: Change Request Workflow Guide

Complete guide to using the Change Request system for managing business rule changes and tracking implementation status.

## Overview

Phase 12 provides a comprehensive workflow for:
- **Gap Analysis**: Identify gaps between documented business rules and actual code implementation
- **Change Requests**: Create, approve, and track change requests for business rule modifications
- **Impact Analysis**: Analyze the impact of proposed changes on existing code and tests
- **Implementation Tracking**: Track implementation status of approved change requests
- **Task Integration**: Link change requests to tasks for implementation tracking

## Table of Contents

1. [Quick Start](#quick-start)
2. [Gap Analysis Workflow](#gap-analysis-workflow)
3. [Change Request Lifecycle](#change-request-lifecycle)
4. [Impact Analysis](#impact-analysis)
5. [Approval/Rejection Process](#approvalrejection-process)
6. [Implementation Tracking](#implementation-tracking)
7. [Integration with Tasks](#integration-with-tasks)
8. [Best Practices](#best-practices)
9. [Examples](#examples)

## Quick Start

### 1. Run Gap Analysis

```bash
# Analyze gaps between business rules and code
curl -X POST "http://localhost:8080/api/v1/knowledge/gap-analysis" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "your-project-id",
    "codebase_path": "."
  }'
```

### 2. Create Change Request

```bash
curl -X POST "http://localhost:8080/api/v1/change-requests" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "knowledge_item_id": "knowledge-item-id",
    "type": "modification",
    "current_state": {...},
    "proposed_state": {...}
  }'
```

### 3. Approve Change Request

```bash
curl -X POST "http://localhost:8080/api/v1/change-requests/{id}/approve" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "approved_by": "admin@example.com"
  }'
```

## Gap Analysis Workflow

### What is Gap Analysis?

Gap analysis identifies discrepancies between:
- **Documented Business Rules**: Rules extracted from documents (Phase 4)
- **Actual Code Implementation**: What's actually implemented in the codebase

### Running Gap Analysis

**API Endpoint**: `POST /api/v1/knowledge/gap-analysis`

**Request:**
```json
{
  "project_id": "uuid",
  "codebase_path": ".",
  "options": {
    "include_tests": true,
    "severity_filter": ["critical", "high"]
  }
}
```

**Response:**
```json
{
  "success": true,
  "report": {
    "gaps": [
      {
        "type": "missing_impl",
        "knowledge_item_id": "uuid",
        "rule_title": "User authentication required",
        "description": "Business rule is documented but not implemented in code",
        "severity": "critical",
        "recommendation": "Implement authentication middleware"
      }
    ],
    "summary": {
      "total_gaps": 10,
      "critical": 2,
      "high": 5,
      "medium": 3
    }
  }
}
```

### Gap Types

1. **missing_impl**: Rule documented but not implemented
2. **missing_doc**: Code exists but not documented
3. **partial_match**: Partial implementation exists
4. **tests_missing**: Implementation exists but no tests

### Using Gap Analysis Results

1. **Review Gaps**: Identify critical gaps that need immediate attention
2. **Create Change Requests**: Convert gaps into change requests
3. **Prioritize**: Focus on critical and high-severity gaps first
4. **Track Progress**: Monitor gap resolution over time

## Change Request Lifecycle

### States

1. **pending_approval**: Change request created, awaiting approval
2. **approved**: Change request approved, ready for implementation
3. **rejected**: Change request rejected, not to be implemented

### Implementation States

1. **pending**: Approved but not started
2. **in_progress**: Currently being implemented
3. **completed**: Implementation finished
4. **blocked**: Blocked by dependencies or issues

### Creating Change Requests

**API Endpoint**: `POST /api/v1/change-requests`

**Request:**
```json
{
  "knowledge_item_id": "uuid",
  "type": "modification",
  "current_state": {
    "rule": "User must be authenticated",
    "implementation": "Basic auth only"
  },
  "proposed_state": {
    "rule": "User must be authenticated",
    "implementation": "JWT-based authentication"
  },
  "impact_analysis": {
    "affected_files": ["src/auth.js"],
    "test_impact": "High - requires test updates"
  }
}
```

**Change Types:**
- **new**: New business rule
- **modification**: Modify existing rule
- **removal**: Remove existing rule
- **unchanged**: No change (for tracking)

## Impact Analysis

### Running Impact Analysis

**API Endpoint**: `POST /api/v1/change-requests/{id}/impact`

**Response:**
```json
{
  "success": true,
  "impact": {
    "affected_files": [
      {
        "file": "src/auth.js",
        "impact_level": "high",
        "changes_required": ["Update authentication logic"]
      }
    ],
    "test_impact": {
      "affected_tests": 5,
      "new_tests_required": 3,
      "tests_to_update": 2
    },
    "dependencies": {
      "blocks": ["TASK-123"],
      "blocked_by": []
    }
  }
}
```

### Impact Levels

- **critical**: Breaks existing functionality
- **high**: Significant changes required
- **medium**: Moderate changes required
- **low**: Minimal changes required

## Approval/Rejection Process

### Approving Change Requests

**API Endpoint**: `POST /api/v1/change-requests/{id}/approve`

**Request:**
```json
{
  "approved_by": "admin@example.com"
}
```

**Response:**
```json
{
  "success": true,
  "change_request": {
    "id": "uuid",
    "status": "approved",
    "approved_by": "admin@example.com",
    "approved_at": "2024-12-11T10:00:00Z"
  }
}
```

### Rejecting Change Requests

**API Endpoint**: `POST /api/v1/change-requests/{id}/reject`

**Request:**
```json
{
  "rejected_by": "admin@example.com",
  "rejection_reason": "Not aligned with business strategy"
}
```

**Response:**
```json
{
  "success": true,
  "change_request": {
    "id": "uuid",
    "status": "rejected",
    "rejected_by": "admin@example.com",
    "rejected_at": "2024-12-11T10:00:00Z",
    "rejection_reason": "Not aligned with business strategy"
  }
}
```

## Implementation Tracking

### Starting Implementation

**API Endpoint**: `POST /api/v1/change-requests/{id}/start`

**Response:**
```json
{
  "success": true,
  "change_request": {
    "id": "uuid",
    "implementation_status": "in_progress"
  }
}
```

### Updating Implementation Status

**API Endpoint**: `POST /api/v1/change-requests/{id}/update`

**Request:**
```json
{
  "implementation_status": "in_progress",
  "implementation_notes": "Working on JWT implementation, 50% complete"
}
```

### Completing Implementation

**API Endpoint**: `POST /api/v1/change-requests/{id}/complete`

**Request:**
```json
{
  "implementation_notes": "JWT authentication fully implemented and tested"
}
```

**Response:**
```json
{
  "success": true,
  "change_request": {
    "id": "uuid",
    "implementation_status": "completed",
    "completed_at": "2024-12-11T15:00:00Z"
  }
}
```

## Integration with Tasks

### Creating Tasks from Change Requests

When a change request is approved, tasks can be automatically created:

**API Endpoint**: `POST /api/v1/tasks/from-change-request`

**Request:**
```json
{
  "change_request_id": "uuid",
  "project_id": "uuid"
}
```

**Response:**
```json
{
  "success": true,
  "tasks_created": [
    "TASK-001",
    "TASK-002"
  ]
}
```

### Linking Tasks to Change Requests

**API Endpoint**: `POST /api/v1/tasks/{taskId}/links`

**Request:**
```json
{
  "link_type": "change_request",
  "linked_id": "change-request-id"
}
```

### Syncing Status

Task status automatically syncs with change request implementation status:
- Task completed → Change request implementation status updated
- Change request completed → Related tasks marked as complete

## Best Practices

### 1. Document Current State

Always document the current state when creating change requests:
```json
{
  "current_state": {
    "implementation": "Basic auth",
    "limitations": "No token refresh"
  }
}
```

### 2. Provide Detailed Proposed State

Include specific implementation details:
```json
{
  "proposed_state": {
    "implementation": "JWT-based auth",
    "features": ["Token refresh", "Role-based access"]
  }
}
```

### 3. Run Impact Analysis First

Always run impact analysis before approval:
1. Understand affected files
2. Identify test requirements
3. Check dependencies
4. Estimate effort

### 4. Use Implementation Notes

Keep detailed implementation notes:
```json
{
  "implementation_notes": "Day 1: Set up JWT library. Day 2: Implement token generation. Day 3: Add refresh logic."
}
```

### 5. Link Related Tasks

Link all related tasks to change requests for better tracking.

## Examples

### Example 1: Complete Workflow

```bash
# 1. Run gap analysis
curl -X POST "http://localhost:8080/api/v1/knowledge/gap-analysis" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"project_id": "uuid", "codebase_path": "."}'

# 2. Create change request from gap
curl -X POST "http://localhost:8080/api/v1/change-requests" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "knowledge_item_id": "uuid",
    "type": "modification",
    "current_state": {...},
    "proposed_state": {...}
  }'

# 3. Run impact analysis
curl -X POST "http://localhost:8080/api/v1/change-requests/{id}/impact" \
  -H "Authorization: Bearer $API_KEY"

# 4. Approve change request
curl -X POST "http://localhost:8080/api/v1/change-requests/{id}/approve" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"approved_by": "admin@example.com"}'

# 5. Start implementation
curl -X POST "http://localhost:8080/api/v1/change-requests/{id}/start" \
  -H "Authorization: Bearer $API_KEY"

# 6. Create tasks
curl -X POST "http://localhost:8080/api/v1/tasks/from-change-request" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"change_request_id": "uuid", "project_id": "uuid"}'

# 7. Update implementation status
curl -X POST "http://localhost:8080/api/v1/change-requests/{id}/update" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"implementation_status": "in_progress", "implementation_notes": "50% complete"}'

# 8. Complete implementation
curl -X POST "http://localhost:8080/api/v1/change-requests/{id}/complete" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{"implementation_notes": "Fully implemented and tested"}'
```

### Example 2: Dashboard View

```bash
# Get change requests dashboard
curl -X GET "http://localhost:8080/api/v1/change-requests/dashboard?project_id=uuid" \
  -H "Authorization: Bearer $API_KEY"
```

**Response:**
```json
{
  "success": true,
  "dashboard": {
    "pending_approval": 5,
    "approved": 10,
    "in_progress": 3,
    "completed": 25,
    "rejected": 2
  }
}
```

## Additional Resources

- [PHASE_12_GUIDE.md](./PHASE_12_GUIDE.md) - Phase 12 technical guide
- [TASK_DEPENDENCY_SYSTEM.md](./TASK_DEPENDENCY_SYSTEM.md) - Task system integration
- [HUB_API_REFERENCE.md](./HUB_API_REFERENCE.md) - Complete API reference









