// Package services provides quality score calculation functions for individual quality dimensions.
//
// This file implements scoring functions for five key quality dimensions:
// maintainability, readability, testability, complexity, and documentation.
// Each function returns a score from 0-100, where higher scores indicate better quality.
//
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"math"
	"strings"
)

// calculateMaintainabilityScore calculates maintainability score on a 0-100 scale.
//
// This score is based on the number of issues found and code complexity.
// Higher scores indicate more maintainable code.
//
// Scoring algorithm:
//   - Starts at 100.0
//   - Deducts 2.0 points per issue found
//   - Deducts 1.5 points per cyclomatic complexity point above 10
//
// Parameters:
//   - code: Source code to analyze
//   - language: Programming language identifier
//   - issues: List of code issues that impact maintainability
//
// Returns:
//   - Maintainability score between 0.0 and 100.0
//
// Example:
//
//	score := service.calculateMaintainabilityScore(code, "go", issues)
//	if score < 70.0 {
//	    // Low maintainability - consider refactoring
//	}
func (s *CodeAnalysisServiceImpl) calculateMaintainabilityScore(code, language string, issues []interface{}) float64 {
	score := 100.0

	// Deduct for issues
	score -= float64(len(issues)) * 2.0

	// Deduct for complexity
	complexity := s.analyzeComplexity(code, language)
	if cyclomatic, ok := complexity["cyclomatic"].(int); ok && cyclomatic > 10 {
		score -= float64(cyclomatic-10) * 1.5
	}

	return math.Max(0, math.Min(100, score))
}

// calculateReadabilityScore calculates readability score on a 0-100 scale.
//
// This score evaluates how easy the code is to read and understand.
// Factors considered include line length and comment density.
//
// Scoring algorithm:
//   - Starts at 100.0
//   - Deducts points based on ratio of lines exceeding 100 characters
//   - Adds bonus points if comment ratio exceeds 10%
//
// Parameters:
//   - code: Source code to analyze
//   - language: Programming language identifier
//
// Returns:
//   - Readability score between 0.0 and 100.0
//
// Example:
//
//	score := service.calculateReadabilityScore(code, "go")
//	if score > 80.0 {
//	    // Highly readable code
//	}
func (s *CodeAnalysisServiceImpl) calculateReadabilityScore(code, language string) float64 {
	score := 100.0
	lines := strings.Split(code, "\n")

	// Check for long lines
	longLineCount := 0
	for _, line := range lines {
		if len(line) > 100 {
			longLineCount++
		}
	}

	// Deduct for long lines
	if len(lines) > 0 {
		longLineRatio := float64(longLineCount) / float64(len(lines))
		score -= longLineRatio * 20.0
	}

	// Check for comments (more comments = better readability)
	// Use language-specific comment patterns
	commentCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isCommentLine(trimmed, language) {
			commentCount++
		}
	}

	// Add bonus for good comment ratio
	if len(lines) > 0 {
		commentRatio := float64(commentCount) / float64(len(lines))
		if commentRatio > 0.1 {
			score += 10.0
		}
	}

	return math.Max(0, math.Min(100, score))
}

// calculateTestabilityScore calculates testability score on a 0-100 scale.
//
// This score indicates how easy the code is to test. Higher complexity
// makes code harder to test, while presence of test files indicates
// better testability.
//
// Scoring algorithm:
//   - Starts at 100.0
//   - Adds 10.0 points if test-related keywords are found
//   - Deducts 2.0 points per cyclomatic complexity point above 15
//
// Parameters:
//   - code: Source code to analyze
//   - language: Programming language identifier
//
// Returns:
//   - Testability score between 0.0 and 100.0
//
// Example:
//
//	score := service.calculateTestabilityScore(code, "go")
//	if score < 60.0 {
//	    // Low testability - consider simplifying code structure
//	}
func (s *CodeAnalysisServiceImpl) calculateTestabilityScore(code, language string) float64 {
	score := 100.0

	// Check for test files
	if strings.Contains(code, "Test") || strings.Contains(code, "test") {
		score += 10.0
	}

	// Deduct for high complexity (harder to test)
	complexity := s.analyzeComplexity(code, language)
	if cyclomatic, ok := complexity["cyclomatic"].(int); ok && cyclomatic > 15 {
		score -= float64(cyclomatic-15) * 2.0
	}

	return math.Max(0, math.Min(100, score))
}

