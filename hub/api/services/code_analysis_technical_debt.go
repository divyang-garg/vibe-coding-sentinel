// Package services provides technical debt estimation functions for code analysis.
//
// This file implements technical debt estimation based on code issues, duplicates,
// and orphaned code. It calculates effort estimates in hours, categorizes debt,
// and provides priority rankings for debt resolution.
//
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// estimateTechnicalDebt estimates technical debt in hours and provides detailed breakdown.
//
// This function analyzes code issues, duplicate code, and orphaned code to estimate
// the total technical debt in development hours. It categorizes debt by type,
// prioritizes issues, and calculates cost estimates.
//
// Debt estimation rules:
//   - Issues: Based on severity (critical: 4h, high: 2h, medium: 1h, low: 0.5h)
//   - Duplicates: 2 hours per duplicate function
//   - Orphaned code: 0.5 hours per orphaned item
//
// The function also calculates:
//   - Debt ratio: ratio of debt hours to estimated total development time
//   - Estimated cost: assuming $100/hour developer rate
//   - Payoff time: estimated days to resolve (assuming 8 hours/day)
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - code: Source code to analyze
//   - language: Programming language identifier
//   - issues: List of code issues with severity and type
//   - duplicates: List of duplicate code patterns
//   - orphaned: List of orphaned/unused code elements
//
// Returns:
//   - TechnicalDebtEstimate with total hours, categories, priority issues, cost, and payoff time
//   - Empty TechnicalDebtEstimate if context is cancelled
//
// Example:
//
//	debt := service.estimateTechnicalDebt(ctx, code, "go", issues, duplicates, orphaned)
//	if debt.TotalDebtHours > 40.0 {
//	    // Significant technical debt - prioritize refactoring
//	}
func (s *CodeAnalysisServiceImpl) estimateTechnicalDebt(ctx context.Context, code, language string, issues, duplicates, orphaned []interface{}) TechnicalDebtEstimate {
	// Check context cancellation
	if ctx.Err() != nil {
		return TechnicalDebtEstimate{}
	}

	// Get language-specific effort multiplier
	effortMultiplier := getLanguageEffortMultiplier(language)

	debtByCategory := make(map[string]float64)
	var priorityIssues []DebtIssue
	totalDebt := 0.0

	// Estimate debt from issues
	for _, issue := range issues {
		if issueMap, ok := issue.(map[string]interface{}); ok {
			severity := "medium"
			if s, ok := issueMap["severity"].(string); ok {
				severity = s
			}

			issueType := "general"
			if t, ok := issueMap["type"].(string); ok {
				issueType = t
			}

			line := 0
			if l, ok := issueMap["line"].(int); ok {
				line = l
			}

			message := ""
			if m, ok := issueMap["message"].(string); ok {
				message = m
			}

			// Estimate hours based on severity, adjusted for language complexity
			hours := s.estimateIssueEffort(severity, issueType) * effortMultiplier
			totalDebt += hours
			debtByCategory[issueType] += hours

			priorityIssues = append(priorityIssues, DebtIssue{
				Type:        issueType,
				Severity:    severity,
				Line:        line,
				Message:     message,
				EffortHours: hours,
				Priority:    s.determinePriority(severity, hours),
			})
		}
	}

	// Estimate debt from duplicates (adjusted for language complexity)
	duplicateHours := float64(len(duplicates)) * 2.0 * effortMultiplier // Base: 2 hours per duplicate
	totalDebt += duplicateHours
	debtByCategory["duplicates"] = duplicateHours

	// Estimate debt from orphaned code (adjusted for language complexity)
	orphanedHours := float64(len(orphaned)) * 0.5 * effortMultiplier // Base: 0.5 hours per orphaned item
	totalDebt += orphanedHours
	debtByCategory["orphaned"] = orphanedHours

	// Sort priority issues by priority and effort
	sort.Slice(priorityIssues, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		if priorityOrder[priorityIssues[i].Priority] != priorityOrder[priorityIssues[j].Priority] {
			return priorityOrder[priorityIssues[i].Priority] > priorityOrder[priorityIssues[j].Priority]
		}
		return priorityIssues[i].EffortHours > priorityIssues[j].EffortHours
	})

	// Limit to top 20 priority issues
	if len(priorityIssues) > 20 {
		priorityIssues = priorityIssues[:20]
	}

	// Calculate debt ratio (debt hours / estimated total development time)
	// Adjust per-line estimate based on language complexity
	lines := len(strings.Split(code, "\n"))
	baseHoursPerLine := 0.1 // Base estimate: 0.1 hours per line
	hoursPerLine := baseHoursPerLine * effortMultiplier
	estimatedTotalHours := float64(lines) * hoursPerLine
	debtRatio := 0.0
	if estimatedTotalHours > 0 {
		debtRatio = totalDebt / estimatedTotalHours
		if debtRatio > 1.0 {
			debtRatio = 1.0
		}
	}

	// Estimate cost (assuming $100/hour developer rate)
	estimatedCost := fmt.Sprintf("$%.2f", totalDebt*100.0)

	// Estimate payoff time (assuming 8 hours/day)
	payoffDays := totalDebt / 8.0
	payoffTime := fmt.Sprintf("%.1f days", payoffDays)

	return TechnicalDebtEstimate{
		TotalDebtHours: totalDebt,
		DebtByCategory: debtByCategory,
		PriorityIssues: priorityIssues,
		EstimatedCost:  estimatedCost,
		PayoffTime:     payoffTime,
		DebtRatio:      debtRatio,
	}
}
