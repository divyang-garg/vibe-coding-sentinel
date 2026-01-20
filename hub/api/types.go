// Package main - Sentinel Hub API
//
// This file contains all shared type definitions used across modules.
// Types are organized by phase:
//   - Phase 3: Document
//   - Phase 4: KnowledgeItem
//   - Phase 10: TestRequirement
//   - Phase 12: ChangeRequest, GapAnalysisReport
//   - Phase 14A: ComprehensiveValidation
//   - Phase 14E: Task, TaskDependency, TaskVerification, TaskLink
//
// All types use snake_case for JSON tags to match database column names.
// Status values are defined in constants.go to ensure consistency.

package main

import (
	"time"

	"sentinel-hub-api/feature_discovery"
	"sentinel-hub-api/models"
)

// Type aliases from feature_discovery package
type EndpointInfo = feature_discovery.EndpointInfo
type ComponentInfo = feature_discovery.ComponentInfo
type TableInfo = feature_discovery.TableInfo
type DatabaseLayerTables = feature_discovery.DatabaseLayerTables

// Type aliases from models package
type LLMConfig = models.LLMConfig
type LLMUsage = models.LLMUsage
type Project = models.Project

// ListChangeRequestsRequest represents the request for listing change requests
type ListChangeRequestsRequest struct {
	StatusFilter string `json:"status"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

// ApproveChangeRequestRequest represents the request for approving a change request
type ApproveChangeRequestRequest struct {
	ApprovedBy string `json:"approved_by"`
}

// RejectChangeRequestRequest represents the request for rejecting a change request
type RejectChangeRequestRequest struct {
	RejectedBy string `json:"rejected_by"`
	Reason     string `json:"reason"`
}

// ImpactAnalysisRequest represents the request for impact analysis
type ImpactAnalysisRequest struct {
	CodebasePath string `json:"codebasePath"`
}

// UpdateImplementationRequest represents the request for updating implementation status
type UpdateImplementationRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

// StartImplementationRequest represents the request for starting implementation
type StartImplementationRequest struct {
	Notes string `json:"notes"`
}

// CompleteImplementationRequest represents the request for completing implementation
type CompleteImplementationRequest struct {
	Notes string `json:"notes"`
}

// ListChangeRequestsResponse represents the response for listing change requests
type ListChangeRequestsResponse struct {
	ChangeRequests []ChangeRequest `json:"change_requests"`
	Total          int             `json:"total"`
	Limit          int             `json:"limit"`
	Offset         int             `json:"offset"`
	HasNext        bool            `json:"has_next"`
	HasPrevious    bool            `json:"has_previous"`
}

// ImpactAnalysisResponse represents the response for impact analysis
type ImpactAnalysisResponse struct {
	Success bool            `json:"success"`
	Impact  *ImpactAnalysis `json:"impact"`
}

// GapAnalysisRequest represents the request for gap analysis
type GapAnalysisRequest struct {
	ProjectID    string                 `json:"project_id"`
	CodebasePath string                 `json:"codebasePath"`
	Options      map[string]interface{} `json:"options"`
}

// GapAnalysisResponse represents the response for gap analysis
type GapAnalysisResponse struct {
	Success  bool               `json:"success"`
	Report   *GapAnalysisReport `json:"report"`
	ReportID string             `json:"report_id,omitempty"`
}

// =============================================================================
// Phase 4: Knowledge Base Types
// =============================================================================

// KnowledgeItem represents extracted knowledge from documents (Phase 4)
type KnowledgeItem struct {
	ID             string                 `json:"id"`
	DocumentID     string                 `json:"document_id"`
	Type           string                 `json:"type"` // business_rule, entity, glossary, journey
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	Confidence     float64                `json:"confidence"`
	SourcePage     int                    `json:"source_page,omitempty"`
	Status         string                 `json:"status"` // pending, approved, rejected
	ApprovedBy     *string                `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time             `json:"approved_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	StructuredData map[string]interface{} `json:"structured_data,omitempty"`
}

// =============================================================================
// Phase 10: Test Enforcement Types
// =============================================================================

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

// =============================================================================
// Phase 12: Gap Analysis Types
// =============================================================================

// GapType represents the type of gap
type GapType string

const (
	GapMissingImpl  GapType = "missing_impl"
	GapMissingDoc   GapType = "missing_doc"
	GapPartial      GapType = "partial_match"
	GapTestsMissing GapType = "tests_missing"
)

// Gap represents a single gap between documentation and code
type Gap struct {
	Type            GapType                `json:"type"`
	KnowledgeItemID string                 `json:"knowledge_item_id"`
	RuleTitle       string                 `json:"rule_title"`
	FilePath        string                 `json:"file_path,omitempty"`
	LineNumber      int                    `json:"line_number,omitempty"`
	Description     string                 `json:"description"`
	Evidence        map[string]interface{} `json:"evidence,omitempty"`
	Recommendation  string                 `json:"recommendation"`
	Severity        string                 `json:"severity"` // critical, high, medium, low
}

// GapAnalysisReport contains the complete gap analysis results (Phase 12)
type GapAnalysisReport struct {
	ProjectID string                 `json:"project_id"`
	Gaps      []Gap                  `json:"gaps"`
	Summary   map[string]interface{} `json:"summary"`
	CreatedAt string                 `json:"created_at"`
}

// =============================================================================
// Phase 15: Intent & Simple Language Types
// =============================================================================

// IntentType represents the type of unclear intent
type IntentType string

const (
	IntentLocationUnclear IntentType = "location_unclear"
	IntentEntityUnclear   IntentType = "entity_unclear"
	IntentActionConfirm   IntentType = "action_confirm"
	IntentAmbiguous       IntentType = "ambiguous"
	IntentClear           IntentType = "clear" // No clarification needed
)

// IntentAnalysisRequest represents the request for intent analysis
type IntentAnalysisRequest struct {
	Prompt         string `json:"prompt"`
	CodebasePath   string `json:"codebasePath,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`
	IncludeContext bool   `json:"includeContext,omitempty"`
}

