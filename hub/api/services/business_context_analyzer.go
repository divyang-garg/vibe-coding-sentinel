// Phase 14A: Business Context Analyzer
// Validates code against business rules, user journeys, and entities

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// BusinessContextFinding represents a finding from business context analysis
type BusinessContextFinding struct {
	Type      string `json:"type"` // "business_rule_violation", "user_journey_mismatch", "entity_validation_failure"
	RuleID    string `json:"rule_id,omitempty"`
	RuleTitle string `json:"rule_title,omitempty"`
	Location  string `json:"location"`
	Issue     string `json:"issue"`
	Severity  string `json:"severity"` // "critical", "high", "medium", "low"
}

// analyzeBusinessContext analyzes code against business rules
// Phase 14D: Now accepts codebaseHash and config for caching support
func analyzeBusinessContext(ctx context.Context, projectID string, feature *DiscoveredFeature, codebaseHash string, config *LLMConfig) ([]BusinessContextFinding, error) {
	findings := []BusinessContextFinding{}

	// Extract business rules
	rules, err := extractBusinessRules(ctx, projectID, nil, codebaseHash, config)
	if err != nil {
		return nil, fmt.Errorf("failed to extract business rules: %w", err)
	}

	// For each discovered feature component, check against business rules
	// This is a simplified version - in production, would use LLM for semantic matching

	// Check UI components
	if feature.UILayer != nil {
		for _, component := range feature.UILayer.Components {
			// Check if component violates any business rules
			for _, rule := range rules {
				// Simplified matching - in production, use LLM for semantic analysis
				if matchesComponentToRule(component, rule) {
					// Check for violations (simplified)
					if hasViolation(component, rule) {
						findings = append(findings, BusinessContextFinding{
							Type:      "business_rule_violation",
							RuleID:    rule.ID,
							RuleTitle: rule.Title,
							Location:  component.Path,
							Issue:     fmt.Sprintf("Component %s may violate business rule: %s", component.Name, rule.Title),
							Severity:  "high",
						})
					}
				}
			}
		}
	}

	// Check API endpoints
	if feature.APILayer != nil {
		for _, endpoint := range feature.APILayer.Endpoints {
			for _, rule := range rules {
				if matchesEndpointToRule(endpoint, rule) {
					if hasViolation(endpoint, rule) {
						findings = append(findings, BusinessContextFinding{
							Type:      "business_rule_violation",
							RuleID:    rule.ID,
							RuleTitle: rule.Title,
							Location:  endpoint.File,
							Issue:     fmt.Sprintf("Endpoint %s %s may violate business rule: %s", endpoint.Method, endpoint.Path, rule.Title),
							Severity:  "high",
						})
					}
				}
			}
		}
	}

	// Check business logic functions
	if feature.LogicLayer != nil {
		for _, function := range feature.LogicLayer.Functions {
			for _, rule := range rules {
				if matchesFunctionToRule(function, rule) {
					if hasViolation(function, rule) {
						findings = append(findings, BusinessContextFinding{
							Type:      "business_rule_violation",
							RuleID:    rule.ID,
							RuleTitle: rule.Title,
							Location:  function.File,
							Issue:     fmt.Sprintf("Function %s may violate business rule: %s", function.Name, rule.Title),
							Severity:  "critical",
						})
					}
				}
			}
		}
	}

	return findings, nil
}

// checkJourneyAdherence checks if feature adheres to user journey steps
// Phase 14D: Now accepts codebaseHash and config for caching support
func checkJourneyAdherence(ctx context.Context, projectID string, feature *DiscoveredFeature, codebaseHash string, config *LLMConfig) ([]BusinessContextFinding, error) {
	findings := []BusinessContextFinding{}

	// Extract user journeys
	journeys, err := extractUserJourneys(ctx, projectID, nil, codebaseHash, config)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user journeys: %w", err)
	}

	// For each journey, check if all steps are implemented
	for _, journey := range journeys {
		// Parse journey steps (simplified - would parse structured data in production)
		steps := parseJourneySteps(journey.Content)

		// Check if each step has corresponding implementation
		for _, step := range steps {
			missing := checkStepImplementation(step, feature)
			if len(missing) > 0 {
				findings = append(findings, BusinessContextFinding{
					Type:      "user_journey_mismatch",
					RuleID:    journey.ID,
					RuleTitle: journey.Title,
					Location:  "feature",
					Issue:     fmt.Sprintf("Journey step '%s' is missing implementation: %v", step, missing),
					Severity:  "high",
				})
			}
		}
	}

	return findings, nil
}

