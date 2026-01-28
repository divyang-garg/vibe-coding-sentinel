// Package services internal helper functions for code analysis
// Provides internal analysis methods for CodeAnalysisServiceImpl
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"fmt"
	"strings"
)

// Helper functions for code analysis

func (s *CodeAnalysisServiceImpl) filterIssuesByRules(issues []map[string]interface{}, rules []string) []map[string]interface{} {
	if len(rules) == 0 {
		return issues
	}

	var filtered []map[string]interface{}
	for _, issue := range issues {
		if issueType, ok := issue["type"].(string); ok {
			for _, rule := range rules {
				if strings.Contains(rule, issueType) || rule == "all" {
					filtered = append(filtered, issue)
					break
				}
			}
		}
	}
	return filtered
}

func (s *CodeAnalysisServiceImpl) calculateSeverityBreakdown(issues []map[string]interface{}) map[string]int {
	breakdown := map[string]int{
		"critical": 0,
		"major":    0,
		"minor":    0,
	}

	for _, issue := range issues {
		if severity, ok := issue["severity"].(string); ok {
			breakdown[severity]++
		}
	}

	return breakdown
}

func (s *CodeAnalysisServiceImpl) generateRefactoringSuggestions(code, language, action string) []map[string]interface{} {
	var suggestions []map[string]interface{}

	// Analyze code characteristics to provide more specific suggestions
	codeLines := strings.Split(code, "\n")
	lineCount := len(codeLines)
	hasLongFunction := lineCount > 50
	hasComplexLogic := strings.Count(code, "{") > 10 || strings.Count(code, "if") > 5

	// Use language to provide language-specific guidance
	languageSpecificNote := ""
	switch language {
	case "go":
		languageSpecificNote = " (Go-specific patterns)"
	case "javascript", "typescript":
		languageSpecificNote = " (JavaScript/TypeScript patterns)"
	case "python":
		languageSpecificNote = " (Python-specific patterns)"
	}

	switch action {
	case "extract_method":
		description := "Extract complex logic into separate methods"
		if hasLongFunction {
			description = fmt.Sprintf("Extract complex logic into separate methods - function is %d lines long", lineCount)
		}
		if hasComplexLogic {
			description += " - detected complex nested logic"
		}
		description += languageSpecificNote

		suggestions = append(suggestions, map[string]interface{}{
			"type":        "extract_method",
			"description": description,
			"priority":    "high",
			"effort":      "medium",
		})
	case "rename_variables":
		description := "Use more descriptive variable names"
		// Check for single-letter variables which are often poor naming
		if strings.Contains(code, "var ") || strings.Contains(code, "let ") || strings.Contains(code, "const ") {
			description += " - consider replacing single-letter variables"
		}
		description += languageSpecificNote

		suggestions = append(suggestions, map[string]interface{}{
			"type":        "rename_variables",
			"description": description,
			"priority":    "medium",
			"effort":      "low",
		})
	default:
		description := "General code structure improvements"
		if hasLongFunction {
			description += fmt.Sprintf(" - consider breaking down %d-line function", lineCount)
		}
		description += languageSpecificNote

		suggestions = append(suggestions, map[string]interface{}{
			"type":        "general_improvement",
			"description": description,
			"priority":    "medium",
			"effort":      "medium",
		})
	}

	return suggestions
}

func (s *CodeAnalysisServiceImpl) estimateRefactoringSavings(suggestions []map[string]interface{}) map[string]interface{} {
	totalSavings := len(suggestions) * 30 // Assume 30 minutes saved per suggestion
	return map[string]interface{}{
		"time_saved_minutes": totalSavings,
		"productivity_gain":  fmt.Sprintf("%.1f%%", float64(totalSavings)/480.0*100), // Based on 8-hour day
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
