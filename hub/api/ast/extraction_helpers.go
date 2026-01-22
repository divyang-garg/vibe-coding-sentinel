// Package ast provides helper functions for AST extraction
// Complies with CODING_STANDARDS.md: Utility modules max 250 lines
package ast

import (
	"strings"
	"unicode"

	sitter "github.com/smacker/go-tree-sitter"
)

// extractFunctionNameFromParent extracts function name from parent node
// Used for arrow functions and function expressions assigned to variables
func extractFunctionNameFromParent(node *sitter.Node, code string, language string) string {
	// For arrow functions and function expressions, we need to look at the parent
	// to find the variable name they're assigned to
	parent := node.Parent()
	if parent == nil {
		return ""
	}

	// Check if parent is a variable_declarator
	if parent.Type() == "variable_declarator" {
		// Find the identifier (variable name) in the variable_declarator
		for i := 0; i < int(parent.ChildCount()); i++ {
			child := parent.Child(i)
			if child != nil && child.Type() == "identifier" {
				return safeSlice(code, child.StartByte(), child.EndByte())
			}
		}
	}

	// Check if parent is an assignment_expression
	if parent.Type() == "assignment_expression" {
		// Find the left side (variable name)
		for i := 0; i < int(parent.ChildCount()); i++ {
			child := parent.Child(i)
			if child != nil && child.Type() == "identifier" {
				return safeSlice(code, child.StartByte(), child.EndByte())
			}
		}
	}

	// Check if parent is a property_definition (for object methods)
	if parent.Type() == "property_definition" || parent.Type() == "pair" {
		// Find the property name (key)
		for i := 0; i < int(parent.ChildCount()); i++ {
			child := parent.Child(i)
			if child != nil && (child.Type() == "property_identifier" || child.Type() == "string" || child.Type() == "identifier") {
				name := safeSlice(code, child.StartByte(), child.EndByte())
				// Remove quotes if it's a string
				if len(name) >= 2 && name[0] == '"' && name[len(name)-1] == '"' {
					name = name[1 : len(name)-1]
				}
				return name
			}
		}
	}

	// For named function expressions, check if the function itself has a name
	if node.Type() == "function_expression" {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child != nil && child.Type() == "identifier" {
				return safeSlice(code, child.StartByte(), child.EndByte())
			}
		}
	}

	return ""
}

// extractClassNameFromParent extracts class name from a class_definition node
func extractClassNameFromParent(classNode *sitter.Node, code string) string {
	if classNode == nil || classNode.Type() != "class_definition" {
		return ""
	}
	// Find the class name identifier
	for i := 0; i < int(classNode.ChildCount()); i++ {
		child := classNode.Child(i)
		if child != nil && child.Type() == "identifier" {
			return safeSlice(code, child.StartByte(), child.EndByte())
		}
	}
	return ""
}

// extractParameters extracts function parameters from an AST node
func extractParameters(node *sitter.Node, code string, language string) []ParameterInfo {
	var parameters []ParameterInfo

	switch language {
	case "go":
		parameters = extractGoParameters(node, code)
	case "javascript", "typescript":
		parameters = extractJavaScriptParameters(node, code, language)
	case "python":
		parameters = extractPythonParameters(node, code)
	}

	return parameters
}

// extractGoParameters extracts parameters from Go function nodes
func extractGoParameters(node *sitter.Node, code string) []ParameterInfo {
	var parameters []ParameterInfo

	// Find parameter_list child
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "parameter_list" {
			// Extract parameters from parameter_list
			for j := 0; j < int(child.ChildCount()); j++ {
				paramNode := child.Child(j)
				if paramNode == nil {
					continue
				}

				if paramNode.Type() == "parameter_declaration" {
					param := extractGoParameter(paramNode, code)
					if param.Name != "" {
						parameters = append(parameters, param)
					}
				}
			}
			break
		}
	}

	return parameters
}

