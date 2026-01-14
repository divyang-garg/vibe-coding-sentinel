// Phase 11: Code-Documentation Comparison
// Implementation Status Tracking and Business Rules Comparison

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
)

// =============================================================================
// DATA STRUCTURES
// =============================================================================

// StatusMarker represents a phase status marker extracted from documentation
type StatusMarker struct {
	Phase      string      `json:"phase"`
	Line       int         `json:"line"`
	Status     string      `json:"status"`      // "COMPLETE", "PENDING", "STUB", "PARTIAL"
	StatusIcon string      `json:"status_icon"` // "✅", "⏳", "🔴", "⚠️"
	Tasks      []PhaseTask `json:"tasks"`
	FilePath   string      `json:"file_path"`
}

// PhaseTask represents a task within a phase (renamed to avoid collision with Phase 14E Task)
type PhaseTask struct {
	Name     string `json:"name"`
	Days     string `json:"days"`
	Status   string `json:"status"` // "Done", "Pending", etc.
	Line     int    `json:"line"`
	Priority string `json:"priority,omitempty"`
}

// ImplementationEvidence represents evidence of code implementation
type ImplementationEvidence struct {
	Feature     string           `json:"feature"`
	Files       []string         `json:"files"`
	Functions   []string         `json:"functions"`
	Endpoints   []string         `json:"endpoints"`
	Tests       []string         `json:"tests"`
	Confidence  float64          `json:"confidence"`   // 0.0 to 1.0
	LineNumbers map[string][]int `json:"line_numbers"` // file -> line numbers
}

// Discrepancy represents a discrepancy between documentation and code
type Discrepancy struct {
	Type           string                 `json:"type"` // "status_mismatch", "missing_impl", "missing_doc", "partial_match"
	Phase          string                 `json:"phase"`
	Feature        string                 `json:"feature,omitempty"`
	DocStatus      string                 `json:"doc_status"`
	CodeStatus     string                 `json:"code_status"`
	Evidence       ImplementationEvidence `json:"evidence"`
	Recommendation string                 `json:"recommendation"`
	FilePath       string                 `json:"file_path"`
	LineNumber     int                    `json:"line_number"`
}

// DocSyncReport represents a complete doc-sync report
type DocSyncReport struct {
	ID            string        `json:"id"`
	ProjectID     string        `json:"project_id"`
	ReportType    string        `json:"report_type"` // "status_tracking", "business_rules", "all"
	InSync        []InSyncItem  `json:"in_sync"`
	Discrepancies []Discrepancy `json:"discrepancies"`
	Summary       SummaryStats  `json:"summary"`
	CreatedAt     time.Time     `json:"created_at"`
}

