// Fixed import structure
package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// LLMConfig contains LLM provider configuration
type LLMConfig struct {
	ID               string                 `json:"id,omitempty"`
	Provider         string                 `json:"provider"`
	APIKey           string                 `json:"api_key"` // Decrypted for use
	Model            string                 `json:"model"`
	KeyType          string                 `json:"key_type"`
	CostOptimization CostOptimizationConfig `json:"cost_optimization,omitempty"`
}

// CostOptimizationConfig contains cost optimization settings
type CostOptimizationConfig struct {
	UseCache          bool    `json:"use_cache"`
	CacheTTLHours     int     `json:"cache_ttl_hours"`
	ProgressiveDepth  bool    `json:"progressive_depth"`
	MaxCostPerRequest float64 `json:"max_cost_per_request,omitempty"`
}

// LLMUsage tracks token usage and costs
type LLMUsage struct {
	ID            string  `json:"id"`
	ProjectID     string  `json:"project_id"`
	ValidationID  string  `json:"validation_id,omitempty"`
	Provider      string  `json:"provider"`
	Model         string  `json:"model"`
	TokensUsed    int     `json:"tokens_used"`
	EstimatedCost float64 `json:"estimated_cost"`
	CreatedAt     string  `json:"created_at"`
}

// getEncryptionKey retrieves or generates the encryption key for API keys
func getEncryptionKey() ([]byte, error) {
	// In production, this should be stored securely (e.g., in a secrets manager)
	// For now, use an environment variable or generate a key
	keyStr := os.Getenv("SENTINEL_ENCRYPTION_KEY")
	if keyStr == "" {
		// Generate a key (32 bytes for AES-256)
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
		// In production, this should be persisted securely
		return key, nil
	}

	// Decode base64 key
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key format: %w", err)
	}

	// Ensure key is 32 bytes (AES-256)
	if len(key) != 32 {
		// Hash to 32 bytes
		hash := sha256.Sum256(key)
		key = hash[:]
	}

	return key, nil
}

// encryptAPIKey encrypts an API key using AES-256
func encryptAPIKey(apiKey string) ([]byte, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(apiKey), nil)
	return ciphertext, nil
}

// decryptAPIKey decrypts an API key using AES-256
func decryptAPIKey(encrypted []byte) (string, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// getLLMConfig retrieves LLM configuration for a project
