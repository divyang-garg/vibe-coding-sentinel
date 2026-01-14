# Phase 14A: Comprehensive Feature Analysis Guide

## Overview

Phase 14A provides comprehensive feature analysis across all 7 layers of a software system (Business, UI, API, Database, Logic, Integration, Tests). It automatically discovers features, analyzes each layer for issues, verifies end-to-end flows, and generates actionable checklists.

## Features

### 1. Feature Discovery

Automatically discovers features across all layers by detecting:
- **UI Layer**: React, Vue, Angular components
- **API Layer**: Express, FastAPI, Django, Gin endpoints
- **Database Layer**: SQL migrations, Prisma, TypeORM schemas
- **Business Logic Layer**: Functions using AST analysis
- **Integration Layer**: External API calls
- **Test Layer**: Test files and scenarios

**Location**: `hub/api/feature_discovery.go`

**Usage**:
```go
feature, err := OrchestrateFeatureDiscovery(ctx, featureName, codebasePath, files)
```

### 2. Layer-Specific Analyzers

#### Business Context Analyzer
Validates code against business rules, user journeys, and entities.

**Location**: `hub/api/business_context_analyzer.go`

**Checks**:
- Business rule violations
- User journey adherence
- Entity validation failures

#### UI Layer Analyzer
Analyzes UI components for validation, error handling, and accessibility.

**Location**: `hub/api/ui_analyzer.go`

**Supports**:
- React/Next.js components
- Vue components
- Angular components
- Accessibility checks (ARIA, semantic HTML, keyboard navigation)

**Checks**:
- Missing form validation
- Missing error handling
- Missing loading states
- Accessibility issues (missing alt text, labels, focus management)

#### API Layer Analyzer
Analyzes API endpoints for security, validation, and contract compliance.

**Location**: `hub/api/api_analyzer.go`

**Checks**:
- Security vulnerabilities (reuses security analyzer)
- Missing input validation
- Missing error handling
- API contract violations

#### Database Layer Analyzer
Analyzes database schema for constraints, indexes, and data integrity.

**Location**: `hub/api/database_analyzer.go`

**Supports**:
- Prisma schemas
- TypeORM schemas
- Raw SQL migrations

**Checks**:
- Missing primary keys
- Missing foreign key constraints
- Missing indexes
- Nullable columns that should be required

#### Business Logic Analyzer
Analyzes business logic functions with AST and LLM-based semantic analysis.

**Location**: `hub/api/logic_analyzer.go`

**Checks**:
- Missing error handling
- Semantic errors (null references, type mismatches, logic errors)
- Edge case handling
- LLM-powered semantic analysis (when configured)

#### Integration Layer Analyzer
Analyzes external API integrations for error handling, retry logic, and contracts.

**Location**: `hub/api/integration_analyzer.go`

**Checks**:
- Missing error handling
- Missing retry logic
- Missing timeout configuration
- Contract mismatches

#### Test Layer Analyzer
Analyzes test coverage and test scenarios.

**Location**: `hub/api/test_analyzer.go`

**Checks**:
- Test coverage (reuses test coverage tracker)
- Missing test scenarios
- Incomplete test assertions

### 3. End-to-End Flow Verification

Verifies complete user journeys across all layers, identifying breakpoints where flows are broken.

**Location**: `hub/api/flow_verifier.go`

**Features**:
- Flow detection (UI → API → Logic → Database)
- Breakpoint identification
- Integration point verification
- Flow status tracking (complete/broken/partial)

**Breakpoint Types**:
- Missing error handling
- Missing validation
- Missing rollback (for database operations)
- Missing timeout (for integrations)
- Contract mismatches

### 4. LLM Integration

#### Model Selection
Automatically selects appropriate LLM model based on task criticality:
- **Critical tasks** (business rule validation, security analysis, semantic analysis): High-accuracy models (GPT-4, Claude Opus)
- **Non-critical tasks**: Cheaper/faster models (GPT-3.5-turbo, Claude Haiku)

**Location**: `hub/api/llm_integration.go`

#### Cost Optimization
Implements progressive depth analysis with caching:
- **Surface depth**: Pattern-based analysis only (no LLM)
- **Medium depth**: Cheaper models with caching
- **Deep depth**: High-accuracy models with caching

