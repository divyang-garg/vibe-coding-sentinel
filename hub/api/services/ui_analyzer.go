// Phase 14A: UI Layer Analyzer
// Analyzes UI components for validation, error handling, and accessibility

package services

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// UILayerFinding represents a finding from UI layer analysis
type UILayerFinding struct {
	Type     string `json:"type"`     // "missing_validation", "missing_error_handling", "accessibility_issue"
	Location string `json:"location"` // Component file path
	Issue    string `json:"issue"`
	Severity string `json:"severity"` // "critical", "high", "medium", "low"
}

// analyzeUILayer analyzes UI components based on framework
func analyzeUILayer(ctx context.Context, feature *DiscoveredFeature) ([]UILayerFinding, error) {
	findings := []UILayerFinding{}

	if feature.UILayer == nil {
		return findings, nil
	}

	framework := feature.UILayer.Framework

	for _, component := range feature.UILayer.Components {
		// Read component file
		data, err := os.ReadFile(component.Path)
		if err != nil {
			LogWarn(ctx, "Failed to read component file %s: %v", component.Path, err)
			continue
		}

		content := string(data)

		// Framework-specific analysis
		switch framework {
		case "react", "nextjs":
			reactFindings := analyzeReactComponents(content, component)
			findings = append(findings, reactFindings...)
		case "vue":
			vueFindings := analyzeVueComponents(content, component)
			findings = append(findings, vueFindings...)
		case "angular":
			angularFindings := analyzeAngularComponents(content, component)
			findings = append(findings, angularFindings...)
		}

		// Accessibility checks (framework-agnostic)
		accessibilityFindings := checkAccessibility(content, component)
		findings = append(findings, accessibilityFindings...)
	}

	return findings, nil
}

// analyzeReactComponents analyzes React/Next.js components
func analyzeReactComponents(content string, component ComponentInfo) []UILayerFinding {
	findings := []UILayerFinding{}

	// Check for form validation
	if component.Type == "form" {
		hasValidation := strings.Contains(content, "react-hook-form") ||
			strings.Contains(content, "formik") ||
			strings.Contains(content, "yup") ||
			strings.Contains(content, "zod")

		if !hasValidation {
			findings = append(findings, UILayerFinding{
				Type:     "missing_validation",
				Location: component.Path,
				Issue:    fmt.Sprintf("Form component %s may be missing validation", component.Name),
				Severity: "high",
			})
		}
	}

	// Check for error handling
	hasErrorHandling := strings.Contains(content, "error") &&
		(strings.Contains(content, "catch") || strings.Contains(content, "ErrorBoundary") || strings.Contains(content, "try"))

	if !hasErrorHandling {
		findings = append(findings, UILayerFinding{
			Type:     "missing_error_handling",
			Location: component.Path,
			Issue:    fmt.Sprintf("Component %s may be missing error handling", component.Name),
			Severity: "medium",
		})
	}

	// Check for loading states
	if strings.Contains(content, "useState") || strings.Contains(content, "useEffect") {
		hasLoadingState := strings.Contains(content, "loading") || strings.Contains(content, "isLoading") || strings.Contains(content, "pending")

		if !hasLoadingState && strings.Contains(content, "fetch") || strings.Contains(content, "axios") {
			findings = append(findings, UILayerFinding{
				Type:     "missing_loading_state",
				Location: component.Path,
				Issue:    fmt.Sprintf("Component %s makes async calls but may be missing loading state", component.Name),
				Severity: "low",
			})
		}
	}

	return findings
}