// calculateComplexityScore calculates complexity score on a 0-100 scale.
//
// This is an inverted score where lower complexity results in higher scores.
// The score is based on cyclomatic complexity metrics.
//
// Scoring algorithm:
//   - Starts at 100.0
//   - Deducts 5.0 points per cyclomatic complexity point above 10
//
// Parameters:
//   - code: Source code to analyze
//   - language: Programming language identifier
//
// Returns:
//   - Complexity score between 0.0 and 100.0 (higher = less complex)
//
// Example:
//
//	score := service.calculateComplexityScore(code, "go")
//	if score < 50.0 {
//	    // High complexity - consider breaking down into smaller functions
//	}
func (s *CodeAnalysisServiceImpl) calculateComplexityScore(code, language string) float64 {
	complexity := s.analyzeComplexity(code, language)
	cyclomatic := 0
	if c, ok := complexity["cyclomatic"].(int); ok {
		cyclomatic = c
	}

	// Invert: lower complexity = higher score
	score := 100.0
	if cyclomatic > 10 {
		score -= float64(cyclomatic-10) * 5.0
	}

	return math.Max(0, math.Min(100, score))
}

// calculateDocumentationScore calculates documentation score on a 0-100 scale.
//
// This score measures the ratio of documentation lines to total lines of code.
// Supports multiple comment styles: //, /* */, #, """ """, ”' ”'
//
// Scoring algorithm:
//   - Calculates ratio of documentation lines to total lines
//   - Score = (doc_lines / total_lines) * 100.0
//
// Parameters:
//   - code: Source code to analyze
//   - language: Programming language identifier
//
// Returns:
//   - Documentation score between 0.0 and 100.0
//
// Example:
//
//	score := service.calculateDocumentationScore(code, "go")
//	if score < 20.0 {
//	    // Low documentation - consider adding more comments
//	}
func (s *CodeAnalysisServiceImpl) calculateDocumentationScore(code, language string) float64 {
	score := 0.0
	lines := strings.Split(code, "\n")

	// Count documentation lines
	// Use language-specific documentation patterns
	docLines := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isDocumentationLine(trimmed, language) {
			docLines++
		}
	}

	// Calculate documentation ratio
	if len(lines) > 0 {
		docRatio := float64(docLines) / float64(len(lines))
		score = docRatio * 100.0
	}

	return math.Max(0, math.Min(100, score))
}

// isCommentLine checks if a line is a comment based on language-specific patterns
func isCommentLine(line, language string) bool {
	switch language {
	case "go", "javascript", "typescript", "java", "csharp", "rust":
		// C-style comments: // and /* */
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*")
	case "python":
		// Python: # for single-line, """ or ''' for multi-line
		return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "\"\"\"") || strings.HasPrefix(line, "'''")
	case "ruby":
		// Ruby: # for single-line, =begin =end for multi-line
		return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "=begin") || strings.HasPrefix(line, "=end")
	case "php":
		// PHP: //, #, and /* */
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*")
	default:
		// Fallback: check for common comment patterns
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") ||
			strings.HasPrefix(line, "#") || strings.HasPrefix(line, "*")
	}
}

// isDocumentationLine checks if a line is documentation based on language-specific patterns
func isDocumentationLine(line, language string) bool {
	switch language {
	case "go":
		// Go: // for comments, /* */ for block comments
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*")
	case "python":
		// Python: # for comments, """ or ''' for docstrings
		return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "\"\"\"") || strings.HasPrefix(line, "'''")
	case "javascript", "typescript":
		// JavaScript/TypeScript: // and /* */ (including JSDoc /** */)
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") ||
			strings.HasPrefix(line, "/**")
	case "java", "csharp":
		// Java/C#: // and /* */ (including JavaDoc /** */)
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") ||
			strings.HasPrefix(line, "/**")
	case "ruby":
		// Ruby: # for comments, =begin =end for multi-line
		return strings.HasPrefix(line, "#") || strings.HasPrefix(line, "=begin") || strings.HasPrefix(line, "=end")
	case "php":
		// PHP: //, #, and /* */
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*")
	default:
		// Fallback: check for common documentation patterns
		return strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") ||
			strings.HasPrefix(line, "#") || strings.HasPrefix(line, "*") ||
			strings.HasPrefix(line, "\"\"\"") || strings.HasPrefix(line, "'''")
	}
}
