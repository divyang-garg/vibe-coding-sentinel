# Error Handling Fix Plan

**Date:** January 23, 2026  
**Status:** 📋 **PLANNING**  
**Priority:** High (Critical for Production Readiness)  
**Compliance:** All fixes must comply with `docs/external/CODING_STANDARDS.md`

---

## Executive Summary

This plan addresses all issues identified in `ERROR_HANDLING_CONSISTENCY_REVIEW.md` while ensuring full compliance with the coding standards. The plan is organized by priority and includes detailed implementation steps, code examples, and testing requirements.

**Total Issues:** 7  
**Files Affected:** 30+  
**Estimated Effort:** 2-3 weeks  
**Risk Level:** Low (well-defined changes)

---

## Phase 1: Critical Fixes (Week 1)

### Fix 1.1: Replace `sql.ErrNoRows` Comparisons with `errors.Is()`

**Priority:** 🔴 **CRITICAL**  
**Impact:** High - Fixes broken error handling for wrapped errors  
**Effort:** Medium - 24 files, ~50 locations  
**Compliance:** Section 4.1 (Error Wrapping)

#### Issue Description
Direct comparison (`err == sql.ErrNoRows`) fails when errors are wrapped with `%w`. This breaks error handling throughout the codebase.

#### Implementation Plan

**Step 1: Create Helper Function (Compliance: Section 8.2 - Service Organization)**

Create a shared utility function to handle database "not found" errors consistently:

**File:** `hub/api/repository/errors.go` (NEW FILE)
```go
// Package repository provides error handling utilities for database operations.
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"sentinel-hub-api/models"
)

// HandleNotFoundError checks if an error is sql.ErrNoRows and returns
// a structured NotFoundError if so, otherwise wraps the error with context.
//
// This function ensures consistent error handling across all repository methods
// and properly handles wrapped errors using errors.Is() as required by
// CODING_STANDARDS.md Section 4.1.
func HandleNotFoundError(err error, resource, id string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return &models.NotFoundError{
			Resource: resource,
			Message:  fmt.Sprintf("%s not found: %s", resource, id),
		}
	}

	return fmt.Errorf("failed to find %s %s: %w", resource, id, err)
}

// WrapDatabaseError wraps a database error with context about the operation.
// Complies with CODING_STANDARDS.md Section 4.1: Error Wrapping
func WrapDatabaseError(err error, operation, resource string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return &models.NotFoundError{
			Resource: resource,
			Message:  fmt.Sprintf("%s not found during %s", resource, operation),
		}
	}

	return fmt.Errorf("database error during %s on %s: %w", operation, resource, err)
}
```

**Step 2: Update All Repository Files**

For each file with `err == sql.ErrNoRows`, replace with the helper function:

**Pattern to Replace:**
```go
// ❌ OLD (Non-compliant)
err := row.Scan(...)
if err != nil {
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("task not found: %s", id)
    }
    return nil, fmt.Errorf("failed to find task: %w", err)
}
```

**Replacement:**
```go
// ✅ NEW (Compliant with CODING_STANDARDS.md Section 4.1)
import "sentinel-hub-api/repository"

err := row.Scan(...)
if err != nil {
    return nil, repository.HandleNotFoundError(err, "task", id)
}
```

**Files to Update (24 files):**

1. `hub/api/services/helpers.go:167`
   ```go
   // Before
   if err == sql.ErrNoRows {
       return nil, fmt.Errorf("task not found")
   }
   
   // After
   if err != nil {
       return nil, repository.HandleNotFoundError(err, "task", taskID)
   }
   ```

2. `hub/api/utils/task_integrations_core.go` (6 instances)
   - Update each instance to use `repository.HandleNotFoundError()`
   - Ensure proper resource name ("task", "project", etc.)

