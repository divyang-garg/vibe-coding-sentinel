// Package services provides quality metrics calculation functions for code analysis.
//
// This file implements comprehensive quality metrics calculation including overall
// quality scores, maintainability index calculation, and aggregation of various
// code quality dimensions. All functions use context for cancellation support
// and follow CODING_STANDARDS.md requirements.
//
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"context"
	"math"
)

// calculateQualityMetrics calculates comprehensive quality metrics for code analysis.
//
// This function aggregates multiple quality dimensions (maintainability, readability,
// testability, complexity, documentation) into a single QualityMetrics structure.
// It processes issues, duplicates, and orphaned code to provide a holistic view
// of code quality.
//
// The overall score is calculated as a weighted average:
//   - Maintainability: 30%
//   - Readability: 20%
//   - Testability: 20%
//   - Complexity: 20%
//   - Documentation: 10%
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - code: Source code to analyze
//   - language: Programming language identifier
//   - issues: List of code issues found during analysis
//   - duplicates: List of duplicate code patterns
//   - orphaned: List of orphaned/unused code elements
//
// Returns:
//   - QualityMetrics struct with comprehensive quality scores and breakdowns
//   - Empty QualityMetrics if context is cancelled
//
// Example:
//
//	metrics := service.calculateQualityMetrics(ctx, code, "go", issues, duplicates, orphaned)
//	if metrics.OverallScore > 80.0 {
//	    // High quality code
//	}
func (s *CodeAnalysisServiceImpl) calculateQualityMetrics(ctx context.Context, code, language string, issues, duplicates, orphaned []interface{}) QualityMetrics {
	// Check context cancellation
	if ctx.Err() != nil {
		return QualityMetrics{}
	}

	// Calculate base scores
	maintainability := s.calculateMaintainabilityScore(code, language, issues)
	readability := s.calculateReadabilityScore(code, language)
	testability := s.calculateTestabilityScore(code, language)
	complexity := s.calculateComplexityScore(code, language)
	documentation := s.calculateDocumentationScore(code, language)

	// Count issues by severity
	severityBreakdown := make(map[string]int)
	issueCount := 0
	for _, issue := range issues {
		if issueMap, ok := issue.(map[string]interface{}); ok {
			if severity, ok := issueMap["severity"].(string); ok {
				severityBreakdown[severity]++
				issueCount++
			}
		}
	}

	// Add duplicates and orphaned to counts
	issueCount += len(duplicates) + len(orphaned)

	// Calculate category scores
	categoryScores := map[string]float64{
		"maintainability": maintainability,
		"readability":     readability,
		"testability":     testability,
		"complexity":      complexity,
		"documentation":   documentation,
	}

	// Calculate overall score (weighted average)
	overallScore := (maintainability*0.3 + readability*0.2 + testability*0.2 + complexity*0.2 + documentation*0.1)

	return QualityMetrics{
		OverallScore:      overallScore,
		Maintainability:   maintainability,
		Readability:       readability,
		Testability:       testability,
		Complexity:        complexity,
		Documentation:     documentation,
		IssueCount:        issueCount,
		SeverityBreakdown: severityBreakdown,
		CategoryScores:    categoryScores,
	}
}

// calculateMaintainabilityIndex calculates maintainability index (MI) on a 0-100 scale.
//
// The maintainability index is a composite metric based on Halstead complexity measures.
// It provides an indication of how maintainable the code is, with higher values
// indicating better maintainability.
//
// The formula used is a simplified version of the standard MI calculation:
//
//	MI = 171 - 5.2 * ln(avg_complexity) - 0.23 * ln(avg_lines) - 16.2 * ln(avg_functions)
//
// Where:
//   - avg_complexity: Average cyclomatic complexity per function
//   - avg_lines: Average lines of code per function
//   - avg_functions: Total number of functions
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - code: Source code to analyze
//   - language: Programming language identifier
//
// Returns:
//   - Maintainability index value between 0.0 and 100.0
//   - 0.0 if code or language is empty, or if context is cancelled
//
// Example:
//
//	mi := service.calculateMaintainabilityIndex(ctx, code, "go")
//	if mi < 50.0 {
//	    // Low maintainability - refactoring recommended
//	}
func (s *CodeAnalysisServiceImpl) calculateMaintainabilityIndex(ctx context.Context, code, language string) float64 {
	if code == "" || language == "" {
		return 0.0
	}

	// Check context cancellation
	if ctx.Err() != nil {
		return 0.0
	}

	// Get complexity metrics
	complexity := s.analyzeComplexity(code, language)
	cyclomatic := 0
	if c, ok := complexity["cyclomatic"].(int); ok {
		cyclomatic = c
	}

	// Get function count
	functionCount := 0
	if fc, ok := complexity["functions"].(int); ok {
		functionCount = fc
	}

	// Get lines count
	lines := 0
	if l, ok := complexity["lines"].(int); ok {
		lines = l
	}

	// Calculate maintainability index using Halstead complexity
	// Simplified formula: MI = 171 - 5.2 * ln(avg_complexity) - 0.23 * ln(avg_lines) - 16.2 * ln(avg_functions)
	avgComplexity := float64(cyclomatic)
	if functionCount > 0 {
		avgComplexity = float64(cyclomatic) / float64(functionCount)
	}

	avgLines := float64(lines)
	if functionCount > 0 {
		avgLines = float64(lines) / float64(functionCount)
	}

	// Calculate MI (simplified version)
	mi := 171.0 - 5.2*math.Log(math.Max(avgComplexity, 1.0)) - 0.23*math.Log(math.Max(avgLines, 1.0)) - 16.2*math.Log(math.Max(float64(functionCount), 1.0))

	// Normalize to 0-100 scale
	mi = math.Max(0, math.Min(100, mi))

	return mi
}
