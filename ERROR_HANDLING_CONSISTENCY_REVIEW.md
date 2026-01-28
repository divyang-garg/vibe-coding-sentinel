# Error Handling Consistency Review

**Date:** January 23, 2026  
**Status:** ⚠️ **INCONSISTENCIES FOUND**  
**Priority:** Medium (should be addressed before production)

---

## Executive Summary

This review analyzes error handling patterns across all packages in the Sentinel codebase. While the codebase generally follows good error handling practices, several inconsistencies were identified that should be standardized for production readiness.

**Key Findings:**
- ✅ Good: Error wrapping with `%w` is used in most places
- ⚠️ Inconsistent: Database error handling (`sql.ErrNoRows`) uses direct comparison instead of `errors.Is()`
- ⚠️ Inconsistent: Duplicate error handler implementations exist
- ⚠️ Inconsistent: HTTP error response formatting varies
- ⚠️ Missing: Centralized error logging with context
- ✅ Good: Structured error types are defined and used

---

## 1. Error Wrapping Patterns

### ✅ **GOOD: Consistent Use of `%w` for Error Wrapping**

Most packages correctly use `fmt.Errorf` with `%w` to preserve error context:

**Examples:**
```go
// hub/api/services/helpers.go
return nil, fmt.Errorf("failed to get task: %w", err)

// hub/api/services/code_analysis_service.go
return nil, fmt.Errorf("failed to analyze security: %w", err)

// internal/cli/extract.go
return fmt.Errorf("extraction failed: %w", err)
```

**Status:** ✅ **COMPLIANT** - Most code follows this pattern

### ⚠️ **ISSUE: Some Direct Error Returns Without Context**

**Location:** Various service files

**Issue:** Some errors are returned directly without adding context:

```go
// hub/api/repository/task_repository_core.go:105
if err != nil {
    return nil, err  // No context added
}
```

**Recommendation:** Always wrap errors with context:
```go
if err != nil {
    return nil, fmt.Errorf("failed to find task by ID %s: %w", id, err)
}
```

---

## 2. Database Error Handling

### ❌ **CRITICAL: Inconsistent `sql.ErrNoRows` Handling**

**Issue:** Direct comparison with `sql.ErrNoRows` instead of using `errors.Is()`

**Current Pattern (Inconsistent):**
```go
// Found in 20+ locations across hub/api/services and hub/api/utils
if err == sql.ErrNoRows {
    return nil, fmt.Errorf("task not found: %s", id)
}
```

**Problem:** 
- Direct comparison (`==`) doesn't work with wrapped errors
- If an error is wrapped with `fmt.Errorf("...: %w", err)`, the comparison fails
- This breaks error handling when errors are properly wrapped

**Locations Affected:**
- `hub/api/services/helpers.go:167`
- `hub/api/utils/task_integrations_core.go` (6 instances)
- `hub/api/services/dependency_detector_helpers.go:49`
- `hub/api/services/test_service.go` (3 instances)
- `hub/api/services/change_request_manager.go:45`
- `hub/api/services/test_sandbox.go:226`
- `hub/api/services/implementation_tracker.go:84`
- `hub/api/services/test_validator.go:219`
- `hub/api/services/test_coverage_tracker.go:220`
- `hub/api/services/policy.go:153`
- `hub/api/services/knowledge_service.go:242`
- `hub/api/services/mutation_engine.go:332`
- `hub/api/test_validator.go:247`
- `hub/api/repository/shared_storage.go` (4 instances)
- `hub/api/repository/task_storage.go:53`

**Correct Pattern:**
```go
import (
    "database/sql"
    "errors"
)

if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, &models.NotFoundError{
            Resource: "task",
            Message:  fmt.Sprintf("task not found: %s", id),
        }
    }
    return nil, fmt.Errorf("failed to find task: %w", err)
}
```

**Impact:** High - This can cause "not found" errors to be treated as internal server errors when errors are properly wrapped.

**Recommendation:** Replace all `err == sql.ErrNoRows` with `errors.Is(err, sql.ErrNoRows)`

---

## 3. HTTP Error Response Handling

### ⚠️ **ISSUE: Duplicate Error Handler Implementations**

**Problem:** Two separate error handler files exist with similar functionality:

1. `hub/api/error_handler.go` (package `main`)
2. `hub/api/handlers/error_handler.go` (package `handlers`)

**Analysis:**
- Both define the same error types: `ValidationError`, `NotFoundError`, `DatabaseError`, `InternalError`, `ExternalServiceError`
- Both have `WriteErrorResponse()` functions with similar logic
- The handlers package version has a local `LogError()` implementation
- The main package version references a global `LogError()` function

