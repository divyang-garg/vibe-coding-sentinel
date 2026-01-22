// Package middleware - Input validation middleware
// Complies with CODING_STANDARDS.md: HTTP middleware max 300 lines
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sentinel-hub-api/validation"
)

// ValidationMiddlewareConfig holds validation configuration
type ValidationMiddlewareConfig struct {
	Validator validation.Validator
	MaxSize   int64 // Maximum request body size in bytes
}

// ValidationMiddleware creates middleware for request validation
func ValidationMiddleware(config ValidationMiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip validation for GET/HEAD/DELETE with no body
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "DELETE" {
				// Validate query parameters only (if needed)
				next.ServeHTTP(w, r)
				return
			}

			// Check request size
			if config.MaxSize > 0 && r.ContentLength > config.MaxSize {
				http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Read and parse request body
			bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, config.MaxSize))
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			// Skip validation if body is empty
			if len(bodyBytes) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Parse JSON
			var body map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &body); err != nil {
				http.Error(w, "Invalid JSON format", http.StatusBadRequest)
				return
			}

			// Validate if validator is provided
			if config.Validator != nil {
				if err := config.Validator.Validate(body); err != nil {
					// Return validation error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": err.Error(),
						"field": getFieldFromError(err),
					})
					return
				}
			}

			// Re-create body for handlers
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			r.ContentLength = int64(len(bodyBytes))

			next.ServeHTTP(w, r)
		})
	}
}

// getFieldFromError extracts field name from validation error
func getFieldFromError(err error) string {
	if valErr, ok := err.(*validation.ValidationError); ok {
		return valErr.Field
	}
	return ""
}
