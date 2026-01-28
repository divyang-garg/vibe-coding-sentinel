// Business Logic Analyzer - Helper Functions
// Code extraction, parsing, and utility functions
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// extractFunctionCodeAST extracts function code using AST analysis
func extractFunctionCodeAST(fullCode string, functionName string, _ int, filePath string) string {
	// Determine language from file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	language := ""
	switch ext {
	case ".go":
		language = "go"
	case ".js", ".jsx":
		language = "javascript"
	case ".ts", ".tsx":
		language = "typescript"
	case ".py":
		language = "python"
	default:
		return "" // Unsupported language, fallback to simple extraction
	}

	// Use AST analyzer to find the function
	parser, err := getParser(language)
	if err != nil {
		return "" // Fallback to simple extraction
	}

	tree, err := parser.ParseCtx(context.Background(), nil, []byte(fullCode))
	if err != nil {
		return "" // Fallback to simple extraction
	}
	defer tree.Close()

	rootNode := tree.RootNode()

	// Traverse AST to find the specific function
	var targetNode *sitter.Node
	TraverseAST(rootNode, func(node *sitter.Node) bool {
		var foundName string
		var isFunction bool

		switch language {
		case "go":
			if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" || child.Type() == "field_identifier" {
							foundName = fullCode[child.StartByte():child.EndByte()]
							isFunction = true
							break
						}
					}
				}
			}
		case "javascript", "typescript":
			if node.Type() == "function_declaration" || node.Type() == "function" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil {
						if child.Type() == "identifier" || child.Type() == "property_identifier" {
							foundName = fullCode[child.StartByte():child.EndByte()]
							isFunction = true
							break
						}
					}
				}
			}
		case "python":
			if node.Type() == "function_definition" {
				for i := 0; i < int(node.ChildCount()); i++ {
					child := node.Child(i)
					if child != nil && child.Type() == "identifier" {
						foundName = fullCode[child.StartByte():child.EndByte()]
						isFunction = true
						break
					}
				}
			}
		}

		if isFunction && foundName == functionName {
			targetNode = node
			return false // Stop traversal
		}
		return true // Continue traversal
	})

	if targetNode != nil {
		// Extract the function code from AST node
		return fullCode[targetNode.StartByte():targetNode.EndByte()]
	}

	return "" // Function not found, fallback to simple extraction
}

