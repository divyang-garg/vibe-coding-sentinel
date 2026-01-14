# Phase 12: Requirements Lifecycle Management Guide

## Overview

Phase 12 provides comprehensive requirements lifecycle management capabilities, tracking requirement changes and ensuring code stays in sync with documented business rules. This phase enables automatic detection of discrepancies between documentation and code, manages change requests, and tracks implementation status.

## Key Features

- **Gap Analysis**: Identify discrepancies between documented business rules and code implementation
- **Change Detection**: Automatically detect changes when documents are re-ingested
- **Change Request Workflow**: Manage approval/rejection of requirement changes
- **Impact Analysis**: Analyze the impact of changes on code and tests
- **Implementation Tracking**: Monitor the status of implementing approved changes

## Gap Analysis

### What is Gap Analysis?

Gap analysis identifies discrepancies between documented business rules (in the knowledge base) and actual code implementation. It detects four types of gaps:

1. **Missing Implementation** (`missing_impl`): Business rule is documented but not implemented in code
2. **Missing Documentation** (`missing_doc`): Code exists but is not documented as a business rule
3. **Partial Match** (`partial_match`): Business rule is partially implemented
4. **Tests Missing** (`tests_missing`): Business rule has no test coverage

### How to Use Gap Analysis

#### Command Line

```bash
# Run gap analysis
sentinel knowledge gap-analysis --project-id <project-id> --codebase-path ./src

# Include test coverage check
sentinel knowledge gap-analysis --project-id <project-id> --codebase-path ./src --include-tests

# Enable reverse check (find undocumented code)
sentinel knowledge gap-analysis --project-id <project-id> --codebase-path ./src --reverse-check

# Output as JSON
sentinel knowledge gap-analysis --project-id <project-id> --codebase-path ./src --format json
```

#### API Endpoint

```http
POST /api/v1/knowledge/gap-analysis
Content-Type: application/json
Authorization: Bearer <token>

{
  "projectId": "uuid",
  "codebasePath": "/path/to/codebase",
  "options": {
    "includeTests": true,
    "reverseCheck": true
  }
}
```

**Response Format**:

