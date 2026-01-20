// Package feature_discovery - Angular component extraction
// Complies with CODING_STANDARDS.md: HTTP Handlers max 300 lines
package feature_discovery

import (
	"regexp"
	"strings"
)

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

// isAngularFormComponent checks if content is an Angular form component
func isAngularFormComponent(content string) bool {
	return strings.Contains(content, "FormGroup") ||
		strings.Contains(content, "ngForm") ||
		strings.Contains(content, "FormBuilder")
}
