// Package services - LLM configuration service
// Complies with CODING_STANDARDS.md: Business Services max 400 lines
package services

import (
	"context"
	"fmt"
	"strings"

	"sentinel-hub-api/models"
)

// LLMServiceImpl implements LLMService interface
type LLMServiceImpl struct {
	// Dependencies can be added here if needed
}

// NewLLMService creates a new LLM service
func NewLLMService() LLMService {
	return &LLMServiceImpl{}
}

// ValidateConfig validates LLM configuration
func (s *LLMServiceImpl) ValidateConfig(ctx context.Context, config models.LLMConfig) (*models.ValidateLLMConfigResponse, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var errors []string
	var warnings []string

	// Validate provider
	validProviders := map[string]bool{
		"openai":    true,
		"anthropic": true,
		"google":    true,
		"azure":     true,
	}
	if !validProviders[strings.ToLower(config.Provider)] {
		errors = append(errors, fmt.Sprintf("Invalid provider: %s. Must be one of: openai, anthropic, google, azure", config.Provider))
	}

	// Validate API key format
	if config.APIKey == "" {
		errors = append(errors, "API key is required")
	} else if len(config.APIKey) < 10 {
		errors = append(errors, "API key appears to be too short")
	}

	// Validate model
	if config.Model == "" {
		errors = append(errors, "Model is required")
	}

	// Validate cost optimization settings
	if config.CostOptimization.CacheTTLHours < 0 {
		errors = append(errors, "Cache TTL hours must be non-negative")
	} else if config.CostOptimization.CacheTTLHours > 168 { // 7 days
		warnings = append(warnings, "Cache TTL is very long (>7 days), consider shorter TTL for better cache freshness")
	}

	if config.CostOptimization.MaxCostPerRequest < 0 {
		errors = append(errors, "Max cost per request must be non-negative")
	} else if config.CostOptimization.MaxCostPerRequest > 100 {
		warnings = append(warnings, "Max cost per request is very high (>$100), consider a lower limit")
	}

	// Validate key type
	if config.KeyType != "" {
		validKeyTypes := map[string]bool{
			"api_key":         true,
			"service_account": true,
			"oauth":           true,
		}
		if !validKeyTypes[strings.ToLower(config.KeyType)] {
			warnings = append(warnings, fmt.Sprintf("Unknown key type: %s", config.KeyType))
		}
	}

	valid := len(errors) == 0

	return &models.ValidateLLMConfigResponse{
		Valid:    valid,
		Errors:   errors,
		Warnings: warnings,
	}, nil
}
