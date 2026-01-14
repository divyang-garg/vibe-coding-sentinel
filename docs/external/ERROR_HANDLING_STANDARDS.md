# Error Handling Standards

## Overview

This document defines the standard error handling patterns for the Sentinel project, ensuring consistency, clarity, and proper error propagation across all components.

## Standard Error Message Format

### Format Structure

```
<Component>: <Action> failed: <Reason> [<Context>]
```

### Examples

- `Hub API: Failed to fetch hook policy: database connection timeout [org_id: abc-123]`
- `Agent: Failed to send telemetry: network error after 3 retries [endpoint: /api/v1/telemetry]`
- `Validation: Invalid UUID format: xyz-456 [field: org_id]`

## Logging Standards

### When to Use Each Log Level

#### `log.Printf()` / `logInfo()`
- **Use for**: Informational messages, successful operations, normal flow
- **Examples**:
  - "Hook execution recorded successfully"
  - "Cache hit for policy org_id: abc-123"
  - "Telemetry sent to Hub"

#### `logWarn()`
- **Use for**: Recoverable errors, degraded functionality, retry attempts
- **Examples**:
  - "HTTP request failed (attempt 2/3), retrying in 200ms..."
  - "Cache entry expired, refetching from Hub"
  - "Database query timeout, using cached value"

#### `logError()` / `log.Printf("ERROR: ...")`
- **Use for**: Unrecoverable errors, critical failures, data corruption
- **Examples**:
  - "ERROR: Database connection pool exhausted"
  - "ERROR: Failed to save audit report: disk full"
  - "ERROR: Invalid configuration: missing required field 'apiKey'"

### Logging Context

Always include relevant context in log messages:

```go
// Good
logWarn("Error fetching hook policy after retries: %v [org_id: %s, agent_id: %s]", err, orgID, agentID)

// Bad
logWarn("Error: %v", err)
```

## Error Context Propagation

### Pattern: Wrap Errors with Context

```go
// Good - preserves original error with context
if err != nil {
    return fmt.Errorf("failed to fetch hook policy for org_id %s: %w", orgID, err)
}

// Bad - loses original error
if err != nil {
    return fmt.Errorf("failed to fetch hook policy")
}
```

### Pattern: Add Context at Each Layer

```go
// Layer 1: Database
rows, err := db.Query(query, args...)
if err != nil {
    return fmt.Errorf("database query failed: %w", err)
}

// Layer 2: Handler
policy, err := fetchPolicy(orgID)
if err != nil {
    return fmt.Errorf("failed to fetch policy for org_id %s: %w", orgID, err)
}

// Layer 3: HTTP Handler
policy, err := getHookPolicy(orgID)
if err != nil {
    log.Printf("Error fetching hook policy: %v", err)
    http.Error(w, "Failed to fetch policy", http.StatusInternalServerError)
    return
}
```

## Error Aggregation Patterns

### Pattern: Collect Multiple Errors

```go
var errors []string

if err := validateUUID(orgID); err != nil {
    errors = append(errors, fmt.Sprintf("org_id: %v", err))
}
if err := validateRequired("agent_id", agentID); err != nil {
    errors = append(errors, fmt.Sprintf("agent_id: %v", err))
}

if len(errors) > 0 {
    return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
}
```

### Pattern: Return First Error or Aggregate

```go
func validateRequest(req *Request) error {
    var errs []error
    
    if err := validateUUID(req.OrgID); err != nil {
        errs = append(errs, err)
    }
    if err := validateHookType(req.HookType); err != nil {
        errs = append(errors, err)
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("validation failed: %v", errs)
    }
    return nil
}
```

## HTTP Error Responses

### Standard HTTP Status Codes

- **400 Bad Request**: Invalid input, validation errors
- **401 Unauthorized**: Missing or invalid authentication
- **403 Forbidden**: Valid auth but insufficient permissions
- **404 Not Found**: Resource doesn't exist
- **409 Conflict**: Resource conflict (e.g., duplicate entry)
- **500 Internal Server Error**: Unexpected server error
- **503 Service Unavailable**: Service temporarily unavailable

### Error Response Format

```json
{
  "error": "Validation failed",
  "message": "org_id: invalid UUID format",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "org_id",
    "value": "invalid-uuid"
  }
}
```

### Implementation Example

