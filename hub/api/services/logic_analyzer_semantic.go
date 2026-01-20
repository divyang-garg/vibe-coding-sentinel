// Business Logic Analyzer - Semantic Analysis Functions
// Performs LLM-based semantic analysis of business logic functions
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"fmt"
	"os"
)

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
