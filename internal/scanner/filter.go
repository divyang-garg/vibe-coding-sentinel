// Package scanner provides false positive filtering
// Complies with CODING_STANDARDS.md: Utilities max 250 lines
package scanner

// isInComment checks if a match is inside a comment
func isInComment(line string, match string) bool {
	commentPatterns := []string{"//", "/*", "#", "--", "*"}
	matchIdx := findStringIndex(line, match)
	if matchIdx < 0 {
		return false
	}

	// Check if the pattern itself starts with a comment marker
	for _, cp := range commentPatterns {
		if hasPrefix(match, cp) {
			return true
		}
	}

	// Check if there's a comment marker before the match
	for _, cp := range commentPatterns {
		commentIdx := findStringIndex(line, cp)
		if commentIdx >= 0 && commentIdx < matchIdx {
			// Check if it's a real comment (not in a string)
			if !isInStringLiteral(line[:commentIdx]) {
				return true
			}
		}
	}

	return false
}

// isInStringLiteral checks if the end of a line is inside a string literal
func isInStringLiteral(line string) bool {
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false
	escaped := false

	for i := 0; i < len(line); i++ {
		if escaped {
			escaped = false
			continue
		}

		if line[i] == '\\' {
			escaped = true
			continue
		}

		if line[i] == '\'' && !inDoubleQuote && !inBacktick {
			inSingleQuote = !inSingleQuote
		} else if line[i] == '"' && !inSingleQuote && !inBacktick {
			inDoubleQuote = !inDoubleQuote
		} else if line[i] == '`' && !inSingleQuote && !inDoubleQuote {
			inBacktick = !inBacktick
		}
	}

	return inSingleQuote || inDoubleQuote || inBacktick
}

// isInStringLiteralAt checks if a specific position is inside a string literal
func isInStringLiteralAt(line string, pos int) bool {
	if pos < 0 || pos >= len(line) {
		return false
	}

	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for i := 0; i < pos && i < len(line); i++ {
		if escaped {
			escaped = false
			continue
		}

		if line[i] == '\\' {
			escaped = true
			continue
		}

		if line[i] == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		} else if line[i] == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		}
	}

	return inSingleQuote || inDoubleQuote
}

// isInDocComment checks if a match is in a documentation comment
func isInDocComment(lines []string, lineNum int, match string) bool {
	// Check if preceding lines are doc comments
	docPatterns := []string{"///", "/**", "#", "//"}
	for i := lineNum - 1; i >= 0 && i >= lineNum-3; i-- {
		if i >= len(lines) {
			continue
		}
		line := lines[i]
		for _, dp := range docPatterns {
			if hasPrefix(line, dp) {
				return true
			}
		}
	}
	return false
}

// shouldFilterFinding determines if a finding should be filtered as false positive
func shouldFilterFinding(finding Finding, line string, allLines []string) bool {
	// Filter if in comment
	if isInComment(line, finding.Pattern) {
		return true
	}

	// Filter if in string literal (for test/doc files)
	if isInStringLiteralAt(line, finding.Column-1) {
		// Allow secrets in string literals to be detected, but filter other patterns
		if finding.Type != "secrets" && finding.Type != "high_entropy_secret" {
			return true
		}
	}

	// Filter if in documentation comment
	if isInDocComment(allLines, finding.Line-1, finding.Pattern) {
		return true
	}

	return false
}

// FilterFalsePositives filters out findings that are likely false positives
func FilterFalsePositives(findings []Finding, fileContent string) []Finding {
	lines := splitLines(fileContent)
	filtered := make([]Finding, 0, len(findings))

	for _, finding := range findings {
		lineIdx := finding.Line - 1
		if lineIdx < 0 || lineIdx >= len(lines) {
			// Keep finding if we can't verify
			filtered = append(filtered, finding)
			continue
		}

		line := lines[lineIdx]
		if !shouldFilterFinding(finding, line, lines) {
			filtered = append(filtered, finding)
		}
	}

	return filtered
}

// Helper functions
func hasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}
