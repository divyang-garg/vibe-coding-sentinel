// Package cli provides audit helper functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/divyang-garg/sentinel-hub-api/internal/hub"
	"github.com/divyang-garg/sentinel-hub-api/internal/scanner"
)

// saveResults saves results to a file
func saveResults(result *scanner.Result, filename, format string) error {
	var data []byte

	switch format {
	case "json":
		data = []byte(scanner.FormatJSON(result))
	case "html":
		data = []byte(scanner.FormatHTML(result))
	case "md", "markdown":
		data = []byte(scanner.FormatMarkdown(result))
	case "text":
		data = []byte(scanner.FormatText(result))
	default:
		return fmt.Errorf("unsupported output format: %s (supported: json, html, md, text)", format)
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(filename, data, 0644)
}

// getHubURL returns the Hub URL from environment or config
func getHubURL() string {
	if url := os.Getenv("SENTINEL_HUB_URL"); url != "" {
		return url
	}
	return "http://localhost:8080" // Default
}

// getAPIKey returns the API key from environment or config
func getAPIKey() string {
	return os.Getenv("SENTINEL_HUB_API_KEY") // Empty if not set
}

// performHubAnalysis performs Hub-based AST analysis
func performHubAnalysis(client *hub.Client, codebasePath string, verbose bool) ([]scanner.Finding, error) {
	findings := make([]scanner.Finding, 0)

	// Collect code files for analysis
	files, err := collectCodeFiles(codebasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}

	if verbose {
		fmt.Printf("  Analyzing %d files with Hub AST engine...\n", len(files))
	}

	// Analyze each file (in production, would batch this)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Determine language
		lang := detectLanguage(file)
		if lang == "" {
			continue
		}

		// Request AST analysis
		req := &hub.ASTAnalysisRequest{
			Code:     string(content),
			Language: lang,
			Analyses: []string{"duplicate_functions", "complexity", "security"},
		}

		resp, err := client.AnalyzeAST(req)
		if err != nil {
			// Hub error, skip this file
			continue
		}

		// Convert Hub findings to scanner findings
		for _, astFinding := range resp.Findings {
			findings = append(findings, scanner.Finding{
				Type:     astFinding.Type,
				Severity: convertHubSeverity(astFinding.Severity),
				File:     file,
				Line:     astFinding.Line,
				Message:  astFinding.Message,
				Pattern:  "", // Hub findings don't have pattern field
			})
		}
	}

	return findings, nil
}

// collectCodeFiles collects all code files in the path
func collectCodeFiles(basePath string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if isCodeFile(filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// detectLanguage detects programming language from file extension
func detectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)
	langMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".ts":   "typescript",
		".py":   "python",
		".java": "java",
		".rb":   "ruby",
		".php":  "php",
	}
	return langMap[ext]
}

// isCodeFile checks if extension is a code file
func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".py": true, ".java": true, ".rb": true, ".php": true,
	}
	return codeExts[ext]
}

// convertHubSeverity converts Hub severity to scanner severity
func convertHubSeverity(hubSeverity string) scanner.Severity {
	switch hubSeverity {
	case "critical":
		return scanner.SeverityCritical
	case "high":
		return scanner.SeverityHigh
	case "medium":
		return scanner.SeverityMedium
	case "low":
		return scanner.SeverityLow
	default:
		return scanner.SeverityWarning
	}
}

// mergeHubFindings merges Hub AST findings into scanner results
func mergeHubFindings(result *scanner.Result, hubFindings []scanner.Finding) *scanner.Result {
	// Add Hub findings to result
	result.Findings = append(result.Findings, hubFindings...)

	// Recalculate summary
	for _, finding := range hubFindings {
		result.Summary[finding.Type]++
		if finding.Severity == scanner.SeverityCritical {
			result.Success = false
		}
	}

	return result
}
