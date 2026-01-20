// Pattern Learning - Output Functions
// File generation and output formatting functions
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package patterns

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/divyang-garg/sentinel-hub-api/internal/config"
)

// generateOutputFiles generates pattern output files
func generateOutputFiles(patterns *PatternData) error {
	// Create directories
	if err := os.MkdirAll(".sentinel", 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(".cursor/rules", 0755); err != nil {
		return err
	}

	// Generate patterns.json
	patternsJSON, err := json.MarshalIndent(patterns, "", "  ")
	if err != nil {
		return err
	}
	if err := config.WriteFile(".sentinel/patterns.json", string(patternsJSON)); err != nil {
		return err
	}

	// Generate project-patterns.md
	markdown := generateCursorRules(patterns)
	if err := config.WriteFile(".cursor/rules/project-patterns.md", markdown); err != nil {
		return fmt.Errorf("failed to write project patterns: %w", err)
	}

	// Generate business-rules.md if business rules are present
	if patterns.BusinessRules != nil && len(patterns.BusinessRules.Rules) > 0 {
		businessRulesMarkdown := generateBusinessRulesForCursor(patterns.BusinessRules.Rules)
		if err := config.WriteFile(".cursor/rules/business-rules.md", businessRulesMarkdown); err != nil {
			return fmt.Errorf("failed to write business rules: %w", err)
		}
	}

	return nil
}

// generateCursorRules generates Cursor rules markdown
func generateCursorRules(patterns *PatternData) string {
	var buf strings.Builder
	buf.WriteString("# Project Patterns\n\n")
	buf.WriteString("This file contains learned patterns from the codebase.\n\n")

	primaryLang := findPrimaryLanguage(patterns)
	if primaryLang != "" {
		buf.WriteString(fmt.Sprintf("## Primary Language: %s\n\n", primaryLang))
	}

	if len(patterns.Frameworks) > 0 {
		buf.WriteString("## Frameworks\n\n")
		for fw := range patterns.Frameworks {
			buf.WriteString(fmt.Sprintf("- %s\n", fw))
		}
		buf.WriteString("\n")
	}

	if len(patterns.NamingPatterns) > 0 {
		buf.WriteString("## Naming Conventions\n\n")
		for pattern := range patterns.NamingPatterns {
			buf.WriteString(fmt.Sprintf("- %s\n", pattern))
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// getKeys returns all keys from a map
func getKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
