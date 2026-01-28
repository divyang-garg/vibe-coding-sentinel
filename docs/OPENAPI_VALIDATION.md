# OpenAPI Contract Validation Guide

## Overview

The OpenAPI contract validation system provides production-ready validation of API endpoints against OpenAPI/Swagger specifications. It uses the `libopenapi` library for enterprise-grade parsing and validation.

## Features

- **Full OpenAPI Support**: OpenAPI 2.0, 3.0, 3.1, 3.2
- **$ref Resolution**: Automatic resolution of schema references
- **Deep Schema Validation**: Parameters, request body, responses, security
- **Code Analysis**: AST-based schema extraction from code
- **Framework Support**: Go (Gin, Echo), Express.js, FastAPI
- **Performance**: Caching and optimization for large contracts
- **Error Reporting**: Detailed findings with contract paths and suggested fixes

## Usage

### Basic Validation

```go
import (
    "context"
    "sentinel-hub-api/services"
)

ctx := context.Background()

// Validate endpoints against contract
endpoints := []services.EndpointInfo{
    {
        Method: "GET",
        Path:   "/users/:id",
        File:   "handlers/users.go",
        Parameters: []services.ParameterInfo{
            {Name: "id", Type: "path", DataType: "int", Required: true},
        },
        Responses: []services.ResponseInfo{
            {StatusCode: 200},
        },
    },
}

findings, err := services.ValidateAPIContracts(ctx, ".", endpoints)
if err != nil {
    log.Fatal(err)
}

for _, finding := range findings {
    fmt.Printf("Type: %s\n", finding.Type)
    fmt.Printf("Issue: %s\n", finding.Issue)
    fmt.Printf("Severity: %s\n", finding.Severity)
    fmt.Printf("Contract Path: %s\n", finding.ContractPath)
    if finding.SuggestedFix != "" {
        fmt.Printf("Suggested Fix: %s\n", finding.SuggestedFix)
    }
}
```

### Parsing Contracts

```go
ctx := context.Background()

// Parse OpenAPI contract (with caching)
contract, err := services.GetCachedContract(ctx, "openapi.yaml")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("OpenAPI Version: %s\n", contract.Version)
fmt.Printf("Endpoints: %d\n", len(contract.Endpoints))

for _, endpoint := range contract.Endpoints {
    fmt.Printf("  %s %s\n", endpoint.Method, endpoint.Path)
}
```

### Schema Extraction

```go
ctx := context.Background()

endpoint := services.EndpointInfo{
    Method:  "POST",
    Path:    "/users",
    File:    "handlers/users.go",
    Handler: "CreateUserHandler",
}

// Extract request schema from code
requestSchema, err := services.ExtractRequestSchema(ctx, endpoint)
if err != nil {
    log.Fatal(err)
}

if requestSchema != nil {
    fmt.Printf("Request Schema Type: %s\n", requestSchema.Type)
    for name, prop := range requestSchema.Properties {
        fmt.Printf("  %s: %s\n", name, prop.Type)
    }
}
```

## Contract File Detection

The validator automatically looks for contract files in the following order:
1. `openapi.yaml`
2. `openapi.json`
3. `swagger.yaml`
4. `swagger.json`

Place your contract file in the codebase root directory.

## Validation Types

### Parameter Validation

Validates:
- Required parameters are present
- Parameter types match contract
- Parameter locations (path, query, header, cookie)
- Enum values
- Pattern constraints (regex)
- Min/Max constraints

### Request Body Validation

Validates:
- Required request body is present
- Content-Type matches contract
- Schema structure matches
- Required fields are present

### Response Validation

Validates:
- Response status codes match contract
- Response schemas match contract
- Content-Type matches
- Response headers match

### Security Validation

Validates:
- Security requirements are implemented
- Security schemes match contract
- Authentication methods are present

## Error Reporting

Each finding includes:
- **Type**: Type of mismatch (contract_mismatch, missing_auth, etc.)
- **Location**: Code file location
- **Issue**: Description of the issue
- **Severity**: critical, high, medium, low
- **ContractPath**: JSON path in contract (e.g., #/paths/~1users/get/parameters/0)
- **SuggestedFix**: Suggested fix for the issue
- **Details**: Additional details map

## Performance

### Caching

Contracts are automatically cached with a default TTL of 5 minutes. Cache is invalidated when:
- TTL expires
- Contract file is modified

```go
cache := services.GetContractCache()
cache.SetTTL(10 * time.Minute) // Custom TTL
cache.ClearCache()              // Clear all cached contracts
```

### Performance Targets

- Parse 1000-endpoint contract: < 1 second
- Validate 100 endpoints: < 500ms
- Memory usage: < 50MB for typical contract

## Framework Support

### Go (Gin, Echo)

Automatically extracts schemas from Go struct definitions:

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

### Express.js

Extracts schemas from Joi or Zod validation:

```javascript
const schema = Joi.object({
    name: Joi.string().required(),
    email: Joi.string().email().required()
});
```

### FastAPI

Extracts schemas from Pydantic models:

```python
class CreateUserRequest(BaseModel):
    name: str
    email: str
```

## Best Practices

1. **Keep contracts up to date**: Update OpenAPI contracts when endpoints change
2. **Use $ref for reusability**: Define reusable schemas in components
3. **Validate in CI/CD**: Add contract validation to your CI pipeline
4. **Monitor performance**: Use caching for large contracts
5. **Review findings**: Address high/critical severity findings first

## Troubleshooting

### Contract Not Found

If no contract file is found, validation is skipped (not an error). Ensure:
- Contract file is in the codebase root
- File name matches: `openapi.yaml`, `openapi.json`, `swagger.yaml`, or `swagger.json`

### $ref Resolution Errors

If $ref resolution fails:
- Check that referenced schemas exist in components
- Verify external references are accessible
- Ensure circular references are handled correctly

### Performance Issues

If validation is slow:
- Enable caching (automatic by default)
- Check contract file size
- Consider splitting large contracts

## Examples

See test files for comprehensive examples:
- `hub/api/services/openapi_parser_test.go`
- `hub/api/services/openapi_integration_test.go`
- `hub/api/services/schema_validator_test.go`
