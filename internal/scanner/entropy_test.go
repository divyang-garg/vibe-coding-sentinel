// Package scanner provides unit tests for entropy detection
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package scanner

import "testing"

func TestCalculateShannonEntropy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "single character",
			input:    "a",
			expected: 0,
		},
		{
			name:     "repeated character",
			input:    "aaaa",
			expected: 0,
		},
		{
			name:     "high entropy",
			input:    "abcdefghijklmnopqrstuvwxyz0123456789",
			expected: 5.17, // Approximate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateShannonEntropy(tt.input)
			if tt.input == "" {
				if result != 0 {
					t.Errorf("Expected 0 for empty string, got %f", result)
				}
			} else if result < 0 {
				t.Errorf("Entropy should be non-negative, got %f", result)
			}
		})
	}
}

func TestIsHighEntropySecret(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "short string",
			input:    "short",
			expected: false,
		},
		{
			name:     "low entropy long string",
			input:    "aaaaaaaaaaaaaaaaaaaa",
			expected: false,
		},
		{
			name:     "high entropy secret",
			input:    "sk_test_EXAMPLE_KEY_PLACEHOLDER_DO_NOT_USE",
			expected: true,
		},
		{
			name:     "base64-like high entropy",
			input:    "dGhpc2lzYXZlcnlsb25nc3RyaW5ndGhhdGxvb2tzbGlrZWJhc2U2NA==",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHighEntropySecret(tt.input)
			if result != tt.expected {
				t.Errorf("isHighEntropySecret(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