**Location**: `hub/api/llm_cache.go`

**Features**:
- Response caching (24-hour TTL)
- Progressive depth (only use LLM when needed)
- Token usage tracking
- Cost estimation

### 5. Result Aggregation

Combines findings from all analyzers into actionable checklists and summaries.

**Location**: `hub/api/result_aggregator.go`

**Outputs**:
- Prioritized checklist items
- Summary statistics
- Formatted reports
- Database storage for trending

## API Endpoints

### POST /api/v1/analyze/comprehensive

Performs comprehensive feature analysis.

**Request**:
```json
{
  "feature": "user-authentication",
  "mode": "auto",
  "codebasePath": "/path/to/codebase",
  "depth": "medium",
  "includeBusinessContext": true
}
```

**Parameters**:
- `feature` (required): Feature name to analyze
- `mode` (required): "auto" or "manual"
- `codebasePath` (required for auto mode): Path to codebase
- `files` (optional, for manual mode): Map of layer to file paths
- `depth` (optional): "surface", "medium", or "deep" (default: "medium")
- `includeBusinessContext` (optional): Include business rule validation (default: false)

**Response**:
```json
{
  "validation_id": "uuid",
  "feature": "user-authentication",
  "summary": {
    "total_findings": 15,
    "critical": 2,
    "high": 5,
    "medium": 6,
    "low": 2,
    "flows_verified": 3,
    "flows_broken": 1,
    "analysis_time": "2.5s"
  },
  "checklist": [
    {
      "priority": "critical",
      "layer": "api",
      "issue": "Missing input validation",
      "location": "api/auth.go:45"
    }
  ],
  "layer_analysis": {
    "business": [...],
    "ui": [...],
    "api": [...],
    "database": [...],
    "logic": [...],
    "integration": [...],
    "test": [...]
  },
  "flows": [
    {
      "name": "Login Flow",
      "status": "broken",
      "steps": [...],
      "breakpoints": [...]
    }
  ]
}
```

### GET /api/v1/validations/{id}

Retrieves a specific validation report.

**Response**:
```json
{
  "id": "uuid",
  "project_id": "uuid",
  "feature": "user-authentication",
  "report": {...},
  "created_at": "2024-01-01T00:00:00Z"
}
```

### GET /api/v1/validations?project={id}

Lists validation reports for a project with pagination.

**Query Parameters**:
- `project` (required): Project ID
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response**:
```json
{
  "validations": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

## Usage Examples

### Basic Analysis

```bash
curl -X POST http://localhost:8080/api/v1/analyze/comprehensive \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "feature": "user-authentication",
    "mode": "auto",
    "codebasePath": "/path/to/codebase",
    "depth": "medium"
  }'
```

### Deep Analysis with Business Context

```bash
curl -X POST http://localhost:8080/api/v1/analyze/comprehensive \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "feature": "payment-processing",
    "mode": "auto",
    "codebasePath": "/path/to/codebase",
    "depth": "deep",
    "includeBusinessContext": true
  }'
```

### Manual Mode (Specify Files)

```bash
curl -X POST http://localhost:8080/api/v1/analyze/comprehensive \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "feature": "user-profile",
    "mode": "manual",
    "files": {
      "ui": ["src/components/Profile.jsx"],
      "api": ["api/profile.go"],
      "database": ["migrations/001_profile.sql"]
    },
    "depth": "medium"
  }'
```

### Retrieve Validation Report

```bash
curl -X GET http://localhost:8080/api/v1/validations/{validation_id} \
  -H "Authorization: Bearer $API_KEY"
