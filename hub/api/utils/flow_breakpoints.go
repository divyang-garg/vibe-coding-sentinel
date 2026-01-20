// Package utils - Flow breakpoint detection
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package utils

import (
	"context"
	"fmt"

	"sentinel-hub-api/feature_discovery"
)

// identifyBreakpoints identifies breakpoints in a flow
func identifyBreakpoints(ctx context.Context, flow *Flow, feature *feature_discovery.DiscoveredFeature) []Breakpoint {
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
func verifyIntegrationPoints(ctx context.Context, flows []Flow, feature *feature_discovery.DiscoveredFeature) ([]Breakpoint, error) {
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