// extractGoParameter extracts a single Go parameter
func extractGoParameter(paramNode *sitter.Node, code string) ParameterInfo {
	param := ParameterInfo{}

	// Find identifier (parameter name) and type_identifier (parameter type)
	for i := 0; i < int(paramNode.ChildCount()); i++ {
		child := paramNode.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "identifier" && param.Name == "" {
			param.Name = safeSlice(code, child.StartByte(), child.EndByte())
		} else if child.Type() == "type_identifier" || child.Type() == "pointer_type" {
			param.Type = safeSlice(code, child.StartByte(), child.EndByte())
		} else if child.Type() == "parameter_list" {
			// This might be a variadic parameter or complex type
			param.Type = safeSlice(code, child.StartByte(), child.EndByte())
		}
	}

	return param
}

// extractJavaScriptParameters extracts parameters from JavaScript/TypeScript function nodes
func extractJavaScriptParameters(node *sitter.Node, code string, language string) []ParameterInfo {
	var parameters []ParameterInfo

	// Find formal_parameters child
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "formal_parameters" {
			// Extract parameters from formal_parameters
			for j := 0; j < int(child.ChildCount()); j++ {
				paramNode := child.Child(j)
				if paramNode == nil {
					continue
				}

				if paramNode.Type() == "identifier" || paramNode.Type() == "required_parameter" {
					param := extractJavaScriptParameter(paramNode, code, language)
					if param.Name != "" {
						parameters = append(parameters, param)
					}
				} else if paramNode.Type() == "rest_parameter" {
					// Handle rest parameters (...args)
					param := extractRestParameter(paramNode, code)
					if param.Name != "" {
						parameters = append(parameters, param)
					}
				}
			}
			break
		}
	}

	return parameters
}

// extractJavaScriptParameter extracts a single JavaScript/TypeScript parameter
func extractJavaScriptParameter(paramNode *sitter.Node, code string, language string) ParameterInfo {
	param := ParameterInfo{}

	// Find identifier (parameter name)
	for i := 0; i < int(paramNode.ChildCount()); i++ {
		child := paramNode.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "identifier" && param.Name == "" {
			param.Name = safeSlice(code, child.StartByte(), child.EndByte())
		} else if child.Type() == "type_annotation" {
			// Extract type from type annotation (TypeScript)
			typeNode := child.Child(0)
			if typeNode != nil {
				param.Type = safeSlice(code, typeNode.StartByte(), typeNode.EndByte())
			}
		}
	}

	// If paramNode itself is an identifier (simple parameter)
	if paramNode.Type() == "identifier" && param.Name == "" {
		param.Name = safeSlice(code, paramNode.StartByte(), paramNode.EndByte())
	}

	return param
}

// extractRestParameter extracts a rest parameter (...args)
func extractRestParameter(paramNode *sitter.Node, code string) ParameterInfo {
	param := ParameterInfo{Type: "..."}

	// Find identifier after ...
	for i := 0; i < int(paramNode.ChildCount()); i++ {
		child := paramNode.Child(i)
		if child != nil && child.Type() == "identifier" {
			param.Name = safeSlice(code, child.StartByte(), child.EndByte())
			break
		}
	}

	return param
}

// extractPythonParameters extracts parameters from Python function nodes
func extractPythonParameters(node *sitter.Node, code string) []ParameterInfo {
	var parameters []ParameterInfo

	// Find parameters child
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "parameters" {
			// Extract parameters from parameters node
			for j := 0; j < int(child.ChildCount()); j++ {
				paramNode := child.Child(j)
				if paramNode == nil {
					continue
				}

				if paramNode.Type() == "identifier" || paramNode.Type() == "typed_parameter" {
					param := extractPythonParameter(paramNode, code)
					if param.Name != "" {
						parameters = append(parameters, param)
					}
				} else if paramNode.Type() == "typed_default_parameter" {
					// Handle parameters with default values
					param := extractPythonParameter(paramNode, code)
					if param.Name != "" {
						parameters = append(parameters, param)
					}
				}
			}
			break
		}
	}

	return parameters
}

