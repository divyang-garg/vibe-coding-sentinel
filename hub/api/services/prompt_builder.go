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
	default:
		return "- Consider all aspects of code quality\n- Include maintainability and readability analysis"
	}
}
