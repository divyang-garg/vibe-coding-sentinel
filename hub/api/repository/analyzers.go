// Package repository contains analyzer implementations for dependency and impact analysis.
package repository

import (
	"context"
	"fmt"
	"sentinel-hub-api/models"
	"time"
)

// DependencyAnalyzerImpl implements DependencyAnalyzer
type DependencyAnalyzerImpl struct{}

// NewDependencyAnalyzer creates a new dependency analyzer instance
func NewDependencyAnalyzer() *DependencyAnalyzerImpl {
	return &DependencyAnalyzerImpl{}
}

// AnalyzeDependencies analyzes task dependencies and builds a dependency graph
func (a *DependencyAnalyzerImpl) AnalyzeDependencies(ctx context.Context, tasks []models.Task) (*models.TaskDependencyGraph, error) {
	graph := &models.TaskDependencyGraph{
		Tasks:          make(map[string]*models.Task),
		Dependencies:   []models.TaskDependency{},
		ExecutionOrder: []string{},
		IsValid:        true,
		Errors:         []string{},
	}

	// Add all tasks to the graph
	for i := range tasks {
		task := tasks[i]
		graph.Tasks[task.ID] = &task
	}

	// In a real implementation, this would analyze the task relationships
	// For now, return a basic valid graph
	graph.ExecutionOrder = make([]string, len(tasks))
	for i, task := range tasks {
		graph.ExecutionOrder[i] = task.ID
	}

	return graph, nil
}

// DetectCycles detects circular dependencies in the dependency graph
func (a *DependencyAnalyzerImpl) DetectCycles(ctx context.Context, dependencies []models.TaskDependency) ([][]string, error) {
	// Basic cycle detection implementation
	// In production, this would use topological sort or DFS
	var cycles [][]string

	// Simple check for self-dependencies
	for _, dep := range dependencies {
		if dep.TaskID == dep.DependsOnTaskID {
			cycles = append(cycles, []string{dep.TaskID})
		}
	}

	return cycles, nil
}

// CreateExecutionPlan creates an execution plan for interdependent tasks
func (a *DependencyAnalyzerImpl) CreateExecutionPlan(ctx context.Context, graph *models.TaskDependencyGraph) (*models.TaskExecutionPlan, error) {
	plan := &models.TaskExecutionPlan{
		Tasks:        make([]models.Task, 0, len(graph.Tasks)),
		Dependencies: graph.Dependencies,
		Batches:      [][]string{graph.ExecutionOrder}, // Single batch for simplicity
		RiskFactors:  []string{},
		CreatedAt:    time.Now(),
	}

	// Convert tasks map to slice
	for _, task := range graph.Tasks {
		plan.Tasks = append(plan.Tasks, *task)
	}

	// Estimate duration (simplified)
	if len(plan.Tasks) > 0 {
		plan.EstimatedDuration = 30 * time.Minute // 30 minutes per task estimate
	}

	return plan, nil
}

// ImpactAnalyzerImpl implements ImpactAnalyzer
type ImpactAnalyzerImpl struct{}

// NewImpactAnalyzer creates a new impact analyzer instance
func NewImpactAnalyzer() *ImpactAnalyzerImpl {
	return &ImpactAnalyzerImpl{}
}

// AnalyzeImpact analyzes the impact of task changes
func (a *ImpactAnalyzerImpl) AnalyzeImpact(ctx context.Context, taskID string, changeType string, tasks []models.Task, dependencies []models.TaskDependency) (*models.TaskImpactAnalysis, error) {
	analysis := &models.TaskImpactAnalysis{
		ID:                    fmt.Sprintf("impact_%s_%d", taskID, time.Now().Unix()),
		TaskID:                taskID,
		ChangeType:            changeType,
		ImpactScope:           "project",
		AffectedTasks:         []string{},
		RiskLevel:             "low",
		RiskFactors:           []string{},
		MitigationSuggestions: []string{"Review changes carefully", "Test thoroughly"},
		EstimatedImpactTime:   60, // 1 hour estimate
		ConfidenceScore:       0.8,
		AnalyzedAt:            time.Now().Format(time.RFC3339),
	}

	// Find the changed task
	var changedTask *models.Task
	for i := range tasks {
		if tasks[i].ID == taskID {
			changedTask = &tasks[i]
			break
		}
	}

	if changedTask == nil {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	// Analyze impact on dependent tasks
	for _, dep := range dependencies {
		if dep.DependsOnTaskID == taskID {
			// This task depends on the changed task
			for _, task := range tasks {
				if task.ID == dep.TaskID {
					impact := models.TaskImpact{
						TaskID:      task.ID,
						TaskTitle:   task.Title,
						ImpactType:  "dependency",
						Severity:    "medium",
						Description: fmt.Sprintf("Depends on changed task '%s'", changedTask.Title),
						TimeImpact:  30, // 30 minutes additional work
						Confidence:  0.9,
					}
					analysis.AffectedTasks = append(analysis.AffectedTasks, impact.TaskID)
					break
				}
			}
		}
	}

	// Adjust risk level based on number of affected tasks
	if len(analysis.AffectedTasks) > 5 {
		analysis.RiskLevel = "high"
		analysis.RiskFactors = append(analysis.RiskFactors, "High number of dependent tasks")
	} else if len(analysis.AffectedTasks) > 2 {
		analysis.RiskLevel = "medium"
		analysis.RiskFactors = append(analysis.RiskFactors, "Multiple dependent tasks")
	}

	return analysis, nil
}

// CalculateRiskLevel calculates the overall risk level for a set of impacts
func (a *ImpactAnalyzerImpl) CalculateRiskLevel(ctx context.Context, impacts []models.TaskImpact) string {
	if len(impacts) == 0 {
		return "low"
	}

	maxSeverity := "low"
	totalTimeImpact := 0

	for _, impact := range impacts {
		totalTimeImpact += impact.TimeImpact

		// Determine highest severity
		switch impact.Severity {
		case "critical":
			return "critical"
		case "high":
			maxSeverity = "high"
		case "medium":
			if maxSeverity == "low" {
				maxSeverity = "medium"
			}
		}
	}

	// Adjust based on total time impact
	if totalTimeImpact > 480 { // 8 hours
		return "critical"
	} else if totalTimeImpact > 240 { // 4 hours
		if maxSeverity == "low" {
			maxSeverity = "medium"
		}
	}

	return maxSeverity
}

// SuggestMitigations provides suggestions for mitigating identified risks
func (a *ImpactAnalyzerImpl) SuggestMitigations(ctx context.Context, impacts []models.TaskImpact) []string {
	suggestions := []string{
		"Review all changes carefully before deployment",
		"Consider phased rollout if impacts are extensive",
		"Ensure comprehensive testing of affected functionality",
	}

	// Add specific suggestions based on impact types
	hasDependencyImpacts := false
	hasBreakingChanges := false

	for _, impact := range impacts {
		if impact.ImpactType == "dependency" {
			hasDependencyImpacts = true
		}
		if impact.Severity == "critical" {
			hasBreakingChanges = true
		}
	}

	if hasDependencyImpacts {
		suggestions = append(suggestions,
			"Update dependent task documentation",
			"Notify stakeholders of dependent tasks")
	}

	if hasBreakingChanges {
		suggestions = append(suggestions,
			"Schedule maintenance window",
			"Prepare rollback plan",
			"Consider feature flags for gradual rollout")
	}

	return suggestions
}
