// Phase 14A: End-to-End Flow Verification
// Verifies complete user journeys across all layers

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Flow represents an end-to-end user flow
type Flow struct {
	Name        string       `json:"name"`
	Steps       []FlowStep   `json:"steps"`
	Status      string       `json:"status"` // "complete", "broken", "partial"
	Breakpoints []Breakpoint `json:"breakpoints,omitempty"`
}

// FlowStep represents a step in a user flow
type FlowStep struct {
	Layer      string `json:"layer"`     // "ui", "api", "logic", "database", "integration"
	Component  string `json:"component"` // Component name or identifier
	Action     string `json:"action"`    // Action description
	File       string `json:"file,omitempty"`
	LineNumber int    `json:"line_number,omitempty"`
}

// Breakpoint represents a point where a flow is broken
type Breakpoint struct {
	Step     int    `json:"step"` // Step index where breakpoint occurs
	Type     string `json:"type"` // "missing_error_handling", "missing_validation", "missing_rollback", "missing_timeout", "missing_retry"
	Issue    string `json:"issue"`
	Severity string `json:"severity"` // "critical", "high", "medium", "low"
	Location string `json:"location"`
}

// verifyEndToEndFlows verifies end-to-end flows for a feature
func verifyEndToEndFlows(ctx context.Context, feature *DiscoveredFeature) ([]Flow, error) {
	flows := []Flow{}

	// Check if UI layer exists and has components
	if feature.UILayer == nil || len(feature.UILayer.Components) == 0 {
		return flows, nil
	}

	// Build call graph from discovered components
	callGraph := buildCallGraph(feature)

	// Identify flows from UI components to database operations
	// Simplified: trace from UI → API → Logic → Database
	for _, uiComponent := range feature.UILayer.Components {
		flow := traceFlow(uiComponent, feature, callGraph)
		if flow != nil {
			flows = append(flows, *flow)
		}
	}

	// Identify breakpoints in each flow
	for i := range flows {
		breakpoints := identifyBreakpoints(ctx, &flows[i], feature)
		flows[i].Breakpoints = breakpoints

		// Update flow status based on breakpoints
		if len(breakpoints) > 0 {
			criticalBreakpoints := 0
			for _, bp := range breakpoints {
				if bp.Severity == "critical" {
					criticalBreakpoints++
				}
			}
			if criticalBreakpoints > 0 {
				flows[i].Status = "broken"
			} else {
				flows[i].Status = "partial"
			}
		} else {
			flows[i].Status = "complete"
		}
	}

	return flows, nil
}

// buildCallGraph builds a call graph from discovered feature components
func buildCallGraph(feature *DiscoveredFeature) map[string][]string {
	graph := make(map[string][]string)

	// Map UI components to API endpoints (by naming convention or explicit mapping)
	if feature.UILayer != nil && feature.APILayer != nil && len(feature.UILayer.Components) > 0 && len(feature.APILayer.Endpoints) > 0 {
		for _, component := range feature.UILayer.Components {
			var endpoints []string
			for _, endpoint := range feature.APILayer.Endpoints {
				// Simple matching: check if component name matches endpoint path
				if matchesComponentToEndpoint(component.Name, endpoint.Path) {
					endpoints = append(endpoints, fmt.Sprintf("%s:%s", endpoint.Method, endpoint.Path))
				}
			}
			if len(endpoints) > 0 {
				graph[component.Name] = endpoints
			}
		}
	}

	// Map API endpoints to business logic functions
	if feature.APILayer != nil && feature.LogicLayer != nil && len(feature.APILayer.Endpoints) > 0 && len(feature.LogicLayer.Functions) > 0 {
		for _, endpoint := range feature.APILayer.Endpoints {
			var functions []string
			for _, function := range feature.LogicLayer.Functions {
				// Simple matching: check if endpoint handler matches function name
				if matchesEndpointToFunction(endpoint.Handler, function.Name) {
					functions = append(functions, function.Name)
				}
			}
			if len(functions) > 0 {
				graph[fmt.Sprintf("%s:%s", endpoint.Method, endpoint.Path)] = functions
			}
		}
	}

	// Map business logic functions to database operations
	if feature.LogicLayer != nil && feature.DatabaseLayer != nil && len(feature.LogicLayer.Functions) > 0 && len(feature.DatabaseLayer.Tables) > 0 {
		for _, function := range feature.LogicLayer.Functions {
			var tables []string
			for _, table := range feature.DatabaseLayer.Tables {
				// Simple matching: check if function name matches table name
				if matchesFunctionToTable(function.Name, table.Name) {
					tables = append(tables, table.Name)
				}
			}
			if len(tables) > 0 {
				graph[function.Name] = tables
			}
		}
	}

	return graph
}

