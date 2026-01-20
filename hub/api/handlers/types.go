// Package handlers types
// Type definitions for HTTP handlers
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package handlers

import "time"

// Project represents a project (local definition for handlers)
type Project struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

// Task represents a task (local definition for handlers)
type Task struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	FilePath    string    `json:"file_path"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateTaskRequest represents a request to create a task
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Source      string `json:"source"`
	Priority    string `json:"priority,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Priority    *string `json:"priority,omitempty"`
}

// ListTasksRequest represents a request to list tasks
type ListTasksRequest struct {
	Limit           int      `json:"limit"`
	Offset          int      `json:"offset"`
	Status          string   `json:"status,omitempty"`
	StatusFilter    string   `json:"status_filter,omitempty"`
	PriorityFilter  string   `json:"priority_filter,omitempty"`
	SourceFilter    string   `json:"source_filter,omitempty"`
	AssignedTo      *string  `json:"assigned_to,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	IncludeArchived *bool    `json:"include_archived,omitempty"`
}

// ListTasksResponse represents a response with tasks
type ListTasksResponse struct {
	Tasks   []Task `json:"tasks"`
	Total   int    `json:"total"`
	HasMore bool   `json:"has_more"`
}

// VerifyTaskRequest represents a request to verify a task
type VerifyTaskRequest struct {
	CodebasePath string `json:"codebase_path"`
	Force        bool   `json:"force,omitempty"`
}

// VerifyTaskResponse represents a task verification response
type VerifyTaskResponse struct {
	Verified bool   `json:"verified"`
	Message  string `json:"message"`
}

// HandlerConfig represents handler configuration
type HandlerConfig struct {
	Limits LimitsConfig `json:"limits"`
}

// LimitsConfig represents limits configuration
type LimitsConfig struct {
	MaxTaskTitleLength       int `json:"max_task_title_length"`
	MaxTaskDescriptionLength int `json:"max_task_description_length"`
	MaxTaskListLimit         int `json:"max_task_list_limit"`
	DefaultTaskListLimit     int `json:"default_task_list_limit"`
}

var handlerConfig = &HandlerConfig{
	Limits: LimitsConfig{
		MaxTaskTitleLength:       200,
		MaxTaskDescriptionLength: 4000,
		MaxTaskListLimit:         100,
		DefaultTaskListLimit:     20,
	},
}

// GetConfig returns handler configuration
func GetConfig() *HandlerConfig {
	return handlerConfig
}

// TaskDetector represents a task detector (stub interface)
type TaskDetector struct{}

// DetectedTask represents a detected task
type DetectedTask struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Source      string `json:"source"`
}
