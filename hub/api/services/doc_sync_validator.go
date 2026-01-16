// Package doc_sync_validator - Validation functions for documentation synchronization
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// validateFeatureFlags validates feature flags match documentation
func validateFeatureFlags(featuresDocPath string, codebasePath string) []Discrepancy {
	var discrepancies []Discrepancy

	content, err := os.ReadFile(featuresDocPath)
	if err != nil {
		log.Printf("Failed to read FEATURES.md: %v", err)
		return discrepancies
	}

	// Extract flags from documentation (look for --flag patterns)
	flagPattern := regexp.MustCompile(`--(\w+)\s*`)
	docFlags := make(map[string]bool)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		matches := flagPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				docFlags[match[1]] = true
			}
		}
	}

	// Search codebase for flag usage
	agentPath := filepath.Join(codebasePath, "synapsevibsentinel.sh")
	agentContent, err := os.ReadFile(agentPath)
	if err != nil {
		log.Printf("Failed to read agent file: %v", err)
		return discrepancies
	}

	codeFlags := make(map[string]bool)
	codeLines := strings.Split(string(agentContent), "\n")
	for _, line := range codeLines {
		matches := flagPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				codeFlags[match[1]] = true
			}
		}
	}

	// Find undocumented flags
	for flag := range codeFlags {
		if !docFlags[flag] {
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "missing_doc",
				Feature:        fmt.Sprintf("Flag: --%s", flag),
				DocStatus:      "MISSING",
				CodeStatus:     "EXISTS",
				Recommendation: fmt.Sprintf("Document flag --%s in FEATURES.md", flag),
			})
		}
	}

	// Find documented but missing flags
	for flag := range docFlags {
		if !codeFlags[flag] {
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "missing_impl",
				Feature:        fmt.Sprintf("Flag: --%s", flag),
				DocStatus:      "DOCUMENTED",
				CodeStatus:     "MISSING",
				Recommendation: fmt.Sprintf("Implement flag --%s or remove from documentation", flag),
			})
		}
	}

	return discrepancies
}

// validateAPIEndpoints validates API endpoints match documentation
func validateAPIEndpoints(roadmapPath string, mainGoPath string) []Discrepancy {
	var discrepancies []Discrepancy

	// Extract endpoints from roadmap
	content, err := os.ReadFile(roadmapPath)
	if err != nil {
		log.Printf("Failed to read roadmap: %v", err)
		return discrepancies
	}

	// Look for endpoint tables in roadmap
	endpointPattern := regexp.MustCompile(`\|\s*(POST|GET|PUT|DELETE)\s*\|\s*(/api/v1/[^\s|]+)`)
	docEndpoints := make(map[string]bool)
	matches := endpointPattern.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) >= 3 {
			endpoint := fmt.Sprintf("%s %s", match[1], match[2])
			docEndpoints[endpoint] = true
		}
	}

	// Extract endpoints from main.go
	mainContent, err := os.ReadFile(mainGoPath)
	if err != nil {
		log.Printf("Failed to read main.go: %v", err)
		return discrepancies
	}

	codeEndpoints := make(map[string]bool)
	codeEndpointPattern := regexp.MustCompile(`r\.(Post|Get|Put|Delete)\(["']([^"']+)["']`)
	codeMatches := codeEndpointPattern.FindAllStringSubmatch(string(mainContent), -1)
	for _, match := range codeMatches {
		if len(match) >= 3 {
			method := strings.ToUpper(match[1])
			path := match[2]
			endpoint := fmt.Sprintf("%s %s", method, path)
			codeEndpoints[endpoint] = true
		}
	}

	// Find undocumented endpoints
	for endpoint := range codeEndpoints {
		if !docEndpoints[endpoint] {
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "missing_doc",
				Feature:        fmt.Sprintf("Endpoint: %s", endpoint),
				DocStatus:      "MISSING",
				CodeStatus:     "EXISTS",
				Recommendation: fmt.Sprintf("Document endpoint %s in IMPLEMENTATION_ROADMAP.md", endpoint),
				FilePath:       mainGoPath,
			})
		}
	}

	// Find documented but missing endpoints
	for endpoint := range docEndpoints {
		if !codeEndpoints[endpoint] {
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "missing_impl",
				Feature:        fmt.Sprintf("Endpoint: %s", endpoint),
				DocStatus:      "DOCUMENTED",
				CodeStatus:     "MISSING",
				Recommendation: fmt.Sprintf("Implement endpoint %s or remove from documentation", endpoint),
				FilePath:       roadmapPath,
			})
		}
	}

	return discrepancies
}

