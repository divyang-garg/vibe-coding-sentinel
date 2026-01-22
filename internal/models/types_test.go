// Package models provides additional tests for error types
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error_Additional(t *testing.T) {
	t.Run("formats error message with field", func(t *testing.T) {
		err := &ValidationError{
			Field:   "email",
			Message: "invalid format",
		}

		msg := err.Error()
		assert.Contains(t, msg, "email")
		assert.Contains(t, msg, "invalid format")
	})

	t.Run("formats error message with field and value", func(t *testing.T) {
		err := &ValidationError{
			Field:   "age",
			Value:   150,
			Message: "must be between 0 and 120",
		}

		msg := err.Error()
		assert.Contains(t, msg, "age")
		assert.Contains(t, msg, "must be between 0 and 120")
	})
}

func TestNotFoundError_Error_Additional(t *testing.T) {
	t.Run("uses custom message when provided", func(t *testing.T) {
		err := &NotFoundError{
			Resource: "user",
			ID:       "123",
			Message:  "Custom not found message",
		}

		msg := err.Error()
		assert.Equal(t, "Custom not found message", msg)
	})

	t.Run("formats message with resource and ID", func(t *testing.T) {
		err := &NotFoundError{
			Resource: "task",
			ID:       42,
		}

		msg := err.Error()
		assert.Contains(t, msg, "task")
		assert.Contains(t, msg, "42")
	})

	t.Run("formats message with resource only", func(t *testing.T) {
		err := &NotFoundError{
			Resource: "project",
		}

		msg := err.Error()
		assert.Contains(t, msg, "project")
		assert.Contains(t, msg, "not found")
	})
}

func TestAuthenticationError_Error_Additional(t *testing.T) {
	t.Run("returns message", func(t *testing.T) {
		err := &AuthenticationError{
			Message: "Invalid credentials",
		}

		msg := err.Error()
		assert.Equal(t, "Invalid credentials", msg)
	})

	t.Run("handles empty message", func(t *testing.T) {
		err := &AuthenticationError{}

		msg := err.Error()
		assert.Equal(t, "", msg)
	})
}

func TestAuthorizationError_Error_Additional(t *testing.T) {
	t.Run("returns message", func(t *testing.T) {
		err := &AuthorizationError{
			Message: "Access denied",
		}

		msg := err.Error()
		assert.Equal(t, "Access denied", msg)
	})
}

func TestNotImplementedError_Error_Additional(t *testing.T) {
	t.Run("uses custom message when provided", func(t *testing.T) {
		err := &NotImplementedError{
			Feature: "feature-x",
			Message: "Custom message",
		}

		msg := err.Error()
		assert.Equal(t, "Custom message", msg)
	})

	t.Run("formats default message with feature", func(t *testing.T) {
		err := &NotImplementedError{
			Feature: "advanced-search",
		}

		msg := err.Error()
		assert.Contains(t, msg, "advanced-search")
		assert.Contains(t, msg, "not implemented")
	})
}

func TestRateLimitError_Error_Additional(t *testing.T) {
	t.Run("returns message", func(t *testing.T) {
		err := &RateLimitError{
			Message: "Rate limit exceeded",
		}

		msg := err.Error()
		assert.Equal(t, "Rate limit exceeded", msg)
	})
}

func TestErrorSeverity(t *testing.T) {
	t.Run("defines all severity levels", func(t *testing.T) {
		severities := []ErrorSeverity{
			ErrorSeverityInfo,
			ErrorSeverityLow,
			ErrorSeverityMedium,
			ErrorSeverityHigh,
			ErrorSeverityCritical,
		}

		for _, sev := range severities {
			assert.GreaterOrEqual(t, int(sev), 0)
			assert.LessOrEqual(t, int(sev), 4)
		}
	})
}

func TestErrorClassification(t *testing.T) {
	t.Run("creates error classification", func(t *testing.T) {
		classification := &ErrorClassification{
			Severity:      ErrorSeverityHigh,
			Category:      "validation",
			Recovery:      "retry",
			Retryable:     true,
			UserVisible:   true,
			Context:       map[string]interface{}{"field": "email"},
			Suggestions:   []string{"Check email format"},
			RelatedErrors: []string{"ERR001"},
			ErrorCode:     400,
		}

		assert.Equal(t, ErrorSeverityHigh, classification.Severity)
		assert.Equal(t, "validation", classification.Category)
		assert.True(t, classification.Retryable)
		assert.True(t, classification.UserVisible)
		assert.Equal(t, 400, classification.ErrorCode)
	})
}

func TestJSONMap(t *testing.T) {
	t.Run("creates and accesses JSON map", func(t *testing.T) {
		m := JSONMap{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}

		assert.Equal(t, "value1", m["key1"])
		assert.Equal(t, 42, m["key2"])
		assert.Equal(t, true, m["key3"])
	})
}

func TestJSONArray(t *testing.T) {
	t.Run("creates and accesses JSON array", func(t *testing.T) {
		arr := JSONArray{"item1", 42, true}

		assert.Equal(t, "item1", arr[0])
		assert.Equal(t, 42, arr[1])
		assert.Equal(t, true, arr[2])
	})
}
