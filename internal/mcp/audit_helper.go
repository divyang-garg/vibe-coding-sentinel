// Package mcp provides audit helper to avoid circular dependencies
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package mcp

import (
	"github.com/divyang-garg/sentinel-hub-api/internal/scanner"
)

// runAuditScan executes a security audit scan
func runAuditScan(path string, vibeCheck, deep bool) (map[string]interface{}, error) {
	opts := scanner.ScanOptions{
		CodebasePath: path,
		CIMode:       true,
		VibeCheck:    vibeCheck,
		Deep:         deep,
		Offline:      true, // MCP runs offline by default
	}

	result, err := scanner.Scan(opts)
	if err != nil {
		return nil, err
	}

	// Convert result to map
	return map[string]interface{}{
		"status":    getStatus(result.Success),
		"path":      path,
		"findings":  len(result.Findings),
		"summary":   result.Summary,
		"success":   result.Success,
		"timestamp": result.Timestamp,
		"details":   result.Findings,
	}, nil
}

// runVibeCheck executes vibe coding detection
func runVibeCheck(path string) (map[string]interface{}, error) {
	opts := scanner.ScanOptions{
		CodebasePath: path,
		CIMode:       true,
		VibeCheck:    true,
		VibeOnly:     true,
		Offline:      true,
	}

	result, err := scanner.Scan(opts)
	if err != nil {
		return nil, err
	}

	// Filter to only vibe issues
	vibeIssues := make([]scanner.Finding, 0)
	for _, finding := range result.Findings {
		// Vibe-related finding types
		if isVibeIssue(finding.Type) {
			vibeIssues = append(vibeIssues, finding)
		}
	}

	return map[string]interface{}{
		"status": getStatus(len(vibeIssues) == 0),
		"path":   path,
		"issues": vibeIssues,
		"count":  len(vibeIssues),
	}, nil
}

// isVibeIssue checks if a finding type is vibe-related
func isVibeIssue(findingType string) bool {
	vibeTypes := map[string]bool{
		"duplicate_code":     true,
		"orphaned_code":      true,
		"large_file":         true,
		"inconsistent_style": true,
		"missing_tests":      true,
	}
	return vibeTypes[findingType]
}

// getStatus converts success boolean to status string
func getStatus(success bool) string {
	if success {
		return "passed"
	}
	return "failed"
}
