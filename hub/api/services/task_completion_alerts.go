// Task Completion - Alert Functions
// Sends alerts for task completion events
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
)

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
