# AST Analysis Architecture

## Overview

The AST (Abstract Syntax Tree) analysis system provides comprehensive code analysis capabilities using Tree-sitter parsers to build ASTs and perform various analyses.

## Architecture Components

### 1. Parser Layer (`hub/api/ast/parsers.go`)

- **Purpose**: Language parser management
- **Responsibilities**:
  - Initialize and cache Tree-sitter parsers
  - Support for Go, JavaScript, TypeScript, Python
  - Parser normalization and caching

### 2. Core Analysis (`hub/api/ast/analysis.go`)

- **Purpose**: Single-file AST analysis
- **Responsibilities**:
  - Parse code into AST
  - Perform requested analyses
  - Cache results for performance
  - Return findings and statistics

### 3. Cross-File Analysis (`hub/api/ast/cross_file.go`)

- **Purpose**: Multi-file and cross-file analysis
- **Components**:
  - **Symbol Table** (`symbol_table.go`): Tracks symbols across files
  - **Dependency Graph** (`dependency_graph.go`): Tracks file dependencies
- **Capabilities**:
  - Unused export detection
  - Undefined reference detection
  - Circular dependency detection
  - Cross-file duplicate detection

### 4. Security Analysis (`hub/api/ast/security_analysis.go`)

- **Purpose**: Security vulnerability detection
- **Detectors**:
  - SQL Injection (`detection_sql_injection.go`)
  - XSS (`detection_xss.go`)
  - Command Injection (`detection_command_injection.go`)
  - Insecure Crypto (`detection_crypto.go`)
  - Secrets Detection (`detection_secrets.go`)

### 5. Service Layer (`hub/api/services/ast_service.go`)

- **Purpose**: Business logic for AST operations
- **Responsibilities**:
  - Request validation
  - Coordinate AST package calls
  - Convert AST types to API models
  - Error handling

### 6. Handler Layer (`hub/api/handlers/ast_handler.go`)

- **Purpose**: HTTP request handling
- **Responsibilities**:
  - Request parsing and validation
  - Response formatting
  - HTTP status codes
  - Error responses

## Data Flow

```
HTTP Request
    ↓
AST Handler (validation)
    ↓
AST Service (business logic)
    ↓
AST Package (analysis)
    ├── Parser (parse code)
    ├── Analysis Engine (perform analyses)
    ├── Security Detectors (if security analysis)
    └── Cross-File Analyzer (if multi-file)
    ↓
AST Service (convert to models)
    ↓
AST Handler (format response)
    ↓
HTTP Response
```

## Analysis Types

### Single-File Analyses

1. **Duplicates**: Detect duplicate function definitions
2. **Unused**: Detect unused variables
3. **Unreachable**: Detect unreachable code
4. **Orphaned**: Detect orphaned code
5. **Empty Catch**: Detect empty catch blocks
6. **Missing Await**: Detect missing await in async functions
7. **Brace Mismatch**: Detect syntax errors

### Cross-File Analyses

1. **Unused Exports**: Exported symbols never used externally
2. **Undefined Refs**: References to undefined symbols
3. **Circular Dependencies**: Circular import/require chains
4. **Cross-File Duplicates**: Duplicate functions across files

### Security Analyses

1. **SQL Injection**: String concatenation in SQL queries
2. **XSS**: Unescaped user input in HTML output
3. **Command Injection**: User input in shell commands
4. **Insecure Crypto**: Weak hash algorithms (MD5, SHA1)
5. **Hardcoded Secrets**: API keys, passwords in code

## Performance Optimizations

1. **Parser Caching**: Parsers are cached and reused
2. **Result Caching**: Analysis results cached for 5 minutes
3. **Parallel Processing**: Multi-file analysis uses goroutines
4. **Early Termination**: Stop analysis on critical errors

## Error Handling

- **Parser Errors**: Return error with language support info
- **Analysis Errors**: Continue with available analyses
- **Validation Errors**: Return 400 with field-specific errors
- **Service Errors**: Return 500 with generic error message

## Extensibility

### Adding New Languages

1. Add parser initialization in `parsers.go`
2. Add language-specific extraction in `symbol_table.go`
3. Add language-specific detection in security detectors

### Adding New Analyses

1. Create detection function in appropriate file
2. Add to analysis switch in `analysis.go`
3. Update supported analyses in service

### Adding New Security Detectors

1. Create detector file (e.g., `detection_new_vuln.go`)
2. Implement detection function
3. Add to `security_analysis.go`

## Compliance

All code follows CODING_STANDARDS.md:
- Handlers: Max 300 lines
- Services: Max 400 lines
- Detection modules: Max 250 lines
- Tests: Max 500 lines
- Proper error wrapping
- Interface-based design