// extractFunctionCode extracts the code for a specific function (fallback method)
func extractFunctionCode(fullCode string, functionName string, startLine int) string {
	// Simplified extraction - used as fallback when AST extraction fails
	lines := strings.Split(fullCode, "\n")

	// Find function start (simplified)
	startIdx := 0
	if startLine > 0 && startLine <= len(lines) {
		startIdx = startLine - 1
	}

	// Verify function name matches at the start line
	if startIdx < len(lines) && functionName != "" {
		line := lines[startIdx]
		// Check for Go function declaration: "func FunctionName" or "func (r Receiver) FunctionName"
		if strings.Contains(line, "func ") {
			// Extract function name from line
			funcIdx := strings.Index(line, "func ")
			if funcIdx >= 0 {
				remaining := strings.TrimSpace(line[funcIdx+5:])
				// Handle method receiver: "func (r Type) MethodName"
				if strings.HasPrefix(remaining, "(") {
					// Find closing paren and method name
					parenEnd := strings.Index(remaining, ")")
					if parenEnd > 0 && parenEnd < len(remaining)-1 {
						remaining = strings.TrimSpace(remaining[parenEnd+1:])
					}
				}
				// Extract function name (before opening paren or space)
				parts := strings.Fields(remaining)
				if len(parts) > 0 {
					extractedName := strings.TrimSpace(parts[0])
					// Remove any trailing characters like opening paren
					extractedName = strings.TrimRight(extractedName, "(")
					// Verify it matches expected function name
					if extractedName != functionName {
						// Function name doesn't match - return empty to indicate failure
						return ""
					}
				}
			}
		} else if strings.Contains(line, "function ") || strings.Contains(line, "function(") {
			// JavaScript/TypeScript: "function functionName" or "const functionName = function"
			// Basic check - if function name appears in the line
			if !strings.Contains(line, functionName) {
				// Function name not found in line - return empty
				return ""
			}
		} else if strings.Contains(line, "def ") {
			// Python: "def functionName"
			funcIdx := strings.Index(line, "def ")
			if funcIdx >= 0 {
				remaining := strings.TrimSpace(line[funcIdx+4:])
				parts := strings.Fields(remaining)
				if len(parts) > 0 {
					extractedName := strings.TrimSpace(parts[0])
					extractedName = strings.TrimRight(extractedName, "(")
					if extractedName != functionName {
						return ""
					}
				}
			}
		}
	}

	// Extract function (look for closing brace or next function)
	endIdx := len(lines)
	for i := startIdx; i < len(lines); i++ {
		// Simple heuristic: look for function declaration or closing brace
		if i > startIdx && (strings.Contains(lines[i], "func ") ||
			(strings.Contains(lines[i], "}") && strings.Count(lines[i], "{") < strings.Count(lines[i], "}"))) {
			// Check if this is the end of our function
			braceCount := 0
			for j := startIdx; j <= i; j++ {
				braceCount += strings.Count(lines[j], "{")
				braceCount -= strings.Count(lines[j], "}")
			}
			if braceCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}

	// Return function code snippet
	if endIdx > startIdx {
		return strings.Join(lines[startIdx:endIdx], "\n")
	}

	return fullCode // Fallback to full code
}

// parseSemanticAnalysisResponse parses LLM response into findings
func parseSemanticAnalysisResponse(_ context.Context, response string, function BusinessLogicFunctionInfo) []LogicLayerFinding {
	findings := []LogicLayerFinding{}

	// Try to parse JSON response
	// This is simplified - would use proper JSON parsing in production
	if strings.Contains(response, "\"issues\"") {
		// Extract issues from JSON (simplified parsing)
		// In production, would use proper JSON unmarshaling
		issueMatches := extractJSONIssues(response)
		for _, issue := range issueMatches {
			findings = append(findings, LogicLayerFinding{
				Type:     issue["type"],
				Location: fmt.Sprintf("%s:%s", function.File, issue["line"]),
				Issue:    issue["description"],
				Severity: issue["severity"],
			})
		}
	} else {
		// Fallback: parse text response
		// Look for issue indicators
		if strings.Contains(strings.ToLower(response), "error") ||
			strings.Contains(strings.ToLower(response), "bug") ||
			strings.Contains(strings.ToLower(response), "issue") {
			findings = append(findings, LogicLayerFinding{
				Type:     "semantic_error",
				Location: fmt.Sprintf("%s:%d", function.File, function.LineNumber),
				Issue:    fmt.Sprintf("LLM analysis found potential issues in function %s: %s", function.Name, truncateString(response, 200)),
				Severity: "medium",
			})
		}
	}

	return findings
}

// SemanticAnalysisResponse represents the structured response from LLM semantic analysis
type SemanticAnalysisResponse struct {
	Issues []SemanticIssue `json:"issues"`
}

// SemanticIssue represents a single issue found in semantic analysis
type SemanticIssue struct {
	Type        string `json:"type"`
	Line        string `json:"line"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// extractJSONIssues extracts issues from JSON response using proper JSON parsing
func extractJSONIssues(jsonStr string) []map[string]string {
	issues := []map[string]string{}

	// Try to find JSON object in the response (may be wrapped in markdown or text)
	jsonStart := strings.Index(jsonStr, "{")
	jsonEnd := strings.LastIndex(jsonStr, "}")

	if jsonStart < 0 || jsonEnd <= jsonStart {
		// No JSON found, return empty
		return issues
	}

	// Extract JSON portion
	jsonPortion := jsonStr[jsonStart : jsonEnd+1]

	// Try to parse as structured JSON
	var response SemanticAnalysisResponse
	if err := json.Unmarshal([]byte(jsonPortion), &response); err == nil {
		// Successfully parsed structured JSON
		for _, issue := range response.Issues {
			issues = append(issues, map[string]string{
				"type":        issue.Type,
				"line":        issue.Line,
				"description": issue.Description,
				"severity":    issue.Severity,
			})
		}
		return issues
	}

	// Fallback: Try to parse as array of issues directly
	var issuesArray []SemanticIssue
	if err := json.Unmarshal([]byte(jsonPortion), &issuesArray); err == nil {
		for _, issue := range issuesArray {
			issues = append(issues, map[string]string{
				"type":        issue.Type,
				"line":        issue.Line,
				"description": issue.Description,
				"severity":    issue.Severity,
			})
		}
		return issues
	}

	// Final fallback: Try to find issues array within JSON
	var genericResponse map[string]interface{}
	if err := json.Unmarshal([]byte(jsonPortion), &genericResponse); err == nil {
		if issuesData, ok := genericResponse["issues"].([]interface{}); ok {
			for _, issueData := range issuesData {
				if issueMap, ok := issueData.(map[string]interface{}); ok {
					issue := make(map[string]string)
					if typeVal, ok := issueMap["type"].(string); ok {
						issue["type"] = typeVal
					}
					if lineVal, ok := issueMap["line"].(string); ok {
						issue["line"] = lineVal
					} else if lineVal, ok := issueMap["line"].(float64); ok {
						issue["line"] = fmt.Sprintf("%.0f", lineVal)
					}
					if descVal, ok := issueMap["description"].(string); ok {
						issue["description"] = descVal
					}
					if sevVal, ok := issueMap["severity"].(string); ok {
						issue["severity"] = sevVal
					}
					if len(issue) > 0 {
						issues = append(issues, issue)
					}
				}
			}
		}
	}

	// If all parsing fails, return empty (no issues found or invalid format)
	return issues
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func estimateTokenUsage(text string) int {
	// Simplified token estimation: ~4 characters per token
	return len(text) / 4
}