// InSyncItem represents an item that is in sync
type InSyncItem struct {
	Phase    string `json:"phase"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

// SummaryStats represents summary statistics
type SummaryStats struct {
	TotalPhases       int     `json:"total_phases"`
	InSyncCount       int     `json:"in_sync_count"`
	DiscrepancyCount  int     `json:"discrepancy_count"`
	AverageConfidence float64 `json:"average_confidence"`
}

// DocSyncRequest represents a request for doc-sync analysis
type DocSyncRequest struct {
	ProjectID  string                 `json:"project_id"`
	ReportType string                 `json:"report_type"` // "status_tracking", "business_rules", "all"
	Options    map[string]interface{} `json:"options,omitempty"`
}

// DocSyncResponse represents the response from doc-sync analysis
type DocSyncResponse struct {
	Success       bool          `json:"success"`
	ReportID      string        `json:"report_id,omitempty"`
	InSync        []InSyncItem  `json:"in_sync"`
	Discrepancies []Discrepancy `json:"discrepancies"`
	Summary       SummaryStats  `json:"summary"`
	Message       string        `json:"message,omitempty"`
}

// =============================================================================
// STATUS MARKER PARSER (Task 1)
// =============================================================================

// parseStatusMarkers parses IMPLEMENTATION_ROADMAP.md for phase status markers
func parseStatusMarkers(docPath string) ([]StatusMarker, error) {
	content, err := os.ReadFile(docPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read documentation file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var markers []StatusMarker
	var currentPhase *StatusMarker
	var inTaskTable bool
	var taskHeaders []string

	// Regex patterns
	phasePattern := regexp.MustCompile(`^## Phase (\d+[A-Z]?):\s*(.+?)\s*([✅⏳🔴⚠️])\s*(.+?)$`)
	statusPattern := regexp.MustCompile(`([✅⏳🔴⚠️])\s*(COMPLETE|COMPLETED|Pending|PENDING|STUB|PARTIAL|DONE)`)
	taskTablePattern := regexp.MustCompile(`^\|.*Task.*\|.*Status.*\|`)
	taskRowPattern := regexp.MustCompile(`^\|(.+?)\|(.+?)\|(.+?)\|`)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Check for phase header
		if matches := phasePattern.FindStringSubmatch(trimmed); matches != nil {
			// Save previous phase if exists
			if currentPhase != nil {
				markers = append(markers, *currentPhase)
			}

			// Extract status from header
			statusIcon := matches[3]
			statusText := matches[4]
			status := normalizeStatus(statusIcon, statusText)

			currentPhase = &StatusMarker{
				Phase:      fmt.Sprintf("Phase %s: %s", matches[1], matches[2]),
				Line:       lineNum,
				Status:     status,
				StatusIcon: statusIcon,
				FilePath:   docPath,
				Tasks:      []PhaseTask{},
			}
			inTaskTable = false
			continue
		}

		// Check for task table header
		if taskTablePattern.MatchString(trimmed) {
			inTaskTable = true
			// Extract headers
			parts := strings.Split(trimmed, "|")
			for _, part := range parts {
				if p := strings.TrimSpace(part); p != "" {
					taskHeaders = append(taskHeaders, p)
				}
			}
			continue
		}

		// Parse task rows
		if inTaskTable && currentPhase != nil {
			if matches := taskRowPattern.FindStringSubmatch(trimmed); matches != nil {
				if len(matches) >= 4 {
					taskName := strings.TrimSpace(matches[1])
					taskDays := ""
					taskStatus := strings.TrimSpace(matches[2])
					taskPriority := ""

					// Skip header rows
					if strings.Contains(strings.ToLower(taskName), "task") ||
						strings.Contains(strings.ToLower(taskName), "days") {
						continue
					}

					// Extract days and priority if present
					if len(matches) >= 3 {
						taskDays = strings.TrimSpace(matches[2])
					}
					if len(matches) >= 4 {
						taskStatus = strings.TrimSpace(matches[3])
					}
					if len(matches) >= 5 {
						taskPriority = strings.TrimSpace(matches[4])
					}

					// Check for status markers in task status
					if statusMatches := statusPattern.FindStringSubmatch(taskStatus); statusMatches != nil {
						taskStatus = normalizeStatus(statusMatches[1], statusMatches[2])
					}

					task := PhaseTask{
						Name:     taskName,
						Days:     taskDays,
						Status:   taskStatus,
						Line:     lineNum,
						Priority: taskPriority,
					}

					currentPhase.Tasks = append(currentPhase.Tasks, task)
				}
			} else if strings.HasPrefix(trimmed, "|") && !strings.Contains(trimmed, "---") {
				// Continue parsing table rows
				continue
			} else {
				// End of table
				inTaskTable = false
				taskHeaders = []string{}
			}
		}

		// Check for standalone status markers
		if currentPhase != nil {
			if statusMatches := statusPattern.FindStringSubmatch(trimmed); statusMatches != nil {
				// Update phase status if found
				if currentPhase.Status == "" {
					currentPhase.Status = normalizeStatus(statusMatches[1], statusMatches[2])
					currentPhase.StatusIcon = statusMatches[1]
				}
			}
		}
	}

	// Add last phase
	if currentPhase != nil {
		markers = append(markers, *currentPhase)
	}

	return markers, nil
}

// normalizeStatus normalizes status text to standard format
func normalizeStatus(icon, text string) string {
	text = strings.ToUpper(strings.TrimSpace(text))

	switch {
	case icon == "✅" || strings.Contains(text, "COMPLETE") || strings.Contains(text, "DONE"):
		return "COMPLETE"
	case icon == "⏳" || strings.Contains(text, "PENDING"):
		return "PENDING"
	case icon == "🔴" || strings.Contains(text, "STUB"):
		return "STUB"
	case icon == "⚠️" || strings.Contains(text, "PARTIAL"):
		return "PARTIAL"
	default:
		return "UNKNOWN"
	}
}

