// Dependency Detection Helper Functions
// Utility functions for task reference lookup, code analysis, and cycle detection
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sentinel-hub-api/pkg/database"

	"github.com/google/uuid"
	sitter "github.com/smacker/go-tree-sitter"
)

// Note: sitter import is kept for future tree-sitter integration
// Currently unused but will be needed when AST parsing is implemented

// findTaskByReference finds a task by ID or title reference
func findTaskByReference(ctx context.Context, projectID string, ref string) (string, error) {
	// Try as UUID first
	if _, err := uuid.Parse(ref); err == nil {
		// It's a UUID, check if task exists
		task, err := GetTask(ctx, ref)
		if err == nil && task.ProjectID == projectID {
			return task.ID, nil
		}
	}

	// Try to find by title (partial match)
	query := `
		SELECT id FROM tasks 
		WHERE project_id = $1 AND LOWER(title) LIKE LOWER($2)
		LIMIT 1
	`

	searchPattern := "%" + ref + "%"
	row := database.QueryRowWithTimeout(ctx, db, query, projectID, searchPattern)

	var taskID string
	err := row.Scan(&taskID)
	if err != nil {
		// Check if it's a "not found" error (sql.ErrNoRows)
		if err == sql.ErrNoRows {
			// Task not found is not an error condition, return empty string
			return "", nil
		}
		// Actual database error
		return "", fmt.Errorf("failed to query tasks: %w", err)
	}

	if taskID != "" {
		return taskID, nil
	}

	// Task not found
	return "", nil
}

// getAllProjectTasks gets all tasks for a project
func getAllProjectTasks(ctx context.Context, projectID string) ([]Task, error) {
	req := ListTasksRequest{
		Limit:  1000, // Get all tasks
		Offset: 0,
	}

	response, err := ListTasks(ctx, projectID, req)
	if err != nil {
		return nil, err
	}

	return response.Tasks, nil
}

// calculateKeywordOverlap calculates overlap between two keyword sets
func calculateKeywordOverlap(keywords1, keywords2 []string) float64 {
	if len(keywords1) == 0 || len(keywords2) == 0 {
		return 0.0
	}

	// Create maps for faster lookup
	map1 := make(map[string]bool)
	for _, kw := range keywords1 {
		map1[strings.ToLower(kw)] = true
	}

	overlapCount := 0
	for _, kw := range keywords2 {
		if map1[strings.ToLower(kw)] {
			overlapCount++
		}
	}

	// Return overlap as percentage
	totalUnique := len(map1) + len(keywords2) - overlapCount
	if totalUnique == 0 {
		return 0.0
	}

	return float64(overlapCount) / float64(totalUnique)
}

// checkCodeReference checks if other task's code is referenced in file using AST analysis
func checkCodeReference(codebasePath, filePath string, otherTask *Task) bool {
	if filePath == "" || otherTask.FilePath == "" {
		// Fallback to keyword matching if file paths not available
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	currentFilePath := filepath.Join(codebasePath, filePath)
	otherFilePath := filepath.Join(codebasePath, otherTask.FilePath)

	// Read both files (not used currently as AST parsing is stubbed)
	_, err := os.ReadFile(currentFilePath)
	if err != nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	_, err = os.ReadFile(otherFilePath)
	if err != nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	// Determine language from file extensions (for future AST parsing)
	currentLang := detectLanguageFromFile(filePath)
	otherLang := detectLanguageFromFile(otherTask.FilePath)

	if currentLang == "" || otherLang == "" {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	// Note: AST parsing is currently stubbed out, so we fall back to keyword matching
	// This will be enabled when tree-sitter integration is complete
	// When implemented, this will extract symbols and check for references
	return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
}

// extractSymbolsFromAST extracts symbols (functions, classes) from AST
// Note: Currently returns empty map as AST parsing is stubbed out
// This will be implemented when tree-sitter integration is complete
func extractSymbolsFromAST(code string, language string, filePath string) map[string]bool {
	// AST parsing is currently stubbed out, return empty map
	// This function will be implemented when tree-sitter integration is complete
	return make(map[string]bool)
}

// checkSymbolReferences checks if symbols are referenced in AST
// Note: Currently stubbed out as AST parsing is not yet implemented
// This will be implemented when tree-sitter integration is complete
func checkSymbolReferences(root *sitter.Node, code string, language string, symbols map[string]bool) bool {
	// AST parsing is currently stubbed out, return false
	// This function will be implemented when tree-sitter integration is complete
	return false
}

// extractIdentifierFromNode extracts identifier name from AST node
func extractIdentifierFromNode(node *sitter.Node, code string) string {
	if node == nil {
		return ""
	}

	// Check if node is directly an identifier
	if node.Type() == "identifier" || node.Type() == "property_identifier" || node.Type() == "field_identifier" {
		return code[node.StartByte():node.EndByte()]
	}

	// Traverse children to find identifier
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && (child.Type() == "identifier" || child.Type() == "property_identifier" || child.Type() == "field_identifier") {
			return code[child.StartByte():child.EndByte()]
		}
	}

	return ""
}

// checkCodeReferenceKeywords is fallback keyword-based check
func checkCodeReferenceKeywords(codebasePath, filePath string, otherTask *Task) bool {
	if filePath == "" {
		return false
	}

	fullPath := filepath.Join(codebasePath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return false
	}

	contentStr := strings.ToLower(string(content))
	keywords := extractKeywords(otherTask.Title + " " + otherTask.Description)

	for _, keyword := range keywords {
		if strings.Contains(contentStr, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// storeDependency stores a dependency in the database
func storeDependency(ctx context.Context, dependency TaskDependency) error {
	query := `
		INSERT INTO task_dependencies (
			id, task_id, depends_on_task_id, dependency_type, confidence, created_at
		) VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (task_id, depends_on_task_id) 
		DO UPDATE SET 
			dependency_type = EXCLUDED.dependency_type,
			confidence = EXCLUDED.confidence
	`

	_, err := database.ExecWithTimeout(ctx, db, query,
		dependency.ID, dependency.TaskID, dependency.DependsOnTaskID,
		dependency.DependencyType, dependency.Confidence,
	)

	return err
}

// detectCycle detects circular dependencies using DFS
func detectCycle(ctx context.Context, startTaskID string, dependencies []TaskDependency) (bool, []string) {
	// Build adjacency list
	adj := make(map[string][]string)
	for _, dep := range dependencies {
		adj[dep.TaskID] = append(adj[dep.TaskID], dep.DependsOnTaskID)
	}

	// DFS to detect cycle
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(taskID string) bool
	dfs = func(taskID string) bool {
		visited[taskID] = true
		recStack[taskID] = true
		path = append(path, taskID)

		for _, neighbor := range adj[taskID] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Cycle detected - find the cycle portion
				// Find where the cycle starts in the current path
				cycleStart := -1
				for i, id := range path {
					if id == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					// Extract only the cycle portion: from cycleStart to end, plus the neighbor
					cyclePath := make([]string, 0, len(path)-cycleStart+1)
					cyclePath = append(cyclePath, path[cycleStart:]...)
					cyclePath = append(cyclePath, neighbor)
					path = cyclePath
					return true
				}
			}
		}

		recStack[taskID] = false
		path = path[:len(path)-1]
		return false
	}

	hasCycle := dfs(startTaskID)
	if hasCycle {
		return true, path
	}

	return false, nil
}
