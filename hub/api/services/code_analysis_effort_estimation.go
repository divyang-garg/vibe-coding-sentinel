// Package services provides effort estimation functions for technical debt calculation.
//
// This file implements helper functions for estimating development effort required
// to fix code issues and determine priority levels. These functions are used by
// the technical debt estimation system.
//
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package services

// estimateIssueEffort estimates development effort in hours required to fix an issue.
//
// This function calculates effort based on issue severity and type. Base hours
// are adjusted by type-specific multipliers to account for different complexity
// levels of different issue types.
//
// Base effort hours by severity:
//   - critical: 4.0 hours
//   - high: 2.0 hours
//   - medium: 1.0 hour
//   - low: 0.5 hours
//
// Type multipliers:
//   - security: 1.5x (more complex, requires careful review)
//   - performance: 1.3x (may require profiling and optimization)
//   - bug: 1.2x (standard bug fix)
//   - style: 0.8x (simpler, mostly formatting)
//
// Parameters:
//   - severity: Issue severity level ("critical", "high", "medium", "low")
//   - issueType: Type of issue ("security", "performance", "bug", "style", etc.)
//
// Returns:
//   - Estimated effort in hours (float64)
//   - Default 1.0 hour if severity is unknown
//
// Example:
//
//	hours := service.estimateIssueEffort("high", "security")
//	// Returns: 2.0 * 1.5 = 3.0 hours
func (s *CodeAnalysisServiceImpl) estimateIssueEffort(severity, issueType string) float64 {
	baseHours := map[string]float64{
		"critical": 4.0,
		"high":     2.0,
		"medium":   1.0,
		"low":      0.5,
	}

	hours := baseHours[severity]
	if hours == 0 {
		hours = 1.0 // Default
	}

	// Adjust based on issue type
	typeMultiplier := map[string]float64{
		"security":    1.5,
		"performance": 1.3,
		"bug":         1.2,
		"style":       0.8,
	}

	if multiplier, ok := typeMultiplier[issueType]; ok {
		hours *= multiplier
	}

	return hours
}

// determinePriority determines priority level based on severity and estimated effort.
//
// This function maps issue severity and effort hours to priority levels for
// technical debt management. Higher priority issues should be addressed first.
//
// Priority determination rules:
//   - "high" priority: critical severity OR effort >= 4.0 hours
//   - "medium" priority: high severity OR effort >= 2.0 hours
//   - "low" priority: all other cases
//
// Parameters:
//   - severity: Issue severity level ("critical", "high", "medium", "low")
//   - effortHours: Estimated effort in hours to fix the issue
//
// Returns:
//   - Priority level as string: "high", "medium", or "low"
//
// Example:
//
//	priority := service.determinePriority("critical", 5.0)
//	// Returns: "high"
func (s *CodeAnalysisServiceImpl) determinePriority(severity string, effortHours float64) string {
	if severity == "critical" || effortHours >= 4.0 {
		return "high"
	}
	if severity == "high" || effortHours >= 2.0 {
		return "medium"
	}
	return "low"
}
