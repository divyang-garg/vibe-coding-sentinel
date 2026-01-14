// Package models contains all data models and types for the Sentinel Hub API.
// This package follows the data-only principle - no business logic or external dependencies.
//
// Architecture: Pure data structures with JSON tags and validation tags where appropriate.
package models

import (
	"fmt"
	"time"
)

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

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Resource string `json:"resource"`
	Message  string `json:"message"`
}

// Error returns the error message
func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

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

// DependencyChange represents a change in task dependencies
type DependencyChange struct {
	ID                string                 `json:"id"`
	TaskID            string                 `json:"task_id"`
	DependencyID      string                 `json:"dependency_id,omitempty"`
	ChangeType        string                 `json:"change_type"`
	AffectedChain     []string               `json:"affected_chain"`
	RippleEffects     map[string]interface{} `json:"ripple_effects"`
	CycleRisk         bool                   `json:"cycle_risk"`
	PerformanceImpact string                 `json:"performance_impact"`
	ChangedAt         time.Time              `json:"changed_at"`
}

// ChangeImpactReport represents the impact analysis of a change
type ChangeImpactReport struct {
	TaskID             string              `json:"task_id"`
	ChangeDescription  string              `json:"change_description"`
	ImpactAnalysis     ErrorClassification `json:"impact_analysis"`
	DependencyChanges  []DependencyChange  `json:"dependency_changes"`
	RecommendedActions []string            `json:"recommended_actions"`
	RiskMitigationPlan []string            `json:"risk_mitigation_plan"`
	TimelineImpact     int                 `json:"timeline_impact_days"`
	ConfidenceLevel    float64             `json:"confidence_level"`
	GeneratedAt        time.Time           `json:"generated_at"`
}

// MCPError represents an MCP protocol error
type MCPError struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    JSONMap `json:"data,omitempty"`
}

// Error returns the error message
func (e *MCPError) Error() string {
	return e.Message
}
