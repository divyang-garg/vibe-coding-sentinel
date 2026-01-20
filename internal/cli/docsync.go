// Package cli provides doc-sync command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/divyang-garg/sentinel-hub-api/internal/scanner"
)

// DocSyncIssue represents a documentation-code sync issue
type DocSyncIssue struct {
	Type        string `json:"type"` // missing, outdated, mismatch
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// runDocSync executes the doc-sync command
func runDocSync(args []string) error {
	fmt.Println("🔄 Checking documentation-code synchronization...")

	codebasePath := "."
	if len(args) > 0 {
		codebasePath = args[0]
	}

	issues := []DocSyncIssue{}

	// Check if docs/knowledge exists
	docsPath := filepath.Join(codebasePath, "docs", "knowledge")
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		fmt.Println("⚠️  No docs/knowledge directory found")
		return nil
	}

	// Load knowledge base
	kb, err := loadKnowledge()
	if err == nil && len(kb.Entries) > 0 {
		// Check if knowledge entries reference code that exists
		issues = append(issues, checkKnowledgeReferences(kb, codebasePath)...)
	}

	// Check for code files that might need documentation
	issues = append(issues, checkMissingDocumentation(codebasePath)...)

	// Display results
	if len(issues) == 0 {
		fmt.Println("✅ Documentation and code are in sync")
		return nil
	}

	fmt.Printf("\n📊 Found %d sync issues:\n\n", len(issues))

	byType := make(map[string]int)
	for _, issue := range issues {
		byType[issue.Type]++
		fmt.Printf("[%s] %s\n", issue.Severity, issue.Description)
		if issue.File != "" {
			fmt.Printf("  File: %s", issue.File)
			if issue.Line > 0 {
				fmt.Printf(":%d", issue.Line)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	fmt.Println("Summary:")
	for typ, count := range byType {
		fmt.Printf("  %s: %d\n", typ, count)
	}

	return nil
}

// checkKnowledgeReferences checks if knowledge entries reference existing code
func checkKnowledgeReferences(kb *KnowledgeBase, codebasePath string) []DocSyncIssue {
	issues := []DocSyncIssue{}

	for _, entry := range kb.Entries {
		// Check if source file exists
		if entry.Source != "" && !strings.HasPrefix(entry.Source, "http") {
			sourcePath := filepath.Join(codebasePath, entry.Source)
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				issues = append(issues, DocSyncIssue{
					Type:        "missing",
					File:        entry.Source,
					Description: fmt.Sprintf("Knowledge entry '%s' references missing file: %s", entry.Title, entry.Source),
					Severity:    "warning",
				})
			}
		}
	}

	return issues
}

// checkMissingDocumentation checks for code that might need documentation
func checkMissingDocumentation(codebasePath string) []DocSyncIssue {
	issues := []DocSyncIssue{}

	// Run a basic scan to find potential undocumented code
	opts := scanner.ScanOptions{
		CodebasePath: codebasePath,
		CIMode:       true,
	}

	result, err := scanner.Scan(opts)
	if err != nil {
		return issues
	}

	// Check for files with many findings that might need documentation
	filesWithFindings := make(map[string]int)
	for _, finding := range result.Findings {
		filesWithFindings[finding.File]++
	}

	for file, count := range filesWithFindings {
		if count > 10 {
			// File has many issues, might need documentation
			issues = append(issues, DocSyncIssue{
				Type:        "outdated",
				File:        file,
				Description: fmt.Sprintf("File has %d findings - may need documentation update", count),
				Severity:    "info",
			})
		}
	}

	return issues
}
