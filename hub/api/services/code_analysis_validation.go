// Package services provides syntax validation and error detection functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"sentinel-hub-api/ast"
)

// validateSyntax validates code syntax using AST parsing
// Returns true if code is syntactically valid, false otherwise
func (s *CodeAnalysisServiceImpl) validateSyntax(code, language string) bool {
	if code == "" {
		return true // Empty code is considered valid
	}

	if language == "" {
		return false // Language is required for validation
	}

	// Use AST parser to validate syntax
	parser, err := ast.GetParser(language)
	if err != nil {
		// Unsupported language - cannot validate
		return false
	}

	// Attempt to parse the code
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		// Parse error means invalid syntax
		return false
	}

	if tree == nil {
		return false
	}
	defer tree.Close()

	// If parsing succeeds, syntax is valid
	rootNode := tree.RootNode()
	return rootNode != nil
}

// findSyntaxErrors finds and returns all syntax errors in the code
// Returns a slice of error messages with line numbers
func (s *CodeAnalysisServiceImpl) findSyntaxErrors(code, language string) []string {
	if code == "" || language == "" {
		return []string{}
	}

	var errors []string

	// Use AST parser to detect syntax errors
	parser, err := ast.GetParser(language)
	if err != nil {
		return []string{fmt.Sprintf("Unsupported language: %s", language)}
	}

	// Attempt to parse the code
	ctx := context.Background()
	tree, parseErr := parser.ParseCtx(ctx, nil, []byte(code))
	if parseErr != nil {
		// Extract error information
		errorMsg := parseErr.Error()

		// Try to extract line number from error message
		lineRe := regexp.MustCompile(`line\s+(\d+)`)
		matches := lineRe.FindStringSubmatch(errorMsg)

		if len(matches) > 1 {
			errors = append(errors, fmt.Sprintf("Line %s: %s", matches[1], errorMsg))
		} else {
			errors = append(errors, errorMsg)
		}
		return errors
	}

	if tree == nil {
		return []string{"Failed to parse code: tree is nil"}
	}
	defer tree.Close()

	// Check for parse errors in the tree
	rootNode := tree.RootNode()
	if rootNode == nil {
		return []string{"Failed to get root node from AST"}
	}

	// Use AST analysis to find syntax issues
	findings, _, err := ast.AnalyzeAST(code, language, []string{"brace_mismatch"})
	if err == nil {
		for _, finding := range findings {
			if finding.Type == "brace_mismatch" || finding.Type == "syntax_error" {
				errorMsg := fmt.Sprintf("Line %d: %s", finding.Line, finding.Message)
				errors = append(errors, errorMsg)
			}
		}
	}

	return errors
}

// findPotentialIssues finds potential code quality issues using AST analysis
// Returns a slice of issue messages with severity and suggestions
func (s *CodeAnalysisServiceImpl) findPotentialIssues(code, language string) []string {
	if code == "" || language == "" {
		return []string{}
	}

	var issues []string

	// Use AST analysis to find potential issues
	findings, _, err := ast.AnalyzeAST(code, language, []string{
		"unused", "unreachable", "empty_catch", "missing_await",
	})
	if err != nil {
		// If AST analysis fails, return basic suggestion
		return []string{"Consider adding error handling"}
	}

	// Convert findings to issue messages
	for _, finding := range findings {
		issueMsg := fmt.Sprintf("Line %d: %s", finding.Line, finding.Message)
		if finding.Suggestion != "" {
			issueMsg += fmt.Sprintf(" - Suggestion: %s", finding.Suggestion)
		}
		issues = append(issues, issueMsg)
	}

	// Add code smell detection
	issues = append(issues, s.detectCodeSmells(code, language)...)

	// If no issues found, return a general suggestion
	if len(issues) == 0 {
		return []string{"Consider adding error handling"}
	}

	// Limit to top 50 issues for performance
	if len(issues) > 50 {
		issues = issues[:50]
	}

	return issues
}

// detectCodeSmells detects common code smells in the code
func (s *CodeAnalysisServiceImpl) detectCodeSmells(code, language string) []string {
	var smells []string
	lines := strings.Split(code, "\n")

	// Check for long functions
	functionCount := 0
	for _, line := range lines {
		if s.isFunctionDeclaration(line, language) {
			functionCount++
		}
	}

	if len(lines) > 0 && functionCount > 0 {
		avgLinesPerFunction := len(lines) / functionCount
		if avgLinesPerFunction > 50 {
			smells = append(smells, fmt.Sprintf("Long functions detected: average %d lines per function (recommended: < 30)", avgLinesPerFunction))
		}
	}

	// Check for magic numbers
	magicNumberRe := regexp.MustCompile(`\b\d{3,}\b`)
	for i, line := range lines {
		if magicNumberRe.MatchString(line) && !s.isInStringLiteral(line) {
			smells = append(smells, fmt.Sprintf("Line %d: Magic number detected - consider using named constants", i+1))
		}
	}

	// Check for deep nesting
	maxNesting := 0
	currentNesting := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "if ") || strings.HasPrefix(trimmed, "for ") ||
			strings.HasPrefix(trimmed, "while ") || strings.HasPrefix(trimmed, "switch ") {
			currentNesting++
			if currentNesting > maxNesting {
				maxNesting = currentNesting
			}
		}
		if strings.HasPrefix(trimmed, "}") || strings.HasPrefix(trimmed, "end") {
			if currentNesting > 0 {
				currentNesting--
			}
		}
	}

	if maxNesting > 4 {
		smells = append(smells, fmt.Sprintf("Deep nesting detected: %d levels (recommended: < 4)", maxNesting))
	}

	return smells
}

// isFunctionDeclaration checks if a line is a function declaration
func (s *CodeAnalysisServiceImpl) isFunctionDeclaration(line, language string) bool {
	trimmed := strings.TrimSpace(line)
	switch language {
	case "go":
		return strings.HasPrefix(trimmed, "func ")
	case "javascript", "typescript":
		return strings.HasPrefix(trimmed, "function ") ||
			strings.Contains(trimmed, "=>") ||
			strings.Contains(trimmed, "= function")
	case "python":
		return strings.HasPrefix(trimmed, "def ")
	default:
		return false
	}
}

// isInStringLiteral checks if a number is inside a string literal
func (s *CodeAnalysisServiceImpl) isInStringLiteral(line string) bool {
	// Simple check: if line contains quotes, the number might be in a string
	return strings.Contains(line, `"`) || strings.Contains(line, "'") || strings.Contains(line, "`")
}
