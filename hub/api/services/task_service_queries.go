// Package services - Task service query operations
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/models"
)

// ListTasks retrieves tasks with filtering and pagination
func (s *TaskServiceImpl) ListTasks(ctx context.Context, req models.ListTasksRequest) (*models.ListTasksResponse, error) {
	tasks, total, err := s.taskRepo.FindByProjectID(ctx, req.ProjectID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Tasks are already values, no conversion needed
	taskList := tasks

	response := &models.ListTasksResponse{
		Tasks:   taskList,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: req.Offset+len(tasks) < total,
	}

	return response, nil
}

// AnalyzeTaskImpact analyzes the impact of task changes
func (s *TaskServiceImpl) AnalyzeTaskImpact(ctx context.Context, id string, change models.TaskChange) (*models.TaskImpactAnalysis, error) {
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

	// Get dependencies for impact analysis
	dependencies, err := s.taskRepo.FindDependencies(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	// Use impact analyzer if available
	if s.impactAnalyzer != nil {
		tasks := []models.Task{*task}
		analysis, err := s.impactAnalyzer.AnalyzeImpact(ctx, id, "task_change", tasks, dependencies)
		if err != nil {
			return nil, fmt.Errorf("impact analysis failed: %w", err)
		}
		return analysis, nil
	}

	// Fallback basic analysis
	affectedTaskTitles := []string{task.Title}
	if len(dependencies) > 0 {
		for _, dep := range dependencies {
			affectedTaskTitles = append(affectedTaskTitles, fmt.Sprintf("Task %s", dep.TaskID))
		}
	}

	return &models.TaskImpactAnalysis{
		ID:            fmt.Sprintf("impact_%s_%s", id, "change"),
		TaskID:        id,
		ChangeType:    "task_change",
		ImpactScope:   "task",
		AffectedTasks: affectedTaskTitles,
		RiskLevel:     "medium",
		RiskFactors:   []string{"Task modification", "Dependency impact"},
		MitigationSuggestions: []string{
			"Regular status monitoring",
			"Dependency tracking",
			"Change documentation",
		},
		EstimatedImpactTime: 2,
		ConfidenceScore:     0.7,
		AnalyzedAt:          time.Now().Format(time.RFC3339),
		PrimaryImpact: models.TaskImpact{
			TaskTitle:       task.Title,
			ImpactType:      "direct_change",
			Severity:        "medium",
			Description:     fmt.Sprintf("Task '%s' has been modified", task.Title),
			TimeImpact:      2,
			Confidence:      0.8,
			Recommendations: []string{"Monitor dependent tasks", "Review change impact"},
		},
		CascadeEffects:     []models.TaskImpact{},
		BlockingTasks:      []string{},
		CriticalPathImpact: 0,
	}, nil
}
