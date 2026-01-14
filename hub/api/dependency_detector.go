// Phase 14E: Dependency Detection Engine
// Detects explicit, implicit, integration, and feature-level dependencies

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	sitter "github.com/smacker/go-tree-sitter"
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

// detectExplicitDependencies parses explicit dependencies from task description
func detectExplicitDependencies(ctx context.Context, task *Task) ([]TaskDependency, error) {
	var dependencies []TaskDependency

	// Pattern: "DEPENDS: TASK-123, TASK-456" or "Depends on: TASK-123"
	dependsPattern := regexp.MustCompile(`(?i)(?:depends|dependency)[\s:]+(?:on|upon)?[\s:]*([A-Z0-9-,\s]+)`)

	text := task.Title + " " + task.Description
	matches := dependsPattern.FindStringSubmatch(text)

	if len(matches) > 1 {
		depList := matches[1]
		// Split by comma and clean up
		taskRefs := strings.Split(depList, ",")

		for _, ref := range taskRefs {
			ref = strings.TrimSpace(ref)
			if ref == "" {
				continue
			}

			// Try to find task by ID or title
			depTaskID, err := findTaskByReference(ctx, task.ProjectID, ref)
			if err != nil {
				LogError(ctx, "Failed to find task %s: %v", ref, err)
				continue
			}

			if depTaskID == "" {
				continue
			}

			dependency := TaskDependency{
				ID:              uuid.New().String(),
				TaskID:          task.ID,
				DependsOnTaskID: depTaskID,
				DependencyType:  "explicit",
				Confidence:      0.95, // High confidence for explicit dependencies
			}

			dependencies = append(dependencies, dependency)
		}
	}

	return dependencies, nil
}

// detectImplicitDependencies detects dependencies through code analysis
func detectImplicitDependencies(ctx context.Context, task *Task, codebasePath string) ([]TaskDependency, error) {
	var dependencies []TaskDependency

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)
	if len(keywords) == 0 {
		return dependencies, nil
	}

	// Get all tasks in the project
	allTasks, err := getAllProjectTasks(ctx, task.ProjectID)
	if err != nil {
		return dependencies, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// For each other task, check if it's related
	for _, otherTask := range allTasks {
		if otherTask.ID == task.ID {
			continue
		}

		// Check if keywords overlap
		otherKeywords := extractKeywords(otherTask.Title + " " + otherTask.Description)
		overlap := calculateKeywordOverlap(keywords, otherKeywords)

		if overlap > 0.3 { // 30% keyword overlap suggests dependency
			// Check if other task's code is referenced in this task's file
			hasCodeReference := false
			if task.FilePath != "" {
				hasCodeReference = checkCodeReference(codebasePath, task.FilePath, &otherTask)
			}

			confidence := overlap
			if hasCodeReference {
				confidence = 0.7 // Higher confidence if code reference found
			}

			if confidence > 0.3 {
				dependency := TaskDependency{
					ID:              uuid.New().String(),
					TaskID:          task.ID,
					DependsOnTaskID: otherTask.ID,
					DependencyType:  "implicit",
					Confidence:      confidence,
				}
				dependencies = append(dependencies, dependency)
			}
		}
	}

	return dependencies, nil
}

// detectIntegrationDependencies detects integration dependencies
func detectIntegrationDependencies(ctx context.Context, task *Task, codebasePath string) ([]TaskDependency, error) {
	var dependencies []TaskDependency

	// Check if task mentions integration keywords
	integrationKeywords := []string{"api", "integration", "service", "external", "third-party", "sdk", "client"}
	taskText := strings.ToLower(task.Title + " " + task.Description)

	hasIntegrationKeyword := false
	for _, keyword := range integrationKeywords {
		if strings.Contains(taskText, keyword) {
			hasIntegrationKeyword = true
			break
		}
	}

	if !hasIntegrationKeyword {
		return dependencies, nil // No integration dependency
	}

	// Query comprehensive analysis for integration-related features
	query := `
		SELECT DISTINCT validation_id, feature
		FROM comprehensive_validations
		WHERE project_id = $1 
		AND (
			LOWER(feature) LIKE ANY(ARRAY['%api%', '%integration%', '%service%', '%external%', '%sdk%', '%client%'])
			OR LOWER(findings::text) LIKE ANY(ARRAY['%api%', '%integration%', '%service%', '%external%', '%sdk%', '%client%'])
		)
	`

	rows, err := queryWithTimeout(ctx, query, task.ProjectID)
	if err != nil {
		// If table doesn't exist or query fails, return empty (graceful degradation)
		return dependencies, nil
	}
	defer rows.Close()

	validationFeatures := make(map[string]string) // validation_id -> feature
	for rows.Next() {
		var validationID, feature string
		if err := rows.Scan(&validationID, &feature); err == nil {
			validationFeatures[validationID] = feature
		}
	}

	// Check if task keywords match any integration features
	keywords := extractKeywords(task.Title + " " + task.Description)
	for validationID, feature := range validationFeatures {
		featureKeywords := extractKeywords(feature)
		overlap := calculateKeywordOverlap(keywords, featureKeywords)

		if overlap > 0.3 {
			// Find tasks linked to this comprehensive analysis
			linkQuery := `
				SELECT task_id FROM task_links
				WHERE link_type = 'comprehensive_analysis' AND linked_id = $1
			`
			linkRows, err := queryWithTimeout(ctx, linkQuery, validationID)
			if err == nil {
				defer linkRows.Close()
				for linkRows.Next() {
					var depTaskID string
					if err := linkRows.Scan(&depTaskID); err == nil && depTaskID != task.ID {
						dependency := TaskDependency{
							ID:              uuid.New().String(),
							TaskID:          task.ID,
							DependsOnTaskID: depTaskID,
							DependencyType:  "integration",
							Confidence:      overlap,
						}
						dependencies = append(dependencies, dependency)
					}
				}
			}
		}
	}

	return dependencies, nil
}

