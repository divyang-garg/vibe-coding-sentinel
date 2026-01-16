// Package framework_detection provides framework detection capabilities
// Complies with CODING_STANDARDS.md: Framework detection max 300 lines
package feature_discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// detectUIFramework detects the UI framework used in the codebase
func detectUIFramework(codebasePath string) (string, string, error) {
	packageJSONPath := filepath.Join(codebasePath, "package.json")

	// Check package.json for React, Vue, Angular
	if _, err := os.Stat(packageJSONPath); err == nil {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
					// Check for React
					if react, ok := deps["react"].(string); ok {
						// Check for Next.js
						if _, hasNext := deps["next"]; hasNext {
							return "nextjs", react, nil
						}
						return "react", react, nil
					}
					// Check for Vue
					if vue, ok := deps["vue"].(string); ok {
						return "vue", vue, nil
					}
					// Check for Angular
					if angular, ok := deps["@angular/core"].(string); ok {
						return "angular", angular, nil
					}
				}
			}
		}
	}

	// Check for framework-specific config files
	if _, err := os.Stat(filepath.Join(codebasePath, "vite.config.js")); err == nil {
		return "vite", "", nil // Could be Vue or React with Vite
	}
	if _, err := os.Stat(filepath.Join(codebasePath, "vite.config.ts")); err == nil {
		return "vite", "", nil
	}
	if _, err := os.Stat(filepath.Join(codebasePath, "next.config.js")); err == nil {
		return "nextjs", "", nil
	}
	if _, err := os.Stat(filepath.Join(codebasePath, "angular.json")); err == nil {
		return "angular", "", nil
	}

	return "unknown", "", nil
}

// detectAPIFramework detects the API framework used in the codebase
func detectAPIFramework(codebasePath string) (string, error) {
	packageJSONPath := filepath.Join(codebasePath, "package.json")

	// Check package.json for Express (Node.js)
	if _, err := os.Stat(packageJSONPath); err == nil {
		data, err := os.ReadFile(packageJSONPath)
		if err == nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
					if _, ok := deps["express"]; ok {
						return "express", nil
					}
				}
			}
		}
	}

	// Check for Python frameworks
	requirementsPath := filepath.Join(codebasePath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		data, err := os.ReadFile(requirementsPath)
		if err == nil {
			content := string(data)
			if strings.Contains(content, "fastapi") {
				return "fastapi", nil
			}
			if strings.Contains(content, "django") {
				return "django", nil
			}
			if strings.Contains(content, "flask") {
				return "flask", nil
			}
		}
	}

	// Check for Go files (Gin router)
	goFiles, _ := filepath.Glob(filepath.Join(codebasePath, "**/*.go"))
	if len(goFiles) > 0 {
		// Check for Gin imports
		for _, file := range goFiles {
			data, err := os.ReadFile(file)
			if err == nil {
				content := string(data)
				if strings.Contains(content, "github.com/gin-gonic/gin") {
					return "gin", nil
				}
				if strings.Contains(content, "github.com/go-chi/chi") {
					return "chi", nil
				}
			}
		}
	}

	return "unknown", nil
}

// detectFrameworks performs comprehensive framework detection
func detectFrameworks(codebasePath string) (*FeatureDiscovery, error) {
	// Extract from main feature_discovery.go
	// Implementation will be copied here
	return &FeatureDiscovery{}, nil
}
