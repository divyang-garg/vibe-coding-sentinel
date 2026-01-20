// Package main - LLM prompt generation
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"fmt"
)

// generatePrompt generates a structured prompt based on analysis type and depth
func generatePrompt(analysisType string, depth string, fileContent string) string {
	// Generate structured prompt based on analysis type and depth
	var systemPrompt, userPrompt string

	switch analysisType {
	case "semantic_analysis":
		systemPrompt = "You are an expert code analyzer specializing in semantic analysis. Analyze code for logic errors, edge cases, and potential bugs."
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
	case "business_logic":
		systemPrompt = "You are an expert business logic analyzer. Analyze code for business rule compliance and logic correctness."
		userPrompt = fmt.Sprintf(`Analyze the following business logic code with %s depth:

%s

Identify any violations of business rules, missing validations, or incorrect logic flows.`, depth, fileContent)
	case "error_handling":
		systemPrompt = "You are an expert in error handling analysis. Analyze code for proper error handling patterns."
		userPrompt = fmt.Sprintf(`Analyze the following code for error handling with %s depth:

%s

Identify missing error handling, improper error propagation, or error handling anti-patterns.`, depth, fileContent)
	default:
		systemPrompt = "You are an expert code analyzer."
		userPrompt = fmt.Sprintf(`Analyze the following code for %s with %s depth:

%s`, analysisType, depth, fileContent)
	}

	// Combine into structured prompt
	return fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
}
