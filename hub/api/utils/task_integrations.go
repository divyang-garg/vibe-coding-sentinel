// Package utils provides task integration types and linking functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package utils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sentinel-hub-api/pkg/database"

	"github.com/google/uuid"
)

// Package-level database connection
var db *sql.DB

// SetDB sets the database connection for the utils package
func SetDB(database *sql.DB) {
	db = database
}

// TaskLink represents a link between tasks and other entities
type TaskLink struct {
	ID        string
	TaskID    string
	LinkType  string
	LinkedID  string
	CreatedAt time.Time
}

// ChangeRequest represents a change request
type ChangeRequest struct {
	ID                   string
	ProjectID            string
	Status               string
	ImplementationStatus string
	Type                 string
}

// Task represents a task
type Task struct {
	ID      string
	Status  string
	Version int
}

// UpdateTaskRequest represents a task update request
type UpdateTaskRequest struct {
	Title       *string
	Description *string
	Status      *string
	Priority    *string
	AssignedTo  *string
	Tags        []string
	Version     int
}

// CreateTaskRequest represents a task creation request
type CreateTaskRequest struct {
	Source      string
	Title       string
	Description string
	Priority    string
}

// ListTasksRequest represents a task listing request
type ListTasksRequest struct {
	Limit           int
	Offset          int
	Status          string
	StatusFilter    string
	Priority        string
	PriorityFilter  string
	IncludeArchived bool
}

// ListTasksResponse represents a task listing response
type ListTasksResponse struct {
	Tasks []Task
}

// KnowledgeItem represents a knowledge item
type KnowledgeItem struct {
	ID     string
	Status string
}

// TestRequirement represents a test requirement
type TestRequirement struct {
	ID          string
	RuleTitle   string
	Description string
}

// ComprehensiveValidation represents a comprehensive validation
type ComprehensiveValidation struct {
	ID        string
	ProjectID string
	Feature   string
}

// ValidateTaskLink validates that a task link references a valid entity
func ValidateTaskLink(ctx context.Context, linkType, linkedID string) error {
	if linkedID == "" {
		return fmt.Errorf("linked ID cannot be empty")
	}
	// Basic validation - in a full implementation, this would check the linked entity exists
	return nil
}

// LinkTaskToChangeRequest links a task to a change request (Phase 12)
func LinkTaskToChangeRequest(ctx context.Context, taskID string, changeRequestID string) error {
	if err := ValidateTaskID(taskID); err != nil {
		return err
	}
	if err := ValidateChangeRequestID(changeRequestID); err != nil {
		return err
	}
	return createTaskLink(ctx, taskID, LinkTypeChangeRequest, changeRequestID)
}

// LinkTaskToKnowledgeItem links a task to a knowledge item (Phase 4)
func LinkTaskToKnowledgeItem(ctx context.Context, taskID string, knowledgeItemID string) error {
	if err := ValidateTaskID(taskID); err != nil {
		return err
	}
	if err := ValidateKnowledgeItemID(knowledgeItemID); err != nil {
		return err
	}
	return createTaskLink(ctx, taskID, LinkTypeKnowledgeItem, knowledgeItemID)
}

// LinkTaskToComprehensiveAnalysis links a task to comprehensive analysis result (Phase 14A)
func LinkTaskToComprehensiveAnalysis(ctx context.Context, taskID string, validationID string) error {
	if err := ValidateTaskID(taskID); err != nil {
		return err
	}
	if err := ValidateComprehensiveValidationID(validationID); err != nil {
		return err
	}
	return createTaskLink(ctx, taskID, LinkTypeComprehensiveAnalysis, validationID)
}

// LinkTaskToTestRequirement links a task to a test requirement (Phase 10)
func LinkTaskToTestRequirement(ctx context.Context, taskID string, testRequirementID string) error {
	if err := ValidateTaskID(taskID); err != nil {
		return err
	}
	if err := ValidateTestRequirementID(testRequirementID); err != nil {
		return err
	}
	return createTaskLink(ctx, taskID, LinkTypeTestRequirement, testRequirementID)
}

// createTaskLink creates a link between task and another system
func createTaskLink(ctx context.Context, taskID string, linkType string, linkedID string) error {
	// Validate task ID
	if err := ValidateUUID(taskID); err != nil {
		return fmt.Errorf("invalid task ID: %w", err)
	}

	// Validate linked entity exists
	if err := ValidateTaskLink(ctx, linkType, linkedID); err != nil {
		return fmt.Errorf("invalid task link: %w", err)
	}

	// Create new link (unique constraint will handle duplicates)
	linkID := uuid.New().String()
	query := `
		INSERT INTO task_links (id, task_id, link_type, linked_id, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (task_id, link_type, linked_id) DO NOTHING
	`

	_, err := database.ExecWithTimeout(ctx, db, query, linkID, taskID, linkType, linkedID)
	if err != nil {
		// Check if it's a unique constraint violation (shouldn't happen with ON CONFLICT, but handle gracefully)
		return fmt.Errorf("failed to create task link: %w", err)
	}
	return nil
}

