// Fixed import structure
package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
			updateReq := UpdateTaskRequest{
				Status:  &newStatus,
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
		for _, depID := range deps.BlockedBy {
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
				updateReq := UpdateTaskRequest{
					Status:  stringPtr("blocked"),
					Version: task.Version,
				}
				UpdateTask(ctx, task.ID, updateReq)
			}

			blockedTasks = append(blockedTasks, task)
			sendDependencyBlockingAlert(ctx, task, deps.BlockedBy)
		}
	}

	return blockedTasks, nil
}

// ScheduleVerification schedules task verification based on trigger
func ScheduleVerification(ctx context.Context, trigger string, projectID string, codebasePath string) error {
	switch trigger {
	case "on_commit":
		// Verify tasks in changed files
		return verifyTasksOnCommit(ctx, projectID, codebasePath)
	case "on_push":
		// Verify all pending tasks
		return verifyAllPendingTasks(ctx, projectID, codebasePath)
	case "manual":
		// Manual verification - no-op, handled by endpoint
		return nil
	case "scheduled":
		// Scheduled verification - verify all tasks
		return verifyAllTasks(ctx, projectID, codebasePath)
	default:
		return fmt.Errorf("unknown trigger: %s", trigger)
	}
}

// verifyTasksOnCommit verifies tasks related to changed files
func verifyTasksOnCommit(ctx context.Context, projectID string, codebasePath string) error {
	// Get changed files from git
	changedFiles, err := getChangedFilesFromGit(codebasePath)
	if err != nil {
		LogError(ctx, "Failed to get changed files from git: %v", err)
		// Fallback to verifying all pending tasks
		return verifyAllPendingTasks(ctx, projectID, codebasePath)
	}

	if len(changedFiles) == 0 {
		// No changed files, nothing to verify
		return nil
	}

	// Get all pending tasks
	req := ListTasksRequest{
		StatusFilter: "pending",
		Limit:        1000,
		Offset:       0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Verify only tasks related to changed files
	changedFilesMap := make(map[string]bool)
	for _, file := range changedFiles {
		// Normalize paths for comparison
		normalized := filepath.Clean(file)
		changedFilesMap[normalized] = true
		// Also check without leading ./
		if strings.HasPrefix(normalized, "./") {
			changedFilesMap[normalized[2:]] = true
		}
	}

	for _, task := range response.Tasks {
		// Check if task's file path matches any changed file
		shouldVerify := false
		if task.FilePath != "" {
			taskPath := filepath.Clean(task.FilePath)
			if changedFilesMap[taskPath] {
				shouldVerify = true
			} else {
				// Check if any changed file is in the same directory or matches pattern
				for changedFile := range changedFilesMap {
					if strings.Contains(taskPath, changedFile) || strings.Contains(changedFile, taskPath) {
						shouldVerify = true
						break
					}
				}
			}
		} else {
			// If no file path, verify based on keywords in changed files
			keywords := extractKeywords(task.Title + " " + task.Description)
			for _, changedFile := range changedFiles {
				fullPath := filepath.Join(codebasePath, changedFile)
				content, err := os.ReadFile(fullPath)
				if err == nil {
					contentStr := strings.ToLower(string(content))
					for _, keyword := range keywords {
						if strings.Contains(contentStr, strings.ToLower(keyword)) {
							shouldVerify = true
							break
						}
					}
					if shouldVerify {
						break
					}
				}
			}
		}

		if shouldVerify {
			_, err := VerifyTask(ctx, task.ID, codebasePath, false)
			if err != nil {
				LogError(ctx, "Failed to verify task %s: %v", task.ID, err)
				continue
			}
		}
	}

	return nil
}

// getChangedFilesFromGit gets list of changed files from git
func getChangedFilesFromGit(codebasePath string) ([]string, error) {
	// Execute git diff to get changed files
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = codebasePath

	output, err := cmd.Output()
	if err != nil {
		// If git command fails, try git diff --cached for staged files
		cmd = exec.Command("git", "diff", "--cached", "--name-only")
		cmd.Dir = codebasePath
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get changed files from git: %w", err)
		}
	}

	// Parse output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var changedFiles []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			changedFiles = append(changedFiles, line)
		}
	}

	return changedFiles, nil
}

// verifyAllPendingTasks verifies all pending tasks
func verifyAllPendingTasks(ctx context.Context, projectID string, codebasePath string) error {
	req := ListTasksRequest{
		StatusFilter: "pending",
		Limit:        1000,
		Offset:       0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range response.Tasks {
		_, err := VerifyTask(ctx, task.ID, codebasePath, false)
		if err != nil {
			LogError(ctx, "Failed to verify task %s: %v", task.ID, err)
			continue
		}
	}

	return nil
}

// verifyAllTasks verifies all tasks
func verifyAllTasks(ctx context.Context, projectID string, codebasePath string) error {
	req := ListTasksRequest{
		Limit:  1000,
		Offset: 0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, task := range response.Tasks {
		if task.Status == "completed" {
			continue // Skip completed tasks
		}

		_, err := VerifyTask(ctx, task.ID, codebasePath, false)
		if err != nil {
			LogError(ctx, "Failed to verify task %s: %v", task.ID, err)
			continue
		}
	}

	return nil
}

// sendTaskCompletionAlert sends alert when task is auto-completed
func sendTaskCompletionAlert(ctx context.Context, task *Task) {
	if alertService == nil {
		initAlertService()
	}
	if err := alertService.SendTaskCompletionAlert(ctx, task); err != nil {
		LogError(ctx, "Failed to send task completion alert: %v", err)
	}
}

// sendIncompleteCriticalTaskAlert sends alert for incomplete critical task
func sendIncompleteCriticalTaskAlert(ctx context.Context, task Task) {
	if alertService == nil {
		initAlertService()
	}
	if err := alertService.SendCriticalTaskAlert(ctx, task); err != nil {
		LogError(ctx, "Failed to send critical task alert: %v", err)
	}
}

// sendDependencyBlockingAlert sends alert when task is blocked by dependencies
func sendDependencyBlockingAlert(ctx context.Context, task Task, blockingTaskIDs []string) {
	if alertService == nil {
		initAlertService()
	}
	if err := alertService.SendDependencyBlockingAlert(ctx, task, blockingTaskIDs); err != nil {
		LogError(ctx, "Failed to send dependency blocking alert: %v", err)
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
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