```go
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Message string                 `json:"message"`
    Code    string                 `json:"code,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func sendErrorResponse(w http.ResponseWriter, status int, err error, code string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    resp := ErrorResponse{
        Error:   http.StatusText(status),
        Message: err.Error(),
        Code:    code,
    }
    
    json.NewEncoder(w).Encode(resp)
}
```

## Error Recovery Patterns

### Pattern: Retry with Exponential Backoff

```go
func httpRequestWithRetry(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            backoff := time.Duration(100*(1<<uint(attempt-1))) * time.Millisecond
            logWarn("HTTP request failed (attempt %d/%d), retrying in %v...", attempt, maxRetries+1, backoff)
            time.Sleep(backoff)
        }
        
        resp, err := client.Do(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        
        lastErr = err
        if resp != nil {
            resp.Body.Close()
        }
    }
    
    return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries+1, lastErr)
}
```

### Pattern: Fallback to Cache

```go
policy, err := fetchPolicyFromHub(orgID)
if err != nil {
    logWarn("Failed to fetch policy from Hub: %v, using cached value", err)
    if cached := getCachedPolicy(orgID); cached != nil {
        return cached, nil
    }
    return nil, err
}
```

### Pattern: Graceful Degradation

```go
func performAuditForHook() (*AuditReport, error) {
    report := &AuditReport{
        CheckResults: make(map[string]CheckResult),
    }
    
    // Try security check, but don't fail if it errors
    if err := performSecurityAnalysisWithError(report); err != nil {
        logWarn("Security check failed: %v, continuing with other checks", err)
    }
    
    // Try vibe check, but don't fail if it errors
    if err := detectVibeIssuesWithError(report); err != nil {
        logWarn("Vibe check failed: %v, continuing with other checks", err)
    }
    
    return report, nil
}
```

## Panic Recovery

### Pattern: Recover from Panics

```go
func performSecurityAnalysisWithError(report *AuditReport) error {
    var err error
    defer func() {
        if r := recover(); r != nil {
            errMsg := fmt.Sprintf("Security check panicked: %v", r)
            logWarn("%s", errMsg)
            if report.CheckResults != nil {
                if cr, ok := report.CheckResults["security"]; ok {
                    cr.Success = false
                    cr.Error = errMsg
                    report.CheckResults["security"] = cr
                }
            }
            err = fmt.Errorf(errMsg)
        }
    }()
    
    // Actual implementation
    findings := performSecurityAnalysis()
    // ...
    
    return err
}
```

## Validation Error Patterns

### Pattern: Return Specific Validation Errors

```go
func validateRequest(req *Request) error {
    var errors []string
    
    if err := validateUUID(req.OrgID); err != nil {
        errors = append(errors, fmt.Sprintf("org_id: %v", err))
    }
    if err := validateRequired("agent_id", req.AgentID); err != nil {
        errors = append(errors, fmt.Sprintf("agent_id: %v", err))
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }
    return nil
}
```

## Database Error Patterns

### Pattern: Handle Database Errors Gracefully

```go
rows, err := queryWithTimeout(ctx, query, args...)
if err != nil {
    if err == context.DeadlineExceeded {
        logWarn("Database query timeout: %s", query)
        return nil, fmt.Errorf("query timeout: %w", err)
    }
    if err == sql.ErrNoRows {
        return nil, nil // Not an error, just no results
    }
    log.Printf("ERROR: Database query failed: %v", err)
    return nil, fmt.Errorf("database error: %w", err)
}
defer rows.Close()
```

## Testing Error Handling

### Pattern: Test Error Scenarios

```go
func TestHTTPRetry(t *testing.T) {
    // Test network error retry
    // Test 5xx server error retry
    // Test 4xx client error (no retry)
    // Test max retries exhausted
}
```

## Best Practices

1. **Always include context** in error messages
2. **Preserve original errors** using `%w` verb
3. **Log at appropriate levels** (info/warn/error)
4. **Return actionable errors** with clear messages
5. **Handle panics gracefully** with recovery
6. **Use standard HTTP status codes** consistently
7. **Validate inputs early** and return clear validation errors
8. **Implement retry logic** for transient failures
9. **Provide fallback mechanisms** when possible
10. **Test error scenarios** thoroughly

## Examples

### Complete Example: Handler with Full Error Handling

```go
func hookTelemetryHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var event HookTelemetryEvent
    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        log.Printf("Error decoding hook telemetry: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate inputs
    if err := validateRequired("agent_id", event.AgentID); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if err := validateHookType(event.HookType); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Database operation with timeout
    query := `INSERT INTO hook_executions ...`
    _, err := execWithTimeout(r.Context(), query, args...)
    if err != nil {
        if err == context.DeadlineExceeded {
            logWarn("Database timeout inserting hook execution: %v", err)
            http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
            return
        }
        log.Printf("ERROR: Database error inserting hook execution: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Hook execution recorded",
    })
}
```












