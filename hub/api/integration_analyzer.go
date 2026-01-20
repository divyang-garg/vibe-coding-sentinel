// Phase 14A: Integration Layer Analyzer
// Analyzes external API integrations for error handling, retry logic, and contracts

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

// IntegrationLayerFinding represents a finding from integration layer analysis
type IntegrationLayerFinding struct {
	Type     string `json:"type"`     // "missing_error_handling", "missing_retry", "contract_mismatch"
	Location string `json:"location"` // File path and line number
	Issue    string `json:"issue"`
	Severity string `json:"severity"` // "critical", "high", "medium", "low"
}

// analyzeIntegrationLayer analyzes external API integrations
func analyzeIntegrationLayer(ctx context.Context, feature *DiscoveredFeature) ([]IntegrationLayerFinding, error) {
	findings := []IntegrationLayerFinding{}

	if feature.IntegrationLayer == nil {
		return findings, nil
	}

	for _, integration := range feature.IntegrationLayer.Integrations {
		// Read integration file
		data, err := os.ReadFile(integration.File)
		if err != nil {
			LogWarn(ctx, "Failed to read integration file %s: %v", integration.File, err)
			continue
		}

		content := string(data)

		// Check for error handling
		errorHandlingFindings := checkIntegrationErrorHandling(content, integration)
		findings = append(findings, errorHandlingFindings...)

		// Check for retry logic
		retryFindings := checkRetryLogic(content, integration)
		findings = append(findings, retryFindings...)

		// Check for timeout handling
		timeoutFindings := checkTimeoutHandling(content, integration)
		findings = append(findings, timeoutFindings...)
	}

	// Validate integration contracts if codebase path available
	// NOTE: DiscoveredFeature.Context field not available - skipping contract validation
	// Integration contract validation would require codebase path from another source

	return findings, nil
}

// validateIntegrationContracts validates integration calls against documented contracts
func validateIntegrationContracts(ctx context.Context, codebasePath string, integrations []IntegrationInfo) ([]IntegrationLayerFinding, error) {
	findings := []IntegrationLayerFinding{}

	// Look for API contract documentation
	contractFiles := []string{
		"api-contracts.yaml",
		"api-contracts.json",
		"integration-contracts.yaml",
		"integration-contracts.json",
	}

	var contractData []byte
	var contractFile string
	for _, filename := range contractFiles {
		filepath := filepath.Join(codebasePath, filename)
		if data, err := os.ReadFile(filepath); err == nil {
			contractData = data
			contractFile = filepath
			break
		}
	}

	if contractData == nil {
		// No contract file found - not an error
		return findings, nil
	}

	// Parse contracts
	contractEndpoints := parseIntegrationContracts(contractData)

	// Compare actual integrations with documented contracts
	for _, integration := range integrations {
		contractEndpoint := findMatchingIntegrationContract(integration, contractEndpoints)

		if contractEndpoint == nil {
			findings = append(findings, IntegrationLayerFinding{
				Type:     "contract_mismatch",
				Location: integration.File,
				Issue:    fmt.Sprintf("Integration %s %s is not documented in %s", integration.Method, integration.Endpoint, contractFile),
				Severity: "medium",
			})
			continue
		}

		// Check if request/response formats match (simplified)
		if !integrationSchemasMatch(integration, *contractEndpoint) {
			findings = append(findings, IntegrationLayerFinding{
				Type:     "contract_mismatch",
				Location: integration.File,
				Issue:    fmt.Sprintf("Integration %s %s schema does not match contract", integration.Method, integration.Endpoint),
				Severity: "high",
			})
		}
	}

	return findings, nil
}

// checkRetryLogic checks for retry mechanisms in integration calls
func checkRetryLogic(content string, integration IntegrationInfo) []IntegrationLayerFinding {
	findings := []IntegrationLayerFinding{}

	// Check for retry libraries
	hasRetryLibrary := strings.Contains(content, "axios-retry") ||
		strings.Contains(content, "tenacity") ||
		strings.Contains(content, "retry") ||
		strings.Contains(content, "backoff")

	// Check for manual retry logic
	hasManualRetry := strings.Contains(content, "retry") &&
		(strings.Contains(content, "for") || strings.Contains(content, "while") || strings.Contains(content, "loop"))

	if !hasRetryLibrary && !hasManualRetry {
		findings = append(findings, IntegrationLayerFinding{
			Type:     "missing_retry",
			Location: fmt.Sprintf("%s:%d", integration.File, integration.LineNumber),
			Issue:    fmt.Sprintf("Integration call to %s may be missing retry logic", integration.Endpoint),
			Severity: "high",
		})
	}

	return findings
}

// checkIntegrationErrorHandling checks for error handling in integration calls
func checkIntegrationErrorHandling(content string, integration IntegrationInfo) []IntegrationLayerFinding {
	findings := []IntegrationLayerFinding{}

	// Check for error handling patterns
	hasErrorHandling := strings.Contains(content, "catch") ||
		strings.Contains(content, "error") ||
		strings.Contains(content, "err") ||
		strings.Contains(content, "exception")

	if !hasErrorHandling {
		findings = append(findings, IntegrationLayerFinding{
			Type:     "missing_error_handling",
			Location: fmt.Sprintf("%s:%d", integration.File, integration.LineNumber),
			Issue:    fmt.Sprintf("Integration call to %s may be missing error handling", integration.Endpoint),
			Severity: "critical",
		})
	}

	return findings
}

