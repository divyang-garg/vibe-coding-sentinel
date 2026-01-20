// Phase 14A: Business Logic Analyzer
// Analyzes business logic functions for correctness, error handling, and semantic issues

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/smacker/go-tree-sitter" // Reserved for tree-sitter integration
)

// LogicLayerFinding represents a finding from business logic analysis
type LogicLayerFinding struct {
	Type     string `json:"type"`     // "semantic_error", "missing_error_handling", "signature_mismatch"
	Location string `json:"location"` // File path and line number
	Issue    string `json:"issue"`
	Severity string `json:"severity"` // "critical", "high", "medium", "low"
}

// analyzeBusinessLogic analyzes business logic functions
func analyzeBusinessLogic(ctx context.Context, projectID string, feature *DiscoveredFeature) ([]LogicLayerFinding, error) {
	return analyzeBusinessLogicWithDepth(ctx, projectID, feature, "medium")
}

// analyzeBusinessLogicWithDepth analyzes business logic functions with specified depth
// Phase 14D: Added depth parameter to control LLM usage
func analyzeBusinessLogicWithDepth(ctx context.Context, projectID string, feature *DiscoveredFeature, depth string) ([]LogicLayerFinding, error) {
	findings := []LogicLayerFinding{}

	if feature.LogicLayer == nil {
		return findings, nil
	}

	// Use AST analyzer to analyze functions
	for _, function := range feature.LogicLayer.Functions {
		// Read function file
		data, err := os.ReadFile(function.File)
		if err != nil {
			LogWarn(ctx, "Failed to read function file %s: %v", function.File, err)
			continue
		}

		// Analyze error handling (always runs, no LLM)
		errorHandlingFindings := analyzeErrorHandling(string(data), function)
		findings = append(findings, errorHandlingFindings...)

		// Phase 14D: Skip LLM semantic analysis for surface depth
		if depth == "surface" {
			// Use pattern-based checks only
			semanticFindings := checkSemanticIssues(string(data), function)
			findings = append(findings, semanticFindings...)
			continue
		}

		// Perform semantic analysis with LLM (if configured)
		// Get business rules for context
		var businessRule interface{} = nil
		// In production, would fetch relevant business rules for this function

		// Try semantic analysis with LLM (respects depth parameter)
		semanticFindings, err := semanticAnalysis(ctx, projectID, function, businessRule, depth)
		if err != nil {
			LogWarn(ctx, "Semantic analysis failed for function %s: %v", function.Name, err)
			// Fall back to pattern-based checks
			semanticFindings = checkSemanticIssues(string(data), function)
		}
		findings = append(findings, semanticFindings...)
	}

	return findings, nil
}

// analyzeErrorHandling analyzes error handling in business logic functions
func analyzeErrorHandling(code string, function BusinessLogicFunctionInfo) []LogicLayerFinding {
	findings := []LogicLayerFinding{}

	// Detect error handling patterns based on language
	// This is simplified - would use AST in production

	// Check for try-catch (JavaScript/TypeScript/Python)
	hasTryCatch := strings.Contains(code, "try") && strings.Contains(code, "catch")

	// Check for error returns (Go)
	hasErrorReturn := strings.Contains(code, "error") && strings.Contains(code, "return")

	// Check for error handling in function
	if !hasTryCatch && !hasErrorReturn {
		findings = append(findings, LogicLayerFinding{
			Type:     "missing_error_handling",
			Location: fmt.Sprintf("%s:%d", function.File, function.LineNumber),
			Issue:    fmt.Sprintf("Function %s may be missing error handling", function.Name),
			Severity: "high",
		})
	}

	return findings
}

// checkSemanticIssues checks for semantic issues in business logic
func checkSemanticIssues(code string, function BusinessLogicFunctionInfo) []LogicLayerFinding {
	findings := []LogicLayerFinding{}

	// Simplified semantic checks
	// In production, would use LLM for semantic analysis

	// Check for potential null/undefined issues
	if strings.Contains(code, ".") && !strings.Contains(code, "?.") && !strings.Contains(code, "if") {
		// Potential null reference (simplified check)
		findings = append(findings, LogicLayerFinding{
			Type:     "semantic_error",
			Location: fmt.Sprintf("%s:%d", function.File, function.LineNumber),
			Issue:    fmt.Sprintf("Function %s may have potential null reference issues", function.Name),
			Severity: "medium",
		})
	}

	return findings
}

