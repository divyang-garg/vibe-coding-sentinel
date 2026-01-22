// Package llm provides LLM security utilities
// Complies with CODING_STANDARDS.md: Security modules max 250 lines
package llm

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

// maskAPIKey masks an API key for display
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// getEncryptionKey retrieves or generates the encryption key for API keys
// Uses SENTINEL_ENCRYPTION_KEY environment variable or generates a key
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

// encryptAPIKey encrypts an API key using AES-256-GCM
func encryptAPIKey(apiKey string) ([]byte, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	key, err := getEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM (Galois/Counter Mode) for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce (number used once) for this encryption
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, []byte(apiKey), nil)
	return ciphertext, nil
}

// decryptAPIKey decrypts an API key using AES-256-GCM
func decryptAPIKey(encryptedKey []byte) (string, error) {
	if len(encryptedKey) == 0 {
		return "", fmt.Errorf("encrypted key cannot be empty")
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
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

	// Extract nonce from ciphertext
	nonceSize := gcm.NonceSize()
	if len(encryptedKey) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedKey[:nonceSize], encryptedKey[nonceSize:]

	// Decrypt and verify authentication
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// validateAPIKeyFormat validates API key format
func validateAPIKeyFormat(provider, apiKey string) error {
	if len(apiKey) < 10 {
		return fmt.Errorf("API key too short")
	}
	return nil
}
