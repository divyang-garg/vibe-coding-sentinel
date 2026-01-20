// Package scanner provides unit tests for false positive filtering
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package scanner

import "testing"

func TestIsInComment(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		match    string
		expected bool
	}{
		{
			name:     "single line comment",
			line:     "// This is a comment with apiKey = 'secret'",
			match:    "apiKey",
			expected: true,
		},
		{
			name:     "not in comment",
			line:     "const apiKey = 'secret';",
			match:    "apiKey",
			expected: false,
		},
		{
			name:     "hash comment",
			line:     "# apiKey = 'secret'",
			match:    "apiKey",
			expected: true,
		},
		{
			name:     "SQL comment",
			line:     "-- SELECT * FROM users WHERE apiKey = 'secret'",
			match:    "apiKey",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInComment(tt.line, tt.match)
			if result != tt.expected {
				t.Errorf("isInComment(%q, %q) = %v, want %v", tt.line, tt.match, result, tt.expected)
			}
		})
	}
}

func TestIsInStringLiteral(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "double quoted string - balanced quotes",
			line:     `const msg = "This is a string";`,
			expected: false, // Line ends with semicolon, not inside string
		},
		{
			name:     "single quoted string - balanced quotes",
			line:     `const msg = 'This is a string';`,
			expected: false, // Line ends with semicolon, not inside string
		},
		{
			name:     "not in string",
			line:     `const apiKey = getKey();`,
			expected: false,
		},
		{
			name:     "escaped quotes - balanced",
			line:     `const msg = "He said \"hello\"";`,
			expected: false, // Line ends with semicolon, not inside string
		},
		{
			name:     "unclosed double quote",
			line:     `const msg = "This is a string`,
			expected: true, // Line ends inside string literal
		},
		{
			name:     "unclosed single quote",
			line:     `const msg = 'This is a string`,
			expected: true, // Line ends inside string literal
		},
		{
			name:     "unclosed backtick",
			line:     "const msg = `This is a string",
			expected: true, // Line ends inside string literal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInStringLiteral(tt.line)
			if result != tt.expected {
				t.Errorf("isInStringLiteral(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

func TestFilterFalsePositives(t *testing.T) {
	findings := []Finding{
		{
			Type:     "secrets",
			File:     "test.js",
			Line:     1,
			Pattern:  "apiKey = 'secret'",
			Severity: SeverityCritical,
		},
		{
			Type:     "secrets",
			File:     "test.js",
			Line:     2,
			Pattern:  "// apiKey = 'secret'",
			Severity: SeverityCritical,
		},
	}

	content := `const apiKey = 'secret';
// apiKey = 'secret'`

	filtered := FilterFalsePositives(findings, content)

	// Should filter out the comment finding
	if len(filtered) != 1 {
		t.Errorf("Expected 1 finding after filtering, got %d", len(filtered))
	}

	if filtered[0].Line != 1 {
		t.Errorf("Expected finding on line 1, got line %d", filtered[0].Line)
	}
}
