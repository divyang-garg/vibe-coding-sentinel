// Package services - Task service core functionality
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sentinel-hub-api/models"
)

// TaskServiceImpl implements TaskService
type TaskServiceImpl struct {
	taskRepo       TaskRepository
	depAnalyzer    DependencyAnalyzer
	impactAnalyzer ImpactAnalyzer
}

// NewTaskService creates a new task service instance
// Complies with CODING_STANDARDS.md: Constructor injection with interfaces
func NewTaskService(taskRepo TaskRepository, depAnalyzer DependencyAnalyzer, impactAnalyzer ImpactAnalyzer) TaskService {
	return &TaskServiceImpl{
		taskRepo:       taskRepo,
		depAnalyzer:    depAnalyzer,
		impactAnalyzer: impactAnalyzer,
	}
}

// AnalyzeDependencies analyzes task dependencies and returns a dependency graph
func (s *TaskServiceImpl) AnalyzeDependencies(ctx context.Context, taskID string) (*models.TaskDependencyGraph, error) {
	dependencies, err := s.taskRepo.FindDependencies(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	// Basic dependency analysis
	graph := &models.TaskDependencyGraph{
		Dependencies: dependencies,
		IsValid:      true,
		GeneratedAt:  generateTimestamp().Format(time.RFC3339),
		Tasks:        make(map[string]*models.Task),
	}

	return graph, nil
}

// GetDependencies retrieves task dependencies and returns a response
func (s *TaskServiceImpl) GetDependencies(ctx context.Context, taskID string) (*models.DependencyGraphResponse, error) {
	dependencies, err := s.taskRepo.FindDependencies(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}

	response := &models.DependencyGraphResponse{
		Graph: &models.TaskDependencyGraph{
			Dependencies: dependencies,
			IsValid:      true,
			GeneratedAt:  generateTimestamp().Format(time.RFC3339),
		},
		IsValid:   true,
		Cycles:    [][]string{},
		Generated: generateTimestamp().Format(time.RFC3339),
	}

	return response, nil
}

// GetTaskExecutionPlan creates an execution plan for the given tasks
func (s *TaskServiceImpl) GetTaskExecutionPlan(ctx context.Context, taskIDs []string) (*models.TaskExecutionPlan, error) {
	if len(taskIDs) == 0 {
		return nil, &models.ValidationError{
			Field:   "task_ids",
			Message: "at least one task ID is required",
		}
	}

	// Load tasks
	var tasks []models.Task
	for _, id := range taskIDs {
		task, err := s.taskRepo.FindByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to find task: %w", err)
		}
		if task == nil {
			return nil, fmt.Errorf("failed to find task %s", id)
		}
		tasks = append(tasks, *task)
	}

	plan := &models.TaskExecutionPlan{
		Dependencies: []models.TaskDependency{},
		Batches:      [][]string{taskIDs}, // Simple single batch for now
		Tasks:        tasks,
		RiskFactors:  []string{},
		CreatedAt:    time.Now(),
	}

	return plan, nil
}

// GetTaskImpactAnalysis retrieves the impact analysis for a task
func (s *TaskServiceImpl) GetTaskImpactAnalysis(ctx context.Context, taskID string) (*models.TaskImpactAnalysis, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, &models.NotFoundError{
			Resource: "task",
			Message:  fmt.Sprintf("task not found: %s", taskID),
		}
	}

	// Create basic impact analysis
	analysis := &models.TaskImpactAnalysis{
		ID:                    fmt.Sprintf("impact_%s", taskID),
		TaskID:                taskID,
		ChangeType:            "analysis",
		ImpactScope:           "task",
		AffectedTasks:         []string{taskID},
		RiskLevel:             "low",
		RiskFactors:           []string{},
		MitigationSuggestions: []string{"Regular monitoring"},
		EstimatedImpactTime:   0,
		ConfidenceScore:       0.8,
		AnalyzedAt:            time.Now().Format(time.RFC3339),
		PrimaryImpact: models.TaskImpact{
			TaskID:            taskID,
			ChangeDescription: "Impact analysis requested",
			AffectedTasks:     []string{},
			RiskLevel:         "low",
			TimeImpact:        0,
			Recommendations:   []string{"Monitor task progress"},
		},
		CascadeEffects:       []models.TaskImpact{},
		BlockingTasks:        []string{},
		CriticalPathImpact:   0,
		OverallRiskLevel:     "low",
		MitigationStrategies: []string{"Standard monitoring"},
		AnalysisConfidence:   0.8,
		GeneratedAt:          time.Now().Format(time.RFC3339),
	}

	return analysis, nil
}