3. `hub/api/services/dependency_detector_helpers.go:49`
4. `hub/api/services/test_service.go` (3 instances)
5. `hub/api/services/change_request_manager.go:45`
6. `hub/api/services/test_sandbox.go:226`
7. `hub/api/services/implementation_tracker.go:84`
8. `hub/api/services/test_validator.go:219`
9. `hub/api/services/test_coverage_tracker.go:220`
10. `hub/api/services/policy.go:153`
11. `hub/api/services/knowledge_service.go:242`
12. `hub/api/services/mutation_engine.go:332`
13. `hub/api/test_validator.go:247`
14. `hub/api/repository/shared_storage.go` (4 instances)
15. `hub/api/repository/task_storage.go:53`
16. Additional files from grep results

**Step 3: Add Tests**

**File:** `hub/api/repository/errors_test.go` (NEW FILE)
```go
package repository

import (
	"database/sql"
	"errors"
	"testing"

	"sentinel-hub-api/models"
)

func TestHandleNotFoundError(t *testing.T) {
	t.Run("sql.ErrNoRows returns NotFoundError", func(t *testing.T) {
		err := HandleNotFoundError(sql.ErrNoRows, "task", "123")
		
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Fatalf("expected NotFoundError, got %T", err)
		}
		
		if notFoundErr.Resource != "task" {
			t.Errorf("expected resource 'task', got %s", notFoundErr.Resource)
		}
	})

	t.Run("wrapped sql.ErrNoRows returns NotFoundError", func(t *testing.T) {
		wrappedErr := fmt.Errorf("database query failed: %w", sql.ErrNoRows)
		err := HandleNotFoundError(wrappedErr, "task", "123")
		
		var notFoundErr *models.NotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Fatalf("expected NotFoundError for wrapped error, got %T", err)
		}
	})

	t.Run("other errors are wrapped with context", func(t *testing.T) {
		originalErr := errors.New("connection failed")
		err := HandleNotFoundError(originalErr, "task", "123")
		
		if errors.Is(err, sql.ErrNoRows) {
			t.Error("should not be sql.ErrNoRows")
		}
		
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
```

**Compliance Checklist:**
- ✅ Uses `errors.Is()` for error checking (Section 4.1)
- ✅ Wraps errors with `%w` (Section 4.1)
- ✅ Returns structured error types (Section 4.2)
- ✅ File size < 250 lines (Utilities limit)
- ✅ Includes tests (Section 6.1)

---

### Fix 1.2: Consolidate Duplicate Error Handlers

**Priority:** 🔴 **HIGH**  
**Impact:** Medium - Reduces maintenance burden  
**Effort:** Low - 2 files  
**Compliance:** Section 1.2 (Layer Separation), Section 8.1 (Handler Organization)

#### Issue Description
Two error handler implementations exist:
- `hub/api/error_handler.go` (package `main`)
- `hub/api/handlers/error_handler.go` (package `handlers`)

#### Implementation Plan

**Step 1: Verify Usage**

Check which error handler is actually used:
```bash
grep -r "WriteErrorResponse\|LogErrorWithContext" hub/api --include="*.go" | grep -v test
```

**Step 2: Consolidate to Handlers Package**

Since handlers are in `hub/api/handlers/`, consolidate there:

**File:** `hub/api/handlers/error_handler.go` (UPDATE)
- Keep the handlers package version
- Enhance with missing features from main package version if needed
- Ensure compliance with Section 4.2 (Structured Error Types)

**Step 3: Remove Duplicate**

**File:** `hub/api/error_handler.go` (DELETE)
- Only if not used by main package
- Check `hub/api/main.go` or entry point first

**Step 4: Update BaseHandler**

**File:** `hub/api/handlers/base.go` (UPDATE)
```go
// Update WriteErrorResponse to use standardized format
func (h *BaseHandler) WriteErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	// Use the standardized WriteErrorResponse from error_handler.go
	WriteErrorResponse(w, err, statusCode)
}
```

