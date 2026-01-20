// Package feature_discovery - Express.js endpoint discovery
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// discoverExpressEndpoints discovers Express.js endpoints
func discoverExpressEndpoints(codebasePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// Search for JavaScript/TypeScript files recursively
	jsFiles, _ := findFilesRecursively(codebasePath, "*.js")
	tsFiles, _ := findFilesRecursively(codebasePath, "*.ts")
	allFiles := append(jsFiles, tsFiles...)

	for _, file := range allFiles {
		if isExcludedPath(file) {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)
		fileEndpoints := parseExpressRoutesFromContent(content, file, featureName)
		endpoints = append(endpoints, fileEndpoints...)
	}

	return endpoints
}

// parseExpressRoutesFromContent parses Express route definitions from file content
func parseExpressRoutesFromContent(content string, filePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// HTTP methods supported by Express
	methods := []string{"get", "post", "put", "delete", "patch", "head", "options"}

	for _, method := range methods {
		// Look for app.METHOD( or router.METHOD(
		patterns := []string{
			`app\.` + method + `\(\s*['"]([^'"]+)['"]`,
			`router\.` + method + `\(\s*['"]([^'"]+)['"]`,
			`\.` + method + `\(\s*['"]([^'"]+)['"]`, // More general pattern
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(`(?i)` + pattern) // Case insensitive
			matches := re.FindAllStringSubmatch(content, -1)

			for _, match := range matches {
				if len(match) >= 2 {
					path := match[len(match)-1] // Get the last capture group (the path)

					// Extract handler function and parameters
					handler, params := extractExpressHandlerFromRoute(content, method, path)

					endpoint := EndpointInfo{
						Method:     strings.ToUpper(method),
						Path:       path,
						Handler:    handler,
						File:       filePath,
						Parameters: params,
						Metadata:   make(map[string]string),
					}

					// Check if endpoint matches feature
					if matchesFeature(filepath.Base(filePath), featureName) ||
						matchesFeature(path, featureName) ||
						matchesFeature(handler, featureName) {
						endpoints = append(endpoints, endpoint)
					}
				}
			}
		}
	}

	return endpoints
}

// extractExpressHandlerFromRoute extracts handler function and parameters from Express route
func extractExpressHandlerFromRoute(content string, method string, path string) (string, []ParameterInfo) {
	// Find the route definition line
	routePattern := method + `\(\s*['"]` + regexp.QuoteMeta(path) + `['"]\s*,\s*([^)]+)\)`
	re := regexp.MustCompile(routePattern)
	match := re.FindStringSubmatch(content)

	if len(match) < 2 {
		return "anonymous", []ParameterInfo{}
	}

	handler := strings.TrimSpace(match[1])
	// Extract function name if it's a function reference
	if strings.Contains(handler, "(") {
		handler = "anonymous"
	}

	// Extract path parameters
	params := extractPathParameters(path)

	return handler, params
}

// discoverExpressMiddlewareInCodebase discovers Express middleware
func discoverExpressMiddlewareInCodebase(codebasePath string) []MiddlewareInfo {
	// Implementation would scan for middleware usage
	return []MiddlewareInfo{}
}

// discoverExpressAuthInCodebase discovers Express authentication
func discoverExpressAuthInCodebase(codebasePath string) []AuthInfo {
	return []AuthInfo{}
}
