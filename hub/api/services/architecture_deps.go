// Package architecture_deps - Dependency analysis for architecture validation
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"fmt"
	"path/filepath"
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

// detectDependencyIssues detects dependency issues using the module graph and file metrics.
func detectDependencyIssues(files []FileContent, graph ModuleGraph) []DependencyIssue {
	var issues []DependencyIssue

	ac := GetArchitectureConfig()

	// God module: files exceeding line threshold (use cleaned path for consistency with graph)
	for _, file := range files {
		lines := strings.Split(file.Content, "\n")
		if len(lines) > ac.MaxLines {
			p := filepath.Clean(file.Path)
			issues = append(issues, DependencyIssue{
				Type:        "god_module",
				Severity:    "high",
				Files:       []string{p},
				Description: fmt.Sprintf("File %s is very large (%d lines), indicating it may be doing too much", p, len(lines)),
				Suggestion:  "Consider splitting into smaller, focused modules",
			})
		}
	}

	// Circular dependencies from graph
	cycles := findCyclesInModuleGraph(graph)
	for _, cycle := range cycles {
		issues = append(issues, DependencyIssue{
			Type:        "circular",
			Severity:    "high",
			Files:       cycle,
			Description: fmt.Sprintf("Circular dependency detected: %s", strings.Join(cycle, " -> ")),
			Suggestion:  "Break the cycle by introducing an interface or moving shared code to a separate module",
		})
	}

	// Tight coupling: high fan-out
	fanOut := fanOutByFile(graph)
	for path, n := range fanOut {
		if n > ac.MaxFanOut {
			issues = append(issues, DependencyIssue{
				Type:        "tight_coupling",
				Severity:    "medium",
				Files:       []string{path},
				Description: fmt.Sprintf("File %s has high fan-out (%d dependencies)", path, n),
				Suggestion:  "Reduce dependencies by extracting shared logic or applying dependency inversion",
			})
		}
	}

	return issues
}

func fanOutByFile(graph ModuleGraph) map[string]int {
	out := make(map[string]int)
	for _, e := range graph.Edges {
		if e.From != "" {
			out[e.From]++
		}
	}
	return out
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
