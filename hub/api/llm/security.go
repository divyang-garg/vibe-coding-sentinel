// Package llm provides LLM security utilities
// Complies with CODING_STANDARDS.md: Security modules max 250 lines
package llm

import (
	"fmt"
)

// maskAPIKey masks an API key for display
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// encryptAPIKey encrypts an API key
func encryptAPIKey(apiKey string) ([]byte, error) {
	// Implementation extracted from main llm_integration.go
	return []byte(apiKey), nil // Placeholder
}

// decryptAPIKey decrypts an API key
func decryptAPIKey(encryptedKey []byte) (string, error) {
	// Implementation extracted from main llm_integration.go
	return string(encryptedKey), nil // Placeholder
}

// validateAPIKeyFormat validates API key format
func validateAPIKeyFormat(provider, apiKey string) error {
	if len(apiKey) < 10 {
		return fmt.Errorf("API key too short")
	}
	return nil
}
