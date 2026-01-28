// Phase 14A: API Layer Analyzer
// Analyzes API endpoints for security, validation, error handling, and contract compliance

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// APILayerFinding represents a finding from API layer analysis
type APILayerFinding struct {
	Type     string `json:"type"`     // "missing_auth", "missing_validation", "missing_error_handling", "contract_mismatch"
	Location string `json:"location"` // File path and line number
	Issue    string `json:"issue"`
	Severity string `json:"severity"` // "critical", "high", "medium", "low"
}

// analyzeAPILayer analyzes API endpoints for various issues
func analyzeAPILayer(ctx context.Context, feature *DiscoveredFeature) ([]APILayerFinding, error) {
	findings := []APILayerFinding{}

	if feature.APILayer == nil {
		return findings, nil
	}

	// Reuse security analyzer for security checks
	for _, endpoint := range feature.APILayer.Endpoints {
		// Read endpoint file
		data, err := os.ReadFile(endpoint.File)
		if err != nil {
			LogWarn(ctx, "Failed to read endpoint file %s: %v", endpoint.File, err)
			continue
		}

		// Analyze security (reuse security analyzer)
		securityFindings := analyzeEndpointSecurity(string(data), endpoint)
		findings = append(findings, securityFindings...)

		// Check for input validation
		validationFindings := checkInputValidation(string(data), endpoint)
		findings = append(findings, validationFindings...)

		// Check for error handling
		errorHandlingFindings := checkErrorHandling(string(data), endpoint)
		findings = append(findings, errorHandlingFindings...)

		// Check status code usage
		statusCodeFindings := checkStatusCodeUsage(string(data), endpoint)
		findings = append(findings, statusCodeFindings...)
	}

	return findings, nil
}

// validateAPIContracts validates API endpoints against OpenAPI/Swagger contracts
func validateAPIContracts(ctx context.Context, codebasePath string, endpoints []EndpointInfo) ([]APILayerFinding, error) {
	// Check if context is cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	findings := []APILayerFinding{}

	// Look for OpenAPI/Swagger files
	openAPIFiles := []string{
		"openapi.yaml",
		"openapi.json",
		"swagger.yaml",
		"swagger.json",
	}

	var contractData []byte
	var contractFile string
	for _, filename := range openAPIFiles {
		// Check if context is cancelled before each file read
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		filepath := filepath.Join(codebasePath, filename)
		if data, err := os.ReadFile(filepath); err == nil {
			contractData = data
			contractFile = filepath
			break
		}
	}

	if contractData == nil {
		// No contract file found - not an error, just no validation
		return findings, nil
	}

	// Parse contract (simplified - would use proper OpenAPI parser in production)
	contractEndpoints := parseOpenAPIContract(contractData)

	// Compare actual endpoints with documented contracts
	for _, endpoint := range endpoints {
		// Check if context is cancelled before processing each endpoint
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Find matching contract endpoint
		contractEndpoint := findMatchingContractEndpoint(endpoint, contractEndpoints)

		if contractEndpoint == nil {
			findings = append(findings, APILayerFinding{
				Type:     "contract_mismatch",
				Location: endpoint.File,
				Issue:    fmt.Sprintf("Endpoint %s %s is not documented in %s", endpoint.Method, endpoint.Path, contractFile),
				Severity: "medium",
			})
			continue
		}

		// Check if request/response schemas match (simplified)
		if !schemasMatch(endpoint, contractEndpoint) {
			findings = append(findings, APILayerFinding{
				Type:     "contract_mismatch",
				Location: endpoint.File,
				Issue:    fmt.Sprintf("Endpoint %s %s schema does not match contract", endpoint.Method, endpoint.Path),
				Severity: "high",
			})
		}
	}

	return findings, nil
}

// Helper functions

func analyzeEndpointSecurity(code string, endpoint EndpointInfo) []APILayerFinding {
	findings := []APILayerFinding{}

	// Check for authentication middleware
	// This is framework-specific
	if strings.Contains(code, "app.") || strings.Contains(code, "router.") {
		// Express.js patterns
		if !strings.Contains(code, "auth") && !strings.Contains(code, "authenticate") && !strings.Contains(code, "jwt") {
			findings = append(findings, APILayerFinding{
				Type:     "missing_auth",
				Location: endpoint.File,
				Issue:    fmt.Sprintf("Endpoint %s %s may be missing authentication", endpoint.Method, endpoint.Path),
				Severity: "critical",
			})
		}
	}

	return findings
}

