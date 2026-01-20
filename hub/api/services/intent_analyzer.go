// Intent Analysis Engine - Main Analysis Functions
// Analyzes user prompts to determine intent and clarity
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// SimpleLanguageTemplate defines a template for clarifying questions
type SimpleLanguageTemplate struct {
	Type        IntentType
	Template    string
	MultiChoice bool
}

// GetTemplates returns all available simple language templates
func GetTemplates() map[IntentType]SimpleLanguageTemplate {
	return map[IntentType]SimpleLanguageTemplate{
		IntentLocationUnclear: {
			Type:        IntentLocationUnclear,
			Template:    "Where should this go?\n1. %s\n2. %s",
			MultiChoice: true,
		},
		IntentEntityUnclear: {
			Type:        IntentEntityUnclear,
			Template:    "Which %s?\n1. %s\n2. %s",
			MultiChoice: true,
		},
		IntentActionConfirm: {
			Type:        IntentActionConfirm,
			Template:    "I will %s. Correct? [Y/n]",
			MultiChoice: false,
		},
	}
}

// FormatTemplate formats a template with provided values
func FormatTemplate(template SimpleLanguageTemplate, values ...string) string {
	// Convert []string to []interface{} for fmt.Sprintf
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	return fmt.Sprintf(template.Template, args...)
}

// AnalyzeIntent analyzes a user prompt to determine intent and clarity
func AnalyzeIntent(ctx context.Context, prompt string, contextData *ContextData, projectID string) (*IntentAnalysisResponse, error) {
	// Quick rule-based check for obviously clear prompts
	if isClearPrompt(prompt) {
		return &IntentAnalysisResponse{
			Success:               true,
			RequiresClarification: false,
			IntentType:            IntentClear,
			Confidence:            1.0,
			SuggestedAction:       prompt,
			ResolvedPrompt:        prompt,
		}, nil
	}

	// Use LLM for intent analysis
	llmConfig, err := getLLMConfig(ctx, projectID)
	if err != nil {
		LogWarn(ctx, "Failed to get LLM config, using rule-based fallback: %v", err)
		return analyzeIntentRuleBased(prompt, contextData)
	}

	// Build LLM prompt
	llmPrompt := buildIntentAnalysisPrompt(prompt, contextData)

	// Call LLM
	response, tokensUsed, err := callLLM(ctx, llmConfig, llmPrompt, "intent_analysis")
	if err != nil {
		LogWarn(ctx, "LLM call failed, using rule-based fallback: %v", err)
		return analyzeIntentRuleBased(prompt, contextData)
	}

	// Parse LLM response
	result, err := parseLLMIntentResponse(response)
	if err != nil {
		LogWarn(ctx, "Failed to parse LLM response, using rule-based fallback: %v", err)
		return analyzeIntentRuleBased(prompt, contextData)
	}

	// Track LLM usage
	if projectID != "" {
		usage := &LLMUsage{
			ProjectID:     projectID,
			Provider:      llmConfig.Provider,
			Model:         llmConfig.Model,
			TokensUsed:    tokensUsed,
			EstimatedCost: calculateEstimatedCost(llmConfig.Provider, llmConfig.Model, tokensUsed),
		}
		if err := trackUsage(ctx, usage); err != nil {
			LogWarn(ctx, "Failed to track LLM usage: %v", err)
		}
	}

	return result, nil
}

// isClearPrompt checks if a prompt is clearly specified
func isClearPrompt(prompt string) bool {
	// Check for specific file paths
	if strings.Contains(prompt, "/") || strings.Contains(prompt, "\\") {
		return true
	}

	// Check for specific function/class names
	if strings.Contains(prompt, "function ") || strings.Contains(prompt, "class ") {
		return true
	}

	// Check for specific commands
	clearKeywords := []string{"create", "add", "implement", "fix", "update", "delete", "remove"}
	promptLower := strings.ToLower(prompt)
	for _, keyword := range clearKeywords {
		if strings.Contains(promptLower, keyword) && len(prompt) > 20 {
			return true
		}
	}

	return false
}

