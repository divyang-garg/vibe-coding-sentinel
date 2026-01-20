// Package models - Task request/response models
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import "time"

// CreateTaskRequest represents a request to create a new task
type CreateTaskRequest struct {
	ProjectID              string   `json:"project_id" validate:"required,uuid"`
	Title                  string   `json:"title" validate:"required,min=1,max=500"`
	Description            string   `json:"description,omitempty" validate:"max=5000"`
	Source                 string   `json:"source,omitempty" validate:"omitempty,oneof=cursor manual change_request comprehensive_analysis"`
	FilePath               string   `json:"file_path,omitempty"`
	LineNumber             *int     `json:"line_number,omitempty"`
	Priority               string   `json:"priority,omitempty" validate:"omitempty,oneof=low medium high critical"`
	AssignedTo             *string  `json:"assigned_to,omitempty" validate:"omitempty,email"`
	EstimatedEffort        *int     `json:"estimated_effort,omitempty" validate:"omitempty,min=0"`
	Tags                   []string `json:"tags,omitempty"`
	VerificationConfidence float64  `json:"verification_confidence,omitempty" validate:"min=0,max=1"`
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	Title                  *string  `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description            *string  `json:"description,omitempty" validate:"omitempty,max=5000"`
	FilePath               *string  `json:"file_path,omitempty"`
	LineNumber             *int     `json:"line_number,omitempty"`
	Status                 *string  `json:"status,omitempty" validate:"omitempty,oneof=pending in_progress completed blocked cancelled"`
	Priority               *string  `json:"priority,omitempty" validate:"omitempty,oneof=low medium high critical"`
	AssignedTo             *string  `json:"assigned_to,omitempty" validate:"omitempty,email"`
	EstimatedEffort        *int     `json:"estimated_effort,omitempty" validate:"omitempty,min=0"`
	ActualEffort           *int     `json:"actual_effort,omitempty" validate:"omitempty,min=0"`
	Tags                   []string `json:"tags,omitempty"`
	VerificationConfidence *float64 `json:"verification_confidence,omitempty" validate:"omitempty,min=0,max=1"`
	Version                int      `json:"version"` // For optimistic locking
}

// ListTasksRequest represents a request to list tasks with filtering and pagination
type ListTasksRequest struct {
	ProjectID       string `json:"project_id,omitempty"`
	Status          string `json:"status,omitempty"`
	StatusFilter    string `json:"status_filter,omitempty"` // Alias for Status
	Priority        string `json:"priority,omitempty"`
	PriorityFilter  string `json:"priority_filter,omitempty"` // Alias for Priority
	AssignedTo      string `json:"assigned_to,omitempty"`
	Source          string `json:"source,omitempty"`
	SourceFilter    string `json:"source_filter,omitempty"` // Alias for Source
	IncludeArchived bool   `json:"include_archived,omitempty"`
	Limit           int    `json:"limit"`
	Offset          int    `json:"offset"`
}

// ListTasksResponse represents the response from listing tasks
type ListTasksResponse struct {
	Tasks   []Task `json:"tasks"`
	Total   int    `json:"total"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	HasMore bool   `json:"has_more"`
}

// AddDependencyRequest represents a request to add a task dependency
type AddDependencyRequest struct {
	DependsOnTaskID string  `json:"depends_on_task_id" validate:"required,uuid"`
	DependencyType  string  `json:"dependency_type" validate:"required,oneof=finish_to_start start_to_start finish_to_finish start_to_finish"`
	Confidence      float64 `json:"confidence,omitempty" validate:"min=0,max=1"`
}

// VerifyTaskRequest represents a request to verify task completion
type VerifyTaskRequest struct {
	Status     string                 `json:"status" validate:"required,oneof=pending verified failed"`
	Confidence float64                `json:"confidence,omitempty" validate:"min=0,max=1"`
	VerifiedBy string                 `json:"verified_by,omitempty" validate:"omitempty,email"`
	VerifiedAt time.Time              `json:"verified_at"`
	Notes      string                 `json:"notes,omitempty" validate:"max=1000"`
	Evidence   map[string]interface{} `json:"evidence,omitempty"`
}

// VerifyTaskResponse represents the response from task verification
type VerifyTaskResponse struct {
	Task         *Task             `json:"task"`
	Verification *TaskVerification `json:"verification"`
	Success      bool              `json:"success"`
}

// TaskImpact represents the impact analysis of a task change
type TaskImpact struct {
	TaskID            string   `json:"task_id"`
	TaskTitle         string   `json:"task_title"`
	ImpactType        string   `json:"impact_type"`
	Severity          string   `json:"severity"`
	Description       string   `json:"description"`
	Confidence        float64  `json:"confidence"`
	ChangeDescription string   `json:"change_description"`
	AffectedTasks     []string `json:"affected_tasks"`
	DependencyChain   []string `json:"dependency_chain"`
	RiskLevel         string   `json:"risk_level"`
	TimeImpact        int      `json:"time_impact_days"`
	Recommendations   []string `json:"recommendations"`
	GeneratedAt       string   `json:"generated_at"`
}

// TaskImpactAnalysis represents comprehensive impact analysis
type TaskImpactAnalysis struct {
	ID                    string       `json:"id"`
	TaskID                string       `json:"task_id"`
	ChangeType            string       `json:"change_type"`
	ImpactScope           string       `json:"impact_scope"`
	AffectedTasks         []string     `json:"affected_tasks"`
	RiskLevel             string       `json:"risk_level"`
	RiskFactors           []string     `json:"risk_factors"`
	MitigationSuggestions []string     `json:"mitigation_suggestions"`
	EstimatedImpactTime   int          `json:"estimated_impact_time"`
	ConfidenceScore       float64      `json:"confidence_score"`
	AnalyzedAt            string       `json:"analyzed_at"`
	PrimaryImpact         TaskImpact   `json:"primary_impact"`
	CascadeEffects        []TaskImpact `json:"cascade_effects"`
	BlockingTasks         []string     `json:"blocking_tasks"`
	CriticalPathImpact    int          `json:"critical_path_impact_days"`
	OverallRiskLevel      string       `json:"overall_risk_level"`
	MitigationStrategies  []string     `json:"mitigation_strategies"`
	AnalysisConfidence    float64      `json:"analysis_confidence"`
	GeneratedAt           string       `json:"generated_at"`
}

// DependencyGraphResponse represents the response from dependency graph operations
type DependencyGraphResponse struct {
	Graph     *TaskDependencyGraph `json:"graph"`
	IsValid   bool                 `json:"is_valid"`
	Cycles    [][]string           `json:"cycles,omitempty"`
	Generated string               `json:"generated"`
}