// validateCommands validates commands match documentation
func validateCommands(featuresDocPath string, agentPath string) []Discrepancy {
	var discrepancies []Discrepancy

	content, err := os.ReadFile(featuresDocPath)
	if err != nil {
		log.Printf("Failed to read FEATURES.md: %v", err)
		return discrepancies
	}

	// Extract commands from documentation (look for command tables)
	commandPattern := regexp.MustCompile(`\|\s*` + "`" + `(\w+)` + "`" + `\s*\|`)
	docCommands := make(map[string]bool)
	matches := commandPattern.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) > 1 {
			cmd := strings.TrimSpace(match[1])
			if cmd != "Command" && cmd != "command" {
				docCommands[cmd] = true
			}
		}
	}

	// Extract commands from agent code
	agentContent, err := os.ReadFile(agentPath)
	if err != nil {
		log.Printf("Failed to read agent file: %v", err)
		return discrepancies
	}

	codeCommands := make(map[string]bool)
	casePattern := regexp.MustCompile(`case\s+["'](\w+)["']:`)
	caseMatches := casePattern.FindAllStringSubmatch(string(agentContent), -1)
	for _, match := range caseMatches {
		if len(match) > 1 {
			codeCommands[match[1]] = true
		}
	}

	// Find undocumented commands
	for cmd := range codeCommands {
		if !docCommands[cmd] {
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "missing_doc",
				Feature:        fmt.Sprintf("Command: %s", cmd),
				DocStatus:      "MISSING",
				CodeStatus:     "EXISTS",
				Recommendation: fmt.Sprintf("Document command '%s' in FEATURES.md", cmd),
				FilePath:       agentPath,
			})
		}
	}

	// Find documented but missing commands
	for cmd := range docCommands {
		if !codeCommands[cmd] {
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "missing_impl",
				Feature:        fmt.Sprintf("Command: %s", cmd),
				DocStatus:      "DOCUMENTED",
				CodeStatus:     "MISSING",
				Recommendation: fmt.Sprintf("Implement command '%s' or remove from documentation", cmd),
				FilePath:       featuresDocPath,
			})
		}
	}

	return discrepancies
}

// validateTestCoverage validates test coverage for documented features
func validateTestCoverage(markers []StatusMarker, testDir string) []Discrepancy {
	var discrepancies []Discrepancy

	for _, marker := range markers {
		if marker.Status == "COMPLETE" {
			// Check if tests exist for this phase
			phaseLower := strings.ToLower(marker.Phase)
			testPattern := regexp.MustCompile(`.*test.*` + regexp.QuoteMeta(phaseLower) + `.*\.(go|sh)$`)

			testFiles, err := findTestFiles(testDir, testPattern)
			if err != nil || len(testFiles) == 0 {
				discrepancies = append(discrepancies, Discrepancy{
					Type:           "tests_missing",
					Phase:          marker.Phase,
					DocStatus:      marker.Status,
					CodeStatus:     "MISSING_TESTS",
					Recommendation: fmt.Sprintf("Add tests for %s - phase is marked COMPLETE but no tests found", marker.Phase),
					FilePath:       marker.FilePath,
					LineNumber:     marker.Line,
				})
			}
		}
	}

	return discrepancies
}
