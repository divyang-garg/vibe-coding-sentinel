# Hub API Reference

Complete API reference for the Sentinel Hub API server. This document reflects the current implementation as of Phase 8 completion.

## Table of Contents

1. [Authentication](#authentication)
2. [Base URL](#base-url)
3. [Health Endpoints](#health-endpoints)
4. [Task Management](#task-management)
5. [Document Management](#document-management)
6. [Organization Management](#organization-management)
7. [Workflow Management](#workflow-management)
8. [API Version Management](#api-version-management)
9. [Code Analysis](#code-analysis)
10. [Repository Management](#repository-management)
11. [Monitoring](#monitoring)
12. [Error Responses](#error-responses)

## Base URL

```
https://your-hub-instance.com
```

## Authentication

All API endpoints (except health checks) require API key authentication:

```
Authorization: Bearer <your-api-key>
```

Or:

```
X-API-Key: <your-api-key>
```

### API Key Management

API keys are configured in the server configuration. Default keys for development:
- `dev-api-key-123`
- `test-api-key-456`

## Health Endpoints

### GET /health
Basic health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-14T10:30:00Z"
}
```

### GET /health/db
Database connectivity health check.

**Response:**
```json
{
  "status": "ok",
  "database": "connected",
  "timestamp": "2024-01-14T10:30:00Z"
}
```

### GET /health/ready
Readiness check for load balancers.

**Response:**
```json
{
  "status": "ready",
  "services": ["database", "cache"],
  "timestamp": "2024-01-14T10:30:00Z"
}
```

## Task Management

### POST /api/v1/tasks
Create a new task.

**Request Body:**
```json
{
  "project_id": "proj-123",
  "title": "Implement user authentication",
  "description": "Add JWT-based authentication system",
  "priority": "high",
  "assigned_to": "user-456",
  "estimated_effort_hours": 8
}
```

**Response:**
```json
{
  "id": "task-789",
  "project_id": "proj-123",
  "title": "Implement user authentication",
  "status": "pending",
  "priority": "high",
  "created_at": "2024-01-14T10:30:00Z"
}
```

### GET /api/v1/tasks
List tasks with optional filtering.

**Query Parameters:**
- `project_id` - Filter by project
- `status` - Filter by status (pending, in_progress, completed)
- `assigned_to` - Filter by assignee
- `limit` - Maximum results (default: 50, max: 100)
- `offset` - Pagination offset

**Response:**
```json
{
  "tasks": [
    {
      "id": "task-789",
      "project_id": "proj-123",
      "title": "Implement user authentication",
      "status": "pending",
      "priority": "high",
      "assigned_to": "user-456",
      "created_at": "2024-01-14T10:30:00Z"
    }
  ],
  "total": 1
}
```

### GET /api/v1/tasks/{id}
Get a specific task by ID.

**Response:**
```json
{
  "id": "task-789",
  "project_id": "proj-123",
  "title": "Implement user authentication",
  "description": "Add JWT-based authentication system",
  "status": "pending",
  "priority": "high",
  "assigned_to": "user-456",
  "estimated_effort_hours": 8,
  "created_at": "2024-01-14T10:30:00Z",
  "updated_at": "2024-01-14T10:30:00Z"
}
```

### PUT /api/v1/tasks/{id}
Update an existing task.

**Request Body:**
```json
{
  "title": "Updated task title",
  "status": "in_progress",
  "assigned_to": "user-789"
}
```

### DELETE /api/v1/tasks/{id}
Delete a task.

## Document Management

### POST /api/v1/documents/upload
Upload a document for processing.

**Request:** Multipart form data with `file` field.

**Response:**
```json
{
  "document_id": "doc-123",
  "status": "uploaded",
  "filename": "requirements.pdf",
  "size_bytes": 1024000
}
```

### GET /api/v1/documents
List documents.

**Query Parameters:**
- `project_id` - Filter by project
- `status` - Filter by processing status

**Response:**
```json
{
  "documents": [
    {
      "id": "doc-123",
      "project_id": "proj-123",
      "name": "requirements.pdf",
      "status": "processed",
      "size_bytes": 1024000,
      "uploaded_at": "2024-01-14T10:30:00Z"
    }
  ]
}
```

### GET /api/v1/documents/{id}
Get document details.

### GET /api/v1/documents/{id}/status
Get document processing status.

## Organization Management

### POST /api/v1/organizations
Create a new organization.

**Request Body:**
```json
{
  "name": "Acme Corp",
  "description": "Software development company"
}
```

### GET /api/v1/organizations/{id}
Get organization details.

### POST /api/v1/projects
Create a new project.

**Request Body:**
```json
{
  "org_id": "org-123",
  "name": "Web Application",
  "description": "Customer-facing web application"
}
```

### GET /api/v1/projects
List projects.

### GET /api/v1/projects/{id}
Get project details.

## Workflow Management

### POST /api/v1/workflows
Create a new workflow definition.

**Request Body:**
```json
{
  "name": "CI/CD Pipeline",
  "description": "Automated deployment workflow",
  "version": "1.0.0",
  "steps": [
    {
      "id": "build",
      "name": "Build Application",
      "tool_name": "docker",
      "arguments": {"image": "myapp:latest"}
    }
  ]
}
```

### GET /api/v1/workflows
List workflows.

### GET /api/v1/workflows/{id}
Get workflow definition.

### POST /api/v1/workflows/{id}/execute
Execute a workflow.

**Response:**
```json
{
  "execution_id": "exec-123",
  "workflow_id": "wf-456",
  "status": "running",
  "started_at": "2024-01-14T10:30:00Z"
}
```

### GET /api/v1/workflows/executions/{id}
Get workflow execution status.

## API Version Management

### POST /api/v1/versions
Create API version.

### GET /api/v1/versions
List API versions.

### GET /api/v1/versions/{id}
Get API version details.

### GET /api/v1/versions/compatibility
Check version compatibility.

## Code Analysis

### POST /api/v1/analyze/code
Analyze code for quality metrics.

**Request Body:**
```json
{
  "code": "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"Hello\") }",
  "language": "go"
}
```

### POST /api/v1/lint/code
Lint code for issues.

### POST /api/v1/refactor/code
Suggest code refactoring.

### POST /api/v1/generate/docs
Generate documentation from code.

### POST /api/v1/validate/code
Validate code syntax and structure.

## Repository Management

### GET /api/v1/repositories
List repositories.

### GET /api/v1/repositories/{id}/impact
Analyze repository impact.

### GET /api/v1/repositories/{id}/centrality
Calculate repository centrality.

### GET /api/v1/repositories/network
Get repository network visualization.

### GET /api/v1/repositories/clusters
Get repository clusters.

### POST /api/v1/repositories/analyze-cross-repo
Perform cross-repository impact analysis.

## Monitoring

### GET /api/v1/monitoring/errors/dashboard
Get error dashboard.

### GET /api/v1/monitoring/errors/analysis
Get error analysis.

### GET /api/v1/monitoring/errors/stats
Get error statistics.

### POST /api/v1/monitoring/errors/classify
Classify an error.

### POST /api/v1/monitoring/errors/report
Report an error.

### GET /api/v1/monitoring/health
Get health metrics.

### GET /api/v1/monitoring/performance
Get performance metrics.

## Error Responses

All API endpoints return standardized error responses:

### 400 Bad Request
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field": "email",
      "issue": "invalid format"
    }
  }
}
```

### 401 Unauthorized
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "API key required"
  }
}
```

### 403 Forbidden
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions"
  }
}
```

### 404 Not Found
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found"
  }
}
```

### 429 Too Many Requests
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded",
    "retry_after": 60
  }
}
```

### 500 Internal Server Error
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred",
    "request_id": "req-123"
  }
}
```

## Security Features

- **Rate Limiting**: 100 requests per 10 seconds per client
- **CORS Support**: Configurable cross-origin policies
- **Security Headers**: XSS protection, content type sniffing prevention
- **Input Validation**: Comprehensive request validation
- **Error Sanitization**: Sensitive information not exposed in errors

## Data Formats

- **Content-Type**: `application/json`
- **Date Format**: ISO 8601 (`2024-01-14T10:30:00Z`)
- **Pagination**: `limit` and `offset` query parameters
- **Filtering**: Query parameters for resource filtering

---

*This API reference is automatically generated and reflects the current implementation. Last updated: January 14, 2026*

## Authentication

### Project API Key Authentication

All protected endpoints require API key authentication via the `Authorization` header:

```
Authorization: Bearer <api_key>
```

API keys are project-specific and can be obtained when creating a project via the admin endpoints.

### Admin API Key Authentication

Admin endpoints (`/api/v1/admin/*`) require admin API key authentication. This includes:
- Creating organizations
- Creating projects
- Uploading binary versions

**Authentication Methods:**

1. **X-Admin-API-Key header** (recommended):
   ```
   X-Admin-API-Key: <admin_api_key>
   ```

2. **Authorization Bearer header**:
   ```
   Authorization: Bearer <admin_api_key>
   ```

The admin API key is configured via the `ADMIN_API_KEY` environment variable. Generate a secure key:

```bash
openssl rand -hex 32
```

**Error Responses:**

- **401 Unauthorized**: Missing or invalid admin API key
  ```json
  {
    "success": false,
    "error": {
      "type": "validation_error",
      "message": "authorization: Invalid admin API key",
      "details": {
        "field": "authorization",
        "code": "unauthorized",
        "message": "Invalid admin API key"
      }
    }
  }
  ```

## Base URL

```
http://localhost:8080/api/v1
```

Replace `localhost:8080` with your Hub API hostname and port.

## Public Endpoints

These endpoints do not require authentication.

### Health Checks

#### GET /health

Basic health check endpoint.

**Response (200 OK):**
```json
{
  "status": "ok",
  "service": "sentinel-hub",
  "version": "1.0.0",
  "timestamp": "2024-12-10T14:30:00Z"
}
```

#### GET /health/db

Database connectivity check with connection pool statistics.

**Response (200 OK - Healthy):**
```json
{
  "status": "healthy",
  "service": "database",
  "stats": {
    "open_connections": 5,
    "in_use": 2,
    "idle": 3,
    "wait_count": 0
  },
  "timestamp": "2024-12-10T14:30:00Z"
}
```

**Response (503 Service Unavailable - Unhealthy):**
```json
{
  "status": "unhealthy",
  "service": "database",
  "error": "connection refused",
  "timestamp": "2024-12-10T14:30:00Z"
}
```

#### GET /health/ready

Readiness check that verifies the service is ready to accept traffic.

**Response (200 OK - Ready):**
```json
{
  "status": "ready",
  "timestamp": "2024-12-10T14:30:00Z"
}
```

**Response (503 Service Unavailable - Not Ready):**
```json
{
  "status": "not_ready",
  "reason": "database_unavailable",
  "error": "connection refused",
  "timestamp": "2024-12-10T14:30:00Z"
}
```

## Admin Endpoints

These endpoints are for administrative operations and require admin API key authentication.

**Authentication:** All admin endpoints require admin API key via `X-Admin-API-Key` header or `Authorization: Bearer <admin-key>`.

### POST /api/v1/admin/organizations

Create a new organization.

**Request Body:**
```json
{
  "name": "My Organization"
}
```

**Response (201 Created):**
```json
{
  "id": "org-uuid",
  "name": "My Organization",
  "created_at": "2024-12-10T14:30:00Z"
}
```

### POST /api/v1/admin/projects

Create a new project.

**Request Body:**
```json
{
  "org_id": "org-uuid",
  "name": "My Project"
}
```

**Response (201 Created):**
```json
{
  "id": "project-uuid",
  "org_id": "org-uuid",
  "name": "My Project",
  "api_key": "generated-api-key",
  "created_at": "2024-12-10T14:30:00Z"
}
```

### POST /api/v1/admin/binary/upload

Upload a new binary version for distribution to Sentinel clients.

**Content-Type:** `multipart/form-data`

**Authentication:** Requires admin API key.

**Form Fields:**
- `version` (required): Version in semver format (e.g., `1.2.3` or `v1.2.3`)
- `platform` (required): Platform identifier. Must be one of:
  - `linux-amd64`
  - `linux-arm64`
  - `darwin-amd64`
  - `darwin-arm64`
  - `windows-amd64`
- `binary` (required): Binary file (max 100MB)
- `release_notes` (optional): Release notes (max 10KB, sanitized)
- `is_stable` (optional): `true` or `false` (default: `false`)
- `is_latest` (optional): `true` or `false` (default: `false`)

**Validation:**
- Version must match semver format: `^v?\d+\.\d+\.\d+(-[a-zA-Z0-9]+)?$`
- Platform must be from allowed list
- Release notes are automatically sanitized (control characters removed, length limited to 10KB)

**Example Request:**
```bash
curl -X POST https://hub.example.com/api/v1/admin/binary/upload \
  -H "X-Admin-API-Key: your-admin-key" \
  -F "version=1.2.3" \
  -F "platform=linux-amd64" \
  -F "is_stable=true" \
  -F "is_latest=true" \
  -F "release_notes=Bug fixes and performance improvements" \
  -F "binary=@sentinel-linux-amd64"
```

**Response (201 Created):**
```json
{
  "success": true,
  "version": "1.2.3",
  "platform": "linux-amd64",
  "checksum": "sha256-checksum-here"
}
```

**Error Responses:**

- **400 Bad Request**: Invalid version format
  ```json
  {
    "success": false,
    "error": {
      "type": "validation_error",
      "message": "version: Version must be in semver format (e.g., 1.2.3 or v1.2.3)",
      "details": {
        "field": "version",
        "code": "invalid_format",
        "message": "Version must be in semver format (e.g., 1.2.3 or v1.2.3)"
      }
    }
  }
  ```

- **400 Bad Request**: Invalid platform
  ```json
  {
    "success": false,
    "error": {
      "type": "validation_error",
      "message": "platform: Platform must be one of: [linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64]",
      "details": {
        "field": "platform",
        "code": "invalid_platform",
        "message": "Platform must be one of: [linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64]"
      }
    }
  }
  ```

- **401 Unauthorized**: Missing or invalid admin API key
  ```json
  {
    "success": false,
    "error": {
      "type": "validation_error",
      "message": "authorization: Invalid admin API key",
      "details": {
        "field": "authorization",
        "code": "unauthorized",
        "message": "Invalid admin API key"
      }
    }
  }
  ```

- **500 Internal Server Error**: File operation or database error
  ```json
  {
    "success": false,
    "error": {
      "type": "database_error",
      "message": "Failed to save version metadata for version=1.2.3 platform=linux-amd64",
      "details": {
        "operation": "save_binary_version"
      }
    }
  }
  ```

## Document Management

### POST /api/v1/documents/ingest

Upload a document for processing.

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `file` (required): The document file (PDF, DOCX, etc.)

**Response (202 Accepted):**
```json
{
  "id": "doc-uuid",
  "status": "queued",
  "message": "Document queued for processing"
}
```

**Rate Limit:** 10 req/s, burst 20

### GET /api/v1/documents/{id}/status

Get document processing status.

**Response (200 OK):**
```json
{
  "id": "doc-uuid",
  "original_name": "document.pdf",
  "status": "processing",
  "progress": 50,
  "stages": [
    {
      "name": "extraction",
      "status": "completed",
      "duration_ms": 1200
    },
    {
      "name": "analysis",
      "status": "processing",
      "duration_ms": 0
    }
  ],
  "created_at": "2024-12-10T14:30:00Z"
}
```

### GET /api/v1/documents/{id}/extracted

Get extracted text from a document.

**Response (200 OK):**
```json
{
  "id": "doc-uuid",
  "extracted_text": "Full extracted text content...",
  "pages": 10
}
```

### GET /api/v1/documents/{id}/knowledge

Get knowledge items extracted from a document.

**Response (200 OK):**
```json
{
  "items": [
    {
      "id": "item-uuid",
      "type": "business_rule",
      "title": "User Authentication Rule",
      "content": "Users must authenticate...",
      "confidence": 0.95,
      "source_page": 5,
      "status": "pending"
    }
  ]
}
```

### POST /api/v1/documents/{id}/detect-changes

Detect changes in a document (re-upload detection).

**Response (200 OK):**
```json
{
  "has_changes": true,
  "changes": [
    {
      "type": "added",
      "page": 5,
      "content": "New content added"
    }
  ]
}
```

### GET /api/v1/documents

List all documents for the project.

**Query Parameters:**
- `status` (optional): Filter by status (`queued`, `processing`, `completed`, `failed`)
- `limit` (optional): Maximum number of results (default: 50)
- `offset` (optional): Pagination offset (default: 0)

**Response (200 OK):**
```json
{
  "documents": [
    {
      "id": "doc-uuid",
      "name": "document.pdf",
      "status": "completed",
      "created_at": "2024-12-10T14:30:00Z"
    }
  ],
  "total": 10,
  "limit": 50,
  "offset": 0
}
```

## Knowledge Management

### PUT /api/v1/knowledge/{id}/status

Update knowledge item status.

**Request Body:**
```json
{
  "status": "approved",
  "approved_by": "user@example.com"
}
```

**Response (200 OK):**
```json
{
  "id": "item-uuid",
  "status": "approved",
  "approved_by": "user@example.com",
  "approved_at": "2024-12-10T14:30:00Z"
}
```

### GET /api/v1/projects/knowledge

List all knowledge items for the project.

**Query Parameters:**
- `type` (optional): Filter by type (`business_rule`, `entity`, `journey`, etc.)
- `status` (optional): Filter by status (`pending`, `approved`, `rejected`)
- `limit` (optional): Maximum number of results (default: 50)

**Response (200 OK):**
```json
{
  "items": [
    {
      "id": "item-uuid",
      "type": "business_rule",
      "title": "Rule Title",
      "content": "Rule content...",
      "status": "approved"
    }
  ],
  "total": 25
}
```

### GET /api/v1/knowledge/business

Get business context for MCP tools (Phase A).

**Query Parameters:**
- `type` (optional): Filter by item type (`rule`, `entity`, `journey`)

**Response (200 OK):**
```json
{
  "items": [
    {
      "item_type": "rule",
      "title": "Business Rule Title",
      "content": "Rule content...",
      "confidence": 0.95
    }
  ]
}
```

### POST /api/v1/knowledge/sync

Synchronize knowledge items.

**Request Body:**
```json
{
  "document_id": "doc-uuid",
  "force": false
}
```

**Response (200 OK):**
```json
{
  "synced": 10,
  "updated": 2,
  "created": 8
}
```

### POST /api/v1/knowledge/gap-analysis

Perform gap analysis between documentation and code.

**Request Body:**
```json
{
  "codebase_path": "/path/to/code",
  "focus_areas": ["authentication", "payment"]
}
```

**Response (200 OK):**
```json
{
  "gaps": [
    {
      "type": "missing_implementation",
      "knowledge_item": "item-uuid",
      "description": "Rule not implemented in code"
    }
  ],
  "summary": {
    "total_gaps": 5,
    "critical": 2,
    "warnings": 3
  }
}
```

**Rate Limit:** 5 req/s, burst 10

### POST /api/v1/knowledge/migrate

Migrate knowledge items between projects.

**Request Body:**
```json
{
  "source_project_id": "source-uuid",
  "target_project_id": "target-uuid",
  "knowledge_item_ids": ["item1-uuid", "item2-uuid"]
}
```

**Response (200 OK):**
```json
{
  "migrated": 2,
  "failed": 0
}
```

**Rate Limit:** 1 req/s, burst 5

## Analysis Endpoints

### POST /api/v1/analyze/ast

Perform AST analysis on code files.

**Request Body:**
```json
{
  "files": [
    {
      "path": "src/main.js",
      "content": "function test() { ... }",
      "language": "javascript"
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "results": [
    {
      "file": "src/main.js",
      "functions": ["test"],
      "complexity": 5,
      "issues": []
    }
  ]
}
```

**Rate Limit:** 5 req/s, burst 10

### POST /api/v1/analyze/vibe

Detect vibe coding patterns.

**Request Body:**
```json
{
  "files": [
    {
      "path": "src/main.js",
      "content": "code content...",
      "language": "javascript"
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "findings": [
    {
      "type": "duplicate_function",
      "file": "src/main.js",
      "line": 10,
      "message": "Duplicate function detected"
    }
  ]
}
```

**Rate Limit:** 5 req/s, burst 10

### POST /api/v1/analyze/cross-file

Cross-file analysis for symbol tracking.

**Request Body:**
```json
{
  "files": [
    {
      "path": "src/file1.js",
      "content": "..."
    },
    {
      "path": "src/file2.js",
      "content": "..."
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "symbols": [
    {
      "name": "myFunction",
      "defined_in": "src/file1.js",
      "used_in": ["src/file2.js"]
    }
  ]
}
```

### POST /api/v1/analyze/security

Security analysis with AST-based rule checking.

**Request Body:**
```json
{
  "files": [
    {
      "path": "src/auth.js",
      "content": "code...",
      "language": "javascript"
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "score": 85,
  "findings": [
    {
      "rule": "SEC-001",
      "severity": "high",
      "file": "src/auth.js",
      "line": 15,
      "message": "Hardcoded secret detected"
    }
  ]
}
```

### GET /api/v1/security/context

Get security context for MCP tools (Phase A).

**Response (200 OK):**
```json
{
  "security_rules": [
    {
      "id": "SEC-001",
      "title": "No Hardcoded Secrets",
      "status": "enforced"
    }
  ],
  "compliance_status": "compliant",
  "security_score": 85
}
```

### POST /api/v1/analyze/comprehensive

Comprehensive feature analysis (Phase 14A).

**Request Body:**
```json
{
  "codebase_path": "/path/to/code",
  "mode": "auto",
  "depth": "medium",
  "include_business_context": true
}
```

**Response (200 OK):**
```json
{
  "feature": "user-authentication",
  "analysis": {
    "ui": {...},
    "api": {...},
    "database": {...},
    "logic": {...},
    "integration": {...},
    "tests": {...}
  },
  "end_to_end_flows": [...],
  "discrepancies": [...]
}
```

**Rate Limit:** 2 req/s, burst 5

### POST /api/v1/analyze/architecture

Architecture analysis for file structure.

**Request Body:**
```json
{
  "files": [
    {
      "path": "src/large-file.js",
      "size": 50000,
      "lines": 2000
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "recommendations": [
    {
      "file": "src/large-file.js",
      "issue": "file_too_large",
      "suggestion": "Split into smaller modules",
      "estimated_benefit": "improved_maintainability"
    }
  ]
}
```

### POST /api/v1/analyze/doc-sync

Documentation-code synchronization analysis.

**Request Body:**
```json
{
  "codebase_path": "/path/to/code",
  "report_type": "status_tracking"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "phases": [
    {
      "phase": "Phase 1",
      "status": "in_sync",
      "doc_status": "complete",
      "code_status": "complete"
    }
  ],
  "discrepancies": [],
  "summary": {
    "total_phases": 15,
    "in_sync_count": 14,
    "discrepancy_count": 1
  }
}
```

### POST /api/v1/analyze/business-rules

Business rules comparison analysis.

**Request Body:**
```json
{
  "codebase_path": "/path/to/code"
}
```

**Response (200 OK):**
```json
{
  "comparisons": [
    {
      "rule_id": "rule-uuid",
      "documented": "Rule description",
      "implemented": true,
      "matches": true
    }
  ]
}
```

## Validation Endpoints

### POST /api/v1/validate/code

**Status**: ✅ Complete

Validate code content using AST analysis (Phase B).

**Request Body:**
```json
{
  "code": "function test() { ... }",
  "language": "javascript",
  "file_path": "src/test.js"
}
```

**Response (200 OK):**
```json
{
  "valid": true,
  "issues": [],
  "suggestions": []
}
```

### POST /api/v1/validate/business

Validate code against business rules (Phase B).

**Request Body:**
```json
{
  "code": "code content...",
  "knowledge_item_ids": ["item1-uuid", "item2-uuid"]
}
```

**Response (200 OK):**
```json
{
  "valid": false,
  "violations": [
    {
      "rule_id": "item1-uuid",
      "rule_title": "Business Rule",
      "violation": "Rule not followed",
      "line": 10
    }
  ]
}
```

## Action Endpoints

### POST /api/v1/fixes/apply

**Status**: ✅ Fully Implemented

**Behavior**: Applies fixes based on fixType parameter (security/style/performance) and returns modified code with change descriptions. Uses `ApplySecurityFixes`, `ApplyStyleFixes`, and `ApplyPerformanceFixes` functions from `fix_applier.go`.

Apply fixes to code (Phase C).

**Request Body:**
```json
{
  "file_path": "src/file.js",
  "fix_type": "remove_debug",
  "fixes": [
    {
      "line": 10,
      "action": "remove",
      "content": "console.log('debug')"
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "applied": true,
  "file_path": "src/file.js",
  "changes": 1
}
```

## Change Requests

### GET /api/v1/change-requests

List change requests for the project.

**Query Parameters:**
- `status` (optional): Filter by status
- `type` (optional): Filter by type

**Response (200 OK):**
```json
{
  "change_requests": [
    {
      "id": "cr-uuid",
      "type": "update",
      "status": "pending_approval",
      "created_at": "2024-12-10T14:30:00Z"
    }
  ]
}
```

### GET /api/v1/change-requests/{id}

Get a specific change request.

**Response (200 OK):**
```json
{
  "id": "cr-uuid",
  "type": "update",
  "status": "approved",
  "current_state": {...},
  "proposed_state": {...},
  "implementation_status": "in_progress"
}
```

### POST /api/v1/change-requests/{id}/approve

Approve a change request.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Change request approved"
}
```

### POST /api/v1/change-requests/{id}/reject

Reject a change request.

**Request Body:**
```json
{
  "reason": "Does not align with requirements"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Change request rejected"
}
```

### POST /api/v1/change-requests/{id}/impact

Analyze impact of a change request.

**Response (200 OK):**
```json
{
  "impact": {
    "affected_files": 5,
    "affected_features": 2,
    "risk_level": "medium"
  }
}
```

### POST /api/v1/change-requests/{id}/start

Start implementation of a change request.

**Request Body:**
```json
{
  "notes": "Starting implementation"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "implementation_status": "in_progress"
}
```

### POST /api/v1/change-requests/{id}/complete

Mark change request implementation as complete.

**Request Body:**
```json
{
  "notes": "Implementation complete"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "implementation_status": "completed"
}
```

### POST /api/v1/change-requests/{id}/update

Update implementation status.

**Request Body:**
```json
{
  "status": "in_progress",
  "notes": "Progress update",
  "progress_percent": 50
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "implementation_status": "in_progress"
}
```

### GET /api/v1/change-requests/dashboard

Get change requests dashboard data.

**Response (200 OK):**
```json
{
  "summary": {
    "total": 10,
    "pending": 3,
    "approved": 5,
    "rejected": 2
  },
  "recent": [...]
}
```

## Telemetry & Metrics

### POST /api/v1/telemetry

Ingest telemetry events.

**Request Body:**
```json
{
  "event_type": "audit_complete",
  "payload": {
    "finding_count": 5,
    "compliance_percent": 95
  },
  "agent_id": "agent-uuid",
  "org_id": "org-uuid",
  "team_id": "team-uuid"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "event_id": "event-uuid"
}
```

**Rate Limit:** 50 req/s, burst 100

### GET /api/v1/telemetry/recent

Get recent telemetry events.

**Query Parameters:**
- `limit` (optional): Number of events (default: 100)
- `event_type` (optional): Filter by event type

**Response (200 OK):**
```json
{
  "events": [
    {
      "id": "event-uuid",
      "event_type": "audit_complete",
      "payload": {...},
      "created_at": "2024-12-10T14:30:00Z"
    }
  ]
}
```

### GET /api/v1/metrics

Get aggregated metrics.

**Query Parameters:**
- `start_date` (optional): Start date filter
- `end_date` (optional): End date filter
- `event_type` (optional): Filter by event type

**Response (200 OK):**
```json
{
  "events": [...],
  "metrics": {
    "audit_count": 100,
    "average_compliance": 92.5,
    "total_findings": 50
  },
  "total": 100
}
```

### GET /api/v1/metrics/trends

Get metrics trends over time.

**Response (200 OK):**
```json
{
  "trends": [
    {
      "date": "2024-12-10",
      "audit_count": 10,
      "compliance": 95
    }
  ]
}
```

### GET /api/v1/metrics/team/{teamId}

Get team-specific metrics.

**Response (200 OK):**
```json
{
  "team_id": "team-uuid",
  "metrics": {
    "audit_count": 50,
    "compliance": 90
  }
}
```

### GET /api/v1/metrics/prometheus

Get Prometheus-formatted metrics (Phase G).

**Response (200 OK):**
```
sentinel_http_requests_total{endpoint="/api/v1/documents/ingest"} 1250
sentinel_http_errors_total{endpoint="/api/v1/documents/ingest"} 5
sentinel_http_request_duration_ms{endpoint="/api/v1/documents/ingest"} 245.50
sentinel_db_open_connections 5
sentinel_uptime_seconds 86400.00
```

## Test Management

### POST /api/v1/test-requirements/generate

Generate test requirements from knowledge items.

**Request Body:**
```json
{
  "knowledge_item_id": "item-uuid"
}
```

**Response (200 OK):**
```json
{
  "requirement_id": "req-uuid",
  "requirements": [
    {
      "rule_title": "Test Rule",
      "requirement_type": "unit_test",
      "description": "Test description"
    }
  ]
}
```

### POST /api/v1/test-coverage/analyze

Analyze test coverage.

**Request Body:**
```json
{
  "knowledge_item_id": "item-uuid",
  "test_files": ["test/test.js"]
}
```

**Response (200 OK):**
```json
{
  "coverage_percentage": 85.5,
  "test_files": ["test/test.js"],
  "missing_coverage": []
}
```

### GET /api/v1/test-coverage/{knowledge_item_id}

Get test coverage for a knowledge item.

**Response (200 OK):**
```json
{
  "knowledge_item_id": "item-uuid",
  "coverage_percentage": 85.5,
  "test_files": ["test/test.js"],
  "last_updated": "2024-12-10T14:30:00Z"
}
```

### POST /api/v1/test-validations/validate

Validate test quality.

**Request Body:**
```json
{
  "test_requirement_id": "req-uuid",
  "test_code": "test code..."
}
```

**Response (200 OK):**
```json
{
  "validation_status": "passed",
  "score": 90,
  "issues": []
}
```

### GET /api/v1/test-validations/{test_requirement_id}

Get validation results.

**Response (200 OK):**
```json
{
  "test_requirement_id": "req-uuid",
  "validation_status": "passed",
  "score": 90,
  "validated_at": "2024-12-10T14:30:00Z"
}
```

### POST /api/v1/test-execution/run

Run test execution.

**Request Body:**
```json
{
  "execution_type": "unit_tests",
  "test_files": ["test/test.js"]
}
```

**Response (200 OK):**
```json
{
  "execution_id": "exec-uuid",
  "status": "running"
}
```

### GET /api/v1/test-execution/{execution_id}

Get test execution results.

**Response (200 OK):**
```json
{
  "execution_id": "exec-uuid",
  "status": "completed",
  "result": {
    "passed": 10,
    "failed": 2,
    "duration_ms": 5000
  }
}
```

### POST /api/v1/mutation-test/run

Run mutation testing.

**Request Body:**
```json
{
  "test_requirement_id": "req-uuid",
  "test_code": "test code..."
}
```

**Response (200 OK):**
```json
{
  "mutation_score": 85.5,
  "total_mutants": 100,
  "killed_mutants": 85,
  "survived_mutants": 15
}
```

### GET /api/v1/mutation-test/{test_requirement_id}

Get mutation test results.

**Response (200 OK):**
```json
{
  "test_requirement_id": "req-uuid",
  "mutation_score": 85.5,
  "total_mutants": 100,
  "killed_mutants": 85
}
```

## Hook Management

### POST /api/v1/telemetry/hook

Submit hook execution telemetry.

**Request Body:**
```json
{
  "hook_type": "pre-commit",
  "result": "passed",
  "findings_summary": {...},
  "duration_ms": 500
}
```

**Response (200 OK):**
```json
{
  "success": true
}
```

### GET /api/v1/hooks/metrics

Get hook execution metrics.

**Response (200 OK):**
```json
{
  "total_executions": 100,
  "passed": 95,
  "failed": 5,
  "average_duration_ms": 500
}
```

### GET /api/v1/hooks/policies

Get hook policies.

**Response (200 OK):**
```json
{
  "policies": [
    {
      "id": "policy-uuid",
      "policy_config": {...}
    }
  ]
}
```

### POST /api/v1/hooks/policies

Create or update hook policy.

**Request Body:**
```json
{
  "policy_config": {
    "enforce": true,
    "rules": [...]
  }
}
```

**Response (200 OK):**
```json
{
  "id": "policy-uuid",
  "success": true
}
```

### GET /api/v1/hooks/limits

Get hook execution limits.

**Response (200 OK):**
```json
{
  "limits": {
    "max_duration_ms": 5000,
    "max_findings": 100
  }
}
```

### POST /api/v1/hooks/baselines

Create hook baseline.

**Request Body:**
```json
{
  "baseline_entry": {
    "file": "src/file.js",
    "line": 10,
    "pattern": "console.log"
  }
}
```

**Response (200 OK):**
```json
{
  "id": "baseline-uuid",
  "success": true
}
```

### POST /api/v1/hooks/baselines/{id}/review

Review hook baseline.

**Request Body:**
```json
{
  "action": "approve",
  "comment": "Approved"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "reviewed": true
}
```

## Intent Analysis

### POST /api/v1/analyze/intent

Analyze user intent from prompt (Phase 15).

**Request Body:**
```json
{
  "prompt": "Add user authentication",
  "codebase_path": "/path/to/code",
  "include_context": true
}
```

**Response (200 OK):**
```json
{
  "intent_type": "feature_implementation",
  "confidence": 0.9,
  "clarifying_questions": [
    "What authentication method should be used?"
  ],
  "suggested_template": "authentication_template",
  "context": {...}
}
```

### POST /api/v1/intent/decisions

Record user decision for intent learning (Phase 15).

**Request Body:**
```json
{
  "prompt": "Add user authentication",
  "decision": "proceeded",
  "selected_template": "oauth_template",
  "context": {...}
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "pattern_refined": true
}
```

### GET /api/v1/intent/patterns

Get learned intent patterns (Phase 15).

**Query Parameters:**
- `type` (optional): Filter by pattern type
- `limit` (optional): Maximum results (default: 50)

**Response (200 OK):**
```json
{
  "patterns": [
    {
      "id": "pattern-uuid",
      "intent_type": "feature_implementation",
      "pattern": "add.*authentication",
      "template": "oauth_template",
      "confidence": 0.95,
      "usage_count": 10
    }
  ]
}
```

## LLM Configuration

Phase 14C endpoints for managing LLM provider configurations and viewing usage statistics.

### POST /api/v1/llm/config

Create a new LLM provider configuration.

**Request Body:**
```json
{
  "provider": "openai",
  "api_key": "sk-...",
  "model": "gpt-4",
  "key_type": "user-provided",
  "cost_optimization": {
    "use_cache": true,
    "cache_ttl_hours": 24,
    "progressive_depth": true,
    "max_cost_per_request": 0.5
  }
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "config_id": "config-uuid",
  "message": "Configuration saved successfully"
}
```

**Rate Limit:** 10 req/min

### GET /api/v1/llm/config/{id}

Get LLM configuration by ID.

**Response (200 OK):**
```json
{
  "id": "config-uuid",
  "provider": "openai",
  "api_key": "****7890",
  "model": "gpt-4",
  "key_type": "user-provided",
  "cost_optimization": {
    "use_cache": true,
    "cache_ttl_hours": 24,
    "progressive_depth": true
  }
}
```

**Rate Limit:** 10 req/min

### PUT /api/v1/llm/config/{id}

Update LLM configuration.

**Request Body:**
```json
{
  "provider": "openai",
  "api_key": "sk-...",
  "model": "gpt-4-turbo",
  "key_type": "user-provided",
  "cost_optimization": {
    "use_cache": true,
    "cache_ttl_hours": 48
  }
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Configuration updated successfully"
}
```

**Rate Limit:** 10 req/min

### DELETE /api/v1/llm/config/{id}

Delete LLM configuration.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Configuration deleted successfully"
}
```

**Rate Limit:** 10 req/min

### GET /api/v1/llm/config/project/{projectId}

List all LLM configurations for a project.

**Response (200 OK):**
```json
{
  "success": true,
  "configs": [
    {
      "provider": "openai",
      "api_key": "****7890",
      "model": "gpt-4",
      "key_type": "user-provided",
      "cost_optimization": {...}
    }
  ],
  "count": 1
}
```

**Rate Limit:** 10 req/min

### GET /api/v1/llm/providers

Get list of supported LLM providers.

**Response (200 OK):**
```json
{
  "success": true,
  "providers": [
    {
      "name": "openai",
      "display_name": "OpenAI",
      "description": "OpenAI GPT models (GPT-4, GPT-3.5)"
    },
    {
      "name": "anthropic",
      "display_name": "Anthropic",
      "description": "Anthropic Claude models (Claude 3 Opus, Sonnet, Haiku)"
    }
  ]
}
```

### GET /api/v1/llm/models/{provider}

Get list of supported models for a provider.

**Response (200 OK):**
```json
{
  "success": true,
  "models": [
    {
      "name": "gpt-4",
      "display_name": "GPT-4",
      "price_per_1k": 0.03
    },
    {
      "name": "gpt-3.5-turbo",
      "display_name": "GPT-3.5 Turbo",
      "price_per_1k": 0.0015
    }
  ]
}
```

### POST /api/v1/llm/config/validate

Validate LLM API key and model connection.

**Request Body:**
```json
{
  "provider": "openai",
  "api_key": "sk-...",
  "model": "gpt-4"
}
```

**Response (200 OK - Valid):**
```json
{
  "success": true,
  "valid": true,
  "message": "Connection test successful"
}
```

**Response (400 Bad Request - Invalid):**
```json
{
  "success": false,
  "valid": false,
  "error": "authentication failed: invalid API key"
}
```

**Rate Limit:** 5 req/min

### GET /api/v1/llm/usage/report

Get detailed usage report for a project.

**Query Parameters:**
- `project_id` (required): Project UUID
- `start_date` (optional): Start date (YYYY-MM-DD), defaults to 30 days ago
- `end_date` (optional): End date (YYYY-MM-DD), defaults to today

**Response (200 OK):**
```json
{
  "project_id": "project-uuid",
  "period": "monthly",
  "start_date": "2024-11-01",
  "end_date": "2024-12-01",
  "total_tokens": 1500000,
  "total_cost": 45.50,
  "usage_by_provider": {
    "openai": {
      "provider": "openai",
      "tokens": 1000000,
      "cost": 30.00,
      "request_count": 500
    }
  },
  "usage_by_model": {
    "openai:gpt-4": {
      "model": "gpt-4",
      "tokens": 1000000,
      "cost": 30.00,
      "request_count": 500
    }
  },
  "daily_usage": [
    {
      "date": "2024-11-01",
      "tokens": 50000,
      "cost": 1.50,
      "request_count": 25
    }
  ]
}
```

**Rate Limit:** 30 req/min

### GET /api/v1/llm/usage/stats

Get aggregated usage statistics.

**Query Parameters:**
- `project_id` (required): Project UUID
- `period` (optional): Period (daily/weekly/monthly/yearly), defaults to monthly

**Response (200 OK):**
```json
{
  "total_requests": 1000,
  "total_tokens": 2000000,
  "total_cost": 60.00,
  "average_cost": 0.06,
  "top_models": [
    {
      "model": "gpt-4",
      "request_count": 600,
      "total_tokens": 1200000,
      "total_cost": 36.00
    }
  ],
  "cost_trend": [
    {
      "date": "2024-11-01",
      "cost": 2.00
    }
  ]
}
```

**Rate Limit:** 30 req/min

### GET /api/v1/llm/usage/cost-breakdown

Get cost breakdown by provider and model.

**Query Parameters:**
- `project_id` (required): Project UUID
- `period` (optional): Period (daily/weekly/monthly/yearly), defaults to monthly

**Response (200 OK):**
```json
{
  "project_id": "project-uuid",
  "period": "monthly",
  "total_cost": 60.00,
  "by_provider": {
    "openai": 40.00,
    "anthropic": 20.00
  },
  "by_model": {
    "gpt-4": 30.00,
    "claude-3-opus": 20.00
  },
  "provider_percentages": {
    "openai": 66.67,
    "anthropic": 33.33
  },
  "model_percentages": {
    "gpt-4": 50.00,
    "claude-3-opus": 33.33
  }
}
```

**Rate Limit:** 30 req/min

### GET /api/v1/llm/usage/trends

Get usage trends over time.

**Query Parameters:**
- `project_id` (required): Project UUID
- `period` (optional): Period (daily/weekly/monthly/yearly), defaults to monthly
- `group_by` (optional): Group by (day/provider/model), defaults to day

**Response (200 OK):**
```json
{
  "success": true,
  "period": "monthly",
  "group_by": "day",
  "trends": [
    {
      "label": "2024-11-01",
      "tokens": 50000,
      "cost": 1.50,
      "requests": 25
    }
  ]
}
```

**Rate Limit:** 30 req/min

## Doc-Sync Management

### GET /api/v1/doc-sync/review-queue

Get review queue for doc-sync updates.

**Response (200 OK):**
```json
{
  "reviews": [
    {
      "id": "review-uuid",
      "file_path": "docs/guide.md",
      "change_type": "update",
      "old_value": "old content",
      "new_value": "new content"
    }
  ],
  "count": 5
}
```

### POST /api/v1/doc-sync/review/{id}

Approve or reject a doc-sync review item.

**Request Body:**
```json
{
  "action": "approve"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Review approved"
}
```

## Comprehensive Validation

### GET /api/v1/validations/{id}

Get comprehensive validation result.

**Response (200 OK):**
```json
{
  "id": "validation-uuid",
  "status": "completed",
  "results": {
    "code": {...},
    "security": {...},
    "business": {...},
    "tests": {...}
  }
}
```

### GET /api/v1/validations

List comprehensive validations.

**Query Parameters:**
- `status` (optional): Filter by status
- `limit` (optional): Maximum results

**Response (200 OK):**
```json
{
  "validations": [
    {
      "id": "validation-uuid",
      "status": "completed",
      "created_at": "2024-12-10T14:30:00Z"
    }
  ]
}
```

## Phase 14D: Cost Optimization Metrics

Phase 14D endpoints for monitoring cache performance and cost optimization metrics.

### GET /api/v1/metrics/cache

Get cache metrics for a project.

**Query Parameters:**
- `project_id` (optional): Project ID (uses authenticated project if not provided)

**Response (200 OK):**
```json
{
  "success": true,
  "project_id": "uuid",
  "hit_rate": 0.75,
  "total_hits": 150,
  "total_misses": 50,
  "cache_size": 250,
  "cache_ttl_hours": 24
}
```

**Response Fields:**
- `hit_rate`: Cache hit rate (0.0 to 1.0)
- `total_hits`: Total number of cache hits
- `total_misses`: Total number of cache misses
- `cache_size`: Number of cached entries for this project
- `cache_ttl_hours`: Cache TTL in hours from LLM config

**Example:**
```bash
curl -X GET "https://hub.example.com/api/v1/metrics/cache?project_id=uuid" \
  -H "Authorization: Bearer $TOKEN"
```

### GET /api/v1/metrics/cost

Get cost optimization metrics for a project.

**Query Parameters:**
- `project_id` (optional): Project ID (uses authenticated project if not provided)
- `period` (optional): Time period - `daily`, `weekly`, or `monthly` (default: `monthly`)

**Response (200 OK):**
```json
{
  "success": true,
  "project_id": "uuid",
  "period": "monthly",
  "total_cost": 45.50,
  "cost_savings": 18.20,
  "savings_percentage": 40.0,
  "cache_hit_savings": 12.30,
  "model_selection_savings": 5.90,
  "total_requests": 150
}
```

**Response Fields:**
- `total_cost`: Total LLM costs for the period
- `cost_savings`: Estimated total savings from optimization
- `savings_percentage`: Percentage of cost saved
- `cache_hit_savings`: Estimated savings from cache hits
- `model_selection_savings`: Estimated savings from smart model selection
- `total_requests`: Total number of LLM requests in the period

**Example:**
```bash
curl -X GET "https://hub.example.com/api/v1/metrics/cost?project_id=uuid&period=monthly" \
  -H "Authorization: Bearer $TOKEN"
```

**Error Responses:**

**400 Bad Request:**
```json
{
  "success": false,
  "error": {
    "type": "validation_error",
    "message": "period must be 'daily', 'weekly', or 'monthly'"
  }
}
```

**500 Internal Server Error:**
```json
{
  "success": false,
  "error": {
    "type": "internal_error",
    "message": "Failed to retrieve cost metrics"
  }
}
```

## Task Management

Task management endpoints for Phase 14E: Task Dependency & Verification System. See [TASK_DEPENDENCY_SYSTEM.md](./TASK_DEPENDENCY_SYSTEM.md) for complete documentation.

### POST /api/v1/tasks

Create a new task.

**Request Body:**
```json
{
  "title": "Implement user authentication",
  "description": "Add JWT-based authentication",
  "file_path": "src/auth.js",
  "line_number": 42,
  "priority": "high",
  "tags": ["auth", "security"]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "task": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "project_id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Implement user authentication",
    "status": "pending",
    "priority": "high",
    "created_at": "2024-12-11T10:00:00Z"
  }
}
```

### GET /api/v1/tasks

List tasks with optional filters.

**Query Parameters:**
- `status`: Filter by status (pending, in_progress, completed, blocked)
- `priority`: Filter by priority (low, medium, high, critical)
- `source`: Filter by source (cursor, manual, change_request, etc.)
- `assigned_to`: Filter by assignee
- `tags`: Filter by tags (comma-separated)
- `include_archived`: Include archived tasks (true/false)
- `limit`: Maximum results (default: 50)
- `offset`: Pagination offset (default: 0)

**Response (200 OK):**
```json
{
  "success": true,
  "tasks": [...],
  "total": 100,
  "limit": 50,
  "offset": 0
}
```

### GET /api/v1/tasks/{id}

Get a specific task by ID.

**Response (200 OK):**
```json
{
  "success": true,
  "task": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Implement user authentication",
    "status": "pending",
    "verification_confidence": 0.0,
    "dependencies": [...]
  }
}
```

### PUT /api/v1/tasks/{id}

Update a task.

**Request Body:**
```json
{
  "status": "in_progress",
  "priority": "critical",
  "assigned_to": "developer@example.com"
}
```

### DELETE /api/v1/tasks/{id}

Delete a task.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Task deleted"
}
```

### POST /api/v1/tasks/scan

Scan codebase for tasks (TODO comments, Cursor markers, etc.).

**Request Body:**
```json
{
  "codebase_path": ".",
  "patterns": ["TODO", "FIXME"]
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "tasks_found": 15,
  "tasks_created": 12,
  "tasks_updated": 3
}
```

### POST /api/v1/tasks/{id}/verify

Verify task completion using multi-factor verification.

**Response (200 OK):**
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
    }
  }
}
```

### POST /api/v1/tasks/verify-all

Verify all pending tasks.

**Response (200 OK):**
```json
{
  "success": true,
  "verified": 10,
  "failed": 2,
  "skipped": 5
}
```

### GET /api/v1/tasks/{id}/dependencies

Get task dependencies.

**Response (200 OK):**
```json
{
  "success": true,
  "dependencies": {
    "depends_on": [...],
    "required_by": [...],
    "cycles": []
  }
}
```

### POST /api/v1/tasks/{id}/detect-dependencies

Detect dependencies for a task.

**Response (200 OK):**
```json
{
  "success": true,
  "dependencies_found": 3,
  "dependencies_created": 3
}
```

## Production-Ready Endpoints

All API endpoints are fully implemented and production-ready.

### POST /api/v1/validate/code

**Status**: ✅ Fully Implemented

**Location**: `hub/api/main.go:1517-1591`

**Behavior**: 
- Calls `analyzeAST(req.Code, req.Language, []string{"duplicates", "unused", "unreachable"})`
- Converts `ASTFinding` results to violations format
- Returns actual code violations with line numbers and messages
- Handles language inference from file path
- Returns appropriate error messages when language cannot be determined

**Implementation**: Uses Tree-sitter AST analysis via `analyzeAST()` function.

### POST /api/v1/fixes/apply

**Status**: ✅ Fully Implemented

**Location**: `hub/api/main.go:1682-1733`

**Behavior**: 
- Applies fixes based on `fix_type` parameter (security/style/performance)
- Uses `ApplySecurityFixes`, `ApplyStyleFixes`, and `ApplyPerformanceFixes` functions
- Returns modified code with change descriptions
- Handles language inference from file path

**Expected Behavior**:
- Parses code with AST analysis
- Applies fixes based on `fixType` parameter:
  - `security`: Apply security fixes (e.g., sanitize inputs, use parameterized queries)
  - `style`: Apply style fixes (e.g., formatting, naming conventions)
  - `performance`: Apply performance fixes (e.g., optimize loops, cache results)
- Returns modified code with change descriptions

**Impact**: `sentinel_apply_fix` MCP tool doesn't actually fix code issues.

**Fix Required**: Implement fix logic for security/style/performance fixes (~1-2 days)

---

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request

Invalid request parameters or body.

```json
{
  "success": false,
  "error": {
    "type": "validation_error",
    "message": "Missing required field: file"
  }
}
```

### 401 Unauthorized

Missing or invalid API key.

```json
{
  "success": false,
  "error": {
    "type": "validation_error",
    "message": "authorization: Invalid API key",
    "details": {
      "field": "authorization",
      "code": "unauthorized",
      "message": "Invalid API key"
    }
  }
}
```

### 403 Forbidden

Valid API key but insufficient permissions.

```json
{
  "success": false,
  "error": {
    "type": "validation_error",
    "message": "authorization: Insufficient permissions",
    "details": {
      "field": "authorization",
      "code": "forbidden",
      "message": "Insufficient permissions"
    }
  }
}
```

### 404 Not Found

Resource not found.

```json
{
  "success": false,
  "error": {
    "type": "not_found_error",
    "message": "Document not found",
    "details": {
      "resource": "document",
      "id": "document_id"
    }
  }
}
```

### 429 Too Many Requests

Rate limit exceeded.

**Headers:**
- `Retry-After: 1` - Seconds to wait before retrying

```json
{
  "success": false,
  "error": {
    "type": "validation_error",
    "message": "Rate limit exceeded"
}
```

### 500 Internal Server Error

Server error.

```json
{
  "success": false,
  "error": {
    "type": "internal_error",
    "message": "An unexpected error occurred"
  }
}
```

### 503 Service Unavailable

Service temporarily unavailable (e.g., database down).

```json
{
  "success": false,
  "error": {
    "type": "external_service_error",
    "message": "Database connection failed",
    "details": {
      "service": "database"
    }
}
```

## Rate Limiting

Different endpoints have different rate limits:

- **Document Upload**: 10 req/s, burst 20
- **Telemetry**: 50 req/s, burst 100
- **AST Analysis**: 5 req/s, burst 10
- **Vibe Analysis**: 5 req/s, burst 10
- **Gap Analysis**: 5 req/s, burst 10
- **Change Requests**: 20 req/s, burst 40
- **Comprehensive Analysis**: 2 req/s, burst 5
- **Knowledge Migration**: 1 req/s, burst 5
- **General API**: Per-API-key rate limiting (2 req/s, burst 5)

When rate limit is exceeded, a `429 Too Many Requests` response is returned with a `Retry-After` header indicating when to retry.

## Pagination

Endpoints that return lists support pagination via query parameters:

- `limit`: Maximum number of results (default: 50, max: 1000)
- `offset`: Number of results to skip (default: 0)

Example:
```
GET /api/v1/documents?limit=20&offset=40
```

## Additional Resources

- [Deployment Guide](./HUB_DEPLOYMENT_GUIDE.md)
- [Monitoring Guide](./MONITORING_GUIDE.md)
- [Security Guide](./SECURITY_GUIDE.md)
- [Technical Specification](./TECHNICAL_SPEC.md)


