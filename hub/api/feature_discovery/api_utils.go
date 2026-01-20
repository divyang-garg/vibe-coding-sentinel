// Package feature_discovery - API endpoint utilities
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"regexp"
	"strings"
)

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
		middleware = discoverExpressMiddlewareInCodebase(codebasePath)
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
		auth = discoverExpressAuthInCodebase(codebasePath)
	case "fastapi":
		auth = discoverFastAPIAuth(codebasePath)
	case "gin", "chi":
		auth = discoverGoAuth(codebasePath, framework)
	}

	return auth
}
