// Package services provides AST-based schema extraction from code
// Complies with CODING_STANDARDS.md: Business services max 400 lines per file
package services

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

// CodeSchema represents schema extracted from code
type CodeSchema struct {
	Type        string
	Properties  map[string]CodeProperty
	Required    []string
	Description string
}

// CodeProperty represents a schema property
type CodeProperty struct {
	Type        string
	Format      string
	Description string
	Constraints map[string]interface{} // min, max, pattern, enum
}

// ExtractRequestSchema extracts request body schema from code
// Currently supports Go structs - framework-specific extractors will be added
func ExtractRequestSchema(ctx context.Context, endpoint EndpointInfo) (*CodeSchema, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// For Go code, try to extract struct definitions
	if strings.HasSuffix(endpoint.File, ".go") {
		return extractGoRequestSchema(ctx, endpoint)
	}

	// For other languages, return nil (framework-specific extractors will handle)
	return nil, nil
}

// ExtractResponseSchema extracts response schema from code
// Currently supports Go structs - framework-specific extractors will be added
func ExtractResponseSchema(ctx context.Context, endpoint EndpointInfo) (*CodeSchema, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// For Go code, try to extract struct definitions
	if strings.HasSuffix(endpoint.File, ".go") {
		return extractGoResponseSchema(ctx, endpoint)
	}

	// For other languages, return nil (framework-specific extractors will handle)
	return nil, nil
}

// extractGoRequestSchema extracts request schema from Go code
func extractGoRequestSchema(ctx context.Context, endpoint EndpointInfo) (*CodeSchema, error) {
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

	// Look for struct definitions that might be request types
	// This is a simplified implementation - full implementation would analyze
	// handler function signatures to find request types
	var requestStruct *ast.StructType
	ast.Inspect(f, func(n ast.Node) bool {
		if ctx.Err() != nil {
			return false
		}

		if ts, ok := n.(*ast.TypeSpec); ok {
			if st, ok := ts.Type.(*ast.StructType); ok {
				// Check if struct name suggests it's a request type
				structName := ts.Name.Name
				if strings.Contains(strings.ToLower(structName), "request") ||
					strings.Contains(strings.ToLower(structName), "req") {
					requestStruct = st
					return false
				}
			}
		}
		return true
	})

	if requestStruct == nil {
		return nil, nil // No request struct found
	}

	// Convert struct to schema
	schema := &CodeSchema{
		Type:       "object",
		Properties: make(map[string]CodeProperty),
		Required:   []string{},
	}

	for _, field := range requestStruct.Fields.List {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		fieldName := ""
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		}

		if fieldName == "" {
			continue
		}

		// Extract field type
		fieldType := extractGoType(field.Type)
		property := CodeProperty{
			Type:        fieldType,
			Constraints: make(map[string]interface{}),
		}

		// Extract tags for validation constraints
		if field.Tag != nil {
			tag := field.Tag.Value
			// Parse JSON tag - remove backticks and parse
			tag = strings.Trim(tag, "`")
			if strings.Contains(tag, "json:") {
				// Extract JSON field name and required status
				jsonTag := extractJSONTag(tag)
				if jsonTag.Name != "" && jsonTag.Name != "-" {
					// Use JSON name if different from Go field name
					if jsonTag.Name != fieldName {
						property.Constraints["json_name"] = jsonTag.Name
					}
					// Check if field is omitempty (optional)
					if jsonTag.OmitEmpty {
						// Remove from required if omitempty
						for i, req := range schema.Required {
							if req == fieldName {
								schema.Required = append(schema.Required[:i], schema.Required[i+1:]...)
								break
							}
						}
					}
				}
			}
			// Parse validate tag if present (e.g., validate:"required,min=1")
			if strings.Contains(tag, "validate:") {
				validateTag := extractValidateTag(tag)
				if validateTag.Required {
					// Ensure field is in required list
					found := false
					for _, req := range schema.Required {
						if req == fieldName {
							found = true
							break
						}
					}
					if !found {
						schema.Required = append(schema.Required, fieldName)
					}
				}
				// Add validation constraints
				for k, v := range validateTag.Constraints {
					property.Constraints[k] = v
				}
			}
		}

		// Check if field is required (no pointer, no omitempty)
		if !isPointerType(field.Type) {
			schema.Required = append(schema.Required, fieldName)
		}

		schema.Properties[fieldName] = property
	}

	return schema, nil
}