// semanticAnalysis performs LLM-based semantic analysis
func semanticAnalysis(ctx context.Context, projectID string, function BusinessLogicFunctionInfo, businessRule interface{}, depth string) ([]LogicLayerFinding, error) {
	findings := []LogicLayerFinding{}

	// Get LLM configuration using existing function
	config, err := getLLMConfig(ctx, projectID)
	if err != nil {
		LogWarn(ctx, "Failed to get LLM configuration: %v", err)
		return findings, nil // Continue without LLM analysis
	}

	// Read function code
	code, err := os.ReadFile(function.File)
	if err != nil {
		return findings, fmt.Errorf("failed to read function file: %w", err)
	}

	codeStr := string(code)

	// Extract function code snippet using AST if available, fallback to simple extraction
	functionCode := extractFunctionCodeAST(codeStr, function.Name, function.LineNumber, function.File)
	if functionCode == "" {
		// Fallback to simple extraction
		functionCode = extractFunctionCode(codeStr, function.Name, function.LineNumber)
	}

	// Build semantic analysis prompt
	prompt := buildSemanticAnalysisPrompt(functionCode, function, businessRule)

	// Use progressive depth analysis with caching (pass the function code, not full file)
	// ValidationID will be updated after report creation
	analysisResult, err := analyzeWithProgressiveDepth(ctx, config, functionCode, "semantic_analysis", depth, projectID, "")

	if err != nil {
		// If LLM call fails, fall back to pattern-based analysis
		LogWarn(ctx, "LLM semantic analysis failed, using pattern-based fallback: %v", err)
		return checkSemanticIssues(codeStr, function), nil
	}

	// Parse LLM response
	semanticFindings := parseSemanticAnalysisResponse(ctx, analysisResult, function)
	findings = append(findings, semanticFindings...)

	// Track LLM usage using existing function
	tokensUsed := estimateTokenUsage(prompt + analysisResult)
	cost := calculateEstimatedCost(config.Provider, config.Model, tokensUsed)

	usage := &LLMUsage{
		ProjectID:     projectID,
		Provider:      config.Provider,
		Model:         config.Model,
		TokensUsed:    tokensUsed,
		EstimatedCost: cost,
	}

	if err := trackUsage(ctx, usage); err != nil {
		LogWarn(ctx, "Failed to track LLM usage: %v", err)
	}

	return findings, nil
}

// buildSemanticAnalysisPrompt builds a prompt for semantic analysis
func buildSemanticAnalysisPrompt(functionCode string, function BusinessLogicFunctionInfo, businessRule interface{}) string {
	prompt := fmt.Sprintf(`Analyze the following function for semantic correctness and business rule compliance.

Function: %s
Location: %s:%d
Code:
%s

`, function.Name, function.File, function.LineNumber, functionCode)

	// Add business rule context if available
	if businessRule != nil {
		prompt += fmt.Sprintf("Business Rule Context:\n%v\n\n", businessRule)
	}

	prompt += `Please analyze:
1. Does the function correctly implement the intended business logic?
2. Are there any semantic errors (null references, type mismatches, logic errors)?
3. Does the function handle edge cases appropriately?
4. Are there any potential bugs or issues?

Respond in JSON format:
{
  "issues": [
    {
      "type": "semantic_error|logic_error|missing_validation|edge_case",
      "severity": "critical|high|medium|low",
      "description": "Detailed description of the issue",
      "line": <line_number>
    }
  ]
}`

	return prompt
}

// extractFunctionCodeAST extracts function code using AST analysis
func extractFunctionCodeAST(fullCode string, functionName string, startLine int, filePath string) string {
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
	// NOTE: AST parsing disabled - tree-sitter integration required
	// Use simple pattern matching fallback
	return extractFunctionCodeSimple(fullCode, functionName, language)
}

// extractFunctionCodeSimple extracts function code using simple pattern matching
func extractFunctionCodeSimple(fullCode, funcName, language string) string {
	lines := strings.Split(fullCode, "\n")
	var funcStart, funcEnd int
	var inFunction bool
	var braceCount int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inFunction {
			// Look for function start
			switch language {
			case "go":
				if strings.HasPrefix(trimmed, "func ") && strings.Contains(trimmed, funcName) {
					funcStart = i
					inFunction = true
					braceCount = strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
				}
			case "javascript", "typescript":
				if strings.Contains(trimmed, "function "+funcName) || strings.Contains(trimmed, funcName+" =") {
					funcStart = i
					inFunction = true
					braceCount = strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
				}
			case "python":
				if strings.HasPrefix(trimmed, "def "+funcName) {
					funcStart = i
					inFunction = true
					// Python uses indentation, not braces
					funcEnd = findPythonFunctionEnd(lines, i)
					return strings.Join(lines[funcStart:funcEnd+1], "\n")
				}
			}
		} else {
			// Track braces to find function end
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount <= 0 {
				funcEnd = i
				return strings.Join(lines[funcStart:funcEnd+1], "\n")
			}
		}
	}

	if inFunction && funcEnd == 0 {
		// Return from funcStart to end of file
		return strings.Join(lines[funcStart:], "\n")
	}

	return ""
}

// findPythonFunctionEnd finds the end of a Python function
func findPythonFunctionEnd(lines []string, start int) int {
	if start >= len(lines) {
		return start
	}

	// Get indentation of def line
	defLine := lines[start]
	defIndent := len(defLine) - len(strings.TrimLeft(defLine, " \t"))

	for i := start + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
		if currentIndent <= defIndent && strings.TrimSpace(line) != "" {
			return i - 1
		}
	}
	return len(lines) - 1
}

// Note: extractFunctionCodeASTOriginal was disabled - tree-sitter integration required

// extractFunctionCode extracts the code for a specific function (fallback method)
func extractFunctionCode(fullCode string, functionName string, startLine int) string {
	// Simplified extraction - used as fallback when AST extraction fails
	lines := strings.Split(fullCode, "\n")

	// Find function start (simplified)
	startIdx := 0
	if startLine > 0 && startLine <= len(lines) {
		startIdx = startLine - 1
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
func parseSemanticAnalysisResponse(ctx context.Context, response string, function BusinessLogicFunctionInfo) []LogicLayerFinding {
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
