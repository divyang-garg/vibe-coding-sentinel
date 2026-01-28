// Package services provides refactoring priority calculation functions for code analysis.
//
// This file implements refactoring priority ranking based on code issues, duplicates,
// and orphaned code. It helps developers identify which refactoring tasks should
// be prioritized based on impact, effort, and severity.
//
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"context"
	"fmt"
	"sort"
)

// calculateRefactoringPriority calculates and ranks refactoring priorities.
//
// This function analyzes code issues, duplicates, and orphaned code to generate
// a prioritized list of refactoring recommendations. Priorities are scored and
// sorted by importance.
//
// Priority scoring:
//   - Critical issues: Score 9.0 (highest priority)
//   - High priority issues: Score 7.0
//   - Duplicate code: Score 8.0
//   - Orphaned code: Score 5.0
//
// The function returns up to 10 top priorities, sorted by score in descending order.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - code: Source code to analyze (used for context, not directly analyzed here)
//   - language: Programming language identifier
//   - issues: List of code issues with severity information
//   - duplicates: List of duplicate code patterns
//   - orphaned: List of orphaned/unused code elements
//
// Returns:
//   - Slice of RefactoringPriority structs, sorted by score (highest first)
//   - Empty slice if context is cancelled
//   - Maximum 10 priorities returned
//
// Example:
//
//	priorities := service.calculateRefactoringPriority(ctx, code, "go", issues, duplicates, orphaned)
//	for _, priority := range priorities {
//	    if priority.Priority == "high" {
//	        // Address high priority refactoring first
//	    }
//	}
func (s *CodeAnalysisServiceImpl) calculateRefactoringPriority(ctx context.Context, code, language string, issues, duplicates, orphaned []interface{}) []RefactoringPriority {
	// Check context cancellation
	if ctx.Err() != nil {
		return []RefactoringPriority{}
	}

	// Validate code input
	if code == "" {
		return []RefactoringPriority{}
	}

	var priorities []RefactoringPriority

	// Determine language-specific effort adjustments
	effortMultiplier := getLanguageEffortMultiplier(language)

	// Add priorities for duplicates
	if len(duplicates) > 0 {
		effort := "medium"
		// Adjust effort based on language complexity
		if effortMultiplier > 1.2 {
			effort = "high"
		}
		priorities = append(priorities, RefactoringPriority{
			Type:        "duplicate_code",
			Description: fmt.Sprintf("Remove %d duplicate function(s) in %s", len(duplicates), language),
			Priority:    "high",
			Effort:      effort,
			Impact:      "high",
			Score:       8.0, // High priority score
		})
	}

	// Add priorities for orphaned code
	if len(orphaned) > 0 {
		effort := "low"
		// Adjust effort based on language complexity
		if effortMultiplier > 1.2 {
			effort = "medium"
		}
		priorities = append(priorities, RefactoringPriority{
			Type:        "orphaned_code",
			Description: fmt.Sprintf("Remove %d orphaned code item(s) in %s", len(orphaned), language),
			Priority:    "medium",
			Effort:      effort,
			Impact:      "medium",
			Score:       5.0,
		})
	}

	// Add priorities based on issue severity
	criticalCount := 0
	highCount := 0
	mediumCount := 0

	for _, issue := range issues {
		if issueMap, ok := issue.(map[string]interface{}); ok {
			if severity, ok := issueMap["severity"].(string); ok {
				switch severity {
				case "critical":
					criticalCount++
				case "high":
					highCount++
				case "medium":
					mediumCount++
				}
			}
		}
	}

	if criticalCount > 0 {
		effort := "high"
		// Adjust effort based on language complexity
		if effortMultiplier > 1.3 {
			effort = "high"
		}
		priorities = append(priorities, RefactoringPriority{
			Type:        "critical_issues",
			Description: fmt.Sprintf("Fix %d critical issue(s) in %s code", criticalCount, language),
			Priority:    "high",
			Effort:      effort,
			Impact:      "high",
			Score:       9.0,
		})
	}

	if highCount > 0 {
		effort := "medium"
		// Adjust effort based on language complexity
		if effortMultiplier > 1.2 {
			effort = "high"
		}
		priorities = append(priorities, RefactoringPriority{
			Type:        "high_priority_issues",
			Description: fmt.Sprintf("Fix %d high priority issue(s) in %s code", highCount, language),
			Priority:    "high",
			Effort:      effort,
			Impact:      "high",
			Score:       7.0,
		})
	}

	// Sort by score (descending)
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i].Score > priorities[j].Score
	})

	// Limit to top 10 priorities
	if len(priorities) > 10 {
		priorities = priorities[:10]
	}

	return priorities
}

// getLanguageEffortMultiplier returns an effort multiplier based on language complexity.
// More complex languages may require more effort for refactoring tasks.
func getLanguageEffortMultiplier(language string) float64 {
	switch language {
	case "go":
		// Go is relatively simple with good tooling
		return 1.0
	case "python":
		// Python is simple but dynamic typing can add complexity
		return 1.1
	case "javascript", "typescript":
		// JavaScript/TypeScript can be complex due to ecosystem and dynamic nature
		return 1.2
	case "java", "csharp":
		// Java/C# have more ceremony but good tooling
		return 1.15
	case "rust":
		// Rust has complex ownership rules
		return 1.3
	case "cpp", "c":
		// C/C++ are complex languages
		return 1.4
	default:
		// Default multiplier for unknown languages
		return 1.1
	}
}