**Compliance Checklist:**
- ✅ Single error handler implementation (Section 8.1)
- ✅ Handlers package organization (Section 1.2)
- ✅ Structured error types (Section 4.2)

---

### Fix 1.3: Standardize HTTP Error Response Format

**Priority:** 🔴 **HIGH**  
**Impact:** Medium - Consistent API responses  
**Effort:** Low - Update handlers  
**Compliance:** Section 4.2 (Structured Error Types), Section 8.1 (Handler Organization)

#### Issue Description
Three different error response formats are used:
1. BaseHandler: `{"error": "message"}`
2. handlers.WriteErrorResponse: `{"success": false, "error": {"type": "...", "message": "...", "details": {...}}}`
3. Direct http.Error(): Plain text

#### Implementation Plan

**Step 1: Standardize Error Response Format**

**File:** `hub/api/handlers/error_handler.go` (UPDATE)

Ensure `WriteErrorResponse()` uses the structured format consistently:

```go
// WriteErrorResponse writes a standardized error response.
// Complies with CODING_STANDARDS.md Section 4.2: Structured Error Types
func WriteErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	// Determine error type and status code
	var errorType string
	var errorMessage string
	var errorDetails map[string]interface{}

	// Check for structured error types first (Section 4.2)
	switch e := err.(type) {
	case *models.ValidationError:
		errorType = "validation_error"
		errorMessage = e.Error()
		statusCode = http.StatusBadRequest
		errorDetails = map[string]interface{}{
			"field":   e.Field,
			"message": e.Message,
		}
	case *models.NotFoundError:
		errorType = "not_found_error"
		errorMessage = e.Error()
		statusCode = http.StatusNotFound
		errorDetails = map[string]interface{}{
			"resource": e.Resource,
		}
	// ... other error types
	default:
		errorType = "internal_error"
		errorMessage = err.Error()
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
	}

	// Build standardized response (Section 4.2)
	response := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"type":    errorType,
			"message": errorMessage,
		},
	}

	if errorDetails != nil {
		response["error"].(map[string]interface{})["details"] = errorDetails
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
```

**Step 2: Update BaseHandler**

**File:** `hub/api/handlers/base.go` (UPDATE)
```go
// WriteErrorResponse delegates to standardized error handler
func (h *BaseHandler) WriteErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	WriteErrorResponse(w, err, statusCode)
}
```

**Step 3: Update Middleware**

**File:** `hub/api/middleware/security.go` (UPDATE)
```go
// Replace direct http.Error() with structured response
if !limiter.Allow() {
	WriteErrorResponse(w, &models.RateLimitError{
		Message: "Rate limit exceeded",
	}, http.StatusTooManyRequests)
	return
}
```

**Compliance Checklist:**
- ✅ Structured error response format (Section 4.2)
- ✅ Consistent across all handlers (Section 8.1)
- ✅ Proper HTTP status codes

---

## Phase 2: Important Improvements (Week 2)

### Fix 2.1: Add Context to Repository Errors

**Priority:** 🟡 **MEDIUM**  
**Impact:** Low - Better error messages  
**Effort:** Low - Wrap errors in repositories  
**Compliance:** Section 4.1 (Error Wrapping), Section 1.2 (Layer Separation)

#### Implementation Plan

**Pattern to Apply:**

**Before:**
```go
// hub/api/repository/task_repository_core.go:82
_, err := r.db.Exec(ctx, query, ...)
return err  // Raw database error
```

**After:**
```go
_, err := r.db.Exec(ctx, query, ...)
if err != nil {
	return fmt.Errorf("failed to save task %s: %w", task.ID, err)
}
return nil
```

**Files to Update:**
- All repository files that return raw errors
- Use `repository.WrapDatabaseError()` helper where appropriate

**Compliance Checklist:**
- ✅ Error wrapping with context (Section 4.1)
- ✅ Repository layer only (Section 1.2)

---

### Fix 2.2: Use Structured Error Types in Services