// extractGoResponseSchema extracts response schema from Go code
func extractGoResponseSchema(ctx context.Context, endpoint EndpointInfo) (*CodeSchema, error) {
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

	// Look for handler function to find return type
	var responseType ast.Expr
	ast.Inspect(f, func(n ast.Node) bool {
		if ctx.Err() != nil {
			return false
		}

		if fn, ok := n.(*ast.FuncDecl); ok {
			// Check if function name matches handler
			if strings.Contains(strings.ToLower(fn.Name.Name), strings.ToLower(endpoint.Handler)) ||
				strings.Contains(strings.ToLower(fn.Name.Name), "handler") {
				if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
					// Get first return type (usually the response)
					responseType = fn.Type.Results.List[0].Type
					return false
				}
			}
		}
		return true
	})

	if responseType == nil {
		return nil, nil // No response type found
	}

	// Convert type to schema
	schema := &CodeSchema{
		Type:       extractGoType(responseType),
		Properties: make(map[string]CodeProperty),
		Required:   []string{},
	}

	// If it's a struct type, extract fields
	if st, ok := responseType.(*ast.StructType); ok {
		for _, field := range st.Fields.List {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			fieldName := ""
			if len(field.Names) > 0 {
				fieldName = field.Names[0].Name
			}

			if fieldName == "" {
				continue
			}

			fieldType := extractGoType(field.Type)
			property := CodeProperty{
				Type:        fieldType,
				Constraints: make(map[string]interface{}),
			}

			schema.Properties[fieldName] = property
		}
	}

	return schema, nil
}

// extractGoType extracts OpenAPI type from Go type expression
func extractGoType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return mapGoTypeToOpenAPIType(t.Name)
	case *ast.SelectorExpr:
		// Handle qualified types like time.Time
		if t.Sel != nil {
			return mapGoTypeToOpenAPIType(t.Sel.Name)
		}
	case *ast.ArrayType:
		return "array"
	case *ast.MapType:
		return "object"
	case *ast.StarExpr:
		// Pointer type - recurse
		return extractGoType(t.X)
	case *ast.InterfaceType:
		return "object"
	}
	return "string" // Default
}

// mapGoTypeToOpenAPIType maps Go type to OpenAPI type
func mapGoTypeToOpenAPIType(goType string) string {
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
	case "Time":
		return "string" // With format: date-time
	default:
		return "object" // Struct or custom type
	}
}

// isPointerType checks if a type expression is a pointer
func isPointerType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

// jsonTagInfo represents parsed JSON tag information
type jsonTagInfo struct {
	Name      string
	OmitEmpty bool
}

// extractJSONTag extracts JSON tag information from struct tag
func extractJSONTag(tag string) jsonTagInfo {
	info := jsonTagInfo{}
	
	// Find json: tag
	jsonIdx := strings.Index(tag, "json:")
	if jsonIdx == -1 {
		return info
	}
	
	// Extract the value after json:
	start := jsonIdx + 5 // len("json:")
	end := start
	for end < len(tag) && tag[end] != ' ' && tag[end] != '"' && tag[end] != '`' {
		end++
	}
	
	jsonValue := tag[start:end]
	jsonValue = strings.Trim(jsonValue, `"`)
	
	// Parse JSON tag value (format: "name,omitempty" or just "name")
	parts := strings.Split(jsonValue, ",")
	if len(parts) > 0 {
		info.Name = strings.TrimSpace(parts[0])
	}
	
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "omitempty" {
			info.OmitEmpty = true
		}
	}
	
	return info
}

// validateTagInfo represents parsed validate tag information
type validateTagInfo struct {
	Required   bool
	Constraints map[string]interface{}
}

// extractValidateTag extracts validate tag information from struct tag
func extractValidateTag(tag string) validateTagInfo {
	info := validateTagInfo{
		Constraints: make(map[string]interface{}),
	}
	
	// Find validate: tag
	validateIdx := strings.Index(tag, "validate:")
	if validateIdx == -1 {
		return info
	}
	
	// Extract the value after validate:
	start := validateIdx + 9 // len("validate:")
	end := start
	for end < len(tag) && tag[end] != ' ' && tag[end] != '"' && tag[end] != '`' {
		end++
	}
	
	validateValue := tag[start:end]
	validateValue = strings.Trim(validateValue, `"`)
	
	// Parse validate tag value (format: "required,min=1,max=100")
	parts := strings.Split(validateValue, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "required" {
			info.Required = true
		} else if strings.Contains(part, "=") {
			// Extract constraint (e.g., "min=1" -> {min: 1})
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				// Try to parse as number
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					info.Constraints[key] = num
				} else {
					info.Constraints[key] = value
				}
			}
		} else if part != "" {
			// Boolean constraint (e.g., "email", "url")
			info.Constraints[part] = true
		}
	}
	
	return info
}
