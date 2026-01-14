// Phase 13: Test Case Generation
// Generates test requirements from structured knowledge items

package main

import (
	"fmt"
	"strings"
)

// generateTestRequirements generates test requirements from a structured knowledge item
func generateTestRequirements(item *StructuredKnowledgeItem) []TestRequirement {
	var tests []TestRequirement
	
	// For business rules, generate tests based on constraints
	if item.Specification != nil && len(item.Specification.Constraints) > 0 {
		ruleID := item.ID
		if ruleID == "" {
			ruleID = "BR-UNKNOWN"
		}
		
		testCounter := 1
		
		// Generate happy path test
		happyPathTest := TestRequirement{
			ID:       fmt.Sprintf("%s-T%d", ruleID, testCounter),
			Name:     fmt.Sprintf("test_%s_happy_path", strings.ToLower(strings.ReplaceAll(item.Title, " ", "_"))),
			Type:     "happy_path",
			Priority: "critical",
			Scenario: fmt.Sprintf("Successfully execute %s", item.Title),
			Expected: &ExpectedResult{
				Success: true,
			},
		}
		
		// Build scenario description
		if len(item.Specification.Constraints) > 0 {
			var constraintDescs []string
			for _, c := range item.Specification.Constraints {
				constraintDescs = append(constraintDescs, c.Expression)
			}
			happyPathTest.Scenario = fmt.Sprintf("Execute %s when: %s", item.Title, strings.Join(constraintDescs, " and "))
		}
		
		tests = append(tests, happyPathTest)
		testCounter++
		
		// Generate error case tests for each constraint
		for i, constraint := range item.Specification.Constraints {
			errorTest := TestRequirement{
				ID:       fmt.Sprintf("%s-T%d", ruleID, testCounter),
				Name:     fmt.Sprintf("test_%s_constraint_%d_violation", strings.ToLower(strings.ReplaceAll(item.Title, " ", "_")), i+1),
				Type:     "error_case",
				Priority: "critical",
				Scenario: fmt.Sprintf("Violate constraint: %s", constraint.Expression),
				Expected: &ExpectedResult{
					Success: false,
				},
			}
			
			// Find corresponding error case
			if len(item.Specification.ErrorCases) > i {
				errorCase := item.Specification.ErrorCases[i]
				errorTest.Expected.Error = errorCase.ErrorCode
			} else if len(item.Specification.ErrorCases) > 0 {
				// Use first error case
				errorTest.Expected.Error = item.Specification.ErrorCases[0].ErrorCode
			}
			
			tests = append(tests, errorTest)
			testCounter++
			
			// Generate boundary test for numeric constraints
			if constraint.Type == "time_based" || constraint.Type == "value_based" {
				boundaryTest := TestRequirement{
					ID:       fmt.Sprintf("%s-T%d", ruleID, testCounter),
					Name:     fmt.Sprintf("test_%s_boundary_%d", strings.ToLower(strings.ReplaceAll(item.Title, " ", "_")), i+1),
					Type:     "edge_case",
					Priority: "high",
					Scenario: fmt.Sprintf("Test boundary condition: %s (boundary: %s)", constraint.Expression, constraint.Boundary),
					Expected: &ExpectedResult{
						Success: constraint.Boundary == "inclusive",
					},
				}
				
				if constraint.Boundary == "exclusive" {
					boundaryTest.Expected.Success = false
					if len(item.Specification.ErrorCases) > 0 {
						boundaryTest.Expected.Error = item.Specification.ErrorCases[0].ErrorCode
					}
				}
				
				tests = append(tests, boundaryTest)
				testCounter++
			}
		}
		
		// Generate exception tests
		for i, exception := range item.Specification.Exceptions {
			exceptionTest := TestRequirement{
				ID:       fmt.Sprintf("%s-T%d", ruleID, testCounter),
				Name:     fmt.Sprintf("test_%s_exception_%d", strings.ToLower(strings.ReplaceAll(item.Title, " ", "_")), i+1),
				Type:     "exception_case",
				Priority: "high",
				Scenario: fmt.Sprintf("Apply exception: %s", exception.Condition),
				Expected: &ExpectedResult{
					Success: true,
				},
			}
			
			tests = append(tests, exceptionTest)
			testCounter++
		}
		
		// Ensure minimum 2 tests (happy_path + error_case)
		if len(tests) < 2 {
			// Add a generic error case if we don't have enough
			genericErrorTest := TestRequirement{
				ID:       fmt.Sprintf("%s-T%d", ruleID, testCounter),
				Name:     fmt.Sprintf("test_%s_generic_error", strings.ToLower(strings.ReplaceAll(item.Title, " ", "_"))),
				Type:     "error_case",
				Priority: "critical",
				Scenario: fmt.Sprintf("Generic error case for %s", item.Title),
				Expected: &ExpectedResult{
					Success: false,
				},
			}
			
			if len(item.Specification.ErrorCases) > 0 {
				genericErrorTest.Expected.Error = item.Specification.ErrorCases[0].ErrorCode
			}
			
			tests = append(tests, genericErrorTest)
		}
	}
	
	return tests
}











