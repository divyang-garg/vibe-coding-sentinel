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

func TestFormatJSON_EdgeCases(t *testing.T) {
	t.Run("handles empty result", func(t *testing.T) {
		result := &Result{
			Success:   true,
			Findings:  []Finding{},
			Summary:   make(map[string]int),
			Timestamp: "2024-01-01T00:00:00Z",
		}

		json := FormatJSON(result)
		// json.MarshalIndent adds spaces, so check for both formats
		if !strings.Contains(json, `"success"`) || (!strings.Contains(json, `"success": true`) && !strings.Contains(json, `"success":true`)) {
			t.Error("JSON should contain success field")
		}
		if !strings.Contains(json, `"findings"`) || (!strings.Contains(json, `"findings": []`) && !strings.Contains(json, `"findings":[]`)) {
			t.Error("JSON should contain empty findings array")
		}
	})

	t.Run("handles result with all fields", func(t *testing.T) {
		result := &Result{
			Success: false,
			Findings: []Finding{
				{
					Type:     "secrets",
					Severity: SeverityCritical,
					File:     "test.js",
					Line:     10,
					Column:   5,
					Message:  "Test message",
					Pattern:  "test pattern",
					Code:     "test code",
				},
			},
			Summary: map[string]int{
				"secrets": 1,
			},
			Timestamp: "2024-01-01T00:00:00Z",
		}

		json := FormatJSON(result)
		if !strings.Contains(json, `"test.js"`) {
			t.Error("JSON should contain filename")
		}
		// json.MarshalIndent adds spaces, so check for both formats
		if !strings.Contains(json, `"column"`) || (!strings.Contains(json, `"column": 5`) && !strings.Contains(json, `"column":5`)) {
			t.Error("JSON should contain column when present")
		}
		if !strings.Contains(json, `"code"`) || (!strings.Contains(json, `"code": "test code"`) && !strings.Contains(json, `"code":"test code"`)) {
			t.Error("JSON should contain code when present")
		}
	})
}

func TestFormatHTML_EdgeCases(t *testing.T) {
	t.Run("handles HTML escaping", func(t *testing.T) {
		result := &Result{
			Success: false,
			Findings: []Finding{
				{
					Type:     "xss",
					Severity: SeverityHigh,
					File:     "test.html",
					Line:     1,
					Message:  "XSS vulnerability: <script>alert('xss')</script>",
					Pattern:  "<script>alert('xss')</script>",
				},
			},
			Summary: map[string]int{
				"xss": 1,
			},
			Timestamp: "2024-01-01T00:00:00Z",
		}

		html := FormatHTML(result)
		// Should escape HTML in message and pattern
		if strings.Contains(html, "<script>") {
			t.Error("HTML should escape script tags in content")
		}
		if !strings.Contains(html, "&lt;script&gt;") {
			t.Error("HTML should escape script tags")
		}
	})

	t.Run("handles empty summary", func(t *testing.T) {
		result := &Result{
			Success:   true,
			Findings:  []Finding{},
			Summary:   nil,
			Timestamp: "2024-01-01T00:00:00Z",
		}

		html := FormatHTML(result)
		if !strings.Contains(html, "<html>") {
			t.Error("HTML should be generated even with empty summary")
		}
	})
}

func TestFormatMarkdown_EdgeCases(t *testing.T) {
	t.Run("handles markdown special characters", func(t *testing.T) {
		result := &Result{
			Success: false,
			Findings: []Finding{
				{
					Type:     "secrets",
					Severity: SeverityCritical,
					File:     "test.md",
					Line:     1,
					Message:  "Pattern with *asterisks* and _underscores_",
					Pattern:  "**bold** and `code`",
				},
			},
			Summary: map[string]int{
				"secrets": 1,
			},
			Timestamp: "2024-01-01T00:00:00Z",
		}

		md := FormatMarkdown(result)
		if !strings.Contains(md, "*asterisks*") {
			t.Error("Markdown should preserve special characters")
		}
		if !strings.Contains(md, "```") {
			t.Error("Markdown should format pattern in code block")
		}
	})
}
