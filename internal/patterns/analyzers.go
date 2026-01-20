// Package patterns provides code analysis functionality
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package patterns

import (
	"os"
	"path/filepath"
	"strings"
)

// analyzeImportPatterns analyzes import patterns in code files
func analyzeImportPatterns(path, content string, patterns *PatternData) {
	ext := filepath.Ext(path)
	lines := strings.Split(content, "\n")

	// Only analyze supported languages
	if ext != ".js" && ext != ".jsx" && ext != ".ts" && ext != ".tsx" && ext != ".py" && ext != ".go" {
		return
	}

	absoluteCount := 0
	relativeCount := 0
	defaultCount := 0
	namedCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect import statements based on language
		if ext == ".go" && strings.HasPrefix(trimmed, "import ") {
			// Go imports - check for absolute vs relative (rare in Go, usually absolute)
			if strings.Contains(trimmed, "\"") || strings.Contains(trimmed, "`") {
				absoluteCount++
				namedCount++
			}
		} else if (ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx") && strings.Contains(trimmed, "import ") {
			// JavaScript/TypeScript imports
			if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "import{") {
				// Check if absolute (starts with @/, /, or package name) or relative (starts with ./ or ../)
				if strings.Contains(trimmed, "from") {
					fromPart := strings.Split(trimmed, "from")
					if len(fromPart) > 1 {
						source := strings.TrimSpace(fromPart[1])
						source = strings.Trim(source, "'\"`;")
						if strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") {
							relativeCount++
						} else if !strings.HasPrefix(source, "@") {
							absoluteCount++
						}
					}
				}

				// Check for default vs named imports
				if strings.Contains(trimmed, "import ") && strings.Contains(trimmed, "from") {
					importPart := strings.Split(trimmed, "from")[0]
					if strings.Contains(importPart, "{") {
						namedCount++
					} else if !strings.Contains(importPart, "*") {
						defaultCount++
					}
				}

				// Collect examples (limit to 5)
				if len(patterns.ImportPatterns.Examples) < 5 {
					if len(trimmed) < 100 {
						patterns.ImportPatterns.Examples = append(patterns.ImportPatterns.Examples, trimmed)
					}
				}
			}
		} else if ext == ".py" && strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
			// Python imports - usually absolute unless explicitly relative
			if strings.HasPrefix(trimmed, "from .") || strings.HasPrefix(trimmed, "from ..") {
				relativeCount++
			} else {
				absoluteCount++
			}

			if strings.HasPrefix(trimmed, "import ") {
				defaultCount++
			} else {
				namedCount++
			}
		}
	}

	// Determine dominant style
	total := absoluteCount + relativeCount
	if total > 0 {
		if absoluteCount > relativeCount*2 {
			patterns.ImportPatterns.Style = "absolute"
		} else if relativeCount > absoluteCount*2 {
			patterns.ImportPatterns.Style = "relative"
		} else {
			patterns.ImportPatterns.Style = "mixed"
		}
	}

	patterns.ImportPatterns.DefaultImports += defaultCount
	patterns.ImportPatterns.NamedImports += namedCount

	// Detect barrel files (index.ts, index.js)
	filename := filepath.Base(path)
	if filename == "index.ts" || filename == "index.js" || filename == "index.tsx" || filename == "index.jsx" {
		dir := filepath.Dir(path)
		if !contains(patterns.ImportPatterns.BarrelFiles, dir) {
			patterns.ImportPatterns.BarrelFiles = append(patterns.ImportPatterns.BarrelFiles, dir)
		}
	}
}

// analyzeCodeStyle analyzes code style patterns
func analyzeCodeStyle(path, content string, patterns *PatternData) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return
	}

	// Analyze indentation
	tabCount := 0
	space2Count := 0
	space4Count := 0
	singleQuoteCount := 0
	doubleQuoteCount := 0
	semicolonCount := 0
	noSemicolonCount := 0

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Check indentation (first non-empty line)
		firstChar := line[0]
		if firstChar == '\t' {
			tabCount++
		} else if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
			space2Count++
		} else if strings.HasPrefix(line, "    ") {
			space4Count++
		}

		// Check quotes (for JS/TS/Python)
		ext := filepath.Ext(path)
		if ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" || ext == ".py" {
			singleQuoteCount += strings.Count(line, "'")
			doubleQuoteCount += strings.Count(line, "\"")

			// Check semicolons (for JS/TS)
			if ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" {
				trimmed := strings.TrimSpace(line)
				if strings.HasSuffix(trimmed, ";") && !strings.HasPrefix(trimmed, "//") {
					semicolonCount++
				} else if len(trimmed) > 0 && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "/*") {
					noSemicolonCount++
				}
			}
		}
	}

	// Determine dominant style
	if tabCount > space2Count && tabCount > space4Count {
		patterns.CodeStyle.IndentStyle = "tabs"
	} else if space2Count > space4Count {
		patterns.CodeStyle.IndentStyle = "spaces"
		patterns.CodeStyle.IndentSize = 2
	} else if space4Count > 0 {
		patterns.CodeStyle.IndentStyle = "spaces"
		patterns.CodeStyle.IndentSize = 4
	}

	// Compare counts with 20% threshold
	if singleQuoteCount > 0 && doubleQuoteCount == 0 {
		patterns.CodeStyle.QuoteStyle = "single"
	} else if doubleQuoteCount > 0 && singleQuoteCount == 0 {
		patterns.CodeStyle.QuoteStyle = "double"
	} else if singleQuoteCount > doubleQuoteCount {
		patterns.CodeStyle.QuoteStyle = "single"
	} else if doubleQuoteCount > singleQuoteCount {
		patterns.CodeStyle.QuoteStyle = "double"
	}

	if semicolonCount > noSemicolonCount*2 {
		patterns.CodeStyle.Semicolons = "always"
	} else if noSemicolonCount > semicolonCount*2 {
		patterns.CodeStyle.Semicolons = "never"
	} else {
		patterns.CodeStyle.Semicolons = "optional"
	}

	// Check line endings
	if strings.Contains(content, "\r\n") {
		patterns.CodeStyle.LineEnding = "crlf"
	} else {
		patterns.CodeStyle.LineEnding = "lf"
	}
}

// analyzeFolderStructure analyzes folder structure patterns
func analyzeFolderStructure(codebasePath string, patterns *PatternData) {
	// Common folder structure patterns
	structurePatterns := map[string]string{
		"components":  "src/components/",
		"features":    "src/features/",
		"services":    "src/services/",
		"utils":       "src/utils/",
		"hooks":       "src/hooks/",
		"pages":       "src/pages/",
		"routes":      "src/routes/",
		"middleware":  "src/middleware/",
		"models":      "src/models/",
		"controllers": "src/controllers/",
		"views":       "src/views/",
		"tests":       "tests/",
		"test":        "test/",
		"__tests__":   "__tests__/",
	}

	err := filepath.Walk(codebasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}

		// Skip common directories
		if shouldSkipPath(path) {
			return filepath.SkipDir
		}

		// Check if directory matches known patterns
		dirName := filepath.Base(path)
		for pattern, prefix := range structurePatterns {
			if strings.Contains(path, prefix) || dirName == pattern {
				examples := patterns.ProjectStructure[pattern]
				if !contains(examples, path) && len(examples) < 10 {
					patterns.ProjectStructure[pattern] = append(examples, path)
				}
			}
		}

		return nil
	})

	if err != nil {
		// Non-fatal error
		return
	}
}
