// Package models contains workflow-related data models.
// This file defines all workflow orchestration domain entities following the data-only principle.
package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// WorkflowStatus represents workflow execution status
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// String returns the string representation of WorkflowStatus
func (s WorkflowStatus) String() string {
	return string(s)
}

// IsValid checks if the WorkflowStatus is valid
func (s WorkflowStatus) IsValid() bool {
	switch s {
	case WorkflowStatusPending, WorkflowStatusRunning, WorkflowStatusCompleted, WorkflowStatusFailed, WorkflowStatusCancelled:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler
func (s WorkflowStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements json.Unmarshaler
func (s *WorkflowStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = WorkflowStatus(str)
	if !s.IsValid() {
		return fmt.Errorf("invalid workflow status: %s", str)
	}
	return nil
}

// StepStatus represents the status of a workflow step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// String returns the string representation of StepStatus
func (s StepStatus) String() string {
	return string(s)
}

// IsValid checks if the StepStatus is valid
func (s StepStatus) IsValid() bool {
	switch s {
	case StepStatusPending, StepStatusRunning, StepStatusCompleted, StepStatusFailed, StepStatusSkipped:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler
func (s StepStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements json.Unmarshaler
func (s *StepStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = StepStatus(str)
	if !s.IsValid() {
		return fmt.Errorf("invalid step status: %s", str)
	}
	return nil
}

// WorkflowDefinition represents a workflow definition
type WorkflowDefinition struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Steps        []WorkflowStep         `json:"steps"`
	InputSchema  map[string]interface{} `json:"input_schema"`
	OutputSchema map[string]interface{} `json:"output_schema"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	ToolName   string                 `json:"tool_name"`
	Arguments  map[string]interface{} `json:"arguments"`
	DependsOn  []string               `json:"depends_on,omitempty"`
	Condition  string                 `json:"condition,omitempty"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
	RetryCount int                    `json:"retry_count,omitempty"`
}

// WorkflowExecution represents workflow execution state
type WorkflowExecution struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      WorkflowStatus         `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       *WorkflowError         `json:"error,omitempty"`
	RequestID   string                 `json:"request_id"`
	Progress    int                    `json:"progress,omitempty"` // 0-100
	Steps       []StepResult           `json:"steps,omitempty"`
}

// GetProgress returns the execution progress (0-100)
func (e *WorkflowExecution) GetProgress() int {
	if e.Progress > 0 {
		return e.Progress
	}
	// Calculate progress based on completed steps
	if len(e.Steps) == 0 {
		return 0
	}
	completed := 0
	for _, step := range e.Steps {
		if step.Status == "completed" {
			completed++
		}
	}
	return (completed * 100) / len(e.Steps)
}

// WorkflowError represents a workflow execution error
type WorkflowError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	StepID  string `json:"step_id,omitempty"`
}

// StepAttempt represents an attempt to execute a workflow step
type StepAttempt struct {
	AttemptNumber int                    `json:"attempt_number"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Success       bool                   `json:"success"`
	Duration      time.Duration          `json:"duration"`
	Output        map[string]interface{} `json:"output,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

// StepResult represents the result of executing a workflow step
type StepResult struct {
	StepID      string                 `json:"step_id"`
	Status      StepStatus             `json:"status"`
	Attempts    []StepAttempt          `json:"attempts,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	RetryCount  int                    `json:"retry_count"`
}
