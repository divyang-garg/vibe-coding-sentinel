// Package llm provides LLM security utilities tests
package llm

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
)

func TestEncryptDecryptAPIKey(t *testing.T) {
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

	testCases := []struct {
		name    string
		apiKey  string
		wantErr bool
	}{
		{
			name:    "valid API key",
			apiKey:  "sk-test1234567890abcdefghijklmnopqrstuvwxyz",
			wantErr: false,
		},
		{
			name:    "short API key",
			apiKey:  "sk-test123",
			wantErr: false,
		},
		{
			name:    "long API key",
			apiKey:  "sk-" + string(make([]byte, 200)),
			wantErr: false,
		},
		{
			name:    "special characters",
			apiKey:  "sk-test!@#$%^&*()_+-=[]{}|;:,.<>?",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := encryptAPIKey(tc.apiKey)
			if (err != nil) != tc.wantErr {
				t.Errorf("encryptAPIKey() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Verify encrypted data is different from plaintext
			if string(encrypted) == tc.apiKey {
				t.Error("encrypted data should not equal plaintext")
			}

			// Verify encrypted data is longer (includes nonce)
			if len(encrypted) <= len(tc.apiKey) {
				t.Error("encrypted data should be longer than plaintext (includes nonce)")
			}

			// Decrypt
			decrypted, err := decryptAPIKey(encrypted)
			if err != nil {
				t.Errorf("decryptAPIKey() error = %v", err)
				return
			}

			// Verify decrypted matches original
			if decrypted != tc.apiKey {
				t.Errorf("decrypted = %q, want %q", decrypted, tc.apiKey)
			}
		})
	}
}

func TestEncryptAPIKey_EmptyKey(t *testing.T) {
	_, err := encryptAPIKey("")
	if err == nil {
		t.Error("encryptAPIKey() with empty key should return error")
	}
}

func TestDecryptAPIKey_InvalidData(t *testing.T) {
	testCases := []struct {
		name        string
		encrypted   []byte
		description string
	}{
		{
			name:        "empty data",
			encrypted:   []byte{},
			description: "should fail on empty encrypted data",
		},
		{
			name:        "too short",
			encrypted:   []byte{1, 2, 3},
			description: "should fail on data shorter than nonce size",
		},
		{
			name:        "invalid ciphertext",
			encrypted:   make([]byte, 50),
			description: "should fail on invalid ciphertext",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decryptAPIKey(tc.encrypted)
			if err == nil {
				t.Errorf("decryptAPIKey() should fail for %s", tc.description)
			}
		})
	}
}

func TestEncryptDecryptAPIKey_DifferentKeys(t *testing.T) {
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

	apiKey := "sk-test1234567890"

	// Encrypt the same key twice
	encrypted1, err := encryptAPIKey(apiKey)
	if err != nil {
		t.Fatalf("encryptAPIKey() error = %v", err)
	}

	encrypted2, err := encryptAPIKey(apiKey)
	if err != nil {
		t.Fatalf("encryptAPIKey() error = %v", err)
	}

	// Encrypted values should be different (due to random nonce)
	if string(encrypted1) == string(encrypted2) {
		t.Error("encrypted values should be different (nonce is random)")
	}

	// But both should decrypt to the same value
	decrypted1, err := decryptAPIKey(encrypted1)
	if err != nil {
		t.Fatalf("decryptAPIKey() error = %v", err)
	}

	decrypted2, err := decryptAPIKey(encrypted2)
	if err != nil {
		t.Fatalf("decryptAPIKey() error = %v", err)
	}

	if decrypted1 != apiKey || decrypted2 != apiKey {
		t.Error("both decrypted values should match original")
	}
}

func TestMaskAPIKey(t *testing.T) {
	testCases := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "long key",
			apiKey:   "sk-test1234567890abcdefghijklmnop",
			expected: "sk-t****mnop", // First 4 + **** + last 4
		},
		{
			name:     "short key",
			apiKey:   "sk-test",
			expected: "****",
		},
		{
			name:     "exactly 8 chars",
			apiKey:   "12345678",
			expected: "****",
		},
		{
			name:     "9 chars",
			apiKey:   "123456789",
			expected: "1234****6789", // First 4 + **** + last 4
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := maskAPIKey(tc.apiKey)
			if result != tc.expected {
				t.Errorf("maskAPIKey() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestValidateAPIKeyFormat(t *testing.T) {
	testCases := []struct {
		name      string
		provider  string
		apiKey    string
		wantError bool
	}{
		{
			name:      "valid long key",
			provider:  "openai",
			apiKey:    "sk-test1234567890abcdefghijklmnop",
			wantError: false,
		},
		{
			name:      "valid minimum length",
			provider:  "openai",
			apiKey:    "sk-test1234", // 12 chars
			wantError: false,
		},
		{
			name:      "too short",
			provider:  "openai",
			apiKey:    "sk-test", // 7 chars
			wantError: true,
		},
		{
			name:      "empty key",
			provider:  "openai",
			apiKey:    "",
			wantError: true,
		},
		{
			name:      "exactly 9 chars",
			provider:  "anthropic",
			apiKey:    "sk-test123", // 10 chars
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAPIKeyFormat(tc.provider, tc.apiKey)
			if (err != nil) != tc.wantError {
				t.Errorf("validateAPIKeyFormat() error = %v, wantError %v", err, tc.wantError)
			}
			if err != nil && !tc.wantError {
				t.Errorf("validateAPIKeyFormat() unexpected error: %v", err)
			}
		})
	}
}
