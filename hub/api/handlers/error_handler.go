// Error Handler
// Standardized error handling and error types across the application

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"sentinel-hub-api/models"
	"sentinel-hub-api/pkg"
)

// LogError logs an error message using the package logger
func LogError(ctx context.Context, msg string, args ...interface{}) {
	pkg.LogError(ctx, msg, args...)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// HTTPStatus returns the HTTP status code for validation errors
func (e *ValidationError) HTTPStatus() int {
	return http.StatusBadRequest
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
	Message  string
}

func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s with ID '%s' not found", e.Resource, e.ID)
}

// HTTPStatus returns the HTTP status code for not found errors
func (e *NotFoundError) HTTPStatus() int {
	return http.StatusNotFound
}

// DatabaseError represents a database operation error
type DatabaseError struct {
	Operation     string
	Message       string
	Code          string
	OriginalError error
}

// InternalError represents an internal server error
type InternalError struct {
	Message string
	Code    string
}

func (e *InternalError) Error() string {
	return e.Message
}

func (e *InternalError) HTTPStatus() int {
	return http.StatusInternalServerError
}

func (e *DatabaseError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("database error in %s: %s (original: %v)", e.Operation, e.Message, e.OriginalError)
	}
	return fmt.Sprintf("database error in %s: %s", e.Operation, e.Message)
}

// HTTPStatus returns the HTTP status code for database errors
func (e *DatabaseError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// ExternalServiceError represents an external service error
type ExternalServiceError struct {
	Service    string
	Message    string
	StatusCode int
}

func (e *ExternalServiceError) Error() string {
	return fmt.Sprintf("external service error (%s): %s", e.Service, e.Message)
}

// HTTPStatus returns the HTTP status code for external service errors
func (e *ExternalServiceError) HTTPStatus() int {
	if e.StatusCode > 0 {
		return e.StatusCode
	}
	return http.StatusBadGateway
}

// WriteErrorResponse writes a standardized error response.
// The statusCode parameter may be 0 for unknown errors, in which case it will be set to 500.
func WriteErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	// Determine error type and status code
	var errorType string
	var errorMessage string
	var errorDetails map[string]interface{}

	switch e := err.(type) {
	case *ValidationError:
		errorType = "validation_error"
		errorMessage = e.Error()
		statusCode = e.HTTPStatus()
		errorDetails = map[string]interface{}{
			"field":   e.Field,
			"code":    e.Code,
			"message": e.Message,
		}
	case *models.ValidationError:
		errorType = "validation_error"
		errorMessage = e.Error()
		statusCode = http.StatusBadRequest
		errorDetails = map[string]interface{}{
			"field":   e.Field,
			"message": e.Message,
		}
	case *NotFoundError:
		errorType = "not_found_error"
		errorMessage = e.Error()
		statusCode = e.HTTPStatus()
		errorDetails = map[string]interface{}{
			"resource": e.Resource,
			"id":       e.ID,
		}
	case *models.NotFoundError:
		errorType = "not_found_error"
		errorMessage = e.Error()
		statusCode = http.StatusNotFound
		errorDetails = map[string]interface{}{
			"resource": e.Resource,
		}
	case *DatabaseError:
		errorType = "database_error"
		errorMessage = e.Error()
		statusCode = e.HTTPStatus()
		errorDetails = map[string]interface{}{
			"operation": e.Operation,
		}
	case *ExternalServiceError:
		errorType = "external_service_error"
		errorMessage = e.Error()
		statusCode = e.HTTPStatus()
		errorDetails = map[string]interface{}{
			"service": e.Service,
		}
	case *InternalError:
		errorType = "internal_error"
		errorMessage = e.Error()
		statusCode = e.HTTPStatus()
		errorDetails = map[string]interface{}{
			"code": e.Code,
		}
	case *models.RateLimitError:
		errorType = "rate_limit_error"
		errorMessage = e.Error()
		statusCode = http.StatusTooManyRequests
		errorDetails = map[string]interface{}{
			"retry_after": e.RetryAfter,
			"reset_time":  e.ResetTime,
		}
	default:
		errorType = "internal_error"
		errorMessage = err.Error()
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
	}

	// Build response
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Cannot call http.Error after WriteHeader, so log the error
		// The response has already been sent, so we can only log
		// In production, this should use proper logging
		_ = err // Log error in production: log.Printf("Failed to encode error response: %v", err)
	}
}

// WriteJSONResponse writes a JSON response with status code
func WriteJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Cannot call http.Error after WriteHeader, so log the error
		// The response has already been sent, so we can only log
		// In production, this should use proper logging
		_ = err // Log error in production: log.Printf("Failed to encode JSON response: %v", err)
	}
}

// LogErrorWithContext logs an error with request context
func LogErrorWithContext(ctx context.Context, err error, message string, additionalContext ...interface{}) {
	// Get request ID from context if available
	requestID := "unknown"
	if id, ok := ctx.Value("request_id").(string); ok {
		requestID = id
	}

	// Get user context if available
	userID := "unknown"
	if id, ok := ctx.Value("user_id").(string); ok {
		userID = id
	}

	// Get stack trace
	stackTrace := getStackTrace()

	// Format additional context if provided
	contextStr := ""
	if len(additionalContext) > 0 {
		if contextMap, ok := additionalContext[0].(map[string]interface{}); ok {
			for k, v := range contextMap {
				contextStr += fmt.Sprintf(", %s=%v", k, v)
			}
		}
	} else {
		// Variadic parameter may be empty - this is intentional and valid
		_ = additionalContext
	}

	// Log error with context using structured logger
	pkg.LogErrorWithErr(ctx, err, message, 
		"request_id", requestID,
		"user_id", userID,
		"context", contextStr,
		"stack_trace", stackTrace)
}

// getStackTrace returns a formatted stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// RecoverFromPanic handles panics gracefully
func RecoverFromPanic(w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		// Log panic
		LogErrorWithContext(r.Context(), fmt.Errorf("panic: %v", rec), "Panic recovered")

		// Write error response
		err := fmt.Errorf("internal server error: panic occurred")
		WriteErrorResponse(w, err, http.StatusInternalServerError)
	}
}
