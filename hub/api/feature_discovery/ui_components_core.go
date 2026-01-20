// Package feature_discovery - UI component discovery core
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// discoverUIComponents discovers UI components in the codebase
// Supports React, Vue, Angular with comprehensive analysis
func discoverUIComponents(ctx context.Context, codebasePath string, featureName string, framework string) (*UILayerComponents, error) {
	components := []ComponentInfo{}

	// Determine file extensions based on framework
	extensions := getUIExtensions(framework)

	// Search for component files recursively
	for _, ext := range extensions {
		files, err := findFilesRecursively(codebasePath, "*"+ext)
		if err != nil {
			continue
		}

		for _, file := range files {
			// Skip excluded directories
			if isExcludedPath(file) {
				continue
			}

			// Check if file matches feature keywords
			fileName := filepath.Base(file)
			if matchesFeature(fileName, featureName) {
				component := extractComponentInfo(file, framework, featureName)
				if component != nil {
					components = append(components, *component)
				}
			}
		}
	}

	// Detect styling frameworks
	stylingFrameworks := detectStylingFrameworks(codebasePath)

	// Build component hierarchy
	hierarchy := buildComponentHierarchy(components, codebasePath)

	return &UILayerComponents{
		Components: components,
		Framework:  framework,
		Styling:    extractStylingNames(stylingFrameworks),
		Hierarchy:  hierarchy,
	}, nil
}

// getUIExtensions returns file extensions for the given framework
func getUIExtensions(framework string) []string {
	switch framework {
	case "react", "nextjs":
		return []string{".tsx", ".jsx", ".ts", ".js"}
	case "vue":
		return []string{".vue"}
	case "angular":
		return []string{".ts", ".component.ts"}
	default:
		// Try all common extensions
		return []string{".tsx", ".jsx", ".vue", ".ts", ".js"}
	}
}

// isExcludedPath checks if a path should be excluded from analysis
func isExcludedPath(filePath string) bool {
	excludedDirs := []string{
		"node_modules",
		"build",
		"dist",
		".next",
		".nuxt",
		"coverage",
		".git",
		"__pycache__",
	}

	for _, dir := range excludedDirs {
		if strings.Contains(filePath, dir) {
			return true
		}
	}

	return false
}

// extractComponentInfo extracts comprehensive component information
func extractComponentInfo(filePath string, framework string, featureName string) *ComponentInfo {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	content := string(data)
	fileName := filepath.Base(filePath)

	// Handle component naming based on framework
	var name string
	switch framework {
	case "angular":
		// Remove .component.ts extension for Angular
		name = strings.TrimSuffix(strings.TrimSuffix(fileName, ".component.ts"), ".component")
	default:
		name = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	}

	component := &ComponentInfo{
		Name:         name,
		Path:         filePath,
		Type:         "component",
		Framework:    framework,
		Props:        []PropInfo{},
		State:        []StateInfo{},
		Methods:      []string{},
		Dependencies: []string{},
		Metadata:     make(map[string]string),
	}

	// Extract framework-specific information
	switch framework {
	case "react", "nextjs":
		extractReactComponentInfoFromContent(component, content)
	case "vue":
		extractVueComponentInfo(component, content)
	case "angular":
		extractAngularComponentInfo(component, content)
	}

	return component
}

// matchesFeature checks if a file name matches feature keywords
func matchesFeature(fileName string, featureName string) bool {
	// If feature name is empty, match all
	if featureName == "" {
		return true
	}

	fileNameLower := strings.ToLower(fileName)
	featureLower := strings.ToLower(featureName)

	// Extract keywords from feature name
	keywords := extractFeatureKeywordsFromName(featureLower)

	for _, keyword := range keywords {
		if strings.Contains(fileNameLower, keyword) {
			return true
		}
	}

	return false
}

// extractFeatureKeywordsFromName extracts keywords from a feature name
func extractFeatureKeywordsFromName(featureName string) []string {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
	}

	words := strings.Fields(featureName)
	var keywords []string
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && !stopWords[strings.ToLower(word)] {
			keywords = append(keywords, strings.ToLower(word))
		}
	}

	return keywords
}
