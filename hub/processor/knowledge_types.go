// Phase 13: Structured Knowledge Types
// Defines Go structs matching KNOWLEDGE_SCHEMA.md

package main

import "time"

// StructuredKnowledgeItem represents a knowledge item following the standardized schema
type StructuredKnowledgeItem struct {
	ID          string                 `json:"id"`
	Version     string                 `json:"version"`
	Type        string                 `json:"type"` // business_rule, entity, api_contract, user_journey, glossary
	Status      string                 `json:"status,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Priority    string                 `json:"priority,omitempty"`
	
	// Business Rule specific fields
	Specification *Specification `json:"specification,omitempty"`
	
	// Entity specific fields
	Fields        []EntityField      `json:"fields,omitempty"`
	Relationships []Relationship     `json:"relationships,omitempty"`
	Invariants    []Invariant        `json:"invariants,omitempty"`
	BusinessRules []string           `json:"business_rules,omitempty"`
	
	// API Contract specific fields
	Endpoint      string             `json:"endpoint,omitempty"`
	Method        string             `json:"method,omitempty"`
	Authentication *Authentication   `json:"authentication,omitempty"`
	RateLimiting  *RateLimiting      `json:"rate_limiting,omitempty"`
	Request       *APIRequest        `json:"request,omitempty"`
	Response      map[string]APIResponse `json:"response,omitempty"`
	ImplementsRules []string         `json:"implements_rules,omitempty"`
	SecurityRules   []string         `json:"security_rules,omitempty"`
	
	// User Journey specific fields
	Name          string             `json:"name,omitempty"`
	Actor         string             `json:"actor,omitempty"`
	Goal          string             `json:"goal,omitempty"`
	Preconditions []string           `json:"preconditions,omitempty"`
	Steps         []JourneyStep      `json:"steps,omitempty"`
	Postconditions []string          `json:"postconditions,omitempty"`
	
	// Glossary specific fields
	Term          string             `json:"term,omitempty"`
	Definition    string             `json:"definition,omitempty"`
	Context       string             `json:"context,omitempty"`
	RelatedTerms  []string           `json:"related_terms,omitempty"`
	Examples      []string           `json:"examples,omitempty"`
	
	// Common fields
	TestRequirements []TestRequirement `json:"test_requirements,omitempty"`
	Traceability     TraceabilityInfo `json:"traceability,omitempty"`
	Metadata         *Metadata        `json:"metadata,omitempty"`
	AmbiguityFlags   []AmbiguityFlag  `json:"ambiguity_flags,omitempty"`
}

// Specification represents the specification section of a business rule
type Specification struct {
	Trigger      string      `json:"trigger,omitempty"`
	Preconditions []string   `json:"preconditions,omitempty"`
	Constraints  []Constraint `json:"constraints"`
	Exceptions   []Exception `json:"exceptions,omitempty"`
	SideEffects  []SideEffect `json:"side_effects,omitempty"`
	ErrorCases   []ErrorCase `json:"error_cases,omitempty"`
}

// Constraint represents a constraint in a business rule
type Constraint struct {
	ID          string `json:"id"`
	Type        string `json:"type"` // time_based, value_based, state_based, relationship_based
	Expression  string `json:"expression"`
	Pseudocode  string `json:"pseudocode"`
	Boundary    string `json:"boundary"` // inclusive, exclusive
	Unit        string `json:"unit,omitempty"` // hours, minutes, days, currency, count
}

// Exception represents an exception to a constraint
type Exception struct {
	ID                string   `json:"id"`
	Condition         string   `json:"condition"`
	ModifiedConstraint string  `json:"modified_constraint,omitempty"`
	AppliesTo         []string `json:"applies_to,omitempty"`
	Source            string   `json:"source,omitempty"`
}

// SideEffect represents a side effect of a business rule
type SideEffect struct {
	Action     string                 `json:"action"`
	Condition  string                 `json:"condition"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Async      bool                   `json:"async,omitempty"`
	Required   bool                   `json:"required,omitempty"`
}

