// Package code_extractors provides Express.js code schema extraction
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

// ExtractExpressSchema extracts schema from Express.js code
// Uses pattern matching to find request/response types
func ExtractExpressSchema(ctx context.Context, endpoint services.EndpointInfo) (*services.CodeSchema, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Read the JavaScript/TypeScript file
	data, err := os.ReadFile(endpoint.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", endpoint.File, err)
	}

	code := string(data)

	// Look for request body validation (Joi, Zod, etc.)
	schema := &services.CodeSchema{
		Type:       "object",
		Properties: make(map[string]services.CodeProperty),
		Required:   []string{},
	}

	// Pattern for Joi validation
	joiPattern := regexp.MustCompile(`Joi\.object\([^)]*\)\.keys\(([^)]+)\)`)
	if matches := joiPattern.FindStringSubmatch(code); len(matches) > 1 {
		// Extract field definitions from Joi schema
		fields := matches[1]
		fieldPattern := regexp.MustCompile(`(\w+):\s*Joi\.(\w+)\(\)`)
		fieldMatches := fieldPattern.FindAllStringSubmatch(fields, -1)

		for _, match := range fieldMatches {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			fieldName := match[1]
			fieldType := match[2]

			property := services.CodeProperty{
				Type:        mapJoiTypeToOpenAPI(fieldType),
				Constraints: make(map[string]interface{}),
			}

			// Check if required (look for .required())
			if strings.Contains(fields, fieldName+".required()") {
				schema.Required = append(schema.Required, fieldName)
			}

			schema.Properties[fieldName] = property
		}
	}

	// Pattern for Zod validation
	zodPattern := regexp.MustCompile(`z\.object\([^)]*\)\.shape\(([^)]+)\)`)
	if matches := zodPattern.FindStringSubmatch(code); len(matches) > 1 {
		// Extract field definitions from Zod schema
		fields := matches[1]
		fieldPattern := regexp.MustCompile(`(\w+):\s*z\.(\w+)\(\)`)
		fieldMatches := fieldPattern.FindAllStringSubmatch(fields, -1)

		for _, match := range fieldMatches {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			fieldName := match[1]
			fieldType := match[2]

			property := services.CodeProperty{
				Type:        mapZodTypeToOpenAPI(fieldType),
				Constraints: make(map[string]interface{}),
			}

			schema.Properties[fieldName] = property
		}
	}

	if len(schema.Properties) == 0 {
		return nil, nil // No schema found
	}

	return schema, nil
}

// mapJoiTypeToOpenAPI maps Joi type to OpenAPI type
func mapJoiTypeToOpenAPI(joiType string) string {
	switch strings.ToLower(joiType) {
	case "string":
		return "string"
	case "number":
		return "number"
	case "integer", "int":
		return "integer"
	case "boolean", "bool":
		return "boolean"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "string"
	}
}

// mapZodTypeToOpenAPI maps Zod type to OpenAPI type
func mapZodTypeToOpenAPI(zodType string) string {
	switch strings.ToLower(zodType) {
	case "string":
		return "string"
	case "number":
		return "number"
	case "int":
		return "integer"
	case "boolean", "bool":
		return "boolean"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return "string"
	}
}
