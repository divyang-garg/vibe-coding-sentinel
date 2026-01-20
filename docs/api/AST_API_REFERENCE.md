# AST API Reference

## Overview

The AST (Abstract Syntax Tree) API provides comprehensive code analysis capabilities including single-file analysis, multi-file analysis, security vulnerability detection, and cross-file dependency analysis.

## Base URL

```
/api/v1/ast
```

## Authentication

All endpoints require authentication via API key in the `Authorization` header:

```
Authorization: Bearer <api_key>
```

## Endpoints

### 1. Single-File AST Analysis

**POST** `/api/v1/ast/analyze`

Performs AST analysis on a single code file.

#### Request Body

```json
{
  "code": "package main\nfunc test() {}\n",
  "language": "go",
  "analyses": ["duplicates", "unused", "unreachable"],
  "file_path": "main.go"
}
```

#### Parameters

- `code` (string, required): Source code to analyze
- `language` (string, required): Programming language (`go`, `javascript`, `typescript`, `python`)
- `analyses` (array, optional): List of analyses to perform. If empty, performs all analyses.
  - `duplicates`: Detect duplicate functions
  - `unused`: Detect unused variables
  - `unreachable`: Detect unreachable code
  - `orphaned`: Detect orphaned code
  - `empty_catch`: Detect empty catch blocks
  - `missing_await`: Detect missing await in async functions
  - `brace_mismatch`: Detect brace/bracket mismatches
- `file_path` (string, optional): File path for context

#### Response

```json
{
  "findings": [
    {
      "type": "unused_variable",
      "severity": "medium",
      "line": 5,
      "column": 10,
      "end_line": 5,
      "end_column": 15,
      "message": "Variable 'x' is declared but never used",
      "code": "var x int",
      "suggestion": "Remove unused variable or use it",
      "confidence": 0.95,
      "auto_fix_safe": true,
      "fix_type": "delete",
      "reasoning": "Variable has no references"
    }
  ],
  "stats": {
    "parse_time_ms": 10,
    "analysis_time_ms": 50,
    "nodes_visited": 150
  },
  "language": "go",
  "file_path": "main.go"
}
```

### 2. Multi-File AST Analysis

**POST** `/api/v1/ast/analyze/multi`

Performs AST analysis across multiple files.

#### Request Body

```json
{
  "files": [
    {
      "path": "file1.go",
      "content": "package main\nfunc test() {}\n",
      "language": "go"
    },
    {
      "path": "file2.go",
      "content": "package main\nfunc test2() {}\n",
      "language": "go"
    }
  ],
  "analyses": ["duplicates", "unused"],
  "project_root": "/path/to/project"
}
```

#### Response

```json
{
  "findings": [...],
  "stats": {
    "parse_time_ms": 20,
    "analysis_time_ms": 100,
    "nodes_visited": 300
  },
  "files": ["file1.go", "file2.go"]
}
```

### 3. Security Analysis

**POST** `/api/v1/ast/analyze/security`

Performs security-focused AST analysis to detect vulnerabilities.

#### Request Body

```json
{
  "code": "package main\nfunc test(id string) {\n  db.Query(\"SELECT * FROM users WHERE id = \" + id)\n}\n",
  "language": "go",
  "severity": "critical"
}
```

#### Parameters

- `code` (string, required): Source code to analyze
- `language` (string, required): Programming language
- `severity` (string, optional): Filter by severity (`critical`, `high`, `medium`, `low`, `all`). Default: `all`

#### Response

```json
{
  "findings": [...],
  "stats": {
    "parse_time_ms": 15,
    "analysis_time_ms": 80,
    "nodes_visited": 200
  },
  "vulnerabilities": [
    {
      "type": "sql_injection",
      "severity": "critical",
      "line": 3,
      "column": 10,
      "message": "Potential SQL injection: string concatenation in SQL query",
      "code": "db.Query(\"SELECT * FROM users WHERE id = \" + id)",
      "description": "SQL query constructed using string concatenation with user input",
      "remediation": "Use parameterized queries (e.g., db.Query with ? placeholders)",
      "confidence": 0.9
    }
  ],
  "risk_score": 85.5
}
```

