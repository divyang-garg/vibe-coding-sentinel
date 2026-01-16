// Fixed import structure
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"sentinel-hub-api/pkg/database"
)

// ChecklistItem represents an actionable item in the analysis checklist
type ChecklistItem struct {
	ID          string `json:"id"`
	Category    string `json:"category"` // "business", "ui", "api", "database", "logic", "integration", "tests"
	Severity    string `json:"severity"` // "critical", "high", "medium", "low"
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Remediation string `json:"remediation"`
	AutoFixable bool   `json:"auto_fixable"`
}

// AnalysisSummary contains summary statistics
type AnalysisSummary struct {
	TotalFindings int            `json:"total_findings"`
	BySeverity    map[string]int `json:"by_severity"`
	ByLayer       map[string]int `json:"by_layer"`
	FlowsVerified int            `json:"flows_verified"`
	FlowsBroken   int            `json:"flows_broken"`
	AnalysisTime  time.Duration  `json:"analysis_time"`
}

// CombinedFindings contains all findings from all layers
type CombinedFindings struct {
	Business    []BusinessContextFinding  `json:"business"`
	UI          []UILayerFinding          `json:"ui"`
	API         []APILayerFinding         `json:"api"`
	Database    []DatabaseLayerFinding    `json:"database"`
	Logic       []LogicLayerFinding       `json:"logic"`
	Integration []IntegrationLayerFinding `json:"integration"`
	Test        []TestLayerFinding        `json:"test"`
}

