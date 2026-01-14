# Phase 13: Knowledge Schema Standardization Guide

## Overview

Phase 13 enhances the existing LLM knowledge extraction (from Phase 4) to extract structured knowledge following the standardized schema defined in `KNOWLEDGE_SCHEMA.md`. This ensures consistent, machine-readable knowledge items with constraints, pseudocode, boundaries, test requirements, and traceability.

## Features

### 1. Schema Validation

All extracted knowledge items are validated against a JSON schema to ensure they follow the standardized format.

**Location**: `hub/processor/schema_validator.go`

**Usage**:
```go
err := validateStructuredKnowledgeItem(&item)
if err != nil {
    // Handle validation error
}
```

### 2. Enhanced Extraction Prompts

Prompts have been enhanced to extract structured knowledge with:
- Constraints with pseudocode
- Boundary specification (inclusive/exclusive)
- Test requirements (minimum 2 per rule)
- Traceability information
- Ambiguity flags

**Location**: `hub/processor/prompts.go`

### 3. Boundary Specification

Constraints automatically detect and normalize boundary specifications (inclusive vs exclusive).

**Location**: `hub/processor/boundary_detector.go`

**Example**:
- "Order age < 24 hours" → boundary: "exclusive"
- "Order age <= 24 hours" → boundary: "inclusive"

### 4. Ambiguity Handling

Automatically detects ambiguous constraints and flags them for clarification.

**Location**: `hub/processor/ambiguity_analyzer.go`

**Detects**:
- Vague time references ("soon", "later")
- Unclear boundaries ("around", "approximately")
- Missing units ("24" without "hours"/"days")
- Multiple interpretations ("may" vs "must")

### 5. Test Case Generation

Automatically generates test requirements from business rules:
- Happy path tests
- Error case tests
- Boundary tests
- Exception tests

**Location**: `hub/processor/test_generator.go`

### 6. Database Schema

Added `structured_data` JSONB column to `knowledge_items` table for storing structured knowledge.

**Migration**: Automatically applied on Hub startup

## Usage

### Document Processing

Documents are automatically processed with structured extraction when uploaded:

```bash
# Upload document (automatic structured extraction)
curl -X POST http://localhost:8080/api/v1/documents/ingest \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@requirements.pdf"
```

### Migration

Migrate existing knowledge items to structured format:

```bash
# Run migration
cd hub/processor
go run main.go --migrate
```

### API Endpoints

**Get Knowledge Items**:
```bash
GET /api/v1/documents/{id}/knowledge
```

Response includes both `content` (backward compatible) and `structured_data` (JSONB).

## Structured Knowledge Format

### Business Rule Example

```json
{
  "id": "BR-001",
  "version": "1.0.0",
  "status": "active",
  "title": "Order Cancellation Window",
  "description": "Orders can only be cancelled within 24 hours",
  "specification": {
    "constraints": [
      {
        "id": "C1",
        "type": "time_based",
        "expression": "Order age < 24 hours",
        "pseudocode": "Date.now() - order.createdAt < 24 * 60 * 60 * 1000",
        "boundary": "exclusive",
        "unit": "hours"
      }
    ]
  },
  "test_requirements": [
    {
      "id": "BR-001-T1",
      "name": "test_cancel_within_window",
      "type": "happy_path",
      "scenario": "Cancel order within 24 hours"
    },
    {
      "id": "BR-001-T2",
      "name": "test_cancel_after_window",
      "type": "error_case",
      "scenario": "Reject cancel after 24 hours"
    }
  ],
  "traceability": {
    "source_document": "requirements.pdf",
    "source_page": 12,
    "source_quote": "Orders can be cancelled within 24 hours..."
  }
}
```

## Configuration

### Environment Variables

- `AZURE_AI_ENDPOINT` - Azure AI Foundry endpoint
- `AZURE_AI_KEY` - Azure AI Foundry API key
- `AZURE_AI_DEPLOYMENT` - Deployment name (default: "claude-opus-4-5")
- `OLLAMA_HOST` - Ollama host (default: "http://localhost:11434")

## Testing

Run Phase 13 tests:

```bash
# Unit tests
./tests/unit/schema_validator_test.sh
./tests/unit/prompts_test.sh
./tests/unit/migration_test.sh

# Integration tests
./tests/integration/phase13_e2e_test.sh
```

## Troubleshooting

### Validation Errors

If validation fails, check:
1. Required fields are present
2. Constraint boundaries are specified (inclusive/exclusive)
3. Test requirements meet minimum (2 per rule)

### Ambiguity Flags

Items with ambiguity flags are set to `status = "needs_clarification"`. Review and clarify ambiguous constraints.

### Migration Issues

If migration fails:
1. Check database connection
2. Verify `structured_data` column exists
3. Review migration logs for specific errors

## References

- `KNOWLEDGE_SCHEMA.md` - Complete schema specification
- `hub/processor/schemas/knowledge_schema.json` - JSON schema definition
- `hub/processor/main.go` - Main extraction logic