// extractPythonParameter extracts a single Python parameter
func extractPythonParameter(paramNode *sitter.Node, code string) ParameterInfo {
	param := ParameterInfo{}

	// Find identifier (parameter name)
	for i := 0; i < int(paramNode.ChildCount()); i++ {
		child := paramNode.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "identifier" && param.Name == "" {
			param.Name = safeSlice(code, child.StartByte(), child.EndByte())
		} else if child.Type() == "type" {
			// Extract type hint
			param.Type = safeSlice(code, child.StartByte(), child.EndByte())
		}
	}

	// If paramNode itself is an identifier (simple parameter)
	if paramNode.Type() == "identifier" && param.Name == "" {
		param.Name = safeSlice(code, paramNode.StartByte(), paramNode.EndByte())
	}

	return param
}

// extractReturnType extracts function return type from an AST node
func extractReturnType(node *sitter.Node, code string, language string) string {
	switch language {
	case "go":
		return extractGoReturnType(node, code)
	case "typescript":
		return extractTypeScriptReturnType(node, code)
	case "python":
		return extractPythonReturnType(node, code)
	case "javascript":
		// JavaScript doesn't have explicit return types
		return ""
	default:
		return ""
	}
}

// extractGoReturnType extracts return type from Go function nodes
func extractGoReturnType(node *sitter.Node, code string) string {
	// Look for type_identifier after parameter_list
	foundParams := false
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "parameter_list" {
			foundParams = true
			continue
		}

		if foundParams {
			// After parameter_list, look for return type
			if child.Type() == "type_identifier" {
				return safeSlice(code, child.StartByte(), child.EndByte())
			} else if child.Type() == "parameter_list" {
				// Multiple return values: (type1, type2)
				return safeSlice(code, child.StartByte(), child.EndByte())
			} else if child.Type() == "pointer_type" {
				return safeSlice(code, child.StartByte(), child.EndByte())
			}
		}
	}

	return ""
}

// extractTypeScriptReturnType extracts return type from TypeScript function nodes
func extractTypeScriptReturnType(node *sitter.Node, code string) string {
	// Look for type_annotation after formal_parameters
	foundParams := false
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "formal_parameters" {
			foundParams = true
			continue
		}

		if foundParams && child.Type() == "type_annotation" {
			// Extract type from type_annotation
			typeNode := child.Child(0)
			if typeNode != nil {
				return safeSlice(code, typeNode.StartByte(), typeNode.EndByte())
			}
			return safeSlice(code, child.StartByte(), child.EndByte())
		}
	}

	return ""
}

// extractPythonReturnType extracts return type from Python function nodes
func extractPythonReturnType(node *sitter.Node, code string) string {
	// Look for type annotation after -> in function signature
	// Python type hints: def func() -> int:
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "type" {
			// Found type annotation
			return safeSlice(code, child.StartByte(), child.EndByte())
		}
	}

	return ""
}

// extractDocumentation extracts function documentation and comments
func extractDocumentation(node *sitter.Node, code string, language string) string {
	switch language {
	case "go":
		return extractGoDocumentation(node, code)
	case "python":
		return extractPythonDocumentation(node, code)
	case "javascript", "typescript":
		return extractJavaScriptDocumentation(node, code)
	default:
		return ""
	}
}

// extractGoDocumentation extracts Go doc comments
func extractGoDocumentation(node *sitter.Node, code string) string {
	// Look for comment nodes before the function
	parent := node.Parent()
	if parent == nil {
		return ""
	}

	var comments []string
	// Traverse siblings to find preceding comments
	for i := 0; i < int(parent.ChildCount()); i++ {
		sibling := parent.Child(i)
		if sibling == nil {
			continue
		}

		// If we found our function node, stop looking
		if sibling.StartByte() == node.StartByte() {
			break
		}

		// Check if this sibling is a comment
		if sibling.Type() == "comment" {
			comment := safeSlice(code, sibling.StartByte(), sibling.EndByte())
			// Remove // prefix and trim
			comment = strings.TrimSpace(comment)
			if strings.HasPrefix(comment, "//") {
				comment = strings.TrimSpace(comment[2:])
			}
			if comment != "" {
				comments = append(comments, comment)
			}
		}
	}

	// Reverse to get comments in order (they're added in reverse)
	if len(comments) > 0 {
		// Reverse slice
		for i, j := 0, len(comments)-1; i < j; i, j = i+1, j-1 {
			comments[i], comments[j] = comments[j], comments[i]
		}
		return strings.Join(comments, "\n")
	}

	return ""
}