// analyzeVueComponents analyzes Vue components
func analyzeVueComponents(content string, component ComponentInfo) []UILayerFinding {
	findings := []UILayerFinding{}

	// Check for form validation
	if component.Type == "form" {
		hasValidation := strings.Contains(content, "vuelidate") ||
			strings.Contains(content, "vee-validate") ||
			strings.Contains(content, "v-validate")

		if !hasValidation {
			findings = append(findings, UILayerFinding{
				Type:     "missing_validation",
				Location: component.Path,
				Issue:    fmt.Sprintf("Form component %s may be missing validation", component.Name),
				Severity: "high",
			})
		}
	}

	// Check for error handling in template
	hasErrorDisplay := strings.Contains(content, "v-if") && strings.Contains(content, "error")

	if !hasErrorDisplay && (strings.Contains(content, "$emit") || strings.Contains(content, "axios")) {
		findings = append(findings, UILayerFinding{
			Type:     "missing_error_handling",
			Location: component.Path,
			Issue:    fmt.Sprintf("Component %s may be missing error display in template", component.Name),
			Severity: "medium",
		})
	}

	return findings
}

// analyzeAngularComponents analyzes Angular components
func analyzeAngularComponents(content string, component ComponentInfo) []UILayerFinding {
	findings := []UILayerFinding{}

	// Check for form validation
	if component.Type == "form" {
		hasValidation := strings.Contains(content, "FormGroup") ||
			strings.Contains(content, "Validators") ||
			strings.Contains(content, "ngForm")

		if !hasValidation {
			findings = append(findings, UILayerFinding{
				Type:     "missing_validation",
				Location: component.Path,
				Issue:    fmt.Sprintf("Form component %s may be missing validation", component.Name),
				Severity: "high",
			})
		}
	}

	// Check for error handling
	hasErrorHandling := strings.Contains(content, "catchError") ||
		strings.Contains(content, "error") && strings.Contains(content, "subscribe")

	if !hasErrorHandling && strings.Contains(content, "HttpClient") {
		findings = append(findings, UILayerFinding{
			Type:     "missing_error_handling",
			Location: component.Path,
			Issue:    fmt.Sprintf("Component %s may be missing error handling for HTTP calls", component.Name),
			Severity: "medium",
		})
	}

	return findings
}

// checkAccessibility checks for accessibility issues
func checkAccessibility(content string, component ComponentInfo) []UILayerFinding {
	findings := []UILayerFinding{}

	// Check for ARIA attributes
	hasAriaLabels := strings.Contains(content, "aria-label") ||
		strings.Contains(content, "aria-labelledby") ||
		strings.Contains(content, "aria-describedby")

	// Check for semantic HTML
	hasSemanticHTML := strings.Contains(content, "<button") ||
		strings.Contains(content, "<input") ||
		strings.Contains(content, "<select") ||
		strings.Contains(content, "<textarea")

	// Check for images without alt text
	if strings.Contains(content, "<img") && !strings.Contains(content, "alt=") {
		findings = append(findings, UILayerFinding{
			Type:     "accessibility_issue",
			Location: component.Path,
			Issue:    fmt.Sprintf("Component %s contains images without alt text", component.Name),
			Severity: "high",
		})
	}

	// Check for buttons/clickable elements without labels
	if (strings.Contains(content, "onClick") || strings.Contains(content, "@click") || strings.Contains(content, "(click)")) &&
		!hasAriaLabels && !hasSemanticHTML {
		findings = append(findings, UILayerFinding{
			Type:     "accessibility_issue",
			Location: component.Path,
			Issue:    fmt.Sprintf("Component %s may have clickable elements without proper labels", component.Name),
			Severity: "medium",
		})
	}

	// Check for focus management
	if strings.Contains(content, "modal") || strings.Contains(content, "dialog") || strings.Contains(content, "popup") {
		hasFocusManagement := strings.Contains(content, "focus") || strings.Contains(content, "tabIndex")

		if !hasFocusManagement {
			findings = append(findings, UILayerFinding{
				Type:     "accessibility_issue",
				Location: component.Path,
				Issue:    fmt.Sprintf("Component %s appears to be a modal/dialog but may be missing focus management", component.Name),
				Severity: "medium",
			})
		}
	}

	return findings
}
