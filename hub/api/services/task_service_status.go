// Package services - Task service status and bulk operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/models"
)

// AssignTask assigns a task to a user
func (s *TaskServiceImpl) AssignTask(ctx context.Context, id string, userID string, version int) (*models.Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Version check for optimistic locking
	if task.Version != version {
		return nil, fmt.Errorf("task version conflict")
	}

	task.AssignedTo = &userID
	task.UpdatedAt = time.Now()
	task.Version++

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

// CompleteTask marks a task as completed
func (s *TaskServiceImpl) CompleteTask(ctx context.Context, id string, actualEffort *int, version int) (*models.Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Version check for optimistic locking
	if task.Version != version {
		return nil, fmt.Errorf("task version conflict")
	}

	task.Status = models.TaskStatusCompleted
	if actualEffort != nil {
		task.ActualEffort = actualEffort
	}
	task.CompletedAt = &time.Time{}
	*task.CompletedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.Version++

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

// UpdateTaskStatus updates the status of a task
func (s *TaskServiceImpl) UpdateTaskStatus(ctx context.Context, id string, status string, version int) (*models.Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// Version check for optimistic locking
	if task.Version != version {
		return nil, fmt.Errorf("task version conflict")
	}

	task.Status = models.TaskStatus(status)
	task.UpdatedAt = time.Now()
	task.Version++

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

// BulkUpdateTasks updates multiple tasks in bulk
func (s *TaskServiceImpl) BulkUpdateTasks(ctx context.Context, updates []models.TaskChange) ([]*models.Task, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}

	var updatedTasks []*models.Task
	for _, update := range updates {
		task, err := s.taskRepo.FindByID(ctx, update.TaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to find task %s: %w", update.TaskID, err)
		}
		if task == nil {
			continue // Skip if task not found
		}

		// Apply update (simplified - would need more logic for real implementation)
		task.UpdatedAt = time.Now()
		task.Version++

		if err := s.taskRepo.Update(ctx, task); err != nil {
			return nil, fmt.Errorf("failed to update task %s: %w", update.TaskID, err)
		}

		updatedTasks = append(updatedTasks, task)
	}

	return updatedTasks, nil
}

// ScanTasks retrieves all tasks for a project (used for scanning/analysis)
func (s *TaskServiceImpl) ScanTasks(ctx context.Context, projectID string) ([]models.Task, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}

	// Get all tasks for the project (simplified - would need pagination in real implementation)
	tasks, _, err := s.taskRepo.FindByProjectID(ctx, projectID, models.ListTasksRequest{
		ProjectID: projectID,
		Limit:     1000, // Large limit for scanning
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan tasks: %w", err)
	}

	return tasks, nil
}
