// Package feature_discovery - FastAPI endpoint discovery
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// discoverFastAPIEndpointsInCodebase discovers FastAPI endpoints
func discoverFastAPIEndpointsInCodebase(codebasePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	pyFiles, _ := findFilesRecursively(codebasePath, "*.py")

	for _, file := range pyFiles {
		if isExcludedPath(file) {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		content := string(data)
		fileEndpoints := parseFastAPIRoutes(content, file, featureName)
		endpoints = append(endpoints, fileEndpoints...)
	}

	return endpoints
}

// parseFastAPIRoutes parses FastAPI route definitions
func parseFastAPIRoutes(content string, filePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	methods := []string{"get", "post", "put", "delete", "patch"}

	for _, method := range methods {
		// Look for @app.METHOD( or @router.METHOD(
		decoratorPattern := `@(?:app|router)\.` + method + `\(['"]([^'"]+)['"]`
		re := regexp.MustCompile(decoratorPattern)
		matches := re.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) >= 2 {
				path := match[1]

				// Find the function that follows the decorator
				funcPattern := `def\s+(\w+)\s*\([^)]*\):`
				funcRe := regexp.MustCompile(funcPattern)
				funcMatch := funcRe.FindStringSubmatch(content)

				handler := "anonymous"
				if len(funcMatch) >= 2 {
					handler = funcMatch[1]
				}

				// Extract parameters and responses from function
				params := extractFastAPIParameters(content, handler)
				responses := extractFastAPIResponses(content, handler)

				endpoint := EndpointInfo{
					Method:     strings.ToUpper(method),
					Path:       path,
					Handler:    handler,
					File:       filePath,
					Parameters: params,
					Responses:  responses,
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

// extractFastAPIParameters extracts parameters from FastAPI function
func extractFastAPIParameters(content string, handler string) []ParameterInfo {
	params := []ParameterInfo{}

	// Find function definition
	funcPattern := `def\s+` + handler + `\s*\(([^)]*)\):`
	re := regexp.MustCompile(funcPattern)
	match := re.FindStringSubmatch(content)

	if len(match) >= 2 {
		paramsStr := match[1]

		// Extract parameter definitions
		paramRe := regexp.MustCompile(`(\w+)\s*:\s*([^=,]+)`)
		paramMatches := paramRe.FindAllStringSubmatch(paramsStr, -1)

		for _, paramMatch := range paramMatches {
			if len(paramMatch) >= 3 {
				paramType := strings.TrimSpace(paramMatch[2])
				dataType := extractPythonType(paramType)

				param := ParameterInfo{
					Name:     paramMatch[1],
					Type:     "query", // Default, could be path, body, etc.
					DataType: dataType,
					Required: !strings.Contains(paramType, "Optional"),
				}
				params = append(params, param)
			}
		}
	}

	return params
}

// extractFastAPIResponses extracts response information from FastAPI function
func extractFastAPIResponses(content string, handler string) []ResponseInfo {
	responses := []ResponseInfo{}

	// Look for status_code parameters in decorators
	statusPattern := `status_code\s*=\s*(\d+)`
	re := regexp.MustCompile(statusPattern)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			statusCode := 200 // Default
			if statusCodeStr := match[1]; statusCodeStr != "" {
				// Simple conversion, would need proper parsing
				if statusCodeStr == "200" {
					statusCode = 200
				} else if statusCodeStr == "201" {
					statusCode = 201
				} else if statusCodeStr == "400" {
					statusCode = 400
				} else if statusCodeStr == "404" {
					statusCode = 404
				} else if statusCodeStr == "500" {
					statusCode = 500
				}
			}

			response := ResponseInfo{
				StatusCode:  statusCode,
				Description: getStatusDescription(statusCode),
			}
			responses = append(responses, response)
		}
	}

	return responses
}

// discoverFastAPIMiddleware discovers FastAPI middleware
func discoverFastAPIMiddleware(codebasePath string) []MiddlewareInfo {
	return []MiddlewareInfo{}
}

// discoverFastAPIAuth discovers FastAPI authentication
func discoverFastAPIAuth(codebasePath string) []AuthInfo {
	return []AuthInfo{}
}
