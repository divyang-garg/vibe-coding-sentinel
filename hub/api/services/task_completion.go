// Task Completion - Main Completion Functions
// Automatically completes tasks based on verification confidence and checks for blocking issues
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/pkg/database"
)

// AutoCompletionConfig represents configuration for auto-completion
type AutoCompletionConfig struct {
	HighConfidenceThreshold         float64 // Default: 0.8
	MediumConfidenceThreshold       float64 // Default: 0.5
	RequireHumanApprovalForCritical bool    // Default: true
}

// DefaultAutoCompletionConfig returns default configuration
func DefaultAutoCompletionConfig() AutoCompletionConfig {
	return AutoCompletionConfig{
		HighConfidenceThreshold:         0.8,
		MediumConfidenceThreshold:       0.5,
		RequireHumanApprovalForCritical: true,
	}
}

// AutoCompleteTasks automatically completes tasks based on verification confidence
func AutoCompleteTasks(ctx context.Context, projectID string, config AutoCompletionConfig) ([]string, error) {
	// Get all pending and in_progress tasks
	req := ListTasksRequest{
		StatusFilter: "", // Get all non-completed tasks
		Limit:        1000,
		Offset:       0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	var completedTaskIDs []string

	for _, task := range response.Tasks {
		if task.Status == "completed" {
			continue
		}

		// Skip if no verification confidence
		if task.VerificationConfidence == 0.0 {
			continue
		}

		// Check if task should be auto-completed
		shouldComplete := false
		newStatus := task.Status

		if task.VerificationConfidence >= config.HighConfidenceThreshold {
			// High confidence - auto-complete (unless critical and requires approval)
			if task.Priority == "critical" && config.RequireHumanApprovalForCritical {
				// Mark as in_progress and alert
				newStatus = "in_progress"
				LogInfo(ctx, "Task %s has high confidence (%.2f) but requires human approval (critical)", task.ID, task.VerificationConfidence)
			} else {
				shouldComplete = true
				newStatus = "completed"
			}
		} else if task.VerificationConfidence >= config.MediumConfidenceThreshold {
			// Medium confidence - mark as in_progress
			newStatus = "in_progress"
		}

		// Update task status if changed
		if newStatus != task.Status {
			statusStr := string(newStatus)
			updateReq := UpdateTaskRequest{
				Status:  &statusStr,
				Version: task.Version,
			}

			updatedTask, err := UpdateTask(ctx, task.ID, updateReq)
			if err != nil {
				LogError(ctx, "Failed to update task %s: %v", task.ID, err)
				continue
			}

			if shouldComplete {
				completedTaskIDs = append(completedTaskIDs, task.ID)
				LogInfo(ctx, "Task %s auto-completed (confidence: %.2f)", task.ID, task.VerificationConfidence)
			} else {
				LogInfo(ctx, "Task %s status updated to %s (confidence: %.2f)", task.ID, newStatus, task.VerificationConfidence)
			}

			// Send alert if needed
			if shouldComplete {
				sendTaskCompletionAlert(ctx, updatedTask)
			}
		}
	}

	return completedTaskIDs, nil
}

// CheckIncompleteCriticalTasks checks for incomplete critical tasks and sends alerts
func CheckIncompleteCriticalTasks(ctx context.Context, projectID string) ([]Task, error) {
	req := ListTasksRequest{
		StatusFilter:   "", // Get all non-completed
		PriorityFilter: "critical",
		Limit:          1000,
		Offset:         0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	var incompleteCriticalTasks []Task

	for _, task := range response.Tasks {
		if task.Status != "completed" {
			incompleteCriticalTasks = append(incompleteCriticalTasks, task)

			// Send alert
			sendIncompleteCriticalTaskAlert(ctx, task)
		}
	}

	return incompleteCriticalTasks, nil
}

// CheckDependencyBlocking checks for tasks blocked by dependencies
func CheckDependencyBlocking(ctx context.Context, projectID string) ([]Task, error) {
	req := ListTasksRequest{
		StatusFilter: "", // Get all tasks
		Limit:        1000,
		Offset:       0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	var blockedTasks []Task

	for _, task := range response.Tasks {
		if task.Status == "completed" {
			continue
		}

		// Get dependencies
		deps, err := GetTaskDependencies(ctx, task.ID)
		if err != nil {
			continue
		}

		// Check if any dependency is incomplete
		isBlocked := false
		// Extract blocked by list from dependency graph
		blockedByList := []string{}
		if deps.Graph != nil && len(deps.Graph.Dependencies) > 0 {
			for _, dep := range deps.Graph.Dependencies {
				if dep.TaskID == task.ID {
					blockedByList = append(blockedByList, dep.DependsOnTaskID)
				}
			}
		}
		for _, depID := range blockedByList {
			depTask, err := GetTask(ctx, depID)
			if err != nil {
				continue
			}

			if depTask.Status != "completed" {
				isBlocked = true
				break
			}
		}

		if isBlocked {
			// Update task status to blocked if not already
			if task.Status != "blocked" {
				blockedStatus := "blocked"
				updateReq := UpdateTaskRequest{
					Status:  &blockedStatus,
					Version: task.Version,
				}
				UpdateTask(ctx, task.ID, updateReq)
			}

			blockedTasks = append(blockedTasks, task)
			sendDependencyBlockingAlert(ctx, task, blockedByList)
		}
	}

	return blockedTasks, nil
}

// ArchiveTask archives a completed task (soft delete)
func ArchiveTask(ctx context.Context, taskID string) error {
	task, err := GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status != "completed" {
		return fmt.Errorf("can only archive completed tasks")
	}

	// Set archived_at timestamp
	now := time.Now()
	query := `
		UPDATE tasks 
		SET archived_at = $1, updated_at = $2
		WHERE id = $3 AND archived_at IS NULL
	`

	// Invalidate cache
	InvalidateTaskCache(taskID)

	_, err = database.ExecWithTimeout(ctx, db, query, now, now, taskID)
	if err != nil {
		return fmt.Errorf("failed to archive task: %w", err)
	}

	return nil
}
