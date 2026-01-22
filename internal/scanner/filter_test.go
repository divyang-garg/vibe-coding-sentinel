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

func TestIsInStringLiteralAt(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		pos      int
		expected bool
	}{
		{
			name:     "position out of bounds (negative)",
			line:     `const msg = "test";`,
			pos:      -1,
			expected: false,
		},
		{
			name:     "position out of bounds (too large)",
			line:     `const msg = "test";`,
			pos:      100,
			expected: false,
		},
		{
			name:     "position inside double quotes",
			line:     `const msg = "test";`,
			pos:      15,
			expected: true,
		},
		{
			name:     "position inside single quotes",
			line:     `const msg = 'test';`,
			pos:      15,
			expected: true,
		},
		{
			name:     "position outside quotes",
			line:     `const msg = "test";`,
			pos:      10,
			expected: false,
		},
		{
			name:     "position after closing quote",
			line:     `const msg = "test";`,
			pos:      20,
			expected: false,
		},
		{
			name:     "escaped quotes",
			line:     `const msg = "He said \"hello\"";`,
			pos:      20,
			expected: true,
		},
		{
			name:     "unclosed quote",
			line:     `const msg = "test`,
			pos:      15,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInStringLiteralAt(tt.line, tt.pos)
			if result != tt.expected {
				t.Errorf("isInStringLiteralAt(%q, %d) = %v, want %v", tt.line, tt.pos, result, tt.expected)
			}
		})
	}
}

func TestIsInDocComment(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		lineNum  int
		match    string
		expected bool
	}{
		{
			name:     "preceding line has /// doc comment",
			lines:    []string{"/// This is a doc comment", "const apiKey = 'secret';"},
			lineNum:  1,
			match:    "apiKey",
			expected: true,
		},
		{
			name:     "preceding line has /** doc comment",
			lines:    []string{"/** This is a doc comment */", "const apiKey = 'secret';"},
			lineNum:  1,
			match:    "apiKey",
			expected: true,
		},
		{
			name:     "preceding line has # doc comment",
			lines:    []string{"# This is a doc comment", "apiKey = 'secret'"},
			lineNum:  1,
			match:    "apiKey",
			expected: true,
		},
		{
			name:     "preceding line has // comment",
			lines:    []string{"// This is a comment", "const apiKey = 'secret';"},
			lineNum:  1,
			match:    "apiKey",
			expected: true,
		},
		{
			name:     "no doc comment",
			lines:    []string{"const x = 1;", "const apiKey = 'secret';"},
			lineNum:  1,
			match:    "apiKey",
			expected: false,
		},
		{
			name:     "line number out of bounds",
			lines:    []string{"const apiKey = 'secret';"},
			lineNum:  5,
			match:    "apiKey",
			expected: false,
		},
		{
			name:     "doc comment 2 lines before",
			lines:    []string{"/// Doc comment", "const x = 1;", "const apiKey = 'secret';"},
			lineNum:  2,
			match:    "apiKey",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInDocComment(tt.lines, tt.lineNum, tt.match)
			if result != tt.expected {
				t.Errorf("isInDocComment(%v, %d, %q) = %v, want %v", tt.lines, tt.lineNum, tt.match, result, tt.expected)
			}
		})
	}
}

func TestShouldFilterFinding(t *testing.T) {
	tests := []struct {
		name     string
		finding  Finding
		line     string
		allLines []string
		expected bool
	}{
		{
			name: "filter finding in comment",
			finding: Finding{
				Type:    "secrets",
				Pattern: "apiKey",
				Line:    1,
				Column:  10,
			},
			line:     "// apiKey = 'secret'",
			allLines: []string{"// apiKey = 'secret'"},
			expected: true,
		},
		{
			name: "filter non-secret finding in string literal",
			finding: Finding{
				Type:    "debug",
				Pattern: "console.log",
				Line:    1,
				Column:  15,
			},
			line:     `const msg = "console.log('test')";`,
			allLines: []string{`const msg = "console.log('test')";`},
			expected: true,
		},
		{
			name: "keep secret finding in string literal",
			finding: Finding{
				Type:    "secrets",
				Pattern: "apiKey",
				Line:    1,
				Column:  15,
			},
			line:     `const msg = "apiKey = 'secret'";`,
			allLines: []string{`const msg = "apiKey = 'secret'";`},
			expected: false,
		},
		{
			name: "filter finding in doc comment",
			finding: Finding{
				Type:    "secrets",
				Pattern: "apiKey",
				Line:    2,
				Column:  10,
			},
			line:     "const apiKey = 'secret';",
			allLines: []string{"/// This is a doc comment", "const apiKey = 'secret';"},
			expected: true,
		},
		{
			name: "keep finding not in comment or string",
			finding: Finding{
				Type:    "secrets",
				Pattern: "apiKey",
				Line:    1,
				Column:  10,
			},
			line:     "const apiKey = 'secret';",
			allLines: []string{"const apiKey = 'secret';"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldFilterFinding(tt.finding, tt.line, tt.allLines)
			if result != tt.expected {
				t.Errorf("shouldFilterFinding() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterFalsePositives_EdgeCases(t *testing.T) {
	t.Run("handles invalid line numbers", func(t *testing.T) {
		findings := []Finding{
			{
				Type:    "secrets",
				Line:    0, // Invalid line number
				Pattern: "apiKey",
			},
			{
				Type:    "secrets",
				Line:    100, // Line number out of bounds
				Pattern: "apiKey",
			},
		}

		content := "const apiKey = 'secret';"
		filtered := FilterFalsePositives(findings, content)

		// Should keep findings we can't verify
		if len(filtered) != 2 {
			t.Errorf("Expected 2 findings, got %d", len(filtered))
		}
	})

	t.Run("handles empty content", func(t *testing.T) {
		findings := []Finding{
			{
				Type:    "secrets",
				Line:    1,
				Pattern: "apiKey",
			},
		}

		filtered := FilterFalsePositives(findings, "")
		if len(filtered) != 1 {
			t.Errorf("Expected 1 finding, got %d", len(filtered))
		}
	})
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		prefix   string
		expected bool
	}{
		{
			name:     "has prefix",
			s:        "hello world",
			prefix:   "hello",
			expected: true,
		},
		{
			name:     "does not have prefix",
			s:        "hello world",
			prefix:   "world",
			expected: false,
		},
		{
			name:     "prefix longer than string",
			s:        "hi",
			prefix:   "hello",
			expected: false,
		},
		{
			name:     "empty prefix",
			s:        "hello",
			prefix:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			prefix:   "hello",
			expected: false,
		},
		{
			name:     "exact match",
			s:        "hello",
			prefix:   "hello",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPrefix(tt.s, tt.prefix)
			if result != tt.expected {
				t.Errorf("hasPrefix(%q, %q) = %v, want %v", tt.s, tt.prefix, result, tt.expected)
			}
		})
	}
}
