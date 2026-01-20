// Package feature_discovery - UI component hierarchy building
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"os"
	"regexp"
	"strings"
)

// buildComponentHierarchy builds a component dependency tree
func buildComponentHierarchy(components []ComponentInfo, codebasePath string) ComponentTree {
	tree := ComponentTree{
		Imports: make(map[string][]string),
		Exports: make(map[string][]string),
	}

	// Analyze imports and exports for each component
	for _, component := range components {
		filePath := component.Path

		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		contentStr := string(content)

		// Extract imports
		importMatches := regexp.MustCompile(`import\s+.*?\s+from\s+['"]([^'"]+)['"]`).FindAllStringSubmatch(contentStr, -1)
		var imports []string
		for _, match := range importMatches {
			if len(match) > 1 {
				importPath := strings.Trim(match[1], "./")
				if !strings.HasPrefix(importPath, "react") && !strings.HasPrefix(importPath, "@types") {
					imports = append(imports, importPath)
				}
			}
		}
		tree.Imports[filePath] = imports

		// Extract exports
		exportMatches := regexp.MustCompile(`export\s+(?:default\s+)?(?:const|function|class)\s+(\w+)`).FindAllStringSubmatch(contentStr, -1)
		var exports []string
		for _, match := range exportMatches {
			if len(match) > 1 {
				exports = append(exports, match[1])
			}
		}
		tree.Exports[filePath] = exports
	}

	return tree
}