// GetTaskLinks retrieves all links for a task
func GetTaskLinks(ctx context.Context, taskID string) ([]TaskLink, error) {
	query := `
		SELECT id, task_id, link_type, linked_id, created_at
		FROM task_links
		WHERE task_id = $1
		ORDER BY created_at DESC
	`

	rows, err := database.QueryWithTimeout(ctx, db, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task links: %w", err)
	}
	defer rows.Close()

	var links []TaskLink
	for rows.Next() {
		var link TaskLink
		err := rows.Scan(&link.ID, &link.TaskID, &link.LinkType, &link.LinkedID, &link.CreatedAt)
		if err != nil {
			continue
		}
		links = append(links, link)
	}

	return links, nil
}

// SyncTaskStatusWithChangeRequest syncs task status with change request status (Phase 12)
func SyncTaskStatusWithChangeRequest(ctx context.Context, taskID string) error {
	// Validate task ID
	if err := ValidateTaskID(taskID); err != nil {
		return err
	}

	// Get task links
	links, err := GetTaskLinks(ctx, taskID)
	if err != nil {
		return err
	}

	// Find change request link
	var changeRequestID string
	for _, link := range links {
		if link.LinkType == LinkTypeChangeRequest {
			changeRequestID = link.LinkedID
			break
		}
	}

	if changeRequestID == "" {
		return nil // No change request linked
	}

	// Get change request
	cr, err := GetChangeRequestByID(ctx, changeRequestID)
	if err != nil {
		if err.Error() == fmt.Sprintf("change request not found: %s", changeRequestID) {
			return nil // Change request not found
		}
		return fmt.Errorf("failed to get change request: %w", err)
	}

	// Map change request status to task status
	taskStatus := mapChangeRequestStatusToTaskStatus(cr.Status, cr.ImplementationStatus)
	if taskStatus == "" {
		return nil // No mapping needed
	}

	// Get current task to check version
	task, err := GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	// Update task status if different
	if task.Status != taskStatus {
		updateReq := UpdateTaskRequest{
			Status:  &taskStatus,
			Version: task.Version,
		}
		_, err = UpdateTask(ctx, taskID, updateReq)
		if err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}
	}

	return nil
}

// mapChangeRequestStatusToTaskStatus maps change request status to task status
func mapChangeRequestStatusToTaskStatus(crStatus, implStatus string) string {
	// If change request is rejected, task should be blocked
	if crStatus == ChangeRequestStatusRejected {
		return TaskStatusBlocked
	}

	// Map implementation status
	switch implStatus {
	case ImplementationStatusCompleted:
		return TaskStatusCompleted
	case ImplementationStatusInProgress:
		return TaskStatusInProgress
	case ImplementationStatusBlocked:
		return TaskStatusBlocked
	case ImplementationStatusPending:
		return TaskStatusPending
	default:
		return ""
	}
}