func checkInputValidation(code string, endpoint EndpointInfo) []APILayerFinding {
	findings := []APILayerFinding{}

	// Check for validation middleware/libraries
	if !strings.Contains(code, "validate") && !strings.Contains(code, "validator") && !strings.Contains(code, "joi") && !strings.Contains(code, "zod") {
		findings = append(findings, APILayerFinding{
			Type:     "missing_validation",
			Location: endpoint.File,
			Issue:    fmt.Sprintf("Endpoint %s %s may be missing input validation", endpoint.Method, endpoint.Path),
			Severity: "high",
		})
	}

	return findings
}

func checkErrorHandling(code string, endpoint EndpointInfo) []APILayerFinding {
	findings := []APILayerFinding{}

	// Check for error handling patterns
	if !strings.Contains(code, "try") && !strings.Contains(code, "catch") && !strings.Contains(code, "error") && !strings.Contains(code, "err") {
		findings = append(findings, APILayerFinding{
			Type:     "missing_error_handling",
			Location: endpoint.File,
			Issue:    fmt.Sprintf("Endpoint %s %s may be missing error handling", endpoint.Method, endpoint.Path),
			Severity: "high",
		})
	}

	return findings
}

func checkStatusCodeUsage(code string, endpoint EndpointInfo) []APILayerFinding {
	findings := []APILayerFinding{}

	// Check for appropriate status codes
	hasStatusCodes := strings.Contains(code, "200") || strings.Contains(code, "201") || strings.Contains(code, "400") || strings.Contains(code, "404") || strings.Contains(code, "500")

	if !hasStatusCodes {
		findings = append(findings, APILayerFinding{
			Type:     "missing_status_codes",
			Location: endpoint.File,
			Issue:    fmt.Sprintf("Endpoint %s %s may not be using appropriate HTTP status codes", endpoint.Method, endpoint.Path),
			Severity: "medium",
		})
	}

	return findings
}

// parseOpenAPIContract parses OpenAPI/Swagger contract files and extracts endpoint information
// Supports both OpenAPI 3.0 and Swagger 2.0 formats
func parseOpenAPIContract(data []byte) []map[string]interface{} {
	var contract map[string]interface{}
	
	// Try YAML first (most common format)
	if err := yaml.Unmarshal(data, &contract); err != nil {
		// Fall back to JSON
		if err := json.Unmarshal(data, &contract); err != nil {
			// If both fail, return empty
			return []map[string]interface{}{}
		}
	}

	var endpoints []map[string]interface{}
	
	// Check OpenAPI version
	openAPIVersion, hasOpenAPI := contract["openapi"].(string)
	swaggerVersion, hasSwagger := contract["swagger"].(string)
	
	if hasOpenAPI && strings.HasPrefix(openAPIVersion, "3.") {
		// OpenAPI 3.0 format
		endpoints = parseOpenAPI3(contract)
	} else if hasSwagger && strings.HasPrefix(swaggerVersion, "2.") {
		// Swagger 2.0 format
		endpoints = parseSwagger2(contract)
	}
	
	return endpoints
}

// parseOpenAPI3 extracts endpoints from OpenAPI 3.0 specification
func parseOpenAPI3(contract map[string]interface{}) []map[string]interface{} {
	var endpoints []map[string]interface{}
	
	paths, ok := contract["paths"].(map[string]interface{})
	if !ok {
		return endpoints
	}
	
	for path, pathItem := range paths {
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}
		
		// OpenAPI 3.0 uses HTTP methods as keys (get, post, put, delete, patch, etc.)
		for method, operation := range pathItemMap {
			// Skip non-method keys like parameters, servers, etc.
			if !isHTTPMethod(method) {
				continue
			}
			
			operationMap, ok := operation.(map[string]interface{})
			if !ok {
				continue
			}
			
			endpoint := map[string]interface{}{
				"method": strings.ToUpper(method),
				"path":   path,
			}
			
			// Extract operation ID if available
			if operationID, ok := operationMap["operationId"].(string); ok {
				endpoint["operationId"] = operationID
			}
			
			// Extract summary/description
			if summary, ok := operationMap["summary"].(string); ok {
				endpoint["summary"] = summary
			}
			if description, ok := operationMap["description"].(string); ok {
				endpoint["description"] = description
			}
			
			// Extract parameters
			if parameters, ok := operationMap["parameters"].([]interface{}); ok {
				endpoint["parameters"] = parameters
			}
			
			// Extract responses
			if responses, ok := operationMap["responses"].(map[string]interface{}); ok {
				endpoint["responses"] = responses
			}
			
			endpoints = append(endpoints, endpoint)
		}
	}
	
	return endpoints
}

