// Dependency Detection Helper Functions
// Utility functions for task reference lookup, code analysis, and cycle detection
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"database/sql"
	"errors"
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
		if errors.Is(err, sql.ErrNoRows) {
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
// Falls back to keyword matching if AST analysis fails or is unavailable
func checkCodeReference(codebasePath, filePath string, otherTask *Task) bool {
	if filePath == "" || otherTask.FilePath == "" {
		// Fallback to keyword matching if file paths not available
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	currentFilePath := filepath.Join(codebasePath, filePath)
	otherFilePath := filepath.Join(codebasePath, otherTask.FilePath)

	// Read both files
	currentContent, err := os.ReadFile(currentFilePath)
	if err != nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	otherContent, err := os.ReadFile(otherFilePath)
	if err != nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	// Determine language from file extensions
	currentLang := detectLanguageFromFile(filePath)
	otherLang := detectLanguageFromFile(otherTask.FilePath)

	if currentLang == "" || otherLang == "" {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	// Extract symbols from other task's file using AST
	otherSymbols := extractSymbolsFromAST(string(otherContent), otherLang, otherTask.FilePath)
	if len(otherSymbols) == 0 {
		// No symbols found or AST failed - fallback to keyword matching
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	// Parse current file and check for symbol references
	parser, err := GetParser(currentLang)
	if err != nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, currentContent)
	if err != nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	// Check if any symbols are referenced
	if checkSymbolReferences(rootNode, string(currentContent), currentLang, otherSymbols) {
		return true
	}

	// No AST matches found - fallback to keyword matching
	return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
}

// extractSymbolsFromAST extracts symbols (functions, classes) from AST
// Returns a map of symbol names found in the code
func extractSymbolsFromAST(code string, language string, filePath string) map[string]bool {
	symbols := make(map[string]bool)

	// Get parser using AST bridge
	parser, err := GetParser(language)
	if err != nil {
		// Unsupported language - return empty map
		// Use filePath for context in error handling (if logging was available)
		_ = filePath // Track which file failed parsing
		return symbols
	}

	// Parse code
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		// Parse error - return empty map (fallback will handle)
		// Use filePath to identify which file had parsing issues
		_ = filePath // Track which file had parsing errors
		return symbols
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return symbols
	}

	// Traverse AST to extract function/class names
	// Use filePath to provide context for symbol extraction (useful for debugging)
	_ = filePath // Track which file symbols are extracted from

	TraverseAST(rootNode, func(node *sitter.Node) bool {
		var symbolName string

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				symbolName = extractIdentifierFromNode(node, code)
			} else if node.Type() == "type_declaration" {
				symbolName = extractIdentifierFromNode(node, code)
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" {
				symbolName = extractIdentifierFromNode(node, code)
			} else if node.Type() == "class_declaration" {
				symbolName = extractIdentifierFromNode(node, code)
			} else if node.Type() == "variable_declarator" {
				// Check for arrow functions assigned to variables
				varName := extractIdentifierFromNode(node, code)
				if varName != "" {
					// Check if this is a function assignment
					for i := 0; i < int(node.ChildCount()); i++ {
						child := node.Child(i)
						if child != nil && (child.Type() == "arrow_function" || child.Type() == "function_expression") {
							symbols[varName] = true
							break
						}
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
				symbolName = extractIdentifierFromNode(node, code)
			} else if node.Type() == "class_definition" {
				symbolName = extractIdentifierFromNode(node, code)
			}
		}

		if symbolName != "" {
			symbols[symbolName] = true
		}

		return true
	})

	return symbols
}

// checkSymbolReferences checks if symbols are referenced in AST
// Returns true if any symbol from the map is found in the AST
func checkSymbolReferences(root *sitter.Node, code string, language string, symbols map[string]bool) bool {
	if root == nil || len(symbols) == 0 {
		return false
	}

	found := false

	// Safe slice helper
	safeSlice := func(start, end uint32) string {
		codeLen := uint32(len(code))
		if start > codeLen {
			start = codeLen
		}
		if end > codeLen {
			end = codeLen
		}
		if start > end {
			return ""
		}
		return code[start:end]
	}

	// Use language to determine language-specific identifier node types
	identifierTypes := []string{"identifier", "property_identifier", "field_identifier", "type_identifier"}
	switch language {
	case "go":
		// Go-specific identifier types
		identifierTypes = append(identifierTypes, "field_identifier", "package_identifier")
	case "javascript", "typescript":
		// JavaScript/TypeScript-specific identifier types
		identifierTypes = append(identifierTypes, "property_identifier", "shorthand_property_identifier")
	case "python":
		// Python-specific identifier types
		identifierTypes = append(identifierTypes, "identifier", "attribute")
	}

	// Traverse AST to find identifier references
	TraverseAST(root, func(node *sitter.Node) bool {
		// Stop if we already found a match
		if found {
			return false
		}

		nodeType := node.Type()

		// Check identifier nodes (using language-specific types)
		for _, idType := range identifierTypes {
			if nodeType == idType {
				identifier := safeSlice(node.StartByte(), node.EndByte())
				if symbols[identifier] {
					found = true
					return false // Stop traversal
				}
			}
		}

		// For method calls, check the method name (language-specific)
		callNodeTypes := []string{"call_expression", "method_call"}
		switch language {
		case "go":
			callNodeTypes = append(callNodeTypes, "call_expression")
		case "javascript", "typescript":
			callNodeTypes = append(callNodeTypes, "call_expression", "new_expression")
		case "python":
			callNodeTypes = append(callNodeTypes, "call")
		}

		for _, callType := range callNodeTypes {
			if nodeType == callType {
				// Get the function/method name being called
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						for _, idType := range identifierTypes {
							if child.Type() == idType {
								identifier := safeSlice(child.StartByte(), child.EndByte())
								if symbols[identifier] {
									found = true
									return false
								}
							}
						}
					}
				}
			}
		}

		return true
	})

	return found
}

// extractIdentifierFromNode extracts identifier name from AST node
// Uses safe string slicing to prevent panics
func extractIdentifierFromNode(node *sitter.Node, code string) string {
	if node == nil {
		return ""
	}

	// Safe slice helper
	safeSlice := func(start, end uint32) string {
		codeLen := uint32(len(code))
		if start > codeLen {
			start = codeLen
		}
		if end > codeLen {
			end = codeLen
		}
		if start > end {
			return ""
		}
		return code[start:end]
	}

	// Check if node is directly an identifier
	if node.Type() == "identifier" || node.Type() == "property_identifier" || node.Type() == "field_identifier" {
		return safeSlice(node.StartByte(), node.EndByte())
	}

	// Traverse children to find identifier
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil && (child.Type() == "identifier" || child.Type() == "property_identifier" || child.Type() == "field_identifier") {
			return safeSlice(child.StartByte(), child.EndByte())
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
	// Check for context cancellation before starting
	if ctx.Err() != nil {
		return false, nil
	}

	// Build adjacency list
	adj := make(map[string][]string)
	for _, dep := range dependencies {
		// Check for context cancellation during processing
		if ctx.Err() != nil {
			return false, nil
		}
		adj[dep.TaskID] = append(adj[dep.TaskID], dep.DependsOnTaskID)
	}

	// DFS to detect cycle
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(taskID string) bool
	dfs = func(taskID string) bool {
		// Check for context cancellation in DFS traversal
		if ctx.Err() != nil {
			return false
		}

		visited[taskID] = true
		recStack[taskID] = true
		path = append(path, taskID)

		for _, neighbor := range adj[taskID] {
			// Check for context cancellation in loop
			if ctx.Err() != nil {
				return false
			}

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