```json
{
  "success": true,
  "report_id": "550e8400-e29b-41d4-a716-446655440000",
  "report": {
    "project_id": "uuid",
    "gaps": [ ... ],
    "summary": { ... },
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Response Fields**:
- `success`: Boolean indicating if the request was successful
- `report_id`: UUID of the stored gap report (empty string if storage failed)
- `report`: The gap analysis report object

### Understanding Gap Reports

A gap report contains:

- **Gaps**: Array of gap objects, each containing:
  - `type`: Type of gap (missing_impl, missing_doc, partial_match, tests_missing)
  - `knowledge_item_id`: ID of the related business rule (if applicable)
  - `rule_title`: Title of the business rule
  - `file_path`: File path where gap was detected (for missing_doc)
  - `line_number`: Line number where gap was detected (for missing_doc)
  - `description`: Human-readable description of the gap
  - `recommendation`: Suggested action to resolve the gap
  - `severity`: Severity level (critical, high, medium, low)
- **Summary**: Statistics about gaps:
  - `total`: Total number of gaps
  - `by_type`: Count of gaps by type
  - `by_severity`: Count of gaps by severity

### Cache Functionality

Gap analysis results are automatically cached to improve performance on repeated requests. The cache:

- **Key**: Combination of `project_id` and `codebase_path`
- **TTL**: 1 hour (configurable via `GapAnalysisCacheTTL` environment variable)
- **Behavior**: 
  - First request: Performs full analysis and caches the result
  - Subsequent requests: Returns cached result if available and not expired
  - Cache invalidation: Automatically invalidated when codebase changes are detected

**Cache Benefits**:
- Faster response times for repeated analyses
- Reduced computational load on the Hub
- Consistent results for the same codebase snapshot

**Cache Limitations**:
- Cache is in-memory only (not persisted across Hub restarts)
- Cache does not detect code changes automatically (manual invalidation may be needed)
- Cache is per-Hub instance (not shared across multiple Hub instances)

### Gap Report Persistence

Gap analysis reports are automatically stored in the database for historical tracking and audit purposes.

**Database Storage**:
- **Table**: `gap_reports`
- **Fields**:
  - `id`: UUID (primary key)
  - `project_id`: UUID (foreign key to projects)
  - `gaps`: JSONB (array of gap objects)
  - `summary`: JSONB (summary statistics)
  - `created_at`: Timestamp

**Response Format**:
The API response includes a `report_id` field that can be used to retrieve the stored report:

```json
{
  "success": true,
  "report": { ... },
  "report_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Use Cases**:
- Track gap analysis history over time
- Compare gap reports between different codebase versions
- Audit compliance with business rules
- Generate trend reports on gap resolution

### Example Gap Report

```json
{
  "success": true,
  "report_id": "550e8400-e29b-41d4-a716-446655440000",
  "report": {
    "project_id": "uuid",
    "gaps": [
      {
        "type": "missing_impl",
        "knowledge_item_id": "ki-123",
        "rule_title": "Process Payment",
        "description": "Business rule 'Process Payment' is documented but not implemented in code",
        "recommendation": "Implement business rule 'Process Payment' in code",
        "severity": "high"
      },
      {
        "type": "missing_doc",
        "rule_title": "validateOrder",
        "file_path": "src/services/order.ts",
        "line_number": 45,
        "description": "Function 'validateOrder' at src/services/order.ts:45 is not documented as a business rule",
        "recommendation": "Document business rule for 'validateOrder' or remove if not needed",
        "severity": "medium"
      }
    ],
    "summary": {
      "total": 2,
      "by_type": {
        "missing_impl": 1,
        "missing_doc": 1,
        "partial_match": 0,
        "tests_missing": 0
      },
      "by_severity": {
        "critical": 0,
        "high": 1,
        "medium": 1,
        "low": 0
      }
    },
    "created_at": "2024-01-15T10:30:00Z"
  },
  "report_id": "gr-uuid"
}
```

## Change Detection

### How Change Detection Works

When a document is uploaded or re-ingested:

1. Knowledge items are extracted from the document
2. Existing knowledge items for the document are loaded from the database
3. Changes are detected by comparing:
   - Content hash (for exact matches)
   - Title (for fallback matching)
   - Content differences (for modifications)
4. Change requests are automatically generated for:
   - **New items**: Items present in new document but not in existing knowledge
   - **Modified items**: Items with same title but different content
   - **Removed items**: Items present in existing knowledge but not in new document

### When Changes are Detected

Change detection occurs automatically when:
- A document is uploaded via the `/api/v1/documents/ingest` endpoint
- Knowledge extraction completes successfully
- The processor automatically triggers change detection after saving knowledge items

**Note**: Change detection can be disabled by setting the `ENABLE_AUTO_CHANGE_DETECTION=false` environment variable in the processor configuration.

### Change Request Workflow

1. **Detection**: Change request is created automatically
2. **Pending Approval**: Change request status is `pending_approval`
3. **Review**: Review change request details and impact analysis
4. **Approval/Rejection**: Approve or reject the change request
5. **Implementation** (if approved): Track implementation status

## Change Requests

### Creating Change Requests

Change requests are created automatically during document ingestion. They can also be created manually via the API:

```http
POST /api/v1/change-requests
Content-Type: application/json
Authorization: Bearer <token>

{
  "project_id": "uuid",
  "knowledge_item_id": "ki-123",
  "type": "modification",
  "current_state": {
    "title": "Process Payment",
    "content": "Payment must be validated"
  },
  "proposed_state": {
    "title": "Process Payment",
    "content": "Payment must be validated and authorized"
  }
}
```

### Approving/Rejecting Changes

#### Approve Change Request

```http
PUT /api/v1/change-requests/{id}/approve
Content-Type: application/json
Authorization: Bearer <token>

{
  "approved_by": "user@example.com"
}
```

#### Reject Change Request

```http
PUT /api/v1/change-requests/{id}/reject
Content-Type: application/json
Authorization: Bearer <token>

{
  "rejected_by": "user@example.com",
  "reason": "Not aligned with business goals"
}
```

### Impact Analysis

Before approving a change request, analyze its impact:

```http
GET /api/v1/change-requests/{id}/impact?codebasePath=/path/to/codebase
Authorization: Bearer <token>
```

Response includes:
- **Affected Code**: Files and functions that need to be modified
- **Affected Tests**: Test files that need to be updated
- **Estimated Effort**: Estimated time to implement the change

### Implementation Tracking

Track the implementation status of approved changes:

#### Start Implementation

```http
PUT /api/v1/change-requests/{id}/implementation/start
Content-Type: application/json
Authorization: Bearer <token>

{
  "notes": "Starting implementation"
}
```

#### Update Implementation Status

```http
PUT /api/v1/change-requests/{id}/implementation-status
Content-Type: application/json
Authorization: Bearer <token>

{
  "status": "in_progress",
  "notes": "50% complete"
}
```

#### Complete Implementation

```http
PUT /api/v1/change-requests/{id}/implementation/complete
Content-Type: application/json
Authorization: Bearer <token>

{
  "notes": "Implementation complete"
}
```

#### Get Implementation Status

```http
GET /api/v1/change-requests/{id}/implementation-status
Authorization: Bearer <token>
```

## API Endpoints

### Gap Analysis

- `POST /api/v1/knowledge/gap-analysis` - Run gap analysis

### Change Requests

- `POST /api/v1/change-requests` - Create change request
- `GET /api/v1/change-requests` - List change requests (with filters)
- `GET /api/v1/change-requests/{id}` - Get change request details
- `PUT /api/v1/change-requests/{id}/approve` - Approve change request
- `PUT /api/v1/change-requests/{id}/reject` - Reject change request
- `GET /api/v1/change-requests/{id}/impact` - Get impact analysis
- `PUT /api/v1/change-requests/{id}/implementation/start` - Start implementation
- `PUT /api/v1/change-requests/{id}/implementation/complete` - Complete implementation
- `PUT /api/v1/change-requests/{id}/implementation-status` - Update implementation status
- `GET /api/v1/change-requests/{id}/implementation-status` - Get implementation status

## Database Schema

### gap_reports Table

Stores gap analysis reports:

```sql
CREATE TABLE gap_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    gaps JSONB NOT NULL,
    summary JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### change_requests Table

Stores change requests:

```sql
CREATE TABLE change_requests (
    id VARCHAR(50) PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    knowledge_item_id UUID REFERENCES knowledge_items(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL,
    current_state JSONB,
    proposed_state JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending_approval',
    implementation_status VARCHAR(20),
    implementation_notes TEXT,
    impact_analysis JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    approved_by VARCHAR(255),
    approved_at TIMESTAMP,
    rejected_by VARCHAR(255),
    rejected_at TIMESTAMP,
    rejection_reason TEXT
);
```

## Examples

### Complete Workflow Example

1. **Upload Document**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/documents/upload \
     -H "Authorization: Bearer token" \
     -F "file=@business_rules.md" \
     -F "projectId=uuid"
   ```

2. **Check for Change Requests**:
   ```bash
   sentinel knowledge changes --status pending
   ```

3. **Analyze Impact**:
   ```bash
   sentinel knowledge impact CR-001 --codebase-path ./src
   ```

4. **Approve Change Request**:
   ```bash
   sentinel knowledge approve CR-001 --approved-by user@example.com
   ```

5. **Start Implementation**:
   ```bash
   sentinel knowledge track CR-001 start --notes "Starting implementation"
   ```

6. **Run Gap Analysis**:
   ```bash
   sentinel knowledge gap-analysis --project-id uuid --codebase-path ./src
   ```

7. **Complete Implementation**:
   ```bash
   sentinel knowledge track CR-001 complete --notes "Implementation complete"
   ```

### API Usage Examples

#### List Change Requests with Filters

```bash
curl -X GET "http://localhost:8080/api/v1/change-requests?status=pending&limit=10&offset=0" \
  -H "Authorization: Bearer token"
```

#### Get Change Request Details

```bash
curl -X GET "http://localhost:8080/api/v1/change-requests/CR-001" \
  -H "Authorization: Bearer token"
```

#### Run Gap Analysis

```bash
curl -X POST http://localhost:8080/api/v1/knowledge/gap-analysis \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token" \
  -d '{
    "projectId": "uuid",
    "codebasePath": "./src",
    "options": {
      "includeTests": true,
      "reverseCheck": true
    }
  }'
```

## Best Practices

1. **Regular Gap Analysis**: Run gap analysis regularly to catch discrepancies early
2. **Review Change Requests**: Always review change requests before approval
3. **Impact Analysis**: Always perform impact analysis before approving changes
4. **Track Implementation**: Keep implementation status up to date
5. **Document Changes**: Document why changes were made in rejection reasons or implementation notes

## Troubleshooting

### Gap Analysis Returns No Results

- Ensure project ID is correct
- Verify codebase path is accessible
- Check that business rules exist in knowledge base

### Change Requests Not Created

- Verify document was successfully ingested
- Check that knowledge items were extracted
- Ensure document ID matches existing document

### Impact Analysis Empty

- Verify codebase path is correct
- Check that code contains functions matching business rule keywords
- Ensure AST analysis is working correctly

### Cache Not Working

- Verify cache TTL is set correctly (default: 1 hour)
- Check Hub logs for cache write/read messages
- Ensure same `project_id` and `codebase_path` are used for cache hits
- Cache is in-memory only - restarting Hub clears cache

### Gap Report Not Persisted

- Check Hub logs for storage errors
- Verify database connection is working
- Ensure `gap_reports` table exists (run migrations)
- Check that `report_id` is returned in API response (may be empty if storage fails)

