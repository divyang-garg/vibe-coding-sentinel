// Package feature_discovery - UI styling framework detection
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// detectStylingFrameworks detects CSS frameworks used in the project
func detectStylingFrameworks(codebasePath string) []StylingFramework {
	var frameworks []StylingFramework

	// Check package.json for CSS framework dependencies
	packageJSONPath := filepath.Join(codebasePath, "package.json")
	if _, err := os.Stat(packageJSONPath); err == nil {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
					// Check for Tailwind CSS
					if version, ok := deps["tailwindcss"].(string); ok {
						frameworks = append(frameworks, StylingFramework{
							Name:    "Tailwind CSS",
							Version: version,
							Type:    "utility",
							Usage:   "utility",
						})
					}

					// Check for styled-components
					if version, ok := deps["styled-components"].(string); ok {
						frameworks = append(frameworks, StylingFramework{
							Name:    "styled-components",
							Version: version,
							Type:    "css-in-js",
							Usage:   "component",
						})
					}

					// Check for SCSS/Sass
					if _, ok := deps["sass"].(string); ok || deps["node-sass"] != nil {
						frameworks = append(frameworks, StylingFramework{
							Name:  "SCSS/Sass",
							Type:  "scss",
							Usage: "global",
						})
					}

					// Check for Less
					if _, ok := deps["less"].(string); ok {
						frameworks = append(frameworks, StylingFramework{
							Name:  "Less",
							Type:  "less",
							Usage: "global",
						})
					}
				}
			}
		}
	}

	// Check for CSS files and analyze their content
	cssFiles, _ := findFilesRecursively(codebasePath, "*.css")
	scssFiles, _ := findFilesRecursively(codebasePath, "*.scss")
	lessFiles, _ := findFilesRecursively(codebasePath, "*.less")

	// Analyze CSS files for framework patterns
	for _, files := range [][]string{cssFiles, scssFiles, lessFiles} {
		for _, file := range files {
			if !isExcludedPath(file) {
				content, err := os.ReadFile(file)
				if err == nil {
					framework := analyzeCSSContent(string(content), filepath.Ext(file))
					if framework != nil {
						frameworks = append(frameworks, *framework)
					}
				}
			}
		}
	}

	return frameworks
}

// analyzeCSSContent analyzes CSS content for framework patterns
func analyzeCSSContent(content string, ext string) *StylingFramework {
	lines := strings.Split(content, "\n")

	// Check for Tailwind utility classes
	tailwindClasses := []string{"flex", "grid", "text-center", "bg-", "text-", "p-", "m-", "w-", "h-"}
	tailwindCount := 0

	for _, line := range lines {
		for _, class := range tailwindClasses {
			if strings.Contains(line, class) {
				tailwindCount++
				if tailwindCount > 5 { // Threshold for considering it Tailwind
					return &StylingFramework{
						Name:  "Tailwind CSS",
						Type:  "utility",
						Usage: "utility",
					}
				}
			}
		}
	}

	// Check for SCSS features
	if ext == ".scss" && (strings.Contains(content, "&") || strings.Contains(content, "@mixin")) {
		return &StylingFramework{
			Name:  "SCSS/Sass",
			Type:  "scss",
			Usage: "global",
		}
	}

	return nil
}

// extractStylingNames extracts framework names from StylingFramework slice
func extractStylingNames(frameworks []StylingFramework) []string {
	names := make([]string, len(frameworks))
	for i, fw := range frameworks {
		names[i] = fw.Name
	}
	return names
}
