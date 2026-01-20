// Package ast provides orphaned code detection
// Complies with CODING_STANDARDS.md: Detection modules max 250 lines
package ast

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// detectOrphanedCode finds code that is never executed
func detectOrphanedCode(root *sitter.Node, code string, language string) []ASTFinding {
	findings := []ASTFinding{}

	// Basic implementation: look for functions that are defined but never called
	// This is a simplified version - full implementation would require call graph analysis

	switch language {
	case "go":
		findings = detectOrphanedCodeGo(root, code)
	}

	return findings
}

// collectInterfaceMethods finds all methods that implement interfaces
func collectInterfaceMethods(root *sitter.Node, code string) map[string]bool {
	interfaceMethods := make(map[string]bool)

	traverseAST(root, func(node *sitter.Node) bool {
		// Method declarations are likely interface implementations
		if node.Type() == "method_declaration" {
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "field_identifier" {
					methodName := safeSlice(code, child.StartByte(), child.EndByte())
					interfaceMethods[methodName] = true
					break
				}
			}
		}
		return true
	})

	return interfaceMethods
}

// detectOrphanedCodeGo finds orphaned functions in Go code
func detectOrphanedCodeGo(root *sitter.Node, code string) []ASTFinding {
	findings := []ASTFinding{}
	functionNames := make(map[string]*sitter.Node)
	functionCalls := make(map[string]bool)
	interfaceMethods := collectInterfaceMethods(root, code)

	// First pass: collect all function definitions
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			// Get function name
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					funcName := safeSlice(code, child.StartByte(), child.EndByte())
					functionNames[funcName] = node
					break
				}
			}
		}
		return true
	})

	// Second pass: collect all function calls and method calls
	traverseAST(root, func(node *sitter.Node) bool {
		if node.Type() == "call_expression" {
			// Get function name being called
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" {
						// Direct function call: funcName()
						funcName := safeSlice(code, child.StartByte(), child.EndByte())
						functionCalls[funcName] = true
						break
					} else if child.Type() == "selector_expression" {
						// Method call: obj.Method()
						// Extract method name from selector
						for j := 0; j < int(child.ChildCount()); j++ {
							methodChild := child.Child(j)
							if methodChild != nil && methodChild.Type() == "field_identifier" {
								methodName := safeSlice(code, methodChild.StartByte(), methodChild.EndByte())
								functionCalls[methodName] = true
								break
							}
						}
					}
				}
			}
		}
		return true
	})

	// Check for orphaned functions (simplified - doesn't handle method calls, etc.)
	config := DefaultConfig()
	for funcName, node := range functionNames {
		// Skip interface methods and excluded functions
		if interfaceMethods[funcName] {
			continue
		}
		if !functionCalls[funcName] && !shouldExcludeFunction(funcName, config) {
			startLine, startCol := getLineColumn(code, int(node.StartByte()))
			endLine, endCol := getLineColumn(code, int(node.EndByte()))

			findings = append(findings, ASTFinding{
				Type:        "orphaned_code",
				Severity:    "info",
				Line:        startLine,
				Column:      startCol,
				EndLine:     endLine,
				EndColumn:   endCol,
				Message:     fmt.Sprintf("Potentially orphaned function: '%s' is defined but never called", funcName),
				Code:        safeSlice(code, node.StartByte(), node.EndByte()),
				Suggestion:  fmt.Sprintf("Consider removing unused function '%s' or ensure it is called", funcName),
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
