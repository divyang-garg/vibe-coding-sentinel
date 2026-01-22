// Package llm provides integration tests for encryption
package llm

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
)

// TestEncryptionIntegration validates that encryption works end-to-end
// This simulates the actual usage in config.go
func TestEncryptionIntegration(t *testing.T) {
	// Set up a test encryption key
	originalKey := os.Getenv("SENTINEL_ENCRYPTION_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("SENTINEL_ENCRYPTION_KEY", originalKey)
		} else {
			os.Unsetenv("SENTINEL_ENCRYPTION_KEY")
		}
	}()

	// Generate a test key
	testKey := make([]byte, 32)
	if _, err := rand.Read(testKey); err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	os.Setenv("SENTINEL_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(testKey))

	// Simulate the flow in config.go: saveLLMConfig
	originalAPIKey := "sk-test-api-key-1234567890abcdefghijklmnopqrstuvwxyz"

	// Step 1: Encrypt (as done in saveLLMConfig)
	encryptedKey, err := encryptAPIKey(originalAPIKey)
	if err != nil {
		t.Fatalf("encryptAPIKey() failed: %v", err)
	}

	// Verify it's encrypted (not plaintext)
	if string(encryptedKey) == originalAPIKey {
		t.Fatal("API key was not encrypted - still in plaintext!")
	}

	// Verify encrypted data is longer (includes nonce + ciphertext)
	if len(encryptedKey) <= len(originalAPIKey) {
		t.Fatal("Encrypted data should be longer than plaintext")
	}

	// Step 2: Store encryptedKey in database (simulated)
	// In real code: INSERT INTO llm_configurations (api_key_encrypted) VALUES ($1)
	storedEncrypted := encryptedKey

	// Step 3: Retrieve from database (simulated)
	retrievedEncrypted := storedEncrypted

	// Step 4: Decrypt (as done in ListLLMConfigs)
	decryptedAPIKey, err := decryptAPIKey(retrievedEncrypted)
	if err != nil {
		t.Fatalf("decryptAPIKey() failed: %v", err)
	}

	// Step 5: Verify decrypted matches original
	if decryptedAPIKey != originalAPIKey {
		t.Fatalf("Decrypted API key doesn't match original: got %q, want %q", decryptedAPIKey, originalAPIKey)
	}

	// Step 6: Mask for display (as done in ListLLMConfigs)
	maskedKey := maskAPIKey(decryptedAPIKey)
	if maskedKey == decryptedAPIKey {
		t.Error("Masked key should not equal original key")
	}
	if len(maskedKey) >= len(decryptedAPIKey) {
		t.Error("Masked key should be shorter or same length")
	}

	t.Logf("✅ Encryption integration test passed")
	t.Logf("   Original: %s", originalAPIKey)
	t.Logf("   Encrypted length: %d bytes", len(encryptedKey))
	t.Logf("   Decrypted: %s", decryptedAPIKey)
	t.Logf("   Masked: %s", maskedKey)
}

// TestEncryptionSecurity validates security properties
func TestEncryptionSecurity(t *testing.T) {
	// Set up a test encryption key
	originalKey := os.Getenv("SENTINEL_ENCRYPTION_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("SENTINEL_ENCRYPTION_KEY", originalKey)
		} else {
			os.Unsetenv("SENTINEL_ENCRYPTION_KEY")
		}
	}()

	// Generate a test key
	testKey := make([]byte, 32)
	if _, err := rand.Read(testKey); err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	os.Setenv("SENTINEL_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(testKey))

	apiKey := "sk-secret-api-key-12345"

	// Test 1: Same plaintext produces different ciphertext (nonce is random)
	enc1, _ := encryptAPIKey(apiKey)
	enc2, _ := encryptAPIKey(apiKey)
	if string(enc1) == string(enc2) {
		t.Error("Same plaintext should produce different ciphertext (nonce is random)")
	}

	// Test 2: Both decrypt to same value
	dec1, _ := decryptAPIKey(enc1)
	dec2, _ := decryptAPIKey(enc2)
	if dec1 != apiKey || dec2 != apiKey {
		t.Error("Both encrypted versions should decrypt to same value")
	}

	// Test 3: Tampered ciphertext fails to decrypt
	tampered := make([]byte, len(enc1))
	copy(tampered, enc1)
	tampered[len(tampered)-1] ^= 0xFF // Flip last byte
	_, err := decryptAPIKey(tampered)
	if err == nil {
		t.Error("Tampered ciphertext should fail to decrypt (GCM authentication)")
	}

	// Test 4: Wrong key cannot decrypt
	os.Setenv("SENTINEL_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	_, err = decryptAPIKey(enc1)
	if err == nil {
		t.Error("Wrong encryption key should fail to decrypt")
	}

	t.Log("✅ Security properties validated")
}
