// Phase 13: Boundary Detection
// Detects and normalizes boundary specifications in constraints

package main

import (
	"fmt"
	"regexp"
	"strings"
)

// detectBoundary detects whether a constraint expression uses inclusive or exclusive boundary
func detectBoundary(expression string) (boundary string, err error) {
	expression = strings.ToLower(strings.TrimSpace(expression))
	
	// Check for explicit boundary indicators
	if strings.Contains(expression, "exclusive") || strings.Contains(expression, "strictly") {
		return "exclusive", nil
	}
	if strings.Contains(expression, "inclusive") || strings.Contains(expression, "including") {
		return "inclusive", nil
	}
	
	// Check for comparison operators
	exclusivePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\s*<\s*`),           // <
		regexp.MustCompile(`\s*>\s*`),           // >
		regexp.MustCompile(`\s*before\s+`),      // before
		regexp.MustCompile(`\s*after\s+`),       // after
		regexp.MustCompile(`\s*less than\s+`),   // less than
		regexp.MustCompile(`\s*more than\s+`),   // more than
		regexp.MustCompile(`\s*under\s+`),       // under
		regexp.MustCompile(`\s*over\s+`),        // over
	}
	
	inclusivePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\s*<=\s*`),          // <=
		regexp.MustCompile(`\s*>=\s*`),          // >=
		regexp.MustCompile(`\s*at most\s+`),     // at most
		regexp.MustCompile(`\s*at least\s+`),    // at least
		regexp.MustCompile(`\s*up to\s+`),       // up to
		regexp.MustCompile(`\s*within\s+`),      // within (usually inclusive)
		regexp.MustCompile(`\s*not more than\s+`), // not more than
		regexp.MustCompile(`\s*not less than\s+`), // not less than
	}
	
	// Check for exclusive patterns
	for _, pattern := range exclusivePatterns {
		if pattern.MatchString(expression) {
			return "exclusive", nil
		}
	}
	
	// Check for inclusive patterns
	for _, pattern := range inclusivePatterns {
		if pattern.MatchString(expression) {
			return "inclusive", nil
		}
	}
	
	// Check for "between" which is typically inclusive
	if strings.Contains(expression, "between") {
		return "inclusive", nil
	}
	
	// Check for "exactly" which is inclusive
	if strings.Contains(expression, "exactly") {
		return "inclusive", nil
	}
	
	// If we can't determine, return error to flag as ambiguous
	return "", fmt.Errorf("ambiguous boundary: cannot determine if inclusive or exclusive")
}

// normalizeBoundary normalizes boundary specification in a constraint
func normalizeBoundary(constraint *Constraint) error {
	if constraint.Boundary != "" {
		// Already specified, validate it
		if constraint.Boundary != "inclusive" && constraint.Boundary != "exclusive" {
			return fmt.Errorf("invalid boundary value: %s (must be 'inclusive' or 'exclusive')", constraint.Boundary)
		}
		return nil
	}
	
	// Try to detect from expression
	boundary, err := detectBoundary(constraint.Expression)
	if err != nil {
		// Ambiguous - will be flagged by ambiguity analyzer
		return err
	}
	
	constraint.Boundary = boundary
	return nil
}