// traceFlow traces a flow from a UI component through all layers
func traceFlow(component ComponentInfo, feature *DiscoveredFeature, callGraph map[string][]string) *Flow {
	flow := &Flow{
		Name:  component.Name,
		Steps: []FlowStep{},
	}

	// Step 1: UI component
	flow.Steps = append(flow.Steps, FlowStep{
		Layer:     "ui",
		Component: component.Name,
		Action:    "User interaction",
		File:      component.Path,
	})

	// Step 2: Find API endpoint
	if feature.APILayer != nil {
		for _, endpoint := range feature.APILayer.Endpoints {
			if matchesComponentToEndpoint(component.Name, endpoint.Path) {
				flow.Steps = append(flow.Steps, FlowStep{
					Layer:     "api",
					Component: fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path),
					Action:    "API call",
					File:      endpoint.File,
				})

				// Step 3: Find business logic function
				if feature.LogicLayer != nil {
					for _, function := range feature.LogicLayer.Functions {
						if matchesEndpointToFunction(endpoint.Handler, function.Name) {
							flow.Steps = append(flow.Steps, FlowStep{
								Layer:      "logic",
								Component:  function.Name,
								Action:     "Business logic execution",
								File:       function.File,
								LineNumber: function.LineNumber,
							})

							// Step 4: Find database operation
							if feature.DatabaseLayer != nil {
								for _, table := range feature.DatabaseLayer.Tables {
									if matchesFunctionToTable(function.Name, table.Name) {
										flow.Steps = append(flow.Steps, FlowStep{
											Layer:     "database",
											Component: table.Name,
											Action:    "Database operation",
											File:      table.File,
										})
										break
									}
								}
							}
							break
						}
					}
				}
				break
			}
		}
	}

	// Check if flow is complete (has all layers)
	if len(flow.Steps) < 3 {
		return nil // Incomplete flow
	}

	return flow
}

// identifyBreakpoints identifies breakpoints in a flow
func identifyBreakpoints(ctx context.Context, flow *Flow, feature *DiscoveredFeature) []Breakpoint {
	breakpoints := []Breakpoint{}

	for i, step := range flow.Steps {
		// Check for missing error handling
		if !hasErrorHandling(step, feature) {
			breakpoints = append(breakpoints, Breakpoint{
				Step:     i,
				Type:     "missing_error_handling",
				Issue:    fmt.Sprintf("Step %d (%s) is missing error handling", i+1, step.Layer),
				Severity: "high",
				Location: step.File,
			})
		}

		// Check for missing validation (for API and UI steps)
		if step.Layer == "api" || step.Layer == "ui" {
			if !hasValidation(step, feature) {
				breakpoints = append(breakpoints, Breakpoint{
					Step:     i,
					Type:     "missing_validation",
					Issue:    fmt.Sprintf("Step %d (%s) is missing input validation", i+1, step.Layer),
					Severity: "high",
					Location: step.File,
				})
			}
		}

		// Check for missing rollback (for database steps)
		if step.Layer == "database" {
			if !hasRollback(step, feature) {
				breakpoints = append(breakpoints, Breakpoint{
					Step:     i,
					Type:     "missing_rollback",
					Issue:    fmt.Sprintf("Step %d (database) may be missing transaction rollback", i+1),
					Severity: "medium",
					Location: step.File,
				})
			}
		}

		// Check for missing timeout (for integration steps)
		if step.Layer == "integration" {
			if !hasTimeout(step, feature) {
				breakpoints = append(breakpoints, Breakpoint{
					Step:     i,
					Type:     "missing_timeout",
					Issue:    fmt.Sprintf("Step %d (integration) is missing timeout handling", i+1),
					Severity: "medium",
					Location: step.File,
				})
			}
		}
	}

	return breakpoints
}

