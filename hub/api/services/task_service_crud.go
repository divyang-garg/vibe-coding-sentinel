// Package services - Task service CRUD operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/models"
)

// CreateTask creates a new task
func (s *TaskServiceImpl) CreateTask(ctx context.Context, req models.CreateTaskRequest) (*models.Task, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project ID is required")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	now := time.Now()
	task := &models.Task{
		ID:                     generateID(),
		ProjectID:              req.ProjectID,
		Source:                 req.Source,
		Title:                  req.Title,
		Description:            req.Description,
		FilePath:               req.FilePath,
		LineNumber:             req.LineNumber,
		Status:                 models.TaskStatusPending,
		Priority:               models.TaskPriority(req.Priority),
		AssignedTo:             req.AssignedTo,
		EstimatedEffort:        req.EstimatedEffort,
		Tags:                   req.Tags,
		VerificationConfidence: req.VerificationConfidence,
		CreatedAt:              now,
		UpdatedAt:              now,
		Version:                1,
	}

	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}

// GetTaskByID retrieves a task by ID
func (s *TaskServiceImpl) GetTaskByID(ctx context.Context, id string) (*models.Task, error) {
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

	return task, nil
}

// UpdateTask updates a task
func (s *TaskServiceImpl) UpdateTask(ctx context.Context, id string, req models.UpdateTaskRequest) (*models.Task, error) {
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

	// Apply updates
	if req.Title != nil && *req.Title != "" {
		task.Title = *req.Title
	}
	if req.Description != nil && *req.Description != "" {
		task.Description = *req.Description
	}
	if req.FilePath != nil && *req.FilePath != "" {
		task.FilePath = *req.FilePath
	}
	if req.LineNumber != nil && *req.LineNumber != 0 {
		task.LineNumber = req.LineNumber
	}
	if req.Status != nil && *req.Status != "" {
		task.Status = models.TaskStatus(*req.Status)
	}
	if req.Priority != nil && *req.Priority != "" {
		task.Priority = models.TaskPriority(*req.Priority)
	}
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		task.AssignedTo = req.AssignedTo
	}
	if req.EstimatedEffort != nil {
		task.EstimatedEffort = req.EstimatedEffort
	}
	if req.ActualEffort != nil {
		task.ActualEffort = req.ActualEffort
	}

	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

// DeleteTask deletes a task
func (s *TaskServiceImpl) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("task ID is required")
	}

	// Check if task exists
	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}

	// Delete associated dependencies first
	dependencies, err := s.taskRepo.FindDependencies(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find task dependencies: %w", err)
	}

	for _, dep := range dependencies {
		if err := s.taskRepo.DeleteDependency(ctx, dep.ID); err != nil {
			return fmt.Errorf("failed to delete dependency %s: %w", dep.ID, err)
		}
	}

	// Delete task
	if err := s.taskRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}
