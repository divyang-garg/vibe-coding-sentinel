// Phase 14E: Dependency Detection Engine - Main Functions
// Detects explicit, implicit, integration, and feature-level dependencies
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"fmt"
	"time"

	"sentinel-hub-api/pkg/database"
)

// DetectDependencies detects all types of dependencies for a task
func DetectDependencies(ctx context.Context, taskID string, codebasePath string) ([]TaskDependency, error) {
	task, err := GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var dependencies []TaskDependency

	// 1. Explicit dependencies (from task description)
	explicitDeps, err := detectExplicitDependencies(ctx, task)
	if err != nil {
		LogError(ctx, "Failed to detect explicit dependencies: %v", err)
	} else {
		dependencies = append(dependencies, explicitDeps...)
	}

	// 2. Implicit dependencies (from code analysis)
	implicitDeps, err := detectImplicitDependencies(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Failed to detect implicit dependencies: %v", err)
	} else {
		dependencies = append(dependencies, implicitDeps...)
	}

	// 3. Integration dependencies (if Phase 14A available)
	integrationDeps, err := detectIntegrationDependencies(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Failed to detect integration dependencies: %v", err)
	} else {
		dependencies = append(dependencies, integrationDeps...)
	}

	// 4. Feature-level dependencies (if Phase 14A available)
	featureDeps, err := detectFeatureDependencies(ctx, task, codebasePath)
	if err != nil {
		LogError(ctx, "Failed to detect feature dependencies: %v", err)
	} else {
		dependencies = append(dependencies, featureDeps...)
	}

	// Store dependencies
	for _, dep := range dependencies {
		if err := storeDependency(ctx, dep); err != nil {
			LogError(ctx, "Failed to store dependency: %v", err)
		}
	}

	return dependencies, nil
}

// GetTaskDependencies retrieves dependencies for a task (with caching)
func GetTaskDependencies(ctx context.Context, taskID string) (*DependencyGraphResponse, error) {
	// Check cache first
	if cachedDeps, found := GetCachedDependencies(taskID); found {
		return cachedDeps, nil
	}

	query := `
		SELECT id, task_id, depends_on_task_id, dependency_type, confidence
		FROM task_dependencies
		WHERE task_id = $1
	`

	rows, err := database.QueryWithTimeout(ctx, db, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %w", err)
	}
	defer rows.Close()

	var dependencies []TaskDependency
	blockedBy := make(map[string]bool)
	blocks := make(map[string]bool)

	for rows.Next() {
		var dep TaskDependency
		err := rows.Scan(
			&dep.ID, &dep.TaskID, &dep.DependsOnTaskID,
			&dep.DependencyType, &dep.Confidence,
		)
		if err != nil {
			continue
		}
		dependencies = append(dependencies, dep)
		blockedBy[dep.DependsOnTaskID] = true
	}

	// Find tasks that depend on this task (reverse lookup)
	reverseQuery := `
		SELECT task_id FROM task_dependencies
		WHERE depends_on_task_id = $1
	`
	reverseRows, err := database.QueryWithTimeout(ctx, db, reverseQuery, taskID)
	if err == nil {
		defer reverseRows.Close()
		for reverseRows.Next() {
			var blockingTaskID string
			if err := reverseRows.Scan(&blockingTaskID); err == nil {
				blocks[blockingTaskID] = true
			}
		}
	}

	// Convert maps to slices
	blockedByList := make([]string, 0, len(blockedBy))
	for id := range blockedBy {
		blockedByList = append(blockedByList, id)
	}

	blocksList := make([]string, 0, len(blocks))
	for id := range blocks {
		blocksList = append(blocksList, id)
	}

	// Check for cycles
	hasCycle, cyclePath := detectCycle(ctx, taskID, dependencies)

	// Build TaskDependencyGraph
	graph := &TaskDependencyGraph{
		Tasks:          make(map[string]*Task),
		Dependencies:   dependencies,
		ExecutionOrder: []string{},
		Cycles:         [][]string{},
		IsValid:        !hasCycle,
		GeneratedAt:    time.Now().Format(time.RFC3339),
	}
	if hasCycle && len(cyclePath) > 0 {
		graph.Cycles = [][]string{cyclePath}
	}

	response := &DependencyGraphResponse{
		Graph:     graph,
		IsValid:   !hasCycle,
		Cycles:    graph.Cycles,
		Generated: graph.GeneratedAt,
	}

	// Cache the dependencies
	SetCachedDependencies(taskID, response)

	return response, nil
}
