// Package scanner provides tests for parallel utility functions
package scanner

import (
	"testing"
)

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
		{-1, "-1"},
		{-42, "-42"},
		{12345, "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := itoa(tt.input)
			if result != tt.expected {
				t.Errorf("itoa(%d) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single line", "hello", []string{"hello"}},
		{"two lines", "hello\nworld", []string{"hello", "world"}},
		{"empty string", "", []string{}},
		{"newline only", "\n", []string{""}},
		{"multiple lines", "a\nb\nc", []string{"a", "b", "c"}},
		{"no trailing newline", "a\nb", []string{"a", "b"}},
		{"trailing newline", "a\nb\n", []string{"a", "b"}}, // splitLines doesn't add empty for trailing newline
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitLines() length = %d, want %d", len(result), len(tt.expected))
			}
			for i, expected := range tt.expected {
				if i < len(result) && result[i] != expected {
					t.Errorf("splitLines()[%d] = %s, want %s", i, result[i], expected)
				}
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
		{"no spaces", "hello", "hello"},
		{"leading spaces", "  hello", "hello"},
		{"trailing spaces", "hello  ", "hello"},
		{"both sides", "  hello  ", "hello"},
		{"tabs", "\thello\t", "hello"},
		{"carriage return", "hello\r", "hello"},
		{"mixed whitespace", " \t hello \t ", "hello"},
		{"only spaces", "   ", ""},
		{"empty string", "", ""},
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
