# API Validation Rules

This document describes the input validation rules for all Sentinel Hub API endpoints.

**Last Updated:** January 20, 2026

---

## Overview

All API endpoints implement comprehensive input validation to ensure:
- Data integrity
- Security (SQL injection, XSS prevention)
- Performance (request size limits)
- User experience (clear error messages)

---

## Validation Framework

The validation framework uses a layered approach:
1. **Middleware Validation** - Request-level validation (size, format)
2. **Field Validation** - Field-specific rules (type, length, pattern)
3. **Business Logic Validation** - Domain-specific rules (service layer)

---

## Common Validation Rules

### String Fields

| Rule | Description | Example |
|------|-------------|---------|
| **Required** | Field must be present and non-empty | `name` in CreateProject |
| **MinLength** | Minimum character count | `title` must be at least 1 character |
| **MaxLength** | Maximum character count | `title` max 500 characters |
| **Pattern** | Regex pattern matching | Project names: alphanumeric + spaces, hyphens, underscores |
| **Enum** | Must match one of allowed values | `status`: "pending", "in_progress", "completed", "archived" |

### Numeric Fields

| Rule | Description | Example |
|------|-------------|---------|
| **Required** | Field must be present | `limit` in pagination |
| **Min** | Minimum value | `limit` must be >= 1 |
| **Max** | Maximum value | `limit` must be <= 1000 |
| **Type** | Integer or float | `age`: integer, `price`: float |

### Special Formats

| Type | Pattern | Example |
|------|---------|---------|
| **Email** | RFC 5322 compliant | `user@example.com` |
| **UUID** | Standard UUID format | `550e8400-e29b-41d4-a716-446655440000` |
| **URL** | HTTP/HTTPS URL | `https://example.com` |

---

## Endpoint-Specific Validation

### Task Endpoints

#### POST `/api/v1/tasks` - Create Task

**Request Body:**
```json
{
  "title": "string (required, 1-500 chars)",
  "description": "string (optional, max 5000 chars)",
  "status": "string (required, enum: pending|in_progress|completed|archived)",
  "priority": "string (optional, enum: low|medium|high|critical)",
  "source": "string (optional, max 100 chars)",
  "file_path": "string (optional, max 1000 chars)"
}
```

**Validation Rules:**
- `title`: Required, 1-500 characters
- `description`: Optional, max 5000 characters
- `status`: Required, must be one of: "pending", "in_progress", "completed", "archived"
- `priority`: Optional, must be one of: "low", "medium", "high", "critical"
- `source`: Optional, max 100 characters
- `file_path`: Optional, max 1000 characters

**Example Valid Request:**
```json
{
  "title": "Fix authentication bug",
  "description": "User cannot log in with API key",
  "status": "pending",
  "priority": "high"
}
```

**Example Error Response:**
```json
{
  "error": "validation failed for field 'title': title is required",
  "field": "title"
}
```

#### PUT `/api/v1/tasks/{id}` - Update Task

**Request Body:**
```json
{
  "title": "string (optional, 1-500 chars)",
  "description": "string (optional, max 5000 chars)",
  "status": "string (optional, enum: pending|in_progress|completed|archived)",
  "priority": "string (optional, enum: low|medium|high|critical)"
}
```

**Validation Rules:**
- All fields are optional
- If provided, same rules as Create Task apply

#### GET `/api/v1/tasks` - List Tasks

**Query Parameters:**
- `status_filter`: Optional, enum: "pending"|"in_progress"|"completed"|"archived"
- `priority_filter`: Optional, enum: "low"|"medium"|"high"|"critical"
- `source_filter`: Optional, max 100 characters
- `limit`: Optional, integer, 1-1000 (default: 100)
- `offset`: Optional, integer, >= 0 (default: 0)

---

### Project Endpoints

#### POST `/api/v1/projects` - Create Project

**Request Body:**
```json
{
  "name": "string (required, 1-255 chars, alphanumeric + spaces/hyphens/underscores)"
}
```

**Validation Rules:**
- `name`: Required, 1-255 characters
- Pattern: `^[a-zA-Z0-9\s\-_]+$` (alphanumeric, spaces, hyphens, underscores only)

**Example Valid Request:**
```json
{
  "name": "My Project"
}
```

**Example Invalid Request:**
```json
{
  "name": "Project@123"  // Invalid: contains @ symbol
}
```

---

### Organization Endpoints

#### POST `/api/v1/organizations` - Create Organization

**Request Body:**
```json
{
  "name": "string (required, 1-255 chars)",
  "description": "string (optional, max 2000 chars)"
}
```

**Validation Rules:**
- `name`: Required, 1-255 characters
- `description`: Optional, max 2000 characters

---

## Security Validation

### SQL Injection Prevention

All string inputs are checked for SQL injection patterns:
- `UNION`, `SELECT`, `INSERT`, `UPDATE`, `DELETE`
- `DROP`, `CREATE`, `ALTER`, `EXEC`, `EXECUTE`
- `SCRIPT` keywords

**Example:**
```json
{
  "query": "'; DROP TABLE users; --"  // ❌ Rejected
}
```

**Error Response:**
```json
{
  "error": "validation failed for field 'query': potentially unsafe input detected",
  "field": "query"
}
```

### XSS Prevention

All user inputs are sanitized:
- Null bytes removed
- Control characters filtered (except newline and tab)
- Whitespace trimmed

### Path Traversal Prevention

File paths are validated to prevent directory traversal:
- `..` sequences are rejected
- Absolute paths validated
- Path normalization applied

---

## Request Size Limits

| Endpoint Type | Max Size | Rationale |
|---------------|----------|-----------|
| Task endpoints | 10 MB | Allow for large descriptions |
| Project/Organization | 5 MB | Standard CRUD operations |
| Document upload | Configurable | Separate endpoint |

**Error Response (413):**
```json
{
  "error": "Request body too large: 15728640 bytes (maximum: 10485760 bytes)"
}
```

---

## Error Responses

### Validation Error Format

```json
{
  "error": "validation failed for field '<field>': <message>",
  "field": "<field_name>"
}
```

### HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| **400** | Bad Request - Validation failed | Missing required field |
| **413** | Request Entity Too Large | Body exceeds size limit |
| **422** | Unprocessable Entity | Invalid data format |

---

## Best Practices

### For API Consumers

1. **Always validate on client side** before sending requests
2. **Handle validation errors gracefully** - show user-friendly messages
3. **Respect field length limits** - truncate if necessary
4. **Use appropriate data types** - strings for text, numbers for numeric values
5. **Follow enum values** - use exact values from documentation

### For API Developers

1. **Add validation to all endpoints** - never trust client input
2. **Use validation middleware** - consistent validation across endpoints
3. **Provide clear error messages** - help users fix issues
4. **Document validation rules** - keep this document updated
5. **Test edge cases** - empty strings, null values, boundary conditions

---

## Testing Validation

### Manual Testing

Use curl or Postman to test validation:

```bash
# Valid request
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"title": "Test", "status": "pending"}'

# Invalid request (missing required field)
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"status": "pending"}'
```

### Automated Testing

Validation is tested with:
- Unit tests for each validator
- Integration tests for endpoints
- Security tests for injection attempts

---

## Changelog

### 2026-01-20
- Initial validation framework implementation
- Task, Project, Organization endpoint validation
- SQL injection and XSS prevention
- Request size limits

---

## Support

For questions or issues with validation:
1. Check this documentation
2. Review error messages carefully
3. Contact API support team

---

**Note:** This document is maintained alongside the codebase. Validation rules may change with API updates.