**Priority:** 🟡 **MEDIUM**  
**Impact:** Medium - Better error handling  
**Effort:** Medium - Update service methods  
**Compliance:** Section 4.2 (Structured Error Types), Section 1.2 (Layer Separation)

#### Implementation Plan

**Pattern to Apply:**

**Before:**
```go
// Current
return nil, fmt.Errorf("task not found")
```

**After:**
```go
// Should be
return nil, &models.NotFoundError{
	Resource: "task",
	Message:  "task not found",
}
```

**Files to Update:**
- All service files that return generic errors for "not found" cases
- All service files that return validation errors
- Map to appropriate structured error types

**Compliance Checklist:**
- ✅ Structured error types (Section 4.2)
- ✅ Service layer only (Section 1.2)

---

### Fix 2.3: Implement Structured Logging

**Priority:** 🟡 **MEDIUM**  
**Impact:** Medium - Better observability  
**Effort:** Medium - Integrate structured logger  
**Compliance:** Section 4.3 (Logging Levels)

#### Implementation Plan

**Step 1: Choose Structured Logger**

Recommendation: Use `zerolog` (lightweight, fast, structured)

**File:** `hub/api/pkg/logging/logger.go` (NEW FILE)
```go
// Package logging provides structured logging for the Sentinel Hub API.
// Complies with CODING_STANDARDS.md Section 4.3: Logging Levels
package logging

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

func NewLogger() *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()

	return &Logger{logger: logger}
}

// Debug logs detailed diagnostic information (Section 4.3)
func (l *Logger) Debug(ctx context.Context, msg string, fields ...interface{}) {
	l.logger.Debug().Fields(fields).Msg(msg)
}

// Info logs normal operational messages (Section 4.3)
func (l *Logger) Info(ctx context.Context, msg string, fields ...interface{}) {
	l.logger.Info().Fields(fields).Msg(msg)
}

// Warn logs unexpected but recoverable events (Section 4.3)
func (l *Logger) Warn(ctx context.Context, msg string, fields ...interface{}) {
	l.logger.Warn().Fields(fields).Msg(msg)
}

// Error logs error conditions requiring attention (Section 4.3)
func (l *Logger) Error(ctx context.Context, err error, msg string, fields ...interface{}) {
	l.logger.Error().
		Err(err).
		Fields(fields).
		Msg(msg)
}

// WithContext adds request context to logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	requestID := "unknown"
	if id, ok := ctx.Value("request_id").(string); ok {
		requestID = id
	}

	userID := "unknown"
	if id, ok := ctx.Value("user_id").(string); ok {
		userID = id
	}

	return &Logger{
		logger: l.logger.With().
			Str("request_id", requestID).
			Str("user_id", userID).
			Logger(),
	}
}
```

**Step 2: Update Error Handler**

**File:** `hub/api/handlers/error_handler.go` (UPDATE)
```go
// Replace log.Printf with structured logger
func LogErrorWithContext(ctx context.Context, err error, message string, additionalContext ...interface{}) {
	logger := logging.NewLogger().WithContext(ctx)
	logger.Error(ctx, err, message, additionalContext...)
}
```

**Step 3: Update All Log Calls**

Replace all `log.Printf`, `log.Println`, etc. with structured logger calls.

**Compliance Checklist:**
- ✅ Structured logging (Section 4.3)
- ✅ Proper log levels (Section 4.3)
- ✅ Context preservation

---

## Phase 3: Nice-to-Have Improvements (Week 3)

### Fix 3.1: Improve CLI Error Messages

**Priority:** 🟢 **LOW**  
**Impact:** Low - Better user experience  
**Effort:** Low - Update error messages  
**Compliance:** Section 4.1 (Error Wrapping)

#### Implementation Plan

**Pattern to Apply:**

**Before:**
```go
return fmt.Errorf("failed to unmarshal JSON: %w", err)
```

**After:**
```go
return fmt.Errorf("failed to parse response: %w", err)
```

