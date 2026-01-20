// Package scanner provides tests for report formatters
// Complies with CODING_STANDARDS.md: Test file max 500 lines
package scanner

import (
	"strings"
	"testing"
)

func TestFormatHTML(t *testing.T) {
	result := &Result{
		Success: false,
		Findings: []Finding{
			{
				Type:     "secrets",
				Severity: SeverityCritical,
				File:     "test.js",
				Line:     10,
				Message:  "Test finding",
				Pattern:  "test pattern",
			},
		},
		Summary: map[string]int{
			"secrets": 1,
		},
		Timestamp: "2024-01-01T00:00:00Z",
	}

	html := FormatHTML(result)

	if !strings.Contains(html, "<html>") {
		t.Error("HTML output should contain <html> tag")
	}

	if !strings.Contains(html, "test.js") {
		t.Error("HTML output should contain filename")
	}

	if !strings.Contains(html, "Test finding") {
		t.Error("HTML output should contain finding message")
	}

	if !strings.Contains(html, "critical") && !strings.Contains(html, "CRITICAL") {
		t.Error("HTML output should contain severity")
	}
}

func TestFormatMarkdown(t *testing.T) {
	result := &Result{
		Success: true,
		Findings: []Finding{
			{
				Type:     "debug",
				Severity: SeverityWarning,
				File:     "app.js",
				Line:     5,
				Message:  "Console.log detected",
			},
		},
		Summary: map[string]int{
			"debug": 1,
		},
		Timestamp: "2024-01-01T00:00:00Z",
	}

	md := FormatMarkdown(result)

	if !strings.Contains(md, "# Sentinel Security Audit Report") {
		t.Error("Markdown should contain report header")
	}

	if !strings.Contains(md, "app.js") {
		t.Error("Markdown should contain filename")
	}

	if !strings.Contains(md, "Console.log detected") {
		t.Error("Markdown should contain finding message")
	}

	if !strings.Contains(md, "✅") || !strings.Contains(md, "PASSED") {
		t.Error("Markdown should indicate success")
	}
}

func TestFormatJSON(t *testing.T) {
	result := &Result{
		Success: false,
		Findings: []Finding{
			{
				Type:     "eval_usage",
				Severity: SeverityCritical,
				File:     "script.js",
				Line:     20,
				Message:  "eval() usage detected",
			},
		},
		Summary: map[string]int{
			"eval_usage": 1,
		},
		Timestamp: "2024-01-01T00:00:00Z",
	}

	json := FormatJSON(result)

	// JSON uses Go's encoding which may include spaces
	if !strings.Contains(json, `"success"`) {
		t.Error("JSON should contain success field")
	}

	if !strings.Contains(json, "script.js") {
		t.Error("JSON should contain filename")
	}

	if !strings.Contains(json, "eval() usage detected") {
		t.Error("JSON should contain finding message")
	}
}

func TestFormatText(t *testing.T) {
	result := &Result{
		Success:   true,
		Findings:  []Finding{},
		Summary:   make(map[string]int),
		Timestamp: "2024-01-01T00:00:00Z",
	}

	text := FormatText(result)

	if !strings.Contains(text, "Sentinel Security Audit") {
		t.Error("Text should contain report header")
	}

	if !strings.Contains(text, "Status: ✅ PASSED") {
		t.Error("Text should indicate success")
	}

	if !strings.Contains(text, "Total Findings: 0") {
		t.Error("Text should show finding count")
	}
}

func TestFormatText_WithFindings(t *testing.T) {
	result := &Result{
		Success: false,
		Findings: []Finding{
			{
				Type:     "secrets",
				Severity: SeverityCritical,
				File:     "config.js",
				Line:     15,
				Message:  "API key detected",
				Pattern:  "apiKey = 'secret'",
			},
			{
				Type:     "debug",
				Severity: SeverityWarning,
				File:     "app.js",
				Line:     5,
				Message:  "Console.log detected",
			},
		},
		Summary: map[string]int{
			"secrets": 1,
			"debug":   1,
		},
		Timestamp: "2024-01-01T00:00:00Z",
	}

	text := FormatText(result)

	if !strings.Contains(text, "❌ FAILED") {
		t.Error("Text should indicate failure")
	}

	if !strings.Contains(text, "Total Findings: 2") {
		t.Error("Text should show correct finding count")
	}

	if !strings.Contains(text, "config.js:15") {
		t.Error("Text should show file and line number")
	}

	if !strings.Contains(text, "API key detected") {
		t.Error("Text should show finding message")
	}
}
