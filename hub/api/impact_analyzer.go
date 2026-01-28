// Phase 12: Impact Analysis Module
// Analyzes impact of change requests on code and tests

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sentinel-hub-api/ast"
)

// ImpactAnalysis represents the impact analysis result
type ImpactAnalysis struct {
	AffectedCode    []CodeLocation `json:"affected_code"`
	AffectedTests   []TestLocation `json:"affected_tests"`
	EstimatedEffort string         `json:"estimated_effort"`
}

// CodeLocation represents a code location affected by a change
type CodeLocation struct {
	FilePath     string `json:"file_path"`
	FunctionName string `json:"function_name"`
	LineNumbers  []int  `json:"line_numbers"`
}

// TestLocation represents a test location affected by a change
type TestLocation struct {
	FilePath string `json:"file_path"`
	TestName string `json:"test_name"`
}

// analyzeCodeImpact finds code affected by a change request
func analyzeCodeImpact(ctx context.Context, changeRequest *ChangeRequest, projectID string, codebasePath string) ([]CodeLocation, error) {
	var locations []CodeLocation

	// Check for context cancellation
	if ctx.Err() != nil {
		return locations, ctx.Err()
	}

	// Extract business rule from change request
	var rule KnowledgeItem
	if changeRequest.CurrentState != nil {
		// Try to reconstruct knowledge item from current state
		if title, ok := changeRequest.CurrentState["title"].(string); ok {
			rule.Title = title
		}
		if content, ok := changeRequest.CurrentState["content"].(string); ok {
			rule.Content = content
		}
	}
	if changeRequest.ProposedState != nil {
		// Use proposed state if available
		if title, ok := changeRequest.ProposedState["title"].(string); ok {
			rule.Title = title
		}
		if content, ok := changeRequest.ProposedState["content"].(string); ok {
			rule.Content = content
		}
	}

	// Use Phase 11 business rule detection
	// Note: projectID could be used for project-specific rule detection in the future
	evidence := detectBusinessRuleImplementation(rule, codebasePath)

	// Log analysis start with context
	LogInfo(ctx, "Analyzing code impact for project %s", projectID)

	// Convert evidence to CodeLocation
	for _, file := range evidence.Files {
		// Get line numbers for this file from map
		var lineNumbers []int
		if lines, ok := evidence.LineNumbers[file]; ok {
			lineNumbers = lines
		} else if len(evidence.Functions) > 0 {
			// Fallback: use function line numbers
			for _, funcName := range evidence.Functions {
				if lines, ok := evidence.LineNumbers[funcName]; ok {
					lineNumbers = append(lineNumbers, lines...)
				}
			}
		}
		if len(lineNumbers) == 0 {
			lineNumbers = []int{0} // Default if no line numbers
		}

		locations = append(locations, CodeLocation{
			FilePath:     file,
			FunctionName: extractFunctionNameFromFile(file, evidence.Functions),
			LineNumbers:  lineNumbers,
		})
	}

	// Also search for functions using AST
	for _, funcName := range evidence.Functions {
		// Find file containing this function using AST
		file := findFileForFunction(funcName, codebasePath)
		if file != "" {
			// Use AST to get exact line numbers for the function
			lineNumbers := getFunctionLineNumbers(file, funcName, codebasePath)
			if len(lineNumbers) == 0 {
				lineNumbers = []int{0} // Fallback if AST extraction fails
			}
			locations = append(locations, CodeLocation{
				FilePath:     file,
				FunctionName: funcName,
				LineNumbers:  lineNumbers,
			})
		}
	}

	return locations, nil
}

