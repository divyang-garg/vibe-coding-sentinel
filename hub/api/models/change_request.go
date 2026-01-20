// Package models contains data structures for the Hub API
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package models

import "time"

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeNew       ChangeType = "new"
	ChangeModified  ChangeType = "modified"
	ChangeRemoved   ChangeType = "removal"
	ChangeUnchanged ChangeType = "unchanged"
)

// ChangeRequest represents a change request (Phase 12)
type ChangeRequest struct {
	ID                   string                 `json:"id"`
	ProjectID            string                 `json:"project_id"`
	KnowledgeItemID      *string                `json:"knowledge_item_id,omitempty"`
	Type                 ChangeType             `json:"type"`
	CurrentState         map[string]interface{} `json:"current_state,omitempty"`
	ProposedState        map[string]interface{} `json:"proposed_state,omitempty"`
	Status               string                 `json:"status"`                // pending_approval, approved, rejected
	ImplementationStatus string                 `json:"implementation_status"` // pending, in_progress, completed, blocked
	ImplementationNotes  string                 `json:"implementation_notes,omitempty"`
	ImpactAnalysis       map[string]interface{} `json:"impact_analysis,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	ApprovedBy           *string                `json:"approved_by,omitempty"`
	ApprovedAt           *time.Time             `json:"approved_at,omitempty"`
	RejectedBy           *string                `json:"rejected_by,omitempty"`
	RejectedAt           *time.Time             `json:"rejected_at,omitempty"`
	RejectionReason      string                 `json:"rejection_reason,omitempty"`
}

// ComprehensiveValidation represents a comprehensive feature analysis result (Phase 14A)
type ComprehensiveValidation struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"project_id"`
	ValidationID  string                 `json:"validation_id"`
	Feature       string                 `json:"feature"`
	Mode          string                 `json:"mode"`
	Depth         string                 `json:"depth"`
	Findings      map[string]interface{} `json:"findings,omitempty"`
	Summary       map[string]interface{} `json:"summary,omitempty"`
	LayerAnalysis map[string]interface{} `json:"layer_analysis,omitempty"`
	EndToEndFlows []interface{}          `json:"end_to_end_flows,omitempty"`
	Checklist     []interface{}          `json:"checklist,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

// TestRequirement represents a test requirement (Phase 10)
type TestRequirement struct {
	ID              string    `json:"id"`
	KnowledgeItemID string    `json:"knowledge_item_id"`
	RuleTitle       string    `json:"rule_title"`
	RequirementType string    `json:"requirement_type"` // happy_path, edge_case, error_case
	Description     string    `json:"description"`
	CodeFunction    string    `json:"code_function,omitempty"`
	Priority        string    `json:"priority"` // critical, high, medium, low
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