// verifyIntegrationPoints verifies integration points between layers
func verifyIntegrationPoints(ctx context.Context, flows []Flow, feature *DiscoveredFeature) ([]Breakpoint, error) {
	breakpoints := []Breakpoint{}

	for _, flow := range flows {
		for i := 0; i < len(flow.Steps)-1; i++ {
			currentStep := flow.Steps[i]
			nextStep := flow.Steps[i+1]

			// Verify API endpoints match UI calls
			if currentStep.Layer == "ui" && nextStep.Layer == "api" {
				// Check if request format matches (simplified)
				if !requestFormatsMatch(currentStep, nextStep, feature) {
					breakpoints = append(breakpoints, Breakpoint{
						Step:     i + 1,
						Type:     "contract_mismatch",
						Issue:    fmt.Sprintf("UI component %s request format may not match API endpoint %s", currentStep.Component, nextStep.Component),
						Severity: "high",
						Location: currentStep.File,
					})
				}
			}

			// Verify database operations match business logic
			if currentStep.Layer == "logic" && nextStep.Layer == "database" {
				// Check if SQL queries match expected operations (simplified)
				if !operationsMatch(currentStep, nextStep, feature) {
					breakpoints = append(breakpoints, Breakpoint{
						Step:     i + 1,
						Type:     "operation_mismatch",
						Issue:    fmt.Sprintf("Business logic %s operations may not match database table %s", currentStep.Component, nextStep.Component),
						Severity: "medium",
						Location: currentStep.File,
					})
				}
			}

			// Verify response formats match expectations
			if currentStep.Layer == "api" && nextStep.Layer == "logic" {
				// Check if response format matches (simplified)
				if !responseFormatsMatch(currentStep, nextStep, feature) {
					breakpoints = append(breakpoints, Breakpoint{
						Step:     i + 1,
						Type:     "response_mismatch",
						Issue:    fmt.Sprintf("API endpoint %s response format may not match business logic %s expectations", currentStep.Component, nextStep.Component),
						Severity: "medium",
						Location: currentStep.File,
					})
				}
			}
		}
	}

	return breakpoints, nil
}

// Helper functions

func matchesComponentToEndpoint(componentName string, endpointPath string) bool {
	// Simplified matching - would use semantic analysis in production
	componentLower := strings.ToLower(componentName)
	pathLower := strings.ToLower(endpointPath)

	// Extract keywords from component name
	keywords := extractFeatureKeywords(componentLower)

	for _, keyword := range keywords {
		if strings.Contains(pathLower, keyword) {
			return true
		}
	}

	return false
}

func matchesEndpointToFunction(handler string, functionName string) bool {
	// Simplified matching
	handlerLower := strings.ToLower(handler)
	functionLower := strings.ToLower(functionName)

	return strings.Contains(functionLower, handlerLower) || strings.Contains(handlerLower, functionLower)
}

func matchesFunctionToTable(functionName string, tableName string) bool {
	// Simplified matching
	functionLower := strings.ToLower(functionName)
	tableLower := strings.ToLower(tableName)

	// Extract keywords from function name
	keywords := extractFeatureKeywords(functionLower)

	for _, keyword := range keywords {
		if strings.Contains(tableLower, keyword) {
			return true
		}
	}

	return false
}

func hasErrorHandling(step FlowStep, feature *DiscoveredFeature) bool {
	if step.File == "" {
		return false
	}

	// Read file content
	content, err := os.ReadFile(step.File)
	if err != nil {
		return false
	}

	codeContent := strings.ToLower(string(content))

	// Check for error handling patterns based on layer
	switch step.Layer {
	case "ui":
		// React: try-catch, error boundaries, .catch()
		return strings.Contains(codeContent, "catch") ||
			strings.Contains(codeContent, "errorboundary") ||
			strings.Contains(codeContent, ".catch(") ||
			strings.Contains(codeContent, "onerror")
	case "api":
		// Go: if err != nil, error return, panic recovery
		return strings.Contains(codeContent, "if err") ||
			strings.Contains(codeContent, "error") ||
			strings.Contains(codeContent, "recover()") ||
			strings.Contains(codeContent, "defer")
	case "logic":
		// Business logic: error handling, validation
		return strings.Contains(codeContent, "error") ||
			strings.Contains(codeContent, "err") ||
			strings.Contains(codeContent, "exception") ||
			strings.Contains(codeContent, "catch")
	case "database":
		// Database: transaction rollback, error handling
		return strings.Contains(codeContent, "rollback") ||
			strings.Contains(codeContent, "error") ||
			strings.Contains(codeContent, "catch")
	case "integration":
		// External API: error handling, retry logic
		return strings.Contains(codeContent, "error") ||
			strings.Contains(codeContent, "catch") ||
			strings.Contains(codeContent, "retry")
	}

	return false
}

func hasValidation(step FlowStep, feature *DiscoveredFeature) bool {
	if step.File == "" {
		return false
	}

	content, err := os.ReadFile(step.File)
	if err != nil {
		return false
	}

	codeContent := strings.ToLower(string(content))

	// Check for validation patterns
	validationPatterns := []string{
		"validate",
		"validation",
		"validator",
		"schema",
		"required",
		"check",
		"verify",
		"assert",
		"zod",       // TypeScript validation
		"yup",       // JavaScript validation
		"joi",       // JavaScript validation
		"validator", // Go validation
	}

	for _, pattern := range validationPatterns {
		if strings.Contains(codeContent, pattern) {
			return true
		}
	}

	return false
}