// IntentAnalysisResponse represents the response from intent analysis
type IntentAnalysisResponse struct {
	Success               bool       `json:"success"`
	RequiresClarification bool       `json:"requires_clarification"`
	IntentType            IntentType `json:"intent_type"`
	Confidence            float64    `json:"confidence"`
	DecisionID            string     `json:"decision_id,omitempty"`
	ClarifyingQuestion    string     `json:"clarifying_question,omitempty"`
	Options               []string   `json:"options,omitempty"`
	SuggestedAction       string     `json:"suggested_action,omitempty"`
	ResolvedPrompt        string     `json:"resolved_prompt,omitempty"`
}

// IntentDecisionRequest represents a user's decision
type IntentDecisionRequest struct {
	DecisionID        string                 `json:"decision_id"`
	UserChoice        string                 `json:"user_choice"`
	ResolvedPrompt    string                 `json:"resolved_prompt,omitempty"`
	AdditionalContext map[string]interface{} `json:"additional_context,omitempty"`
}

// IntentDecision represents a stored decision
type IntentDecision struct {
	ID                 string                 `json:"id"`
	ProjectID          string                 `json:"project_id"`
	OriginalPrompt     string                 `json:"original_prompt"`
	IntentType         IntentType             `json:"intent_type"`
	ClarifyingQuestion string                 `json:"clarifying_question"`
	UserChoice         string                 `json:"user_choice"`
	ResolvedPrompt     string                 `json:"resolved_prompt"`
	ContextData        map[string]interface{} `json:"context_data"`
	CreatedAt          string                 `json:"created_at"`
}

