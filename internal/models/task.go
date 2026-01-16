// Package models contains task-related data models
// Complies with CODING_STANDARDS.md: Data models max 200 lines
package models

import (
	"time"
)

// Task represents a tracked task
type Task struct {
	ID                     int          `json:"id" db:"id"`
	ProjectID              string       `json:"project_id" db:"project_id"`
	Source                 string       `json:"source" db:"source"`
	Title                  string       `json:"title" db:"title"`
	Description            string       `json:"description,omitempty" db:"description"`
	FilePath               string       `json:"file_path,omitempty" db:"file_path"`
	LineNumber             *int         `json:"line_number,omitempty" db:"line_number"`
	Status                 TaskStatus   `json:"status" db:"status"`
	Priority               TaskPriority `json:"priority" db:"priority"`
	AssignedTo             *string      `json:"assigned_to,omitempty" db:"assigned_to"`
	EstimatedEffort        *int         `json:"estimated_effort,omitempty" db:"estimated_effort"`
	ActualEffort           *int         `json:"actual_effort,omitempty" db:"actual_effort"`
	Tags                   []string     `json:"tags,omitempty" db:"tags"`
	VerificationConfidence float64      `json:"verification_confidence" db:"verification_confidence"`
	CreatedAt              time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time    `json:"updated_at" db:"updated_at"`
	CompletedAt            *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
	VerifiedAt             *time.Time   `json:"verified_at,omitempty" db:"verified_at"`
	ArchivedAt             *time.Time   `json:"archived_at,omitempty" db:"archived_at"`
	Version                int          `json:"version" db:"version"`
}

// TaskStatus represents task status types
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskPriority represents task priority levels
type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityMedium   TaskPriority = "medium"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

// Validate validates task data
func (t *Task) Validate() error {
	if t.ProjectID == "" {
		return &ValidationError{Field: "project_id", Message: "project_id is required"}
	}
	if t.Title == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	return nil
}

// IsCompleted returns true if the task is completed
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

// CanTransitionTo checks if the task can transition to the given status
func (t *Task) CanTransitionTo(newStatus TaskStatus) bool {
	switch t.Status {
	case TaskStatusPending:
		return newStatus == TaskStatusInProgress || newStatus == TaskStatusCancelled
	case TaskStatusInProgress:
		return newStatus == TaskStatusCompleted || newStatus == TaskStatusBlocked || newStatus == TaskStatusCancelled
	case TaskStatusBlocked:
		return newStatus == TaskStatusInProgress || newStatus == TaskStatusCancelled
	case TaskStatusCompleted, TaskStatusCancelled:
		return false // Terminal states
	default:
		return false
	}
}