// ComprehensiveAnalysisReport contains the complete analysis results
type ComprehensiveAnalysisReport struct {
	ValidationID  string                 `json:"validation_id"`
	Feature       string                 `json:"feature"`
	Mode          string                 `json:"mode"`
	Depth         string                 `json:"depth"`
	Summary       *AnalysisSummary       `json:"summary"`
	Findings      CombinedFindings       `json:"findings"`
	Checklist     []ChecklistItem        `json:"checklist"`
	LayerAnalysis map[string]interface{} `json:"layer_analysis"`
	EndToEndFlows []interface{}          `json:"end_to_end_flows,omitempty"`
	HubURL        string                 `json:"hub_url,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

// generateChecklist aggregates findings from all layers into prioritized checklist
func generateChecklist(
	businessFindings []BusinessContextFinding,
	uiFindings []UILayerFinding,
	apiFindings []APILayerFinding,
	dbFindings []DatabaseLayerFinding,
	logicFindings []LogicLayerFinding,
	integrationFindings []IntegrationLayerFinding,
	testFindings []TestLayerFinding,
) []ChecklistItem {
	checklist := []ChecklistItem{}

	// Process business context findings
	for _, finding := range businessFindings {
		item := ChecklistItem{
			ID:          generateID(),
			Category:    "business",
			Severity:    finding.Severity,
			Title:       finding.RuleTitle,
			Description: finding.Issue,
			Location:    finding.Location,
			Remediation: generateRemediation(finding.Type, finding.Issue),
			AutoFixable: false, // Business rule violations require manual review
		}
		checklist = append(checklist, item)
	}

	// Process UI layer findings
	for _, finding := range uiFindings {
		item := ChecklistItem{
			ID:          generateID(),
			Category:    "ui",
			Severity:    finding.Severity,
			Title:       getFindingTitle(finding.Type),
			Description: finding.Issue,
			Location:    finding.Location,
			Remediation: generateRemediation(finding.Type, finding.Issue),
			AutoFixable: finding.Type == "accessibility_issue", // Some UI issues can be auto-fixed
		}
		checklist = append(checklist, item)
	}

	// Process API layer findings
	for _, finding := range apiFindings {
		item := ChecklistItem{
			ID:          generateID(),
			Category:    "api",
			Severity:    finding.Severity,
			Title:       getFindingTitle(finding.Type),
			Description: finding.Issue,
			Location:    finding.Location,
			Remediation: generateRemediation(finding.Type, finding.Issue),
			AutoFixable: false, // API issues require code changes
		}
		checklist = append(checklist, item)
	}

	// Process database layer findings
	for _, finding := range dbFindings {
		item := ChecklistItem{
			ID:          generateID(),
			Category:    "database",
			Severity:    finding.Severity,
			Title:       getFindingTitle(finding.Type),
			Description: finding.Issue,
			Location:    finding.Location,
			Remediation: generateRemediation(finding.Type, finding.Issue),
			AutoFixable: finding.Type == "missing_index", // Index creation can be automated
		}
		checklist = append(checklist, item)
	}

	// Process logic layer findings
	for _, finding := range logicFindings {
		item := ChecklistItem{
			ID:          generateID(),
			Category:    "logic",
			Severity:    finding.Severity,
			Title:       getFindingTitle(finding.Type),
			Description: finding.Issue,
			Location:    finding.Location,
			Remediation: generateRemediation(finding.Type, finding.Issue),
			AutoFixable: false, // Logic issues require manual review
		}
		checklist = append(checklist, item)
	}

	// Process integration layer findings
	for _, finding := range integrationFindings {
		item := ChecklistItem{
			ID:          generateID(),
			Category:    "integration",
			Severity:    finding.Severity,
			Title:       getFindingTitle(finding.Type),
			Description: finding.Issue,
			Location:    finding.Location,
			Remediation: generateRemediation(finding.Type, finding.Issue),
			AutoFixable: false, // Integration issues require manual review
		}
		checklist = append(checklist, item)
	}

	// Process test layer findings
	for _, finding := range testFindings {
		item := ChecklistItem{
			ID:          generateID(),
			Category:    "tests",
			Severity:    finding.Severity,
			Title:       getFindingTitle(finding.Type),
			Description: finding.Issue,
			Location:    finding.Location,
			Remediation: generateRemediation(finding.Type, finding.Issue),
			AutoFixable: false, // Test issues require manual review
		}
		checklist = append(checklist, item)
	}

	return checklist
}

// getFindingTitle returns a human-readable title for a finding type
func getFindingTitle(findingType string) string {
	titleMap := map[string]string{
		// UI findings
		"missing_validation":     "Missing Input Validation",
		"missing_error_handling": "Missing Error Handling",
		"accessibility_issue":    "Accessibility Issue",
		// API findings
		"missing_auth":      "Missing Authentication",
		"contract_mismatch": "API Contract Mismatch",
		// Database findings
		"missing_constraint":   "Missing Database Constraint",
		"missing_index":        "Missing Database Index",
		"data_integrity_issue": "Data Integrity Issue",
		// Logic findings
		"semantic_error":     "Semantic Error",
		"signature_mismatch": "Function Signature Mismatch",
		// Integration findings
		"missing_retry": "Missing Retry Logic",
		// Test findings
		"missing_coverage":   "Missing Test Coverage",
		"weak_assertion":     "Weak Test Assertion",
		"missing_edge_case":  "Missing Edge Case Test",
		"missing_error_case": "Missing Error Case Test",
	}

	if title, ok := titleMap[findingType]; ok {
		return title
	}
	return findingType // Fallback to type name
}

// generateSummary creates summary statistics from all findings
func generateSummary(
	checklist []ChecklistItem,
	flowsVerified int,
	flowsBroken int,
	analysisTime time.Duration,
) *AnalysisSummary {
	summary := &AnalysisSummary{
		TotalFindings: len(checklist),
		BySeverity:    make(map[string]int),
		ByLayer:       make(map[string]int),
		FlowsVerified: flowsVerified,
		FlowsBroken:   flowsBroken,
		AnalysisTime:  analysisTime,
	}

	// Count by severity
	for _, item := range checklist {
		summary.BySeverity[item.Severity]++
		summary.ByLayer[item.Category]++
	}

	return summary
}

// formatReport combines all analysis results into a comprehensive report
func formatReport(
	ctx context.Context,
	projectID string,
	feature string,
	mode string,
	depth string,
	checklist []ChecklistItem,
	summary *AnalysisSummary,
	layerAnalysis map[string]interface{},
	endToEndFlows []interface{},
	hubURLBase string,
) (*ComprehensiveAnalysisReport, error) {
	// Generate validation ID
	validationID := generateValidationID()

	// Generate Hub URL
	hubURL := fmt.Sprintf("%s/validations/%s", hubURLBase, validationID)

	now := time.Now()

	// Extract findings from layerAnalysis
	combinedFindings := CombinedFindings{}
	if layerAnalysis != nil {
		if business, ok := layerAnalysis["business"].([]BusinessContextFinding); ok {
			combinedFindings.Business = business
		}
		if ui, ok := layerAnalysis["ui"].([]UILayerFinding); ok {
			combinedFindings.UI = ui
		}
		if api, ok := layerAnalysis["api"].([]APILayerFinding); ok {
			combinedFindings.API = api
		}
		if database, ok := layerAnalysis["database"].([]DatabaseLayerFinding); ok {
			combinedFindings.Database = database
		}
		if logic, ok := layerAnalysis["logic"].([]LogicLayerFinding); ok {
			combinedFindings.Logic = logic
		}
		if integration, ok := layerAnalysis["integration"].([]IntegrationLayerFinding); ok {
			combinedFindings.Integration = integration
		}
		if test, ok := layerAnalysis["test"].([]TestLayerFinding); ok {
			combinedFindings.Test = test
		}
	}

	report := &ComprehensiveAnalysisReport{
		ValidationID:  validationID,
		Feature:       feature,
		Mode:          mode,
		Depth:         depth,
		Summary:       summary,
		Findings:      combinedFindings,
		Checklist:     checklist,
		LayerAnalysis: layerAnalysis,
		EndToEndFlows: endToEndFlows,
		HubURL:        hubURL,
		CreatedAt:     now,
		CompletedAt:   &now,
	}

	return report, nil
}

// storeComprehensiveValidation stores the validation report in the database
func storeComprehensiveValidation(ctx context.Context, report *ComprehensiveAnalysisReport, projectID string) error {
	// Marshal JSONB fields
	findingsJSON, err := marshalJSONB(report.Findings)
	if err != nil {
		return fmt.Errorf("failed to marshal findings: %w", err)
	}

	summaryJSON, err := marshalJSONB(report.Summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	layerAnalysisJSON, err := marshalJSONB(report.LayerAnalysis)
	if err != nil {
		return fmt.Errorf("failed to marshal layer analysis: %w", err)
	}

	var flowsJSON string
	if report.EndToEndFlows != nil {
		flowsJSON, err = marshalJSONB(report.EndToEndFlows)
		if err != nil {
			return fmt.Errorf("failed to marshal flows: %w", err)
		}
	}

	checklistJSON, err := marshalJSONB(report.Checklist)
	if err != nil {
		return fmt.Errorf("failed to marshal checklist: %w", err)
	}

	query := `
		INSERT INTO comprehensive_validations (
			project_id, validation_id, feature, mode, depth,
			findings, summary, layer_analysis, end_to_end_flows, checklist,
			created_at, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = database.ExecWithTimeout(ctx, db, query,
		projectID,
		report.ValidationID,
		report.Feature,
		report.Mode,
		report.Depth,
		findingsJSON,
		summaryJSON,
		layerAnalysisJSON,
		flowsJSON,
		checklistJSON,
		report.CreatedAt,
		report.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store validation: %w", err)
	}

	return nil
}

// Helper functions

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateValidationID() string {
	// Format: VAL-XXX (3 random hex bytes)
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("VAL-%s", hex.EncodeToString(b))
}

func generateRemediation(findingType string, issue string) string {
	// Generate remediation suggestions based on finding type
	switch findingType {
	case "business_rule_violation":
		return "Review business rule and ensure code implementation matches requirements"
	case "user_journey_mismatch":
		return "Implement missing journey steps or update journey documentation"
	case "entity_validation_failure":
		return "Verify entity structure matches documented schema"
	default:
		return "Review and fix the identified issue"
	}
}