**Files to Update:**
- `internal/cli/extract.go`
- `internal/cli/audit.go`
- Other CLI files with technical error messages

**Compliance Checklist:**
- ✅ User-friendly messages
- ✅ Error wrapping preserved (Section 4.1)

---

## Testing Requirements

### Unit Tests

For each fix, add unit tests:

1. **Error Wrapping Tests:**
   ```go
   func TestErrorWrapping(t *testing.T) {
       originalErr := sql.ErrNoRows
       wrappedErr := fmt.Errorf("failed to find task: %w", originalErr)
       
       assert.True(t, errors.Is(wrappedErr, sql.ErrNoRows))
   }
   ```

2. **Error Type Mapping Tests:**
   ```go
   func TestErrorToHTTPStatus(t *testing.T) {
       err := &models.NotFoundError{Resource: "task"}
       status := getHTTPStatus(err)
       assert.Equal(t, http.StatusNotFound, status)
   }
   ```

3. **Repository Error Handling Tests:**
   ```go
   func TestHandleNotFoundError(t *testing.T) {
       err := repository.HandleNotFoundError(sql.ErrNoRows, "task", "123")
       assert.IsType(t, &models.NotFoundError{}, err)
   }
   ```

### Integration Tests

Add integration tests for:
- Database connection failures
- "Not found" scenarios
- Validation errors
- External service failures

**Compliance:** Section 6.1 (Test Coverage Requirements)

---

## Implementation Checklist

### Phase 1: Critical Fixes (Week 1)
- [ ] Create `hub/api/repository/errors.go` with helper functions
- [ ] Add tests for error helper functions
- [ ] Update all 24 files with `err == sql.ErrNoRows` comparisons
- [ ] Verify duplicate error handler usage
- [ ] Consolidate error handlers to `hub/api/handlers/error_handler.go`
- [ ] Remove duplicate `hub/api/error_handler.go` (if unused)
- [ ] Standardize `WriteErrorResponse()` format
- [ ] Update `BaseHandler.WriteErrorResponse()`
- [ ] Update middleware to use structured error responses
- [ ] Run all tests and verify fixes

### Phase 2: Important Improvements (Week 2)
- [ ] Add context to all repository error returns
- [ ] Update services to return structured error types
- [ ] Create structured logging package
- [ ] Update error handler to use structured logger
- [ ] Replace all log calls with structured logger
- [ ] Run all tests and verify improvements

### Phase 3: Nice-to-Have (Week 3)
- [ ] Improve CLI error messages
- [ ] Add error handling documentation
- [ ] Final review and testing

---

## Risk Assessment

### Low Risk
- All changes are well-defined
- Backward compatible (error types remain the same)
- Comprehensive test coverage planned

### Mitigation
- Implement in phases
- Test after each phase
- Rollback plan: Git revert if issues arise

---

## Success Criteria

1. ✅ All `err == sql.ErrNoRows` replaced with `errors.Is()`
2. ✅ Single error handler implementation
3. ✅ Consistent HTTP error response format
4. ✅ All repository errors wrapped with context
5. ✅ Services return structured error types
6. ✅ Structured logging implemented
7. ✅ All tests passing
8. ✅ Code coverage maintained at 80%+
9. ✅ Compliance with CODING_STANDARDS.md verified

---

## Compliance Verification

After implementation, verify compliance with:

- [ ] Section 4.1: Error Wrapping - All errors wrapped with `%w`
- [ ] Section 4.2: Structured Error Types - Used consistently
- [ ] Section 4.3: Logging Levels - Structured logging implemented
- [ ] Section 1.2: Layer Separation - No violations
- [ ] Section 8.1: Handler Organization - Consistent patterns
- [ ] Section 6.1: Test Coverage - 80%+ maintained

---

**Plan Owner:** Engineering Team  
**Review Date:** After Phase 1 completion  
**Next Steps:** Begin Phase 1 implementation
