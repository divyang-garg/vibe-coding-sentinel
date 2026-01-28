// Package code_extractors provides FastAPI code schema extraction
// Complies with CODING_STANDARDS.md: Business services max 400 lines per file
package code_extractors

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"sentinel-hub-api/services"
)

// ExtractFastAPISchema extracts schema from FastAPI code
// Uses pattern matching to find Pydantic models and function signatures
func ExtractFastAPISchema(ctx context.Context, endpoint services.EndpointInfo) (*services.CodeSchema, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Read the Python file
	data, err := os.ReadFile(endpoint.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", endpoint.File, err)
	}

	code := string(data)

	// Look for Pydantic models
	schema := &services.CodeSchema{
		Type:       "object",
		Properties: make(map[string]services.CodeProperty),
		Required:   []string{},
	}

	// Pattern for Pydantic BaseModel
	modelPattern := regexp.MustCompile(`class\s+(\w+)\s*\([^)]*BaseModel[^)]*\):\s*([^}]+)`)
	matches := modelPattern.FindStringSubmatch(code)

	if len(matches) > 2 {
		modelBody := matches[2]

		// Extract field definitions
		fieldPattern := regexp.MustCompile(`(\w+):\s*(\w+)(?:\[([^\]]+)\])?`)
		fieldMatches := fieldPattern.FindAllStringSubmatch(modelBody, -1)

		for _, match := range fieldMatches {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			fieldName := match[1]
			fieldType := match[2]
			isOptional := strings.Contains(modelBody, fieldName+": Optional") ||
				strings.Contains(modelBody, fieldName+": "+fieldType+" | None")

			property := services.CodeProperty{
				Type:        mapPythonTypeToOpenAPI(fieldType),
				Constraints: make(map[string]interface{}),
			}

			if !isOptional {
				schema.Required = append(schema.Required, fieldName)
			}

			schema.Properties[fieldName] = property
		}
	}

	// Also look for function parameters with type hints
	funcPattern := regexp.MustCompile(`def\s+\w+\s*\([^)]*(\w+):\s*(\w+)[^)]*\)`)
	funcMatches := funcPattern.FindAllStringSubmatch(code, -1)

	for _, match := range funcMatches {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		paramName := match[1]
		paramType := match[2]

		// Skip common parameters
		if paramName == "request" || paramName == "response" || paramName == "db" {
			continue
		}

		property := services.CodeProperty{
			Type:        mapPythonTypeToOpenAPI(paramType),
			Constraints: make(map[string]interface{}),
		}

		schema.Properties[paramName] = property
		schema.Required = append(schema.Required, paramName)
	}

	if len(schema.Properties) == 0 {
		return nil, nil // No schema found
	}

	return schema, nil
}

// mapPythonTypeToOpenAPI maps Python type to OpenAPI type
func mapPythonTypeToOpenAPI(pythonType string) string {
	switch strings.ToLower(pythonType) {
	case "str", "string":
		return "string"
	case "int", "integer":
		return "integer"
	case "float", "double":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "list", "array":
		return "array"
	case "dict", "object":
		return "object"
	default:
		return "object" // Assume custom class/object
	}
}