// checkTimeoutHandling checks for timeout configuration
func checkTimeoutHandling(content string, integration IntegrationInfo) []IntegrationLayerFinding {
	findings := []IntegrationLayerFinding{}

	// Check for timeout configuration
	hasTimeout := strings.Contains(content, "timeout") ||
		strings.Contains(content, "Timeout") ||
		strings.Contains(content, "TIMEOUT")

	if !hasTimeout {
		findings = append(findings, IntegrationLayerFinding{
			Type:     "missing_timeout",
			Location: fmt.Sprintf("%s:%d", integration.File, integration.LineNumber),
			Issue:    fmt.Sprintf("Integration call to %s may be missing timeout configuration", integration.Endpoint),
			Severity: "medium",
		})
	}

	return findings
}

// IntegrationContract represents a documented API contract
type IntegrationContract struct {
	Endpoint    string                 `json:"endpoint" yaml:"endpoint"`
	Method      string                 `json:"method" yaml:"method"`
	Request     map[string]interface{} `json:"request,omitempty" yaml:"request,omitempty"`
	Response    map[string]interface{} `json:"response,omitempty" yaml:"response,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty" yaml:"headers,omitempty"`
	QueryParams map[string]string      `json:"query_params,omitempty" yaml:"query_params,omitempty"`
}

// Helper functions

func parseIntegrationContracts(data []byte) []IntegrationContract {
	contracts := []IntegrationContract{}

	// Try JSON first
	var jsonContracts struct {
		Contracts []IntegrationContract `json:"contracts"`
		Endpoints []IntegrationContract `json:"endpoints"`
	}
	if err := json.Unmarshal(data, &jsonContracts); err == nil {
		if len(jsonContracts.Contracts) > 0 {
			contracts = jsonContracts.Contracts
		} else if len(jsonContracts.Endpoints) > 0 {
			contracts = jsonContracts.Endpoints
		}
		return contracts
	}

	// Try YAML
	var yamlContracts struct {
		Contracts []IntegrationContract `yaml:"contracts"`
		Endpoints []IntegrationContract `yaml:"endpoints"`
	}
	if err := yaml.Unmarshal(data, &yamlContracts); err == nil {
		if len(yamlContracts.Contracts) > 0 {
			contracts = yamlContracts.Contracts
		} else if len(yamlContracts.Endpoints) > 0 {
			contracts = yamlContracts.Endpoints
		}
		return contracts
	}

	// Try parsing as array directly
	var contractsArray []IntegrationContract
	if err := json.Unmarshal(data, &contractsArray); err == nil {
		return contractsArray
	}
	// Try YAML array parsing
	if err := yaml.Unmarshal(data, &contractsArray); err == nil {
		return contractsArray
	}

	return contracts
}

func findMatchingIntegrationContract(integration IntegrationInfo, contracts []IntegrationContract) *IntegrationContract {
	// Normalize endpoint for comparison
	normalizeEndpoint := func(endpoint string) string {
		endpoint = strings.TrimSpace(endpoint)
		endpoint = strings.TrimSuffix(endpoint, "/")
		return strings.ToLower(endpoint)
	}

	integrationEndpoint := normalizeEndpoint(integration.Endpoint)
	integrationMethod := strings.ToUpper(strings.TrimSpace(integration.Method))

	for i := range contracts {
		contractEndpoint := normalizeEndpoint(contracts[i].Endpoint)
		contractMethod := strings.ToUpper(strings.TrimSpace(contracts[i].Method))

		// Exact match
		if contractEndpoint == integrationEndpoint && contractMethod == integrationMethod {
			return &contracts[i]
		}

		// Partial match (endpoint matches, method is wildcard or empty)
		if contractEndpoint == integrationEndpoint && (contractMethod == "" || contractMethod == "*" || contractMethod == "ANY") {
			return &contracts[i]
		}

		// Pattern matching (contract endpoint contains wildcards)
		if strings.Contains(contractEndpoint, "*") {
			pattern := strings.ReplaceAll(contractEndpoint, "*", ".*")
			if matched, _ := filepath.Match(pattern, integrationEndpoint); matched {
				if contractMethod == "" || contractMethod == "*" || contractMethod == integrationMethod {
					return &contracts[i]
				}
			}
		}
	}

	return nil
}

func integrationSchemasMatch(integration IntegrationInfo, contract IntegrationContract) bool {
	// Read integration file to extract actual request/response
	content, err := os.ReadFile(integration.File)
	if err != nil {
		// Can't read file, assume match (conservative)
		return true
	}

	codeContent := string(content)

	// Check if request format matches
	if contract.Request != nil {
		// Look for request body patterns in code
		hasRequestBody := strings.Contains(codeContent, "body") ||
			strings.Contains(codeContent, "data") ||
			strings.Contains(codeContent, "payload") ||
			strings.Contains(codeContent, "request")

		// If contract requires request body but code doesn't have it, mismatch
		if required, ok := contract.Request["required"].(bool); ok && required && !hasRequestBody {
			return false
		}
	}

	// Check if response format matches
	if contract.Response != nil {
		// Look for response handling patterns
		hasResponseHandling := strings.Contains(codeContent, "response") ||
			strings.Contains(codeContent, "result") ||
			strings.Contains(codeContent, "data")

		// If contract defines response but code doesn't handle it, mismatch
		if required, ok := contract.Response["required"].(bool); ok && required && !hasResponseHandling {
			return false
		}
	}

	// Check headers if specified
	if len(contract.Headers) > 0 {
		for headerName, headerValue := range contract.Headers {
			// Check if header is present in code
			headerPattern := strings.ToLower(headerName)
			if !strings.Contains(strings.ToLower(codeContent), headerPattern) {
				// Header might be set via library, so this is a soft check
				// Only fail if it's a required header
				if strings.Contains(strings.ToLower(headerValue), "required") {
					return false
				}
			}
		}
	}

	return true
}
