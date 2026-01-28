// Package code_extractors provides framework-specific code schema extraction
// Complies with CODING_STANDARDS.md: Business services max 400 lines per file
package code_extractors

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"sentinel-hub-api/services"
)

// ExtractGoSchema extracts schema from Go code (Gin, Echo frameworks)
func ExtractGoSchema(ctx context.Context, endpoint services.EndpointInfo) (*services.CodeSchema, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Read the Go file
	data, err := os.ReadFile(endpoint.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", endpoint.File, err)
	}

	// Parse Go file
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, endpoint.File, data, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	// Look for handler function
	var handlerFunc *ast.FuncDecl
	ast.Inspect(f, func(n ast.Node) bool {
		if ctx.Err() != nil {
			return false
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			// Match by handler name or function signature
			if endpoint.Handler != "" && fn.Name.Name == endpoint.Handler {
				handlerFunc = fn
				return false
			}
			// Or match by common handler patterns
			if strings.Contains(strings.ToLower(fn.Name.Name), "handler") ||
				strings.Contains(strings.ToLower(fn.Name.Name), strings.ToLower(endpoint.Method)) {
				handlerFunc = fn
				return false
			}
		}
		return true
	})

	if handlerFunc == nil {
		return nil, nil // Handler not found
	}

	// Extract request and response types from function signature
	schema := &services.CodeSchema{
		Type:       "object",
		Properties: make(map[string]services.CodeProperty),
		Required:   []string{},
	}

	// Analyze function parameters (request types)
	if handlerFunc.Type.Params != nil {
		for _, param := range handlerFunc.Type.Params.List {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			// Look for context, request struct, etc.
			if ident, ok := param.Type.(*ast.Ident); ok {
				if ident.Name == "Context" {
					continue // Skip context parameter
				}
			}

			// Extract struct type from parameter
			if structType := extractStructType(param.Type); structType != nil {
				for _, field := range structType.Fields.List {
					fieldName := ""
					if len(field.Names) > 0 {
						fieldName = field.Names[0].Name
					}

					if fieldName == "" {
						continue
					}

					property := services.CodeProperty{
						Type:        extractGoTypeString(field.Type),
						Constraints: make(map[string]interface{}),
					}

					// Check if required (not pointer)
					if !isPointerType(field.Type) {
						schema.Required = append(schema.Required, fieldName)
					}

					schema.Properties[fieldName] = property
				}
			}
		}
	}

	return schema, nil
}

// extractStructType extracts struct type from type expression
func extractStructType(expr ast.Expr) *ast.StructType {
	switch t := expr.(type) {
	case *ast.Ident:
		// Need to look up the type definition
		return nil
	case *ast.SelectorExpr:
		// Qualified type
		return nil
	case *ast.StarExpr:
		// Pointer - recurse
		return extractStructType(t.X)
	case *ast.StructType:
		return t
	}
	return nil
}

// extractGoTypeString extracts type string from Go type expression
func extractGoTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return mapGoTypeToOpenAPI(t.Name)
	case *ast.SelectorExpr:
		if t.Sel != nil {
			return mapGoTypeToOpenAPI(t.Sel.Name)
		}
	case *ast.ArrayType:
		return "array"
	case *ast.MapType:
		return "object"
	case *ast.StarExpr:
		return extractGoTypeString(t.X)
	}
	return "string"
}

// mapGoTypeToOpenAPI maps Go type to OpenAPI type
func mapGoTypeToOpenAPI(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "object"
	}
}

// isPointerType checks if type is a pointer
func isPointerType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}