// extractPythonDocumentation extracts Python docstrings
func extractPythonDocumentation(node *sitter.Node, code string) string {
	// Python docstrings are the first expression_statement in the function body
	// Look for string literal in function body
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "block" {
			// Look in block for expression_statement with string
			for j := 0; j < int(child.ChildCount()); j++ {
				stmt := child.Child(j)
				if stmt == nil {
					continue
				}

				if stmt.Type() == "expression_statement" {
					// Check if it's a string literal (docstring)
					for k := 0; k < int(stmt.ChildCount()); k++ {
						expr := stmt.Child(k)
						if expr != nil && (expr.Type() == "string" || expr.Type() == "string_literal") {
							docstring := safeSlice(code, expr.StartByte(), expr.EndByte())
							// Remove quotes
							if len(docstring) >= 3 && docstring[0] == '"' && docstring[len(docstring)-1] == '"' {
								docstring = docstring[1 : len(docstring)-1]
							} else if len(docstring) >= 6 && docstring[0:3] == `"""` && docstring[len(docstring)-3:] == `"""` {
								docstring = docstring[3 : len(docstring)-3]
							}
							return docstring
						}
					}
				}
			}
			break
		}
	}

	return ""
}

// extractJavaScriptDocumentation extracts JSDoc and comments
func extractJavaScriptDocumentation(node *sitter.Node, code string) string {
	// Look for comment nodes before the function
	parent := node.Parent()
	if parent == nil {
		return ""
	}

	var comments []string
	// Traverse siblings to find preceding comments
	for i := 0; i < int(parent.ChildCount()); i++ {
		sibling := parent.Child(i)
		if sibling == nil {
			continue
		}

		// If we found our function node, stop looking
		if sibling.StartByte() == node.StartByte() {
			break
		}

		// Check if this sibling is a comment
		if sibling.Type() == "comment" {
			comment := safeSlice(code, sibling.StartByte(), sibling.EndByte())
			comment = strings.TrimSpace(comment)
			// Handle JSDoc /** ... */
			if strings.HasPrefix(comment, "/**") && strings.HasSuffix(comment, "*/") {
				comment = comment[3 : len(comment)-2]
				comment = strings.TrimSpace(comment)
				// Remove leading * from each line
				lines := strings.Split(comment, "\n")
				for i, line := range lines {
					lines[i] = strings.TrimSpace(strings.TrimPrefix(line, "*"))
				}
				comment = strings.Join(lines, "\n")
			} else if strings.HasPrefix(comment, "//") {
				comment = strings.TrimSpace(comment[2:])
			}
			if comment != "" {
				comments = append(comments, comment)
			}
		}
	}

	// Reverse to get comments in order
	if len(comments) > 0 {
		for i, j := 0, len(comments)-1; i < j; i, j = i+1, j-1 {
			comments[i], comments[j] = comments[j], comments[i]
		}
		return strings.Join(comments, "\n")
	}

	return ""
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
		// In JavaScript/TypeScript, check for private methods (#methodName)
		if strings.HasPrefix(funcName, "#") {
			return "private"
		}
		// Check for protected methods (not standard but sometimes used)
		if strings.HasPrefix(funcName, "_") && !strings.HasPrefix(funcName, "__") {
			return "protected"
		}
		// For now, assume all others are public
		return "public"
	case "python":
		// In Python, functions starting with underscore are private
		// Extract just the method name (after the dot if it's a class method)
		methodName := funcName
		if idx := strings.LastIndex(funcName, "."); idx >= 0 {
			methodName = funcName[idx+1:]
		}
		if strings.HasPrefix(methodName, "__") && strings.HasSuffix(methodName, "__") {
			// Magic methods are public
			return "public"
		}
		if strings.HasPrefix(methodName, "_") {
			return "private"
		}
		return "public"
	default:
		return "public"
	}
}
