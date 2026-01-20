// Package ast fuzz tests
// Complies with CODING_STANDARDS.md: Tests max 500 lines
package ast

import (
	"testing"
)

// FuzzAnalyzeAST fuzzes the AnalyzeAST function to ensure no panics
func FuzzAnalyzeAST(f *testing.F) {
	// Add seed corpus
	f.Add("func test() {}", "go")
	f.Add("function test() {}", "javascript")
	f.Add("async function test() { fetch('/api'); }", "javascript")
	f.Add("def test(): pass", "python")
	f.Add("interface Test { name: string; }", "typescript")
	f.Add("", "go")               // Empty code
	f.Add("invalid code {", "go") // Invalid code
	f.Add("package main\nfunc main() {\n\tfmt.Println(\"hello\")\n}", "go")

	// Fuzz function - should never panic
	f.Fuzz(func(t *testing.T, code string, lang string) {
		// Should never panic due to panic recovery wrapper
		findings, stats, err := AnalyzeAST(code, lang, nil)

		// If error occurs, that's okay - we just don't want panics
		if err != nil {
			return
		}

		// Basic sanity checks
		if findings == nil {
			t.Error("Findings should not be nil (should be empty slice, not nil)")
		}

		if stats.NodesVisited < 0 {
			t.Error("NodesVisited should be >= 0")
		}

		if stats.ParseTime < 0 {
			t.Error("ParseTime should be >= 0")
		}

		if stats.AnalysisTime < 0 {
			t.Error("AnalysisTime should be >= 0")
		}
	})
}

// FuzzSafeSlice fuzzes the safeSlice function to ensure no panics
func FuzzSafeSlice(f *testing.F) {
	f.Add("hello world", uint32(0), uint32(5))
	f.Add("test", uint32(10), uint32(20)) // Out of bounds
	f.Add("", uint32(0), uint32(0))       // Empty string
	f.Add("abc", uint32(5), uint32(3))    // start > end

	f.Fuzz(func(t *testing.T, code string, start uint32, end uint32) {
		// Should never panic
		result := safeSlice(code, start, end)

		// Basic sanity check
		if len(result) > len(code) {
			t.Errorf("Result length %d exceeds code length %d", len(result), len(code))
		}
	})
}

// FuzzGetLineColumn fuzzes the getLineColumn function to ensure no panics
func FuzzGetLineColumn(f *testing.F) {
	f.Add("hello world", 0)
	f.Add("line1\nline2\nline3", 10)
	f.Add("", 0)
	f.Add("test", 100) // Out of bounds

	f.Fuzz(func(t *testing.T, code string, byteOffset int) {
		// Should never panic
		line, col := getLineColumn(code, byteOffset)

		// Basic sanity checks
		if line < 1 {
			t.Errorf("Line number should be >= 1, got %d", line)
		}

		if col < 1 {
			t.Errorf("Column number should be >= 1, got %d", col)
		}
	})
}