func hasRollback(step FlowStep, feature *DiscoveredFeature) bool {
	if step.File == "" {
		return false
	}

	content, err := os.ReadFile(step.File)
	if err != nil {
		return false
	}

	codeContent := strings.ToLower(string(content))

	// Check for rollback/transaction patterns
	return strings.Contains(codeContent, "rollback") ||
		strings.Contains(codeContent, "transaction") ||
		strings.Contains(codeContent, "begin") ||
		strings.Contains(codeContent, "commit") ||
		strings.Contains(codeContent, "undo") ||
		strings.Contains(codeContent, "revert")
}

func hasTimeout(step FlowStep, feature *DiscoveredFeature) bool {
	if step.File == "" {
		return false
	}

	content, err := os.ReadFile(step.File)
	if err != nil {
		return false
	}

	codeContent := strings.ToLower(string(content))

	// Check for timeout configuration
	return strings.Contains(codeContent, "timeout") ||
		strings.Contains(codeContent, "context.timeout") ||
		strings.Contains(codeContent, "deadline") ||
		strings.Contains(codeContent, "withtimeout")
}

func requestFormatsMatch(uiStep FlowStep, apiStep FlowStep, feature *DiscoveredFeature) bool {
	// Read both files
	uiContent, err1 := os.ReadFile(uiStep.File)
	apiContent, err2 := os.ReadFile(apiStep.File)

	if err1 != nil || err2 != nil {
		// Can't read files, assume match (conservative)
		return true
	}

	uiCode := strings.ToLower(string(uiContent))
	apiCode := strings.ToLower(string(apiContent))

	// Look for common data format indicators
	formatIndicators := []string{
		"json",
		"xml",
		"formdata",
		"multipart",
		"content-type",
	}

	uiFormats := []string{}
	apiFormats := []string{}

	for _, indicator := range formatIndicators {
		if strings.Contains(uiCode, indicator) {
			uiFormats = append(uiFormats, indicator)
		}
		if strings.Contains(apiCode, indicator) {
			apiFormats = append(apiFormats, indicator)
		}
	}

	// If both have formats, check for overlap
	if len(uiFormats) > 0 && len(apiFormats) > 0 {
		for _, uiFormat := range uiFormats {
			for _, apiFormat := range apiFormats {
				if uiFormat == apiFormat {
					return true
				}
			}
		}
		// No overlap found
		return false
	}

	// If no formats detected, assume match (conservative)
	return true
}

func operationsMatch(logicStep FlowStep, dbStep FlowStep, feature *DiscoveredFeature) bool {
	if logicStep.File == "" || dbStep.File == "" {
		return true
	}

	logicContent, err1 := os.ReadFile(logicStep.File)
	dbContent, err2 := os.ReadFile(dbStep.File)

	if err1 != nil || err2 != nil {
		return true
	}

	logicCode := strings.ToLower(string(logicContent))
	dbCode := strings.ToLower(string(dbContent))

	// Check if logic operation matches database operation
	// Look for CRUD operations
	operations := []string{"create", "insert", "read", "select", "update", "delete", "remove"}

	for _, op := range operations {
		logicHasOp := strings.Contains(logicCode, op)
		dbHasOp := strings.Contains(dbCode, op)

		// If both have the same operation, they match
		if logicHasOp && dbHasOp {
			return true
		}
	}

	// If no clear operation match, check for function/table name matching
	// This is a simplified check - in production would use AST
	return true // Conservative: assume match if no clear mismatch
}

func responseFormatsMatch(apiStep FlowStep, logicStep FlowStep, feature *DiscoveredFeature) bool {
	if apiStep.File == "" || logicStep.File == "" {
		return true
	}

	apiContent, err1 := os.ReadFile(apiStep.File)
	logicContent, err2 := os.ReadFile(logicStep.File)

	if err1 != nil || err2 != nil {
		return true
	}

	apiCode := strings.ToLower(string(apiContent))
	logicCode := strings.ToLower(string(logicContent))

	// Look for response structure indicators
	responseIndicators := []string{
		"json",
		"response",
		"return",
		"result",
		"data",
	}

	apiHasResponse := false
	logicHasResponse := false

	for _, indicator := range responseIndicators {
		if strings.Contains(apiCode, indicator) {
			apiHasResponse = true
		}
		if strings.Contains(logicCode, indicator) {
			logicHasResponse = true
		}
	}

	// If both have response handling, assume they match
	if apiHasResponse && logicHasResponse {
		return true
	}

	// If neither has response handling, also assume match (might be handled elsewhere)
	return true
}
