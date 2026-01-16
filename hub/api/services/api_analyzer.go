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

func parseOpenAPIContract(data []byte) []map[string]interface{} {
	// Simplified parsing - would use proper OpenAPI parser in production
	// For now, return empty
	return []map[string]interface{}{}
}

func findMatchingContractEndpoint(endpoint EndpointInfo, contractEndpoints []map[string]interface{}) map[string]interface{} {
	// Simplified matching - would do proper path matching in production
	return nil
}

func schemasMatch(endpoint EndpointInfo, contractEndpoint map[string]interface{}) bool {
	// Simplified - would do proper schema comparison in production
	return true
}
