// Package services provides helper functions for schema validation.
//
// This file contains file I/O and utility functions used by the schema validator.
// Separated to keep main schema_validator.go file under 400 lines.
//
// Complies with CODING_STANDARDS.md: Utilities max 250 lines per file
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// readEndpointSource reads the endpoint source file with context cancellation support.
func readEndpointSource(ctx context.Context, filePath string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Resolve file path (handle relative/absolute)
	absPath, err := resolveFilePath(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve file path %s: %w", filePath, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("endpoint source file not found: %s", absPath)
	}

	// Read file with context cancellation support
	data, err := readFileWithContext(ctx, absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read endpoint source file %s: %w", absPath, err)
	}

	return string(data), nil
}

// readFileWithContext reads a file with context cancellation support.
func readFileWithContext(ctx context.Context, filePath string) ([]byte, error) {
	// Use goroutine with context cancellation
	var data []byte
	var readErr error
	done := make(chan struct{})

	go func() {
		defer close(done)
		data, readErr = os.ReadFile(filePath)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return data, readErr
	}
}

// resolveFilePath resolves a file path to an absolute path.
func resolveFilePath(filePath string) (string, error) {
	if filepath.IsAbs(filePath) {
		return filePath, nil
	}

	// Try to resolve relative to current working directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve relative path: %w", err)
	}

	return absPath, nil
}

// detectLanguageFromFileForSchemaValidator detects the programming language from file extension.
// This is a wrapper around the existing detectLanguageFromFile in helpers.go to avoid redeclaration.
func detectLanguageFromFileForSchemaValidator(filePath string) string {
	// Use the existing function from helpers.go
	return detectLanguageFromFile(filePath)
}

// mapContractLocationToEndpointType maps OpenAPI parameter location to endpoint parameter type
func mapContractLocationToEndpointType(location string) string {
	switch location {
	case "path":
		return "path"
	case "query":
		return "query"
	case "header":
		return "header"
	case "cookie":
		return "cookie"
	case "body":
		return "body"
	default:
		return location
	}
}

// mapEndpointTypeToContractLocation maps endpoint parameter type to OpenAPI parameter location
func mapEndpointTypeToContractLocation(paramType string) string {
	switch paramType {
	case "path":
		return "path"
	case "query":
		return "query"
	case "header":
		return "header"
	case "cookie":
		return "cookie"
	case "body":
		return "body"
	default:
		return paramType
	}
}

// mapOpenAPITypeToEndpointType maps OpenAPI schema type to endpoint data type
func mapOpenAPITypeToEndpointType(openAPIType string) string {
	switch openAPIType {
	case "string":
		return "string"
	case "integer", "number":
		return "int"
	case "boolean":
		return "bool"
	case "array":
		return "array"
	case "object":
		return "object"
	default:
		return ""
	}
}

// normalizePathForJSONPath normalizes path for JSON path reference
func normalizePathForJSONPath(path string) string {
	// Replace path parameters with ~1 for JSON path encoding
	normalized := strings.ReplaceAll(path, "{", "~1")
	normalized = strings.ReplaceAll(normalized, "}", "")
	// Replace / with ~1
	normalized = strings.ReplaceAll(normalized, "/", "~1")
	return normalized
}

// getExpectedStatusCodes returns a string of expected status codes
func getExpectedStatusCodes(responses map[string]ContractResponse) string {
	codes := make([]string, 0, len(responses))
	for code := range responses {
		codes = append(codes, code)
	}
	return strings.Join(codes, ", ")
}

// getContentTypes returns a slice of content type strings from a content types map
func getContentTypes(contentTypes map[string]*base.Schema) []string {
	types := make([]string, 0, len(contentTypes))
	for ct := range contentTypes {
		types = append(types, ct)
	}
	return types
}

// validateSecurityMetadata provides fallback validation using endpoint metadata
// when AST analysis is unavailable.
func validateSecurityMetadata(ctx context.Context, endpoint EndpointInfo, contract ContractEndpoint, existingFindings []APILayerFinding) []APILayerFinding {
	findings := existingFindings

	// Use existing simplified validation as fallback
	hasSecurity := len(endpoint.Auth) > 0
	if !hasSecurity && len(contract.Security) > 0 {
		findings = append(findings, APILayerFinding{
			Type:         "contract_mismatch",
			Location:     endpoint.File,
			Issue:        fmt.Sprintf("Security requirements defined in contract but not found in endpoint metadata for %s %s", endpoint.Method, endpoint.Path),
			Severity:     "critical",
			ContractPath: fmt.Sprintf("#/paths/%s/%s/security", normalizePathForJSONPath(endpoint.Path), strings.ToLower(endpoint.Method)),
			SuggestedFix: fmt.Sprintf("Add security implementation to endpoint %s %s", endpoint.Method, endpoint.Path),
			Details: map[string]string{
				"validation_method": "metadata_fallback",
				"note":              "AST analysis unavailable, using metadata-based validation",
			},
		})
	}

	// Validate security schemes match (existing logic)
	for _, contractSec := range contract.Security {
		if ctx.Err() != nil {
			return findings
		}

		for _, scheme := range contractSec.Schemes {
			schemeFound := false
			for _, endpointAuth := range endpoint.Auth {
				if strings.EqualFold(endpointAuth, scheme) {
					schemeFound = true
					break
				}
			}

			if !schemeFound {
				findings = append(findings, APILayerFinding{
					Type:         "contract_mismatch",
					Location:     endpoint.File,
					Issue:        fmt.Sprintf("Security scheme '%s' required by contract but not found in endpoint %s %s", scheme, endpoint.Method, endpoint.Path),
					Severity:     "critical",
					ContractPath: fmt.Sprintf("#/paths/%s/%s/security/%s", normalizePathForJSONPath(endpoint.Path), strings.ToLower(endpoint.Method), scheme),
					SuggestedFix: fmt.Sprintf("Add security scheme '%s' to endpoint %s %s", scheme, endpoint.Method, endpoint.Path),
				})
			}
		}
	}

	return findings
}

// createMissingSecurityFinding creates a finding for missing security implementation.
func createMissingSecurityFinding(endpoint EndpointInfo, scheme string, contract ContractEndpoint) APILayerFinding {
	return APILayerFinding{
		Type:         "contract_mismatch",
		Location:     endpoint.File,
		Issue:        fmt.Sprintf("Security scheme '%s' required by contract but not detected in code for endpoint %s %s", scheme, endpoint.Method, endpoint.Path),
		Severity:     "critical",
		ContractPath: fmt.Sprintf("#/paths/%s/%s/security/%s", normalizePathForJSONPath(endpoint.Path), strings.ToLower(endpoint.Method), scheme),
		SuggestedFix: fmt.Sprintf("Add security scheme '%s' implementation to endpoint %s %s", scheme, endpoint.Method, endpoint.Path),
		Details: map[string]string{
			"validation_method": "ast_analysis",
			"scheme":            scheme,
		},
	}
}