// IntentPattern represents a learned pattern
type IntentPattern struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id,omitempty"`
	PatternType string                 `json:"pattern_type"`
	PatternData map[string]interface{} `json:"pattern_data"`
	Frequency   int                    `json:"frequency"`
	LastUsed    string                 `json:"last_used"`
	CreatedAt   string                 `json:"created_at"`
}

// =============================================================================
// Phase 14E: Task Dependency & Verification System Types
// =============================================================================

// Task represents a tracked task
type Task struct {
	ID                     string     `json:"id"`
	ProjectID              string     `json:"project_id"`
	Source                 string     `json:"source"` // 'cursor', 'manual', 'change_request', 'comprehensive_analysis'
	Title                  string     `json:"title"`
	Description            string     `json:"description,omitempty"`
	FilePath               string     `json:"file_path,omitempty"`
	LineNumber             *int       `json:"line_number,omitempty"`
	Status                 string     `json:"status"`   // 'pending', 'in_progress', 'completed', 'blocked'
	Priority               string     `json:"priority"` // 'low', 'medium', 'high', 'critical'
	AssignedTo             *string    `json:"assigned_to,omitempty"`
	EstimatedEffort        *int       `json:"estimated_effort,omitempty"` // in hours
	ActualEffort           *int       `json:"actual_effort,omitempty"`    // in hours
	Tags                   []string   `json:"tags,omitempty"`
	VerificationConfidence float64    `json:"verification_confidence"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	VerifiedAt             *time.Time `json:"verified_at,omitempty"`
	ArchivedAt             *time.Time `json:"archived_at,omitempty"`
	Version                int        `json:"version"` // for optimistic locking
}

// TaskDependency represents a dependency between tasks
type TaskDependency struct {
	ID              string    `json:"id"`
	TaskID          string    `json:"task_id"`
	DependsOnTaskID string    `json:"depends_on_task_id"`
	DependencyType  string    `json:"dependency_type"` // 'explicit', 'implicit', 'integration', 'feature'
	Confidence      float64   `json:"confidence"`      // 0.0-1.0
	CreatedAt       time.Time `json:"created_at"`
}

// TaskVerification represents a verification result for a task
type TaskVerification struct {
	ID               string                 `json:"id"`
	TaskID           string                 `json:"task_id"`
	VerificationType string                 `json:"verification_type"` // 'code_existence', 'code_usage', 'test_coverage', 'integration'
	Status           string                 `json:"status"`            // 'pending', 'verified', 'failed'
	Confidence       float64                `json:"confidence"`        // 0.0-1.0
	Evidence         map[string]interface{} `json:"evidence,omitempty"`
	RetryCount       int                    `json:"retry_count"`
	VerifiedAt       *time.Time             `json:"verified_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

// TaskLink represents a link to other systems
type TaskLink struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	LinkType  string    `json:"link_type"` // 'change_request', 'knowledge_item', 'comprehensive_analysis', 'test_requirement'
	LinkedID  string    `json:"linked_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTaskRequest represents the request for creating a task
type CreateTaskRequest struct {
	Source      string   `json:"source"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	FilePath    string   `json:"file_path,omitempty"`
	LineNumber  *int     `json:"line_number,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	AssignedTo  *string  `json:"assigned_to,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateTaskRequest represents the request for updating a task
type UpdateTaskRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Priority    *string  `json:"priority,omitempty"`
	AssignedTo  *string  `json:"assigned_to,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Version     int      `json:"version"` // for optimistic locking
}

// ListTasksRequest represents the request for listing tasks
type ListTasksRequest struct {
	StatusFilter    string   `json:"status,omitempty"`
	PriorityFilter  string   `json:"priority,omitempty"`
	SourceFilter    string   `json:"source,omitempty"`
	AssignedTo      *string  `json:"assigned_to,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	IncludeArchived *bool    `json:"include_archived,omitempty"`
	Limit           int      `json:"limit"`
	Offset          int      `json:"offset"`
}

