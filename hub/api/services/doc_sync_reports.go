// Package doc_sync_reports - Report generation for documentation synchronization
// Complies with CODING_STANDARDS.md: Utilities max 250 lines

package services

import (
	"fmt"
	"strings"
	"time"
)

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
