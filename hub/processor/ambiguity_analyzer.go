// Phase 13: Ambiguity Analysis
// Detects ambiguous constraints and rules

package main

import (
	"fmt"
	"regexp"
	"strings"
)

// analyzeAmbiguity analyzes a knowledge item for ambiguous content
func analyzeAmbiguity(item *StructuredKnowledgeItem) []AmbiguityFlag {
	var flags []AmbiguityFlag
	
	// Analyze constraints if this is a business rule
	if item.Specification != nil {
		for i, constraint := range item.Specification.Constraints {
			// Check for vague time references
			if constraint.Type == "time_based" {
				vagueTimePatterns := []string{
					"soon", "later", "eventually", "sometime", "quickly",
					"immediately", "promptly", "asap", "as soon as possible",
				}
				for _, vague := range vagueTimePatterns {
					if strings.Contains(strings.ToLower(constraint.Expression), vague) {
						flags = append(flags, AmbiguityFlag{
							Field:           fmt.Sprintf("specification.constraints[%d].expression", i),
							Interpretations: []string{
								"Could mean within minutes",
								"Could mean within hours",
								"Could mean within days",
							},
							ClarificationQuestion: fmt.Sprintf("What does '%s' mean exactly? Please specify a concrete time period.", vague),
						})
						break
					}
				}
			}
			
			// Check for unclear boundaries
			unclearBoundaryPatterns := []string{
				"around", "approximately", "about", "roughly", "nearly",
				"close to", "near", "almost",
			}
			for _, unclear := range unclearBoundaryPatterns {
				if strings.Contains(strings.ToLower(constraint.Expression), unclear) {
					flags = append(flags, AmbiguityFlag{
						Field:           fmt.Sprintf("specification.constraints[%d].expression", i),
						Interpretations: []string{
							"Could be interpreted as inclusive",
							"Could be interpreted as exclusive",
							"Could have a tolerance range",
						},
						ClarificationQuestion: fmt.Sprintf("What does '%s' mean exactly? Is there a tolerance range?", unclear),
					})
					break
				}
			}
			
			// Check for missing units
			if constraint.Type == "time_based" || constraint.Type == "value_based" {
				// Look for numbers without units
				numberPattern := regexp.MustCompile(`\b\d+\b`)
				matches := numberPattern.FindAllString(constraint.Expression, -1)
				for _, match := range matches {
					// Check if unit is specified nearby
					unitPatterns := []string{"hour", "minute", "day", "week", "month", "year", "second"}
					hasUnit := false
					lowerExpr := strings.ToLower(constraint.Expression)
					for _, unit := range unitPatterns {
						if strings.Contains(lowerExpr, unit) {
							hasUnit = true
							break
						}
					}
					
					if !hasUnit && constraint.Unit == "" {
						flags = append(flags, AmbiguityFlag{
							Field:           fmt.Sprintf("specification.constraints[%d].unit", i),
							Interpretations: []string{
								"Could be hours",
								"Could be days",
								"Could be minutes",
							},
							ClarificationQuestion: fmt.Sprintf("What unit does '%s' refer to? (hours, days, minutes, etc.)", match),
						})
					}
				}
			}
			
			// Check for missing boundary specification
			if constraint.Boundary == "" {
				flags = append(flags, AmbiguityFlag{
					Field:           fmt.Sprintf("specification.constraints[%d].boundary", i),
					Interpretations: []string{
						"Could be inclusive (<=)",
						"Could be exclusive (<)",
					},
					ClarificationQuestion: fmt.Sprintf("Is the boundary inclusive (<=) or exclusive (<) for constraint: %s?", constraint.Expression),
				})
			}
		}
		
		// Check for multiple interpretations in expressions
		ambiguousWords := []string{"may", "might", "could", "should", "can", "possibly"}
		for i, constraint := range item.Specification.Constraints {
			for _, word := range ambiguousWords {
				if strings.Contains(strings.ToLower(constraint.Expression), word) {
					flags = append(flags, AmbiguityFlag{
						Field:           fmt.Sprintf("specification.constraints[%d].expression", i),
						Interpretations: []string{
							"Could mean 'must' (mandatory)",
							"Could mean 'may' (optional)",
							"Could mean 'should' (recommended)",
						},
						ClarificationQuestion: fmt.Sprintf("Does '%s' mean 'must', 'may', or 'should'?", word),
					})
					break
				}
			}
		}
	}
	
	// Check for vague descriptions
	if item.Description != "" {
		vagueDescPatterns := []string{
			"as needed", "when appropriate", "if necessary", "when possible",
			"reasonable", "appropriate", "suitable",
		}
		for _, vague := range vagueDescPatterns {
			if strings.Contains(strings.ToLower(item.Description), vague) {
				flags = append(flags, AmbiguityFlag{
					Field:           "description",
					Interpretations: []string{
						"Could have multiple interpretations",
						"Needs concrete criteria",
					},
					ClarificationQuestion: fmt.Sprintf("What does '%s' mean exactly? Please provide concrete criteria.", vague),
				})
				break
			}
		}
	}
	
	return flags
}

