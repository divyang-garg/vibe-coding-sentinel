// Package services - Depth-aware prompt builder
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"fmt"
	"strings"
)

// buildDepthAwarePrompt builds a prompt with depth-aware detail level
// depth: 1=basic, 2=detailed, 3=comprehensive
func buildDepthAwarePrompt(basePrompt string, depth int, taskType string) string {
	switch depth {
	case 1: // Basic - minimal analysis
		return buildBasicPrompt(basePrompt, taskType)
	case 2: // Detailed - moderate analysis
		return buildDetailedPrompt(basePrompt, taskType)
	case 3: // Comprehensive - thorough analysis
		return buildComprehensivePrompt(basePrompt, taskType)
	default:
		// Default to base prompt if depth is out of range
		return basePrompt
	}
}

// buildBasicPrompt builds a basic prompt for depth level 1
func buildBasicPrompt(basePrompt string, taskType string) string {
	instructions := "Provide a brief, high-level analysis focusing on the most critical issues only."

	return fmt.Sprintf(`%s

Analysis Instructions:
- %s
- Focus on critical issues only
- Keep responses concise
- Limit to top 3-5 findings`, basePrompt, instructions)
}

// buildDetailedPrompt builds a detailed prompt for depth level 2
func buildDetailedPrompt(basePrompt string, taskType string) string {
	instructions := "Provide a thorough analysis covering major issues, patterns, and recommendations."

	return fmt.Sprintf(`%s

Analysis Instructions:
- %s
- Cover major issues and patterns
- Include specific examples
- Provide actionable recommendations
- Consider edge cases`, basePrompt, instructions)
}

// buildComprehensivePrompt builds a comprehensive prompt for depth level 3
func buildComprehensivePrompt(basePrompt string, taskType string) string {
	instructions := "Provide an exhaustive analysis covering all aspects including edge cases, optimizations, and best practices."

	additionalContext := getAdditionalContextForTaskType(taskType)

	return fmt.Sprintf(`%s

Analysis Instructions:
- %s
- Cover all issues including edge cases
- Include optimization opportunities
- Reference best practices and patterns
- Provide detailed examples and explanations
- Consider performance, security, and maintainability
%s`, basePrompt, instructions, additionalContext)
}

// getAdditionalContextForTaskType returns additional context based on task type
func getAdditionalContextForTaskType(taskType string) string {
	switch strings.ToLower(taskType) {
	case "semantic_analysis":
		return "- Analyze control flow, data flow, and logic correctness\n- Check for potential null pointer exceptions\n- Identify race conditions and concurrency issues"
	case "business_logic":
		return "- Verify business rule compliance\n- Check validation logic completeness\n- Identify missing error handling for business rules"
	case "error_handling":
		return "- Analyze error propagation paths\n- Check error recovery mechanisms\n- Verify error logging and monitoring"
	case "security":
		return "- Check for injection vulnerabilities\n- Verify authentication and authorization\n- Analyze input validation and sanitization"
	case "performance":
		return "- Identify performance bottlenecks\n- Check for inefficient algorithms\n- Analyze resource usage patterns"
	case "maintainability":
		return "- Assess code readability and structure\n- Check for code duplication\n- Evaluate documentation quality"
	case "architecture":
		return "- Analyze design patterns and structure\n- Check for architectural violations\n- Evaluate separation of concerns"
	default:
		return "- Consider all aspects of code quality\n- Include maintainability and readability analysis"
	}
}

// GeneratePrompt generates a unified prompt for all analysis types
// Supports both main package types (semantic_analysis, business_logic, error_handling)
// and services package types (security, performance, maintainability, architecture)
// Depth levels: "surface"/"quick" (brief), "medium" (detailed), "deep" (comprehensive)
func GeneratePrompt(analysisType string, depth string, fileContent string) string {
	var systemPrompt string
	var userPrompt string

	// System prompt based on analysis type
	switch analysisType {
	// Main package analysis types
	case "semantic_analysis":
		systemPrompt = "You are an expert code analyzer specializing in semantic analysis. Analyze code for logic errors, edge cases, and potential bugs."
	case "business_logic":
		systemPrompt = "You are an expert business logic analyzer. Analyze code for business rule compliance and logic correctness."
	case "error_handling":
		systemPrompt = "You are an expert in error handling analysis. Analyze code for proper error handling patterns."
	// Services package analysis types
	case "security":
		systemPrompt = "You are a security analysis expert. Analyze code for security vulnerabilities and best practices."
	case "performance":
		systemPrompt = "You are a performance optimization expert. Analyze code for performance issues and optimization opportunities."
	case "maintainability":
		systemPrompt = "You are a code quality expert. Analyze code for maintainability, readability, and best practices."
	case "architecture":
		systemPrompt = "You are an architecture expert. Analyze code structure, design patterns, and architectural concerns."
	default:
		systemPrompt = "You are a code analysis expert. Analyze the provided code."
	}

	// Normalize depth levels: "surface" maps to "quick" for consistency
	normalizedDepth := depth
	if depth == "surface" {
		normalizedDepth = "quick"
	}

	// User prompt based on depth
	switch normalizedDepth {
	case "quick":
		userPrompt = fmt.Sprintf("Quick analysis: Provide a brief summary of key findings in the following code:\n\n%s", fileContent)
	case "medium":
		// For semantic_analysis, business_logic, error_handling, use detailed JSON format
		if analysisType == "semantic_analysis" {
			userPrompt = fmt.Sprintf(`Analyze the following code with %s depth for semantic issues:

%s

Provide your analysis in JSON format with the following structure:
{
  "issues": [
    {
      "type": "error_type",
      "line": "line_number",
      "description": "detailed description",
      "severity": "low|medium|high"
    }
  ]
}`, depth, fileContent)
		} else if analysisType == "business_logic" {
			userPrompt = fmt.Sprintf(`Analyze the following business logic code with %s depth:

%s

Identify any violations of business rules, missing validations, or incorrect logic flows.`, depth, fileContent)
		} else if analysisType == "error_handling" {
			userPrompt = fmt.Sprintf(`Analyze the following code for error handling with %s depth:

%s

Identify missing error handling, improper error propagation, or error handling anti-patterns.`, depth, fileContent)
		} else {
			userPrompt = fmt.Sprintf("Medium analysis: Analyze the code and provide findings with examples. Focus on the most important issues:\n\n%s", fileContent)
		}
	case "deep":
		// For semantic_analysis, business_logic, error_handling, use detailed format
		if analysisType == "semantic_analysis" {
			userPrompt = fmt.Sprintf(`Analyze the following code with %s depth for semantic issues:

%s

Provide your analysis in JSON format with the following structure:
{
  "issues": [
    {
      "type": "error_type",
      "line": "line_number",
      "description": "detailed description",
      "severity": "low|medium|high"
    }
  ]
}`, depth, fileContent)
		} else if analysisType == "business_logic" {
			userPrompt = fmt.Sprintf(`Analyze the following business logic code with %s depth:

%s

Identify any violations of business rules, missing validations, or incorrect logic flows.`, depth, fileContent)
		} else if analysisType == "error_handling" {
			userPrompt = fmt.Sprintf(`Analyze the following code for error handling with %s depth:

%s

Identify missing error handling, improper error propagation, or error handling anti-patterns.`, depth, fileContent)
		} else {
			userPrompt = fmt.Sprintf("Deep analysis: Perform comprehensive analysis of the code. Include detailed findings, recommendations, and examples:\n\n%s", fileContent)
		}
	default:
		userPrompt = fmt.Sprintf("Analyze the following code:\n\n%s", fileContent)
	}

	// Combine into structured prompt
	return fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
}
