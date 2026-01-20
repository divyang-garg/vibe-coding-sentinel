// Package config provides extraction configuration
// Complies with CODING_STANDARDS.md: Config modules max 200 lines
package config

import "fmt"

// ExtractionConfig holds extraction-specific configuration
type ExtractionConfig struct {
	LLMProvider              string
	LLMModel                 string
	LLMAPIKey                string
	CacheDir                 string
	CacheTTLHours            int
	BatchSize                int
	MaxRetries               int
	CircuitBreakerThreshold  int
	CircuitBreakerTimeoutSec int
}

// LoadExtractionConfig loads extraction configuration from environment
func LoadExtractionConfig() (*ExtractionConfig, error) {
	cfg := &ExtractionConfig{
		LLMProvider:             getEnv("LLM_PROVIDER", "openai"),
		LLMModel:                getEnv("LLM_MODEL", ""),
		LLMAPIKey:               getEnv("LLM_API_KEY", ""),
		CacheDir:                getEnv("SENTINEL_CACHE_DIR", ""),
		CacheTTLHours:           getEnvAsInt("SENTINEL_CACHE_TTL_HOURS", 24),
		BatchSize:               getEnvAsInt("SENTINEL_BATCH_SIZE", 4000),
		MaxRetries:              getEnvAsInt("SENTINEL_MAX_RETRIES", 3),
		CircuitBreakerThreshold: getEnvAsInt("SENTINEL_CB_THRESHOLD", 5),
		CircuitBreakerTimeoutSec: getEnvAsInt("SENTINEL_CB_TIMEOUT_SEC", 60),
	}

	// Validate required fields
	if cfg.LLMAPIKey == "" && cfg.LLMProvider != "ollama" {
		return nil, fmt.Errorf("LLM_API_KEY is required when LLM_PROVIDER is not 'ollama'")
	}

	return cfg, nil
}
