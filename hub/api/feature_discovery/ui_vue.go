// Package feature_discovery - Vue component extraction
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"regexp"
	"strings"
)

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

// isVueFormComponent checks if content is a Vue form component
func isVueFormComponent(content string) bool {
	return strings.Contains(content, "vuelidate") ||
		strings.Contains(content, "vee-validate") ||
		strings.Contains(content, "vue-form")
}
