// Package scanner provides entropy-based secret detection
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package scanner

import (
	"math"
	"regexp"
)

// calculateShannonEntropy calculates the Shannon entropy of a string
// Higher entropy indicates more randomness, which is characteristic of secrets
func calculateShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	// Count character frequencies
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}

	// Calculate entropy: H = -sum(p * log2(p))
	var entropy float64
	length := float64(len(s))
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// isHighEntropySecret determines if a string is likely a secret based on entropy
// Secrets typically have high entropy (>4.5) and sufficient length (>=20 chars)
func isHighEntropySecret(s string) bool {
	if len(s) < 20 {
		return false
	}

	entropy := calculateShannonEntropy(s)
	return entropy > 4.5
}

// detectEntropySecrets scans content for high-entropy strings that might be secrets
func detectEntropySecrets(content string, filePath string) []Finding {
	var findings []Finding

	// Pattern to match potential secret-like strings
	// Matches: variable assignments with long alphanumeric values
	secretPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:api[_-]?key|secret|token|password|auth[_-]?token)\s*[=:]\s*["']?([a-zA-Z0-9+/=]{20,})["']?`),
		regexp.MustCompile(`["']([a-zA-Z0-9+/=]{32,})["']`), // Long base64-like strings
		regexp.MustCompile(`\b([a-f0-9]{40,})\b`),           // Long hex strings (like SHA1/SHA256)
	}

	lines := splitLines(content)
	for lineNum, line := range lines {
		for _, pattern := range secretPatterns {
			matches := pattern.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					candidate := match[1]
					// Check if it's high entropy
					if isHighEntropySecret(candidate) {
						// Find column position
						idx := findStringIndex(line, candidate)
						column := 0
						if idx >= 0 {
							column = idx + 1
						}

						finding := Finding{
							Type:     "high_entropy_secret",
							Severity: SeverityCritical,
							File:     filePath,
							Line:     lineNum + 1,
							Column:   column,
							Message:  "High-entropy string detected (likely secret or API key)",
							Pattern:  trimSpace(line),
							Code:     trimSpace(line),
						}
						findings = append(findings, finding)
					}
				}
			}
		}
	}

	return findings
}

// findStringIndex finds the index of substring in string (helper to avoid strings import)
func findStringIndex(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
