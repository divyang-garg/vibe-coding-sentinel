// Package feature_discovery provides comprehensive UI component discovery
// Complies with CODING_STANDARDS.md: UI components max 300 lines
package feature_discovery

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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
		extractReactComponentInfo(component, content)
	case "vue":
		extractVueComponentInfo(component, content)
	case "angular":
		extractAngularComponentInfo(component, content)
	}

	return component
}

// extractReactComponentInfo extracts React component information
func extractReactComponentInfo(component *ComponentInfo, content string) {
	// Detect component type
	if isReactFunctionComponent(content) {
		component.Type = "functional"
		component.Metadata["pattern"] = "function"
	} else if isReactClassComponent(content) {
		component.Type = "class"
		component.Metadata["pattern"] = "class"
	} else {
		// Default to component if type detection fails
		component.Type = "component"
	}

	// Extract props
	component.Props = extractReactProps(content)

	// Extract state
	component.State = extractReactState(content)

	// Extract methods
	component.Methods = extractReactMethods(content)

	// Detect special types
	if isReactFormComponent(content) {
		component.Type = "form"
	}
	if isReactPageComponent(content) {
		component.Type = "page"
	}

	// Extract dependencies
	component.Dependencies = extractReactDependencies(content)
}

// extractVueComponentInfo extracts Vue component information
func extractVueComponentInfo(component *ComponentInfo, content string) {
	// Detect component type
	if strings.Contains(content, "export default") {
		component.Type = "component"
	}

	// Extract props from props definition
	if propsMatch := regexp.MustCompile(`props:\s*\{([^}]+)\}`).FindStringSubmatch(content); len(propsMatch) > 1 {
		propsStr := propsMatch[1]
		propNames := regexp.MustCompile(`(\w+):`).FindAllStringSubmatch(propsStr, -1)
		for _, match := range propNames {
			if len(match) > 1 {
				component.Props = append(component.Props, PropInfo{Name: match[1]})
			}
		}
	}

	// Extract data properties as state
	if dataMatch := regexp.MustCompile(`data\(\)\s*\{[^}]*return\s*\{([^}]+)\}`).FindStringSubmatch(content); len(dataMatch) > 1 {
		dataStr := dataMatch[1]
		stateNames := regexp.MustCompile(`(\w+):`).FindAllStringSubmatch(dataStr, -1)
		for _, match := range stateNames {
			if len(match) > 1 {
				component.State = append(component.State, StateInfo{Name: match[1]})
			}
		}
	}

	// Extract methods
	if methodsMatch := regexp.MustCompile(`methods:\s*\{([^}]+)\}`).FindStringSubmatch(content); len(methodsMatch) > 1 {
		methodsStr := methodsMatch[1]
		methodNames := regexp.MustCompile(`(\w+)\(`).FindAllStringSubmatch(methodsStr, -1)
		for _, match := range methodNames {
			if len(match) > 1 {
				component.Methods = append(component.Methods, match[1])
			}
		}
	}

	// Detect form components
	if isVueFormComponent(content) {
		component.Type = "form"
	}
}

// extractAngularComponentInfo extracts Angular component information
func extractAngularComponentInfo(component *ComponentInfo, content string) {
	// Check for @Component decorator
	if strings.Contains(content, "@Component") {
		component.Type = "component"
	}

	// Extract selector from @Component decorator
	if selectorMatch := regexp.MustCompile(`selector:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(selectorMatch) > 1 {
		component.Metadata["selector"] = selectorMatch[1]
	}

	// Extract templateUrl and styleUrls
	if templateMatch := regexp.MustCompile(`templateUrl:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(templateMatch) > 1 {
		component.Metadata["template"] = templateMatch[1]
	}

	if styleMatch := regexp.MustCompile(`styleUrls:\s*\[([^\]]+)\]`).FindStringSubmatch(content); len(styleMatch) > 1 {
		component.Metadata["styles"] = strings.TrimSpace(styleMatch[1])
	}

	// Extract methods - look for function declarations in the class
	methodMatches := regexp.MustCompile(`(\w+)\s*\([^)]*\)\s*:\s*void\s*\{`).FindAllStringSubmatch(content, -1)
	for _, match := range methodMatches {
		if len(match) > 1 {
			methodName := match[1]
			// Skip lifecycle methods and constructor
			if methodName != "constructor" && !strings.HasPrefix(methodName, "ng") {
				component.Methods = append(component.Methods, methodName)
			}
		}
	}

	// Detect form components
	if isAngularFormComponent(content) {
		component.Type = "form"
	}
}

// Helper functions for component type detection
func isReactFunctionComponent(content string) bool {
	// Check for function component pattern: function ComponentName or const ComponentName =
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "function ") && strings.Contains(line, "React.FC") {
			return true
		}
		if strings.HasPrefix(line, "const ") && strings.Contains(line, " = ") && strings.Contains(content, "React.FC") {
			return true
		}
	}
	return false
}

func isReactClassComponent(content string) bool {
	return strings.Contains(content, "class ") && strings.Contains(content, "extends React.Component")
}