// =============================================================================
// CODE IMPLEMENTATION DETECTOR (Task 2)
// =============================================================================

// detectImplementation scans codebase for implementation evidence
func detectImplementation(feature string, codebasePath string) ImplementationEvidence {
	evidence := ImplementationEvidence{
		Feature:     feature,
		Files:       []string{},
		Functions:   []string{},
		Endpoints:   []string{},
		Tests:       []string{},
		LineNumbers: make(map[string][]int),
	}

	// Search for function definitions
	funcPattern := regexp.MustCompile(`func\s+(\w+).*?\(`)
	featureLower := strings.ToLower(feature)

	// Search in hub/api directory
	hubAPIPath := filepath.Join(codebasePath, "hub", "api")
	if files, err := findGoFiles(hubAPIPath); err == nil {
		for _, file := range files {
			if funcs, lines := findFunctionsInFile(file, funcPattern); len(funcs) > 0 {
				for i, fn := range funcs {
					fnLower := strings.ToLower(fn)
					if strings.Contains(fnLower, featureLower) || strings.Contains(featureLower, fnLower) {
						evidence.Functions = append(evidence.Functions, fn)
						evidence.Files = appendIfNotExists(evidence.Files, file)
						if evidence.LineNumbers[file] == nil {
							evidence.LineNumbers[file] = []int{}
						}
						evidence.LineNumbers[file] = append(evidence.LineNumbers[file], lines[i])
					}
				}
			}
		}
	}

	// Search for API endpoints
	endpointPattern := regexp.MustCompile(`r\.(Post|Get|Put|Delete)\(["']([^"']+)["']`)
	if endpoints, files, lines := findEndpointsInFile(filepath.Join(codebasePath, "hub", "api", "main.go"), endpointPattern); len(endpoints) > 0 {
		for i, endpoint := range endpoints {
			endpointLower := strings.ToLower(endpoint)
			if strings.Contains(endpointLower, featureLower) || strings.Contains(featureLower, endpointLower) {
				evidence.Endpoints = append(evidence.Endpoints, endpoint)
				evidence.Files = appendIfNotExists(evidence.Files, files[i])
				if evidence.LineNumbers[files[i]] == nil {
					evidence.LineNumbers[files[i]] = []int{}
				}
				evidence.LineNumbers[files[i]] = append(evidence.LineNumbers[files[i]], lines[i])
			}
		}
	}

	// Search for test files
	testPattern := regexp.MustCompile(`.*_test\.(go|sh)$`)
	if testFiles, err := findTestFiles(filepath.Join(codebasePath, "tests"), testPattern); err == nil {
		for _, testFile := range testFiles {
			testFileLower := strings.ToLower(testFile)
			if strings.Contains(testFileLower, featureLower) {
				evidence.Tests = append(evidence.Tests, testFile)
			}
		}
	}

	// Calculate confidence score
	evidence.Confidence = calculateConfidence(evidence)

	return evidence
}

// Helper functions for code detection

func findGoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func findFunctionsInFile(filePath string, pattern *regexp.Regexp) ([]string, []int) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(string(content), "\n")
	var funcs []string
	var lineNums []int

	for i, line := range lines {
		if matches := pattern.FindStringSubmatch(line); len(matches) > 1 {
			funcs = append(funcs, matches[1])
			lineNums = append(lineNums, i+1)
		}
	}

	return funcs, lineNums
}

func findEndpointsInFile(filePath string, pattern *regexp.Regexp) ([]string, []string, []int) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, nil
	}

	lines := strings.Split(string(content), "\n")
	var endpoints []string
	var files []string
	var lineNums []int

	for i, line := range lines {
		if matches := pattern.FindStringSubmatch(line); len(matches) >= 3 {
			method := matches[1]
			path := matches[2]
			endpoint := fmt.Sprintf("%s %s", method, path)
			endpoints = append(endpoints, endpoint)
			files = append(files, filePath)
			lineNums = append(lineNums, i+1)
		}
	}

	return endpoints, files, lineNums
}