// Helper functions (simplified implementations)

func matchesComponentToRule(component ComponentInfo, rule KnowledgeItem) bool {
	// Simplified matching - check if component name or path contains rule keywords
	keywords := extractFeatureKeywords(strings.ToLower(rule.Title))
	componentLower := strings.ToLower(component.Name + " " + component.Path)

	for _, keyword := range keywords {
		if strings.Contains(componentLower, keyword) {
			return true
		}
	}
	return false
}

func matchesEndpointToRule(endpoint EndpointInfo, rule KnowledgeItem) bool {
	keywords := extractFeatureKeywords(strings.ToLower(rule.Title))
	endpointLower := strings.ToLower(endpoint.Path + " " + endpoint.File)

	for _, keyword := range keywords {
		if strings.Contains(endpointLower, keyword) {
			return true
		}
	}
	return false
}

func matchesFunctionToRule(function BusinessLogicFunctionInfo, rule KnowledgeItem) bool {
	keywords := extractFeatureKeywords(strings.ToLower(rule.Title))
	functionLower := strings.ToLower(function.Name + " " + function.File)

	for _, keyword := range keywords {
		if strings.Contains(functionLower, keyword) {
			return true
		}
	}
	return false
}

func hasViolation(component interface{}, rule KnowledgeItem) bool {
	// Check if component violates business rule by analyzing code
	// This is a simplified implementation - in production would use LLM for semantic analysis

	var filePath string
	var codeContent string

	// Extract file path and code based on component type
	switch c := component.(type) {
	case ComponentInfo:
		filePath = c.Path
	case EndpointInfo:
		filePath = c.File
	case BusinessLogicFunctionInfo:
		filePath = c.File
	default:
		return false // Unknown type, assume no violation
	}

	if filePath == "" {
		return false
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false // Can't read file, assume no violation
	}

	codeContent = strings.ToLower(string(content))
	ruleContent := strings.ToLower(rule.Content + " " + rule.Title)

	// Extract keywords from rule
	ruleKeywords := extractFeatureKeywords(ruleContent)

	// Check for negative patterns (indicating violation)
	negativePatterns := []string{
		"not allowed",
		"forbidden",
		"prohibited",
		"must not",
		"should not",
		"cannot",
		"must never",
	}

	// Check if code contains negative patterns from rule
	for _, keyword := range ruleKeywords {
		for _, pattern := range negativePatterns {
			if strings.Contains(ruleContent, pattern+" "+keyword) {
				// Rule says "must not X", check if code does X
				if strings.Contains(codeContent, keyword) {
					return true // Violation detected
				}
			}
		}
	}

	// Check for required patterns (rule says "must X", check if code has X)
	requiredPatterns := []string{
		"must",
		"required",
		"should",
		"shall",
	}

	for _, keyword := range ruleKeywords {
		for _, pattern := range requiredPatterns {
			if strings.Contains(ruleContent, pattern+" "+keyword) {
				// Rule says "must X", check if code has X
				if !strings.Contains(codeContent, keyword) {
					return true // Violation: required pattern missing
				}
			}
		}
	}

	return false // No violation detected
}

