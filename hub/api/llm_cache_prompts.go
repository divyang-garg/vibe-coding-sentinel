// Package main - LLM prompt generation
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package main

import (
	"sentinel-hub-api/services"
)

// generatePrompt generates a structured prompt based on analysis type and depth
// This is a wrapper that delegates to the unified prompt builder in the services package
func generatePrompt(analysisType string, depth string, fileContent string) string {
	return services.GeneratePrompt(analysisType, depth, fileContent)
}
