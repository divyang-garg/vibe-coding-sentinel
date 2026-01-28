// Package main - Flow verification core orchestration
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"context"
	"fmt"
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

	// Use callGraph if available for efficient traversal, otherwise fall back to direct feature traversal
	// Note: len() for nil maps is defined as zero, so nil check is unnecessary
	if len(callGraph) > 0 {
		// Step 2: Find API endpoint using call graph
		endpoints, hasEndpoints := callGraph[component.Name]
		if hasEndpoints && len(endpoints) > 0 {
			// Use first matching endpoint from call graph
			endpointKey := endpoints[0]

			// Find endpoint details from feature
			if feature.APILayer != nil {
				for _, endpoint := range feature.APILayer.Endpoints {
					if fmt.Sprintf("%s:%s", endpoint.Method, endpoint.Path) == endpointKey {
						flow.Steps = append(flow.Steps, FlowStep{
							Layer:     "api",
							Component: fmt.Sprintf("%s %s", endpoint.Method, endpoint.Path),
							Action:    "API call",
							File:      endpoint.File,
						})

						// Step 3: Find business logic function using call graph
						functions, hasFunctions := callGraph[endpointKey]
						if hasFunctions && len(functions) > 0 {
							functionName := functions[0]
							// Find function details from feature
							if feature.LogicLayer != nil {
								for _, function := range feature.LogicLayer.Functions {
									if function.Name == functionName {
										flow.Steps = append(flow.Steps, FlowStep{
											Layer:      "logic",
											Component:  function.Name,
											Action:     "Business logic execution",
											File:       function.File,
											LineNumber: function.LineNumber,
										})

										// Step 4: Find database operation using call graph
										tables, hasTables := callGraph[functionName]
										if hasTables && len(tables) > 0 {
											tableName := tables[0]
											// Find table details from feature
											if feature.DatabaseLayer != nil {
												for _, table := range feature.DatabaseLayer.Tables {
													if table.Name == tableName {
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
										}
										break
									}
								}
							}
						}
						break
					}
				}
			}
		}
	}

	// Fallback: If callGraph is empty or doesn't have the component, use direct feature traversal
	if len(flow.Steps) == 1 {
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
	}

	// Check if flow is complete (has all layers)
	if len(flow.Steps) < 3 {
		return nil // Incomplete flow
	}

	return flow
}
