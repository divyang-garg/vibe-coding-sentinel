// Package ast provides async/await detection functionality
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectEmptyCatchBlocks finds empty catch blocks
func detectEmptyCatchBlocks(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	TraverseAST(root, func(node *sitter.Node) bool {
		var catchBody *sitter.Node
		var isCatch bool

		switch language {
		case "javascript", "typescript":
			if node.Type() == "catch_clause" {
				isCatch = true
				// Find the catch block body
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "statement_block" {
						catchBody = child
						break
					}
				}
			}
		case "python":
			if node.Type() == "except_clause" {
				isCatch = true
				// Find the except block body
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && (child.Type() == "block" || child.Type() == "suite") {
						catchBody = child
						break
					}
				}
			}
		}

		if isCatch && catchBody != nil {
			// Check if body is empty (no statements or only comments/pass)
			hasStatements := false
			for i := 0; i < int(catchBody.ChildCount()); i++ {
				child := catchBody.Child(i)
				if child != nil {
					nodeType := child.Type()
					// Ignore comments and whitespace
					if nodeType != "comment" && nodeType != "line_comment" && nodeType != "block_comment" {
						// Check if it's a meaningful statement
						childCode := strings.TrimSpace(safeSlice(code, child.StartByte(), child.EndByte()))
						// For Python, "pass" is a no-op statement - treat as empty
						if language == "python" && (childCode == "pass" || strings.HasSuffix(childCode, "pass")) {
							// pass statement - don't count as meaningful statement
							continue
						}
						if childCode != "" && childCode != "{" && childCode != "}" {
							hasStatements = true
							break
						}
					}
				}
			}

			if !hasStatements {
				startLine, startCol := getLineColumn(code, int(catchBody.StartByte()))
				endLine, endCol := getLineColumn(code, int(catchBody.EndByte()))

				findings = append(findings, ASTFinding{
					Type:        "empty_catch",
					Severity:    "warning",
					Line:        startLine,
					Column:      startCol,
					EndLine:     endLine,
					EndColumn:   endCol,
					Message:     "Empty catch/except block detected - errors are silently ignored",
					Code:        safeSlice(code, catchBody.StartByte(), catchBody.EndByte()),
					Suggestion:  "Add error handling logic or logging to the catch block",
					Confidence:  0.5,        // Initial, needs validation
					AutoFixSafe: false,      // Safe default
					FixType:     "refactor", // Appropriate type
					Reasoning:   "Pending codebase validation",
					Validated:   false,
				})
			}
		}

		return true
	})

	return findings
}

// detectMissingAwait finds missing await keywords in async functions.
// Uses registry when a detector is registered; otherwise falls back to language check.
func detectMissingAwait(root *sitter.Node, code string, language string) []ASTFinding {
	if d := GetLanguageDetector(language); d != nil {
		return d.DetectAsync(root, code)
	}
	findings := []ASTFinding{}
	// Only relevant for JavaScript/TypeScript when no detector registered
	if language != "javascript" && language != "typescript" {
		return findings
	}
	return detectMissingAwaitJS(root, code)
}

// detectMissingAwaitJS finds missing await keywords in JS/TS async functions
func detectMissingAwaitJS(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	asyncFunctions := make(map[*sitter.Node]*sitter.Node)

	TraverseAST(root, func(node *sitter.Node) bool {
		isAsyncFunc := false
		var bodyNode *sitter.Node

		if node.Type() == "function_declaration" || node.Type() == "arrow_function" || node.Type() == "function_expression" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "async" {
						isAsyncFunc = true
					}
					if child.Type() == "statement_block" || child.Type() == "expression_statement" {
						bodyNode = child
					}
				}
			}
			if isAsyncFunc && bodyNode != nil {
				asyncFunctions[node] = bodyNode
			}
		}
		return true
	})

	for _, bodyNode := range asyncFunctions {
		TraverseAST(bodyNode, func(node *sitter.Node) bool {
			if node.Type() == "call_expression" {
				parent := node.Parent()
				isAwaited := parent != nil && parent.Type() == "await_expression"
				if !isAwaited {
					callCode := safeSlice(code, node.StartByte(), node.EndByte())
					callCodeLower := strings.ToLower(callCode)
					likelyAsync := strings.Contains(callCodeLower, "fetch") ||
						strings.Contains(callCodeLower, ".then") ||
						strings.Contains(callCodeLower, ".catch") ||
						strings.Contains(callCodeLower, "promise")
					if likelyAsync {
						startLine, startCol := getLineColumn(code, int(node.StartByte()))
						endLine, endCol := getLineColumn(code, int(node.EndByte()))
						findings = append(findings, ASTFinding{
							Type:       "missing_await",
							Severity:   "error",
							Line:       startLine,
							Column:     startCol,
							EndLine:    endLine,
							EndColumn:  endCol,
							Message:    "Potential missing await keyword in async function",
							Code:       callCode,
							Suggestion: "Add 'await' keyword before the async call",
						})
					}
				}
			}
			return true
		})
	}
	return findings
}
