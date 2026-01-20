// Package feature_discovery - Django endpoint discovery
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// discoverDjangoEndpoints discovers Django endpoints
func discoverDjangoEndpoints(codebasePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// Look for urls.py files
	urlFiles, _ := findFilesRecursively(codebasePath, "urls.py")

	for _, file := range urlFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)
		fileEndpoints := parseDjangoRoutes(content, file, featureName)
		endpoints = append(endpoints, fileEndpoints...)
	}

	return endpoints
}

// parseDjangoRoutes parses Django URL patterns
func parseDjangoRoutes(content string, filePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// Look for path() or url() patterns
	pathPatterns := []string{
		`path\(\s*['"]([^'"]+)['"]\s*,\s*([^,]+)`,
		`url\(\s*['"]([^'"]+)['"]\s*,\s*([^,]+)`,
	}

	for _, pattern := range pathPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) >= 3 {
				path := match[1]
				view := strings.TrimSpace(match[2])

				// Extract view function name
				viewRe := regexp.MustCompile(`(\w+)`)
				viewMatch := viewRe.FindStringSubmatch(view)
				handler := "unknown"
				if len(viewMatch) >= 2 {
					handler = viewMatch[1]
				}

				endpoint := EndpointInfo{
					Method:     "GET", // Django defaults to GET, but can support others
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

	return endpoints
}

// discoverDjangoMiddleware discovers Django middleware
func discoverDjangoMiddleware(codebasePath string) []MiddlewareInfo {
	return []MiddlewareInfo{}
}
