// Package hub provides HTTP client for communicating with Sentinel Hub API
// Complies with CODING_STANDARDS.md: Client/integration max 300 lines
package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client provides methods to communicate with Sentinel Hub API
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	timeout    time.Duration
}

// NewClient creates a new Hub API client
func NewClient(baseURL, apiKey string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default Hub URL
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  apiKey,
		timeout: 30 * time.Second,
	}
}

// IsAvailable checks if the Hub is reachable
func (c *Client) IsAvailable() bool {
	if c.baseURL == "" {
		return false
	}

	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// AnalyzeAST sends code to Hub for AST analysis
func (c *Client) AnalyzeAST(req *ASTAnalysisRequest) (*ASTAnalysisResponse, error) {
	if !c.IsAvailable() {
		return nil, fmt.Errorf("hub not available")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/ast/analyze",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hub returned status %d", resp.StatusCode)
	}

	var result ASTAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AnalyzeVibe sends code to Hub for vibe coding detection
func (c *Client) AnalyzeVibe(req *VibeAnalysisRequest) (*VibeAnalysisResponse, error) {
	if !c.IsAvailable() {
		return nil, fmt.Errorf("hub not available")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/vibe/analyze",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hub returned status %d", resp.StatusCode)
	}

	var result VibeAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// AnalyzeStructure sends code to Hub for file structure analysis
func (c *Client) AnalyzeStructure(req *StructureAnalysisRequest) (*StructureAnalysisResponse, error) {
	if !c.IsAvailable() {
		return nil, fmt.Errorf("hub not available")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/structure/analyze",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hub returned status %d", resp.StatusCode)
	}

	var result StructureAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetHookPolicy retrieves hook execution policy from Hub
func (c *Client) GetHookPolicy(orgID string) (*HookPolicy, error) {
	if !c.IsAvailable() {
		// Return default policy when Hub unavailable
		return &HookPolicy{
			AuditEnabled:         true,
			VibeCheckEnabled:     true,
			SecurityCheckEnabled: true,
			FileSizeCheckEnabled: true,
			AllowOverride:        true,
			MaxOverridesPerDay:   10,
		}, nil
	}

	resp, err := c.httpClient.Get(
		fmt.Sprintf("%s/api/v1/hooks/policies?org_id=%s", c.baseURL, orgID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hub returned status %d", resp.StatusCode)
	}

	var policy HookPolicy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &policy, nil
}

// SendTelemetry sends telemetry data to Hub
func (c *Client) SendTelemetry(data *TelemetryData) error {
	if !c.IsAvailable() {
		// Silently fail when Hub unavailable
		return nil
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/v1/telemetry",
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		// Don't fail the operation if telemetry fails
		return nil
	}
	defer resp.Body.Close()

	return nil
}
