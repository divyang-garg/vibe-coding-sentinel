// Package models contains shared types and errors
// Complies with CODING_STANDARDS.md: Type definitions max 200 lines
package models

import (
	"fmt"
	"time"
)

// ValidationError represents validation errors
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Resource string      `json:"resource"`
	ID       interface{} `json:"id,omitempty"`
	Message  string      `json:"message"`
}

// Error returns the error message
func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ID != nil {
		return fmt.Sprintf("%s with id %v not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// AuthenticationError represents authentication failures
type AuthenticationError struct {
	Message string `json:"message"`
}

// Error returns the error message
func (e *AuthenticationError) Error() string {
	return e.Message
}

// AuthorizationError represents authorization failures
type AuthorizationError struct {
	Message  string `json:"message"`
	Resource string `json:"resource,omitempty"`
	Action   string `json:"action,omitempty"`
}

// Error returns the error message
func (e *AuthorizationError) Error() string {
	return e.Message
}

// JSONMap represents a flexible JSON object
type JSONMap map[string]interface{}

// JSONArray represents a flexible JSON array
type JSONArray []interface{}

// ErrorSeverity represents error severity levels
type ErrorSeverity int

const (
	ErrorSeverityInfo ErrorSeverity = iota
	ErrorSeverityLow
	ErrorSeverityMedium
	ErrorSeverityHigh
	ErrorSeverityCritical
)

// ErrorClassification represents comprehensive error classification data
type ErrorClassification struct {
	Severity      ErrorSeverity          `json:"severity"`
	Category      string                 `json:"category"`
	Recovery      string                 `json:"recovery_strategy"`
	Retryable     bool                   `json:"retryable"`
	UserVisible   bool                   `json:"user_visible"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Suggestions   []string               `json:"suggestions,omitempty"`
	RelatedErrors []string               `json:"related_errors,omitempty"`
	ErrorCode     int                    `json:"error_code"`
	Timestamp     time.Time              `json:"timestamp"`
	RequestID     string                 `json:"request_id,omitempty"`
	ToolName      string                 `json:"tool_name,omitempty"`
}

// NotImplementedError represents features that are not yet implemented
type NotImplementedError struct {
	Feature  string `json:"feature"`
	Message  string `json:"message"`
	Resource string `json:"resource,omitempty"`
}

// Error returns the error message
func (e *NotImplementedError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "feature not implemented: " + e.Feature
}

// RateLimitError represents rate limiting errors
type RateLimitError struct {
	Message    string    `json:"message"`
	RetryAfter int       `json:"retry_after"`
	ResetTime  time.Time `json:"reset_time"`
}

// Error returns the error message
func (e *RateLimitError) Error() string {
	return e.Message
}
