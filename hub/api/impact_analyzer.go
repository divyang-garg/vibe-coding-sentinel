// Phase 12: Impact Analysis Module
// Analyzes impact of change requests on code and tests

package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	evidence := detectBusinessRuleImplementation(rule, codebasePath)

	// Convert evidence to CodeLocation
	for _, file := range evidence.Files {
		lineNumbers := evidence.LineNumbers[file]
		if len(lineNumbers) == 0 {
			lineNumbers = []int{0} // Default if no line numbers
		}

		locations = append(locations, CodeLocation{
			FilePath:     file,
			FunctionName: extractFunctionNameFromFile(file, evidence.Functions),
			LineNumbers:  lineNumbers,
		})
	}

	// Also search for functions
	for _, funcName := range evidence.Functions {
		// Find file containing this function
		file := findFileForFunction(funcName, codebasePath)
		if file != "" {
			locations = append(locations, CodeLocation{
				FilePath:     file,
				FunctionName: funcName,
				LineNumbers:  []int{0}, // Would need AST to get exact lines
			})
		}
	}

	return locations, nil
}

// analyzeTestImpact finds tests affected by a change request
func analyzeTestImpact(ctx context.Context, changeRequest *ChangeRequest, knowledgeItemID string) ([]TestLocation, error) {
	var locations []TestLocation

	if knowledgeItemID == "" {
		return locations, nil
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
	// Simple search - would be better with AST
	// For now, return empty and let caller handle
	return ""
}
