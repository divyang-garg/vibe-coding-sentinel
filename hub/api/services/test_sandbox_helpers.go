// Test Sandbox - Helper Functions
// Utility functions for Docker operations and validation
// Complies with CODING_STANDARDS.md: Business Services max 400 lines

package services

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// checkDockerAvailable checks if Docker is available
func checkDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// getDockerImage returns the appropriate Docker image for a language
func getDockerImage(language string) string {
	switch strings.ToLower(language) {
	case "go", "golang":
		return "golang:1.21-alpine"
	case "javascript", "js", "typescript", "ts":
		return "node:20-alpine"
	case "python", "py":
		return "python:3.11-alpine"
	default:
		return "golang:1.21-alpine" // Default fallback
	}
}

// needsNetworkAccess determines if network access is needed for test execution
func needsNetworkAccess(language string, deps []DependencyFile) bool {
	// Check if dependencies file exists and requires network
	for _, dep := range deps {
		if dep.Path == "package.json" || dep.Path == "requirements.txt" || dep.Path == "go.mod" {
			// If dependencies file exists, likely need network for installation
			return true
		}
	}
	return false
}

// validateDependencyFile validates a dependency file before writing
func validateDependencyFile(dep DependencyFile) error {
	switch dep.Path {
	case "go.mod":
		if !strings.HasPrefix(dep.Content, "module ") {
			return fmt.Errorf("invalid go.mod: missing module declaration")
		}
	case "package.json":
		var pkg map[string]interface{}
		if err := json.Unmarshal([]byte(dep.Content), &pkg); err != nil {
			return fmt.Errorf("invalid package.json: %w", err)
		}
	case "requirements.txt":
		// Basic validation - check for common patterns
		if len(dep.Content) == 0 {
			return fmt.Errorf("empty requirements.txt")
		}
		// Check for basic format (package names, version specifiers)
		lines := strings.Split(dep.Content, "\n")
		hasValidLine := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				hasValidLine = true
				break
			}
		}
		if !hasValidLine {
			return fmt.Errorf("requirements.txt has no valid package declarations")
		}
	}
	return nil
}

// getTestCommand returns the test command for a language
func getTestCommand(language string, customCommand string) string {
	if customCommand != "" {
		return customCommand
	}

	switch strings.ToLower(language) {
	case "go", "golang":
		return "go test -v ./..."
	case "javascript", "js", "typescript", "ts":
		return "npm test"
	case "python", "py":
		return "pytest -v"
	default:
		return "go test -v ./..."
	}
}

// getDirPath extracts directory path from file path
func getDirPath(filePath string) string {
	lastSlash := strings.LastIndex(filePath, "/")
	if lastSlash == -1 {
		return "."
	}
	return filePath[:lastSlash]
}
