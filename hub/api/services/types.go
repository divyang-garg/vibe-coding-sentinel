// Package services type aliases
// Imports types from models package for use in services
// Complies with CODING_STANDARDS.md: Data Models max 200 lines
package services

import (
	"sentinel-hub-api/feature_discovery"
	"sentinel-hub-api/models"
)

// Type aliases from models package
type Task = models.Task
type VerifyTaskResponse = models.VerifyTaskResponse
type DependencyGraphResponse = models.DependencyGraphResponse
type KnowledgeItem = models.KnowledgeItem
type ChangeType = models.ChangeType
type ChangeRequest = models.ChangeRequest
type TestRequirement = models.TestRequirement
type LLMConfig = models.LLMConfig
type TaskDependency = models.TaskDependency
type Discrepancy = models.Discrepancy
type ImplementationEvidence = models.ImplementationEvidence
type StatusMarker = models.StatusMarker
type DocSyncRequest = models.DocSyncRequest
type DocSyncResponse = models.DocSyncResponse
type DocSyncReport = models.DocSyncReport
type Project = models.Project
type ListTasksRequest = models.ListTasksRequest
type ListTasksResponse = models.ListTasksResponse
type TaskDependencyGraph = models.TaskDependencyGraph
type UpdateTaskRequest = models.UpdateTaskRequest
type PhaseTask = models.PhaseTask
type InSyncItem = models.InSyncItem
type SummaryStats = models.SummaryStats
type LLMUsage = models.LLMUsage

// Type aliases from feature_discovery package
type DiscoveredFeature = feature_discovery.DiscoveredFeature
type EndpointInfo = feature_discovery.EndpointInfo
type ComponentInfo = feature_discovery.ComponentInfo
type BusinessLogicFunctionInfo = feature_discovery.BusinessLogicFunctionInfo
type IntegrationInfo = feature_discovery.IntegrationInfo
type TableInfo = feature_discovery.TableInfo
type DatabaseLayerTables = feature_discovery.DatabaseLayerTables
type TestFileInfo = feature_discovery.TestFileInfo

// ChangeType constants (aliased from models package)
const (
	ChangeNew       = models.ChangeNew
	ChangeModified  = models.ChangeModified
	ChangeRemoved   = models.ChangeRemoved
	ChangeUnchanged = models.ChangeUnchanged
)

// IntentType represents intent classification types
type IntentType string

// IntentType constants
const (
	IntentLocationUnclear IntentType = "location_unclear"
	IntentEntityUnclear   IntentType = "entity_unclear"
	IntentActionConfirm   IntentType = "action_confirm"
	IntentClear           IntentType = "clear"
	IntentAmbiguous       IntentType = "ambiguous"
)

// GapType represents the type of gap
type GapType string

const (
	GapCodeMissing   GapType = "code_missing"
	GapTestsMissing  GapType = "tests_missing"
	GapDocIncomplete GapType = "doc_incomplete"
	GapDocOutdated   GapType = "doc_outdated"
	GapMissingImpl   GapType = "missing_impl"
	GapPartial       GapType = "partial_match"
	GapMissingDoc    GapType = "missing_doc"
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

// GapAnalysisReport contains the complete gap analysis results
type GapAnalysisReport struct {
	ProjectID string                 `json:"project_id"`
	Gaps      []Gap                  `json:"gaps"`
	Summary   map[string]interface{} `json:"summary"`
	CreatedAt string                 `json:"created_at"`
}

// ImpactAnalysis represents the impact analysis result
type ImpactAnalysis struct {
	AffectedCode    []CodeLocation `json:"affected_code"`
	AffectedTests   []TestLocation `json:"affected_tests"`
	EstimatedEffort string         `json:"estimated_effort"`
}

// CodeLocation represents a code location affected by a change
type CodeLocation struct {
	FilePath     string `json:"file_path"`
	FunctionName string `json:"function_name"`
	LineNumbers  []int  `json:"line_numbers"`
}

// TestLocation represents a test location affected by a change
type TestLocation struct {
	FilePath string `json:"file_path"`
	TestName string `json:"test_name"`
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

// ComprehensiveAnalysisRequest represents a request for comprehensive feature analysis
type ComprehensiveAnalysisRequest struct {
	ProjectID              string   `json:"project_id"`
	Feature                string   `json:"feature,omitempty"`
	CodebasePath           string   `json:"codebase_path,omitempty"`
	Mode                   string   `json:"mode,omitempty"`  // "auto" or "manual"
	Depth                  string   `json:"depth,omitempty"` // "shallow" or "deep"
	IncludeBusinessContext bool     `json:"include_business_context,omitempty"`
	Files                  []string `json:"files,omitempty"`
}

// IntentAnalysisRequest represents a request for intent analysis
type IntentAnalysisRequest struct {
	Prompt         string                 `json:"prompt"`
	CodebasePath   string                 `json:"codebase_path"`
	IncludeContext bool                   `json:"include_context,omitempty"`
	ContextData    map[string]interface{} `json:"context_data,omitempty"`
	ProjectID      string                 `json:"project_id,omitempty"`
}

// BusinessRulesAnalysisRequest represents a request for business rules analysis
type BusinessRulesAnalysisRequest struct {
	ProjectID    string   `json:"project_id"`
	CodebasePath string   `json:"codebase_path"`
	RuleIDs      []string `json:"rule_ids,omitempty"`
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

// GapAnalysisRequest represents a request for gap analysis
type GapAnalysisRequest struct {
	ProjectID    string                 `json:"project_id"`
	CodebasePath string                 `json:"codebase_path"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// ListKnowledgeItemsRequest represents a request to list knowledge items
type ListKnowledgeItemsRequest struct {
	ProjectID string `json:"project_id"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// BusinessContextRequest represents a request for business context
type BusinessContextRequest struct {
	ProjectID string   `json:"project_id"`
	Feature   string   `json:"feature,omitempty"`
	Entity    string   `json:"entity,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
}

// BusinessContextResponse represents business context response
type BusinessContextResponse struct {
	Rules            []KnowledgeItem        `json:"rules"`
	Entities         []KnowledgeItem        `json:"entities"`
	UserJourneys     []KnowledgeItem        `json:"user_journeys"`
	Constraints      []string               `json:"constraints"`
	SideEffects      []string               `json:"side_effects"`
	SecurityRules    []string               `json:"security_rules"`
	TestRequirements int                    `json:"test_requirements"`
	Context          map[string]interface{} `json:"context,omitempty"`
}

// SyncKnowledgeRequest represents a request to sync knowledge items
type SyncKnowledgeRequest struct {
	ProjectID        string   `json:"project_id"`
	KnowledgeItemIDs []string `json:"knowledge_item_ids,omitempty"`
	Force            bool     `json:"force,omitempty"`
}

// SyncKnowledgeResponse represents the result of knowledge sync
type SyncKnowledgeResponse struct {
	SyncedCount int      `json:"synced_count"`
	FailedCount int      `json:"failed_count"`
	SyncedItems []string `json:"synced_items,omitempty"`
	FailedItems []string `json:"failed_items,omitempty"`
	Message     string   `json:"message"`
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