func parseJourneySteps(journeyContent string) []string {
	steps := []string{}

	// Try to parse as JSON first
	var jsonJourney struct {
		Steps []string `json:"steps"`
		Flow  []string `json:"flow"`
	}
	if err := json.Unmarshal([]byte(journeyContent), &jsonJourney); err == nil {
		if len(jsonJourney.Steps) > 0 {
			return jsonJourney.Steps
		}
		if len(jsonJourney.Flow) > 0 {
			return jsonJourney.Flow
		}
	}

	// Try to parse as markdown list
	lines := strings.Split(journeyContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Check for markdown list items
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "1. ") {
			step := strings.TrimPrefix(line, "- ")
			step = strings.TrimPrefix(step, "* ")
			step = strings.TrimPrefix(step, "1. ")
			step = strings.TrimSpace(step)
			if step != "" {
				steps = append(steps, step)
			}
		}
	}

	// If no structured format found, try to extract steps from text
	if len(steps) == 0 {
		// Look for numbered steps or action verbs
		sentences := strings.Split(journeyContent, ".")
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			// Look for action verbs at start of sentence
			actionVerbs := []string{"click", "enter", "submit", "select", "navigate", "view", "create", "update", "delete"}
			for _, verb := range actionVerbs {
				if strings.HasPrefix(strings.ToLower(sentence), verb+" ") {
					steps = append(steps, sentence)
					break
				}
			}
		}
	}

	return steps
}

func checkStepImplementation(step string, feature *DiscoveredFeature) []string {
	missing := []string{}
	stepLower := strings.ToLower(step)

	// Extract keywords from step
	stepKeywords := extractFeatureKeywords(stepLower)

	// Check if step has UI implementation
	hasUI := false
	if feature.UILayer != nil {
		for _, component := range feature.UILayer.Components {
			componentLower := strings.ToLower(component.Name + " " + component.Path)
			for _, keyword := range stepKeywords {
				if strings.Contains(componentLower, keyword) {
					hasUI = true
					break
				}
			}
			if hasUI {
				break
			}
		}
	}
	if !hasUI && (strings.Contains(stepLower, "click") || strings.Contains(stepLower, "view") || strings.Contains(stepLower, "navigate") || strings.Contains(stepLower, "select")) {
		missing = append(missing, "ui")
	}

	// Check if step has API implementation
	hasAPI := false
	if feature.APILayer != nil {
		for _, endpoint := range feature.APILayer.Endpoints {
			endpointLower := strings.ToLower(endpoint.Path + " " + endpoint.File)
			for _, keyword := range stepKeywords {
				if strings.Contains(endpointLower, keyword) {
					hasAPI = true
					break
				}
			}
			if hasAPI {
				break
			}
		}
	}
	if !hasAPI && (strings.Contains(stepLower, "submit") || strings.Contains(stepLower, "send") || strings.Contains(stepLower, "request") || strings.Contains(stepLower, "api")) {
		missing = append(missing, "api")
	}

	// Check if step has Logic implementation
	hasLogic := false
	if feature.LogicLayer != nil {
		for _, function := range feature.LogicLayer.Functions {
			functionLower := strings.ToLower(function.Name + " " + function.File)
			for _, keyword := range stepKeywords {
				if strings.Contains(functionLower, keyword) {
					hasLogic = true
					break
				}
			}
			if hasLogic {
				break
			}
		}
	}
	if !hasLogic && (strings.Contains(stepLower, "process") || strings.Contains(stepLower, "calculate") || strings.Contains(stepLower, "validate") || strings.Contains(stepLower, "business")) {
		missing = append(missing, "logic")
	}

	// Check if step has Database implementation
	hasDB := false
	if feature.DatabaseLayer != nil {
		for _, table := range feature.DatabaseLayer.Tables {
			tableLower := strings.ToLower(table.Name + " " + table.File)
			for _, keyword := range stepKeywords {
				if strings.Contains(tableLower, keyword) {
					hasDB = true
					break
				}
			}
			if hasDB {
				break
			}
		}
	}
	if !hasDB && (strings.Contains(stepLower, "save") || strings.Contains(stepLower, "store") || strings.Contains(stepLower, "database") || strings.Contains(stepLower, "persist")) {
		missing = append(missing, "database")
	}

	return missing
}
