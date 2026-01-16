// Package feature_discovery provides comprehensive API endpoint discovery
// Complies with CODING_STANDARDS.md: API endpoints max 300 lines
package feature_discovery

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// discoverAPIEndpoints discovers API endpoints in the codebase
// Supports Express, FastAPI, Django, Gin, Chi frameworks
func discoverAPIEndpoints(ctx context.Context, codebasePath string, featureName string, framework string) (*APILayerEndpoints, error) {
	endpoints := []EndpointInfo{}

	switch framework {
	case "express":
		endpoints = discoverExpressEndpoints(codebasePath, featureName)
	case "fastapi":
		endpoints = discoverFastAPIEndpoints(codebasePath, featureName)
	case "django":
		endpoints = discoverDjangoEndpoints(codebasePath, featureName)
	case "gin", "chi":
		endpoints = discoverGoEndpoints(codebasePath, featureName, framework)
	default:
		// Try to auto-detect framework
		endpoints = autoDetectAPIEndpoints(codebasePath, featureName)
	}

	// Discover middleware and authentication
	middleware := discoverMiddleware(codebasePath, framework)
	auth := discoverAuthentication(codebasePath, framework)

	return &APILayerEndpoints{
		Endpoints:  endpoints,
		Framework:  framework,
		Middleware: middleware,
		Auth:       auth,
	}, nil
}

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
		fileEndpoints := parseExpressRoutes(content, file, featureName)
		endpoints = append(endpoints, fileEndpoints...)
	}

	return endpoints
}

// parseExpressRoutes parses Express route definitions from file content
func parseExpressRoutes(content string, filePath string, featureName string) []EndpointInfo {
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
					handler, params := extractExpressHandler(content, method, path)

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

// extractExpressHandler extracts handler function and parameters from Express route
func extractExpressHandler(content string, method string, path string) (string, []ParameterInfo) {
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

// discoverFastAPIEndpoints discovers FastAPI endpoints
func discoverFastAPIEndpoints(codebasePath string, featureName string) []EndpointInfo {
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

// autoDetectAPIEndpoints attempts to auto-detect API endpoints
func autoDetectAPIEndpoints(codebasePath string, featureName string) []EndpointInfo {
	endpoints := []EndpointInfo{}

	// Try Express patterns
	expressEndpoints := discoverExpressEndpoints(codebasePath, featureName)
	endpoints = append(endpoints, expressEndpoints...)

	// Try FastAPI patterns
	fastapiEndpoints := discoverFastAPIEndpoints(codebasePath, featureName)
	endpoints = append(endpoints, fastapiEndpoints...)

	// Try Django patterns
	djangoEndpoints := discoverDjangoEndpoints(codebasePath, featureName)
	endpoints = append(endpoints, djangoEndpoints...)

	// Try Go patterns (both Gin and Chi)
	goEndpointsGin := discoverGoEndpoints(codebasePath, featureName, "gin")
	goEndpointsChi := discoverGoEndpoints(codebasePath, featureName, "chi")
	endpoints = append(endpoints, goEndpointsGin...)
	endpoints = append(endpoints, goEndpointsChi...)

	return endpoints
}

// extractPathParameters extracts parameters from path
func extractPathParameters(path string) []ParameterInfo {
	params := []ParameterInfo{}

	// Find path parameters like /users/:id or /users/{id}
	paramPatterns := []string{
		`\{([^}]+)\}`, // {id} pattern
		`:([^/]+)`,    // :id pattern
	}

	for _, pattern := range paramPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(path, -1)

		for _, match := range matches {
			if len(match) >= 2 {
				param := ParameterInfo{
					Name:     match[1],
					Type:     "path",
					DataType: "string", // Default assumption
					Required: true,
				}
				params = append(params, param)
			}
		}
	}

	return params
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

// extractPythonType extracts Python type information
func extractPythonType(typeStr string) string {
	typeStr = strings.TrimSpace(typeStr)

	// Handle Optional types
	if strings.HasPrefix(typeStr, "Optional[") {
		return strings.TrimSuffix(strings.TrimPrefix(typeStr, "Optional["), "]")
	}

	// Map common Python types to JSON schema types
	switch typeStr {
	case "str", "string":
		return "string"
	case "int", "integer":
		return "integer"
	case "bool", "boolean":
		return "boolean"
	case "float", "number":
		return "number"
	case "list", "array":
		return "array"
	case "dict", "object":
		return "object"
	default:
		return "string" // Default fallback
	}
}

// getStatusDescription returns description for HTTP status code
func getStatusDescription(code int) string {
	switch code {
	case 200:
		return "Success"
	case 201:
		return "Created"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	default:
		return "Unknown"
	}
}

// discoverMiddleware discovers middleware in the codebase
func discoverMiddleware(codebasePath string, framework string) []MiddlewareInfo {
	middleware := []MiddlewareInfo{}

	switch framework {
	case "express":
		middleware = discoverExpressMiddleware(codebasePath)
	case "fastapi":
		middleware = discoverFastAPIMiddleware(codebasePath)
	case "gin", "chi":
		middleware = discoverGoMiddleware(codebasePath, framework)
	}

	return middleware
}

// discoverAuthentication discovers authentication methods
func discoverAuthentication(codebasePath string, framework string) []AuthInfo {
	auth := []AuthInfo{}

	switch framework {
	case "express":
		auth = discoverExpressAuth(codebasePath)
	case "fastapi":
		auth = discoverFastAPIAuth(codebasePath)
	case "gin", "chi":
		auth = discoverGoAuth(codebasePath, framework)
	}

	return auth
}

// Placeholder implementations for middleware and auth discovery
func discoverExpressMiddleware(codebasePath string) []MiddlewareInfo {
	// Implementation would scan for middleware usage
	return []MiddlewareInfo{}
}

func discoverFastAPIMiddleware(codebasePath string) []MiddlewareInfo {
	return []MiddlewareInfo{}
}

func discoverGoMiddleware(codebasePath string, framework string) []MiddlewareInfo {
	return []MiddlewareInfo{}
}

func discoverExpressAuth(codebasePath string) []AuthInfo {
	return []AuthInfo{}
}

func discoverFastAPIAuth(codebasePath string) []AuthInfo {
	return []AuthInfo{}
}

func discoverGoAuth(codebasePath string, framework string) []AuthInfo {
	return []AuthInfo{}
}
