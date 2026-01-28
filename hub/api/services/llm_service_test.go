// Package services - Unit tests for LLM service
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package services

import (
	"context"
	"testing"

	"sentinel-hub-api/models"

	"github.com/stretchr/testify/assert"
)

func TestLLMService_ValidateConfig_Valid(t *testing.T) {
	service := NewLLMService()

	config := models.LLMConfig{
		Provider: "openai",
		APIKey:   "sk-test12345678901234567890",
		Model:    "gpt-4",
		KeyType:  "api_key",
		CostOptimization: models.CostOptimizationConfig{
			UseCache:         true,
			CacheTTLHours:    24,
			ProgressiveDepth: true,
		},
	}

	result, err := service.ValidateConfig(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestLLMService_ValidateConfig_InvalidProvider(t *testing.T) {
	service := NewLLMService()

	config := models.LLMConfig{
		Provider: "invalid_provider",
		APIKey:   "sk-test12345678901234567890",
		Model:    "gpt-4",
	}

	result, err := service.ValidateConfig(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestLLMService_ValidateConfig_MissingAPIKey(t *testing.T) {
	service := NewLLMService()

	config := models.LLMConfig{
		Provider: "openai",
		Model:    "gpt-4",
	}

	result, err := service.ValidateConfig(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "API key")
}

func TestLLMService_ValidateConfig_ShortAPIKey(t *testing.T) {
	service := NewLLMService()

	config := models.LLMConfig{
		Provider: "openai",
		APIKey:   "short",
		Model:    "gpt-4",
	}

	result, err := service.ValidateConfig(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestLLMService_ValidateConfig_MissingModel(t *testing.T) {
	service := NewLLMService()

	config := models.LLMConfig{
		Provider: "openai",
		APIKey:   "sk-test12345678901234567890",
	}

	result, err := service.ValidateConfig(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors[0], "Model")
}

func TestLLMService_ValidateConfig_InvalidCacheTTL(t *testing.T) {
	service := NewLLMService()

	config := models.LLMConfig{
		Provider: "openai",
		APIKey:   "sk-test12345678901234567890",
		Model:    "gpt-4",
		CostOptimization: models.CostOptimizationConfig{
			CacheTTLHours: -1,
		},
	}

	result, err := service.ValidateConfig(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestLLMService_ValidateConfig_LongCacheTTLWarning(t *testing.T) {
	service := NewLLMService()

	config := models.LLMConfig{
		Provider: "openai",
		APIKey:   "sk-test12345678901234567890",
		Model:    "gpt-4",
		CostOptimization: models.CostOptimizationConfig{
			CacheTTLHours: 200, // > 7 days
		},
	}

	result, err := service.ValidateConfig(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.NotEmpty(t, result.Warnings)
}

func TestLLMService_ValidateConfig_HighCostWarning(t *testing.T) {
	service := NewLLMService()

	config := models.LLMConfig{
		Provider: "openai",
		APIKey:   "sk-test12345678901234567890",
		Model:    "gpt-4",
		CostOptimization: models.CostOptimizationConfig{
			MaxCostPerRequest: 150.0, // > $100
		},
	}

	result, err := service.ValidateConfig(context.Background(), config)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Valid)
	assert.NotEmpty(t, result.Warnings)
}

func TestLLMService_ValidateConfig_ContextCancellation(t *testing.T) {
	service := NewLLMService()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	config := models.LLMConfig{
		Provider: "openai",
		APIKey:   "sk-test12345678901234567890",
		Model:    "gpt-4",
	}

	result, err := service.ValidateConfig(ctx, config)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, context.Canceled, err)
}

func TestLLMService_ValidateConfig_AllProviders(t *testing.T) {
	service := NewLLMService()

	providers := []string{"openai", "anthropic", "google", "azure"}

	for _, provider := range providers {
		config := models.LLMConfig{
			Provider: provider,
			APIKey:   "sk-test12345678901234567890",
			Model:    "gpt-4",
		}

		result, err := service.ValidateConfig(context.Background(), config)

		assert.NoError(t, err, "Provider: %s", provider)
		assert.NotNil(t, result, "Provider: %s", provider)
		assert.True(t, result.Valid, "Provider: %s", provider)
	}
}