// buildIntentAnalysisPrompt builds the LLM prompt for intent analysis
func buildIntentAnalysisPrompt(prompt string, contextData *ContextData) string {
	builder := strings.Builder{}
	builder.WriteString("Analyze the following user prompt and determine if it needs clarification.\n\n")
	builder.WriteString("User Prompt: " + prompt + "\n\n")

	if contextData != nil {
		if len(contextData.RecentFiles) > 0 {
			builder.WriteString("Recent files: " + strings.Join(contextData.RecentFiles[:minInt(5, len(contextData.RecentFiles))], ", ") + "\n")
		}
		if len(contextData.BusinessRules) > 0 {
			builder.WriteString("Business rules: " + strings.Join(contextData.BusinessRules[:minInt(5, len(contextData.BusinessRules))], ", ") + "\n")
		}
		if len(contextData.CodePatterns) > 0 {
			builder.WriteString("Code patterns: " + strings.Join(contextData.CodePatterns, ", ") + "\n")
		}
	}

	builder.WriteString("\nDetermine:\n")
	builder.WriteString("1. Is the prompt clear or vague?\n")
	builder.WriteString("2. What type of clarification is needed? (location_unclear, entity_unclear, action_confirm, ambiguous, or clear)\n")
	builder.WriteString("3. If clarification is needed, what clarifying question should be asked?\n")
	builder.WriteString("4. What are the likely options for the user to choose from?\n")
	builder.WriteString("\nRespond in JSON format:\n")
	builder.WriteString(`{"requires_clarification": true/false, "intent_type": "location_unclear|entity_unclear|action_confirm|ambiguous|clear", "confidence": 0.0-1.0, "clarifying_question": "...", "options": ["option1", "option2"], "suggested_action": "..."}`)

	return builder.String()
}

// parseLLMIntentResponse parses the LLM response into IntentAnalysisResponse
func parseLLMIntentResponse(response string) (*IntentAnalysisResponse, error) {
	// Try to extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[jsonStart : jsonEnd+1]
	var result struct {
		RequiresClarification bool     `json:"requires_clarification"`
		IntentType            string   `json:"intent_type"`
		Confidence            float64  `json:"confidence"`
		ClarifyingQuestion    string   `json:"clarifying_question"`
		Options               []string `json:"options"`
		SuggestedAction       string   `json:"suggested_action"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	intentType := IntentType(result.IntentType)
	if intentType == "" {
		intentType = IntentAmbiguous
	}

	responseObj := &IntentAnalysisResponse{
		Success:               true,
		RequiresClarification: result.RequiresClarification,
		IntentType:            intentType,
		Confidence:            result.Confidence,
		ClarifyingQuestion:    result.ClarifyingQuestion,
		Options:               result.Options,
		SuggestedAction:       result.SuggestedAction,
	}

	// Generate clarifying question from template if needed
	if responseObj.RequiresClarification && responseObj.ClarifyingQuestion == "" {
		templates := GetTemplates()
		template, ok := templates[responseObj.IntentType]
		if ok && len(responseObj.Options) >= 2 {
			responseObj.ClarifyingQuestion = FormatTemplate(template, responseObj.Options[0], responseObj.Options[1])
		}
	}

	return responseObj, nil
}

// getLLMConfig retrieves the LLM configuration for a project
func getLLMConfig(ctx context.Context, projectID string) (*LLMConfig, error) {
	configs, err := listLLMConfigs(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list LLM configs: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no LLM configuration found for project %s", projectID)
	}

	// Return the first (most recent) configuration
	return configs[0], nil
}

// analyzeIntentRuleBased performs rule-based intent analysis as fallback
func analyzeIntentRuleBased(prompt string, contextData *ContextData) (*IntentAnalysisResponse, error) {
	promptLower := strings.ToLower(prompt)

	// Check for location uncertainty
	locationKeywords := []string{"add", "create", "put", "place", "where"}
	hasLocationKeyword := false
	for _, keyword := range locationKeywords {
		if strings.Contains(promptLower, keyword) {
			hasLocationKeyword = true
			break
		}
	}

	if hasLocationKeyword && len(prompt) < 30 {
		// Suggest common locations based on context
		options := []string{"src/", "lib/", "app/"}
		if contextData != nil && len(contextData.CodePatterns) > 0 {
			options = contextData.CodePatterns[:minInt(2, len(contextData.CodePatterns))]
		}

		templates := GetTemplates()
		template := templates[IntentLocationUnclear]
		question := FormatTemplate(template, options[0], options[1])

		return &IntentAnalysisResponse{
			Success:               true,
			RequiresClarification: true,
			IntentType:            IntentLocationUnclear,
			Confidence:            0.6,
			ClarifyingQuestion:    question,
			Options:               options,
		}, nil
	}

	// Default: prompt seems clear
	return &IntentAnalysisResponse{
		Success:               true,
		RequiresClarification: false,
		IntentType:            IntentClear,
		Confidence:            0.7,
		SuggestedAction:       prompt,
		ResolvedPrompt:        prompt,
	}, nil
}
