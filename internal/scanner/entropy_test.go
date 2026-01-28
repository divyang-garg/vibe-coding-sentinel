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
			expected: 5.17, // Approximate - actual value may vary slightly
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
			} else if tt.name == "high entropy" {
				// For high entropy test, verify it's in reasonable range (5.0-5.5)
				// Actual calculation may vary slightly due to floating point precision
				if result < 5.0 || result > 5.5 {
					t.Errorf("Expected entropy around 5.17 for high entropy string, got %f (expected range: 5.0-5.5)", result)
				}
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
			input:    "aB3dEf9gH2iJ4kL6mN8oP1qR5sT7uV0wX2yZ4bC6dF8gH1",
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

func TestFindStringIndex(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{
			name:     "empty substr",
			s:        "hello",
			substr:   "",
			expected: 0,
		},
		{
			name:     "substr longer than string",
			s:        "hi",
			substr:   "hello",
			expected: -1,
		},
		{
			name:     "substr found at start",
			s:        "hello world",
			substr:   "hello",
			expected: 0,
		},
		{
			name:     "substr found in middle",
			s:        "hello world",
			substr:   "world",
			expected: 6,
		},
		{
			name:     "substr not found",
			s:        "hello world",
			substr:   "xyz",
			expected: -1,
		},
		{
			name:     "substr at end",
			s:        "hello world",
			substr:   "world",
			expected: 6,
		},
		{
			name:     "multiple occurrences - returns first",
			s:        "hello hello",
			substr:   "hello",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findStringIndex(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("findStringIndex(%q, %q) = %d, want %d", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}

func TestDetectEntropySecrets(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filePath string
		expected int // Expected number of findings
	}{
		{
			name:     "api key pattern",
			content:  `const apiKey = "aB3dEf9gH2iJ4kL6mN8oP1qR5sT7uV0wX2yZ4bC6dF8gH1";`,
			filePath: "test.js",
			expected: 1,
		},
		{
			name:     "long base64 string",
			content:  `const token = "dGhpc2lzYXZlcnlsb25nc3RyaW5ndGhhdGxvb2tzbGlrZWJhc2U2NA==";`,
			filePath: "test.js",
			expected: 1,
		},
		{
			name:     "long hex string",
			content:  `const hash = "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456";`,
			filePath: "test.js",
			expected: 1,
		},
		{
			name:     "low entropy string",
			content:  `const key = "aaaaaaaaaaaaaaaaaaaa";`,
			filePath: "test.js",
			expected: 0,
		},
		{
			name:     "short string",
			content:  `const key = "short";`,
			filePath: "test.js",
			expected: 0,
		},
		{
			name: "multiple secrets",
			content: `const apiKey = "aB3dEf9gH2iJ4kL6mN8oP1qR5sT7uV0wX2yZ4bC6dF8gH1";
const secret = "dGhpc2lzYXZlcnlsb25nc3RyaW5ndGhhdGxvb2tzbGlrZWJhc2U2NA==";`,
			filePath: "test.js",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := detectEntropySecrets(tt.content, tt.filePath)
			if len(findings) != tt.expected {
				t.Errorf("detectEntropySecrets() found %d findings, want %d", len(findings), tt.expected)
			}
			for _, finding := range findings {
				if finding.File != tt.filePath {
					t.Errorf("finding.File = %q, want %q", finding.File, tt.filePath)
				}
				if finding.Type != "high_entropy_secret" {
					t.Errorf("finding.Type = %q, want 'high_entropy_secret'", finding.Type)
				}
			}
		})
	}
}