// ErrorCase represents an error case for a business rule
type ErrorCase struct {
	Condition    string `json:"condition"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	HTTPStatus   int    `json:"http_status,omitempty"`
	Recoverable  bool   `json:"recoverable,omitempty"`
}

// TestRequirement represents a test requirement for a knowledge item
type TestRequirement struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Type              string                 `json:"type"` // happy_path, error_case, edge_case, exception_case
	Priority          string                 `json:"priority,omitempty"`
	Scenario          string                 `json:"scenario"`
	Setup             map[string]interface{} `json:"setup,omitempty"`
	Action            string                 `json:"action,omitempty"`
	Expected          *ExpectedResult        `json:"expected,omitempty"`
	AssertionsRequired []string              `json:"assertions_required,omitempty"`
}

// ExpectedResult represents the expected result of a test
type ExpectedResult struct {
	Success     bool                   `json:"success,omitempty"`
	ReturnValue map[string]interface{} `json:"return_value,omitempty"`
	SideEffects []string               `json:"side_effects,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// TraceabilityInfo represents traceability information
type TraceabilityInfo struct {
	SourceDocument string    `json:"source_document"`
	SourceSection  string    `json:"source_section,omitempty"`
	SourcePage     int       `json:"source_page,omitempty"`
	SourceQuote    string    `json:"source_quote,omitempty"`
	Stakeholder    string    `json:"stakeholder,omitempty"`
	ApprovedDate   string    `json:"approved_date,omitempty"`
	RelatedRules   []string  `json:"related_rules,omitempty"`
	Implements     []string  `json:"implements,omitempty"`
}

// Metadata represents metadata for a knowledge item
type Metadata struct {
	CreatedAt             time.Time `json:"created_at,omitempty"`
	CreatedBy             string    `json:"created_by,omitempty"`
	Confidence            float64   `json:"confidence,omitempty"`
	NeedsClarification    bool      `json:"needs_clarification,omitempty"`
	ClarificationQuestions []string  `json:"clarification_questions,omitempty"`
}

// AmbiguityFlag represents an ambiguity flag for a knowledge item
type AmbiguityFlag struct {
	Field                string   `json:"field"`
	Interpretations      []string `json:"interpretations"`
	ClarificationQuestion string  `json:"clarification_question,omitempty"`
}

// EntityField represents a field in an entity
type EntityField struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Required    bool                   `json:"required,omitempty"`
	Unique      bool                   `json:"unique,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Description string                 `json:"description,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}

// Relationship represents a relationship between entities
type Relationship struct {
	Entity     string `json:"entity"`
	Type       string `json:"type"` // one-to-one, one-to-many, many-to-many
	ForeignKey string `json:"foreign_key,omitempty"`
	Inverse    string `json:"inverse,omitempty"`
	Cascade    string `json:"cascade,omitempty"` // delete, nullify, restrict
}

// Invariant represents an invariant for an entity
type Invariant struct {
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Message   string `json:"message,omitempty"`
}

// Authentication represents authentication requirements for an API
type Authentication struct {
	Required bool     `json:"required,omitempty"`
	Type     string   `json:"type,omitempty"` // bearer, api_key, basic
	Scopes   []string `json:"scopes,omitempty"`
}

// RateLimiting represents rate limiting for an API
type RateLimiting struct {
	Enabled           bool `json:"enabled,omitempty"`
	RequestsPerMinute int  `json:"requests_per_minute,omitempty"`
}

// APIRequest represents an API request specification
type APIRequest struct {
	Params  map[string]interface{} `json:"params,omitempty"`
	Query   map[string]interface{} `json:"query,omitempty"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Body    map[string]interface{} `json:"body,omitempty"`
}

// APIResponse represents an API response specification
type APIResponse struct {
	Description string                 `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Examples    []interface{}          `json:"examples,omitempty"`
}

// JourneyStep represents a step in a user journey
type JourneyStep struct {
	ID            string          `json:"id"`
	Action        string          `json:"action"`
	SystemResponse string         `json:"system_response"`
	DecisionPoints []DecisionPoint `json:"decision_points,omitempty"`
}

// DecisionPoint represents a decision point in a journey step
type DecisionPoint struct {
	Condition string `json:"condition"`
	NextStep  string `json:"next_step"`
}











