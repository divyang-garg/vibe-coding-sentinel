// Package feature_discovery - React component extraction
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"path/filepath"
	"regexp"
	"strings"
)

// extractReactComponentInfoFromContent extracts React component information
func extractReactComponentInfoFromContent(component *ComponentInfo, content string) {
	// Detect component type
	if isReactFunctionComponentType(content) {
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

// isReactFunctionComponentType checks if content is a React function component
func isReactFunctionComponentType(content string) bool {
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

// isReactClassComponent checks if content is a React class component
func isReactClassComponent(content string) bool {
	return strings.Contains(content, "class ") && strings.Contains(content, "extends React.Component")
}

// isReactFormComponent checks if content is a React form component
func isReactFormComponent(content string) bool {
	return strings.Contains(content, "react-hook-form") ||
		strings.Contains(content, "formik") ||
		strings.Contains(content, "useForm")
}

// isReactPageComponent checks if content is a React page component
func isReactPageComponent(content string) bool {
	fileName := filepath.Base(strings.TrimSuffix(content, filepath.Ext(content)))
	return strings.Contains(strings.ToLower(fileName), "page") ||
		strings.Contains(strings.ToLower(fileName), "index") ||
		strings.Contains(content, "export default function")
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
