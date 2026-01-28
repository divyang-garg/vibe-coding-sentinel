// Package ast provides duplicate detection functionality
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func detectBraceMismatch(tree *sitter.Tree, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	if tree == nil {
		return findings
	}

	rootNode := tree.RootNode()
	if rootNode == nil {
		return findings
	}

	// Get language-specific delimiter patterns
	delimiterPatterns := getLanguageDelimiterPatterns(language)

	// Tree-sitter reports parse errors as ERROR nodes
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "ERROR" || node.IsError() || node.HasError() {
			// This is a parse error - likely brace/bracket mismatch
			startLine, startCol := getLineColumn(code, int(node.StartByte()))
			endLine, endCol := getLineColumn(code, int(node.EndByte()))

			// Determine type of mismatch based on surrounding code and language patterns
			errorCode := safeSlice(code, node.StartByte(), node.EndByte())
			mismatchType := detectMismatchType(errorCode, delimiterPatterns)

			// Generate language-specific suggestion
			suggestion := getLanguageSpecificSuggestion(mismatchType, language)

			findings = append(findings, ASTFinding{
				Type:       "brace_mismatch",
				Severity:   "error",
				Line:       startLine,
				Column:     startCol,
				EndLine:    endLine,
				EndColumn:  endCol,
				Message:    fmt.Sprintf("Parse error detected in %s code - likely mismatched %s", language, mismatchType),
				Code:       errorCode,
				Suggestion: suggestion,
			})
		}

		return true
	})

	return findings
}

// getLanguageDelimiterPatterns returns language-specific delimiter patterns for mismatch detection
func getLanguageDelimiterPatterns(language string) map[string][]string {
	switch language {
	case "go":
		return map[string][]string{
			"brace":      {"{", "}"},
			"bracket":    {"[", "]"},
			"parenthesis": {"(", ")"},
		}
	case "javascript", "typescript":
		return map[string][]string{
			"brace":      {"{", "}"},
			"bracket":    {"[", "]"},
			"parenthesis": {"(", ")"},
			"template":   {"`", "`"},
		}
	case "python":
		return map[string][]string{
			"bracket":    {"[", "]"},
			"parenthesis": {"(", ")"},
			"brace":      {"{", "}"},
		}
	case "java":
		return map[string][]string{
			"brace":      {"{", "}"},
			"bracket":    {"[", "]"},
			"parenthesis": {"(", ")"},
		}
	default:
		// Generic patterns for unknown languages
		return map[string][]string{
			"brace":      {"{", "}"},
			"bracket":    {"[", "]"},
			"parenthesis": {"(", ")"},
		}
	}
}

// detectMismatchType determines the type of delimiter mismatch based on error code and language patterns
func detectMismatchType(errorCode string, patterns map[string][]string) string {
	// Check each delimiter type in order of priority
	for mismatchType, delimiters := range patterns {
		for _, delim := range delimiters {
			if strings.Contains(errorCode, delim) {
				return mismatchType
			}
		}
	}

	// Fallback to generic detection
	if strings.Contains(errorCode, "[") || strings.Contains(errorCode, "]") {
		return "bracket"
	} else if strings.Contains(errorCode, "(") || strings.Contains(errorCode, ")") {
		return "parenthesis"
	}
	return "brace"
}

// getLanguageSpecificSuggestion provides language-specific suggestions for fixing delimiter mismatches
func getLanguageSpecificSuggestion(mismatchType, language string) string {
	baseSuggestion := fmt.Sprintf("Check for mismatched %ss in the code around this location", mismatchType)

	switch language {
	case "go":
		switch mismatchType {
		case "brace":
			return baseSuggestion + ". In Go, ensure all '{' have matching '}' and check function/method declarations."
		case "bracket":
			return baseSuggestion + ". In Go, check array/slice declarations and indexing operations."
		case "parenthesis":
			return baseSuggestion + ". In Go, verify function calls, type assertions, and control structures."
		}
	case "javascript", "typescript":
		switch mismatchType {
		case "brace":
			return baseSuggestion + ". In JavaScript/TypeScript, check object literals, destructuring, and block statements."
		case "bracket":
			return baseSuggestion + ". In JavaScript/TypeScript, verify array literals, destructuring, and property access."
		case "parenthesis":
			return baseSuggestion + ". In JavaScript/TypeScript, check function calls, arrow functions, and expressions."
		case "template":
			return baseSuggestion + ". In JavaScript/TypeScript, verify template literals (backticks) are properly closed."
		}
	case "python":
		switch mismatchType {
		case "bracket":
			return baseSuggestion + ". In Python, check list/dict comprehensions and indexing operations."
		case "parenthesis":
			return baseSuggestion + ". In Python, verify function calls, tuples, and generator expressions."
		case "brace":
			return baseSuggestion + ". In Python, check dictionary literals and set comprehensions."
		}
	case "java":
		switch mismatchType {
		case "brace":
			return baseSuggestion + ". In Java, ensure all '{' have matching '}' in class/method/block declarations."
		case "bracket":
			return baseSuggestion + ". In Java, check array declarations and generic type parameters."
		case "parenthesis":
			return baseSuggestion + ". In Java, verify method calls, constructor invocations, and control structures."
		}
	}

	return baseSuggestion
}
