// Package services - Task service dependency management
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"

	"sentinel-hub-api/models"
)

// AddDependency adds a dependency between tasks
func (s *TaskServiceImpl) AddDependency(ctx context.Context, taskID string, req models.AddDependencyRequest) (*models.TaskDependency, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if req.DependsOnTaskID == "" {
		return nil, fmt.Errorf("depends on task ID is required")
	}

	// Verify both tasks exist
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, &models.NotFoundError{
			Resource: "task",
			Message:  "task not found",
		}
	}

	dependsOnTask, err := s.taskRepo.FindByID(ctx, req.DependsOnTaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find depends on task: %w", err)
	}
	if dependsOnTask == nil {
		return nil, &models.NotFoundError{
			Resource: "task",
			Message:  "depends on task not found",
		}
	}

	// Check for cycles if dependency analyzer is available
	if s.depAnalyzer != nil {
		// Get existing dependencies for cycle detection
		existingDeps, err := s.taskRepo.FindDependencies(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing dependencies: %w", err)
		}

		// Add the new dependency for cycle checking
		testDeps := append(existingDeps, models.TaskDependency{
			TaskID:          taskID,
			DependsOnTaskID: req.DependsOnTaskID,
		})

		cycles, err := s.depAnalyzer.DetectCycles(ctx, testDeps)
		if err != nil {
			return nil, fmt.Errorf("failed to check for cycles: %w", err)
		}
		if len(cycles) > 0 {
			return nil, fmt.Errorf("dependency would create a cycle")
		}
	}

	// Create dependency
	dependency := &models.TaskDependency{
		ID:              generateID(),
		TaskID:          taskID,
		DependsOnTaskID: req.DependsOnTaskID,
		DependencyType:  req.DependencyType,
		Confidence:      req.Confidence,
		CreatedAt:       generateTimestamp(),
	}

	if err := s.taskRepo.SaveDependency(ctx, dependency); err != nil {
		return nil, fmt.Errorf("failed to save dependency: %w", err)
	}

	return dependency, nil
}

// VerifyTask verifies task completion
func (s *TaskServiceImpl) VerifyTask(ctx context.Context, id string, req models.VerifyTaskRequest) (*models.VerifyTaskResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	task, err := s.taskRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, &models.NotFoundError{
			Resource: "task",
			Message:  "task not found",
		}
	}

	// Create verification record
	verification := &models.TaskVerification{
		ID:         generateID(),
		TaskID:     id,
		Status:     models.VerificationStatus(req.Status),
		Confidence: req.Confidence,
		VerifiedBy: req.VerifiedBy,
		VerifiedAt: &req.VerifiedAt,
		Notes:      req.Notes,
		Evidence:   req.Evidence,
		CreatedAt:  generateTimestamp(),
	}

	if err := s.taskRepo.SaveVerification(ctx, verification); err != nil {
		return nil, fmt.Errorf("failed to save verification: %w", err)
	}

	// Update task verification confidence
	task.VerificationConfidence = req.Confidence
	task.UpdatedAt = generateTimestamp()
	task.Version++

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return &models.VerifyTaskResponse{
		Task:         task,
		Verification: verification,
		Success:      true,
	}, nil
}