func isReactFormComponent(content string) bool {
	return strings.Contains(content, "react-hook-form") ||
		strings.Contains(content, "formik") ||
		strings.Contains(content, "useForm")
}

func isReactPageComponent(content string) bool {
	fileName := filepath.Base(strings.TrimSuffix(content, filepath.Ext(content)))
	return strings.Contains(strings.ToLower(fileName), "page") ||
		strings.Contains(strings.ToLower(fileName), "index") ||
		strings.Contains(content, "export default function")
}

func isVueFormComponent(content string) bool {
	return strings.Contains(content, "vuelidate") ||
		strings.Contains(content, "vee-validate") ||
		strings.Contains(content, "vue-form")
}

func isAngularFormComponent(content string) bool {
	return strings.Contains(content, "FormGroup") ||
		strings.Contains(content, "ngForm") ||
		strings.Contains(content, "FormBuilder")
}

// extractReactProps extracts props from React component
func extractReactProps(content string) []PropInfo {
	var props []PropInfo

	// Extract from TypeScript interface
	if propInterface := regexp.MustCompile(`interface\s+\w+Props\s*\{([^}]+)\}`).FindStringSubmatch(content); len(propInterface) > 1 {
		propStr := propInterface[1]
		propMatches := regexp.MustCompile(`(\w+)(?:\?\s*)?:\s*([^;]+)`).FindAllStringSubmatch(propStr, -1)
		for _, match := range propMatches {
			if len(match) >= 3 {
				prop := PropInfo{
					Name: match[1],
					Type: match[2],
				}
				// Check if optional
				if strings.Contains(match[0], "?") {
					prop.Required = false
				} else {
					prop.Required = true
				}
				props = append(props, prop)
			}
		}
	}

	// Extract from destructured parameters
	if paramMatch := regexp.MustCompile(`\{([^}]+)\}`).FindStringSubmatch(content); len(paramMatch) > 1 {
		paramStr := paramMatch[1]
		paramNames := regexp.MustCompile(`(\w+)(?:\s*=\s*[^,}]*)?`).FindAllStringSubmatch(paramStr, -1)
		for _, match := range paramNames {
			if len(match) > 1 {
				found := false
				for _, prop := range props {
					if prop.Name == match[1] {
						found = true
						break
					}
				}
				if !found {
					props = append(props, PropInfo{Name: match[1]})
				}
			}
		}
	}

	return props
}

// extractReactState extracts state from React component
func extractReactState(content string) []StateInfo {
	var state []StateInfo

	// Extract useState calls
	stateMatches := regexp.MustCompile(`(?:const|let)\s*\[(\w+),\s*set\w+\]\s*=\s*useState\(([^)]*)\)`).FindAllStringSubmatch(content, -1)
	for _, match := range stateMatches {
		if len(match) >= 3 {
			state = append(state, StateInfo{
				Name:  match[1],
				Value: match[2],
			})
		}
	}

	// Extract this.state references (class components)
	if stateMatch := regexp.MustCompile(`this\.state\s*=\s*\{([^}]+)\}`).FindStringSubmatch(content); len(stateMatch) > 1 {
		stateStr := stateMatch[1]
		stateNames := regexp.MustCompile(`(\w+):`).FindAllStringSubmatch(stateStr, -1)
		for _, match := range stateNames {
			if len(match) > 1 {
				state = append(state, StateInfo{Name: match[1]})
			}
		}
	}

	return state
}

// extractReactMethods extracts methods from React component
func extractReactMethods(content string) []string {
	var methods []string

	// Extract function declarations
	funcMatches := regexp.MustCompile(`(?:function\s+|const\s+\w+\s*=\s*)\s*(\w+)\s*\(`).FindAllStringSubmatch(content, -1)
	for _, match := range funcMatches {
		if len(match) > 1 {
			methodName := match[1]
			// Skip React hooks and built-in functions
			if !strings.HasPrefix(methodName, "use") && methodName != "function" {
				methods = append(methods, methodName)
			}
		}
	}

	return methods
}

// extractReactDependencies extracts component dependencies
func extractReactDependencies(content string) []string {
	var deps []string

	// Extract import statements
	importMatches := regexp.MustCompile(`import\s+.*?\s+from\s+['"]([^'"]+)['"]`).FindAllStringSubmatch(content, -1)
	for _, match := range importMatches {
		if len(match) > 1 {
			dep := strings.Trim(match[1], "./")
			if !strings.HasPrefix(dep, "react") && !strings.HasPrefix(dep, "@types") {
				deps = append(deps, dep)
			}
		}
	}

	return deps
}

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

// matchesFeature checks if a file name matches feature keywords
func matchesFeature(fileName string, featureName string) bool {
	// If feature name is empty, match all
	if featureName == "" {
		return true
	}

	fileNameLower := strings.ToLower(fileName)
	featureLower := strings.ToLower(featureName)

	// Extract keywords from feature name
	keywords := extractFeatureKeywords(featureLower)

	for _, keyword := range keywords {
		if strings.Contains(fileNameLower, keyword) {
			return true
		}
	}

	return false
}

// extractFeatureKeywords extracts keywords from a feature name
func extractFeatureKeywords(featureName string) []string {
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
