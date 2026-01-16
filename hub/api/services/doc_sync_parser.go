// Package doc_sync_parser - Status marker parsing for documentation synchronization
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

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
