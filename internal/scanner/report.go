// Package scanner provides report formatting functionality
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package scanner

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
)

// FormatHTML formats scan results as HTML
func FormatHTML(result *Result) string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>Sentinel Security Audit Report</title>\n")
	sb.WriteString("<style>\n")
	sb.WriteString("body { font-family: Arial, sans-serif; margin: 20px; }\n")
	sb.WriteString(".header { background: #f0f0f0; padding: 15px; border-radius: 5px; }\n")
	sb.WriteString(".finding { margin: 10px 0; padding: 10px; border-left: 4px solid; }\n")
	sb.WriteString(".critical { border-color: #d32f2f; background: #ffebee; }\n")
	sb.WriteString(".high { border-color: #f57c00; background: #fff3e0; }\n")
	sb.WriteString(".medium { border-color: #fbc02d; background: #fffde7; }\n")
	sb.WriteString(".low { border-color: #388e3c; background: #e8f5e9; }\n")
	sb.WriteString(".summary { margin: 20px 0; }\n")
	sb.WriteString("</style>\n</head>\n<body>\n")

	sb.WriteString("<div class=\"header\">\n")
	sb.WriteString("<h1>Sentinel Security Audit Report</h1>\n")
	sb.WriteString(fmt.Sprintf("<p><strong>Timestamp:</strong> %s</p>\n", html.EscapeString(result.Timestamp)))
	sb.WriteString("<p><strong>Status:</strong> ")
	if result.Success {
		sb.WriteString("<span style=\"color: green;\">✅ PASSED</span>")
	} else {
		sb.WriteString("<span style=\"color: red;\">❌ FAILED</span>")
	}
	sb.WriteString("</p>\n")
	sb.WriteString(fmt.Sprintf("<p><strong>Total Findings:</strong> %d</p>\n", len(result.Findings)))
	sb.WriteString("</div>\n")

	// Summary section
	if len(result.Summary) > 0 {
		sb.WriteString("<div class=\"summary\">\n<h2>Summary by Type</h2>\n<ul>\n")
		keys := make([]string, 0, len(result.Summary))
		for k := range result.Summary {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("<li><strong>%s:</strong> %d</li>\n", html.EscapeString(k), result.Summary[k]))
		}
		sb.WriteString("</ul>\n</div>\n")
	}

	// Findings section
	if len(result.Findings) > 0 {
		sb.WriteString("<h2>Findings</h2>\n")
		for _, finding := range result.Findings {
			severityClass := strings.ToLower(string(finding.Severity))
			sb.WriteString(fmt.Sprintf("<div class=\"finding %s\">\n", severityClass))
			sb.WriteString(fmt.Sprintf("<p><strong>[%s]</strong> %s:%d</p>\n",
				html.EscapeString(string(finding.Severity)),
				html.EscapeString(finding.File),
				finding.Line))
			sb.WriteString(fmt.Sprintf("<p>%s</p>\n", html.EscapeString(finding.Message)))
			if finding.Pattern != "" {
				sb.WriteString(fmt.Sprintf("<pre>%s</pre>\n", html.EscapeString(finding.Pattern)))
			}
			sb.WriteString("</div>\n")
		}
	}

	sb.WriteString("</body>\n</html>\n")
	return sb.String()
}

// FormatMarkdown formats scan results as Markdown
func FormatMarkdown(result *Result) string {
	var sb strings.Builder

	sb.WriteString("# Sentinel Security Audit Report\n\n")
	sb.WriteString(fmt.Sprintf("**Timestamp:** %s\n\n", result.Timestamp))
	sb.WriteString("**Status:** ")
	if result.Success {
		sb.WriteString("✅ PASSED\n\n")
	} else {
		sb.WriteString("❌ FAILED\n\n")
	}
	sb.WriteString(fmt.Sprintf("**Total Findings:** %d\n\n", len(result.Findings)))

	// Summary section
	if len(result.Summary) > 0 {
		sb.WriteString("## Summary by Type\n\n")
		keys := make([]string, 0, len(result.Summary))
		for k := range result.Summary {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("- **%s:** %d\n", k, result.Summary[k]))
		}
		sb.WriteString("\n")
	}

	// Findings section
	if len(result.Findings) > 0 {
		sb.WriteString("## Findings\n\n")
		for _, finding := range result.Findings {
			sb.WriteString(fmt.Sprintf("### [%s] %s:%d\n\n",
				string(finding.Severity),
				finding.File,
				finding.Line))
			sb.WriteString(fmt.Sprintf("%s\n\n", finding.Message))
			if finding.Pattern != "" {
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", finding.Pattern))
			}
		}
	}

	return sb.String()
}

// FormatJSON formats scan results as JSON
func FormatJSON(result *Result) string {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal: %s"}`, err.Error())
	}
	return string(data)
}

// FormatText formats scan results as plain text
func FormatText(result *Result) string {
	var sb strings.Builder

	sb.WriteString("Sentinel Security Audit Report\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", result.Timestamp))
	sb.WriteString("Status: ")
	if result.Success {
		sb.WriteString("✅ PASSED\n")
	} else {
		sb.WriteString("❌ FAILED\n")
	}
	sb.WriteString(fmt.Sprintf("Total Findings: %d\n\n", len(result.Findings)))

	// Summary section
	if len(result.Summary) > 0 {
		sb.WriteString("Summary by Type:\n")
		keys := make([]string, 0, len(result.Summary))
		for k := range result.Summary {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", k, result.Summary[k]))
		}
		sb.WriteString("\n")
	}

	// Findings section
	if len(result.Findings) > 0 {
		sb.WriteString("Findings:\n")
		for i, finding := range result.Findings {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("  ... and %d more findings\n", len(result.Findings)-i))
				break
			}
			sb.WriteString(fmt.Sprintf("  [%s] %s:%d - %s\n",
				string(finding.Severity),
				finding.File,
				finding.Line,
				finding.Message))
			if finding.Pattern != "" {
				sb.WriteString(fmt.Sprintf("    Pattern: %s\n", finding.Pattern))
			}
		}
	}

	return sb.String()
}
