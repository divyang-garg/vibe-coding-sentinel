// Package scanner tests for parallel utility functions
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package scanner

import (
	"testing"
)

func TestItoa(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 123, "123"},
		{"positive_single", 5, "5"},
		{"positive_large", 987654321, "987654321"},
		{"negative", -42, "-42"},
		{"negative_single", -7, "-7"},
		{"negative_large", -123456789, "-123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := itoa(tt.input)
			if result != tt.expected {
				t.Errorf("itoa(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no_whitespace", "hello", "hello"},
		{"leading_spaces", "  hello", "hello"},
		{"trailing_spaces", "hello  ", "hello"},
		{"both", "  hello  ", "hello"},
		{"tabs", "\t\thello\t\t", "hello"},
		{"carriage_return", "hello\r\n", "hello\r\n"}, // trimSpace only trims spaces/tabs, not newlines
		{"mixed", "  \t hello \t  ", "hello"},
		{"only_whitespace", "   ", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimSpace(tt.input)
			if result != tt.expected {
				t.Errorf("trimSpace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
