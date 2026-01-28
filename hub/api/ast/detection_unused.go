// Package ast provides duplicate detection functionality
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectUnusedVariables finds unused variable declarations.
// Uses registry when a detector is registered for the language; otherwise falls back to switch.
func detectUnusedVariables(root *sitter.Node, code string, language string) []ASTFinding {
	if d := GetLanguageDetector(language); d != nil {
		return d.DetectUnused(root, code)
	}
	findings := []ASTFinding{}
	switch language {
	case "go":
		findings = detectUnusedVariablesGo(root, code)
	case "javascript", "typescript":
		findings = detectUnusedVariablesJS(root, code)
	case "python":
		findings = detectUnusedVariablesPython(root, code)
	}
	return findings
}

// detectUnusedVariablesGo finds unused variables in Go code
func detectUnusedVariablesGo(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	variableDeclarations := make(map[string]*sitter.Node)
	declarationPositions := make(map[uint32]bool) // Track byte offsets of declared variable names
	variableUsages := make(map[string]bool)

	// First pass: collect all variable declarations and their positions
	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "short_var_declaration" || node.Type() == "var_declaration" {
			// For both types, collect all identifier nodes within the declaration
			// that are not part of type annotations
			TraverseAST(node, func(declNode *sitter.Node) bool {
				if declNode.Type() == "identifier" {
					// Check if this identifier is in a type context (should be excluded)
					parent := declNode.Parent()
					if parent != nil {
						parentType := parent.Type()
						// Skip identifiers that are type identifiers or in type contexts
						if parentType == "type_identifier" || parentType == "qualified_type" {
							return true
						}
					}

					varName := safeSlice(code, declNode.StartByte(), declNode.EndByte())
					variableDeclarations[varName] = node
					declarationPositions[declNode.StartByte()] = true
				}
				return true
			})
		}
		return true
	})

	// Second pass: collect all variable usages (excluding declaration positions)
	TraverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "identifier" {
			// Skip if this identifier is at a declaration position
			if !declarationPositions[node.StartByte()] {
				varName := safeSlice(code, node.StartByte(), node.EndByte())
				variableUsages[varName] = true
			}
		}
		return true
	})

	// Check for unused variables
	for varName, node := range variableDeclarations {
		if !variableUsages[varName] {
			startLine, startCol := getLineColumn(code, int(node.StartByte()))
			endLine, endCol := getLineColumn(code, int(node.EndByte()))

			findings = append(findings, ASTFinding{
				Type:        "unused_variable",
				Severity:    "warning",
				Line:        startLine,
				Column:      startCol,
				EndLine:     endLine,
				EndColumn:   endCol,
				Message:     fmt.Sprintf("Unused variable: '%s' is declared but never used", varName),
				Code:        safeSlice(code, node.StartByte(), node.EndByte()),
				Suggestion:  fmt.Sprintf("Remove unused variable '%s' or use it in an expression", varName),
				Confidence:  0.5,      // Initial, needs validation
				AutoFixSafe: false,    // Safe default
				FixType:     "delete", // Appropriate type
				Reasoning:   "Pending codebase validation",
				Validated:   false,
			})
		}
	}

	return findings
}