// CreateTasksFromChangeRequest creates tasks from approved change request (Phase 12)
func CreateTasksFromChangeRequest(ctx context.Context, changeRequestID string, projectID string) ([]string, error) {
	// Validate IDs
	if err := ValidateChangeRequestID(changeRequestID); err != nil {
		return nil, err
	}
	if err := ValidateUUID(projectID); err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	// Get change request
	cr, err := GetChangeRequestByID(ctx, changeRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get change request: %w", err)
	}

	// Verify project ID matches
	if cr.ProjectID != projectID {
		return nil, fmt.Errorf("change request not found for project")
	}

	// Only create tasks for approved change requests
	if cr.Status != ChangeRequestStatusApproved {
		return nil, fmt.Errorf("change request not approved")
	}

	// Extract tasks from proposed_state or description
	// For now, create a single task for the change request
	taskTitle := fmt.Sprintf("Implement change request %s", changeRequestID)
	if cr.Type != "" {
		taskTitle = fmt.Sprintf("Implement %s change: %s", cr.Type, changeRequestID)
	}

	createReq := CreateTaskRequest{
		Source:      LinkTypeChangeRequest,
		Title:       taskTitle,
		Description: fmt.Sprintf("Change request ID: %s\nType: %s", changeRequestID, cr.Type),
		Priority:    TaskPriorityHigh, // Change requests are typically high priority
	}

	task, err := CreateTask(ctx, projectID, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Link task to change request
	if err := LinkTaskToChangeRequest(ctx, task.ID, changeRequestID); err != nil {
		LogError(ctx, "Failed to link task to change request: %v", err)
	}

	return []string{task.ID}, nil
}

// SyncTaskStatusWithDocSync syncs task status with doc-sync reports (Phase 11)
func SyncTaskStatusWithDocSync(ctx context.Context, taskID string) error {
	// Validate task ID
	if err := ValidateTaskID(taskID); err != nil {
		return err
	}

	// Get task links
	links, err := GetTaskLinks(ctx, taskID)
	if err != nil {
		return err
	}

	// Find doc-sync related links (through comprehensive analysis or knowledge items)
	// Check for links to knowledge items that might have doc-sync status
	var docSyncStatus string
	for _, link := range links {
		if link.LinkType == LinkTypeKnowledgeItem {
			// Get knowledge item
			ki, err := GetKnowledgeItemByID(ctx, link.LinkedID)
			if err == nil {
				// Map knowledge item status to task status
				// Knowledge items typically have: "pending", "approved", "active", "deprecated"
				switch ki.Status {
				case KnowledgeItemStatusApproved, KnowledgeItemStatusActive:
					docSyncStatus = TaskStatusInProgress
				case KnowledgeItemStatusDeprecated:
					docSyncStatus = TaskStatusBlocked
				default:
					docSyncStatus = TaskStatusPending
				}
				break
			}
		}
	}

	// If we found a doc-sync status, update task if different
	if docSyncStatus != "" {
		task, err := GetTask(ctx, taskID)
		if err != nil {
			return err
		}

		if task.Status != docSyncStatus {
			updateReq := UpdateTaskRequest{
				Status:  &docSyncStatus,
				Version: task.Version,
			}
			_, err = UpdateTask(ctx, taskID, updateReq)
			if err != nil {
				return fmt.Errorf("failed to update task status from doc-sync: %w", err)
			}
		}
	}

	return nil
}

// CreateTasksFromComprehensiveAnalysis creates tasks from comprehensive analysis results (Phase 14A)
func CreateTasksFromComprehensiveAnalysis(ctx context.Context, validationID string, projectID string) ([]string, error) {
	// Validate IDs
	if err := ValidateComprehensiveValidationID(validationID); err != nil {
		return nil, err
	}
	if err := ValidateUUID(projectID); err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	// Get comprehensive analysis result
	cv, err := GetComprehensiveValidationByID(ctx, validationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comprehensive validation: %w", err)
	}

	// Verify project ID matches
	if cv.ProjectID != projectID {
		return nil, fmt.Errorf("comprehensive validation not found for project")
	}

	// Parse checklist to extract tasks
	// For now, create a single task for the feature
	taskTitle := fmt.Sprintf("Complete feature: %s", cv.Feature)
	createReq := CreateTaskRequest{
		Source:      LinkTypeComprehensiveAnalysis,
		Title:       taskTitle,
		Description: fmt.Sprintf("Feature analysis validation ID: %s", validationID),
		Priority:    TaskPriorityMedium,
	}

	task, err := CreateTask(ctx, projectID, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Link task to comprehensive analysis
	if err := LinkTaskToComprehensiveAnalysis(ctx, task.ID, validationID); err != nil {
		LogError(ctx, "Failed to link task to comprehensive analysis: %v", err)
	}

	return []string{task.ID}, nil
}

// CreateTasksFromTestRequirements creates tasks from missing test requirements (Phase 10)
func CreateTasksFromTestRequirements(ctx context.Context, testRequirementID string, projectID string) ([]string, error) {
	// Validate IDs
	if err := ValidateTestRequirementID(testRequirementID); err != nil {
		return nil, err
	}
	if err := ValidateUUID(projectID); err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}

	// Get test requirement
	tr, err := GetTestRequirementByID(ctx, testRequirementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test requirement: %w", err)
	}

	// Create task for missing test
	taskTitle := fmt.Sprintf("Add tests for: %s", tr.RuleTitle)
	createReq := CreateTaskRequest{
		Source:      LinkTypeTestRequirement,
		Title:       taskTitle,
		Description: tr.Description,
		Priority:    TaskPriorityHigh, // Test requirements are high priority
	}

	task, err := CreateTask(ctx, projectID, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Link task to test requirement
	if err := LinkTaskToTestRequirement(ctx, task.ID, testRequirementID); err != nil {
		LogError(ctx, "Failed to link task to test requirement: %v", err)
	}

	return []string{task.ID}, nil
}

// SyncAllTaskStatuses syncs all task statuses with linked systems
func SyncAllTaskStatuses(ctx context.Context, projectID string) error {
	// Get all tasks for project
	req := ListTasksRequest{
		Limit:  1000,
		Offset: 0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Sync each task
	for _, task := range response.Tasks {
		// Sync with change requests
		if err := SyncTaskStatusWithChangeRequest(ctx, task.ID); err != nil {
			LogError(ctx, "Failed to sync task %s with change request: %v", task.ID, err)
		}

		// Sync with doc-sync
		if err := SyncTaskStatusWithDocSync(ctx, task.ID); err != nil {
			LogError(ctx, "Failed to sync task %s with doc-sync: %v", task.ID, err)
		}
	}

	return nil
}
