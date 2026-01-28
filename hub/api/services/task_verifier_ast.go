// Package services - Task Verification AST Extraction Functions
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sentinel-hub-api/ast"

	sitter "github.com/smacker/go-tree-sitter"
)

// extractFunctionCallsAST extracts function calls and references matching keywords using AST analysis
func extractFunctionCallsAST(codebasePath string, keywords []string, sourceFiles []string) ([]string, error) {
	var callSites []string
	var astFailed bool

	// Build a map of keywords for faster lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	// Try AST-based analysis first
	for _, file := range sourceFiles {
		fullPath := filepath.Join(codebasePath, file)

		// Skip test files for call site detection
		if strings.Contains(fullPath, "_test.") || strings.Contains(fullPath, ".test.") {
			continue
		}

		// Read file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			astFailed = true
			continue
		}

		// Detect language from file
		language := detectLanguageFromFileForVerifier(fullPath)
		if language == "" {
			astFailed = true
			continue
		}

		// Parse AST
		parser, err := ast.GetParser(language)
		if err != nil {
			astFailed = true
			continue
		}

		ctx := context.Background()
		tree, err := parser.ParseCtx(ctx, nil, content)
		if err != nil {
			astFailed = true
			continue
		}

		rootNode := tree.RootNode()
		if rootNode == nil {
			tree.Close()
			astFailed = true
			continue
		}

		// Extract call sites from AST
		fileCallSites := extractCallSitesFromAST(rootNode, string(content), language, keywordMap, file)
		callSites = append(callSites, fileCallSites...)

		tree.Close()
	}

	// Fallback to regex if AST failed for all files
	if len(callSites) == 0 && astFailed {
		return extractFunctionCallsRegex(codebasePath, keywords, sourceFiles)
	}

	return callSites, nil
}

// extractCallSitesFromAST extracts function call sites from AST matching keywords
// Returns a slice of call site strings in format "file:line"
func extractCallSitesFromAST(root *sitter.Node, code string, language string, keywordMap map[string]bool, filePath string) []string {
	if root == nil || len(keywordMap) == 0 {
		return []string{}
	}

	var callSites []string
	codeLen := uint32(len(code))

	// Safe slice helper to prevent panics
	safeSlice := func(start, end uint32) string {
		if start > codeLen {
			start = codeLen
		}
		if end > codeLen {
			end = codeLen
		}
		if start > end {
			return ""
		}
		return code[start:end]
	}

	// Traverse AST to find call expressions
	ast.TraverseAST(root, func(node *sitter.Node) bool {
		nodeType := node.Type()

		// Check for call expressions (function calls)
		isCallExpression := false
		switch language {
		case "go":
			isCallExpression = nodeType == "call_expression"
		case "javascript", "typescript":
			isCallExpression = nodeType == "call_expression" || nodeType == "member_expression"
		case "python":
			isCallExpression = nodeType == "call"
		case "java":
			isCallExpression = nodeType == "method_invocation"
		}

		if isCallExpression {
			// Extract function name from call expression
			funcName := extractFunctionNameFromCall(node, code, language, safeSlice)
			if funcName != "" {
				funcNameLower := strings.ToLower(funcName)
				// Check if function name matches any keyword
				for keyword := range keywordMap {
					if strings.Contains(funcNameLower, keyword) || strings.Contains(keyword, funcNameLower) {
						// Get line number for call site
						startPoint := node.StartPoint()
						line := int(startPoint.Row) + 1
						callSite := fmt.Sprintf("%s:%d", filePath, line)
						callSites = append(callSites, callSite)
						break
					}
				}
			}
		}

		return true
	})

	return callSites
}

// extractFunctionNameFromCall extracts the function name from a call expression node
func extractFunctionNameFromCall(node *sitter.Node, code string, language string, safeSlice func(uint32, uint32) string) string {
	if node == nil {
		return ""
	}

	// For Go: call_expression has identifier as first child
	// For JavaScript/TypeScript: call_expression has identifier or member_expression
	// For Python: call has identifier as first child
	// For Java: method_invocation has identifier as first child

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		childType := child.Type()

		switch language {
		case "go":
			if childType == "identifier" || childType == "field_identifier" || childType == "type_identifier" {
				return safeSlice(child.StartByte(), child.EndByte())
			}
		case "javascript", "typescript":
			if childType == "identifier" || childType == "property_identifier" {
				return safeSlice(child.StartByte(), child.EndByte())
			}
			// For method calls like obj.method(), get the method name
			if childType == "member_expression" {
				// Get the property/method name (last child)
				for j := int(child.ChildCount()) - 1; j >= 0; j-- {
					memberChild := child.Child(j)
					if memberChild != nil && (memberChild.Type() == "property_identifier" || memberChild.Type() == "identifier") {
						return safeSlice(memberChild.StartByte(), memberChild.EndByte())
					}
				}
			}
		case "python":
			if childType == "identifier" || childType == "attribute" {
				// For attribute calls like obj.method(), get the attribute name
				if childType == "attribute" {
					for j := int(child.ChildCount()) - 1; j >= 0; j-- {
						attrChild := child.Child(j)
						if attrChild != nil && attrChild.Type() == "identifier" {
							return safeSlice(attrChild.StartByte(), attrChild.EndByte())
						}
					}
				} else {
					return safeSlice(child.StartByte(), child.EndByte())
				}
			}
		case "java":
			if childType == "identifier" {
				return safeSlice(child.StartByte(), child.EndByte())
			}
		}
	}

	return ""
}

// extractIdentifierFromNodeForVerifier extracts identifier name from AST node for task verifier
// Returns the identifier string from the node, or empty string if not an identifier
// Note: This is a task-verifier-specific version to avoid conflict with existing extractIdentifierFromNode
func extractIdentifierFromNodeForVerifier(node *sitter.Node, code string) string {
	if node == nil {
		return ""
	}

	codeLen := uint32(len(code))
	start := node.StartByte()
	end := node.EndByte()

	// Safe bounds checking
	if start > codeLen {
		start = codeLen
	}
	if end > codeLen {
		end = codeLen
	}
	if start > end {
		return ""
	}

	return code[start:end]
}
