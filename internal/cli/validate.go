// Package cli provides validate-rules command implementation
// Complies with CODING_STANDARDS.md: CLI handlers max 300 lines
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runValidateRules validates Cursor rules
func runValidateRules(args []string) error {
	rulesDir := ".cursor/rules"

	// Check if rules directory exists
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return fmt.Errorf("rules directory not found: %s\n\nRun 'sentinel init' to create it.", rulesDir)
	}

	fmt.Println("🔍 Validating Cursor rules...")

	// Read all rule files
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return fmt.Errorf("unable to read rules directory: %w", err)
	}

	validCount := 0
	invalidCount := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		rulePath := filepath.Join(rulesDir, entry.Name())
		if err := validateRuleFile(rulePath); err != nil {
			fmt.Printf("❌ %s: %v\n", entry.Name(), err)
			invalidCount++
		} else {
			fmt.Printf("✅ %s\n", entry.Name())
			validCount++
		}
	}

	fmt.Printf("\n📊 Validation complete: %d valid, %d invalid\n", validCount, invalidCount)

	if invalidCount > 0 {
		return fmt.Errorf("validation completed with errors found")
	}

	return nil
}

// validateRuleFile validates a single rule file
func validateRuleFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read rule file: %w", err)
	}

	content := string(data)

	// Check for YAML frontmatter
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Errorf("missing YAML frontmatter")
	}

	// Extract frontmatter
	parts := strings.SplitN(content[4:], "\n---\n", 2)
	if len(parts) < 2 {
		return fmt.Errorf("malformed YAML frontmatter")
	}

	frontmatter := parts[0]

	// Check for required fields
	requiredFields := []string{"description:", "globs:", "alwaysApply:"}
	for _, field := range requiredFields {
		if !strings.Contains(frontmatter, field) {
			return fmt.Errorf("missing required field: %s", strings.TrimSuffix(field, ":"))
		}
	}

	return nil
}
