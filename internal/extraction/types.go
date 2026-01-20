// Package extraction provides LLM-powered knowledge extraction
package extraction

import "time"

// ExtractRequest represents a knowledge extraction request
type ExtractRequest struct {
	Text       string         `json:"text" validate:"required"`
	Source     string         `json:"source"`
	SchemaType string         `json:"schema_type"` // business_rule|entity|api_contract|user_journey|glossary
	Options    ExtractOptions `json:"options"`
}

// ExtractOptions configures extraction behavior
type ExtractOptions struct {
	UseLLM        bool    `json:"use_llm"`
	UseFallback   bool    `json:"use_fallback"`
	MinConfidence float64 `json:"min_confidence"`
}

// ExtractResult contains extraction results
type ExtractResult struct {
	BusinessRules []BusinessRule     `json:"business_rules,omitempty"`
	Entities      []Entity           `json:"entities,omitempty"`
	APIContracts  []APIContract      `json:"api_contracts,omitempty"`
	UserJourneys  []UserJourney      `json:"user_journeys,omitempty"`
	Glossary      []GlossaryTerm     `json:"glossary,omitempty"`
	Confidence    float64            `json:"confidence"`
	Source        string             `json:"source"` // llm|regex
	Errors        []ExtractionError  `json:"errors,omitempty"`
	Metadata      ExtractionMetadata `json:"metadata"`
}

// BusinessRule per KNOWLEDGE_SCHEMA.md
type BusinessRule struct {
	ID, Version, Status, Title, Description, Priority string
	Specification                                     Specification `json:"specification"`
	Traceability                                      Traceability  `json:"traceability"`
	Confidence                                        float64       `json:"confidence"`
}

// Specification contains rule details
type Specification struct {
	Trigger       string       `json:"trigger,omitempty"`
	Preconditions []string     `json:"preconditions,omitempty"`
	Constraints   []Constraint `json:"constraints"`
	Exceptions    []Exception  `json:"exceptions,omitempty"`
	ErrorCases    []ErrorCase  `json:"error_cases,omitempty"`
}

// Constraint represents a business constraint
type Constraint struct {
	ID, Type, Expression, Pseudocode, Boundary, Unit string
}

// Exception represents a rule exception
type Exception struct {
	ID, Condition, ModifiedConstraint string
}

// ErrorCase represents an error scenario
type ErrorCase struct {
	Condition, ErrorCode, ErrorMessage string
	HTTPStatus                         int `json:"http_status,omitempty"`
}

// Traceability links to source document
type Traceability struct {
	SourceDocument, SourceSection, SourceQuote string
}

// ExtractionError represents an extraction error
type ExtractionError struct {
	Code, Message string
}

// ExtractionMetadata contains extraction metadata
type ExtractionMetadata struct {
	ProcessedAt  time.Time `json:"processed_at"`
	TokensUsed   int       `json:"tokens_used"`
	CacheHit     bool      `json:"cache_hit"`
	ProcessingMs int64     `json:"processing_ms"`
}

// Entity represents a domain entity
type Entity struct {
	ID, Version, Status, Name, Description, Category string
	Fields                                           []EntityField  `json:"fields"`
	Relationships                                    []Relationship `json:"relationships,omitempty"`
	Traceability                                     Traceability   `json:"traceability"`
}

// EntityField represents an entity field/attribute
type EntityField struct {
	Name, Type, Description string
	Required, Unique        bool
	Validation              map[string]interface{} `json:"validation,omitempty"`
}

// Relationship represents entity relationships
type Relationship struct {
	Entity, Type, ForeignKey, Inverse, Cascade string
}

// APIContract represents an API endpoint contract
type APIContract struct {
	ID, Version, Status, Endpoint, Method, Description string
	Request                                            RequestSchema  `json:"request,omitempty"`
	Response                                           ResponseSchema `json:"response,omitempty"`
	Traceability                                       Traceability   `json:"traceability"`
}

// RequestSchema represents API request structure
type RequestSchema struct {
	Params map[string]FieldSchema `json:"params,omitempty"`
	Query  map[string]FieldSchema `json:"query,omitempty"`
	Body   map[string]FieldSchema `json:"body,omitempty"`
}

// ResponseSchema represents API response structure
type ResponseSchema struct {
	StatusCodes map[string]StatusResponse `json:"status_codes"`
}

// FieldSchema represents a field in request/response
type FieldSchema struct {
	Type       string
	Required   bool
	Validation map[string]interface{} `json:"validation,omitempty"`
}

// StatusResponse represents a status code response
type StatusResponse struct {
	Description string
	Body        interface{} `json:"body,omitempty"`
}

// UserJourney represents a user journey/flow
type UserJourney struct {
	ID, Version, Status, Name, Actor, Goal, Description string
	Preconditions, Postconditions                       []string
	Steps                                               []JourneyStep `json:"steps"`
	Traceability                                        Traceability  `json:"traceability"`
}

// JourneyStep represents a step in a user journey
type JourneyStep struct {
	Step                                    int
	ActorAction, SystemResponse, Validation string
	BusinessRules, APICalls                 []string
}

// GlossaryTerm represents a glossary term
type GlossaryTerm struct {
	ID, Term, Definition, Context    string
	Synonyms, RelatedTerms, Examples []string
	Traceability                     Traceability `json:"traceability"`
}
