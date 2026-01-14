// Package models - Monitoring and error handling data models
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import (
	"time"
)

// ErrorReport represents an error report for monitoring
type ErrorReport struct {
	ID         string                 `json:"id"`
	Category   string                 `json:"category" validate:"required"`
	Message    string                 `json:"message" validate:"required"`
	Severity   ErrorSeverity          `json:"severity"`
	Context    map[string]interface{} `json:"context,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	Resolution string                 `json:"resolution,omitempty"`
}
