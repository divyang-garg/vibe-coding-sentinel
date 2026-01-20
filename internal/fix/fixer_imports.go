// Package fix provides import sorting functionality
// Complies with CODING_STANDARDS.md: Business services max 400 lines
package fix

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// sortImports sorts import statements by language
func sortImports(content, path string, fixesApplied *int, modified bool) (string, bool) {
	ext := filepath.Ext(path)

	switch ext {
	case ".go":
		return sortGoImports(content, path, fixesApplied, modified)
	case ".ts", ".tsx", ".js", ".jsx":
		return sortJSImports(content, path, fixesApplied, modified)
	case ".py":
		// Python import sorting deferred - more complex
		return content, modified
	default:
		return content, modified
	}
}

// sortGoImports sorts Go imports into standard library, external, and internal groups
func sortGoImports(content, path string, fixesApplied *int, modified bool) (string, bool) {
	lines := strings.Split(content, "\n")
	var importStart int = -1
	var inImportBlock bool
	var importLines []string
	var beforeImports []string
	importBlockEnd := -1

	// Find import block
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "import ") {
			if !inImportBlock {
				importStart = i
				inImportBlock = true
				beforeImports = lines[:i]

				if trimmed == "import (" {
					// Multi-line import block - continue collecting
					continue
				} else if strings.HasPrefix(trimmed, "import \"") || strings.HasPrefix(trimmed, "import `") {
					// Single import statement
					importLines = append(importLines, strings.TrimPrefix(strings.TrimPrefix(trimmed, "import "), "import "))
					importBlockEnd = i + 1
					break
				}
			}
		} else if inImportBlock {
			if trimmed == ")" {
				importBlockEnd = i + 1
				break
			} else if strings.HasPrefix(trimmed, "\"") || strings.HasPrefix(trimmed, "`") {
				importLines = append(importLines, trimmed)
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
				// End of imports
				importBlockEnd = i
				break
			}
		}
	}

	if importStart == -1 || len(importLines) == 0 {
		return content, modified
	}

	if importBlockEnd == -1 {
		importBlockEnd = len(lines)
	}
	afterImports := lines[importBlockEnd:]

	// Sort imports: stdlib (no dots or starting with dots), external (has domain), internal (relative)
	var stdlib, external, internal []string

	for _, imp := range importLines {
		clean := strings.Trim(imp, "\"`\t ")
		if strings.HasPrefix(clean, ".") || strings.HasPrefix(clean, "/") {
			internal = append(internal, imp)
		} else if strings.Contains(clean, ".") && !strings.HasPrefix(clean, "std") {
			external = append(external, imp)
		} else {
			stdlib = append(stdlib, imp)
		}
	}

	// Sort each group
	sort.Strings(stdlib)
	sort.Strings(external)
	sort.Strings(internal)

	// Rebuild import block
	var newImportLines []string
	if len(stdlib) > 0 {
		newImportLines = append(newImportLines, stdlib...)
	}
	if len(external) > 0 {
		if len(newImportLines) > 0 {
			newImportLines = append(newImportLines, "") // blank line separator
		}
		newImportLines = append(newImportLines, external...)
	}
	if len(internal) > 0 {
		if len(newImportLines) > 0 {
			newImportLines = append(newImportLines, "") // blank line separator
		}
		newImportLines = append(newImportLines, internal...)
	}

	// Reconstruct file
	var newLines []string
	newLines = append(newLines, beforeImports...)
	newLines = append(newLines, "import (")
	for _, imp := range newImportLines {
		if imp == "" {
			newLines = append(newLines, "")
		} else {
			newLines = append(newLines, "\t"+imp)
		}
	}
	newLines = append(newLines, ")")
	newLines = append(newLines, afterImports...)

	fmt.Printf("  Sort imports in %s\n", path)
	(*fixesApplied)++
	return strings.Join(newLines, "\n"), true
}

// sortJSImports sorts JavaScript/TypeScript imports
func sortJSImports(content, path string, fixesApplied *int, modified bool) (string, bool) {
	lines := strings.Split(content, "\n")

	var importLines []string
	var otherLines []string
	var firstNonImport int = -1
	inImports := true

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inImports && (strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "import{")) {
			importLines = append(importLines, line)
		} else if inImports && trimmed == "" && len(importLines) > 0 {
			// Allow blank lines within imports
			importLines = append(importLines, line)
		} else if inImports && len(importLines) > 0 {
			inImports = false
			firstNonImport = i
			otherLines = append(otherLines, line)
		} else {
			otherLines = append(otherLines, line)
		}
	}

	if len(importLines) <= 1 {
		return content, modified
	}

	// Separate external and relative imports
	var externalImports, relativeImports []string

	for _, imp := range importLines {
		trimmed := strings.TrimSpace(imp)
		if trimmed == "" {
			continue
		}
		// Check if it's a relative import (starts with ./ or ../)
		if strings.Contains(imp, "from './") || strings.Contains(imp, "from \"./") ||
			strings.Contains(imp, "from '../") || strings.Contains(imp, "from \"../") {
			relativeImports = append(relativeImports, imp)
		} else {
			externalImports = append(externalImports, imp)
		}
	}

	// Sort each group
	sort.Strings(externalImports)
	sort.Strings(relativeImports)

	// Check if already sorted
	var sortedImports []string
	sortedImports = append(sortedImports, externalImports...)
	if len(externalImports) > 0 && len(relativeImports) > 0 {
		sortedImports = append(sortedImports, "") // blank line separator
	}
	sortedImports = append(sortedImports, relativeImports...)

	// Reconstruct file
	if firstNonImport > 0 {
		var newLines []string
		newLines = append(newLines, sortedImports...)
		if len(sortedImports) > 0 {
			newLines = append(newLines, "") // blank line before rest of code
		}
		newLines = append(newLines, lines[firstNonImport:]...)

		fmt.Printf("  Sort imports in %s\n", path)
		(*fixesApplied)++
		return strings.Join(newLines, "\n"), true
	}

	return content, modified
}
