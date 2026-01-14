// Package models - Task data models
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import (
	"time"
)

// Task represents a task in the system
type Task struct {
	ID                     string       `json:"id" db:"id" validate:"required,uuid"`
	ProjectID              string       `json:"project_id" db:"project_id" validate:"required,uuid"`
	Source                 string       `json:"source" db:"source" validate:"required,oneof=cursor manual change_request comprehensive_analysis"`
	Title                  string       `json:"title" db:"title" validate:"required,min=1,max=500"`
	Description            string       `json:"description,omitempty" db:"description" validate:"max=5000"`
	FilePath               string       `json:"file_path,omitempty" db:"file_path"`
	LineNumber             *int         `json:"line_number,omitempty" db:"line_number"`
	Status                 TaskStatus   `json:"status" db:"status" validate:"required"`
	Priority               TaskPriority `json:"priority" db:"priority" validate:"required"`
	AssignedTo             *string      `json:"assigned_to,omitempty" db:"assigned_to" validate:"omitempty,email"`
	EstimatedEffort        *int         `json:"estimated_effort,omitempty" db:"estimated_effort" validate:"omitempty,min=0"`
	ActualEffort           *int         `json:"actual_effort,omitempty" db:"actual_effort" validate:"omitempty,min=0"`
	Tags                   []string     `json:"tags,omitempty" db:"tags"`
	VerificationConfidence float64      `json:"verification_confidence" db:"verification_confidence" validate:"min=0,max=1"`
	CreatedAt              time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time    `json:"updated_at" db:"updated_at"`
	CompletedAt            *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
	VerifiedAt             *time.Time   `json:"verified_at,omitempty" db:"verified_at"`
	ArchivedAt             *time.Time   `json:"archived_at,omitempty" db:"archived_at"`
	Version                int          `json:"version" db:"version" validate:"min=0"`
}

// TaskDependencyGraph represents a graph of task dependencies
type TaskDependencyGraph struct {
	Tasks          map[string]*Task `json:"tasks"`
	Dependencies   []TaskDependency `json:"dependencies"`
	ExecutionOrder []string         `json:"execution_order"`
	Errors         []string         `json:"errors,omitempty"`
	Cycles         [][]string       `json:"cycles,omitempty"`
	IsValid        bool             `json:"is_valid"`
	GeneratedAt    string           `json:"generated_at"`
}

// TaskExecutionPlan represents a plan for executing tasks
type TaskExecutionPlan struct {
	Dependencies      []TaskDependency `json:"dependencies"`
	Batches           [][]string       `json:"batches"`
	EstimatedDuration time.Duration    `json:"estimated_duration"`
	Tasks             []Task           `json:"tasks"`
	RiskFactors       []string         `json:"risk_factors"`
	CreatedAt         time.Time        `json:"created_at"`
}

// TaskChange represents a change to a task
type TaskChange struct {
	ID            string                 `json:"id"`
	TaskID        string                 `json:"task_id"`
	ChangeType    string                 `json:"change_type"`
	OldValues     map[string]interface{} `json:"old_values,omitempty"`
	NewValues     map[string]interface{} `json:"new_values,omitempty"`
	Justification string                 `json:"justification"`
	Field         string                 `json:"field"`
	OldValue      interface{}            `json:"old_value,omitempty"`
	NewValue      interface{}            `json:"new_value"`
	Description   string                 `json:"description"`
	ChangedBy     string                 `json:"changed_by,omitempty"`
	ChangedAt     time.Time              `json:"changed_at"`
}

// TaskDependency represents a dependency between tasks
type TaskDependency struct {
	ID              string    `json:"id" db:"id" validate:"required,uuid"`
	TaskID          string    `json:"task_id" db:"task_id" validate:"required,uuid"`
	DependsOnTaskID string    `json:"depends_on_task_id" db:"depends_on_task_id" validate:"required,uuid"`
	DependencyType  string    `json:"dependency_type" db:"dependency_type" validate:"required,oneof=finish_to_start start_to_start finish_to_finish start_to_finish"`
	Confidence      float64   `json:"confidence" db:"confidence" validate:"min=0,max=1"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// TaskLink represents a link between tasks
type TaskLink struct {
	ID           string    `json:"id" db:"id" validate:"required,uuid"`
	TaskID       string    `json:"task_id" db:"task_id" validate:"required,uuid"`
	LinkedTaskID string    `json:"linked_task_id" db:"linked_task_id" validate:"required,uuid"`
	LinkType     string    `json:"link_type" db:"link_type" validate:"required,oneof=blocks blocked_by relates_to duplicates"`
	Description  string    `json:"description,omitempty" db:"description" validate:"max=500"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// TaskVerification represents a verification of task completion
type TaskVerification struct {
	ID               string                 `json:"id" db:"id" validate:"required,uuid"`
	TaskID           string                 `json:"task_id" db:"task_id" validate:"required,uuid"`
	Status           VerificationStatus     `json:"status" db:"status" validate:"required"`
	VerificationType string                 `json:"verification_type" db:"verification_type" validate:"required"`
	RetryCount       int                    `json:"retry_count" db:"retry_count" validate:"min=0"`
	Confidence       float64                `json:"confidence" db:"confidence" validate:"min=0,max=1"`
	VerifiedBy       string                 `json:"verified_by,omitempty" db:"verified_by" validate:"omitempty,email"`
	VerifiedAt       *time.Time             `json:"verified_at,omitempty" db:"verified_at"`
	Notes            string                 `json:"notes,omitempty" db:"notes" validate:"max=1000"`
	Evidence         map[string]interface{} `json:"evidence,omitempty" db:"evidence"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}