#### Vulnerability Types

- `sql_injection`: SQL injection vulnerabilities
- `xss`: Cross-site scripting vulnerabilities
- `command_injection`: Command injection vulnerabilities
- `insecure_crypto`: Insecure cryptographic usage
- `hardcoded_secret`: Hardcoded secrets, API keys, or credentials

### 4. Cross-File Dependency Analysis

**POST** `/api/v1/ast/analyze/cross`

Analyzes dependencies and relationships across multiple files.

#### Request Body

```json
{
  "files": [
    {
      "path": "file1.go",
      "content": "package main\nfunc Exported() {}\n",
      "language": "go"
    },
    {
      "path": "file2.go",
      "content": "package main\nimport \"fmt\"\n",
      "language": "go"
    }
  ],
  "project_root": "/path/to/project"
}
```

#### Response

```json
{
  "findings": [...],
  "unused_exports": [
    {
      "name": "Exported",
      "kind": "function",
      "file_path": "file1.go",
      "line": 2,
      "column": 1
    }
  ],
  "undefined_refs": [
    {
      "name": "UndefinedFunc",
      "file_path": "file2.go",
      "line": 5,
      "column": 10,
      "kind": "call"
    }
  ],
  "circular_deps": [
    ["file1.go", "file2.go", "file1.go"]
  ],
  "cross_file_duplicates": [...],
  "stats": {
    "files_analyzed": 2,
    "symbols_found": 10,
    "dependencies_found": 5,
    "analysis_time_ms": 120
  }
}
```

### 5. Get Supported Analyses

**GET** `/api/v1/ast/supported`

Returns list of supported languages and analyses.

#### Response

```json
{
  "languages": [
    {
      "name": "go",
      "aliases": ["golang"],
      "supported": true
    },
    {
      "name": "javascript",
      "aliases": ["js", "jsx"],
      "supported": true
    },
    {
      "name": "typescript",
      "aliases": ["ts", "tsx"],
      "supported": true
    },
    {
      "name": "python",
      "aliases": ["py"],
      "supported": true
    }
  ],
  "analyses": [
    "duplicates",
    "unused",
    "unreachable",
    "orphaned",
    "empty_catch",
    "missing_await",
    "brace_mismatch",
    "unused_exports",
    "undefined_refs",
    "circular_deps",
    "cross_file_duplicates"
  ]
}
```

## Error Responses

All endpoints return standard error responses:

```json
{
  "error": "validation failed for field 'code': Code is required"
}
```

### Status Codes

- `200 OK`: Success
- `400 Bad Request`: Invalid request (missing required fields, invalid format)
- `401 Unauthorized`: Missing or invalid authentication
- `500 Internal Server Error`: Server error during analysis

## Rate Limiting

- Default: 100 requests per 10 seconds per API key
- Security analysis endpoints: 50 requests per 10 seconds

## Examples

### cURL Examples

```bash
# Single-file analysis
curl -X POST https://api.example.com/api/v1/ast/analyze \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "package main\nfunc test() {}\n",
    "language": "go"
  }'

# Security analysis
curl -X POST https://api.example.com/api/v1/ast/analyze/security \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "db.Query(\"SELECT * FROM users WHERE id = \" + id)",
    "language": "go",
    "severity": "critical"
  }'
```

## Best Practices

1. **Use appropriate analyses**: Only request analyses you need to reduce processing time
2. **Batch requests**: Use multi-file endpoints for analyzing multiple files
3. **Handle errors**: Always check error responses and handle them appropriately
4. **Cache results**: AST analysis results can be cached client-side for unchanged code
5. **Rate limiting**: Respect rate limits and implement exponential backoff

## Performance

- Single-file analysis: < 100ms for typical files
- Multi-file analysis: < 500ms for 10 files
- Security analysis: < 200ms for typical files
- Cross-file analysis: < 1s for 20 files