// parseSwagger2 extracts endpoints from Swagger 2.0 specification
func parseSwagger2(contract map[string]interface{}) []map[string]interface{} {
	var endpoints []map[string]interface{}
	
	paths, ok := contract["paths"].(map[string]interface{})
	if !ok {
		return endpoints
	}
	
	for path, pathItem := range paths {
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Swagger 2.0 uses HTTP methods as keys (get, post, put, delete, patch, etc.)
		for method, operation := range pathItemMap {
			// Skip non-method keys like parameters
			if !isHTTPMethod(method) {
				continue
			}
			
			operationMap, ok := operation.(map[string]interface{})
			if !ok {
				continue
			}
			
			endpoint := map[string]interface{}{
				"method": strings.ToUpper(method),
				"path":   path,
			}
			
			// Extract operation ID if available
			if operationID, ok := operationMap["operationId"].(string); ok {
				endpoint["operationId"] = operationID
			}
			
			// Extract summary/description
			if summary, ok := operationMap["summary"].(string); ok {
				endpoint["summary"] = summary
			}
			if description, ok := operationMap["description"].(string); ok {
				endpoint["description"] = description
			}
			
			// Extract parameters
			if parameters, ok := operationMap["parameters"].([]interface{}); ok {
				endpoint["parameters"] = parameters
			}
			
			// Extract responses
			if responses, ok := operationMap["responses"].(map[string]interface{}); ok {
				endpoint["responses"] = responses
			}
			
			endpoints = append(endpoints, endpoint)
		}
	}
	
	return endpoints
}

// isHTTPMethod checks if a string is a valid HTTP method
func isHTTPMethod(method string) bool {
	methods := map[string]bool{
		"get":     true,
		"post":    true,
		"put":     true,
		"delete":  true,
		"patch":   true,
		"head":    true,
		"options": true,
		"trace":   true,
	}
	return methods[strings.ToLower(method)]
}

// findMatchingContractEndpoint finds a matching endpoint in the contract
// Matches by HTTP method and path (with basic path parameter normalization)
func findMatchingContractEndpoint(endpoint EndpointInfo, contractEndpoints []map[string]interface{}) map[string]interface{} {
	endpointMethod := strings.ToUpper(endpoint.Method)
	endpointPath := normalizePath(endpoint.Path)
	
	for _, contractEndpoint := range contractEndpoints {
		contractMethod, ok := contractEndpoint["method"].(string)
		if !ok {
			continue
		}
		
		contractPath, ok := contractEndpoint["path"].(string)
		if !ok {
			continue
		}
		
		// Normalize contract path for comparison
		normalizedContractPath := normalizePath(contractPath)
		
		// Match method and path
		if strings.ToUpper(contractMethod) == endpointMethod && normalizedContractPath == endpointPath {
			return contractEndpoint
		}
	}
	
	return nil
}

// normalizePath normalizes API paths for comparison
// Converts path parameters to a common format (e.g., /users/{id} and /users/:id both become /users/{id})
func normalizePath(path string) string {
	if path == "" {
		return path
	}
	
	// Replace :param with {param} (Express.js style to OpenAPI style)
	normalized := path
	for {
		idx := strings.Index(normalized, ":")
		if idx == -1 || idx == len(normalized)-1 {
			break
		}
		
		// Find the end of the parameter (next / or end of string)
		end := idx + 1
		for end < len(normalized) && normalized[end] != '/' {
			end++
		}
		
		paramName := normalized[idx+1 : end]
		normalized = normalized[:idx] + "{" + paramName + "}" + normalized[end:]
	}
	
	return normalized
}

// schemasMatch compares endpoint schemas with contract schemas
// Currently performs basic validation - full schema comparison would require deeper analysis
func schemasMatch(endpoint EndpointInfo, contractEndpoint map[string]interface{}) bool {
	// Basic validation: check if contract has response definitions
	responses, ok := contractEndpoint["responses"].(map[string]interface{})
	if !ok || len(responses) == 0 {
		// Contract doesn't define responses - consider it a mismatch
		return false
	}
	
	// Check if endpoint has expected response codes
	// For now, we consider it a match if contract has responses defined
	// Full implementation would compare request/response schemas in detail
	
	// If endpoint has response info, do basic comparison
	if len(endpoint.Responses) > 0 {
		// Check if any endpoint response codes match contract response codes
		for _, endpointResponse := range endpoint.Responses {
			statusCodeStr := fmt.Sprintf("%d", endpointResponse.StatusCode)
			if _, exists := responses[statusCodeStr]; !exists {
				// Endpoint returns a status code not in contract
				return false
			}
		}
	}
	
	return true
}