// ListTasksResponse represents the response for listing tasks
type ListTasksResponse struct {
	Tasks       []Task `json:"tasks"`
	Total       int    `json:"total"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
	HasNext     bool   `json:"has_next"`
	HasPrevious bool   `json:"has_previous"`
}

// VerifyTaskRequest represents the request for verifying a task
type VerifyTaskRequest struct {
	Force bool `json:"force,omitempty"` // force re-verification
}

// VerifyTaskResponse represents the response for task verification
type VerifyTaskResponse struct {
	TaskID            string                 `json:"task_id"`
	OverallConfidence float64                `json:"overall_confidence"`
	Verifications     []TaskVerification     `json:"verifications"`
	Status            string                 `json:"status"`
	Evidence          map[string]interface{} `json:"evidence,omitempty"`
}

// AddDependencyRequest represents the request for adding a dependency
type AddDependencyRequest struct {
	DependsOnTaskID string  `json:"depends_on_task_id"`
	DependencyType  string  `json:"dependency_type"`
	Confidence      float64 `json:"confidence,omitempty"`
}

// DependencyGraphResponse represents the dependency graph
type DependencyGraphResponse struct {
	TaskID       string              `json:"task_id"`
	Dependencies []TaskDependency    `json:"dependencies"`
	BlockedBy    []string            `json:"blocked_by"`      // task IDs that block this task
	Blocks       []string            `json:"blocks"`          // task IDs blocked by this task
	Graph        map[string][]string `json:"graph,omitempty"` // adjacency list representation
	HasCycle     bool                `json:"has_cycle"`
	CyclePath    []string            `json:"cycle_path,omitempty"`
}

// =============================================================================
// Phase 14A: Comprehensive Feature Analysis Types
// =============================================================================

// ComprehensiveValidation represents a comprehensive feature analysis result (Phase 14A)
type ComprehensiveValidation struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"project_id"`
	ValidationID  string                 `json:"validation_id"`
	Feature       string                 `json:"feature"`
	Mode          string                 `json:"mode"`
	Depth         string                 `json:"depth"`
	Findings      map[string]interface{} `json:"findings"`
	Summary       map[string]interface{} `json:"summary"`
	LayerAnalysis map[string]interface{} `json:"layer_analysis"`
	EndToEndFlows map[string]interface{} `json:"end_to_end_flows,omitempty"`
	Checklist     map[string]interface{} `json:"checklist"`
	CreatedAt     time.Time              `json:"created_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

// =============================================================================
// Phase 12: Change Request Types
// =============================================================================

// ChangeType represents the type of change
type ChangeType string

const (
	ChangeNew       ChangeType = "new"
	ChangeModified  ChangeType = "modification"
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
	ImplementationNotes  *string                `json:"implementation_notes,omitempty"`
	ImpactAnalysis       map[string]interface{} `json:"impact_analysis,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	ApprovedBy           *string                `json:"approved_by,omitempty"`
	ApprovedAt           *time.Time             `json:"approved_at,omitempty"`
	RejectedBy           *string                `json:"rejected_by,omitempty"`
	RejectedAt           *time.Time             `json:"rejected_at,omitempty"`
	RejectionReason      *string                `json:"rejection_reason,omitempty"`
}

// =============================================================================
// Phase 14A: Discovered Feature Types (for flow verification)
// =============================================================================

// Type aliases for backward compatibility
// Note: Actual types are defined in feature_discovery/feature_types.go
type DiscoveredFeature = feature_discovery.DiscoveredFeature
type LogicLayer = feature_discovery.LogicLayer
type BusinessLogicFunctionInfo = feature_discovery.BusinessLogicFunctionInfo
type IntegrationLayer = feature_discovery.IntegrationLayer
type IntegrationInfo = feature_discovery.IntegrationInfo
type TestLayer = feature_discovery.TestLayer
type TestFileInfo = feature_discovery.TestFileInfo