// detectFeatureDependencies detects feature-level dependencies
func detectFeatureDependencies(ctx context.Context, task *Task, codebasePath string) ([]TaskDependency, error) {
	var dependencies []TaskDependency

	// Query comprehensive analysis for feature dependencies
	query := `
		SELECT validation_id, feature, checklist
		FROM comprehensive_validations
		WHERE project_id = $1
	`

	rows, err := queryWithTimeout(ctx, query, task.ProjectID)
	if err != nil {
		// If table doesn't exist or query fails, return empty (graceful degradation)
		return dependencies, nil
	}
	defer rows.Close()

	type FeatureInfo struct {
		ValidationID string
		Feature      string
		Checklist    string
	}

	var features []FeatureInfo
	for rows.Next() {
		var fi FeatureInfo
		var checklist sql.NullString
		if err := rows.Scan(&fi.ValidationID, &fi.Feature, &checklist); err == nil {
			if checklist.Valid {
				fi.Checklist = checklist.String
			}
			features = append(features, fi)
		}
	}

	// Extract keywords from task
	keywords := extractKeywords(task.Title + " " + task.Description)

	// Find matching features
	for _, feature := range features {
		featureKeywords := extractKeywords(feature.Feature + " " + feature.Checklist)
		overlap := calculateKeywordOverlap(keywords, featureKeywords)

		if overlap > 0.3 {
			// Find tasks linked to this comprehensive analysis
			linkQuery := `
				SELECT task_id FROM task_links
				WHERE link_type = 'comprehensive_analysis' AND linked_id = $1
			`
			linkRows, err := queryWithTimeout(ctx, linkQuery, feature.ValidationID)
			if err == nil {
				defer linkRows.Close()
				for linkRows.Next() {
					var depTaskID string
					if err := linkRows.Scan(&depTaskID); err == nil && depTaskID != task.ID {
						dependency := TaskDependency{
							ID:              uuid.New().String(),
							TaskID:          task.ID,
							DependsOnTaskID: depTaskID,
							DependencyType:  "feature",
							Confidence:      overlap,
						}
						dependencies = append(dependencies, dependency)
					}
				}
			}
		}
	}

	return dependencies, nil
}

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
	row := queryRowWithTimeout(ctx, query, projectID, searchPattern)

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

	// Extract symbols from other task's file
	otherSymbols := extractSymbolsFromAST(string(otherContent), otherLang, otherTask.FilePath)
	if len(otherSymbols) == 0 {
		// No symbols found, fallback to keyword matching
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	// Check if symbols from other file are referenced in current file
	parser, err := getParser(currentLang)
	if err != nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, currentContent)
	if err != nil || tree == nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return checkCodeReferenceKeywords(codebasePath, filePath, otherTask)
	}

	// Check for symbol references in current file
	return checkSymbolReferences(rootNode, string(currentContent), currentLang, otherSymbols)
}

// extractSymbolsFromAST extracts symbols (functions, classes) from AST
func extractSymbolsFromAST(code string, language string, filePath string) map[string]bool {
	symbols := make(map[string]bool)

	parser, err := getParser(language)
	if err != nil {
		return symbols
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil || tree == nil {
		return symbols
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return symbols
	}

	traverseAST(rootNode, func(node *sitter.Node) bool {
		var symbolName string

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "identifier" || child.Type() == "field_identifier") {
						symbolName = code[child.StartByte():child.EndByte()]
						break
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "identifier" || child.Type() == "property_identifier") {
						symbolName = code[child.StartByte():child.EndByte()]
						break
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						symbolName = code[child.StartByte():child.EndByte()]
						break
					}
				}
			}
		}

		if symbolName != "" {
			symbols[strings.ToLower(symbolName)] = true
		}

		return true
	})

	return symbols
}

// checkSymbolReferences checks if symbols are referenced in AST
func checkSymbolReferences(root *sitter.Node, code string, language string, symbols map[string]bool) bool {
	found := false

	traverseAST(root, func(node *sitter.Node) bool {
		if found {
			return false // Stop traversal if already found
		}

		// Check for identifier nodes that might reference symbols
		if node.Type() == "identifier" || node.Type() == "property_identifier" || node.Type() == "field_identifier" {
			identifierName := strings.ToLower(code[node.StartByte():node.EndByte()])
			if symbols[identifierName] {
				found = true
				return false
			}
		}

		// Check for call expressions
		if node.Type() == "call_expression" || node.Type() == "call" {
			// Extract function name from call
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					funcName := extractIdentifierFromNode(child, code)
					if funcName != "" && symbols[strings.ToLower(funcName)] {
						found = true
						return false
					}
				}
			}
		}

		return true
	})

	return found
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

	_, err := execWithTimeout(ctx, query,
		dependency.ID, dependency.TaskID, dependency.DependsOnTaskID,
		dependency.DependencyType, dependency.Confidence,
	)

	return err
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

	rows, err := queryWithTimeout(ctx, query, taskID)
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
	reverseRows, err := queryWithTimeout(ctx, reverseQuery, taskID)
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

	response := &DependencyGraphResponse{
		TaskID:       taskID,
		Dependencies: dependencies,
		BlockedBy:    blockedByList,
		Blocks:       blocksList,
		HasCycle:     hasCycle,
		CyclePath:    cyclePath,
	}

	// Cache the dependencies
	SetCachedDependencies(taskID, response)

	return response, nil
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