```

## LLM Configuration

### Setting Up LLM

Configure LLM for semantic analysis:

```sql
INSERT INTO llm_configurations (project_id, provider, api_key_encrypted, model, key_type, cost_optimization)
VALUES (
  'project-uuid',
  'openai',
  '<encrypted-api-key>',
  'gpt-4',
  'api_key',
  '{"use_cache": true, "cache_ttl_hours": 24, "progressive_depth": true}'::jsonb
);
```

**Supported Providers**:
- `openai`: OpenAI API
- `anthropic`: Anthropic Claude API
- `azure`: Azure OpenAI Service

**Cost Optimization**:
- `use_cache`: Enable response caching (default: true)
- `cache_ttl_hours`: Cache TTL in hours (default: 24)
- `progressive_depth`: Use progressive depth analysis (default: true)
- `max_cost_per_request`: Maximum cost per request (optional)

### Viewing LLM Usage

```sql
SELECT 
  provider,
  model,
  SUM(tokens_used) as total_tokens,
  SUM(estimated_cost) as total_cost,
  COUNT(*) as request_count
FROM llm_usage
WHERE project_id = 'project-uuid'
  AND created_at >= NOW() - INTERVAL '30 days'
GROUP BY provider, model;
```

## Database Schema

### comprehensive_validations

Stores comprehensive analysis results.

```sql
CREATE TABLE comprehensive_validations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id),
  feature VARCHAR(255) NOT NULL,
  mode VARCHAR(50) NOT NULL,
  depth VARCHAR(50) NOT NULL,
  report JSONB NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_comprehensive_validations_project ON comprehensive_validations(project_id);
CREATE INDEX idx_comprehensive_validations_feature ON comprehensive_validations(feature);
CREATE INDEX idx_comprehensive_validations_created ON comprehensive_validations(created_at);
```

### llm_usage

Tracks LLM token usage and costs.

```sql
CREATE TABLE llm_usage (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id),
  validation_id UUID REFERENCES comprehensive_validations(id),
  provider VARCHAR(50) NOT NULL,
  model VARCHAR(100) NOT NULL,
  tokens_used INTEGER NOT NULL,
  estimated_cost DECIMAL(10, 4) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_llm_usage_project ON llm_usage(project_id);
CREATE INDEX idx_llm_usage_validation ON llm_usage(validation_id);
CREATE INDEX idx_llm_usage_created ON llm_usage(created_at);
```

## Performance Considerations

### Analysis Depth

- **Surface**: Fast, pattern-based only (< 10 seconds)
- **Medium**: Includes LLM with caching (< 30 seconds)
- **Deep**: Full LLM analysis (< 60 seconds)

### Caching

LLM responses are cached for 24 hours by default. Cache keys are based on:
- File content hash
- Analysis type
- Depth level

### Optimization Tips

1. Use `surface` depth for quick checks
2. Use `medium` depth for regular analysis
3. Use `deep` depth only when needed
4. Enable LLM caching to reduce costs
5. Use progressive depth to avoid unnecessary LLM calls

## Troubleshooting

### LLM Analysis Not Working

1. Check LLM configuration exists for project:
   ```sql
   SELECT * FROM llm_configurations WHERE project_id = 'project-uuid';
   ```

2. Verify API key is valid:
   ```bash
   # Test connection (if implemented)
   curl -X POST http://localhost:8080/api/v1/llm/test \
     -H "Authorization: Bearer $API_KEY"
   ```

3. Check LLM usage logs:
   ```sql
   SELECT * FROM llm_usage 
   WHERE project_id = 'project-uuid' 
   ORDER BY created_at DESC 
   LIMIT 10;
   ```

### Slow Analysis

1. Check analysis depth (use `surface` for faster results)
2. Verify caching is enabled
3. Check database query performance
4. Monitor LLM API response times

### Missing Findings

1. Verify feature discovery found components
2. Check analyzer logs for errors
3. Ensure codebase path is correct
4. Verify file permissions

## Best Practices

1. **Start with surface depth** for quick overview
2. **Use medium depth** for regular analysis
3. **Use deep depth** only for critical features
4. **Enable business context** for compliance checks
5. **Review checklists** and prioritize critical issues
6. **Track LLM costs** and optimize usage
7. **Store reports** for trending and comparison

## Related Documentation

- [Phase 12 Guide](./PHASE_12_GUIDE.md) - Requirements Lifecycle Management
- [Phase 13 Guide](./PHASE_13_GUIDE.md) - Knowledge Schema Standardization
- [Architecture](./ARCHITECTURE.md) - System Architecture
- [Features](./FEATURES.md) - Complete Feature List










