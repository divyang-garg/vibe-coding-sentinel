// Package services provides code quality and vibe analysis functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"regexp"
	"strings"

	"sentinel-hub-api/ast"
)

// identifyVibeIssues identifies code quality "vibe" issues using AST analysis
// Returns issues related to maintainability, readability, and technical debt
func (s *CodeAnalysisServiceImpl) identifyVibeIssues(code, language string) []interface{} {
	if code == "" || language == "" {
		return []interface{}{}
	}

	var issues []interface{}

	// Use AST analysis to find vibe-related issues
	findings, _, err := ast.AnalyzeAST(code, language, []string{
		"duplicates", "unused", "unreachable", "empty_catch", "missing_await",
	})
	if err != nil {
		return []interface{}{}
	}

	// Convert findings to vibe issues
	for _, finding := range findings {
		issue := map[string]interface{}{
			"type":       finding.Type,
			"severity":   finding.Severity,
			"line":       finding.Line,
			"message":    finding.Message,
			"confidence": finding.Confidence,
		}
		if finding.Suggestion != "" {
			issue["suggestion"] = finding.Suggestion
		}
		issues = append(issues, issue)
	}

	// Limit to top 30 issues for performance
	if len(issues) > 30 {
		issues = issues[:30]
	}

	return issues
}

// findDuplicateFunctions finds duplicate or similar functions using AST analysis
// Returns duplicate function groups with similarity scores
func (s *CodeAnalysisServiceImpl) findDuplicateFunctions(code, language string) []interface{} {
	if code == "" || language == "" {
		return []interface{}{}
	}

	// Use AST analysis to find duplicate functions
	findings, _, err := ast.AnalyzeAST(code, language, []string{"duplicates"})
	if err != nil {
		return []interface{}{}
	}

	var duplicates []interface{}
	duplicateMap := make(map[string][]ast.ASTFinding)

	// Group duplicate findings by function name or code similarity
	for _, finding := range findings {
		if finding.Type == "duplicate_function" {
			// Extract function name from message or code
			funcName := s.extractFunctionNameFromFinding(finding)
			duplicateMap[funcName] = append(duplicateMap[funcName], finding)
		}
	}

	// Convert grouped findings to duplicate function structures
	for funcName, findings := range duplicateMap {
		if len(findings) > 1 {
			lines := make([]int, 0, len(findings))
			for _, f := range findings {
				lines = append(lines, f.Line)
			}

			duplicate := map[string]interface{}{
				"functions":  []string{funcName},
				"similarity": 0.85, // Default similarity, could be calculated more precisely
				"lines":     lines,
				"suggestion": "Consider extracting common logic into a shared function",
			}
			duplicates = append(duplicates, duplicate)
		}
	}

	// Limit to top 20 duplicates
	if len(duplicates) > 20 {
		duplicates = duplicates[:20]
	}

	return duplicates
}

// extractFunctionNameFromFinding extracts function name from AST finding
func (s *CodeAnalysisServiceImpl) extractFunctionNameFromFinding(finding ast.ASTFinding) string {
	// Try to extract from code snippet
	if finding.Code != "" {
		// Look for function declaration patterns
		funcRe := regexp.MustCompile(`(?:func|function|def)\s+(\w+)`)
		matches := funcRe.FindStringSubmatch(finding.Code)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Fallback to message
	if strings.Contains(finding.Message, "function") {
		// Try to extract function name from message
		parts := strings.Fields(finding.Message)
		for i, part := range parts {
			if part == "function" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}

	return "unknown"
}

// findOrphanedCode finds orphaned (unused) code using AST analysis
// Returns unused functions, variables, imports, and dead code
func (s *CodeAnalysisServiceImpl) findOrphanedCode(code, language string) []interface{} {
	if code == "" || language == "" {
		return []interface{}{}
	}

	// Use AST analysis to find orphaned code
	findings, _, err := ast.AnalyzeAST(code, language, []string{"orphaned", "unused"})
	if err != nil {
		return []interface{}{}
	}

	var orphaned []interface{}

	// Extract functions to check if they're exported (potentially used elsewhere)
	functions, _ := ast.ExtractFunctions(code, language, "")

	// Create a map of exported functions (should not be flagged as orphaned)
	exportedMap := make(map[string]bool)
	for _, fn := range functions {
		if fn.Visibility == "exported" || fn.Visibility == "public" {
			exportedMap[fn.Name] = true
		}
	}

	// Convert findings to orphaned code structures
	for _, finding := range findings {
		if finding.Type == "unused_function" || finding.Type == "unused_variable" ||
			finding.Type == "orphaned" || finding.Type == "dead_code" {

			// Skip exported functions (may be used from other files)
			if finding.Type == "unused_function" {
				funcName := s.extractFunctionNameFromFinding(finding)
				if exportedMap[funcName] {
					continue // Skip exported functions
				}
			}

			orphanedItem := map[string]interface{}{
				"type":          finding.Type,
				"name":          s.extractFunctionNameFromFinding(finding),
				"line":          finding.Line,
				"message":       finding.Message,
				"safe_to_remove": finding.Type != "unused_function" || !exportedMap[s.extractFunctionNameFromFinding(finding)],
			}
			orphaned = append(orphaned, orphanedItem)
		}
	}

	// Limit to top 50 orphaned items
	if len(orphaned) > 50 {
		orphaned = orphaned[:50]
	}

	return orphaned
}
