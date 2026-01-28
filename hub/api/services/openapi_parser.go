// Package services provides OpenAPI/Swagger contract parsing and validation using libopenapi.
//
// This package implements production-ready OpenAPI/Swagger contract validation with:
//   - Full support for OpenAPI 2.0, 3.0, 3.1, 3.2 specifications
//   - Automatic $ref reference resolution
//   - Deep schema validation (parameters, request body, responses, security)
//   - AST-based code schema extraction
//   - Framework-specific extractors (Go, Express.js, FastAPI)
//   - Caching for performance optimization
//
// Example usage:
//
//	ctx := context.Background()
//	contract, err := ParseOpenAPIContract(ctx, "openapi.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	endpoints := []EndpointInfo{...}
//	findings, err := validateAPIContracts(ctx, ".", endpoints)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, finding := range findings {
//		fmt.Printf("Issue: %s (severity: %s)\n", finding.Issue, finding.Severity)
//	}
//
// Complies with CODING_STANDARDS.md: Business services max 400 lines per file
package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// OpenAPIContract represents a parsed OpenAPI contract
type OpenAPIContract struct {
	Document  libopenapi.Document
	Version   string // "2.0", "3.0", "3.1", "3.2"
	FilePath  string
	Endpoints []ContractEndpoint
}

// ContractEndpoint represents an endpoint from the contract
type ContractEndpoint struct {
	Method      string
	Path        string
	OperationID string
	Parameters  []ContractParameter
	RequestBody *ContractRequestBody
	Responses   map[string]ContractResponse
	Security    []ContractSecurity
}

// ContractParameter represents a parameter definition
type ContractParameter struct {
	Name        string
	In          string // path, query, header, cookie
	Required    bool
	Schema      *base.Schema
	Description string
}

// ContractRequestBody represents request body schema
type ContractRequestBody struct {
	Required     bool
	ContentTypes map[string]*base.Schema // content-type -> schema
}

// ContractResponse represents response schema
type ContractResponse struct {
	Description  string
	ContentTypes map[string]*base.Schema
	Headers      map[string]*base.Schema
}

// ContractSecurity represents security requirement
type ContractSecurity struct {
	Schemes []string
	Scopes  []string
}

// ParseOpenAPIContract parses an OpenAPI/Swagger contract file
// Returns a parsed contract with all $ref references resolved
func ParseOpenAPIContract(ctx context.Context, filePath string) (*OpenAPIContract, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Read contract file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file %s: %w", filePath, err)
	}

	// Parse with libopenapi (handles YAML/JSON automatically)
	document, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI contract: %w", err)
	}

	// Extract version
	version := extractVersion(document)

	// Extract endpoints with $ref resolution
	endpoints, err := extractEndpoints(ctx, document, version)
	if err != nil {
		return nil, fmt.Errorf("failed to extract endpoints: %w", err)
	}

	return &OpenAPIContract{
		Document:  document,
		Version:   version,
		FilePath:  filePath,
		Endpoints: endpoints,
	}, nil
}

// extractVersion extracts OpenAPI version from document
func extractVersion(document libopenapi.Document) string {
	// Try OpenAPI 3.x first
	if model, err := document.BuildV3Model(); err == nil && model != nil {
		return model.Model.Version
	}

	// Try Swagger 2.0
	if model, err := document.BuildV2Model(); err == nil && model != nil {
		return model.Model.Swagger
	}

	return "unknown"
}

// extractEndpoints extracts all endpoints with resolved $refs
func extractEndpoints(ctx context.Context, document libopenapi.Document, version string) ([]ContractEndpoint, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var endpoints []ContractEndpoint

	// Handle OpenAPI 3.x
	if strings.HasPrefix(version, "3.") {
		v3Endpoints, err := extractOpenAPI3Endpoints(ctx, document)
		if err != nil {
			return nil, fmt.Errorf("failed to extract OpenAPI 3.x endpoints: %w", err)
		}
		endpoints = v3Endpoints
	} else if strings.HasPrefix(version, "2.") {
		// Handle Swagger 2.0
		v2Endpoints, err := extractSwagger2Endpoints(ctx, document)
		if err != nil {
			return nil, fmt.Errorf("failed to extract Swagger 2.0 endpoints: %w", err)
		}
		endpoints = v2Endpoints
	}

	return endpoints, nil
}

// extractOpenAPI3Endpoints and related functions are in openapi_parser_v3.go
// extractSwagger2Endpoints and related functions are in openapi_parser_v2.go
