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

	// Tree-sitter reports parse errors as ERROR nodes
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		if node.Type() == "ERROR" || node.IsError() || node.HasError() {
			// This is a parse error - likely brace/bracket mismatch
			startLine, startCol := getLineColumn(code, int(node.StartByte()))
			endLine, endCol := getLineColumn(code, int(node.EndByte()))

			// Determine type of mismatch based on surrounding code
			errorCode := safeSlice(code, node.StartByte(), node.EndByte())
			mismatchType := "brace"
			if strings.Contains(errorCode, "[") || strings.Contains(errorCode, "]") {
				mismatchType = "bracket"
			} else if strings.Contains(errorCode, "(") || strings.Contains(errorCode, ")") {
				mismatchType = "parenthesis"
			}

			findings = append(findings, ASTFinding{
				Type:       "brace_mismatch",
				Severity:   "error",
				Line:       startLine,
				Column:     startCol,
				EndLine:    endLine,
				EndColumn:  endCol,
				Message:    fmt.Sprintf("Parse error detected - likely mismatched %s", mismatchType),
				Code:       errorCode,
				Suggestion: fmt.Sprintf("Check for mismatched %ss in the code around this location", mismatchType),
			})
		}

		return true
	})

	return findings
}