**Current Usage:**
- Handlers use `handlers.WriteErrorResponse()` from `hub/api/handlers/error_handler.go`
- Base handler uses its own `WriteErrorResponse()` method in `hub/api/handlers/base.go`

**Recommendation:**
1. Consolidate error handlers into a single package (preferably `hub/api/handlers`)
2. Remove duplicate `hub/api/error_handler.go` (if not used)
3. Standardize on one error response format

### ⚠️ **ISSUE: Inconsistent Error Response Format**

**Pattern 1: BaseHandler.WriteErrorResponse()**
```go
// hub/api/handlers/base.go:29
errorResponse := map[string]interface{}{
    "error": err.Error(),
}
```

**Pattern 2: handlers.WriteErrorResponse()**
```go
// hub/api/handlers/error_handler.go:109
response := map[string]interface{}{
    "success": false,
    "error": map[string]interface{}{
        "type":    errorType,
        "message": errorMessage,
        "details": errorDetails,
    },
}
```

**Pattern 3: Direct http.Error()**
```go
// hub/api/middleware/security.go:66
http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
```

**Recommendation:** Standardize on Pattern 2 (structured error response) for all API endpoints.

---

## 4. Error Logging Patterns

### ⚠️ **ISSUE: Inconsistent Error Logging**

**Current State:**
- `hub/api/handlers/error_handler.go` has `LogErrorWithContext()` but uses a simple `log.Printf()`
- No centralized structured logging
- Context information (request ID, user ID) is extracted but logging is basic

**Good Pattern (Exists but needs standardization):**
```go
// hub/api/handlers/error_handler.go:188
func LogErrorWithContext(ctx context.Context, err error, message string, additionalContext ...interface{}) {
    requestID := "unknown"
    if id, ok := ctx.Value("request_id").(string); ok {
        requestID = id
    }
    // ... logs with context
}
```

**Issues:**
1. Not consistently used across all error paths
2. Uses basic `log.Printf()` instead of structured logger
3. Stack traces are captured but may be too verbose for production

**Recommendation:**
1. Use structured logging (e.g., `zerolog`, `zap`, or `logrus`)
2. Ensure all error paths log with context
3. Configure log levels appropriately (ERROR for errors, WARN for recoverable issues)

---

## 5. Structured Error Types

### ✅ **GOOD: Well-Defined Error Types**

**Location:** `hub/api/models/types.go` and `hub/api/handlers/error_handler.go`

**Defined Types:**
- `ValidationError` - Field validation failures
- `NotFoundError` - Resource not found
- `DatabaseError` - Database operation failures
- `InternalError` - Internal server errors
- `ExternalServiceError` - External service failures
- `NotImplementedError` - Feature not implemented
- `RateLimitError` - Rate limiting errors

**Status:** ✅ **GOOD** - Error types are well-defined with HTTP status code methods

**Issue:** Not all code paths use these structured types consistently.

**Recommendation:** 
1. Use structured error types instead of plain `fmt.Errorf()` where appropriate
2. Ensure all handlers check for structured error types and map to appropriate HTTP status codes

---

## 6. Service Layer Error Handling

### ✅ **GOOD: Consistent Error Wrapping in Services**

**Pattern:**
```go
// hub/api/services/task_service_core.go:34
if err != nil {
    return nil, fmt.Errorf("failed to get dependencies: %w", err)
}
```

**Status:** ✅ **COMPLIANT** - Services consistently wrap errors with context

### ⚠️ **ISSUE: Missing Error Type Mapping**

**Issue:** Services return generic errors instead of structured error types:

```go
// Current
return nil, fmt.Errorf("task not found")

// Should be
return nil, &models.NotFoundError{
    Resource: "task",
    Message:  "task not found",
}
```

**Recommendation:** Services should return structured error types that handlers can map to HTTP status codes.

---

## 7. Repository Layer Error Handling

### ⚠️ **ISSUE: Repositories Return Raw Database Errors**

**Current Pattern:**
```go
// hub/api/repository/task_repository_core.go:82
_, err := r.db.Exec(ctx, query, ...)
return err  // Raw database error
```

**Issue:** 
- Raw database errors leak implementation details
- No context about what operation failed
- Handlers can't distinguish between different error types

**Recommendation:**
```go
_, err := r.db.Exec(ctx, query, ...)
if err != nil {
    return fmt.Errorf("failed to save task %s: %w", task.ID, err)
}
```

**For "not found" cases:**
```go
err := row.Scan(...)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, &models.NotFoundError{
            Resource: "task",
            ID:       id,
        }
    }
    return nil, fmt.Errorf("failed to find task: %w", err)
}
```