func findTestFiles(dir string, pattern *regexp.Regexp) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && pattern.MatchString(path) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func appendIfNotExists(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func calculateConfidence(evidence ImplementationEvidence) float64 {
	score := 0.0

	// Base score for function existence
	if len(evidence.Functions) > 0 {
		score += 0.3
	}

	// Additional score for tests
	if len(evidence.Tests) > 0 {
		score += 0.3
	}

	// Additional score for API endpoints
	if len(evidence.Endpoints) > 0 {
		score += 0.3
	}

	// Additional score for multiple files
	if len(evidence.Files) > 1 {
		score += 0.1
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// =============================================================================
// STATUS COMPARISON ENGINE (Task 3)
// =============================================================================

// compareStatus compares documentation status with code evidence
func compareStatus(marker StatusMarker, evidence ImplementationEvidence) []Discrepancy {
	var discrepancies []Discrepancy

	// Determine code status based on evidence
	codeStatus := determineCodeStatus(evidence)

	// Check for mismatches
	if marker.Status == "COMPLETE" && codeStatus == "MISSING" {
		discrepancies = append(discrepancies, Discrepancy{
			Type:           "missing_impl",
			Phase:          marker.Phase,
			DocStatus:      marker.Status,
			CodeStatus:     codeStatus,
			Evidence:       evidence,
			Recommendation: fmt.Sprintf("Implement %s - documentation says COMPLETE but code is missing", marker.Phase),
			FilePath:       marker.FilePath,
			LineNumber:     marker.Line,
		})
	} else if marker.Status == "STUB" && codeStatus == "COMPLETE" {
		discrepancies = append(discrepancies, Discrepancy{
			Type:           "status_mismatch",
			Phase:          marker.Phase,
			DocStatus:      marker.Status,
			CodeStatus:     codeStatus,
			Evidence:       evidence,
			Recommendation: fmt.Sprintf("Update documentation status from STUB to COMPLETE for %s", marker.Phase),
			FilePath:       marker.FilePath,
			LineNumber:     marker.Line,
		})
	} else if marker.Status == "PENDING" && codeStatus == "COMPLETE" {
		discrepancies = append(discrepancies, Discrepancy{
			Type:           "status_mismatch",
			Phase:          marker.Phase,
			DocStatus:      marker.Status,
			CodeStatus:     codeStatus,
			Evidence:       evidence,
			Recommendation: fmt.Sprintf("Update documentation status from PENDING to COMPLETE for %s", marker.Phase),
			FilePath:       marker.FilePath,
			LineNumber:     marker.Line,
		})
	} else if marker.Status == "COMPLETE" && codeStatus == "PARTIAL" {
		discrepancies = append(discrepancies, Discrepancy{
			Type:           "partial_match",
			Phase:          marker.Phase,
			DocStatus:      marker.Status,
			CodeStatus:     codeStatus,
			Evidence:       evidence,
			Recommendation: fmt.Sprintf("Complete implementation for %s - documentation says COMPLETE but code is partial", marker.Phase),
			FilePath:       marker.FilePath,
			LineNumber:     marker.Line,
		})
	}

	return discrepancies
}

// determineCodeStatus determines code status based on evidence
func determineCodeStatus(evidence ImplementationEvidence) string {
	if evidence.Confidence >= 0.9 {
		return "COMPLETE"
	} else if evidence.Confidence >= 0.6 {
		return "PARTIAL"
	} else if evidence.Confidence >= 0.3 {
		return "PARTIAL"
	} else {
		return "MISSING"
	}
}

// =============================================================================
// VALIDATORS (Tasks 4, 5, 6, 7)
// =============================================================================

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

// =============================================================================
// DISCREPANCY REPORT GENERATOR (Task 8)
// =============================================================================

// generateReport generates a formatted doc-sync report
func generateReport(markers []StatusMarker, discrepancies []Discrepancy, projectID string) DocSyncReport {
	reportID := fmt.Sprintf("doc-sync-%d", time.Now().Unix())

	// Build in-sync items
	inSync := []InSyncItem{}
	for _, marker := range markers {
		// Check if this marker has no discrepancies
		hasDiscrepancy := false
		for _, disc := range discrepancies {
			if disc.Phase == marker.Phase {
				hasDiscrepancy = true
				break
			}
		}
		if !hasDiscrepancy {
			inSync = append(inSync, InSyncItem{
				Phase:    marker.Phase,
				Status:   marker.Status,
				Evidence: fmt.Sprintf("Status: %s, Tasks: %d", marker.Status, len(marker.Tasks)),
			})
		}
	}

	// Calculate summary
	summary := SummaryStats{
		TotalPhases:      len(markers),
		InSyncCount:      len(inSync),
		DiscrepancyCount: len(discrepancies),
	}

	if len(discrepancies) > 0 {
		totalConfidence := 0.0
		for _, disc := range discrepancies {
			totalConfidence += disc.Evidence.Confidence
		}
		summary.AverageConfidence = totalConfidence / float64(len(discrepancies))
	}

	return DocSyncReport{
		ID:            reportID,
		ProjectID:     projectID,
		ReportType:    "status_tracking",
		InSync:        inSync,
		Discrepancies: discrepancies,
		Summary:       summary,
		CreatedAt:     time.Now(),
	}
}

// formatReportHumanReadable formats report as human-readable text
func formatReportHumanReadable(report DocSyncReport) string {
	var sb strings.Builder

	sb.WriteString("📋 Documentation-Code Sync Report\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════\n\n")

	// In-sync items
	sb.WriteString("✅ IN SYNC:\n")
	if len(report.InSync) == 0 {
		sb.WriteString("  (none)\n")
	} else {
		for _, item := range report.InSync {
			sb.WriteString(fmt.Sprintf("  - %s (%s)\n", item.Phase, item.Status))
		}
	}

	sb.WriteString("\n")

	// Discrepancies
	sb.WriteString("⚠️  DISCREPANCIES FOUND:\n\n")
	if len(report.Discrepancies) == 0 {
		sb.WriteString("  (none)\n")
	} else {
		for _, disc := range report.Discrepancies {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", disc.Phase, disc.Type))
			sb.WriteString(fmt.Sprintf("    Documentation: %s\n", disc.DocStatus))
			sb.WriteString(fmt.Sprintf("    Code: %s\n", disc.CodeStatus))
			if len(disc.Evidence.Files) > 0 {
				sb.WriteString("    Evidence:\n")
				for _, file := range disc.Evidence.Files {
					sb.WriteString(fmt.Sprintf("      - %s", file))
					if lines, ok := disc.Evidence.LineNumbers[file]; ok && len(lines) > 0 {
						sb.WriteString(fmt.Sprintf(" (lines: %v)", lines))
					}
					sb.WriteString("\n")
				}
			}
			sb.WriteString(fmt.Sprintf("    Recommendation: %s\n", disc.Recommendation))
			sb.WriteString(fmt.Sprintf("    File: %s:%d\n", disc.FilePath, disc.LineNumber))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("═══════════════════════════════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("Summary: %d phases, %d in sync, %d discrepancies\n",
		report.Summary.TotalPhases, report.Summary.InSyncCount, report.Summary.DiscrepancyCount))

	return sb.String()
}

// =============================================================================
// AUTO-UPDATE CAPABILITY (Task 9)
// =============================================================================

// generateUpdateSuggestions generates suggested documentation updates
func generateUpdateSuggestions(discrepancies []Discrepancy) []DocUpdate {
	var updates []DocUpdate

	for _, disc := range discrepancies {
		if disc.Type == "status_mismatch" && disc.CodeStatus == "COMPLETE" {
			update := DocUpdate{
				FilePath:   disc.FilePath,
				LineNumber: disc.LineNumber,
				ChangeType: "status_update",
				OldValue:   disc.DocStatus,
				NewValue:   "COMPLETE",
				Reason:     disc.Recommendation,
			}
			updates = append(updates, update)
		}
	}

	return updates
}

// DocUpdate represents a suggested documentation update
type DocUpdate struct {
	FilePath   string `json:"file_path"`
	LineNumber int    `json:"line_number"`
	ChangeType string `json:"change_type"` // "status_update", "content_update", "add_feature"
	OldValue   string `json:"old_value"`
	NewValue   string `json:"new_value"`
	Reason     string `json:"reason"`
}

// =============================================================================
// BUSINESS RULES COMPARISON (Phase 11B)
// =============================================================================

// compareBusinessRules performs bidirectional comparison between business rules and code
func compareBusinessRules(ctx context.Context, projectID string, codebasePath string) ([]Discrepancy, error) {
	var discrepancies []Discrepancy

	// Extract business rules from knowledge base
	rules, err := extractBusinessRules(ctx, projectID, nil, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract business rules: %w", err)
	}

	// For each rule, check if code implements it
	for _, rule := range rules {
		// Search for rule implementation in code
		evidence := detectBusinessRuleImplementation(rule, codebasePath)

		if evidence.Confidence < 0.3 {
			// Rule documented but not implemented
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "missing_impl",
				Feature:        rule.Title,
				DocStatus:      "DOCUMENTED",
				CodeStatus:     "MISSING",
				Recommendation: fmt.Sprintf("Implement business rule '%s' in code", rule.Title),
			})
		} else if evidence.Confidence < 0.7 {
			// Partially implemented
			discrepancies = append(discrepancies, Discrepancy{
				Type:           "partial_match",
				Feature:        rule.Title,
				DocStatus:      "DOCUMENTED",
				CodeStatus:     "PARTIAL",
				Evidence:       evidence,
				Recommendation: fmt.Sprintf("Complete implementation of business rule '%s'", rule.Title),
			})
		}
	}

	// FUTURE ENHANCEMENT: Reverse check - find code patterns not documented as rules
	// This would require AST analysis to extract business logic patterns from code
	// and compare against documented business rules. This is a bidirectional validation
	// that would help identify undocumented business logic in the codebase.
	// Priority: P2 - Enhancement for future phase

	return discrepancies, nil
}

// detectBusinessRuleImplementation searches codebase for business rule implementation using AST analysis
func detectBusinessRuleImplementation(rule KnowledgeItem, codebasePath string) ImplementationEvidence {
	evidence := ImplementationEvidence{
		Feature:     rule.Title,
		Files:       []string{},
		Functions:   []string{},
		LineNumbers: make(map[string][]int),
	}

	// Extract keywords from rule title
	words := regexp.MustCompile(`\s+|[_-]`).Split(rule.Title, -1)
	var keywords []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 2 {
			wordLower := strings.ToLower(word)
			common := []string{"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}
			isCommon := false
			for _, c := range common {
				if wordLower == c {
					isCommon = true
					break
				}
			}
			if !isCommon {
				keywords = append(keywords, word)
			}
		}
	}

	if len(keywords) == 0 {
		return evidence
	}

	// Build keyword map for faster lookup
	keywordMap := make(map[string]bool)
	for _, kw := range keywords {
		keywordMap[strings.ToLower(kw)] = true
	}

	// Search in Go files using AST analysis
	if files, err := findGoFiles(filepath.Join(codebasePath, "hub", "api")); err == nil {
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			// Try AST-based analysis first
			astMatches := detectBusinessRuleWithAST(string(content), file, keywordMap, keywords)
			if astMatches.Confidence > 0 {
				// AST found matches
				evidence.Files = appendIfNotExists(evidence.Files, file)
				evidence.Functions = append(evidence.Functions, astMatches.Functions...)
				for funcName, lines := range astMatches.LineNumbers {
					evidence.LineNumbers[funcName] = append(evidence.LineNumbers[funcName], lines...)
				}
				evidence.Confidence += astMatches.Confidence
			} else {
				// Fallback to keyword matching
				contentLower := strings.ToLower(string(content))
				matches := 0
				for _, keyword := range keywords {
					if strings.Contains(contentLower, strings.ToLower(keyword)) {
						matches++
					}
				}

				if matches >= len(keywords)/2 {
					evidence.Files = appendIfNotExists(evidence.Files, file)
					evidence.Confidence += 0.2 // Lower confidence for keyword-only matches
				}
			}
		}
	}

	// Cap confidence at 1.0
	if evidence.Confidence > 1.0 {
		evidence.Confidence = 1.0
	}

	return evidence
}

// detectBusinessRuleWithAST uses AST to detect business rule implementation
func detectBusinessRuleWithAST(code string, filePath string, keywordMap map[string]bool, keywords []string) ImplementationEvidence {
	evidence := ImplementationEvidence{
		Files:       []string{},
		Functions:   []string{},
		LineNumbers: make(map[string][]int),
		Confidence:  0.0,
	}

	// Get parser for Go
	parser, err := getParser("go")
	if err != nil {
		return evidence
	}

	// Parse code into AST
	ctx := context.Background()
	tree, err := parser.ParseCtx(ctx, nil, []byte(code))
	if err != nil || tree == nil {
		return evidence
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return evidence
	}

	// Extract function and class definitions from AST
	traverseAST(rootNode, func(node *sitter.Node) bool {
		var funcName string
		var isFunction bool
		var line int

		if node.Type() == "function_declaration" || node.Type() == "method_declaration" {
			// Extract function name
			for i := 0; i < int(node.ChildCount()); i++ {
				child := node.Child(i)
				if child != nil {
					if child.Type() == "identifier" || child.Type() == "field_identifier" {
						funcName = code[child.StartByte():child.EndByte()]
						isFunction = true
						line, _ = getLineColumn(code, int(node.StartByte()))
						break
					}
				}
			}

			if isFunction && funcName != "" {
				funcNameLower := strings.ToLower(funcName)

				// Check if function name matches keywords
				matches := 0
				for keyword := range keywordMap {
					if strings.Contains(funcNameLower, keyword) || strings.Contains(keyword, funcNameLower) {
						matches++
					}
				}

				if matches > 0 {
					// Function name matches - high confidence
					evidence.Functions = appendIfNotExists(evidence.Functions, funcName)
					evidence.LineNumbers[funcName] = append(evidence.LineNumbers[funcName], line)
					evidence.Confidence += 0.5 // High weight for function name matches
				} else {
					// Check function signature and body for keyword matches
					funcCode := code[node.StartByte():node.EndByte()]
					funcCodeLower := strings.ToLower(funcCode)

					keywordMatches := 0
					for _, keyword := range keywords {
						if strings.Contains(funcCodeLower, strings.ToLower(keyword)) {
							keywordMatches++
						}
					}

					if keywordMatches >= len(keywords)/2 {
						// Keywords found in function - medium confidence
						evidence.Functions = appendIfNotExists(evidence.Functions, funcName)
						evidence.LineNumbers[funcName] = append(evidence.LineNumbers[funcName], line)
						evidence.Confidence += 0.3 // Medium weight for keyword matches in function
					}
				}
			}
		}

		return true
	})

	return evidence
}

// =============================================================================
// MAIN DOC-SYNC ANALYSIS FUNCTION
// =============================================================================

// analyzeDocSync performs complete doc-sync analysis
func analyzeDocSync(ctx context.Context, req DocSyncRequest, codebasePath string) (DocSyncResponse, error) {
	roadmapPath := filepath.Join(codebasePath, "docs", "external", "IMPLEMENTATION_ROADMAP.md")

	// Parse status markers
	markers, err := parseStatusMarkers(roadmapPath)
	if err != nil {
		return DocSyncResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse status markers: %v", err),
		}, err
	}

	// Detect implementations for each phase
	var allDiscrepancies []Discrepancy
	for _, marker := range markers {
		evidence := detectImplementation(marker.Phase, codebasePath)
		discrepancies := compareStatus(marker, evidence)
		allDiscrepancies = append(allDiscrepancies, discrepancies...)
	}

	// Validate test coverage
	testCoverageDiscrepancies := validateTestCoverage(markers, filepath.Join(codebasePath, "tests"))
	allDiscrepancies = append(allDiscrepancies, testCoverageDiscrepancies...)

	// Validate feature flags, API endpoints, and commands if requested
	if req.ReportType == "all" || req.ReportType == "status_tracking" {
		featuresDocPath := filepath.Join(codebasePath, "docs", "external", "FEATURES.md")
		flagDiscrepancies := validateFeatureFlags(featuresDocPath, codebasePath)
		allDiscrepancies = append(allDiscrepancies, flagDiscrepancies...)

		mainGoPath := filepath.Join(codebasePath, "hub", "api", "main.go")
		endpointDiscrepancies := validateAPIEndpoints(roadmapPath, mainGoPath)
		allDiscrepancies = append(allDiscrepancies, endpointDiscrepancies...)

		agentPath := filepath.Join(codebasePath, "synapsevibsentinel.sh")
		commandDiscrepancies := validateCommands(featuresDocPath, agentPath)
		allDiscrepancies = append(allDiscrepancies, commandDiscrepancies...)
	}

	// Business rules comparison if requested
	if req.ReportType == "all" || req.ReportType == "business_rules" {
		businessRuleDiscrepancies, err := compareBusinessRules(ctx, req.ProjectID, codebasePath)
		if err != nil {
			log.Printf("Business rules comparison failed: %v", err)
		} else {
			allDiscrepancies = append(allDiscrepancies, businessRuleDiscrepancies...)
		}
	}

	// Generate report
	report := generateReport(markers, allDiscrepancies, req.ProjectID)

	// Store report in database
	reportID, err := storeDocSyncReport(ctx, report)
	if err != nil {
		log.Printf("Failed to store report: %v", err)
		reportID = report.ID
	}

	// Store suggested updates if fix mode is enabled
	var updateCount int
	if fixMode, ok := req.Options["fix"].(bool); ok && fixMode {
		updates := generateUpdateSuggestions(allDiscrepancies)
		if len(updates) > 0 {
			updateIDs, err := storeDocSyncUpdates(ctx, reportID, req.ProjectID, updates)
			if err != nil {
				log.Printf("Failed to store some updates: %v", err)
			}
			updateCount = len(updateIDs)
			log.Printf("Stored %d suggested updates for review", updateCount)
		}
	}

	return DocSyncResponse{
		Success:       true,
		ReportID:      reportID,
		InSync:        report.InSync,
		Discrepancies: report.Discrepancies,
		Summary:       report.Summary,
		Message:       fmt.Sprintf("Analyzed %d phases, found %d discrepancies, stored %d updates", len(markers), len(allDiscrepancies), updateCount),
	}, nil
}

