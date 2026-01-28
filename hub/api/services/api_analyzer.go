// Phase 14A: API Layer Analyzer
// Analyzes API endpoints for security, validation, error handling, and contract compliance

package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// APILayerFinding represents a finding from API layer analysis
type APILayerFinding struct {
	Type         string            `json:"type"`          // "missing_auth", "missing_validation", "missing_error_handling", "contract_mismatch"
	Location     string            `json:"location"`      // File path and line number
	Issue        string            `json:"issue"`         // Description of the issue
	Severity     string            `json:"severity"`     // "critical", "high", "medium", "low"
	ContractPath string            `json:"contract_path,omitempty"` // JSON path in contract (e.g., #/paths/~1users/get/parameters/0)
	CodeLocation string            `json:"code_location,omitempty"` // Code location (file:line)
	SuggestedFix string            `json:"suggested_fix,omitempty"` // Suggested fix for the issue
	Details      map[string]string `json:"details,omitempty"`       // Additional details
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
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Find contract file
	contractFile, err := findContractFile(ctx, codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to find contract file: %w", err)
	}
	if contractFile == "" {
		// No contract found - not an error, just no validation
		return []APILayerFinding{}, nil
	}

	// Parse with libopenapi (using cache)
	contract, err := GetCachedContract(ctx, contractFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract: %w", err)
	}

	// Validate endpoints against contract using schema validator
	findings := []APILayerFinding{}
	for _, endpoint := range endpoints {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// validateEndpointAgainstContract is implemented in schema_validator.go
		validationFindings := validateEndpointAgainstContract(ctx, endpoint, contract)
		findings = append(findings, validationFindings...)
	}

	return findings, nil
}

// findContractFile locates OpenAPI/Swagger contract files
func findContractFile(ctx context.Context, codebasePath string) (string, error) {
	// Check context
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	openAPIFiles := []string{
		"openapi.yaml",
		"openapi.json",
		"swagger.yaml",
		"swagger.json",
	}

	for _, filename := range openAPIFiles {
		filePath := filepath.Join(codebasePath, filename)
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		}
	}

	return "", nil // Not found, but not an error
}

// validateEndpointAgainstContract is now implemented in schema_validator.go
// This function is kept for backward compatibility and delegates to the schema validator

// findMatchingContractEndpoint finds a matching endpoint in the contract
// Matches by HTTP method and path (with basic path parameter normalization)
func findMatchingContractEndpoint(endpoint EndpointInfo, contract *OpenAPIContract) *ContractEndpoint {
	endpointMethod := strings.ToUpper(endpoint.Method)
	endpointPath := normalizePath(endpoint.Path)

	for i := range contract.Endpoints {
		contractEndpoint := &contract.Endpoints[i]
		if strings.ToUpper(contractEndpoint.Method) == endpointMethod {
			contractPath := normalizePath(contractEndpoint.Path)
			if contractPath == endpointPath {
				return contractEndpoint
			}
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