---

## 8. CLI Error Handling

### ✅ **GOOD: Consistent Error Wrapping**

**Pattern:**
```go
// internal/cli/extract.go:37
return fmt.Errorf("document parsing failed: %w", err)
```

**Status:** ✅ **COMPLIANT** - CLI consistently wraps errors

### ⚠️ **ISSUE: Error Messages Not User-Friendly**

**Issue:** Some error messages are too technical for end users:

```go
// Too technical
return fmt.Errorf("failed to unmarshal JSON: %w", err)

// Better
return fmt.Errorf("failed to parse response: %w", err)
```

**Recommendation:** Use user-friendly error messages in CLI, with technical details in logs.

---

## Recommendations Summary

### High Priority

1. **Replace `sql.ErrNoRows` comparisons with `errors.Is()`**
   - **Impact:** High - Fixes broken error handling for wrapped errors
   - **Effort:** Medium - 20+ locations to update
   - **Files:** All files in `hub/api/services/` and `hub/api/repository/`

2. **Consolidate duplicate error handlers**
   - **Impact:** Medium - Reduces maintenance burden
   - **Effort:** Low - Remove unused file, standardize on one implementation
   - **Files:** `hub/api/error_handler.go`, `hub/api/handlers/error_handler.go`

3. **Standardize HTTP error response format**
   - **Impact:** Medium - Consistent API responses
   - **Effort:** Low - Update handlers to use standardized format
   - **Files:** All handler files

### Medium Priority

4. **Add structured logging**
   - **Impact:** Medium - Better observability
   - **Effort:** Medium - Integrate structured logger, update all log calls
   - **Files:** All packages

5. **Use structured error types in services**
   - **Impact:** Medium - Better error handling
   - **Effort:** Medium - Update service methods to return structured errors
   - **Files:** All service files

6. **Add context to repository errors**
   - **Impact:** Low - Better error messages
   - **Effort:** Low - Wrap errors in repositories
   - **Files:** All repository files

### Low Priority

7. **Improve CLI error messages**
   - **Impact:** Low - Better user experience
   - **Effort:** Low - Update error messages
   - **Files:** `internal/cli/` files

---

## Compliance with CODING_STANDARDS.md

### ✅ **COMPLIANT:**
- Error wrapping with `%w` (Section 4.1)
- Structured error types (Section 4.2)

### ⚠️ **NEEDS IMPROVEMENT:**
- Error handling consistency (Section 4.1 - should use `errors.Is()` for error checking)
- Logging levels (Section 4.3 - needs structured logging implementation)

---

## Action Items

### Immediate (This Week)
- [ ] Replace all `err == sql.ErrNoRows` with `errors.Is(err, sql.ErrNoRows)`
- [ ] Consolidate duplicate error handler files
- [ ] Standardize HTTP error response format

### Short-term (Next 2 Weeks)
- [ ] Implement structured logging
- [ ] Update services to return structured error types
- [ ] Add context to repository error returns

### Long-term (Next Month)
- [ ] Improve CLI error messages
- [ ] Add error handling tests
- [ ] Document error handling patterns

---

## Testing Recommendations

1. **Add tests for error wrapping:**
   ```go
   func TestErrorWrapping(t *testing.T) {
       originalErr := sql.ErrNoRows
       wrappedErr := fmt.Errorf("failed to find task: %w", originalErr)
       
       // Should be able to detect original error
       assert.True(t, errors.Is(wrappedErr, sql.ErrNoRows))
   }
   ```

2. **Add tests for error type mapping:**
   ```go
   func TestErrorToHTTPStatus(t *testing.T) {
       err := &models.NotFoundError{Resource: "task"}
       status := getHTTPStatus(err)
       assert.Equal(t, http.StatusNotFound, status)
   }
   ```

3. **Add integration tests for error scenarios:**
   - Test database connection failures
   - Test "not found" scenarios
   - Test validation errors
   - Test external service failures

---

## Conclusion

The codebase demonstrates good error handling practices overall, with consistent use of error wrapping and well-defined error types. However, several inconsistencies need to be addressed:

1. **Critical:** Database error handling must use `errors.Is()` instead of direct comparison
2. **Important:** Duplicate error handlers should be consolidated
3. **Important:** HTTP error responses should be standardized
4. **Recommended:** Structured logging should be implemented
5. **Recommended:** Services should return structured error types

Addressing these issues will improve error handling consistency, maintainability, and production readiness.

---

**Reviewer:** AI Assistant  
**Next Review:** After fixes are implemented