// analyzeTestImpact finds tests affected by a change request
func analyzeTestImpact(ctx context.Context, changeRequest *ChangeRequest, knowledgeItemID string) ([]TestLocation, error) {
	var locations []TestLocation

	// Check for context cancellation
	if ctx.Err() != nil {
		return locations, ctx.Err()
	}

	if knowledgeItemID == "" {
		// Try to get knowledge item ID from change request if not provided
		if changeRequest != nil && changeRequest.KnowledgeItemID != nil && *changeRequest.KnowledgeItemID != "" {
			knowledgeItemID = *changeRequest.KnowledgeItemID
		} else {
			return locations, nil
		}
	}

	// Use Phase 10 test coverage tracker
	query := `
		SELECT DISTINCT tc.test_file_path, tc.test_function_name
		FROM test_coverage tc
		WHERE tc.knowledge_item_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, getQueryTimeout())
	defer cancel()

	rows, err := queryWithTimeout(ctx, query, knowledgeItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query test coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filePath sql.NullString
		var testName sql.NullString

		if err := rows.Scan(&filePath, &testName); err != nil {
			LogWarn(ctx, "Error scanning test coverage: %v", err)
			continue
		}

		if filePath.Valid {
			locations = append(locations, TestLocation{
				FilePath: filePath.String,
				TestName: testName.String,
			})
		}
	}

	return locations, nil
}

// estimateEffort calculates estimated effort based on impact
func estimateEffort(impact *ImpactAnalysis) string {
	files := len(impact.AffectedCode)
	functions := 0
	for _, loc := range impact.AffectedCode {
		if loc.FunctionName != "" {
			functions++
		}
	}
	testFiles := len(impact.AffectedTests)

	// Heuristic: (files * 0.5) + (functions * 0.25) + (testFiles * 0.3) hours
	hours := float64(files)*0.5 + float64(functions)*0.25 + float64(testFiles)*0.3

	if hours < 1 {
		return fmt.Sprintf("%.1f hour", hours)
	} else if hours < 8 {
		return fmt.Sprintf("%.1f hours", hours)
	} else {
		days := hours / 8
		return fmt.Sprintf("%.1f day(s)", days)
	}
}

// storeImpactAnalysis stores impact analysis results in the database
func storeImpactAnalysis(ctx context.Context, changeRequestID string, impact *ImpactAnalysis) error {
	impactJSON, err := marshalJSONB(impact)
	if err != nil {
		LogError(ctx, "Failed to marshal impact analysis for change request %s: %v", changeRequestID, err)
		return fmt.Errorf("failed to marshal impact analysis: %w", err)
	}
	query := `UPDATE change_requests SET impact_analysis = $1 WHERE id = $2`
	_, err = execWithTimeout(ctx, query, impactJSON, changeRequestID)
	if err != nil {
		LogError(ctx, "Failed to store impact analysis for change request %s: %v", changeRequestID, err)
		return fmt.Errorf("failed to store impact analysis for change request %s: %w", changeRequestID, err)
	}
	LogInfo(ctx, "Successfully stored impact analysis for change request %s", changeRequestID)
	return nil
}

// analyzeImpact performs complete impact analysis
func analyzeImpact(ctx context.Context, changeRequestID string, projectID string, codebasePath string) (*ImpactAnalysis, error) {
	// Validate inputs
	if err := ValidateUUID(changeRequestID); err != nil {
		return nil, fmt.Errorf("invalid change request ID: %w", err)
	}
	if err := ValidateUUID(projectID); err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}
	if err := ValidateDirectory(codebasePath); err != nil {
		return nil, fmt.Errorf("invalid codebase path: %w", err)
	}

	// Load change request
	changeRequest, err := getChangeRequest(ctx, changeRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to load change request %s: %w", changeRequestID, err)
	}
	if changeRequest == nil {
		return nil, fmt.Errorf("change request %s not found", changeRequestID)
	}

	// Analyze code impact
	codeLocations, err := analyzeCodeImpact(ctx, changeRequest, projectID, codebasePath)
	if err != nil {
		LogWarn(ctx, "Error analyzing code impact: %v", err)
		codeLocations = []CodeLocation{} // Continue with empty
	}

	// Analyze test impact
	knowledgeItemID := ""
	if changeRequest.KnowledgeItemID != nil {
		knowledgeItemID = *changeRequest.KnowledgeItemID
	}
	testLocations, err := analyzeTestImpact(ctx, changeRequest, knowledgeItemID)
	if err != nil {
		LogWarn(ctx, "Error analyzing test impact: %v", err)
		testLocations = []TestLocation{} // Continue with empty
	}

	impact := &ImpactAnalysis{
		AffectedCode:  codeLocations,
		AffectedTests: testLocations,
	}

	// Estimate effort
	impact.EstimatedEffort = estimateEffort(impact)

	return impact, nil
}

// Helper functions

func extractFunctionNameFromFile(filePath string, functions []string) string {
	// Try to find matching function for this file
	for _, funcName := range functions {
		if strings.Contains(filePath, funcName) || strings.Contains(funcName, filePath) {
			return funcName
		}
	}
	if len(functions) > 0 {
		return functions[0]
	}
	return ""
}

func findFileForFunction(funcName string, codebasePath string) string {
	// Validate inputs
	if funcName == "" || codebasePath == "" {
		return ""
	}

	// Use AST to search through codebase for the function
	supportedExts := map[string]string{
		".go":  "go",
		".js":  "javascript",
		".jsx": "javascript",
		".ts":  "typescript",
		".tsx": "typescript",
		".py":  "python",
	}

	var foundFile string
	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			// Skip common directories that shouldn't be scanned
			dirName := info.Name()
			if dirName == "vendor" || dirName == "node_modules" || dirName == ".git" ||
				dirName == ".idea" || dirName == ".vscode" || dirName == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has supported extension
		ext := strings.ToLower(filepath.Ext(path))
		language, isSupported := supportedExts[ext]
		if !isSupported {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		// Use AST to extract functions from this file
		functions, err := ast.ExtractFunctions(string(content), language, funcName)
		if err != nil {
			return nil // Skip files with parsing errors
		}

		// Check if the function we're looking for is in this file
		for _, fn := range functions {
			if fn.Name == funcName {
				foundFile = path
				return filepath.SkipAll // Found it, stop searching
			}
		}

		return nil
	})

	if err != nil {
		return "" // Return empty on walk error
	}

	return foundFile
}

// getFunctionLineNumbers uses AST to get exact line numbers for a function
func getFunctionLineNumbers(filePath string, funcName string, codebasePath string) []int {
	// Validate that the file is within the codebase path for security
	if codebasePath != "" {
		// Ensure filePath is within codebasePath to prevent path traversal
		relPath, err := filepath.Rel(codebasePath, filePath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			// File is outside codebase path - security check failed
			return []int{}
		}
	}

	// Determine language from file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	language := ""
	switch ext {
	case ".go":
		language = "go"
	case ".js", ".jsx":
		language = "javascript"
	case ".ts", ".tsx":
		language = "typescript"
	case ".py":
		language = "python"
	default:
		return []int{} // Unsupported language
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []int{}
	}

	// Use AST to extract the function and get its line number
	functions, err := ast.ExtractFunctions(string(content), language, funcName)
	if err != nil {
		return []int{}
	}

	// Find exact match and return line numbers
	for _, fn := range functions {
		if fn.Name == funcName {
			// Return line range for the function
			if fn.EndLine > fn.Line {
				lines := make([]int, 0, fn.EndLine-fn.Line+1)
				for i := fn.Line; i <= fn.EndLine; i++ {
					lines = append(lines, i)
				}
				return lines
			}
			return []int{fn.Line}
		}
	}

	return []int{}
}