// storeDocSyncReport stores report in database
func storeDocSyncReport(ctx context.Context, report DocSyncReport) (string, error) {
	discrepanciesJSON, err := json.Marshal(report.Discrepancies)
	if err != nil {
		return "", err
	}

	summaryJSON, err := json.Marshal(report.Summary)
	if err != nil {
		return "", err
	}

	query := `
		INSERT INTO doc_sync_reports (id, project_id, report_type, discrepancies, summary, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var reportID string
	err = db.QueryRowContext(ctx, query,
		report.ID,
		report.ProjectID,
		report.ReportType,
		string(discrepanciesJSON),
		string(summaryJSON),
		report.CreatedAt,
	).Scan(&reportID)

	if err != nil {
		return "", fmt.Errorf("failed to store report: %w", err)
	}

	return reportID, nil
}

// storeDocSyncUpdate stores a suggested update in the database
func storeDocSyncUpdate(ctx context.Context, reportID string, projectID string, update DocUpdate) (string, error) {
	query := `
		INSERT INTO doc_sync_updates (id, report_id, project_id, file_path, change_type, old_value, new_value, line_number, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id
	`

	var updateID string
	err := db.QueryRowContext(ctx, query,
		reportID,
		projectID,
		update.FilePath,
		update.ChangeType,
		update.OldValue,
		update.NewValue,
		update.LineNumber,
	).Scan(&updateID)

	if err != nil {
		return "", fmt.Errorf("failed to store update: %w", err)
	}

	return updateID, nil
}

// storeDocSyncUpdates stores multiple updates in the database
func storeDocSyncUpdates(ctx context.Context, reportID string, projectID string, updates []DocUpdate) ([]string, error) {
	var updateIDs []string
	for _, update := range updates {
		updateID, err := storeDocSyncUpdate(ctx, reportID, projectID, update)
		if err != nil {
			log.Printf("Failed to store update for %s: %v", update.FilePath, err)
			continue
		}
		updateIDs = append(updateIDs, updateID)
	}
	return updateIDs, nil
}
