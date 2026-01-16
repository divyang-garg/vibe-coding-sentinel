// Package architecture_deps - Dependency analysis for architecture validation
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"fmt"
	"strings"
)

func detectModuleType(path string) string {
	pathLower := strings.ToLower(path)
	if strings.Contains(pathLower, "component") {
		return "component"
	}
	if strings.Contains(pathLower, "service") {
		return "service"
	}
	if strings.Contains(pathLower, "util") || strings.Contains(pathLower, "helper") {
		return "utility"
	}
	if strings.Contains(pathLower, "test") {
		return "test"
	}
	return "module"
}

// detectDependencyIssues detects dependency issues in the codebase
func detectDependencyIssues(files []FileContent, graph ModuleGraph) []DependencyIssue {
	var issues []DependencyIssue

	// Simple detection: look for circular imports (basic pattern matching)
	// Full implementation would build actual dependency graph
	for _, file := range files {
		// Check for very large files (potential god module)
		lines := strings.Split(file.Content, "\n")
		if len(lines) > 1000 {
			issues = append(issues, DependencyIssue{
				Type:        "god_module",
				Severity:    "high",
				Files:       []string{file.Path},
				Description: fmt.Sprintf("File %s is very large (%d lines), indicating it may be doing too much", file.Path, len(lines)),
				Suggestion:  "Consider splitting into smaller, focused modules",
			})
		}
	}

	return issues
}

// generateRecommendations generates recommendations based on analysis
func generateRecommendations(oversizedFiles []FileAnalysisResult, issues []DependencyIssue) []string {
	var recommendations []string

	if len(oversizedFiles) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d oversized files. Consider refactoring to improve maintainability.", len(oversizedFiles)))
	}

	if len(issues) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d dependency issues. Review and refactor to reduce coupling.", len(issues)))
	}

	if len(oversizedFiles) == 0 && len(issues) == 0 {
		recommendations = append(recommendations, "No major architecture issues detected. Keep file sizes manageable as codebase grows.")
	}

	return recommendations
}
