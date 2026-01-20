// Package feature_discovery - Go (Gin/Chi) endpoint discovery
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// discoverGoEndpoints discovers Go (Gin/Chi) endpoints
func discoverGoEndpoints(codebasePath string, featureName string, framework string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	goFiles, _ := findFilesRecursively(codebasePath, "*.go")

	for _, file := range goFiles {
		if isExcludedPath(file) {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)
		fileEndpoints := parseGoRoutes(content, file, featureName, framework)
		endpoints = append(endpoints, fileEndpoints...)
	}

	return endpoints
}

// parseGoRoutes parses Go (Gin/Chi) route definitions
func parseGoRoutes(content string, filePath string, featureName string, framework string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// For Gin, methods are like GET, POST, etc. (uppercase)
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		// Look for router.Method( patterns
		patterns := []string{
			`r\.` + method + `\(\s*['"]([^'"]+)['"]`,
			`router\.` + method + `\(\s*['"]([^'"]+)['"]`,
			`\b` + method + `\(\s*['"]([^'"]+)['"]`, // Word boundary + general pattern
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllStringSubmatch(content, -1)

			for _, match := range matches {
				if len(match) >= 2 {
					path := match[len(match)-1] // Get the last capture group (the path)

					// Extract handler function
					handler := extractGoHandler(content, method, path)

					endpoint := EndpointInfo{
						Method:     method,
						Path:       path,
						Handler:    handler,
						File:       filePath,
						Parameters: extractPathParameters(path),
						Metadata:   make(map[string]string),
					}

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

// extractGoHandler extracts handler function name from Go route definition
func extractGoHandler(content string, method string, path string) string {
	// Look for the route definition and extract handler
	patterns := []string{
		`r\.` + method + `\(\s*['"]` + regexp.QuoteMeta(path) + `['"]\s*,\s*([^)]+)\)`,
		`router\.` + method + `\(\s*['"]` + regexp.QuoteMeta(path) + `['"]\s*,\s*([^)]+)\)`,
		method + `\(\s*['"]` + regexp.QuoteMeta(path) + `['"]\s*,\s*([^)]+)\)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(content)
		if len(match) >= 2 {
			handlerFunc := strings.TrimSpace(match[1])
			// Extract function name, handling various formats
			funcRe := regexp.MustCompile(`(\w+)(\s*\([^)]*\))?$`)
			funcMatch := funcRe.FindStringSubmatch(handlerFunc)
			if len(funcMatch) >= 2 {
				return funcMatch[1]
			}
			return strings.TrimSpace(handlerFunc)
		}
	}

	return "anonymous"
}

// discoverGoMiddleware discovers Go middleware
func discoverGoMiddleware(codebasePath string, framework string) []MiddlewareInfo {
	return []MiddlewareInfo{}
}

// discoverGoAuth discovers Go authentication
func discoverGoAuth(codebasePath string, framework string) []AuthInfo {
	return []AuthInfo{}
}
