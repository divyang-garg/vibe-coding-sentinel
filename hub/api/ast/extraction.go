// Package ast provides function extraction capabilities for AST-based code analysis.
//
// This package provides APIs for extracting function information from source code
// using Abstract Syntax Tree (AST) parsing. It supports multiple programming languages
// including Go, JavaScript, TypeScript, and Python.
//
// Example usage:
//
//	// Extract all functions matching a keyword
//	functions, err := ast.ExtractFunctions(code, "go", "calculate")
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, fn := range functions {
//		fmt.Printf("Found function: %s at line %d\n", fn.Name, fn.Line)
//	}
//
//	// Extract a specific function by name
//	funcInfo, err := ast.ExtractFunctionByName(code, "go", "calculateTotal")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Function: %s, Visibility: %s\n", funcInfo.Name, funcInfo.Visibility)
//
//	// Detect language from code or file path
//	language := ast.DetectLanguage(code, "main.go")
//	fmt.Printf("Detected language: %s\n", language)
//
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
)

// ExtractFunctions extracts all functions from code matching the keyword
// Returns a slice of FunctionInfo for functions whose names contain the keyword (case-insensitive partial match)
// If keyword is empty, returns all functions
func ExtractFunctions(code string, language string, keyword string) ([]FunctionInfo, error) {
	parser, err := getParser(language)
	if err != nil {
		return nil, fmt.Errorf("unsupported language: %w", err)
	}

	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil {
		// Return empty slice, not error, for backward compatibility
		return []FunctionInfo{}, nil
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return []FunctionInfo{}, nil
	}

	var functions []FunctionInfo
	keywordLower := strings.ToLower(keyword)

	traverseAST(rootNode, func(node *sitter.Node) bool {
		funcInfo := extractFunctionFromNode(node, code, language)
		if funcInfo != nil {
			// Match keyword (case-insensitive partial match)
			if keyword == "" || strings.Contains(strings.ToLower(funcInfo.Name), keywordLower) {
				functions = append(functions, *funcInfo)
			}
		}
		return true
	})

	return functions, nil
}

// ExtractFunctionByName extracts a specific function by exact name match
// Returns the FunctionInfo for the function or an error if not found
func ExtractFunctionByName(code string, language string, funcName string) (*FunctionInfo, error) {
	functions, err := ExtractFunctions(code, language, "")
	if err != nil {
		return nil, fmt.Errorf("failed to extract functions: %w", err)
	}

	for i := range functions {
		if functions[i].Name == funcName {
			return &functions[i], nil
		}
	}

	return nil, fmt.Errorf("function '%s' not found", funcName)
}

// extractFunctionFromNode extracts function information from an AST node
// Returns nil if the node is not a function definition
func extractFunctionFromNode(node *sitter.Node, code string, language string) *FunctionInfo {
	var funcName string
	var isFunction bool
	var startByte, endByte uint32

	switch language {
	case "go":
		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			startByte = node.StartByte()
			endByte = node.EndByte()
			// For method_declaration, format is: receiver method_name
			// For function_declaration, format is: func_name
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" {
						funcName = safeSlice(code, child.StartByte(), child.EndByte())
						isFunction = true
						break
					} else if child.Type() == "parameter_list" {
						// This is a method receiver - get the method name after it
						continue
					} else if child.Type() == "field_identifier" {
						// Method name in method_declaration
						funcName = safeSlice(code, child.StartByte(), child.EndByte())
						isFunction = true
						break
					}
				}
			}
		}
	case "javascript", "typescript":
		if node.Type() == "function_declaration" || node.Type() == "function" {
			startByte = node.StartByte()
			endByte = node.EndByte()
			// Find the function name identifier
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					funcName = safeSlice(code, child.StartByte(), child.EndByte())
					isFunction = true
					break
				}
			}
		}
	case "python":
		if node.Type() == "function_definition" {
			startByte = node.StartByte()
			endByte = node.EndByte()
			// Find the function name identifier
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil && child.Type() == "identifier" {
					funcName = safeSlice(code, child.StartByte(), child.EndByte())
					isFunction = true
					break
				}
			}
		}
	}

	if !isFunction || funcName == "" {
		return nil
	}

	// Calculate line and column positions
	startLine, startCol := getLineColumn(code, int(startByte))
	endLine, endCol := getLineColumn(code, int(endByte))

	// Extract full function code
	funcCode := safeSlice(code, startByte, endByte)

	// Determine visibility
	visibility := determineVisibility(funcName, language)

	funcInfo := &FunctionInfo{
		Name:       funcName,
		Language:   language,
		Line:       startLine,
		Column:     startCol,
		EndLine:    endLine,
		EndColumn:  endCol,
		Code:       funcCode,
		Visibility: visibility,
		Metadata:   make(map[string]string),
	}

	return funcInfo
}

// determineVisibility determines if a function is public/exported or private
func determineVisibility(funcName string, language string) string {
	if funcName == "" {
		return "private"
	}

	switch language {
	case "go":
		// In Go, exported functions start with uppercase
		if len(funcName) > 0 && unicode.IsUpper(rune(funcName[0])) {
			return "exported"
		}
		return "private"
	case "javascript", "typescript":
		// In JavaScript/TypeScript, there's no strict visibility, but we can check for export
		// For now, assume all are public (can be enhanced later)
		return "public"
	case "python":
		// In Python, functions starting with underscore are private
		if strings.HasPrefix(funcName, "_") {
			return "private"
		}
		return "public"
	default:
		return "public"
	}
}
